package repository_test // Changed package declaration

import (
	"context"
	"erp-system/internal/accounting/models"
	"erp-system/internal/accounting/repository"
	// app_errors "erp-system/pkg/errors" // Alias to avoid conflict - REMOVED as unused directly by suite methods
	"testing"
	"time"

	"github.com/google/uuid"
	// "github.com/stretchr/testify/assert" // REMOVED as suite.Assert() or s.Assert() is used
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	app_errors "erp-system/pkg/errors" // Re-add if specific error type checks are needed outside suite methods
)

// JournalEntryRepositoryIntegrationTestSuite defines the suite for JournalEntryRepository integration tests.
type JournalEntryRepositoryIntegrationTestSuite struct {
	suite.Suite
	db   *gorm.DB
	repo repository.JournalEntryRepository
	coaRepo repository.ChartOfAccountRepository // For setting up accounts
	ctx  context.Context

	// Some common accounts for tests
	cashAccount    *models.ChartOfAccount
	revenueAccount *models.ChartOfAccount
	expenseAccount *models.ChartOfAccount
}

// SetupSuite runs once before all tests in the suite.
func (s *JournalEntryRepositoryIntegrationTestSuite) SetupSuite() {
	s.T().Log("Setting up suite for JournalEntryRepository integration tests...")
	s.db = dbInstance_acc_repo // Use the DB instance from the local setup
	s.repo = repository.NewJournalEntryRepository(s.db)
	s.coaRepo = repository.NewChartOfAccountRepository(s.db) // Initialize COA repo
	s.ctx = context.Background()
	s.T().Log("Suite setup complete.")
}

// SetupTest runs before each test in the suite.
func (s *JournalEntryRepositoryIntegrationTestSuite) SetupTest() {
	s.T().Logf("Setting up test: %s", s.T().Name())
	resetAccRepoTables(s.T(), s.db) // Use the reset function from the local setup

	// Create common accounts for use in tests
	s.cashAccount = s.createTestAccount("1010", "Cash Test", models.Asset)
	s.revenueAccount = s.createTestAccount("4010", "Revenue Test", models.Revenue)
	s.expenseAccount = s.createTestAccount("5010", "Expense Test", models.Expense)
	s.T().Logf("Test setup complete for: %s", s.T().Name())
}

// Helper to create accounts for tests
func (s *JournalEntryRepositoryIntegrationTestSuite) createTestAccount(code, name string, accType models.AccountType) *models.ChartOfAccount {
	acc := models.ChartOfAccount{AccountCode: code, AccountName: name, AccountType: accType, IsActive: true}
	createdAcc, err := s.coaRepo.Create(s.ctx, &acc)
	s.Require().NoError(err, "Failed to create test account %s", code)
	return createdAcc
}


// TestJournalEntryRepositoryIntegration runs the entire suite.
func TestJournalEntryRepositoryIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping accounting repository integration tests in short mode.")
	}
	t.Log("Starting JournalEntryRepositoryIntegration Test Suite...")
	suite.Run(t, new(JournalEntryRepositoryIntegrationTestSuite))
	t.Log("JournalEntryRepositoryIntegration Test Suite finished.")
}


func (s *JournalEntryRepositoryIntegrationTestSuite) TestCreateJournalEntry_WithLines() {
	s.T().Log("Running TestCreateJournalEntry_WithLines")
	entry := models.JournalEntry{
		EntryDate:   time.Now(),
		Description: "Test Sale Transaction",
		Status:      models.StatusDraft,
		JournalLines: []models.JournalLine{
			{AccountID: s.cashAccount.ID, Amount: 150.75, IsDebit: true, Currency: "USD"},
			{AccountID: s.revenueAccount.ID, Amount: 150.75, IsDebit: false, Currency: "USD"},
		},
	}

	createdEntry, err := s.repo.Create(s.ctx, &entry)
	s.NoError(err, "Failed to create journal entry") // Uses suite's s.NoError
	s.NotNil(createdEntry)
	s.NotEqual(uuid.Nil, createdEntry.ID)
	s.Len(createdEntry.JournalLines, 2, "Should have 2 journal lines")

	// Verify from DB
	var fetchedEntry models.JournalEntry
	err = s.db.Preload("JournalLines").First(&fetchedEntry, "id = ?", createdEntry.ID).Error
	s.NoError(err)
	s.Equal(createdEntry.Description, fetchedEntry.Description)
	s.Len(fetchedEntry.JournalLines, 2)
	s.InDelta(150.75, fetchedEntry.JournalLines[0].Amount, 0.001)
}

func (s *JournalEntryRepositoryIntegrationTestSuite) TestGetJournalEntryByID_WithLines() {
	s.T().Log("Running TestGetJournalEntryByID_WithLines")
	seedEntry := models.JournalEntry{
		EntryDate:   time.Now(), Description: "Seed Entry for Get", Status: models.StatusPosted,
		JournalLines: []models.JournalLine{
			{AccountID: s.expenseAccount.ID, Amount: 50.00, IsDebit: true},
			{AccountID: s.cashAccount.ID, Amount: 50.00, IsDebit: false},
		},
	}
	// Use repo.Create to ensure BeforeCreate hooks run and lines are associated
	createdSeedEntry, err := s.repo.Create(s.ctx, &seedEntry)
	s.Require().NoError(err, "Failed to seed entry for GetByID test")


	fetchedEntry, err := s.repo.GetByID(s.ctx, createdSeedEntry.ID)
	s.NoError(err)
	s.NotNil(fetchedEntry)
	s.Equal(createdSeedEntry.Description, fetchedEntry.Description)
	s.Len(fetchedEntry.JournalLines, 2, "Fetched entry should include lines")
	s.Equal(s.expenseAccount.ID, fetchedEntry.JournalLines[0].AccountID)

	// Test Not Found
	nonExistentID := uuid.New()
	_, err = s.repo.GetByID(s.ctx, nonExistentID)
	s.Error(err, "Expected error for non-existent ID")
	// Re-added app_errors for this specific type check
	_, ok := err.(*app_errors.NotFoundError)
    s.True(ok, "Expected NotFoundError type")
}


func (s *JournalEntryRepositoryIntegrationTestSuite) TestUpdateJournalEntry_HeaderAndLines() {
	s.T().Log("Running TestUpdateJournalEntry_HeaderAndLines")
	// 1. Seed an initial entry
	initialEntry := models.JournalEntry{
		EntryDate:   time.Now().Add(-24 * time.Hour),
		Description: "Initial Entry for Update",
		Status:      models.StatusDraft,
		JournalLines: []models.JournalLine{
			{AccountID: s.cashAccount.ID, Amount: 100.00, IsDebit: true},
			{AccountID: s.revenueAccount.ID, Amount: 100.00, IsDebit: false},
		},
	}
	createdInitialEntry, err := s.repo.Create(s.ctx, &initialEntry)
	s.Require().NoError(err, "Failed to create initial entry for update test")
	s.Require().Len(createdInitialEntry.JournalLines, 2, "Initial entry should have 2 lines")
	initialLine1ID := createdInitialEntry.JournalLines[0].ID
	initialLine2ID := createdInitialEntry.JournalLines[1].ID


	// 2. Prepare update: change description, one line amount, add a new line
	updatedDescription := "Updated Entry Description"
	updatedEntryData := models.JournalEntry{
		ID:          createdInitialEntry.ID, // Must match existing ID
		EntryDate:   createdInitialEntry.EntryDate, // Keep same date or update
		Description: updatedDescription,
		Status:      models.StatusDraft, // Keep status or update
		JournalLines: []models.JournalLine{
			// Update existing line 1 (match by ID if GORM's Replace handles it, or ensure full replacement)
			// For GORM's .Association("JournalLines").Replace(), new lines without ID are created,
			// lines with existing ID that are present are updated, lines with existing ID not present are deleted.
			{ID: initialLine1ID, JournalID: createdInitialEntry.ID, AccountID: s.cashAccount.ID, Amount: 120.00, IsDebit: true}, // Amount changed
			// Line 2 is omitted, so it should be deleted by Replace behavior
			// Add a new line
			{JournalID: createdInitialEntry.ID, AccountID: s.expenseAccount.ID, Amount: 30.00, IsDebit: true},
			{JournalID: createdInitialEntry.ID, AccountID: s.revenueAccount.ID, Amount: 150.00, IsDebit: false}, // New balancing credit line
		},
	}
    // Ensure the updated entry is balanced: 120 DR + 30 DR = 150 DR; 150 CR. Balanced.

	// 3. Perform update
	finalUpdatedEntry, err := s.repo.Update(s.ctx, &updatedEntryData)
	s.NoError(err, "Failed to update journal entry")
	s.NotNil(finalUpdatedEntry)
	s.Equal(updatedDescription, finalUpdatedEntry.Description)

	// Verify lines from DB
	var fetchedEntryAfterUpdate models.JournalEntry
	err = s.db.Preload("JournalLines").Order("journal_lines.created_at asc").First(&fetchedEntryAfterUpdate, "id = ?", createdInitialEntry.ID).Error
	s.NoError(err)

	s.Len(fetchedEntryAfterUpdate.JournalLines, 3, "Should have 3 lines after update (1 updated, 1 deleted, 2 new effectively replaced old 2)")

	foundUpdatedCashLine := false
	foundNewLineExpense := false
	foundNewLineRevenue := false

	for _, line := range fetchedEntryAfterUpdate.JournalLines {
		if line.AccountID == s.cashAccount.ID {
			s.InDelta(120.00, line.Amount, 0.001, "Cash line amount should be updated")
			foundUpdatedCashLine = true
		} else if line.AccountID == s.expenseAccount.ID {
			s.InDelta(30.00, line.Amount, 0.001, "New expense line should exist")
			foundNewLineExpense = true
		} else if line.AccountID == s.revenueAccount.ID && !line.IsDebit { // Ensure it's the new credit line
			s.InDelta(150.00, line.Amount, 0.001, "New revenue line should exist")
			foundNewLineRevenue = true
		}
		// Check that initialLine2ID is not present
		s.NotEqual(initialLine2ID, line.ID, "Original line 2 should have been removed")
	}
	s.True(foundUpdatedCashLine, "Updated cash line not found")
	s.True(foundNewLineExpense, "New expense line not found")
	s.True(foundNewLineRevenue, "New revenue line not found")
}

func (s *JournalEntryRepositoryIntegrationTestSuite) TestDeleteJournalEntry_SoftDeleteCascadesToLines() {
	s.T().Log("Running TestDeleteJournalEntry_SoftDeleteCascadesToLines")
	// Seed an entry with lines
	entry := models.JournalEntry{
		EntryDate: time.Now(), Description: "Entry to be soft-deleted", Status: models.StatusDraft,
		JournalLines: []models.JournalLine{
			{AccountID: s.cashAccount.ID, Amount: 10.00, IsDebit: true},
			{AccountID: s.revenueAccount.ID, Amount: 10.00, IsDebit: false},
		},
	}
	createdEntry, err := s.repo.Create(s.ctx, &entry)
	s.Require().NoError(err)
	s.Require().Len(createdEntry.JournalLines, 2)
	lineID1 := createdEntry.JournalLines[0].ID


	// Perform soft delete on the entry
	err = s.repo.Delete(s.ctx, createdEntry.ID)
	s.NoError(err, "Failed to soft-delete journal entry")

	// Verify entry is soft-deleted
	var fetchedEntry models.JournalEntry
	err = s.db.Unscoped().First(&fetchedEntry, "id = ?", createdEntry.ID).Error
	s.NoError(err)
	s.NotNil(fetchedEntry.DeletedAt, "JournalEntry DeletedAt should be set")

	// Verify lines are hard-deleted (as per current repo.Delete implementation)
	var lineCount int64
	s.db.Model(&models.JournalLine{}).Where("journal_id = ?", createdEntry.ID).Count(&lineCount)
	s.Equal(int64(0), lineCount, "JournalLines should be hard-deleted when entry is soft-deleted by current repo logic")

	// Double check a specific line is gone
	var specificLine models.JournalLine
	err = s.db.First(&specificLine, "id = ?", lineID1).Error
	s.Error(err, "Specific line should not be found after entry soft delete (hard delete of lines)")
	s.ErrorIs(err, gorm.ErrRecordNotFound)
}


func (s *JournalEntryRepositoryIntegrationTestSuite) TestListJournalEntries() {
	s.T().Log("Running TestListJournalEntries")
	// Seed some data
	now := time.Now().Truncate(time.Second) // Truncate for easier comparison if needed
	entry1 := models.JournalEntry{
		EntryDate: now.Add(-2 * 24 * time.Hour), Description: "Older Entry Alpha", Status: models.StatusPosted,
		JournalLines: []models.JournalLine{{AccountID: s.cashAccount.ID, Amount: 10, IsDebit: true}, {AccountID: s.revenueAccount.ID, Amount: 10, IsDebit: false}},
	}
	entry2 := models.JournalEntry{
		EntryDate: now.Add(-1 * 24 * time.Hour), Description: "Recent Entry Beta", Status: models.StatusDraft,
		JournalLines: []models.JournalLine{{AccountID: s.expenseAccount.ID, Amount: 20, IsDebit: true}, {AccountID: s.cashAccount.ID, Amount: 20, IsDebit: false}},
	}
	entry3 := models.JournalEntry{
		EntryDate: now, Description: "Current Entry Alpha", Status: models.StatusPosted,
		JournalLines: []models.JournalLine{{AccountID: s.cashAccount.ID, Amount: 30, IsDebit: true}, {AccountID: s.revenueAccount.ID, Amount: 30, IsDebit: false}},
	}
	_, err := s.repo.Create(s.ctx, &entry1); s.Require().NoError(err)
	_, err = s.repo.Create(s.ctx, &entry2); s.Require().NoError(err)
	_, err = s.repo.Create(s.ctx, &entry3); s.Require().NoError(err)


	s.Run("No filters, default pagination", func() {
		entries, total, err := s.repo.List(s.ctx, 0, 10, make(map[string]interface{}))
		s.NoError(err)
		s.Len(entries, 3)
		s.Equal(int64(3), total)
		// Default order is entry_date desc, created_at desc
		s.Equal("Current Entry Alpha", entries[0].Description)
		s.Equal("Recent Entry Beta", entries[1].Description)
	})

	s.Run("Filter by Status DRAFT", func() {
		filters := map[string]interface{}{"status": models.StatusDraft}
		entries, total, err := s.repo.List(s.ctx, 0, 10, filters)
		s.NoError(err)
		s.Len(entries, 1)
		s.Equal(int64(1), total)
		s.Equal("Recent Entry Beta", entries[0].Description)
	})

	s.Run("Filter by Description 'Alpha'", func() {
		filters := map[string]interface{}{"description": "Alpha"} // ILIKE '%Alpha%'
		entries, total, err := s.repo.List(s.ctx, 0, 10, filters)
		s.NoError(err)
		s.Len(entries, 2)
		s.Equal(int64(2), total)
	})

	s.Run("Filter by DateRange", func() {
		dateFrom := now.Add(-3 * 24 * time.Hour)
		dateTo := now.Add(-1 * 24 * time.Hour).Add(12 * time.Hour) // Should include entry1 and entry2
		filters := map[string]interface{}{"date_from": dateFrom, "date_to": dateTo}
		entries, total, err := s.repo.List(s.ctx, 0, 10, filters)
		s.NoError(err)
		s.Len(entries, 2, "Should find two entries in date range")
		s.Equal(int64(2), total)
	})
}


func (s *JournalEntryRepositoryIntegrationTestSuite) TestUpdateJournalEntryStatus() {
	s.T().Log("Running TestUpdateJournalEntryStatus")
	entry := models.JournalEntry{
		EntryDate: time.Now(), Description: "Status Update Test", Status: models.StatusDraft,
		JournalLines: []models.JournalLine{{AccountID: s.cashAccount.ID, Amount: 5, IsDebit: true}, {AccountID: s.revenueAccount.ID, Amount: 5, IsDebit: false}},
	}
	createdEntry, err := s.repo.Create(s.ctx, &entry); s.Require().NoError(err)

	err = s.repo.UpdateJournalEntryStatus(s.ctx, createdEntry.ID, models.StatusPosted)
	s.NoError(err)

	fetchedEntry, err := s.repo.GetByID(s.ctx, createdEntry.ID)
	s.NoError(err)
	s.Equal(models.StatusPosted, fetchedEntry.Status)

	// Test update non-existent entry
	err = s.repo.UpdateJournalEntryStatus(s.ctx, uuid.New(), models.StatusPosted)
	s.Error(err)
	_, ok := err.(*app_errors.NotFoundError)
    s.True(ok, "Expected NotFoundError for updating status of non-existent entry")

}

func (s *JournalEntryRepositoryIntegrationTestSuite) TestGetJournalEntriesForTrialBalance() {
	s.T().Log("Running TestGetJournalEntriesForTrialBalance")
	now := time.Now().Truncate(time.Second)
	// Posted entry within range
	entry1 := models.JournalEntry{ EntryDate: now.AddDate(0, 0, -5), Status: models.StatusPosted, Description: "TB Entry 1",
		JournalLines: []models.JournalLine{
			{AccountID: s.cashAccount.ID, Amount: 100, IsDebit: true, ChartOfAccount: s.cashAccount}, // Ensure ChartOfAccount is pre-filled for test simplicity
			{AccountID: s.revenueAccount.ID, Amount: 100, IsDebit: false, ChartOfAccount: s.revenueAccount},
		}}
	// Draft entry within range (should be ignored)
	entry2 := models.JournalEntry{ EntryDate: now.AddDate(0, 0, -4), Status: models.StatusDraft, Description: "TB Entry 2 Draft",
		JournalLines: []models.JournalLine{{AccountID: s.cashAccount.ID, Amount: 50, IsDebit: true, ChartOfAccount: s.cashAccount}}}
	// Posted entry outside range (before)
	entry3 := models.JournalEntry{ EntryDate: now.AddDate(0, 0, -15), Status: models.StatusPosted, Description: "TB Entry 3 Old",
		JournalLines: []models.JournalLine{{AccountID: s.cashAccount.ID, Amount: 20, IsDebit: true, ChartOfAccount: s.cashAccount}}}
	// Posted entry outside range (after)
	entry4 := models.JournalEntry{ EntryDate: now.AddDate(0, 0, 1), Status: models.StatusPosted, Description: "TB Entry 4 Future",
		JournalLines: []models.JournalLine{{AccountID: s.cashAccount.ID, Amount: 20, IsDebit: true, ChartOfAccount: s.cashAccount}}}


	_, err := s.repo.Create(s.ctx, &entry1); s.Require().NoError(err)
	_, err = s.repo.Create(s.ctx, &entry2); s.Require().NoError(err)
	_, err = s.repo.Create(s.ctx, &entry3); s.Require().NoError(err)
	_, err = s.repo.Create(s.ctx, &entry4); s.Require().NoError(err)

	startDate := now.AddDate(0,0,-10)
	endDate := now.AddDate(0,0,-3) // Should only pick up entry1

	entries, err := s.repo.GetJournalEntriesForTrialBalance(s.ctx, startDate, endDate)
	s.NoError(err)
	s.Len(entries, 1, "Should only fetch one posted entry in the date range")
	if len(entries) == 1 {
		s.Equal("TB Entry 1", entries[0].Description)
		s.Len(entries[0].JournalLines, 2, "TB Entry 1 should have its lines preloaded")
		// Check if ChartOfAccount is preloaded on lines (it should be by the repo method)
		s.NotNil(entries[0].JournalLines[0].ChartOfAccount, "ChartOfAccount on line should be preloaded")
		s.Equal(s.cashAccount.ID, entries[0].JournalLines[0].ChartOfAccount.ID)
	}
}
