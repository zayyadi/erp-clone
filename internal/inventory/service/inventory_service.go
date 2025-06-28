package service

import (
	"context"
	"erp-system/internal/inventory/models"
	repo "erp-system/internal/inventory/repository" // Alias to avoid conflict
	dto "erp-system/internal/inventory/service/dto"
	app_errors "erp-system/pkg/errors"
	"erp-system/pkg/logger"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// InventoryService defines the interface for inventory-related business logic.
type InventoryService interface {
	// Item Management
	CreateItem(ctx context.Context, req dto.CreateItemRequest) (*models.Item, error)
	GetItemByID(ctx context.Context, id uuid.UUID) (*models.Item, error)
	GetItemBySKU(ctx context.Context, sku string) (*models.Item, error)
	UpdateItem(ctx context.Context, id uuid.UUID, req dto.UpdateItemRequest) (*models.Item, error)
	DeleteItem(ctx context.Context, id uuid.UUID) error
	ListItems(ctx context.Context, req dto.ListItemRequest) ([]*models.Item, int64, error)

	// Warehouse Management
	CreateWarehouse(ctx context.Context, req dto.CreateWarehouseRequest) (*models.Warehouse, error)
	GetWarehouseByID(ctx context.Context, id uuid.UUID) (*models.Warehouse, error)
	GetWarehouseByCode(ctx context.Context, code string) (*models.Warehouse, error)
	UpdateWarehouse(ctx context.Context, id uuid.UUID, req dto.UpdateWarehouseRequest) (*models.Warehouse, error)
	DeleteWarehouse(ctx context.Context, id uuid.UUID) error
	ListWarehouses(ctx context.Context, req dto.ListWarehouseRequest) ([]*models.Warehouse, int64, error)

	// Inventory Transactions & Levels
	CreateInventoryAdjustment(ctx context.Context, req dto.CreateInventoryAdjustmentRequest) (*models.InventoryTransaction, error)
	GetInventoryLevels(ctx context.Context, req dto.InventoryLevelRequest) (*dto.InventoryLevelsResponse, error)
	GetItemStockLevelInWarehouse(ctx context.Context, itemID uuid.UUID, warehouseID uuid.UUID, asOfDate time.Time) (float64, error)

	// Reporting (Placeholders for now)
	// GenerateInventoryValuationReport(ctx context.Context, req dto.InventoryValuationReportRequest) (*dto.InventoryValuationReportResponse, error)
}

// inventoryService is an implementation of InventoryService.
type inventoryService struct {
	itemRepo         repo.ItemRepository
	warehouseRepo    repo.WarehouseRepository
	transactionRepo  repo.InventoryTransactionRepository
}

// NewInventoryService creates a new InventoryService.
func NewInventoryService(
	itemRepo repo.ItemRepository,
	warehouseRepo repo.WarehouseRepository,
	transactionRepo repo.InventoryTransactionRepository,
) InventoryService {
	return &inventoryService{
		itemRepo:        itemRepo,
		warehouseRepo:   warehouseRepo,
		transactionRepo: transactionRepo,
	}
}

// --- Item Management Methods ---

func (s *inventoryService) CreateItem(ctx context.Context, req dto.CreateItemRequest) (*models.Item, error) {
	logger.InfoLogger.Printf("Service: Attempting to create item with SKU: %s", req.SKU)

	// Validate ItemType
	validItemType := false
	for _, it := range []models.ItemType{models.RawMaterial, models.FinishedGood, models.WorkInProgress, models.NonInventory} {
		if req.ItemType == it {
			validItemType = true
			break
		}
	}
	if !validItemType {
		return nil, app_errors.NewValidationError(fmt.Sprintf("invalid item type: %s", req.ItemType), "item_type")
	}

	// Check if SKU already exists
	existing, err := s.itemRepo.GetBySKU(ctx, req.SKU)
	if err != nil && !isNotFoundError(err) {
		return nil, err // Propagate internal server error
	}
	if existing != nil {
		return nil, app_errors.NewConflictError(fmt.Sprintf("item with SKU %s already exists", req.SKU))
	}

	item := &models.Item{
		SKU:           req.SKU,
		Name:          req.Name,
		Description:   req.Description,
		UnitOfMeasure: req.UnitOfMeasure,
		ItemType:      req.ItemType,
		IsActive:      req.IsActive, // DTO default is fine, GORM model default handles it if not set
	}
    if !req.IsActive && req.SKU != "" { // If explicitly set to inactive on create
        // This check might be redundant if DTO has default true and user doesn't send it
        // but good if user can explicitly create as inactive.
        // Default handling: if is_active is not in JSON, it's false for bool.
        // We should ensure DTO default or explicit value.
        // The model has `gorm:"default:true"`, so if not set, it becomes true.
        // If req.IsActive is a non-pointer bool, it will be false if not sent in JSON.
        // Let's assume the DTO's bool `IsActive` will be `false` if not provided by client.
        // So, if client wants it active, they must send `is_active: true`.
        // Or, service can default it:
        // item.IsActive = true // Default to true
        // if req.IsActive explicitly set to false by client, then:
        // if json request had "is_active": false, then req.IsActive would be false.
        // This is fine. The model's default:true handles if the field is entirely absent at DB level.
        // For GORM create, if item.IsActive is false (its zero value), GORM might skip it if not careful,
        // relying on DB default. But here, we are setting it from req.IsActive.
    }


	createdItem, err := s.itemRepo.Create(ctx, item)
	if err != nil {
		return nil, err
	}
	return createdItem, nil
}

func (s *inventoryService) GetItemByID(ctx context.Context, id uuid.UUID) (*models.Item, error) {
	return s.itemRepo.GetByID(ctx, id)
}

func (s *inventoryService) GetItemBySKU(ctx context.Context, sku string) (*models.Item, error) {
	return s.itemRepo.GetBySKU(ctx, sku)
}

func (s *inventoryService) UpdateItem(ctx context.Context, id uuid.UUID, req dto.UpdateItemRequest) (*models.Item, error) {
	item, err := s.itemRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		item.Name = *req.Name
	}
	if req.Description != nil {
		item.Description = *req.Description
	}
	if req.UnitOfMeasure != nil {
		item.UnitOfMeasure = *req.UnitOfMeasure
	}
	if req.ItemType != nil {
		validItemType := false
		for _, it := range []models.ItemType{models.RawMaterial, models.FinishedGood, models.WorkInProgress, models.NonInventory} {
			if *req.ItemType == it {
				validItemType = true
				break
			}
		}
		if !validItemType {
			return nil, app_errors.NewValidationError(fmt.Sprintf("invalid item type: %s", *req.ItemType), "item_type")
		}
		item.ItemType = *req.ItemType
	}
	if req.IsActive != nil {
		item.IsActive = *req.IsActive
		// Add logic here if deactivating an item has implications (e.g., stock exists)
		if !(*req.IsActive) {
            // Check if item has stock. If so, prevent deactivation or warn.
            stockLevels, err := s.transactionRepo.GetStockLevelsByItem(ctx, id, time.Now())
            if err != nil {
                logger.ErrorLogger.Printf("Service: Failed to check stock levels for item %s during deactivation: %v", id, err)
                // Decide: proceed with deactivation or return error? For now, let's allow it but log.
            } else {
                for _, level := range stockLevels {
                    if level > 0 { // Using a small tolerance for float might be better: math.Abs(level) > 1e-9
                        logger.WarnLogger.Printf("Service: Item %s (ID: %s) is being deactivated but has stock in some warehouses.", item.SKU, item.ID)
                        // return nil, app_errors.NewConflictError(fmt.Sprintf("cannot deactivate item %s, it has existing stock", item.SKU))
                        break
                    }
                }
            }
        }
	}

	return s.itemRepo.Update(ctx, item)
}

func (s *inventoryService) DeleteItem(ctx context.Context, id uuid.UUID) error {
	item, err := s.itemRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Business logic: Cannot delete item if it has stock or transactions.
	// Check stock levels across all warehouses for this item.
	stockLevels, err := s.transactionRepo.GetStockLevelsByItem(ctx, id, time.Now())
	if err != nil {
		logger.ErrorLogger.Printf("Service: Failed to check stock levels for item %s before deletion: %v", id, err)
		return app_errors.NewInternalServerError("failed to verify item stock before deletion", err)
	}
	for whID, level := range stockLevels {
		if level != 0 { // Using a small tolerance for float might be better: math.Abs(level) > 1e-9
			logger.WarnLogger.Printf("Service: Cannot delete item %s, it has stock (%.3f) in warehouse %s.", item.SKU, level, whID)
			return app_errors.NewConflictError(fmt.Sprintf("cannot delete item %s, it has existing stock", item.SKU))
		}
	}
	// A more thorough check would be to see if ANY transactions exist, even if current stock is zero.
	// For now, checking current stock is a basic safeguard.

	return s.itemRepo.Delete(ctx, id)
}

func (s *inventoryService) ListItems(ctx context.Context, req dto.ListItemRequest) ([]*models.Item, int64, error) {
	filters := make(map[string]interface{})
	if req.Name != "" {
		filters["name"] = req.Name
	}
	if req.SKU != "" {
		filters["sku"] = req.SKU
	}
	if req.ItemType != "" {
		filters["item_type"] = req.ItemType
	}
	if req.IsActive != nil { // Pointer check
		filters["is_active"] = *req.IsActive
	}

	offset := 0
	if req.Page > 0 && req.Limit > 0 {
		offset = (req.Page - 1) * req.Limit
	}
	limit := req.Limit
	if limit == 0 { limit = 20 }


	return s.itemRepo.List(ctx, offset, limit, filters)
}


// --- Warehouse Management Methods ---

func (s *inventoryService) CreateWarehouse(ctx context.Context, req dto.CreateWarehouseRequest) (*models.Warehouse, error) {
	existing, err := s.warehouseRepo.GetByCode(ctx, req.Code)
	if err != nil && !isNotFoundError(err) {
		return nil, err
	}
	if existing != nil {
		return nil, app_errors.NewConflictError(fmt.Sprintf("warehouse with code %s already exists", req.Code))
	}

	warehouse := &models.Warehouse{
		Code:     req.Code,
		Name:     req.Name,
		Location: req.Location,
		IsActive: req.IsActive, // Similar default handling as Item.IsActive
	}
	return s.warehouseRepo.Create(ctx, warehouse)
}

func (s *inventoryService) GetWarehouseByID(ctx context.Context, id uuid.UUID) (*models.Warehouse, error) {
	return s.warehouseRepo.GetByID(ctx, id)
}

func (s *inventoryService) GetWarehouseByCode(ctx context.Context, code string) (*models.Warehouse, error) {
	return s.warehouseRepo.GetByCode(ctx, code)
}

func (s *inventoryService) UpdateWarehouse(ctx context.Context, id uuid.UUID, req dto.UpdateWarehouseRequest) (*models.Warehouse, error) {
	warehouse, err := s.warehouseRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		warehouse.Name = *req.Name
	}
	if req.Location != nil {
		warehouse.Location = *req.Location
	}
	if req.IsActive != nil {
		warehouse.IsActive = *req.IsActive
		if !(*req.IsActive) {
            // Check if warehouse has stock.
            stockLevels, err := s.transactionRepo.GetStockLevelsByWarehouse(ctx, id, time.Now())
             if err != nil {
                logger.ErrorLogger.Printf("Service: Failed to check stock levels for warehouse %s during deactivation: %v", id, err)
            } else {
                for _, level := range stockLevels {
                    if level != 0 {
                         logger.WarnLogger.Printf("Service: Warehouse %s (ID: %s) is being deactivated but has stock.", warehouse.Code, warehouse.ID)
                        // return nil, app_errors.NewConflictError(fmt.Sprintf("cannot deactivate warehouse %s, it has items in stock", warehouse.Code))
                        break
                    }
                }
            }
        }
	}
	return s.warehouseRepo.Update(ctx, warehouse)
}

func (s *inventoryService) DeleteWarehouse(ctx context.Context, id uuid.UUID) error {
	warehouse, err := s.warehouseRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
    // Check if warehouse has stock.
    stockLevels, err := s.transactionRepo.GetStockLevelsByWarehouse(ctx, id, time.Now())
    if err != nil {
        logger.ErrorLogger.Printf("Service: Failed to check stock levels for warehouse %s during deletion: %v", id, err)
        return app_errors.NewInternalServerError("failed to verify warehouse stock before deletion", err)
    }
    for itemID, level := range stockLevels {
        if level != 0 {
            logger.WarnLogger.Printf("Service: Cannot delete warehouse %s, it has stock (item %s: %.3f).", warehouse.Code, itemID, level)
            return app_errors.NewConflictError(fmt.Sprintf("cannot delete warehouse %s, it has items in stock", warehouse.Code))
        }
    }
	return s.warehouseRepo.Delete(ctx, id)
}

func (s *inventoryService) ListWarehouses(ctx context.Context, req dto.ListWarehouseRequest) ([]*models.Warehouse, int64, error) {
	filters := make(map[string]interface{})
	if req.Name != "" { filters["name"] = req.Name }
	if req.Code != "" { filters["code"] = req.Code }
	if req.Location != "" { filters["location"] = req.Location }
	if req.IsActive != nil { filters["is_active"] = *req.IsActive }

	offset := 0
	if req.Page > 0 && req.Limit > 0 { offset = (req.Page - 1) * req.Limit }
	limit := req.Limit
	if limit == 0 { limit = 20 }

	return s.warehouseRepo.List(ctx, offset, limit, filters)
}


// --- Inventory Transactions & Levels Methods ---

func (s *inventoryService) CreateInventoryAdjustment(ctx context.Context, req dto.CreateInventoryAdjustmentRequest) (*models.InventoryTransaction, error) {
	logger.InfoLogger.Printf("Service: Creating inventory adjustment for item %s in warehouse %s", req.ItemID, req.WarehouseID)

	// Validate Item
	item, err := s.itemRepo.GetByID(ctx, req.ItemID)
	if err != nil {
		if isNotFoundError(err) {
			return nil, app_errors.NewValidationError("item not found", "item_id")
		}
		return nil, err
	}
	if !item.IsActive {
		return nil, app_errors.NewValidationError(fmt.Sprintf("item %s is not active", item.SKU), "item_id")
	}
    if item.ItemType == models.NonInventory {
        return nil, app_errors.NewValidationError(fmt.Sprintf("item %s is a non-inventory item and cannot have stock adjustments", item.SKU), "item_id")
    }


	// Validate Warehouse
	warehouse, err := s.warehouseRepo.GetByID(ctx, req.WarehouseID)
	if err != nil {
		if isNotFoundError(err) {
			return nil, app_errors.NewValidationError("warehouse not found", "warehouse_id")
		}
		return nil, err
	}
	if !warehouse.IsActive {
		return nil, app_errors.NewValidationError(fmt.Sprintf("warehouse %s is not active", warehouse.Code), "warehouse_id")
	}

	// Validate AdjustmentType
	switch req.AdjustmentType {
	case models.AdjustStockIn, models.AdjustStockOut:
		// Valid adjustment types
	default:
		return nil, app_errors.NewValidationError(fmt.Sprintf("invalid adjustment type: %s. Must be ADJUST_STOCK_IN or ADJUST_STOCK_OUT.", req.AdjustmentType), "adjustment_type")
	}

	// For AdjustStockOut, check if sufficient stock exists (optional, depending on allow negative stock setting)
	// This check is simplified; a real system might have an "allow negative inventory" setting per item/warehouse.
	if req.AdjustmentType == models.AdjustStockOut {
		currentStock, err := s.transactionRepo.GetStockLevel(ctx, req.ItemID, req.WarehouseID, time.Now())
		if err != nil {
			logger.ErrorLogger.Printf("Service: Failed to get current stock for item %s, warehouse %s: %v", req.ItemID, req.WarehouseID, err)
			return nil, app_errors.NewInternalServerError("failed to verify stock before adjustment", err)
		}
		if currentStock < req.Quantity {
			// For now, let's allow it and log. A real system might block this.
			logger.WarnLogger.Printf("Service: AdjustStockOut for item %s (qty %.3f) in warehouse %s exceeds current stock (%.3f). Negative stock will result.", item.SKU, req.Quantity, warehouse.Code, currentStock)
			// return nil, app_errors.NewConflictError(fmt.Sprintf("insufficient stock for item %s in warehouse %s. Available: %.3f, Requested: %.3f", item.SKU, warehouse.Code, currentStock, req.Quantity))
		}
	}

	transactionDate := time.Now()
	if req.TransactionDate != nil && !(*req.TransactionDate).IsZero() {
		transactionDate = *req.TransactionDate
	}

	transaction := &models.InventoryTransaction{
		ItemID:          req.ItemID,
		WarehouseID:     req.WarehouseID,
		Quantity:        req.Quantity, // Quantity is always positive; type defines effect
		TransactionType: req.AdjustmentType,
		TransactionDate: transactionDate,
		Notes:           req.Notes,
		ReferenceID:     req.ReferenceID,
	}

	return s.transactionRepo.Create(ctx, transaction)
}

func (s *inventoryService) GetInventoryLevels(ctx context.Context, req dto.InventoryLevelRequest) (*dto.InventoryLevelsResponse, error) {
	asOfDate := time.Now()
	if req.AsOfDate != nil && !(*req.AsOfDate).IsZero() {
		asOfDate = *req.AsOfDate
	}

	response := &dto.InventoryLevelsResponse{
		Levels:   make([]dto.ItemStockLevelInfo, 0),
		AsOfDate: asOfDate,
	}

	if req.ItemID != nil && req.WarehouseID != nil { // Specific item in specific warehouse
		item, err := s.itemRepo.GetByID(ctx, *req.ItemID)
		if err != nil { return nil, err }
		warehouse, err := s.warehouseRepo.GetByID(ctx, *req.WarehouseID)
		if err != nil { return nil, err }

		quantity, err := s.transactionRepo.GetStockLevel(ctx, *req.ItemID, *req.WarehouseID, asOfDate)
		if err != nil { return nil, err }
		response.Levels = append(response.Levels, dto.ItemStockLevelInfo{
			ItemID: item.ID, ItemSKU: item.SKU, ItemName: item.Name,
			WarehouseID: warehouse.ID, WarehouseCode: warehouse.Code, WarehouseName: warehouse.Name,
			Quantity: quantity, AsOfDate: asOfDate,
		})
	} else if req.ItemID != nil { // Specific item across all warehouses
		item, err := s.itemRepo.GetByID(ctx, *req.ItemID)
		if err != nil { return nil, err }

		stockPerWarehouse, err := s.transactionRepo.GetStockLevelsByItem(ctx, *req.ItemID, asOfDate)
		if err != nil { return nil, err }

		for whID, qty := range stockPerWarehouse {
			warehouse, err := s.warehouseRepo.GetByID(ctx, whID) // TODO: Optimize, maybe batch fetch warehouses
			if err != nil { logger.ErrorLogger.Printf("Error fetching warehouse %s: %v", whID, err); continue }
			response.Levels = append(response.Levels, dto.ItemStockLevelInfo{
				ItemID: item.ID, ItemSKU: item.SKU, ItemName: item.Name,
				WarehouseID: warehouse.ID, WarehouseCode: warehouse.Code, WarehouseName: warehouse.Name,
				Quantity: qty, AsOfDate: asOfDate,
			})
		}
	} else if req.WarehouseID != nil { // All items in a specific warehouse
		warehouse, err := s.warehouseRepo.GetByID(ctx, *req.WarehouseID)
		if err != nil { return nil, err }

		stockPerItem, err := s.transactionRepo.GetStockLevelsByWarehouse(ctx, *req.WarehouseID, asOfDate)
		if err != nil { return nil, err }

		for itemID, qty := range stockPerItem {
			item, err := s.itemRepo.GetByID(ctx, itemID) // TODO: Optimize, maybe batch fetch items
			if err != nil { logger.ErrorLogger.Printf("Error fetching item %s: %v", itemID, err); continue }
			response.Levels = append(response.Levels, dto.ItemStockLevelInfo{
				ItemID: item.ID, ItemSKU: item.SKU, ItemName: item.Name,
				WarehouseID: warehouse.ID, WarehouseCode: warehouse.Code, WarehouseName: warehouse.Name,
				Quantity: qty, AsOfDate: asOfDate,
			})
		}
	} else { // All items across all warehouses (potentially large, needs pagination or summary)
		// This could be very expensive. For now, let's return an error or suggest more specific filters.
		logger.WarnLogger.Println("Service: GetInventoryLevels called without ItemID or WarehouseID. This query can be very large and is not fully implemented for optimal performance.")
        // Fetch all active items and for each item, fetch its stock across warehouses. This is N+1 like.
        // A better approach would be a single complex query if possible, or paginated results.
        // For simplicity, this case is not fully optimized here.
        // Consider fetching all transactions up to date and aggregating in memory (bad for large datasets)
        // Or iterate through all items, then all warehouses (very bad).
        // Placeholder:
        return nil, app_errors.NewValidationError("Listing all stock levels is not supported without item or warehouse filter in this version. Please provide filters.", "")
	}

	return response, nil
}

func (s *inventoryService) GetItemStockLevelInWarehouse(ctx context.Context, itemID uuid.UUID, warehouseID uuid.UUID, asOfDate time.Time) (float64, error) {
    if asOfDate.IsZero() {
        asOfDate = time.Now()
    }
    // Validate item and warehouse existence first (optional, repo method might handle it)
    _, err := s.itemRepo.GetByID(ctx, itemID)
    if err != nil { return 0, app_errors.NewValidationError("item not found", "item_id")}
    _, err = s.warehouseRepo.GetByID(ctx, warehouseID)
    if err != nil { return 0, app_errors.NewValidationError("warehouse not found", "warehouse_id")}

    return s.transactionRepo.GetStockLevel(ctx, itemID, warehouseID, asOfDate)
}


// --- Helper Functions ---
// (isNotFoundError might be duplicated from accounting service, consider moving to a shared pkg/utils if common)
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*app_errors.NotFoundError)
	return ok
}
