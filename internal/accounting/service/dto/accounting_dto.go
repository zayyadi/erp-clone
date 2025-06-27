package dto

import (
	"erp-system/internal/accounting/models"
	"time"

	"github.com/google/uuid"
)

// --- Chart of Accounts DTOs ---

// CreateChartOfAccountRequest defines the structure for creating a new chart of account.
type CreateChartOfAccountRequest struct {
	AccountCode     string               `json:"account_code" binding:"required,min=1,max=20"`
	AccountName     string               `json:"account_name" binding:"required,min=1,max=100"`
	AccountType     models.AccountType   `json:"account_type" binding:"required"` // Should be validated against enum values
	ParentAccountID *uuid.UUID           `json:"parent_account_id,omitempty"`
	IsActive        bool                 `json:"is_active"` // Defaults to true if omitted by user, handled by service/model
}

// UpdateChartOfAccountRequest defines the structure for updating an existing chart of account.
// Pointers are used to distinguish between a field not being provided and a field being set to its zero value.
type UpdateChartOfAccountRequest struct {
	AccountName     *string              `json:"account_name,omitempty" binding:"omitempty,min=1,max=100"`
	AccountType     *models.AccountType  `json:"account_type,omitempty"` // Should be validated against enum values
	ParentAccountID *uuid.UUID           `json:"parent_account_id,omitempty"` // Allows setting to null by passing explicit null or omitting, or changing
	IsActive        *bool                `json:"is_active,omitempty"`
}

// ListChartOfAccountsRequest defines parameters for listing chart of accounts.
type ListChartOfAccountsRequest struct {
	Page          int                `form:"page,default=1"`
	Limit         int                `form:"limit,default=20"`
	AccountName   string             `form:"account_name,omitempty"`
	AccountType   models.AccountType `form:"account_type,omitempty"`
	IsActive      *bool              `form:"is_active,omitempty"` // Pointer to differentiate not set, true, false
	// Add other filter fields like ParentAccountID, etc. if needed
}


// --- Journal Entry DTOs ---

// JournalLineRequest defines a line item within a journal entry request.
type JournalLineRequest struct {
	ID        uuid.UUID `json:"id,omitempty"` // Used for updates if lines can be individually identified
	AccountID uuid.UUID `json:"account_id" binding:"required"`
	Amount    float64   `json:"amount" binding:"required,gt=0"` // Amount should be positive
	Currency  string    `json:"currency,omitempty"`             // Defaults to USD if empty
	IsDebit   bool      `json:"is_debit"`                       // True for Debit, False for Credit
}

// CreateJournalEntryRequest defines the structure for creating a new journal entry.
type CreateJournalEntryRequest struct {
	EntryDate   time.Time            `json:"entry_date,omitempty"` // Defaults to Now if omitted
	Description string               `json:"description" binding:"max=255"`
	Reference   string               `json:"reference,omitempty" binding:"max=100"`
	Status      models.JournalStatus `json:"status,omitempty"`      // Optional: e.g. "DRAFT", "POSTED". Defaults to DRAFT in service.
	Lines       []JournalLineRequest `json:"lines" binding:"required,min=1,dive"` // dive validates each element in slice
}

// UpdateJournalEntryRequest defines the structure for updating an existing journal entry.
type UpdateJournalEntryRequest struct {
	EntryDate   *time.Time           `json:"entry_date,omitempty"`
	Description *string              `json:"description,omitempty" binding:"omitempty,max=255"`
	Reference   *string              `json:"reference,omitempty" binding:"omitempty,max=100"`
	Status      *models.JournalStatus `json:"status,omitempty"` // e.g. "DRAFT", "POSTED", "VOIDED"
	Lines       *[]JournalLineRequest `json:"lines,omitempty" binding:"omitempty,min=1,dive"` // Pointer to allow omitting lines update
}

// ListJournalEntriesRequest defines parameters for listing journal entries.
type ListJournalEntriesRequest struct {
	Page        int                  `form:"page,default=1"`
	Limit       int                  `form:"limit,default=20"`
	Description string               `form:"description,omitempty"`
	Reference   string               `form:"reference,omitempty"`
	Status      models.JournalStatus `form:"status,omitempty"`
	DateFrom    time.Time            `form:"date_from,omitempty" time_format:"2006-01-02"`
	DateTo      time.Time            `form:"date_to,omitempty" time_format:"2006-01-02"`
	AccountID   uuid.UUID            `form:"account_id,omitempty"` // To filter entries affecting a specific account
}


// --- Reporting DTOs ---

// TrialBalanceRequest defines parameters for generating a trial balance report.
type TrialBalanceRequest struct {
	StartDate                time.Time `json:"start_date,omitempty" form:"start_date,omitempty" time_format:"2006-01-02"` // For period-specific changes, not typical for TB itself
	EndDate                  time.Time `json:"end_date" form:"end_date" binding:"required" time_format:"2006-01-02"`
	IncludeZeroBalanceAccounts bool    `json:"include_zero_balance_accounts,omitempty" form:"include_zero_balance_accounts,omitempty"`
	// Add other parameters like specific subsidiary, department, etc.
}

// TrialBalanceLine represents a single line in the trial balance report.
type TrialBalanceLine struct {
	AccountCode string  `json:"account_code"`
	AccountName string  `json:"account_name"`
	Debit       float64 `json:"debit"`
	Credit      float64 `json:"credit"`
}

// TrialBalanceResponse is the structure for the trial balance report.
type TrialBalanceResponse struct {
	ReportDate   time.Time          `json:"report_date"`
	Lines        []TrialBalanceLine `json:"lines"`
	TotalDebits  float64            `json:"total_debits"`
	TotalCredits float64            `json:"total_credits"`
}


// BalanceSheetRequest (Example structure, can be expanded)
type BalanceSheetRequest struct {
    AsOfDate time.Time `json:"as_of_date" binding:"required" time_format:"2006-01-02"`
    // Filters: Department, Subsidiary, etc.
}

// BalanceSheetSection (Example for BS structure)
type BalanceSheetSection struct {
    Title    string             `json:"title"`
    Accounts []BalanceSheetLine `json:"accounts"`
    Total    float64            `json:"total"`
}

// BalanceSheetLine (Example for BS line)
type BalanceSheetLine struct {
    AccountName string  `json:"account_name"`
    Amount      float64 `json:"amount"`
    // SubLines if there's a hierarchy within the report
}

// BalanceSheetResponse (Example)
type BalanceSheetResponse struct {
    ReportDate     time.Time             `json:"report_date"`
    Assets         BalanceSheetSection   `json:"assets"`
    Liabilities    BalanceSheetSection   `json:"liabilities"`
    Equity         BalanceSheetSection   `json:"equity"`
    TotalLiabilitiesAndEquity float64    `json:"total_liabilities_and_equity"`
    // Verification (TotalAssets == TotalLiabilitiesAndEquity)
}


// ProfitAndLossRequest (Example structure)
type ProfitAndLossRequest struct {
    StartDate time.Time `json:"start_date" binding:"required" time_format:"2006-01-02"`
    EndDate   time.Time `json:"end_date" binding:"required" time_format:"2006-01-02"`
    // Filters: Department, Project, etc.
}

// ProfitAndLossSection (Example for P&L structure)
type ProfitAndLossSection struct {
    Title    string                `json:"title"` // e.g., "Revenue", "Cost of Goods Sold", "Operating Expenses"
    Accounts []ProfitAndLossLine   `json:"accounts"`
    Total    float64               `json:"total"`
}

// ProfitAndLossLine (Example for P&L line)
type ProfitAndLossLine struct {
    AccountName string  `json:"account_name"`
    Amount      float64 `json:"amount"`
}

// ProfitAndLossResponse (Example)
type ProfitAndLossResponse struct {
    ReportPeriod string                 `json:"report_period"` // e.g., "For the period DD/MM/YYYY to DD/MM/YYYY"
    Revenue      ProfitAndLossSection   `json:"revenue"`
    COGS         ProfitAndLossSection   `json:"cogs"`
    GrossProfit  float64                `json:"gross_profit"`
    Expenses     ProfitAndLossSection   `json:"expenses"` // Operating Expenses
    NetIncome    float64                `json:"net_income"`   // Or Net Loss
}

// General API Response Wrappers (Optional, but good practice)

// SuccessResponse wraps a successful API response.
type SuccessResponse struct {
	Status  string      `json:"status"` // e.g., "success"
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse wraps an error API response.
type ErrorResponse struct {
	Status  string `json:"status"` // e.g., "error"
	Message string `json:"message"`
	Details interface{} `json:"details,omitempty"` // e.g., validation errors map[string]string
}
