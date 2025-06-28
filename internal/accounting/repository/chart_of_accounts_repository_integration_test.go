package repository_test // Changed package declaration

import (
	"context"
	"erp-system/internal/accounting/models"
	"erp-system/internal/accounting/repository"
	// Use the global dbInstance from the setup file in the same package (accounting_test)
	// "erp-system/internal/accounting" // This would be if setup was in a different package
	"testing"

	"github.com/google/uuid"
	// "github.com/stretchr/testify/assert" // REMOVED - suite provides assertions
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	// app_errors "erp-system/pkg/errors" // REMOVED - Not actually used for IsType check
)

// ChartOfAccountRepositoryIntegrationTestSuite defines the suite for ChartOfAccountRepository integration tests.
type ChartOfAccountRepositoryIntegrationTestSuite struct {
	suite.Suite
	db   *gorm.DB
	repo repository.ChartOfAccountRepository
	ctx  context.Context
}

// SetupSuite runs once before all tests in the suite.
func (s *ChartOfAccountRepositoryIntegrationTestSuite) SetupSuite() {
	s.T().Log("Setting up suite for ChartOfAccountRepository integration tests...")
	s.db = dbInstance_acc_repo // Use the DB instance from the local setup
	s.repo = repository.NewChartOfAccountRepository(s.db)
	s.ctx = context.Background()
	s.T().Log("Suite setup complete.")
}

// SetupTest runs before each test in the suite.
func (s *ChartOfAccountRepositoryIntegrationTestSuite) SetupTest() {
	s.T().Logf("Setting up test: %s", s.T().Name())
	resetAccRepoTables(s.T(), s.db) // Use the reset function from the local setup
	s.T().Logf("Test setup complete for: %s", s.T().Name())
}

// TearDownTest runs after each test method.
func (s *ChartOfAccountRepositoryIntegrationTestSuite) TearDownTest() {
    s.T().Logf("Tearing down test: %s", s.T().Name())
    // resetTables(s.T(), s.db) // Clean up database after each test
    s.T().Logf("Test teardown complete for: %s", s.T().Name())
}


// TestChartOfAccountRepositoryIntegration runs the entire suite.
func TestChartOfAccountRepositoryIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping accounting repository integration tests in short mode.")
	}
	t.Log("Starting ChartOfAccountRepositoryIntegration Test Suite...")
	suite.Run(t, new(ChartOfAccountRepositoryIntegrationTestSuite))
	t.Log("ChartOfAccountRepositoryIntegration Test Suite finished.")
}


func (s *ChartOfAccountRepositoryIntegrationTestSuite) TestCreateChartOfAccount() {
	s.T().Log("Running TestCreateChartOfAccount")
	account := models.ChartOfAccount{
		AccountCode: "1010",
		AccountName: "Cash Test",
		AccountType: models.Asset,
		IsActive:    true,
	}

	createdAccount, err := s.repo.Create(s.ctx, &account)
	s.NoError(err, "Failed to create chart of account") // Uses suite's s.NoError
	s.NotNil(createdAccount, "Created account should not be nil")
	s.NotEqual(uuid.Nil, createdAccount.ID, "ID should be set")
	s.Equal("1010", createdAccount.AccountCode)

	// Verify it's in the DB
	var fetchedAccount models.ChartOfAccount
	err = s.db.First(&fetchedAccount, "id = ?", createdAccount.ID).Error
	s.NoError(err, "Failed to fetch created account from DB")
	s.Equal(createdAccount.AccountName, fetchedAccount.AccountName)
}

func (s *ChartOfAccountRepositoryIntegrationTestSuite) TestGetChartOfAccountByID() {
	s.T().Log("Running TestGetChartOfAccountByID")
	newAccount := models.ChartOfAccount{
		AccountCode: "1020", AccountName: "AR Test", AccountType: models.Asset, IsActive: true,
	}
	// Manually insert using GORM for setup, or use repo.Create
	s.db.Create(&newAccount)
	s.NotEqual(uuid.Nil, newAccount.ID)

	fetchedAccount, err := s.repo.GetByID(s.ctx, newAccount.ID)
	s.NoError(err, "Error getting account by ID")
	s.NotNil(fetchedAccount)
	s.Equal(newAccount.AccountCode, fetchedAccount.AccountCode)

	// Test Not Found
	nonExistentID := uuid.New()
	_, err = s.repo.GetByID(s.ctx, nonExistentID)
	s.Error(err, "Expected error for non-existent ID")
	// s.True(errors.Is(err, gorm.ErrRecordNotFound), "Expected gorm.ErrRecordNotFound") // Service wraps this
}

func (s *ChartOfAccountRepositoryIntegrationTestSuite) TestGetChartOfAccountByCode() {
	s.T().Log("Running TestGetChartOfAccountByCode")
	newAccount := models.ChartOfAccount{
		AccountCode: "1030", AccountName: "Inventory Test", AccountType: models.Asset, IsActive: true,
	}
	s.db.Create(&newAccount)

	fetchedAccount, err := s.repo.GetByCode(s.ctx, "1030")
	s.NoError(err, "Error getting account by code")
	s.NotNil(fetchedAccount)
	s.Equal(newAccount.ID, fetchedAccount.ID)

	// Test Not Found
	_, err = s.repo.GetByCode(s.ctx, "NON_EXISTENT_CODE")
	s.Error(err, "Expected error for non-existent code")
}

func (s *ChartOfAccountRepositoryIntegrationTestSuite) TestUpdateChartOfAccount() {
	s.T().Log("Running TestUpdateChartOfAccount")
	account := models.ChartOfAccount{
		AccountCode: "1040", AccountName: "Prepaid Expenses", AccountType: models.Asset, IsActive: true,
	}
	s.db.Create(&account)

	account.AccountName = "Updated Prepaid Expenses"
	account.IsActive = false
	updatedAccount, err := s.repo.Update(s.ctx, &account)

	s.NoError(err, "Error updating account")
	s.NotNil(updatedAccount)
	s.Equal("Updated Prepaid Expenses", updatedAccount.AccountName)
	s.False(updatedAccount.IsActive, "IsActive should be false after update")

	// Verify in DB
	var fetchedDbAccount models.ChartOfAccount
	s.db.First(&fetchedDbAccount, "id = ?", account.ID)
	s.Equal("Updated Prepaid Expenses", fetchedDbAccount.AccountName)
	s.False(fetchedDbAccount.IsActive)
}

func (s *ChartOfAccountRepositoryIntegrationTestSuite) TestDeleteChartOfAccount() {
	s.T().Log("Running TestDeleteChartOfAccount")
	account := models.ChartOfAccount{
		AccountCode: "1050", AccountName: "To Be Deleted", AccountType: models.Asset, IsActive: true,
	}
	s.db.Create(&account)

	err := s.repo.Delete(s.ctx, account.ID)
	s.NoError(err, "Error deleting account")

	// Verify soft delete (deleted_at is set)
	var fetchedAccount models.ChartOfAccount
	// Use Unscoped to retrieve soft-deleted records
	err = s.db.Unscoped().First(&fetchedAccount, "id = ?", account.ID).Error
	s.NoError(err, "Error fetching soft-deleted account")
	s.NotNil(fetchedAccount.DeletedAt, "DeletedAt should be set for soft-deleted record")
}

func (s *ChartOfAccountRepositoryIntegrationTestSuite) TestListChartOfAccounts() {
	s.T().Log("Running TestListChartOfAccounts")
	// Seed some data
	accountsToSeed := []models.ChartOfAccount{
		{AccountCode: "L100", AccountName: "List Acc 1", AccountType: models.Asset, IsActive: true},
		{AccountCode: "L101", AccountName: "List Acc 2", AccountType: models.Liability, IsActive: true},
		{AccountCode: "L102", AccountName: "Another Asset", AccountType: models.Asset, IsActive: false}, // Inactive
		{AccountCode: "L103", AccountName: "List Acc 3 Asset", AccountType: models.Asset, IsActive: true},
	}
	for _, acc := range accountsToSeed {
		// Need to create a new variable for the pointer in loop for GORM create
		tempAcc := acc
		err := s.db.Create(&tempAcc).Error
		s.Require().NoError(err, "Failed to seed account for TestListChartOfAccounts")
	}
	s.T().Logf("Seeded %d accounts", len(accountsToSeed))


	s.Run("No filters, default pagination", func() {
		s.T().Log("Subtest: No filters, default pagination")
		accounts, total, err := s.repo.List(s.ctx, 0, 10, make(map[string]interface{}))
		s.NoError(err)
		s.Len(accounts, 4, "Should list all (including inactive by default if filter not specified)")
		s.Equal(int64(4), total) // Total includes all, active or not, if no filter
	})

	s.Run("Filter by AccountType Asset", func() {
		s.T().Log("Subtest: Filter by AccountType Asset")
		filters := map[string]interface{}{"account_type": models.Asset}
		accounts, total, err := s.repo.List(s.ctx, 0, 10, filters)
		s.NoError(err)
		// L100, L102 (inactive), L103 are Assets
		s.Len(accounts, 3)
		s.Equal(int64(3), total)
		for _, acc := range accounts {
			s.Equal(models.Asset, acc.AccountType)
		}
	})

	s.Run("Filter by IsActive true", func() {
		s.T().Log("Subtest: Filter by IsActive true")
		filters := map[string]interface{}{"is_active": true}
		accounts, total, err := s.repo.List(s.ctx, 0, 10, filters)
		s.NoError(err)
		// L100, L101, L103 are active
		s.Len(accounts, 3)
		s.Equal(int64(3), total)
		for _, acc := range accounts {
			s.True(acc.IsActive)
		}
	})

	s.Run("Filter by AccountName (partial match)", func() {
		s.T().Log("Subtest: Filter by AccountName (partial match)")
		filters := map[string]interface{}{"account_name": "List Acc"}
		accounts, total, err := s.repo.List(s.ctx, 0, 10, filters)
		s.NoError(err)
		// L100, L101, L103
		s.Len(accounts, 3)
		s.Equal(int64(3), total)
	})

	s.Run("Pagination limit 1, page 1", func() {
		s.T().Log("Subtest: Pagination limit 1, page 1")
		// Order is by account_code asc by default in repo
		accounts, total, err := s.repo.List(s.ctx, 0, 1, make(map[string]interface{}))
		s.NoError(err)
		s.Len(accounts, 1)
		s.Equal(int64(4), total) // Total count should still be all matching accounts
		s.Equal("L100", accounts[0].AccountCode) // Assuming L100 is first by code
	})

	s.Run("Pagination limit 1, page 2", func() {
		s.T().Log("Subtest: Pagination limit 1, page 2")
		accounts, total, err := s.repo.List(s.ctx, 1, 1, make(map[string]interface{})) // offset 1, limit 1
		s.NoError(err)
		s.Len(accounts, 1)
		s.Equal(int64(4), total)
		s.Equal("L101", accounts[0].AccountCode) // Assuming L101 is second by code
	})
}

// TestParentChildRelationship tests creating an account with a parent.
func (s *ChartOfAccountRepositoryIntegrationTestSuite) TestParentChildRelationship() {
	s.T().Log("Running TestParentChildRelationship")
	parentAccount := models.ChartOfAccount{
		AccountCode: "P100", AccountName: "Parent Asset", AccountType: models.Asset, IsActive: true,
	}
	createdParent, err := s.repo.Create(s.ctx, &parentAccount)
	s.NoError(err)
	s.NotNil(createdParent)

	childAccount := models.ChartOfAccount{
		AccountCode:     "C101",
		AccountName:     "Child Asset",
		AccountType:     models.Asset,
		IsActive:        true,
		ParentAccountID: &createdParent.ID,
	}
	createdChild, err := s.repo.Create(s.ctx, &childAccount)
	s.NoError(err)
	s.NotNil(createdChild)
	s.Require().NotNil(createdChild.ParentAccountID, "Child's ParentAccountID should be set")
	s.Equal(createdParent.ID, *createdChild.ParentAccountID)

	// Fetch child and check parent ID
	fetchedChild, err := s.repo.GetByID(s.ctx, createdChild.ID)
	s.NoError(err)
	s.Require().NotNil(fetchedChild.ParentAccountID)
	s.Equal(createdParent.ID, *fetchedChild.ParentAccountID)
}

// TestUniqueAccountCodeConstraint tests the unique constraint on account_code.
func (s *ChartOfAccountRepositoryIntegrationTestSuite) TestUniqueAccountCodeConstraint() {
	s.T().Log("Running TestUniqueAccountCodeConstraint")
	account1 := models.ChartOfAccount{
		AccountCode: "U100", AccountName: "Unique Acc 1", AccountType: models.Asset, IsActive: true,
	}
	_, err := s.repo.Create(s.ctx, &account1)
	s.NoError(err, "First account creation should succeed")

	account2 := models.ChartOfAccount{
		AccountCode: "U100", AccountName: "Unique Acc 2", AccountType: models.Liability, IsActive: true,
	}
	_, err = s.repo.Create(s.ctx, &account2)
	s.Error(err, "Second account creation with same code should fail")
	// The actual error from GORM/Postgres for unique violation might be driver-specific.
	// For Postgres, it's often "ERROR: duplicate key value violates unique constraint (SQLSTATE 23505)"
	// The repository wraps this into an internal server error currently.
	// s.Contains(err.Error(), "duplicate key value violates unique constraint", "Error message should indicate unique constraint violation")
	// Or, if your repo maps this to a custom error:
	// assert.IsType(s.T(), &app_errors.ConflictError{}, err)
}
