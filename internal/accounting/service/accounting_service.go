package service

import (
	"context"
	"erp-system/internal/accounting/models"
	"erp-system/internal/accounting/repository"
	dto "erp-system/internal/accounting/service/dto" // Alias for DTOs
	"erp-system/pkg/errors"
	"erp-system/pkg/logger"
	"fmt"
	"math" // For float comparisons with tolerance
	"time"

	"github.com/google/uuid"
)

// AccountingService defines the interface for accounting-related business logic.
type AccountingService interface {
	// Chart of Accounts
	CreateChartOfAccount(ctx context.Context, req dto.CreateChartOfAccountRequest) (*models.ChartOfAccount, error)
	GetChartOfAccountByID(ctx context.Context, id uuid.UUID) (*models.ChartOfAccount, error)
	GetChartOfAccountByCode(ctx context.Context, code string) (*models.ChartOfAccount, error)
	UpdateChartOfAccount(ctx context.Context, id uuid.UUID, req dto.UpdateChartOfAccountRequest) (*models.ChartOfAccount, error)
	DeleteChartOfAccount(ctx context.Context, id uuid.UUID) error
	ListChartOfAccounts(ctx context.Context, req dto.ListChartOfAccountsRequest) ([]*models.ChartOfAccount, int64, error)

	// Journal Entries
	CreateJournalEntry(ctx context.Context, req dto.CreateJournalEntryRequest) (*models.JournalEntry, error)
	GetJournalEntryByID(ctx context.Context, id uuid.UUID) (*models.JournalEntry, error)
	UpdateJournalEntry(ctx context.Context, id uuid.UUID, req dto.UpdateJournalEntryRequest) (*models.JournalEntry, error)
	DeleteJournalEntry(ctx context.Context, id uuid.UUID) error
	ListJournalEntries(ctx context.Context, req dto.ListJournalEntriesRequest) ([]*models.JournalEntry, int64, error)
	PostJournalEntry(ctx context.Context, id uuid.UUID) (*models.JournalEntry, error)

	// Reporting
	GetTrialBalance(ctx context.Context, req dto.TrialBalanceRequest) (*dto.TrialBalanceResponse, error)
	// GetBalanceSheet(ctx context.Context, date time.Time) (*dto.BalanceSheetResponse, error)
	// GetProfitAndLossStatement(ctx context.Context, startDate, endDate time.Time) (*dto.ProfitAndLossResponse, error)

	// Other specific methods
	GetAccountBalance(ctx context.Context, accountID uuid.UUID, date time.Time) (float64, error)
}

// accountingService is an implementation of AccountingService.
type accountingService struct {
	coaRepo     repository.ChartOfAccountRepository
	journalRepo repository.JournalEntryRepository
	// Potentially other repositories if needed
}

// NewAccountingService creates a new AccountingService.
func NewAccountingService(
	coaRepo repository.ChartOfAccountRepository,
	journalRepo repository.JournalEntryRepository,
) AccountingService {
	return &accountingService{
		coaRepo:     coaRepo,
		journalRepo: journalRepo,
	}
}

// --- Chart of Accounts Methods ---

func (s *accountingService) CreateChartOfAccount(ctx context.Context, req dto.CreateChartOfAccountRequest) (*models.ChartOfAccount, error) {
	logger.InfoLogger.Printf("Service: Attempting to create chart of account with code: %s", req.AccountCode)

	// Validate request
	if req.AccountCode == "" || req.AccountName == "" || req.AccountType == "" {
		logger.WarnLogger.Println("Service: Validation failed for creating chart of account - missing required fields.")
		return nil, errors.NewValidationError("AccountCode, AccountName, and AccountType are required", "")
	}
	// Validate AccountType enum
	validAccountType := false
	for _, at := range []models.AccountType{models.Asset, models.Liability, models.Equity, models.Revenue, models.Expense} {
		if req.AccountType == at {
			validAccountType = true
			break
		}
	}
	if !validAccountType {
		logger.WarnLogger.Printf("Service: Invalid account type provided: %s", req.AccountType)
		return nil, errors.NewValidationError(fmt.Sprintf("invalid account type: %s", req.AccountType), "account_type")
	}

	// Check if account code already exists
	existing, err := s.coaRepo.GetByCode(ctx, req.AccountCode)
	if err != nil && !isNotFoundError(err) { // isNotFoundError checks if it's our custom NotFoundError
		logger.ErrorLogger.Printf("Service: Error checking existing account code %s: %v", req.AccountCode, err)
		return nil, err // Propagate error (should be internal server error from repo)
	}
	if existing != nil {
		logger.WarnLogger.Printf("Service: Account code %s already exists.", req.AccountCode)
		return nil, errors.NewConflictError(fmt.Sprintf("account with code %s already exists", req.AccountCode))
	}

	// Check parent account if provided
	if req.ParentAccountID != nil && *req.ParentAccountID != uuid.Nil {
		parentAccount, err := s.coaRepo.GetByID(ctx, *req.ParentAccountID)
		if err != nil {
			if isNotFoundError(err) {
				logger.WarnLogger.Printf("Service: Parent account with ID %s not found.", *req.ParentAccountID)
				return nil, errors.NewValidationError("parent account not found", "parent_account_id")
			}
			logger.ErrorLogger.Printf("Service: Error fetching parent account %s: %v", *req.ParentAccountID, err)
			return nil, err // Propagate internal server error
		}
		if !parentAccount.IsActive {
			logger.WarnLogger.Printf("Service: Parent account %s is not active.", parentAccount.AccountCode)
			return nil, errors.NewValidationError("parent account is not active", "parent_account_id")
		}
	}

	account := &models.ChartOfAccount{
		AccountCode:     req.AccountCode,
		AccountName:     req.AccountName,
		AccountType:     req.AccountType,
		ParentAccountID: req.ParentAccountID,
		IsActive:        req.IsActive, // Default true if not provided by DTO (DTO should have default)
	}

	createdAccount, err := s.coaRepo.Create(ctx, account)
	if err != nil {
		logger.ErrorLogger.Printf("Service: Error creating chart of account in repository: %v", err)
		return nil, err // Propagate error
	}
	logger.InfoLogger.Printf("Service: Successfully created chart of account with ID: %s", createdAccount.ID)
	return createdAccount, nil
}

func (s *accountingService) GetChartOfAccountByID(ctx context.Context, id uuid.UUID) (*models.ChartOfAccount, error) {
	logger.InfoLogger.Printf("Service: Attempting to get chart of account by ID: %s", id)
	account, err := s.coaRepo.GetByID(ctx, id)
	if err != nil {
		logger.ErrorLogger.Printf("Service: Error getting chart of account by ID %s from repository: %v", id, err)
		return nil, err // Propagate (could be NotFoundError or InternalServerError)
	}
	logger.InfoLogger.Printf("Service: Successfully retrieved chart of account by ID: %s", id)
	return account, nil
}

func (s *accountingService) GetChartOfAccountByCode(ctx context.Context, code string) (*models.ChartOfAccount, error) {
	logger.InfoLogger.Printf("Service: Attempting to get chart of account by code: %s", code)
	account, err := s.coaRepo.GetByCode(ctx, code)
	if err != nil {
		logger.ErrorLogger.Printf("Service: Error getting chart of account by code %s from repository: %v", code, err)
		return nil, err // Propagate
	}
	logger.InfoLogger.Printf("Service: Successfully retrieved chart of account by code: %s", code)
	return account, nil
}

func (s *accountingService) UpdateChartOfAccount(ctx context.Context, id uuid.UUID, req dto.UpdateChartOfAccountRequest) (*models.ChartOfAccount, error) {
	logger.InfoLogger.Printf("Service: Attempting to update chart of account with ID: %s", id)

	account, err := s.coaRepo.GetByID(ctx, id)
	if err != nil {
		logger.ErrorLogger.Printf("Service: Error finding chart of account %s for update: %v", id, err)
		return nil, err // Propagate (could be NotFoundError)
	}

	// Update fields from request if provided
	if req.AccountName != nil {
		account.AccountName = *req.AccountName
	}
	if req.AccountType != nil {
		// Validate AccountType enum
		validAccountType := false
		for _, at := range []models.AccountType{models.Asset, models.Liability, models.Equity, models.Revenue, models.Expense} {
			if *req.AccountType == at {
				validAccountType = true
				break
			}
		}
		if !validAccountType {
			logger.WarnLogger.Printf("Service: Invalid account type provided for update: %s", *req.AccountType)
			return nil, errors.NewValidationError(fmt.Sprintf("invalid account type: %s", *req.AccountType), "account_type")
		}
		account.AccountType = *req.AccountType
	}
	if req.ParentAccountID != nil { // Allows setting ParentAccountID to nil or a new ID
		if *req.ParentAccountID == uuid.Nil { // Check if trying to set to nil explicitly
			account.ParentAccountID = nil
		} else {
			// If setting to a new parent, validate the new parent
			parentAccount, err := s.coaRepo.GetByID(ctx, *req.ParentAccountID)
			if err != nil {
				if isNotFoundError(err) {
					logger.WarnLogger.Printf("Service: Parent account with ID %s not found for update.", *req.ParentAccountID)
					return nil, errors.NewValidationError("parent account not found", "parent_account_id")
				}
				logger.ErrorLogger.Printf("Service: Error fetching parent account %s for update: %v", *req.ParentAccountID, err)
				return nil, err
			}
			if !parentAccount.IsActive {
				logger.WarnLogger.Printf("Service: Parent account %s is not active for update.", parentAccount.AccountCode)
				return nil, errors.NewValidationError("parent account is not active", "parent_account_id")
			}
			if parentAccount.ID == account.ID { // Prevent self-referencing
				logger.WarnLogger.Printf("Service: Cannot set account %s as its own parent.", account.AccountCode)
				return nil, errors.NewValidationError("cannot set account as its own parent", "parent_account_id")
			}
			account.ParentAccountID = req.ParentAccountID
		}
	}
	if req.IsActive != nil {
		account.IsActive = *req.IsActive
		// Add logic here if deactivating an account has other implications (e.g., if it has children, or is used in posted entries)
		if !(*req.IsActive) {
			// Example check: Cannot deactivate if it's a parent of active accounts (this is complex, may need recursive check)
			// Example check: Cannot deactivate if it has recent posted transactions (define "recent")
			logger.InfoLogger.Printf("Service: Account %s (ID: %s) is being deactivated.", account.AccountCode, account.ID)
		}
	}
	// Note: AccountCode is typically not updatable. If it were, need to check for uniqueness.

	updatedAccount, err := s.coaRepo.Update(ctx, account)
	if err != nil {
		logger.ErrorLogger.Printf("Service: Error updating chart of account %s in repository: %v", id, err)
		return nil, err // Propagate
	}
	logger.InfoLogger.Printf("Service: Successfully updated chart of account with ID: %s", updatedAccount.ID)
	return updatedAccount, nil
}

func (s *accountingService) DeleteChartOfAccount(ctx context.Context, id uuid.UUID) error {
	logger.InfoLogger.Printf("Service: Attempting to delete chart of account with ID: %s", id)

	account, err := s.coaRepo.GetByID(ctx, id)
	if err != nil {
		logger.ErrorLogger.Printf("Service: Error finding chart of account %s for deletion: %v", id, err)
		return err // Propagate (could be NotFoundError)
	}

	// Business logic for deletion:
	// 1. Check if account is part of any posted journal entries.
	//    This requires a method in journal repository: e.g., `IsAccountUsedInPostedEntries(accountID) bool`
	//    For simplicity, this check is omitted here but is crucial in a real system.
	//    If used, deletion might be blocked or require archiving.
	// Example:
	/*
		isUsed, err := s.journalRepo.IsAccountUsed(ctx, id) // Assumes such a method exists
		if err != nil {
			logger.ErrorLogger.Printf("Service: Error checking if account %s is used: %v", id, err)
			return errors.NewInternalServerError("failed to check account usage", err)
		}
		if isUsed {
			logger.WarnLogger.Printf("Service: Cannot delete account %s because it is used in journal entries.", account.AccountCode)
			return errors.NewConflictError(fmt.Sprintf("account %s (%s) cannot be deleted because it is in use", account.AccountName, account.AccountCode))
		}
	*/

	// 2. Check if it's a parent account to any other active accounts.
	//    This would require listing child accounts.
	//    For simplicity, also omitted. A common rule is to reassign child accounts or prevent deletion.

	if err := s.coaRepo.Delete(ctx, id); err != nil {
		logger.ErrorLogger.Printf("Service: Error deleting chart of account %s from repository: %v", id, err)
		return err // Propagate
	}
	logger.InfoLogger.Printf("Service: Successfully deleted chart of account with ID: %s", id)
	return nil
}

func (s *accountingService) ListChartOfAccounts(ctx context.Context, req dto.ListChartOfAccountsRequest) ([]*models.ChartOfAccount, int64, error) {
	logger.InfoLogger.Printf("Service: Listing chart of accounts with filters: %+v", req)
	filters := make(map[string]interface{})
	if req.AccountName != "" {
		filters["account_name"] = req.AccountName
	}
	if req.AccountType != "" {
		filters["account_type"] = req.AccountType
	}
    if req.IsActive != nil { // Check if the pointer is not nil
        filters["is_active"] = *req.IsActive
    }


	offset := 0
	if req.Page > 0 && req.Limit > 0 {
		offset = (req.Page - 1) * req.Limit
	}
	limit := req.Limit
	if limit == 0 { // Default limit or handle no limit case
		limit = 20 // Default to 20 if not specified or 0
	}


	accounts, total, err := s.coaRepo.List(ctx, offset, limit, filters)
	if err != nil {
		logger.ErrorLogger.Printf("Service: Error listing chart of accounts from repository: %v", err)
		return nil, 0, err // Propagate
	}
	logger.InfoLogger.Printf("Service: Successfully listed chart of accounts. Found: %d, Total: %d", len(accounts), total)
	return accounts, total, nil
}

// --- Journal Entries Methods ---

func (s *accountingService) CreateJournalEntry(ctx context.Context, req dto.CreateJournalEntryRequest) (*models.JournalEntry, error) {
	logger.InfoLogger.Printf("Service: Attempting to create journal entry with description: %s", req.Description)

	// Validate request
	if req.EntryDate.IsZero() {
		req.EntryDate = time.Now() // Default to now if not provided
	}
	if len(req.Lines) == 0 {
		logger.WarnLogger.Println("Service: Journal entry must have at least one line.")
		return nil, errors.NewValidationError("journal entry must have at least one line", "lines")
	}

	var totalDebits float64
	var totalCredits float64
	journalLines := make([]models.JournalLine, len(req.Lines))

	for i, lineReq := range req.Lines {
		if lineReq.AccountID == uuid.Nil {
			logger.WarnLogger.Printf("Service: Journal line %d has missing account ID.", i+1)
			return nil, errors.NewValidationError(fmt.Sprintf("line %d: account_id is required", i+1), "lines.account_id")
		}
		if lineReq.Amount <= 0 { // Amounts should be positive, IsDebit determines effect
			logger.WarnLogger.Printf("Service: Journal line %d has invalid amount: %.2f", i+1, lineReq.Amount)
			return nil, errors.NewValidationError(fmt.Sprintf("line %d: amount must be positive", i+1), "lines.amount")
		}

		// Validate account ID exists and is active
		account, err := s.coaRepo.GetByID(ctx, lineReq.AccountID)
		if err != nil {
			if isNotFoundError(err) {
				logger.WarnLogger.Printf("Service: Account with ID %s for line %d not found.", lineReq.AccountID, i+1)
				return nil, errors.NewValidationError(fmt.Sprintf("line %d: account with ID %s not found", i+1, lineReq.AccountID), "lines.account_id")
			}
			logger.ErrorLogger.Printf("Service: Error fetching account %s for line %d: %v", lineReq.AccountID, i+1, err)
			return nil, err // Internal server error
		}
		if !account.IsActive {
			logger.WarnLogger.Printf("Service: Account %s (%s) for line %d is not active.", account.AccountCode, account.AccountName, i+1)
			return nil, errors.NewValidationError(fmt.Sprintf("line %d: account %s (%s) is not active", i+1, account.AccountCode, account.AccountName), "lines.account_id")
		}

		journalLines[i] = models.JournalLine{
			AccountID: lineReq.AccountID,
			Amount:    lineReq.Amount,
			Currency:  lineReq.Currency, // TODO: Validate currency code if necessary
			IsDebit:   lineReq.IsDebit,
		}
		if journalLines[i].Currency == "" {
			journalLines[i].Currency = "USD" // Default currency
		}

		if lineReq.IsDebit {
			totalDebits += lineReq.Amount
		} else {
			totalCredits += lineReq.Amount
		}
	}

	// Check if debits equal credits (with a small tolerance for float comparison)
	const tolerance = 1e-9 // Tolerance for floating point comparison
	if math.Abs(totalDebits-totalCredits) > tolerance {
		logger.WarnLogger.Printf("Service: Journal entry debits (%.2f) do not equal credits (%.2f).", totalDebits, totalCredits)
		return nil, errors.NewValidationError(fmt.Sprintf("debits (%.2f) must equal credits (%.2f)", totalDebits, totalCredits), "lines")
	}

	entryStatus := models.StatusDraft // Default status for new entries, can be changed by PostJournalEntry
	if req.Status != "" {             // Allow overriding status if provided and valid (e.g. for import)
		switch req.Status {
		case models.StatusDraft, models.StatusPosted: // Voided typically not set on create
			entryStatus = req.Status
		default:
			logger.WarnLogger.Printf("Service: Invalid status '%s' provided for new journal entry. Defaulting to DRAFT.", req.Status)
			// entryStatus remains StatusDraft
		}
	}


	entry := &models.JournalEntry{
		EntryDate:   req.EntryDate,
		Description: req.Description,
		Reference:   req.Reference,
		Status:      entryStatus,
		JournalLines: journalLines,
	}

	createdEntry, err := s.journalRepo.Create(ctx, entry)
	if err != nil {
		logger.ErrorLogger.Printf("Service: Error creating journal entry in repository: %v", err)
		return nil, err // Propagate
	}
	logger.InfoLogger.Printf("Service: Successfully created journal entry with ID: %s", createdEntry.ID)
	return createdEntry, nil
}

func (s *accountingService) GetJournalEntryByID(ctx context.Context, id uuid.UUID) (*models.JournalEntry, error) {
	logger.InfoLogger.Printf("Service: Attempting to get journal entry by ID: %s", id)
	entry, err := s.journalRepo.GetByID(ctx, id)
	if err != nil {
		logger.ErrorLogger.Printf("Service: Error getting journal entry by ID %s from repository: %v", id, err)
		return nil, err // Propagate
	}
	logger.InfoLogger.Printf("Service: Successfully retrieved journal entry by ID: %s", id)
	return entry, nil
}

func (s *accountingService) UpdateJournalEntry(ctx context.Context, id uuid.UUID, req dto.UpdateJournalEntryRequest) (*models.JournalEntry, error) {
	logger.InfoLogger.Printf("Service: Attempting to update journal entry with ID: %s", id)

	existingEntry, err := s.journalRepo.GetByID(ctx, id)
	if err != nil {
		logger.ErrorLogger.Printf("Service: Error finding journal entry %s for update: %v", id, err)
		return nil, err // Propagate (could be NotFoundError)
	}

	// Business rule: Cannot update a 'POSTED' or 'VOIDED' entry in certain ways.
	// For example, lines might be uneditable, or only description/reference can change.
	// If status is POSTED, only allow changes to non-financial fields or require un-posting first.
	if existingEntry.Status == models.StatusPosted || existingEntry.Status == models.StatusVoided {
		// Simplified: allow updating Description and Reference for Posted/Voided entries
		canUpdate := false
		if req.Description != nil && *req.Description != existingEntry.Description {
			existingEntry.Description = *req.Description
			canUpdate = true
		}
		if req.Reference != nil && *req.Reference != existingEntry.Reference {
			existingEntry.Reference = *req.Reference
			canUpdate = true
		}

		// If only description/reference changed and other financial fields are nil/empty in request
		if (req.EntryDate == nil || (*req.EntryDate).IsZero() || (*req.EntryDate).Equal(existingEntry.EntryDate)) &&
		   (req.Lines == nil || len(*req.Lines) == 0) &&
		   (req.Status == nil || *req.Status == existingEntry.Status) {
			if !canUpdate { // No actual changes requested for allowed fields
				logger.InfoLogger.Printf("Service: No updatable fields provided for posted/voided journal entry %s.", id)
				return existingEntry, nil // No change, return existing
			}
			// Proceed to update only description/reference
			updatedEntry, err := s.journalRepo.Update(ctx, existingEntry) // Repo update must handle partial field updates
			if err != nil {
				logger.ErrorLogger.Printf("Service: Error updating non-financial fields for posted/voided journal entry %s: %v", id, err)
				return nil, err
			}
			logger.InfoLogger.Printf("Service: Successfully updated non-financial fields for posted/voided journal entry %s", id)
			return updatedEntry, nil
		} else {
			logger.WarnLogger.Printf("Service: Journal entry %s is %s and cannot be fully updated. Un-post first or only update specific fields.", id, existingEntry.Status)
			return nil, errors.NewConflictError(fmt.Sprintf("cannot update a %s journal entry in this way", existingEntry.Status))
		}
	}

	// If DRAFT, allow full update
	if req.EntryDate != nil && !(*req.EntryDate).IsZero() {
		existingEntry.EntryDate = *req.EntryDate
	}
	if req.Description != nil {
		existingEntry.Description = *req.Description
	}
	if req.Reference != nil {
		existingEntry.Reference = *req.Reference
	}
	if req.Status != nil { // Handle status change, e.g., DRAFT to DRAFT (no change), or DRAFT to POSTED (use PostJournalEntry)
		if *req.Status == models.StatusPosted && existingEntry.Status == models.StatusDraft {
			// If trying to post via update, redirect to PostJournalEntry logic or handle here
			logger.InfoLogger.Printf("Service: Update request for entry %s includes posting. Will attempt to post.", id)
			// This will be handled by PostJournalEntry if called separately.
			// For now, let's assume status update to POSTED here means it should be posted.
			// The PostJournalEntry method has more dedicated logic.
			// It might be better to disallow status change to POSTED here and force use of PostJournalEntry.
			// For now, let's allow it if it's balanced.
			existingEntry.Status = models.StatusPosted
		} else if *req.Status != existingEntry.Status && *req.Status != models.StatusPosted { // Allow changing to other non-posted statuses if any
			existingEntry.Status = *req.Status
		}
	}


	if req.Lines != nil && len(*req.Lines) > 0 {
		var totalDebits float64
		var totalCredits float64
		updatedLines := make([]models.JournalLine, len(*req.Lines))

		for i, lineReq := range *req.Lines {
			if lineReq.AccountID == uuid.Nil {
				return nil, errors.NewValidationError(fmt.Sprintf("line %d: account_id is required", i+1), "lines.account_id")
			}
			if lineReq.Amount <= 0 {
				return nil, errors.NewValidationError(fmt.Sprintf("line %d: amount must be positive", i+1), "lines.amount")
			}
			account, err := s.coaRepo.GetByID(ctx, lineReq.AccountID)
			if err != nil { /* ... error handling ... */
				if isNotFoundError(err) { return nil, errors.NewValidationError(fmt.Sprintf("line %d: account %s not found", i+1, lineReq.AccountID), "")}
				return nil, err
			}
			if !account.IsActive { /* ... error handling ... */
				return nil, errors.NewValidationError(fmt.Sprintf("line %d: account %s not active", i+1, account.AccountCode), "")
			}

			updatedLines[i] = models.JournalLine{
				// ID might be needed if repo is matching lines by ID for update vs create.
				// If lineReq includes an ID, use it. GORM's association replace handles this.
				ID:        lineReq.ID, // Assumes DTO line includes ID for existing lines
				JournalID: existingEntry.ID, // Ensure JournalID is set for new lines
				AccountID: lineReq.AccountID,
				Amount:    lineReq.Amount,
				Currency:  lineReq.Currency,
				IsDebit:   lineReq.IsDebit,
			}
			if updatedLines[i].Currency == "" {
				updatedLines[i].Currency = "USD"
			}

			if lineReq.IsDebit {
				totalDebits += lineReq.Amount
			} else {
				totalCredits += lineReq.Amount
			}
		}
		const tolerance = 1e-9
		if math.Abs(totalDebits-totalCredits) > tolerance {
			return nil, errors.NewValidationError(fmt.Sprintf("debits (%.2f) must equal credits (%.2f)", totalDebits, totalCredits), "lines")
		}
		existingEntry.JournalLines = updatedLines
	} else if req.Lines != nil && len(*req.Lines) == 0 { // Explicitly empty lines array
        return nil, errors.NewValidationError("journal entry must have at least one line", "lines")
    }


	// If the entry is being marked as POSTED, ensure it's balanced.
    // This check is also done if lines were updated. If only status changed to POSTED, we need to re-check balance.
    if existingEntry.Status == models.StatusPosted {
        if !existingEntry.IsBalanced() { // IsBalanced method on JournalEntry model
            // If lines were not part of this update request, IsBalanced() uses existing lines.
            // If lines were part of request, it uses the new lines.
            currentDebits, currentCredits := existingEntry.TotalDebits(), existingEntry.TotalCredits()
            logger.WarnLogger.Printf("Service: Journal entry %s cannot be posted. Debits (%.2f) do not equal credits (%.2f).", id, currentDebits, currentCredits)
            return nil, errors.NewValidationError(
                fmt.Sprintf("cannot post entry, debits (%.2f) must equal credits (%.2f)", currentDebits, currentCredits),
                "lines",
            )
        }
    }


	updatedEntry, err := s.journalRepo.Update(ctx, existingEntry)
	if err != nil {
		logger.ErrorLogger.Printf("Service: Error updating journal entry %s in repository: %v", id, err)
		return nil, err // Propagate
	}
	logger.InfoLogger.Printf("Service: Successfully updated journal entry with ID: %s", updatedEntry.ID)
	return updatedEntry, nil
}

func (s *accountingService) DeleteJournalEntry(ctx context.Context, id uuid.UUID) error {
	logger.InfoLogger.Printf("Service: Attempting to delete journal entry with ID: %s", id)
	entry, err := s.journalRepo.GetByID(ctx, id)
	if err != nil {
		logger.ErrorLogger.Printf("Service: Error finding journal entry %s for deletion: %v", id, err)
		return err // Propagate (could be NotFoundError)
	}

	// Business rule: Cannot delete a 'POSTED' entry. It must be 'VOIDED' or 'UNPOSTED' first.
	if entry.Status == models.StatusPosted {
		logger.WarnLogger.Printf("Service: Cannot delete journal entry %s because it is POSTED. Void or un-post first.", id)
		return errors.NewConflictError(fmt.Sprintf("cannot delete a POSTED journal entry (ID: %s). Void or un-post first.", id))
	}
	// VOIDED entries can often be deleted (archived by soft delete). DRAFT entries can be deleted.

	if err := s.journalRepo.Delete(ctx, id); err != nil {
		logger.ErrorLogger.Printf("Service: Error deleting journal entry %s from repository: %v", id, err)
		return err // Propagate
	}
	logger.InfoLogger.Printf("Service: Successfully deleted journal entry with ID: %s", id)
	return nil
}

func (s *accountingService) ListJournalEntries(ctx context.Context, req dto.ListJournalEntriesRequest) ([]*models.JournalEntry, int64, error) {
	logger.InfoLogger.Printf("Service: Listing journal entries with filters: %+v", req)
	filters := make(map[string]interface{})
	if req.Description != "" {
		filters["description"] = req.Description
	}
	if req.Reference != "" {
		filters["reference"] = req.Reference
	}
	if req.Status != "" {
		filters["status"] = req.Status
	}
	if !req.DateFrom.IsZero() {
		filters["date_from"] = req.DateFrom
	}
	if !req.DateTo.IsZero() {
		filters["date_to"] = req.DateTo
	}
	// TODO: Add filter by account_id if DTO supports it

	offset := 0
	if req.Page > 0 && req.Limit > 0 {
		offset = (req.Page - 1) * req.Limit
	}
	limit := req.Limit
	if limit == 0 {
		limit = 20 // Default
	}

	entries, total, err := s.journalRepo.List(ctx, offset, limit, filters)
	if err != nil {
		logger.ErrorLogger.Printf("Service: Error listing journal entries from repository: %v", err)
		return nil, 0, err // Propagate
	}
	logger.InfoLogger.Printf("Service: Successfully listed journal entries. Found: %d, Total: %d", len(entries), total)
	return entries, total, nil
}

func (s *accountingService) PostJournalEntry(ctx context.Context, id uuid.UUID) (*models.JournalEntry, error) {
	logger.InfoLogger.Printf("Service: Attempting to post journal entry with ID: %s", id)
	entry, err := s.journalRepo.GetByID(ctx, id)
	if err != nil {
		logger.ErrorLogger.Printf("Service: Error finding journal entry %s for posting: %v", id, err)
		return nil, err // Propagate (could be NotFoundError)
	}

	if entry.Status == models.StatusPosted {
		logger.WarnLogger.Printf("Service: Journal entry %s is already POSTED.", id)
		return entry, nil // Already posted, no action needed
	}
	if entry.Status == models.StatusVoided {
		logger.WarnLogger.Printf("Service: Journal entry %s is VOIDED and cannot be posted.", id)
		return nil, errors.NewConflictError(fmt.Sprintf("cannot post a VOIDED journal entry (ID: %s)", id))
	}

	// Ensure entry is balanced before posting
	if !entry.IsBalanced() {
		debits, credits := entry.TotalDebits(), entry.TotalCredits()
		logger.WarnLogger.Printf("Service: Journal entry %s is not balanced. Debits: %.2f, Credits: %.2f. Cannot post.", id, debits, credits)
		return nil, errors.NewValidationError(fmt.Sprintf("entry is not balanced (Debits: %.2f, Credits: %.2f)", debits, credits), "lines")
	}

	// Additional checks before posting (e.g., all accounts in lines are active)
	for _, line := range entry.JournalLines {
		// Account should have been validated on creation/update, but a check here is good defense
		account, err := s.coaRepo.GetByID(ctx, line.AccountID)
		if err != nil { // Should not happen if data integrity is maintained
			logger.ErrorLogger.Printf("Service: Critical error - account %s in journal entry %s not found during posting: %v", line.AccountID, id, err)
			return nil, errors.NewInternalServerError("error validating account during posting", err)
		}
		if !account.IsActive {
			logger.WarnLogger.Printf("Service: Account %s (%s) in journal entry %s is inactive. Cannot post.", account.AccountCode, account.AccountName, id)
			return nil, errors.NewConflictError(fmt.Sprintf("account %s (%s) is inactive", account.AccountCode, account.AccountName))
		}
	}

	// Update status to POSTED
	// entry.Status = models.StatusPosted // This would be done by UpdateJournalEntryStatus
	// updatedEntry, err := s.journalRepo.Update(ctx, entry) // This might be too broad, use specific status update
	err = s.journalRepo.UpdateJournalEntryStatus(ctx, id, models.StatusPosted)
	if err != nil {
		logger.ErrorLogger.Printf("Service: Error updating journal entry %s status to POSTED in repository: %v", id, err)
		return nil, err
	}

	// Reload the entry to confirm the change
	postedEntry, err := s.journalRepo.GetByID(ctx, id)
	if err != nil {
		logger.ErrorLogger.Printf("Service: Error reloading journal entry %s after posting: %v", id, err)
		return nil, err
	}


	logger.InfoLogger.Printf("Service: Successfully posted journal entry with ID: %s", postedEntry.ID)
	return postedEntry, nil
}

// --- Reporting Methods ---

func (s *accountingService) GetTrialBalance(ctx context.Context, req dto.TrialBalanceRequest) (*dto.TrialBalanceResponse, error) {
	logger.InfoLogger.Printf("Service: Generating Trial Balance for period ending %s", req.EndDate.Format("2006-01-02"))

	if req.EndDate.IsZero() {
		req.EndDate = time.Now() // Default to current date if not specified
	}
	// Start date for trial balance is effectively the beginning of time for cumulative balances,
	// or a specific start if it's a period-specific TB (less common for standard TB).
	// For simplicity, let's assume it's cumulative up to EndDate.
	// The repository method GetJournalEntriesForTrialBalance needs a start and end.
	// For a cumulative trial balance, startDate could be very early, or the logic adapted.
	// Let's make startDate configurable or default to a very early date for full history.
	var startDate time.Time
	if req.StartDate.IsZero() {
		startDate = time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC) // A very early date
	} else {
		startDate = req.StartDate
	}


	// 1. Fetch all posted journal entries up to the EndDate
	// The repo method GetJournalEntriesForTrialBalance should handle this.
	// It needs to fetch entries with status "POSTED" and entry_date <= req.EndDate.
	// It should also preload JournalLines and their associated ChartOfAccount.
	entries, err := s.journalRepo.GetJournalEntriesForTrialBalance(ctx, startDate, req.EndDate)
	if err != nil {
		logger.ErrorLogger.Printf("Service: Error fetching journal entries for trial balance: %v", err)
		return nil, err
	}

	// 2. Aggregate balances for each account
	accountBalances := make(map[uuid.UUID]float64)    // K: AccountID, V: Balance (positive for debit, negative for credit normal balance)
	accountDetails := make(map[uuid.UUID]models.ChartOfAccount) // K: AccountID, V: Account details

	for _, entry := range entries {
		for _, line := range entry.JournalLines {
			if line.ChartOfAccount == nil { // Defensive check
				logger.ErrorLogger.Printf("Service: TrialBalance - Journal line %s is missing ChartOfAccount details. Skipping.", line.ID)
				// This indicates an issue with preloading or data integrity.
				// Potentially fetch it:
				// acc, accErr := s.coaRepo.GetByID(ctx, line.AccountID)
				// if accErr != nil { continue }
				// line.ChartOfAccount = acc
				continue
			}

			if _, exists := accountDetails[line.AccountID]; !exists {
				accountDetails[line.AccountID] = *line.ChartOfAccount
			}

			amount := line.Amount
			if line.IsDebit {
				accountBalances[line.AccountID] += amount
			} else {
				accountBalances[line.AccountID] -= amount
			}
		}
	}

	// 3. Format for response
	var trialBalanceLines []dto.TrialBalanceLine
	var totalDebits float64
	var totalCredits float64

	// It's good practice to list all accounts from CoA, even those with zero balance.
    // So, fetch all active accounts first.
    allAccounts, _, err := s.coaRepo.List(ctx, 0, 0, map[string]interface{}{"is_active": true}) // Get all active accounts
    if err != nil {
        logger.ErrorLogger.Printf("Service: Error fetching all active accounts for trial balance: %v", err)
        return nil, errors.NewInternalServerError("failed to fetch accounts for trial balance", err)
    }

    for _, acc := range allAccounts {
        balance := accountBalances[acc.ID] // Will be 0.0 if no transactions for this account

        debitAmount := 0.0
        creditAmount := 0.0

        // Determine if balance is debit or credit based on account type's normal balance
        // Assets, Expenses normally have Debit balances.
        // Liabilities, Equity, Revenue normally have Credit balances.
        isDebitNormalBalance := acc.AccountType == models.Asset || acc.AccountType == models.Expense

        if isDebitNormalBalance {
            if balance >= 0 { // Normal debit balance or zero
                debitAmount = balance
            } else { // Abnormal credit balance
                creditAmount = -balance // Show as positive credit
            }
        } else { // Credit normal balance (Liability, Equity, Revenue)
            if balance <= 0 { // Normal credit balance or zero
                creditAmount = -balance
            } else { // Abnormal debit balance
                debitAmount = balance // Show as positive debit
            }
        }

        // Only add lines if there's a non-zero balance, or if req.IncludeZeroBalance is true
        if req.IncludeZeroBalanceAccounts || debitAmount != 0 || creditAmount != 0 {
            trialBalanceLines = append(trialBalanceLines, dto.TrialBalanceLine{
                AccountCode: acc.AccountCode,
                AccountName: acc.AccountName,
                Debit:       debitAmount,
                Credit:      creditAmount,
            })
        }

        totalDebits += debitAmount
        totalCredits += creditAmount
    }


	// Sort lines by account code for consistent output (optional)
	// sort.Slice(trialBalanceLines, func(i, j int) bool {
	// 	return trialBalanceLines[i].AccountCode < trialBalanceLines[j].AccountCode
	// })


	// Final check for balance (should always balance if accounting is correct)
	const tolerance = 1e-9
	if math.Abs(totalDebits-totalCredits) > tolerance {
		logger.ErrorLogger.Printf("Service: Trial Balance is out of balance! Debits: %.2f, Credits: %.2f", totalDebits, totalCredits)
		// This is a critical system error if it happens.
		return nil, errors.NewInternalServerError(fmt.Sprintf("trial balance generation failed: totals are unbalanced (D:%.2f, C:%.2f)", totalDebits, totalCredits), nil)
	}

	response := &dto.TrialBalanceResponse{
		ReportDate: req.EndDate,
		Lines:      trialBalanceLines,
		TotalDebits:  totalDebits,
		TotalCredits: totalCredits,
	}

	logger.InfoLogger.Printf("Service: Successfully generated Trial Balance for period ending %s. Total Debits: %.2f, Total Credits: %.2f", req.EndDate.Format("2006-01-02"), totalDebits, totalCredits)
	return response, nil
}


func (s *accountingService) GetAccountBalance(ctx context.Context, accountID uuid.UUID, date time.Time) (float64, error) {
    logger.InfoLogger.Printf("Service: Calculating balance for account %s as of %s", accountID, date.Format("2006-01-02"))

    // Validate account
    account, err := s.coaRepo.GetByID(ctx, accountID)
    if err != nil {
        logger.ErrorLogger.Printf("Service: Error fetching account %s for balance calculation: %v", accountID, err)
        return 0, err // Propagate NotFound or InternalServerError
    }

    // Fetch all journal entries involving this account, up to the specified date, that are POSTED.
    // This requires a repository method that can filter journal entries by account ID and date range.
    // Let's assume journalRepo.GetJournalEntriesByAccountID handles this.
    // We need all entries from the beginning of time up to `date`.
    // The GetJournalEntriesByAccountID method in repo takes offset/limit, set to 0,0 for all.
    // And startDate can be very early date.
    veryEarlyDate := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
    entries, _, err := s.journalRepo.GetJournalEntriesByAccountID(ctx, accountID, 0, 0, veryEarlyDate, date)
    if err != nil {
        logger.ErrorLogger.Printf("Service: Error fetching journal entries for account %s: %v", accountID, err)
        return 0, errors.NewInternalServerError("failed to fetch journal entries for account balance", err)
    }

    var balance float64
    for _, entry := range entries {
        if entry.Status != models.StatusPosted { // Double check, though repo method should filter
            continue
        }
        for _, line := range entry.JournalLines {
            if line.AccountID == accountID {
                if line.IsDebit {
                    balance += line.Amount
                } else {
                    balance -= line.Amount
                }
            }
        }
    }

    // The 'balance' here is net change. Depending on account type, this means different things.
    // For Asset/Expense (debit normal): positive balance is a debit balance.
    // For Liability/Equity/Revenue (credit normal): positive balance (from this calculation) means it's a debit effect,
    // so if it's a credit normal account, its "actual" balance would be -balance.
    // The function should probably return the "natural" balance.
    // E.g. if it's an Asset account, balance = 100 means $100 Debit.
    // If it's a Liability account, balance = -100 means $100 Credit (since calc is Debit-Credit).
    // The current calculation (sum of debits - sum of credits for that account) is standard.

    logger.InfoLogger.Printf("Service: Calculated balance for account %s (%s) as of %s: %.2f", account.AccountCode, account.AccountName, date.Format("2006-01-02"), balance)
    return balance, nil
}


// --- Helper Functions ---

// isNotFoundError checks if the error is of type *errors.NotFoundError.
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*errors.NotFoundError)
	return ok
}
