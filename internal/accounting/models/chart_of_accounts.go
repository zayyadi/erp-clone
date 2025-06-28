package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AccountType defines the type of account (e.g., Asset, Liability).
type AccountType string

const (
	Asset     AccountType = "ASSET"
	Liability AccountType = "LIABILITY"
	Equity    AccountType = "EQUITY"
	Revenue   AccountType = "REVENUE"
	Expense   AccountType = "EXPENSE"
)

// ChartOfAccount represents an account in the chart of accounts.
type ChartOfAccount struct {
	ID              uuid.UUID      `gorm:"type:uuid;primary_key;" json:"id"`
	AccountCode     string         `gorm:"type:varchar(20);not null;uniqueIndex" json:"account_code"`
	AccountName     string         `gorm:"type:varchar(100);not null" json:"account_name"`
	AccountType     AccountType    `gorm:"type:varchar(20);not null;index" json:"account_type"`
	IsActive        bool           `gorm:"not null;default:true;index" json:"is_active"`
	Description     string         `gorm:"type:varchar(255)" json:"description"`
	ParentAccountID *uuid.UUID     `gorm:"type:uuid;index" json:"parent_account_id"`
	CreatedAt       time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for ChartOfAccount model.
func (ChartOfAccount) TableName() string {
	return "chart_of_accounts"
}

// BeforeCreate will set a UUID for the new account.
func (coa *ChartOfAccount) BeforeCreate(tx *gorm.DB) (err error) {
	if coa.ID == uuid.Nil {
		coa.ID = uuid.New()
	}
	return
}
