package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	// "github.com/shopspring/decimal" // Recommended for currency/financial calculations
)

// JournalLine represents a single line item within a journal entry.
type JournalLine struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	JournalID uuid.UUID `gorm:"type:uuid;not null;index" json:"journal_id"` // Foreign key to JournalEntry
	AccountID uuid.UUID `gorm:"type:uuid;not null;index" json:"account_id"` // Foreign key to ChartOfAccount
	// Amount    decimal.Decimal `gorm:"type:numeric(15,2);not null" json:"amount"` // Using decimal for precision
	Amount    float64   `gorm:"type:numeric(15,2);not null" json:"amount"` // Using float64 for now, decimal is better
	Currency  string    `gorm:"type:varchar(3);default:'USD'" json:"currency"`
	IsDebit   bool      `gorm:"not null" json:"is_debit"` // True for debit, False for credit
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	// DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // Soft delete for lines might not always be needed if entry is soft deleted

	// Associations
	JournalEntry   *JournalEntry   `gorm:"foreignKey:JournalID;references:ID" json:"-"` // Pointer to avoid cyclic dependencies if both ways
	ChartOfAccount *ChartOfAccount `gorm:"foreignKey:AccountID;references:ID" json:"chart_of_account,omitempty"` // Made it a pointer and added json tag
}

// TableName specifies the table name for JournalLine model.
func (JournalLine) TableName() string {
	return "journal_lines"
}

// BeforeCreate will set a UUID for the new journal line.
func (jl *JournalLine) BeforeCreate(tx *gorm.DB) (err error) {
	if jl.ID == uuid.Nil {
		jl.ID = uuid.New()
	}
	if jl.Currency == "" {
		jl.Currency = "USD" // Ensure default if not provided
	}
	// Basic validation for amount
	if jl.Amount < 0 {
		return gorm.ErrInvalidData // Or a custom error: amounts should be non-negative
	}
	return
}

/*
// Example of using shopspring/decimal if you add it:
// Need to add `go get github.com/shopspring/decimal`
// And import it in this file.

// JournalLine represents a single line item within a journal entry.
type JournalLineDecimal struct {
	ID        uuid.UUID       `gorm:"type:uuid;primary_key;" json:"id"`
	JournalID uuid.UUID       `gorm:"type:uuid;not null;index" json:"journal_id"`
	AccountID uuid.UUID       `gorm:"type:uuid;not null;index" json:"account_id"`
	Amount    decimal.Decimal `gorm:"type:numeric(15,2);not null" json:"amount"`
	Currency  string          `gorm:"type:varchar(3);default:'USD'" json:"currency"`
	IsDebit   bool            `gorm:"not null" json:"is_debit"`
	CreatedAt time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

func (jld *JournalLineDecimal) BeforeCreate(tx *gorm.DB) (err error) {
	if jld.ID == uuid.Nil {
		jld.ID = uuid.New()
	}
	if jld.Currency == "" {
		jld.Currency = "USD"
	}
	if jld.Amount.LessThan(decimal.Zero) {
		return gorm.ErrInvalidData // Amounts should be non-negative
	}
	return
}
*/
