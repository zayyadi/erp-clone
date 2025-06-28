package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// JournalStatus represents the status of a journal entry (e.g., Draft, Posted, Voided)
type JournalStatus string

const (
	StatusDraft  JournalStatus = "DRAFT"
	StatusPosted JournalStatus = "POSTED"
	StatusVoided JournalStatus = "VOIDED" // Example of an additional status
)

// JournalEntry represents a financial transaction header.
type JournalEntry struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;" json:"id"`
	EntryDate   time.Time      `gorm:"not null" json:"entry_date"`
	Description string         `gorm:"type:varchar(255)" json:"description"`
	Reference   string         `gorm:"type:varchar(100)" json:"reference"`                       // E.g., Invoice number, PO number
	Status      JournalStatus  `gorm:"type:varchar(20);default:'POSTED';not null" json:"status"` // Default to DRAFT might be safer in some flows
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Associations
	JournalLines []JournalLine `gorm:"foreignKey:JournalID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"journal_lines"` // Lines associated with this entry
}

// JournalLine represents a single debit or credit in a journal entry.
// type JournalLine struct {
// 	ID          uuid.UUID      `gorm:"type:uuid;primary_key;" json:"id"`
// 	JournalID   uuid.UUID      `gorm:"type:uuid;not null;index" json:"journal_id"`
// 	AccountID   uuid.UUID      `gorm:"type:uuid;not null;index" json:"account_id"`
// 	Amount      float64        `gorm:"type:decimal(18,4);not null" json:"amount"` // Use decimal in DB for precision
// 	Currency    string         `gorm:"type:varchar(3);not null;default:'USD'" json:"currency"`
// 	IsDebit     bool           `gorm:"not null" json:"is_debit"`
// 	Description string         `gorm:"type:varchar(255)" json:"description"`
// 	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
// 	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
// 	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

// 	// Associations
// 	ChartOfAccount *ChartOfAccount `gorm:"foreignKey:AccountID" json:"chart_of_account,omitempty"`
// }

// TableName specifies the table name for JournalEntry model.
func (JournalEntry) TableName() string {
	return "journal_entries"
}

// BeforeCreate will set a UUID for the new journal entry.
func (je *JournalEntry) BeforeCreate(tx *gorm.DB) (err error) {
	if je.ID == uuid.Nil {
		je.ID = uuid.New()
	}
	// Validate Status
	switch je.Status {
	case StatusDraft, StatusPosted, StatusVoided:
		// valid status
	default:
		// If status is not set or invalid, default to DRAFT or POSTED based on business rule
		// For now, GORM default handles 'POSTED' if not provided.
		// If a specific default is needed on create when an invalid value is passed, set it here.
		// e.g. je.Status = StatusDraft
		if je.Status == "" { // if GORM default is not enough
			je.Status = StatusPosted // As per schema, but DRAFT might be better
		}
	}
	if je.EntryDate.IsZero() {
		je.EntryDate = time.Now()
	}
	return
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
	return
}

// TotalDebits calculates the sum of all debit amounts in the journal lines.
func (je *JournalEntry) TotalDebits() float64 {
	var total float64
	for _, line := range je.JournalLines {
		if line.IsDebit {
			total += line.Amount
		}
	}
	return total
}

// TotalCredits calculates the sum of all credit amounts in the journal lines.
func (je *JournalEntry) TotalCredits() float64 {
	var total float64
	for _, line := range je.JournalLines {
		if !line.IsDebit {
			total += line.Amount
		}
	}
	return total
}

// IsBalanced checks if the journal entry's debits equal its credits.
// Note: This is a simplified check and doesn't handle floating point inaccuracies well.
// For financial calculations, use decimal types or scaled integers.
func (je *JournalEntry) IsBalanced() bool {
	// A small tolerance for float comparison might be needed if using float64 directly
	// const tolerance = 1e-9
	// return math.Abs(je.TotalDebits() - je.TotalCredits()) < tolerance
	return je.TotalDebits() == je.TotalCredits() // Simpler for now
}
