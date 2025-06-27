package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AccountType represents the type of account (e.g., Asset, Liability, Equity, Revenue, Expense)
type AccountType string

const (
	Asset     AccountType = "ASSET"
	Liability AccountType = "LIABILITY"
	Equity    AccountType = "EQUITY"
	Revenue   AccountType = "REVENUE"
	Expense   AccountType = "EXPENSE"
)

// ChartOfAccount represents the structure for an account in the chart of accounts.
type ChartOfAccount struct {
	ID              uuid.UUID      `gorm:"type:uuid;primary_key;" json:"id"`
	AccountCode     string         `gorm:"type:varchar(20);not null;uniqueIndex" json:"account_code"` // Unique index added
	AccountName     string         `gorm:"type:varchar(100);not null" json:"account_name"`
	AccountType     AccountType    `gorm:"type:varchar(50);not null" json:"account_type"` // Using custom type
	ParentAccountID *uuid.UUID     `gorm:"type:uuid;index" json:"parent_account_id"`      // Pointer to allow null, index for FK
	IsActive        bool           `gorm:"default:true" json:"is_active"`
	CreatedAt       time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"` // For soft deletes

	// Associations
	// ParentAccount *ChartOfAccount `gorm:"foreignKey:ParentAccountID;references:ID" json:"parent_account,omitempty"`
	// ChildAccounts []*ChartOfAccount `gorm:"foreignKey:ParentAccountID;references:ID" json:"child_accounts,omitempty"`
	// JournalLines  []JournalLine   `gorm:"foreignKey:AccountID" json:"-"` // Journal lines associated with this account
}

// TableName specifies the table name for ChartOfAccount model.
func (ChartOfAccount) TableName() string {
	return "chart_of_accounts"
}

// BeforeCreate will set a UUID rather than numeric ID.
func (coa *ChartOfAccount) BeforeCreate(tx *gorm.DB) (err error) {
	if coa.ID == uuid.Nil {
		coa.ID = uuid.New()
	}
	// Validate AccountType
	switch coa.AccountType {
	case Asset, Liability, Equity, Revenue, Expense:
		// valid type
	default:
		return gorm.ErrInvalidData // Or a more specific error
	}
	return
}

// Helper function to get ParentAccountID as a string for JSON, if needed
func (coa *ChartOfAccount) GetParentAccountIDString() string {
	if coa.ParentAccountID == nil {
		return ""
	}
	return coa.ParentAccountID.String()
}
