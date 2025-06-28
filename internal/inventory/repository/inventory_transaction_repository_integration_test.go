package repository_test

import (
	"context"
	"erp-system/internal/inventory/models"
	"erp-system/internal/inventory/repository"
	"testing"
	"time"

	"github.com/google/uuid"
	// "github.com/stretchr/testify/assert" // REMOVED - suite provides assertions
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// InventoryTransactionRepositoryIntegrationTestSuite defines the suite for InventoryTransactionRepository integration tests.
type InventoryTransactionRepositoryIntegrationTestSuite struct {
	suite.Suite
	db   *gorm.DB
	repo repository.InventoryTransactionRepository

	// Repos for setting up dependent data
	itemRepo      repository.ItemRepository
	warehouseRepo repository.WarehouseRepository

	ctx context.Context

	// Test data
	testItem1      *models.Item
	testItem2      *models.Item
	testWarehouse1 *models.Warehouse
	testWarehouse2 *models.Warehouse
}

// SetupSuite runs once before all tests in the suite.
func (s *InventoryTransactionRepositoryIntegrationTestSuite) SetupSuite() {
	s.T().Log("Setting up suite for InventoryTransactionRepository integration tests...")
	s.db = dbInstance_inventory // Use the DB instance from item_integration_test_setup_test.go
	s.repo = repository.NewInventoryTransactionRepository(s.db)
	s.itemRepo = repository.NewItemRepository(s.db)
	s.warehouseRepo = repository.NewWarehouseRepository(s.db)
	s.ctx = context.Background()
	s.T().Log("InventoryTransactionRepository Suite setup complete.")
}

// SetupTest runs before each test in the suite.
func (s *InventoryTransactionRepositoryIntegrationTestSuite) SetupTest() {
	s.T().Logf("Setting up test: %s", s.T().Name())
	resetInventoryTables(s.T(), s.db) // Reset relevant tables

	// Create common items and warehouses
	item1 := models.Item{SKU: "TXN_ITEM_1", Name: "Transacting Item 1", UnitOfMeasure: "PCS", ItemType: models.FinishedGood, IsActive: true}
	item2 := models.Item{SKU: "TXN_ITEM_2", Name: "Transacting Item 2", UnitOfMeasure: "PCS", ItemType: models.RawMaterial, IsActive: true}
	s.testItem1, _ = s.itemRepo.Create(s.ctx, &item1)
	s.testItem2, _ = s.itemRepo.Create(s.ctx, &item2)

	wh1 := models.Warehouse{Code: "TXN_WH_1", Name: "Transacting WH 1", IsActive: true}
	wh2 := models.Warehouse{Code: "TXN_WH_2", Name: "Transacting WH 2", IsActive: true}
	s.testWarehouse1, _ = s.warehouseRepo.Create(s.ctx, &wh1)
	s.testWarehouse2, _ = s.warehouseRepo.Create(s.ctx, &wh2)

	s.Require().NotNil(s.testItem1, "Test item 1 setup failed")
	s.Require().NotNil(s.testWarehouse1, "Test warehouse 1 setup failed")

	s.T().Logf("Test setup complete for: %s", s.T().Name())
}

// TestInventoryTransactionRepositoryIntegration runs the entire suite.
func TestInventoryTransactionRepositoryIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping InventoryTransactionRepository integration tests in short mode.")
	}
	t.Log("Starting InventoryTransactionRepositoryIntegration Test Suite...")
	suite.Run(t, new(InventoryTransactionRepositoryIntegrationTestSuite))
	t.Log("InventoryTransactionRepositoryIntegration Test Suite finished.")
}

func (s *InventoryTransactionRepositoryIntegrationTestSuite) TestCreateInventoryTransaction() {
	s.T().Log("Running TestCreateInventoryTransaction")
	transaction := models.InventoryTransaction{
		ItemID:          s.testItem1.ID,
		WarehouseID:     s.testWarehouse1.ID,
		Quantity:        10.0,
		TransactionType: models.ReceiveStock,
		TransactionDate: time.Now(),
	}

	createdTxn, err := s.repo.Create(s.ctx, &transaction)
	s.NoError(err)
	s.NotNil(createdTxn)
	s.NotEqual(uuid.Nil, createdTxn.ID)
	s.Equal(s.testItem1.ID, createdTxn.ItemID)
	s.NotNil(createdTxn.Item, "Created transaction should have Item preloaded")
	s.Equal(s.testItem1.SKU, createdTxn.Item.SKU)

	var fetchedTxn models.InventoryTransaction
	err = s.db.First(&fetchedTxn, "id = ?", createdTxn.ID).Error
	s.NoError(err)
	s.Equal(createdTxn.Quantity, fetchedTxn.Quantity)
}

func (s *InventoryTransactionRepositoryIntegrationTestSuite) TestGetInventoryTransactionByID() {
	s.T().Log("Running TestGetInventoryTransactionByID")
	seedTxn := models.InventoryTransaction{
		ItemID: s.testItem1.ID, WarehouseID: s.testWarehouse1.ID, Quantity: 5, TransactionType: models.AdjustStockIn, TransactionDate: time.Now(),
	}
	s.db.Create(&seedTxn) // Seed directly

	fetchedTxn, err := s.repo.GetByID(s.ctx, seedTxn.ID)
	s.NoError(err)
	s.NotNil(fetchedTxn)
	s.Equal(seedTxn.ID, fetchedTxn.ID)
	s.Equal(s.testItem1.ID, fetchedTxn.ItemID)
	s.NotNil(fetchedTxn.Item, "Fetched transaction by ID should have Item preloaded")
	s.NotNil(fetchedTxn.Warehouse, "Fetched transaction by ID should have Warehouse preloaded")


	_, err = s.repo.GetByID(s.ctx, uuid.New()) // Non-existent ID
	s.Error(err)
}

func (s *InventoryTransactionRepositoryIntegrationTestSuite) TestListInventoryTransactions() {
	s.T().Log("Running TestListInventoryTransactions")
	now := time.Now()
	s.repo.Create(s.ctx, &models.InventoryTransaction{ItemID: s.testItem1.ID, WarehouseID: s.testWarehouse1.ID, Quantity: 10, TransactionType: models.ReceiveStock, TransactionDate: now.Add(-time.Hour)})
	s.repo.Create(s.ctx, &models.InventoryTransaction{ItemID: s.testItem1.ID, WarehouseID: s.testWarehouse1.ID, Quantity: 2, TransactionType: models.IssueStock, TransactionDate: now})
	s.repo.Create(s.ctx, &models.InventoryTransaction{ItemID: s.testItem2.ID, WarehouseID: s.testWarehouse1.ID, Quantity: 20, TransactionType: models.ReceiveStock, TransactionDate: now.Add(-30 * time.Minute)})

	s.Run("No filters", func() {
		txns, total, err := s.repo.List(s.ctx, 0, 10, make(map[string]interface{}))
		s.NoError(err)
		s.Len(txns, 3)
		s.Equal(int64(3), total)
		s.NotNil(txns[0].Item, "Listed transactions should have Item preloaded")
	})

	s.Run("Filter by ItemID", func() {
		filters := map[string]interface{}{"item_id": s.testItem1.ID}
		txns, total, err := s.repo.List(s.ctx, 0, 10, filters)
		s.NoError(err)
		s.Len(txns, 2)
		s.Equal(int64(2), total)
	})

	s.Run("Filter by TransactionType", func() {
		filters := map[string]interface{}{"transaction_type": models.IssueStock}
		txns, total, err := s.repo.List(s.ctx, 0, 10, filters)
		s.NoError(err)
		s.Len(txns, 1)
		s.Equal(int64(1), total)
		s.Equal(models.IssueStock, txns[0].TransactionType)
	})
}

func (s *InventoryTransactionRepositoryIntegrationTestSuite) TestGetStockLevel() {
	s.T().Log("Running TestGetStockLevel")
	now := time.Now()
	// Item 1 in WH1: +10, -3, +5 = 12
	s.repo.Create(s.ctx, &models.InventoryTransaction{ItemID: s.testItem1.ID, WarehouseID: s.testWarehouse1.ID, Quantity: 10, TransactionType: models.ReceiveStock, TransactionDate: now.Add(-2 * time.Hour)})
	s.repo.Create(s.ctx, &models.InventoryTransaction{ItemID: s.testItem1.ID, WarehouseID: s.testWarehouse1.ID, Quantity: 3, TransactionType: models.IssueStock, TransactionDate: now.Add(-time.Hour)})
	s.repo.Create(s.ctx, &models.InventoryTransaction{ItemID: s.testItem1.ID, WarehouseID: s.testWarehouse1.ID, Quantity: 5, TransactionType: models.AdjustStockIn, TransactionDate: now})
	// Item 1 in WH2: +7
	s.repo.Create(s.ctx, &models.InventoryTransaction{ItemID: s.testItem1.ID, WarehouseID: s.testWarehouse2.ID, Quantity: 7, TransactionType: models.ReceiveStock, TransactionDate: now})
	// Item 2 in WH1: +100 (after 'now', should be excluded by date filter if date is 'now')
	s.repo.Create(s.ctx, &models.InventoryTransaction{ItemID: s.testItem2.ID, WarehouseID: s.testWarehouse1.ID, Quantity: 100, TransactionType: models.ReceiveStock, TransactionDate: now.Add(time.Hour)})


	s.Run("Item1 in WH1 as of now", func() {
		level, err := s.repo.GetStockLevel(s.ctx, s.testItem1.ID, s.testWarehouse1.ID, now)
		s.NoError(err)
		s.Equal(12.0, level)
	})

	s.Run("Item1 in WH1 as of one hour ago", func() {
		// Should only include the first transaction (+10)
		level, err := s.repo.GetStockLevel(s.ctx, s.testItem1.ID, s.testWarehouse1.ID, now.Add(-time.Hour).Add(time.Minute)) // ensure it includes the -3 txn
		s.NoError(err)
		s.Equal(7.0, level) // 10 - 3 = 7
	})

	s.Run("Item1 in WH2 as of now", func() {
		level, err := s.repo.GetStockLevel(s.ctx, s.testItem1.ID, s.testWarehouse2.ID, now)
		s.NoError(err)
		s.Equal(7.0, level)
	})

	s.Run("Item2 in WH1 as of now (should be 0)", func() {
		level, err := s.repo.GetStockLevel(s.ctx, s.testItem2.ID, s.testWarehouse1.ID, now)
		s.NoError(err)
		s.Equal(0.0, level) // The +100 is in the future
	})

	s.Run("Item2 in WH1 as of two hours from now (should be 100)", func() {
		level, err := s.repo.GetStockLevel(s.ctx, s.testItem2.ID, s.testWarehouse1.ID, now.Add(2*time.Hour))
		s.NoError(err)
		s.Equal(100.0, level)
	})

	s.Run("Non-existent item or warehouse", func() {
		level, err := s.repo.GetStockLevel(s.ctx, uuid.New(), s.testWarehouse1.ID, now)
		s.NoError(err) // Repo doesn't error if item/wh don't exist, just returns 0 transactions
		s.Equal(0.0, level)
	})
}


func (s *InventoryTransactionRepositoryIntegrationTestSuite) TestGetStockLevelsByItem() {
    s.T().Log("Running TestGetStockLevelsByItem")
    now := time.Now()
    // Item 1: WH1 has 12, WH2 has 7
    s.repo.Create(s.ctx, &models.InventoryTransaction{ItemID: s.testItem1.ID, WarehouseID: s.testWarehouse1.ID, Quantity: 10, TransactionType: models.ReceiveStock, TransactionDate: now.Add(-2 * time.Hour)})
	s.repo.Create(s.ctx, &models.InventoryTransaction{ItemID: s.testItem1.ID, WarehouseID: s.testWarehouse1.ID, Quantity: 3, TransactionType: models.IssueStock, TransactionDate: now.Add(-time.Hour)})
	s.repo.Create(s.ctx, &models.InventoryTransaction{ItemID: s.testItem1.ID, WarehouseID: s.testWarehouse1.ID, Quantity: 5, TransactionType: models.AdjustStockIn, TransactionDate: now})
	s.repo.Create(s.ctx, &models.InventoryTransaction{ItemID: s.testItem1.ID, WarehouseID: s.testWarehouse2.ID, Quantity: 7, TransactionType: models.ReceiveStock, TransactionDate: now})
    // Item 2: WH1 has 50
    s.repo.Create(s.ctx, &models.InventoryTransaction{ItemID: s.testItem2.ID, WarehouseID: s.testWarehouse1.ID, Quantity: 50, TransactionType: models.ReceiveStock, TransactionDate: now})

    levels, err := s.repo.GetStockLevelsByItem(s.ctx, s.testItem1.ID, now)
    s.NoError(err)
    s.NotNil(levels)
    s.Len(levels, 2, "Item 1 should have stock in 2 warehouses")
    s.Equal(12.0, levels[s.testWarehouse1.ID])
    s.Equal(7.0, levels[s.testWarehouse2.ID])

    levelsItem2, err := s.repo.GetStockLevelsByItem(s.ctx, s.testItem2.ID, now)
    s.NoError(err)
    s.Len(levelsItem2, 1)
    s.Equal(50.0, levelsItem2[s.testWarehouse1.ID])
}


func (s *InventoryTransactionRepositoryIntegrationTestSuite) TestGetStockLevelsByWarehouse() {
    s.T().Log("Running TestGetStockLevelsByWarehouse")
    now := time.Now()
    // WH1: Item1 has 12, Item2 has 50
    s.repo.Create(s.ctx, &models.InventoryTransaction{ItemID: s.testItem1.ID, WarehouseID: s.testWarehouse1.ID, Quantity: 10, TransactionType: models.ReceiveStock, TransactionDate: now.Add(-2 * time.Hour)})
	s.repo.Create(s.ctx, &models.InventoryTransaction{ItemID: s.testItem1.ID, WarehouseID: s.testWarehouse1.ID, Quantity: 3, TransactionType: models.IssueStock, TransactionDate: now.Add(-time.Hour)})
	s.repo.Create(s.ctx, &models.InventoryTransaction{ItemID: s.testItem1.ID, WarehouseID: s.testWarehouse1.ID, Quantity: 5, TransactionType: models.AdjustStockIn, TransactionDate: now})
    s.repo.Create(s.ctx, &models.InventoryTransaction{ItemID: s.testItem2.ID, WarehouseID: s.testWarehouse1.ID, Quantity: 50, TransactionType: models.ReceiveStock, TransactionDate: now})
    // WH2: Item1 has 7
	s.repo.Create(s.ctx, &models.InventoryTransaction{ItemID: s.testItem1.ID, WarehouseID: s.testWarehouse2.ID, Quantity: 7, TransactionType: models.ReceiveStock, TransactionDate: now})

    levels, err := s.repo.GetStockLevelsByWarehouse(s.ctx, s.testWarehouse1.ID, now)
    s.NoError(err)
    s.NotNil(levels)
    s.Len(levels, 2, "Warehouse 1 should have stock of 2 items")
    s.Equal(12.0, levels[s.testItem1.ID])
    s.Equal(50.0, levels[s.testItem2.ID])

    levelsWH2, err := s.repo.GetStockLevelsByWarehouse(s.ctx, s.testWarehouse2.ID, now)
    s.NoError(err)
    s.Len(levelsWH2, 1)
    s.Equal(7.0, levelsWH2[s.testItem1.ID])
}
