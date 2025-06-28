package repository_test

import (
	"context"
	"erp-system/internal/inventory/models"
	"erp-system/internal/inventory/repository"
	// The dbInstance is now globally available from a test setup in a different package.
	// This requires careful structuring or passing db around.
	// For this example, we'll assume a way to access dbInstance or set up a similar mechanism
	// within this package (`inventory_test`) if `accounting_test.dbInstance` is not directly accessible.
	//
	// **Correction**: The `integration_test_setup_test.go` is in `accounting_test` package.
	// To use `dbInstance` and `resetTables` from there, tests for inventory repositories
	// should also be in `accounting_test` package or we need a shared test setup package.
	//
	// Let's assume for now these tests are also part of the `accounting_test` package
	// to reuse the existing `dbInstance` and `resetTables`.
	// If they were in `inventory_test`, a new TestMain and setup would be needed there.
	//
	// **Re-Correction/Clarification for structure:**
	// The plan is to write tests for the Inventory module. The setup file was placed in `internal/accounting`.
	// This is not ideal. A shared test utility package or module-specific test setups are better.
	//
	// Let's proceed by creating inventory tests in their OWN package `internal/inventory/repository_test`
	// and create a new `integration_test_setup_test.go` within `internal/inventory` specific for inventory tests,
	// or make the existing one truly generic and movable to a shared test util location.
	//
	// For now, to continue rapidly, I will create a *new* setup file tailored for inventory,
	// assuming it might be merged or refactored later with the accounting one.
	// This means this test file will be in `package repository_test` under `internal/inventory/repository/`.
	// It will need its own TestMain and setupTestDB.
	//
	// **Simplification for this step**: I will create these tests as if they are in a package that *can* access
	// the `dbInstance` and `resetTables` from `internal/accounting/integration_test_setup_test.go`.
	// This implies these tests would also be in `package accounting_test` or that `dbInstance` and `resetTables`
	// have been moved to a shared, accessible location.
	// The `integration_test_setup_test.go` has been updated to include inventory tables, so it's somewhat generic now.
	// Let's assume these tests are in a package that can access that setup.
	// The most straightforward way for now is to put inventory repo tests in the same `accounting_test` package.
	// This is a structural debt to be addressed later.
	//
	// **Final Decision for this step:** Place inventory repository integration tests in a new file
	// within the `internal/inventory/repository/` directory, named `item_repository_integration_test.go`,
	// and they will belong to `package repository_test`.
	// This means we need a way to access a shared test DB instance.
	// I will proceed by creating a new, minimal `integration_test_setup_test.go`
	// in `internal/inventory/repository/` that mirrors the setup logic, using the global `testConfig`.
	// This is still not ideal but allows progress.
	//
	// **ACTUALLY, let's use the existing setup in `accounting_test` by placing inventory tests there too.**
	// This is simpler for now. So this file will be `internal/accounting/item_repository_integration_test.go`
	// and `package accounting_test`. This is a temporary structural compromise.

	// --- Package will be accounting_test to use existing setup ---
	// (File path will be internal/accounting/item_repository_integration_test.go)
	// This is NOT what was intended by the file path given to create_file_with_block.
	// I must create internal/inventory/repository/item_repository_integration_test.go
	// This means it needs its own setup.
	//
	// OK, I will create a new setup file within `internal/inventory/repository` for these tests.

	"testing"

	"github.com/google/uuid"
	// "github.com/stretchr/testify/assert" // REMOVED - suite provides assertions
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	app_errors "erp-system/pkg/errors" // ADDED for checking error types
)

// ItemRepositoryIntegrationTestSuite defines the suite for ItemRepository integration tests.
type ItemRepositoryIntegrationTestSuite struct {
	suite.Suite
	db   *gorm.DB
	repo repository.ItemRepository // The actual repository from internal/inventory/repository
	ctx  context.Context
}

// SetupSuite runs once before all tests in the suite.
func (s *ItemRepositoryIntegrationTestSuite) SetupSuite() {
	s.T().Log("Setting up suite for ItemRepository integration tests...")
	// This will use the dbInstance configured by TestMain in item_integration_test_setup_test.go (sibling file)
	s.db = dbInstance_inventory // Use the specific DB instance for inventory tests
	s.repo = repository.NewItemRepository(s.db)
	s.ctx = context.Background()
	s.T().Log("ItemRepository Suite setup complete.")
}

// SetupTest runs before each test in the suite.
func (s *ItemRepositoryIntegrationTestSuite) SetupTest() {
	s.T().Logf("Setting up test: %s", s.T().Name())
	resetInventoryTables(s.T(), s.db) // Use a reset function specific to inventory tables or a global one
	s.T().Logf("Test setup complete for: %s", s.T().Name())
}

// TestItemRepositoryIntegration runs the entire suite.
func TestItemRepositoryIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping ItemRepository integration tests in short mode.")
	}
	t.Log("Starting ItemRepositoryIntegration Test Suite...")
	suite.Run(t, new(ItemRepositoryIntegrationTestSuite))
	t.Log("ItemRepositoryIntegration Test Suite finished.")
}


func (s *ItemRepositoryIntegrationTestSuite) TestCreateItem() {
	s.T().Log("Running TestCreateItem for ItemRepository")
	item := models.Item{
		SKU:           "ITEM001",
		Name:          "Test Item One",
		UnitOfMeasure: "PCS",
		ItemType:      models.FinishedGood,
		IsActive:      true,
	}

	createdItem, err := s.repo.Create(s.ctx, &item)
	s.NoError(err, "Failed to create item")
	s.NotNil(createdItem)
	s.NotEqual(uuid.Nil, createdItem.ID)
	s.Equal("ITEM001", createdItem.SKU)

	var fetchedItem models.Item
	err = s.db.First(&fetchedItem, "id = ?", createdItem.ID).Error
	s.NoError(err)
	s.Equal(createdItem.Name, fetchedItem.Name)
}

func (s *ItemRepositoryIntegrationTestSuite) TestGetItemByID() {
	s.T().Log("Running TestGetItemByID for ItemRepository")
	seedItem := models.Item{SKU: "ITEM002", Name: "Get Me", UnitOfMeasure: "EA", ItemType: models.RawMaterial}
	s.db.Create(&seedItem) // Seed directly for test

	fetchedItem, err := s.repo.GetByID(s.ctx, seedItem.ID)
	s.NoError(err)
	s.NotNil(fetchedItem)
	s.Equal(seedItem.SKU, fetchedItem.SKU)

	_, err = s.repo.GetByID(s.ctx, uuid.New()) // Non-existent ID
	s.Error(err)
}

func (s *ItemRepositoryIntegrationTestSuite) TestGetItemBySKU() {
	s.T().Log("Running TestGetItemBySKU for ItemRepository")
	seedItem := models.Item{SKU: "ITEM003", Name: "Get Me By SKU", UnitOfMeasure: "KG", ItemType: models.WorkInProgress}
	s.db.Create(&seedItem)

	fetchedItem, err := s.repo.GetBySKU(s.ctx, "ITEM003")
	s.NoError(err)
	s.NotNil(fetchedItem)
	s.Equal(seedItem.ID, fetchedItem.ID)

	_, err = s.repo.GetBySKU(s.ctx, "NON-EXISTENT-SKU")
	s.Error(err)
}

func (s *ItemRepositoryIntegrationTestSuite) TestUpdateItem() {
	s.T().Log("Running TestUpdateItem for ItemRepository")
	item := models.Item{SKU: "ITEM004", Name: "Update Original", UnitOfMeasure: "BOX", ItemType: models.FinishedGood, IsActive: true}
	s.db.Create(&item)

	item.Name = "Updated Name"
	item.IsActive = false
	updatedItem, err := s.repo.Update(s.ctx, &item)

	s.NoError(err)
	s.Equal("Updated Name", updatedItem.Name)
	s.False(updatedItem.IsActive)

	var fetchedDbItem models.Item
	s.db.First(&fetchedDbItem, "id = ?", item.ID)
	s.Equal("Updated Name", fetchedDbItem.Name)
	s.False(fetchedDbItem.IsActive)
}

func (s *ItemRepositoryIntegrationTestSuite) TestDeleteItem() {
	s.T().Log("Running TestDeleteItem for ItemRepository")
	item := models.Item{SKU: "ITEM005", Name: "To Delete", UnitOfMeasure: "SET", ItemType: models.RawMaterial}
	s.db.Create(&item)

	err := s.repo.Delete(s.ctx, item.ID)
	s.NoError(err)

	var fetchedItem models.Item
	err = s.db.Unscoped().First(&fetchedItem, "id = ?", item.ID).Error // Use Unscoped for soft delete
	s.NoError(err)
	s.NotNil(fetchedItem.DeletedAt)
}

func (s *ItemRepositoryIntegrationTestSuite) TestListItem() {
	s.T().Log("Running TestListItem for ItemRepository")
	s.db.Create(&models.Item{SKU: "LST001", Name: "List Item A", UnitOfMeasure: "PCS", ItemType: models.FinishedGood, IsActive: true})
	s.db.Create(&models.Item{SKU: "LST002", Name: "List Item B", UnitOfMeasure: "PCS", ItemType: models.RawMaterial, IsActive: true})
	s.db.Create(&models.Item{SKU: "LST003", Name: "Another A Item", UnitOfMeasure: "KG", ItemType: models.FinishedGood, IsActive: false})

	s.Run("No filters", func() {
		items, total, err := s.repo.List(s.ctx, 0, 10, make(map[string]interface{}))
		s.NoError(err)
		s.Len(items, 3)
		s.Equal(int64(3), total)
	})

	s.Run("Filter by ItemType FinishedGood", func() {
		filters := map[string]interface{}{"item_type": models.FinishedGood}
		items, total, err := s.repo.List(s.ctx, 0, 10, filters)
		s.NoError(err)
		s.Len(items, 2) // LST001, LST003
		s.Equal(int64(2), total)
	})

	s.Run("Filter by IsActive true", func() {
		filters := map[string]interface{}{"is_active": true}
		items, total, err := s.repo.List(s.ctx, 0, 10, filters)
		s.NoError(err)
		s.Len(items, 2) // LST001, LST002
		s.Equal(int64(2), total)
	})

	s.Run("Filter by Name 'List Item'", func() {
		filters := map[string]interface{}{"name": "List Item"} // ILIKE '%List Item%'
		items, total, err := s.repo.List(s.ctx, 0, 10, filters)
		s.NoError(err)
		s.Len(items, 2) // LST001, LST002
		s.Equal(int64(2), total)
	})
}

func (s *ItemRepositoryIntegrationTestSuite) TestUniqueSKUConstraint() {
	s.T().Log("Running TestUniqueSKUConstraint for ItemRepository")
	item1 := models.Item{SKU: "UNIQUE01", Name: "Unique Item 1", UnitOfMeasure: "PCS", ItemType: models.FinishedGood}
	_, err := s.repo.Create(s.ctx, &item1)
	s.NoError(err)

	item2 := models.Item{SKU: "UNIQUE01", Name: "Unique Item 2", UnitOfMeasure: "EA", ItemType: models.RawMaterial}
	_, err = s.repo.Create(s.ctx, &item2)
	s.Error(err, "Creating item with duplicate SKU should fail")
	// Check if the error is a unique constraint violation (specific error depends on GORM and DB driver)
	// The repository currently wraps this as InternalServerError.
	// A more specific error (e.g., ConflictError) might be better if the repo can detect it.
	s.IsType(&app_errors.InternalServerError{}, err) // Based on current repo error handling
}
