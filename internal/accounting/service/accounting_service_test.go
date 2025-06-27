package service_test

import (
	"context"
	"erp-system/internal/accounting/models"
	"erp-system/internal/accounting/repository/mocks"
	"erp-system/internal/accounting/service"
	dto "erp-system/internal/accounting/service/dto"
	app_errors "erp-system/pkg/errors" // Renamed to avoid conflict with std errors
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAccountingService_CreateChartOfAccount(t *testing.T) {
	mockCoaRepo := mocks.NewChartOfAccountRepositoryMock(t)
	// mockJournalRepo := mocks.NewJournalEntryRepositoryMock(t) // Not used in this specific test but needed for service creation

	accountingService := service.NewAccountingService(mockCoaRepo, nil) // Pass nil if journalRepo not used by this method

	ctx := context.Background()
	req := dto.CreateChartOfAccountRequest{
		AccountCode: "1010",
		AccountName: "Test Asset Account",
		AccountType: models.Asset,
		IsActive:    true,
	}

	t.Run("Success", func(t *testing.T) {
		mockCoaRepo.On("GetByCode", ctx, req.AccountCode).Return(nil, app_errors.NewNotFoundError("chart_of_account_code", req.AccountCode)).Once()
		mockCoaRepo.On("Create", ctx, mock.AnythingOfType("*models.ChartOfAccount")).Run(func(args mock.Arguments) {
			arg := args.Get(1).(*models.ChartOfAccount)
			assert.Equal(t, req.AccountCode, arg.AccountCode)
			assert.Equal(t, req.AccountName, arg.AccountName)
			arg.ID = uuid.New() // Simulate DB generating ID
		}).Return(func(ctx context.Context, coa *models.ChartOfAccount) *models.ChartOfAccount {
			return coa
		}, nil).Once()

		account, err := accountingService.CreateChartOfAccount(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, account)
		assert.Equal(t, req.AccountCode, account.AccountCode)
		mockCoaRepo.AssertExpectations(t)
	})

	t.Run("Validation Error - Missing Fields", func(t *testing.T) {
		invalidReq := dto.CreateChartOfAccountRequest{AccountCode: "1020"} // Missing name and type
		_, err := accountingService.CreateChartOfAccount(ctx, invalidReq)
		assert.Error(t, err)
		assert.IsType(t, &app_errors.ValidationError{}, err)
	})

	t.Run("Validation Error - Invalid Account Type", func(t *testing.T) {
		invalidReq := dto.CreateChartOfAccountRequest{
			AccountCode: "1021", AccountName: "Invalid Type Acc", AccountType: "INVALID_TYPE",
		}
		_, err := accountingService.CreateChartOfAccount(ctx, invalidReq)
		assert.Error(t, err)
		assert.IsType(t, &app_errors.ValidationError{}, err)
		assert.Contains(t, err.Error(), "invalid account type")
	})

	t.Run("Conflict Error - Account Code Exists", func(t *testing.T) {
		existingAccount := &models.ChartOfAccount{ID: uuid.New(), AccountCode: req.AccountCode}
		mockCoaRepo.On("GetByCode", ctx, req.AccountCode).Return(existingAccount, nil).Once()

		_, err := accountingService.CreateChartOfAccount(ctx, req)

		assert.Error(t, err)
		assert.IsType(t, &app_errors.ConflictError{}, err)
		mockCoaRepo.AssertExpectations(t)
	})

	t.Run("Error - Parent Account Not Found", func(t *testing.T) {
		parentID := uuid.New()
		reqWithParent := dto.CreateChartOfAccountRequest{
			AccountCode:     "1030",
			AccountName:     "Child Account",
			AccountType:     models.Asset,
			IsActive:        true,
			ParentAccountID: &parentID,
		}
		mockCoaRepo.On("GetByCode", ctx, reqWithParent.AccountCode).Return(nil, app_errors.NewNotFoundError("chart_of_account_code", reqWithParent.AccountCode)).Once()
		mockCoaRepo.On("GetByID", ctx, parentID).Return(nil, app_errors.NewNotFoundError("chart_of_account", parentID.String())).Once()

		_, err := accountingService.CreateChartOfAccount(ctx, reqWithParent)
		assert.Error(t, err)
		assert.IsType(t, &app_errors.ValidationError{}, err) // Service wraps it as validation error
		assert.Contains(t, err.Error(), "parent account not found")
		mockCoaRepo.AssertExpectations(t)
	})

	t.Run("Error - Parent Account Not Active", func(t *testing.T) {
		parentID := uuid.New()
		parentAccount := &models.ChartOfAccount{ID: parentID, AccountCode: "P100", IsActive: false}
		reqWithInactiveParent := dto.CreateChartOfAccountRequest{
			AccountCode:     "1031",
			AccountName:     "Child Account Inactive Parent",
			AccountType:     models.Asset,
			IsActive:        true,
			ParentAccountID: &parentID,
		}
		mockCoaRepo.On("GetByCode", ctx, reqWithInactiveParent.AccountCode).Return(nil, app_errors.NewNotFoundError("chart_of_account_code", reqWithInactiveParent.AccountCode)).Once()
		mockCoaRepo.On("GetByID", ctx, parentID).Return(parentAccount, nil).Once()

		_, err := accountingService.CreateChartOfAccount(ctx, reqWithInactiveParent)
		assert.Error(t, err)
		assert.IsType(t, &app_errors.ValidationError{}, err)
		assert.Contains(t, err.Error(), "parent account is not active")
		mockCoaRepo.AssertExpectations(t)
	})

    t.Run("Repository Create Fails", func(t *testing.T) {
        mockCoaRepo.On("GetByCode", ctx, req.AccountCode).Return(nil, app_errors.NewNotFoundError("chart_of_account_code", req.AccountCode)).Once()
        mockCoaRepo.On("Create", ctx, mock.AnythingOfType("*models.ChartOfAccount")).Return(nil, fmt.Errorf("database error")).Once()

        _, err := accountingService.CreateChartOfAccount(ctx, req)
        assert.Error(t, err)
        assert.EqualError(t, err, "database error") // Propagates direct error
        mockCoaRepo.AssertExpectations(t)
    })
}

func TestAccountingService_UpdateChartOfAccount(t *testing.T) {
    mockCoaRepo := mocks.NewChartOfAccountRepositoryMock(t)
    accountingService := service.NewAccountingService(mockCoaRepo, nil)
    ctx := context.Background()

    accountID := uuid.New()
    originalAccount := &models.ChartOfAccount{
        ID:          accountID,
        AccountCode: "UA01",
        AccountName: "Original Name",
        AccountType: models.Asset,
        IsActive:    true,
    }

    t.Run("Success - Update Name", func(t *testing.T) {
        newName := "Updated Name"
        req := dto.UpdateChartOfAccountRequest{AccountName: &newName}

        mockCoaRepo.On("GetByID", ctx, accountID).Return(originalAccount, nil).Once()
        mockCoaRepo.On("Update", ctx, mock.MatchedBy(func(acc *models.ChartOfAccount) bool {
            return acc.ID == accountID && acc.AccountName == newName
        })).Return(func(ctx context.Context, acc *models.ChartOfAccount) *models.ChartOfAccount {
            return acc // Return the modified account
        }, nil).Once()

        updatedAccount, err := accountingService.UpdateChartOfAccount(ctx, accountID, req)
        assert.NoError(t, err)
        assert.NotNil(t, updatedAccount)
        assert.Equal(t, newName, updatedAccount.AccountName)
        mockCoaRepo.AssertExpectations(t)
    })

    t.Run("Error - Account Not Found", func(t *testing.T) {
        req := dto.UpdateChartOfAccountRequest{}
        mockCoaRepo.On("GetByID", ctx, accountID).Return(nil, app_errors.NewNotFoundError("coa", accountID.String())).Once()

        _, err := accountingService.UpdateChartOfAccount(ctx, accountID, req)
        assert.Error(t, err)
        assert.IsType(t, &app_errors.NotFoundError{}, err)
        mockCoaRepo.AssertExpectations(t)
    })

    t.Run("Error - Invalid Parent Account ID during update", func(t *testing.T) {
        newParentID := uuid.New()
        req := dto.UpdateChartOfAccountRequest{ParentAccountID: &newParentID}

        mockCoaRepo.On("GetByID", ctx, accountID).Return(originalAccount, nil).Once()
        mockCoaRepo.On("GetByID", ctx, newParentID).Return(nil, app_errors.NewNotFoundError("coa", newParentID.String())).Once() // Mocking parent not found

        _, err := accountingService.UpdateChartOfAccount(ctx, accountID, req)
        assert.Error(t, err)
        assert.IsType(t, &app_errors.ValidationError{}, err)
		assert.Contains(t, err.Error(), "parent account not found")
        mockCoaRepo.AssertExpectations(t)
    })
}


func TestAccountingService_CreateJournalEntry(t *testing.T) {
	mockCoaRepo := mocks.NewChartOfAccountRepositoryMock(t)
	mockJournalRepo := mocks.NewJournalEntryRepositoryMock(t)
	accountingService := service.NewAccountingService(mockCoaRepo, mockJournalRepo)
	ctx := context.Background()

	cashAccountID := uuid.New()
	revenueAccountID := uuid.New()

	cashAccount := &models.ChartOfAccount{ID: cashAccountID, AccountCode: "1000", AccountName: "Cash", AccountType: models.Asset, IsActive: true}
	revenueAccount := &models.ChartOfAccount{ID: revenueAccountID, AccountCode: "4000", AccountName: "Revenue", AccountType: models.Revenue, IsActive: true}

	req := dto.CreateJournalEntryRequest{
		EntryDate:   time.Now(),
		Description: "Test Sale",
		Lines: []dto.JournalLineRequest{
			{AccountID: cashAccountID, Amount: 100.00, IsDebit: true, Currency: "USD"},
			{AccountID: revenueAccountID, Amount: 100.00, IsDebit: false, Currency: "USD"},
		},
	}

	t.Run("Success", func(t *testing.T) {
		mockCoaRepo.On("GetByID", ctx, cashAccountID).Return(cashAccount, nil).Once()
		mockCoaRepo.On("GetByID", ctx, revenueAccountID).Return(revenueAccount, nil).Once()

		mockJournalRepo.On("Create", ctx, mock.AnythingOfType("*models.JournalEntry")).Run(func(args mock.Arguments) {
			arg := args.Get(1).(*models.JournalEntry)
			assert.Len(t, arg.JournalLines, 2)
			assert.True(t, arg.IsBalanced()) // Assuming IsBalanced is a helper on model
			arg.ID = uuid.New()              // Simulate DB generating ID
		}).Return(func(ctx context.Context, je *models.JournalEntry) *models.JournalEntry {
			// Make sure lines also get IDs if BeforeCreate hook in model handles it
			for i := range je.JournalLines {
				if je.JournalLines[i].ID == uuid.Nil {
					je.JournalLines[i].ID = uuid.New()
				}
			}
			return je
		}, nil).Once()

		entry, err := accountingService.CreateJournalEntry(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, entry)
		assert.Equal(t, models.StatusDraft, entry.Status) // Default status
		assert.Len(t, entry.JournalLines, 2)
		mockCoaRepo.AssertExpectations(t)
		mockJournalRepo.AssertExpectations(t)
	})

	t.Run("Error - Unbalanced Entry", func(t *testing.T) {
		unbalancedReq := dto.CreateJournalEntryRequest{
			Description: "Unbalanced",
			Lines: []dto.JournalLineRequest{
				{AccountID: cashAccountID, Amount: 100.00, IsDebit: true},
				{AccountID: revenueAccountID, Amount: 90.00, IsDebit: false}, // Unbalanced
			},
		}
		// Mock GetByID for accounts used in lines
		mockCoaRepo.On("GetByID", ctx, cashAccountID).Return(cashAccount, nil).Once()
		mockCoaRepo.On("GetByID", ctx, revenueAccountID).Return(revenueAccount, nil).Once()

		_, err := accountingService.CreateJournalEntry(ctx, unbalancedReq)
		assert.Error(t, err)
		assert.IsType(t, &app_errors.ValidationError{}, err)
		assert.Contains(t, err.Error(), "debits (100.00) must equal credits (90.00)")
		mockCoaRepo.AssertExpectations(t)
		// No call to journalRepo.Create should happen
		mockJournalRepo.AssertNotCalled(t, "Create", ctx, mock.Anything)
	})

    t.Run("Error - Line Account Not Found", func(t *testing.T) {
        nonExistentAccountID := uuid.New()
        reqWithInvalidAccount := dto.CreateJournalEntryRequest{
            Description: "Invalid Account",
            Lines: []dto.JournalLineRequest{
                {AccountID: nonExistentAccountID, Amount: 50.00, IsDebit: true},
                {AccountID: cashAccountID, Amount: 50.00, IsDebit: false},
            },
        }
        mockCoaRepo.On("GetByID", ctx, nonExistentAccountID).Return(nil, app_errors.NewNotFoundError("coa", nonExistentAccountID.String())).Once()
        // mockCoaRepo.On("GetByID", ctx, cashAccountID).Return(cashAccount, nil) // This might or might not be called depending on loop order

        _, err := accountingService.CreateJournalEntry(ctx, reqWithInvalidAccount)
        assert.Error(t, err)
        assert.IsType(t, &app_errors.ValidationError{}, err)
        assert.Contains(t, err.Error(), "not found")
        mockCoaRepo.AssertExpectations(t) // Ensure GetByID for nonExistentAccountID was called
    })


    t.Run("Error - Line Account Not Active", func(t *testing.T) {
        inactiveAccountID := uuid.New()
        inactiveAccount := &models.ChartOfAccount{ID: inactiveAccountID, AccountCode: "INAC", IsActive: false}
        reqWithInactiveAccount := dto.CreateJournalEntryRequest{
            Description: "Inactive Account",
            Lines: []dto.JournalLineRequest{
                {AccountID: inactiveAccountID, Amount: 70.00, IsDebit: true},
                {AccountID: cashAccountID, Amount: 70.00, IsDebit: false},
            },
        }
        mockCoaRepo.On("GetByID", ctx, inactiveAccountID).Return(inactiveAccount, nil).Once()
        // mockCoaRepo.On("GetByID", ctx, cashAccountID).Return(cashAccount, nil) // Might not be called if first account fails

        _, err := accountingService.CreateJournalEntry(ctx, reqWithInactiveAccount)
        assert.Error(t, err)
        assert.IsType(t, &app_errors.ValidationError{}, err)
        assert.Contains(t, err.Error(), "is not active")
        mockCoaRepo.AssertExpectations(t)
    })


    t.Run("Error - Invalid Line Amount (<=0)", func(t *testing.T) {
        reqWithInvalidAmount := dto.CreateJournalEntryRequest{
            Description: "Invalid Amount",
            Lines: []dto.JournalLineRequest{
                {AccountID: cashAccountID, Amount: -10.00, IsDebit: true}, // Invalid amount
                {AccountID: revenueAccountID, Amount: -10.00, IsDebit: false},
            },
        }
        // No need to mock GetByID if validation fails before that
        _, err := accountingService.CreateJournalEntry(ctx, reqWithInvalidAmount)
        assert.Error(t, err)
        assert.IsType(t, &app_errors.ValidationError{}, err)
        assert.Contains(t, err.Error(), "amount must be positive")
    })

    t.Run("Repository Create Fails for Journal Entry", func(t *testing.T) {
        mockCoaRepo.On("GetByID", ctx, cashAccountID).Return(cashAccount, nil).Once()
        mockCoaRepo.On("GetByID", ctx, revenueAccountID).Return(revenueAccount, nil).Once()
        mockJournalRepo.On("Create", ctx, mock.AnythingOfType("*models.JournalEntry")).Return(nil, fmt.Errorf("db create error")).Once()

        _, err := accountingService.CreateJournalEntry(ctx, req)
        assert.Error(t, err)
        assert.EqualError(t, err, "db create error")
        mockCoaRepo.AssertExpectations(t)
        mockJournalRepo.AssertExpectations(t)
    })
}

func TestAccountingService_PostJournalEntry(t *testing.T) {
    mockCoaRepo := mocks.NewChartOfAccountRepositoryMock(t)
    mockJournalRepo := mocks.NewJournalEntryRepositoryMock(t)
    accountingService := service.NewAccountingService(mockCoaRepo, mockJournalRepo)
    ctx := context.Background()

    entryID := uuid.New()
    cashAccountID := uuid.New()
    revenueAccountID := uuid.New()

    draftEntry := &models.JournalEntry{
        ID:     entryID,
        Status: models.StatusDraft,
        JournalLines: []models.JournalLine{
            {AccountID: cashAccountID, Amount: 100, IsDebit: true, ChartOfAccount: &models.ChartOfAccount{ID: cashAccountID, IsActive: true}},
            {AccountID: revenueAccountID, Amount: 100, IsDebit: false, ChartOfAccount: &models.ChartOfAccount{ID: revenueAccountID, IsActive: true}},
        },
    }
    // Pre-populate ChartOfAccount in lines for IsBalanced and account active checks
    // In real scenario, GetByID would populate this.
    // For test, ensure the mock GetByID for entry returns it with lines that have ChartOfAccount.

    t.Run("Success - Post Draft Entry", func(t *testing.T) {
        // Mock GetByID for the entry itself
        mockJournalRepo.On("GetByID", ctx, entryID).Return(draftEntry, nil).Once()

        // Mock GetByID for accounts in lines (for active check)
        mockCoaRepo.On("GetByID", ctx, cashAccountID).Return(&models.ChartOfAccount{ID: cashAccountID, IsActive: true}, nil).Once()
        mockCoaRepo.On("GetByID", ctx, revenueAccountID).Return(&models.ChartOfAccount{ID: revenueAccountID, IsActive: true}, nil).Once()

        // Mock UpdateJournalEntryStatus
        mockJournalRepo.On("UpdateJournalEntryStatus", ctx, entryID, models.StatusPosted).Return(nil).Once()

        // Mock GetByID again for reloading the entry after status update
        postedEntry := *draftEntry // copy
        postedEntry.Status = models.StatusPosted
        mockJournalRepo.On("GetByID", ctx, entryID).Return(&postedEntry, nil).Once()


        entry, err := accountingService.PostJournalEntry(ctx, entryID)
        assert.NoError(t, err)
        assert.NotNil(t, entry)
        assert.Equal(t, models.StatusPosted, entry.Status)
        mockJournalRepo.AssertExpectations(t)
        mockCoaRepo.AssertExpectations(t)
    })

    t.Run("Error - Entry Already Posted", func(t *testing.T) {
        alreadyPostedEntry := &models.JournalEntry{ID: entryID, Status: models.StatusPosted}
        mockJournalRepo.On("GetByID", ctx, entryID).Return(alreadyPostedEntry, nil).Once()

        entry, err := accountingService.PostJournalEntry(ctx, entryID)
        assert.NoError(t, err) // No error, just returns existing entry
        assert.Equal(t, models.StatusPosted, entry.Status)
        mockJournalRepo.AssertExpectations(t)
    })

    t.Run("Error - Cannot Post Voided Entry", func(t *testing.T) {
        voidedEntry := &models.JournalEntry{ID: entryID, Status: models.StatusVoided}
        mockJournalRepo.On("GetByID", ctx, entryID).Return(voidedEntry, nil).Once()

        _, err := accountingService.PostJournalEntry(ctx, entryID)
        assert.Error(t, err)
        assert.IsType(t, &app_errors.ConflictError{}, err)
        mockJournalRepo.AssertExpectations(t)
    })

    t.Run("Error - Entry Not Balanced", func(t *testing.T) {
        unbalancedEntry := &models.JournalEntry{
            ID:     entryID,
            Status: models.StatusDraft,
            JournalLines: []models.JournalLine{
                {AccountID: cashAccountID, Amount: 100, IsDebit: true},
                {AccountID: revenueAccountID, Amount: 90, IsDebit: false},
            },
        }
        mockJournalRepo.On("GetByID", ctx, entryID).Return(unbalancedEntry, nil).Once()

        _, err := accountingService.PostJournalEntry(ctx, entryID)
        assert.Error(t, err)
        assert.IsType(t, &app_errors.ValidationError{}, err)
        assert.Contains(t, err.Error(), "not balanced")
        mockJournalRepo.AssertExpectations(t)
    })

    t.Run("Error - Inactive Account in Line", func(t *testing.T) {
        inactiveCashAccount := &models.ChartOfAccount{ID: cashAccountID, IsActive: false} // Inactive
        entryWithInactiveAccountLine := &models.JournalEntry{
            ID:     entryID,
            Status: models.StatusDraft,
            JournalLines: []models.JournalLine{
                {AccountID: cashAccountID, Amount: 100, IsDebit: true, ChartOfAccount: inactiveCashAccount},
                {AccountID: revenueAccountID, Amount: 100, IsDebit: false, ChartOfAccount: &models.ChartOfAccount{ID: revenueAccountID, IsActive: true}},
            },
        }
        mockJournalRepo.On("GetByID", ctx, entryID).Return(entryWithInactiveAccountLine, nil).Once()
        mockCoaRepo.On("GetByID", ctx, cashAccountID).Return(inactiveCashAccount, nil).Once()
        // mockCoaRepo.On("GetByID", ctx, revenueAccountID).Return(&models.ChartOfAccount{ID: revenueAccountID, IsActive: true}, nil).Once() // May not be called if first fails

        _, err := accountingService.PostJournalEntry(ctx, entryID)
        assert.Error(t, err)
        assert.IsType(t, &app_errors.ConflictError{}, err)
        assert.Contains(t, err.Error(), "is inactive")
        mockJournalRepo.AssertExpectations(t)
        mockCoaRepo.AssertExpectations(t)
    })
}

func TestAccountingService_GetTrialBalance(t *testing.T) {
    mockCoaRepo := mocks.NewChartOfAccountRepositoryMock(t)
    mockJournalRepo := mocks.NewJournalEntryRepositoryMock(t)
    accountingService := service.NewAccountingService(mockCoaRepo, mockJournalRepo)
    ctx := context.Background()

    endDate := time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC)
    startDate := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC) // Default start for cumulative

    // Accounts
    cashAccID, arAccID, apAccID, revAccID, expAccID := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
    allActiveAccounts := []*models.ChartOfAccount{
        {ID: cashAccID, AccountCode: "1010", AccountName: "Cash", AccountType: models.Asset, IsActive: true},
        {ID: arAccID, AccountCode: "1020", AccountName: "Accounts Receivable", AccountType: models.Asset, IsActive: true},
        {ID: apAccID, AccountCode: "2010", AccountName: "Accounts Payable", AccountType: models.Liability, IsActive: true},
        {ID: revAccID, AccountCode: "4010", AccountName: "Service Revenue", AccountType: models.Revenue, IsActive: true},
        {ID: expAccID, AccountCode: "5010", AccountName: "Rent Expense", AccountType: models.Expense, IsActive: true},
    }

    // Journal Entries for Trial Balance
    journalEntries := []models.JournalEntry{
        { // Cash Sale
            ID: uuid.New(), Status: models.StatusPosted, EntryDate: endDate.AddDate(0,0,-10),
            JournalLines: []models.JournalLine{
                {AccountID: cashAccID, Amount: 1000, IsDebit: true, ChartOfAccount: allActiveAccounts[0]},
                {AccountID: revAccID, Amount: 1000, IsDebit: false, ChartOfAccount: allActiveAccounts[3]},
            },
        },
        { // Paid Rent
            ID: uuid.New(), Status: models.StatusPosted, EntryDate: endDate.AddDate(0,0,-5),
            JournalLines: []models.JournalLine{
                {AccountID: expAccID, Amount: 200, IsDebit: true, ChartOfAccount: allActiveAccounts[4]},
                {AccountID: cashAccID, Amount: 200, IsDebit: false, ChartOfAccount: allActiveAccounts[0]},
            },
        },
    }

    t.Run("Success - Basic Trial Balance", func(t *testing.T) {
        req := dto.TrialBalanceRequest{EndDate: endDate, IncludeZeroBalanceAccounts: true}

        mockJournalRepo.On("GetJournalEntriesForTrialBalance", ctx, startDate, endDate).Return(journalEntries, nil).Once()
        mockCoaRepo.On("List", ctx, 0, 0, map[string]interface{}{"is_active": true}).Return(allActiveAccounts, int64(len(allActiveAccounts)), nil).Once()

        tb, err := accountingService.GetTrialBalance(ctx, req)

        assert.NoError(t, err)
        assert.NotNil(t, tb)
        assert.Len(t, tb.Lines, 5) // All active accounts
        assert.Equal(t, tb.TotalDebits, tb.TotalCredits)
        assert.InDelta(t, 1200.00, tb.TotalDebits, 0.001) // Cash 800 DR, AR 0, AP 0, Revenue 1000 CR, Expense 200 DR => 1000 DR, 1000 CR

        // Check specific account balances (Cash: 1000 DR - 200 CR = 800 DR)
        foundCash := false
        for _, line := range tb.Lines {
            if line.AccountCode == "1010" { // Cash
                assert.InDelta(t, 800.00, line.Debit, 0.001)
                assert.InDelta(t, 0.00, line.Credit, 0.001)
                foundCash = true
            }
        }
        assert.True(t, foundCash, "Cash account not found in trial balance")

        mockJournalRepo.AssertExpectations(t)
        mockCoaRepo.AssertExpectations(t)
    })

    t.Run("Success - Trial Balance without zero balance accounts", func(t *testing.T) {
        req := dto.TrialBalanceRequest{EndDate: endDate, IncludeZeroBalanceAccounts: false}

        mockJournalRepo.On("GetJournalEntriesForTrialBalance", ctx, startDate, endDate).Return(journalEntries, nil).Once()
        mockCoaRepo.On("List", ctx, 0, 0, map[string]interface{}{"is_active": true}).Return(allActiveAccounts, int64(len(allActiveAccounts)), nil).Once()

        tb, err := accountingService.GetTrialBalance(ctx, req)
        assert.NoError(t, err)
        assert.NotNil(t, tb)
        // Cash (800 DR), Revenue (1000 CR), Expense (200 DR) have balances. AR and AP are zero.
        assert.Len(t, tb.Lines, 3)
        assert.Equal(t, tb.TotalDebits, tb.TotalCredits)
        assert.InDelta(t, 1000.00, tb.TotalDebits, 0.001) // Cash 800 DR, Expense 200 DR = 1000 DR. Revenue 1000 CR.

        mockJournalRepo.AssertExpectations(t)
        mockCoaRepo.AssertExpectations(t)
    })


    t.Run("Error - Fetching Journal Entries Fails", func(t *testing.T) {
        req := dto.TrialBalanceRequest{EndDate: endDate}
        mockJournalRepo.On("GetJournalEntriesForTrialBalance", ctx, startDate, endDate).Return(nil, fmt.Errorf("db error")).Once()
        // mockCoaRepo.On("List", ctx, 0, 0, mock.Anything).Return(allActiveAccounts, int64(len(allActiveAccounts)), nil) // Might not be called

        _, err := accountingService.GetTrialBalance(ctx, req)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "db error")
        mockJournalRepo.AssertExpectations(t)
    })

    t.Run("Error - Fetching All Accounts Fails", func(t *testing.T) {
        req := dto.TrialBalanceRequest{EndDate: endDate}
        mockJournalRepo.On("GetJournalEntriesForTrialBalance", ctx, startDate, endDate).Return(journalEntries, nil).Once()
        mockCoaRepo.On("List", ctx, 0, 0, map[string]interface{}{"is_active": true}).Return(nil, int64(0), fmt.Errorf("coa list error")).Once()

        _, err := accountingService.GetTrialBalance(ctx, req)
        assert.Error(t, err)
        assert.IsType(t, &app_errors.InternalServerError{}, err)
        assert.Contains(t, err.Error(), "failed to fetch accounts for trial balance")
        mockJournalRepo.AssertExpectations(t)
        mockCoaRepo.AssertExpectations(t)
    })
}


// Add more tests for GetChartOfAccountByID, DeleteChartOfAccount, ListChartOfAccounts,
// GetJournalEntryByID, UpdateJournalEntry, DeleteJournalEntry, ListJournalEntries etc.
// following similar patterns. Remember to reset mocks for each sub-test if necessary or use Once().

func TestAccountingService_GetChartOfAccountByID(t *testing.T) {
    mockCoaRepo := mocks.NewChartOfAccountRepositoryMock(t)
    s := service.NewAccountingService(mockCoaRepo, nil)
    ctx := context.Background()
    testID := uuid.New()

    t.Run("Success", func(t *testing.T) {
        expectedAccount := &models.ChartOfAccount{ID: testID, AccountName: "Test"}
        mockCoaRepo.On("GetByID", ctx, testID).Return(expectedAccount, nil).Once()
        acc, err := s.GetChartOfAccountByID(ctx, testID)
        assert.NoError(t, err)
        assert.Equal(t, expectedAccount, acc)
        mockCoaRepo.AssertExpectations(t)
    })

    t.Run("Not Found", func(t *testing.T) {
        mockCoaRepo.On("GetByID", ctx, testID).Return(nil, app_errors.NewNotFoundError("coa", testID.String())).Once()
        _, err := s.GetChartOfAccountByID(ctx, testID)
        assert.Error(t, err)
        assert.IsType(t, &app_errors.NotFoundError{}, err)
        mockCoaRepo.AssertExpectations(t)
    })
}

func TestAccountingService_DeleteJournalEntry(t *testing.T) {
    mockJournalRepo := mocks.NewJournalEntryRepositoryMock(t)
    s := service.NewAccountingService(nil, mockJournalRepo)
    ctx := context.Background()
    entryID := uuid.New()

    t.Run("Success - Delete Draft Entry", func(t *testing.T) {
        draftEntry := &models.JournalEntry{ID: entryID, Status: models.StatusDraft}
        mockJournalRepo.On("GetByID", ctx, entryID).Return(draftEntry, nil).Once()
        mockJournalRepo.On("Delete", ctx, entryID).Return(nil).Once()

        err := s.DeleteJournalEntry(ctx, entryID)
        assert.NoError(t, err)
        mockJournalRepo.AssertExpectations(t)
    })

    t.Run("Error - Cannot Delete Posted Entry", func(t *testing.T) {
        postedEntry := &models.JournalEntry{ID: entryID, Status: models.StatusPosted}
        mockJournalRepo.On("GetByID", ctx, entryID).Return(postedEntry, nil).Once()

        err := s.DeleteJournalEntry(ctx, entryID)
        assert.Error(t, err)
        assert.IsType(t, &app_errors.ConflictError{}, err)
        mockJournalRepo.AssertExpectations(t) // GetByID was called
        mockJournalRepo.AssertNotCalled(t, "Delete", ctx, entryID) // Delete should not be called
    })

    t.Run("Error - Entry Not Found for Deletion", func(t *testing.T) {
        mockJournalRepo.On("GetByID", ctx, entryID).Return(nil, app_errors.NewNotFoundError("je", entryID.String())).Once()
        err := s.DeleteJournalEntry(ctx, entryID)
        assert.Error(t, err)
        assert.IsType(t, &app_errors.NotFoundError{}, err)
        mockJournalRepo.AssertExpectations(t)
    })
}
