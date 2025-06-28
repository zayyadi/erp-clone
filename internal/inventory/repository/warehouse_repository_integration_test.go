package repository_test

import (
	"context"
	"erp-system/internal/inventory/models"
	"erp-system/internal/inventory/repository"
	"testing"

	"github.com/google/uuid"
	// "github.com/stretchr/testify/assert" // REMOVED - suite provides assertions
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// WarehouseRepositoryIntegrationTestSuite defines the suite for WarehouseRepository integration tests.
type WarehouseRepositoryIntegrationTestSuite struct {
	suite.Suite
	db   *gorm.DB
	repo repository.WarehouseRepository // The actual repository from internal/inventory/repository
	ctx  context.Context
}

// SetupSuite runs once before all tests in the suite.
func (s *WarehouseRepositoryIntegrationTestSuite) SetupSuite() {
	s.T().Log("Setting up suite for WarehouseRepository integration tests...")
	s.db = dbInstance_inventory // Use the DB instance from item_integration_test_setup_test.go
	s.repo = repository.NewWarehouseRepository(s.db)
	s.ctx = context.Background()
	s.T().Log("WarehouseRepository Suite setup complete.")
}

// SetupTest runs before each test in the suite.
func (s *WarehouseRepositoryIntegrationTestSuite) SetupTest() {
	s.T().Logf("Setting up test: %s", s.T().Name())
	resetInventoryTables(s.T(), s.db) // Reset relevant tables
	s.T().Logf("Test setup complete for: %s", s.T().Name())
}

// TestWarehouseRepositoryIntegration runs the entire suite.
func TestWarehouseRepositoryIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping WarehouseRepository integration tests in short mode.")
	}
	t.Log("Starting WarehouseRepositoryIntegration Test Suite...")
	suite.Run(t, new(WarehouseRepositoryIntegrationTestSuite))
	t.Log("WarehouseRepositoryIntegration Test Suite finished.")
}

func (s *WarehouseRepositoryIntegrationTestSuite) TestCreateWarehouse() {
	s.T().Log("Running TestCreateWarehouse for WarehouseRepository")
	warehouse := models.Warehouse{
		Code:     "WH001",
		Name:     "Main Warehouse",
		Location: "123 Main St",
		IsActive: true,
	}

	createdWarehouse, err := s.repo.Create(s.ctx, &warehouse)
	s.NoError(err, "Failed to create warehouse")
	s.NotNil(createdWarehouse)
	s.NotEqual(uuid.Nil, createdWarehouse.ID)
	s.Equal("WH001", createdWarehouse.Code)

	var fetchedWarehouse models.Warehouse
	err = s.db.First(&fetchedWarehouse, "id = ?", createdWarehouse.ID).Error
	s.NoError(err)
	s.Equal(createdWarehouse.Name, fetchedWarehouse.Name)
}

func (s *WarehouseRepositoryIntegrationTestSuite) TestGetWarehouseByID() {
	s.T().Log("Running TestGetWarehouseByID for WarehouseRepository")
	seedWarehouse := models.Warehouse{Code: "WH002", Name: "Secondary Warehouse", Location: "456 Side Ave"}
	s.db.Create(&seedWarehouse)

	fetchedWarehouse, err := s.repo.GetByID(s.ctx, seedWarehouse.ID)
	s.NoError(err)
	s.NotNil(fetchedWarehouse)
	s.Equal(seedWarehouse.Code, fetchedWarehouse.Code)

	_, err = s.repo.GetByID(s.ctx, uuid.New()) // Non-existent ID
	s.Error(err)
}

func (s *WarehouseRepositoryIntegrationTestSuite) TestGetWarehouseByCode() {
	s.T().Log("Running TestGetWarehouseByCode for WarehouseRepository")
	seedWarehouse := models.Warehouse{Code: "WH003", Name: "Warehouse by Code", Location: "789 Code Rd"}
	s.db.Create(&seedWarehouse)

	fetchedWarehouse, err := s.repo.GetByCode(s.ctx, "WH003")
	s.NoError(err)
	s.NotNil(fetchedWarehouse)
	s.Equal(seedWarehouse.ID, fetchedWarehouse.ID)

	_, err = s.repo.GetByCode(s.ctx, "NON-EXISTENT-CODE")
	s.Error(err)
}

func (s *WarehouseRepositoryIntegrationTestSuite) TestUpdateWarehouse() {
	s.T().Log("Running TestUpdateWarehouse for WarehouseRepository")
	warehouse := models.Warehouse{Code: "WH004", Name: "Original WH Name", Location: "Old Location", IsActive: true}
	s.db.Create(&warehouse)

	warehouse.Name = "Updated WH Name"
	warehouse.IsActive = false
	updatedWarehouse, err := s.repo.Update(s.ctx, &warehouse)

	s.NoError(err)
	s.Equal("Updated WH Name", updatedWarehouse.Name)
	s.False(updatedWarehouse.IsActive)

	var fetchedDbWarehouse models.Warehouse
	s.db.First(&fetchedDbWarehouse, "id = ?", warehouse.ID)
	s.Equal("Updated WH Name", fetchedDbWarehouse.Name)
	s.False(fetchedDbWarehouse.IsActive)
}

func (s *WarehouseRepositoryIntegrationTestSuite) TestDeleteWarehouse() {
	s.T().Log("Running TestDeleteWarehouse for WarehouseRepository")
	warehouse := models.Warehouse{Code: "WH005", Name: "Warehouse To Delete", Location: "Delete Me St"}
	s.db.Create(&warehouse)

	err := s.repo.Delete(s.ctx, warehouse.ID)
	s.NoError(err)

	var fetchedWarehouse models.Warehouse
	err = s.db.Unscoped().First(&fetchedWarehouse, "id = ?", warehouse.ID).Error // Use Unscoped for soft delete
	s.NoError(err)
	s.NotNil(fetchedWarehouse.DeletedAt)
}

func (s *WarehouseRepositoryIntegrationTestSuite) TestListWarehouses() {
	s.T().Log("Running TestListWarehouses for WarehouseRepository")
	s.db.Create(&models.Warehouse{Code: "LWH001", Name: "Alpha Warehouse", Location: "East Wing", IsActive: true})
	s.db.Create(&models.Warehouse{Code: "LWH002", Name: "Beta Warehouse", Location: "West Wing", IsActive: true})
	s.db.Create(&models.Warehouse{Code: "LWH003", Name: "Charlie Inactive", Location: "North Wing", IsActive: false})

	s.Run("No filters", func() {
		warehouses, total, err := s.repo.List(s.ctx, 0, 10, make(map[string]interface{}))
		s.NoError(err)
		s.Len(warehouses, 3)
		s.Equal(int64(3), total)
	})

	s.Run("Filter by IsActive true", func() {
		filters := map[string]interface{}{"is_active": true}
		warehouses, total, err := s.repo.List(s.ctx, 0, 10, filters)
		s.NoError(err)
		s.Len(warehouses, 2)
		s.Equal(int64(2), total)
	})

	s.Run("Filter by Name 'Warehouse'", func() {
		filters := map[string]interface{}{"name": "Warehouse"} // ILIKE '%Warehouse%'
		warehouses, total, err := s.repo.List(s.ctx, 0, 10, filters)
		s.NoError(err)
		s.Len(warehouses, 2) // Alpha, Beta
		s.Equal(int64(2), total)
	})
}

func (s *WarehouseRepositoryIntegrationTestSuite) TestUniqueWarehouseCodeConstraint() {
	s.T().Log("Running TestUniqueWarehouseCodeConstraint for WarehouseRepository")
	wh1 := models.Warehouse{Code: "UWH001", Name: "Unique WH 1"}
	_, err := s.repo.Create(s.ctx, &wh1)
	s.NoError(err)

	wh2 := models.Warehouse{Code: "UWH001", Name: "Unique WH 2"}
	_, err = s.repo.Create(s.ctx, &wh2)
	s.Error(err, "Creating warehouse with duplicate code should fail")
	// The repository currently wraps this as InternalServerError.
}
