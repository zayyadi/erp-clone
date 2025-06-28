package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ItemType represents the type of inventory item (e.g., RAW_MATERIAL, FINISHED_GOOD, WIP, NON_INVENTORY)
type ItemType string

const (
	RawMaterial  ItemType = "RAW_MATERIAL"
	FinishedGood ItemType = "FINISHED_GOOD"
	WorkInProgress ItemType = "WIP"          // Work-in-Progress
	NonInventory ItemType = "NON_INVENTORY" // For items not tracked in stock (e.g., services, supplies)
	// Add other types as needed
)

// Item represents an inventory item.
type Item struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key;" json:"id"`
	SKU           string         `gorm:"type:varchar(50);not null;uniqueIndex" json:"sku"` // Stock Keeping Unit, unique
	Name          string         `gorm:"type:varchar(100);not null" json:"name"`
	Description   string         `gorm:"type:text" json:"description,omitempty"`
	UnitOfMeasure string         `gorm:"type:varchar(20);not null" json:"unit_of_measure"` // e.g., PCS, KG, LTR, MTR
	ItemType      ItemType       `gorm:"type:varchar(20);not null" json:"item_type"`
	IsActive      bool           `gorm:"default:true" json:"is_active"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"` // For soft deletes

	// Potential future fields:
	// Barcode         string  `gorm:"type:varchar(100);index" json:"barcode,omitempty"`
	// Brand           string  `gorm:"type:varchar(50)" json:"brand,omitempty"`
	// Category        string  `gorm:"type:varchar(50)" json:"category,omitempty"`
	// StandardCost    *decimal.Decimal `gorm:"type:numeric(15,2)" json:"standard_cost,omitempty"` // If using standard costing
	// PurchaseUoM     string  `gorm:"type:varchar(20)" json:"purchase_uom,omitempty"`
	// SalesUoM        string  `gorm:"type:varchar(20)" json:"sales_uom,omitempty"`
	// MinStockLevel   *float64 `json:"min_stock_level,omitempty"`
	// MaxStockLevel   *float64 `json:"max_stock_level,omitempty"`
	// ReorderPoint    *float64 `json:"reorder_point,omitempty"`
	// LeadTimeDays    *int     `json:"lead_time_days,omitempty"` // For procurement

	// Associations (if needed directly on Item model, often handled by joining through transactions)
	// InventoryTransactions []InventoryTransaction `gorm:"foreignKey:ItemID" json:"-"`
}

// TableName specifies the table name for Item model.
func (Item) TableName() string {
	return "items"
}

// BeforeCreate will set a UUID for the new item.
func (i *Item) BeforeCreate(tx *gorm.DB) (err error) {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	// Validate ItemType
	switch i.ItemType {
	case RawMaterial, FinishedGood, WorkInProgress, NonInventory:
		// valid type
	default:
		// If type is mandatory and an invalid value is passed, return error.
		// If it can be empty, handle that case or set a default.
		if string(i.ItemType) == "" { // Check if it's empty string if that's possible input
			// return gorm.ErrInvalidData // Or a custom error: "item type is required"
			// For now, let's assume it's always provided and valid based on ItemType string consts.
			// If the type is not one of the constants, it's an invalid value.
		}
		// This check might be better in the service layer for clearer error messages to user.
		// For now, just ensuring it's one of the defined constants.
		// If an invalid string is passed, it won't match any case.
		// Depending on DB constraints, this might fail at DB level if type is enum there.
		// For VARCHAR, any string could be inserted if not validated here or by service.
	}
	if i.UnitOfMeasure == "" {
		return gorm.ErrInvalidData // Or custom error: "unit of measure is required"
	}
	return
}
