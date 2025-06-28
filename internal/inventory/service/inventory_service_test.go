package service_test

import (
	"context"
	"erp-system/internal/inventory/models"
	invRepoMock "erp-system/internal/inventory/repository/mocks" // Alias for inventory mocks
	"erp-system/internal/inventory/service"
	dto "erp-system/internal/inventory/service/dto"
	app_errors "erp-system/pkg/errors"
	// "fmt" // REMOVED - no longer used directly
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestInventoryService_CreateItem(t *testing.T) {
	mockItemRepo := invRepoMock.NewItemRepositoryMock(t)
	// Other repos can be nil if not used by the method under test
	invService := service.NewInventoryService(mockItemRepo, nil, nil)
	ctx := context.Background()

	req := dto.CreateItemRequest{
		SKU:           "ITEM001",
		Name:          "Test Item 1",
		UnitOfMeasure: "PCS",
		ItemType:      models.FinishedGood,
		IsActive:      true,
	}

	t.Run("Success", func(t *testing.T) {
		mockItemRepo.On("GetBySKU", ctx, req.SKU).Return(nil, app_errors.NewNotFoundError("item_sku", req.SKU)).Once()
		mockItemRepo.On("Create", ctx, mock.AnythingOfType("*models.Item")).Run(func(args mock.Arguments) {
			arg := args.Get(1).(*models.Item)
			assert.Equal(t, req.SKU, arg.SKU)
			arg.ID = uuid.New() // Simulate DB generating ID
		}).Return(func(_ context.Context, item *models.Item) *models.Item { return item }, nil).Once()

		item, err := invService.CreateItem(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, item)
		assert.Equal(t, req.SKU, item.SKU)
		mockItemRepo.AssertExpectations(t)
	})

	t.Run("Error - SKU Exists", func(t *testing.T) {
		existingItem := &models.Item{ID: uuid.New(), SKU: req.SKU}
		mockItemRepo.On("GetBySKU", ctx, req.SKU).Return(existingItem, nil).Once()

		_, err := invService.CreateItem(ctx, req)
		assert.Error(t, err)
		assert.IsType(t, &app_errors.ConflictError{}, err)
		mockItemRepo.AssertExpectations(t)
	})

	t.Run("Error - Invalid Item Type", func(t *testing.T) {
		invalidReq := req
		invalidReq.ItemType = "INVALID_TYPE"
		_, err := invService.CreateItem(ctx, invalidReq)
		assert.Error(t, err)
		assert.IsType(t, &app_errors.ValidationError{}, err)
		assert.Contains(t, err.Error(), "invalid item type")
	})
}

func TestInventoryService_UpdateItem(t *testing.T) {
	mockItemRepo := invRepoMock.NewItemRepositoryMock(t)
	mockTxnRepo := invRepoMock.NewInventoryTransactionRepositoryMock(t) // Needed for stock check on deactivation
	invService := service.NewInventoryService(mockItemRepo, nil, mockTxnRepo)
	ctx := context.Background()

	itemID := uuid.New()
	originalItem := &models.Item{
		ID: itemID, SKU: "ITEM002", Name: "Original Name", ItemType: models.RawMaterial, IsActive: true,
	}
	newName := "Updated Item Name"
	req := dto.UpdateItemRequest{Name: &newName}

	t.Run("Success - Update Name", func(t *testing.T) {
		mockItemRepo.On("GetByID", ctx, itemID).Return(originalItem, nil).Once()
		mockItemRepo.On("Update", ctx, mock.MatchedBy(func(item *models.Item) bool {
			return item.ID == itemID && item.Name == newName
		})).Return(func(_ context.Context, item *models.Item) *models.Item { return item }, nil).Once()

		updatedItem, err := invService.UpdateItem(ctx, itemID, req)
		assert.NoError(t, err)
		assert.Equal(t, newName, updatedItem.Name)
		mockItemRepo.AssertExpectations(t)
	})

    t.Run("Success - Deactivate Item with No Stock", func(t *testing.T) {
        isActiveFalse := false
        reqDeactivate := dto.UpdateItemRequest{IsActive: &isActiveFalse}

        mockItemRepo.On("GetByID", ctx, itemID).Return(originalItem, nil).Once()
        // Mock GetStockLevelsByItem to return empty map (no stock)
        mockTxnRepo.On("GetStockLevelsByItem", ctx, itemID, mock.AnythingOfType("time.Time")).Return(make(map[uuid.UUID]float64), nil).Once()
        mockItemRepo.On("Update", ctx, mock.MatchedBy(func(item *models.Item) bool {
            return item.ID == itemID && !item.IsActive
        })).Return(func(_ context.Context, item *models.Item) *models.Item {
			item.IsActive = false // ensure the mock returns the change
			return item
		}, nil).Once()

        updatedItem, err := invService.UpdateItem(ctx, itemID, reqDeactivate)
        assert.NoError(t, err)
        assert.False(t, updatedItem.IsActive)
        mockItemRepo.AssertExpectations(t)
        mockTxnRepo.AssertExpectations(t)
    })

    t.Run("Warning (Allow) - Deactivate Item with Stock", func(t *testing.T) {
		// This test assumes the service currently WARNS but ALLOWS deactivation with stock.
		// If it were to block, the error type would be ConflictError.
        isActiveFalse := false
        reqDeactivate := dto.UpdateItemRequest{IsActive: &isActiveFalse}

        stockData := make(map[uuid.UUID]float64)
        stockData[uuid.New()] = 10.0 // Item has stock in some warehouse

        mockItemRepo.On("GetByID", ctx, itemID).Return(originalItem, nil).Once()
        mockTxnRepo.On("GetStockLevelsByItem", ctx, itemID, mock.AnythingOfType("time.Time")).Return(stockData, nil).Once()
        mockItemRepo.On("Update", ctx, mock.MatchedBy(func(item *models.Item) bool {
            return item.ID == itemID && !item.IsActive // Check if IsActive is correctly set to false
        })).Return(func(_ context.Context, item *models.Item) *models.Item {
			item.IsActive = false // Simulate update
			return item
		}, nil).Once()


        updatedItem, err := invService.UpdateItem(ctx, itemID, reqDeactivate)
        assert.NoError(t, err) // No error because we allow it with a warning (logged by service)
        assert.False(t, updatedItem.IsActive)
        mockItemRepo.AssertExpectations(t)
        mockTxnRepo.AssertExpectations(t)
    })
}


func TestInventoryService_DeleteItem(t *testing.T) {
    mockItemRepo := invRepoMock.NewItemRepositoryMock(t)
    mockTxnRepo := invRepoMock.NewInventoryTransactionRepositoryMock(t)
    invService := service.NewInventoryService(mockItemRepo, nil, mockTxnRepo)
    ctx := context.Background()
    itemID := uuid.New()

    t.Run("Success - Delete Item with No Stock", func(t *testing.T) {
        itemToDelete := &models.Item{ID: itemID, SKU: "DEL001"}
        mockItemRepo.On("GetByID", ctx, itemID).Return(itemToDelete, nil).Once()
        mockTxnRepo.On("GetStockLevelsByItem", ctx, itemID, mock.AnythingOfType("time.Time")).Return(make(map[uuid.UUID]float64), nil).Once()
        mockItemRepo.On("Delete", ctx, itemID).Return(nil).Once()

        err := invService.DeleteItem(ctx, itemID)
        assert.NoError(t, err)
        mockItemRepo.AssertExpectations(t)
        mockTxnRepo.AssertExpectations(t)
    })

    t.Run("Error - Delete Item with Stock", func(t *testing.T) {
        // Isolate mocks for this sub-test
        mockItemRepoSub := invRepoMock.NewItemRepositoryMock(t)
        mockTxnRepoSub := invRepoMock.NewInventoryTransactionRepositoryMock(t)
        invServiceSub := service.NewInventoryService(mockItemRepoSub, nil, mockTxnRepoSub)
        ctxSub := context.Background()

        itemToDelete := &models.Item{ID: itemID, SKU: "DEL002"} // itemID from parent scope
        stockData := map[uuid.UUID]float64{uuid.New(): 5.0} // Has stock

        mockItemRepoSub.On("GetByID", ctxSub, itemID).Return(itemToDelete, nil).Once()
        mockTxnRepoSub.On("GetStockLevelsByItem", ctxSub, itemID, mock.AnythingOfType("time.Time")).Return(stockData, nil).Once()

        err := invServiceSub.DeleteItem(ctxSub, itemID)
        assert.Error(t, err)
        assert.IsType(t, &app_errors.ConflictError{}, err)
        assert.Contains(t, err.Error(), "cannot delete item DEL002, it has existing stock")

        // Assert expectations on sub-test mocks
        mockItemRepoSub.AssertExpectations(t)
        mockTxnRepoSub.AssertExpectations(t)
        mockItemRepoSub.AssertNotCalled(t, "Delete", ctxSub, itemID)
    })
}


func TestInventoryService_CreateWarehouse(t *testing.T) {
    mockWarehouseRepo := invRepoMock.NewWarehouseRepositoryMock(t)
    invService := service.NewInventoryService(nil, mockWarehouseRepo, nil)
    ctx := context.Background()

    req := dto.CreateWarehouseRequest{
        Code: "WH001", Name: "Main Warehouse", IsActive: true,
    }

    t.Run("Success", func(t *testing.T) {
        mockWarehouseRepo.On("GetByCode", ctx, req.Code).Return(nil, app_errors.NewNotFoundError("wh_code", req.Code)).Once()
        mockWarehouseRepo.On("Create", ctx, mock.AnythingOfType("*models.Warehouse")).Return(&models.Warehouse{ID: uuid.New(), Code: req.Code, Name: req.Name}, nil).Once()

        wh, err := invService.CreateWarehouse(ctx, req)
        assert.NoError(t, err)
        assert.NotNil(t, wh)
        assert.Equal(t, req.Code, wh.Code)
        mockWarehouseRepo.AssertExpectations(t)
    })

    t.Run("Error - Code Exists", func(t *testing.T) {
        existingWH := &models.Warehouse{ID: uuid.New(), Code: req.Code}
        mockWarehouseRepo.On("GetByCode", ctx, req.Code).Return(existingWH, nil).Once()

        _, err := invService.CreateWarehouse(ctx, req)
        assert.Error(t, err)
        assert.IsType(t, &app_errors.ConflictError{}, err)
        mockWarehouseRepo.AssertExpectations(t)
    })
}


func TestInventoryService_CreateInventoryAdjustment(t *testing.T) {
    mockItemRepo := invRepoMock.NewItemRepositoryMock(t)
    mockWarehouseRepo := invRepoMock.NewWarehouseRepositoryMock(t)
    mockTxnRepo := invRepoMock.NewInventoryTransactionRepositoryMock(t)
    invService := service.NewInventoryService(mockItemRepo, mockWarehouseRepo, mockTxnRepo)
    ctx := context.Background()

    itemID := uuid.New()
    warehouseID := uuid.New()
    item := &models.Item{ID: itemID, SKU: "ADJITEM", IsActive: true, ItemType: models.FinishedGood}
    warehouse := &models.Warehouse{ID: warehouseID, Code: "ADJWH", IsActive: true}

    req := dto.CreateInventoryAdjustmentRequest{
        ItemID:         itemID,
        WarehouseID:    warehouseID,
        AdjustmentType: models.AdjustStockIn,
        Quantity:       10.0,
    }

    t.Run("Success - AdjustStockIn", func(t *testing.T) {
        mockItemRepo.On("GetByID", ctx, itemID).Return(item, nil).Once()
        mockWarehouseRepo.On("GetByID", ctx, warehouseID).Return(warehouse, nil).Once()
        mockTxnRepo.On("Create", ctx, mock.AnythingOfType("*models.InventoryTransaction")).Return(&models.InventoryTransaction{ID: uuid.New(), ItemID: itemID, WarehouseID: warehouseID, Quantity: 10.0, TransactionType: models.AdjustStockIn}, nil).Once()

        txn, err := invService.CreateInventoryAdjustment(ctx, req)
        assert.NoError(t, err)
        assert.NotNil(t, txn)
        assert.Equal(t, models.AdjustStockIn, txn.TransactionType)
        mockItemRepo.AssertExpectations(t)
        mockWarehouseRepo.AssertExpectations(t)
        mockTxnRepo.AssertExpectations(t)
    })

    t.Run("Success - AdjustStockOut (allows negative for now)", func(t *testing.T) {
        reqOut := req
        reqOut.AdjustmentType = models.AdjustStockOut
        mockItemRepo.On("GetByID", ctx, itemID).Return(item, nil).Once()
        mockWarehouseRepo.On("GetByID", ctx, warehouseID).Return(warehouse, nil).Once()
        // Mock GetStockLevel for the check (assuming it's called, current service logic logs but allows)
        mockTxnRepo.On("GetStockLevel", ctx, itemID, warehouseID, mock.AnythingOfType("time.Time")).Return(5.0, nil).Once() // Current stock is 5
        mockTxnRepo.On("Create", ctx, mock.AnythingOfType("*models.InventoryTransaction")).Return(&models.InventoryTransaction{ID: uuid.New(), ItemID: itemID, WarehouseID: warehouseID, Quantity: 10.0, TransactionType: models.AdjustStockOut}, nil).Once()

        txn, err := invService.CreateInventoryAdjustment(ctx, reqOut)
        assert.NoError(t, err) // Service currently allows this with a log. If it blocked, this would be an error.
        assert.NotNil(t, txn)
        assert.Equal(t, models.AdjustStockOut, txn.TransactionType)
        mockItemRepo.AssertExpectations(t)
        mockWarehouseRepo.AssertExpectations(t)
        mockTxnRepo.AssertExpectations(t)
    })


    t.Run("Error - Invalid Adjustment Type", func(t *testing.T) {
        reqInvalidType := req
        reqInvalidType.AdjustmentType = "INVALID_ADJUST_TYPE"
        mockItemRepo.On("GetByID", ctx, itemID).Return(item, nil).Once()
        mockWarehouseRepo.On("GetByID", ctx, warehouseID).Return(warehouse, nil).Once()

        _, err := invService.CreateInventoryAdjustment(ctx, reqInvalidType)
        assert.Error(t, err)
        assert.IsType(t, &app_errors.ValidationError{}, err)
        assert.Contains(t, err.Error(), "invalid adjustment type")
    })

    t.Run("Error - Item Not Active", func(t *testing.T) {
        inactiveItem := &models.Item{ID: itemID, SKU: "ADJITEM_INACT", IsActive: false, ItemType: models.FinishedGood}
        mockItemRepo.On("GetByID", ctx, itemID).Return(inactiveItem, nil).Once()
        // mockWarehouseRepo.On("GetByID", ctx, warehouseID).Return(warehouse, nil) // Might not be called

        _, err := invService.CreateInventoryAdjustment(ctx, req)
        assert.Error(t, err)
        assert.IsType(t, &app_errors.ValidationError{}, err)
        assert.Contains(t, err.Error(), "is not active")
    })

    t.Run("Error - Item is NonInventory type", func(t *testing.T) {
        nonInvItem := &models.Item{ID: itemID, SKU: "ADJITEM_NONINV", IsActive: true, ItemType: models.NonInventory}
        mockItemRepo.On("GetByID", ctx, itemID).Return(nonInvItem, nil).Once()
        // mockWarehouseRepo.On("GetByID", ctx, warehouseID).Return(warehouse, nil) // Might not be called

        _, err := invService.CreateInventoryAdjustment(ctx, req)
        assert.Error(t, err)
        assert.IsType(t, &app_errors.ValidationError{}, err)
        assert.Contains(t, err.Error(), "non-inventory item and cannot have stock adjustments")
    })
}

func TestInventoryService_GetInventoryLevels(t *testing.T) {
    mockItemRepo := invRepoMock.NewItemRepositoryMock(t)
    mockWarehouseRepo := invRepoMock.NewWarehouseRepositoryMock(t)
    mockTxnRepo := invRepoMock.NewInventoryTransactionRepositoryMock(t)
    invService := service.NewInventoryService(mockItemRepo, mockWarehouseRepo, mockTxnRepo)
    ctx := context.Background()

    itemID := uuid.New()
    warehouseID := uuid.New()
    item := &models.Item{ID: itemID, SKU: "LVLITEM", Name: "Level Item"}
    warehouse := &models.Warehouse{ID: warehouseID, Code: "LVLWH", Name: "Level Warehouse"}
    now := time.Now()

    t.Run("Success - Specific Item in Specific Warehouse", func(t *testing.T) {
        req := dto.InventoryLevelRequest{ItemID: &itemID, WarehouseID: &warehouseID, AsOfDate: &now}
        mockItemRepo.On("GetByID", ctx, itemID).Return(item, nil).Once()
        mockWarehouseRepo.On("GetByID", ctx, warehouseID).Return(warehouse, nil).Once()
        mockTxnRepo.On("GetStockLevel", ctx, itemID, warehouseID, now).Return(25.5, nil).Once()

        resp, err := invService.GetInventoryLevels(ctx, req)
        assert.NoError(t, err)
        assert.NotNil(t, resp)
        assert.Len(t, resp.Levels, 1)
        assert.Equal(t, 25.5, resp.Levels[0].Quantity)
        assert.Equal(t, itemID, resp.Levels[0].ItemID)
        assert.Equal(t, warehouseID, resp.Levels[0].WarehouseID)
        mockItemRepo.AssertExpectations(t)
        mockWarehouseRepo.AssertExpectations(t)
        mockTxnRepo.AssertExpectations(t)
    })

    t.Run("Error - GetInventoryLevels without ItemID or WarehouseID", func(t *testing.T) {
        req := dto.InventoryLevelRequest{AsOfDate: &now} // No filters
        _, err := invService.GetInventoryLevels(ctx, req)
        assert.Error(t, err)
        assert.IsType(t, &app_errors.ValidationError{}, err)
        assert.Contains(t, err.Error(), "not supported without item or warehouse filter")
    })
}

// Add more tests for other service methods:
// GetItemByID, GetItemBySKU, ListItems
// GetWarehouseByID, GetWarehouseByCode, UpdateWarehouse, DeleteWarehouse, ListWarehouses
// GetItemStockLevelInWarehouse
// etc.
// Remember to test error paths and edge cases.
