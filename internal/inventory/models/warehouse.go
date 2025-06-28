package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Warehouse represents a physical or logical location where inventory is stored.
type Warehouse struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;" json:"id"`
	Code      string         `gorm:"type:varchar(20);not null;uniqueIndex" json:"code"` // Unique code for the warehouse
	Name      string         `gorm:"type:varchar(100);not null" json:"name"`
	Location  string         `gorm:"type:varchar(255)" json:"location,omitempty"` // Address or description of location
	IsActive  bool           `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // For soft deletes

	// Potential future fields:
	// WarehouseType string `gorm:"type:varchar(30)" json:"warehouse_type,omitempty"` // e.g., Main, Transit, Quarantine, Retail
	// IsDefault     bool   `gorm:"default:false" json:"is_default"` // If there's a default warehouse for operations

	// Associations
	// InventoryTransactions []InventoryTransaction `gorm:"foreignKey:WarehouseID" json:"-"`
}

// TableName specifies the table name for Warehouse model.
func (Warehouse) TableName() string {
	return "warehouses"
}

// BeforeCreate will set a UUID for the new warehouse.
func (w *Warehouse) BeforeCreate(tx *gorm.DB) (err error) {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	if w.Code == "" {
		return gorm.ErrInvalidData // Or custom error "warehouse code is required"
	}
	if w.Name == "" {
		return gorm.ErrInvalidData // Or custom error "warehouse name is required"
	}
	return
}
