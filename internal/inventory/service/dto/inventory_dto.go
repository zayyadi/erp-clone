package dto

import (
	"erp-system/internal/inventory/models"
	"time"

	"github.com/google/uuid"
)

// --- Item DTOs ---

// CreateItemRequest defines the structure for creating a new item.
type CreateItemRequest struct {
	SKU           string           `json:"sku" binding:"required,min=1,max=50"`
	Name          string           `json:"name" binding:"required,min=1,max=100"`
	Description   string           `json:"description,omitempty"`
	UnitOfMeasure string           `json:"unit_of_measure" binding:"required,min=1,max=20"`
	ItemType      models.ItemType  `json:"item_type" binding:"required"` // Validated against enum
	IsActive      bool             `json:"is_active"`                    // Defaults to true if omitted
}

// UpdateItemRequest defines the structure for updating an existing item.
type UpdateItemRequest struct {
	Name          *string          `json:"name,omitempty" binding:"omitempty,min=1,max=100"`
	Description   *string          `json:"description,omitempty"`
	UnitOfMeasure *string          `json:"unit_of_measure,omitempty" binding:"omitempty,min=1,max=20"`
	ItemType      *models.ItemType `json:"item_type,omitempty"` // Validated against enum
	IsActive      *bool            `json:"is_active,omitempty"`
	// SKU is typically not updatable after creation to maintain integrity.
}

// ListItemRequest defines parameters for listing items.
type ListItemRequest struct {
	Page      int             `form:"page,default=1"`
	Limit     int             `form:"limit,default=20"`
	Name      string          `form:"name,omitempty"`
	SKU       string          `form:"sku,omitempty"`
	ItemType  models.ItemType `form:"item_type,omitempty"`
	IsActive  *bool           `form:"is_active,omitempty"` // Pointer to differentiate not set, true, false
}


// --- Warehouse DTOs ---

// CreateWarehouseRequest defines the structure for creating a new warehouse.
type CreateWarehouseRequest struct {
	Code     string `json:"code" binding:"required,min=1,max=20"`
	Name     string `json:"name" binding:"required,min=1,max=100"`
	Location string `json:"location,omitempty" binding:"max=255"`
	IsActive bool   `json:"is_active"` // Defaults to true
}

// UpdateWarehouseRequest defines the structure for updating an existing warehouse.
type UpdateWarehouseRequest struct {
	Name     *string `json:"name,omitempty" binding:"omitempty,min=1,max=100"`
	Location *string `json:"location,omitempty" binding:"omitempty,max=255"`
	IsActive *bool   `json:"is_active,omitempty"`
	// Code is typically not updatable.
}

// ListWarehouseRequest defines parameters for listing warehouses.
type ListWarehouseRequest struct {
	Page     int    `form:"page,default=1"`
	Limit    int    `form:"limit,default=20"`
	Name     string `form:"name,omitempty"`
	Code     string `form:"code,omitempty"`
	Location string `form:"location,omitempty"`
	IsActive *bool  `form:"is_active,omitempty"`
}


// --- Inventory Transaction (Adjustment) DTOs ---

// CreateInventoryAdjustmentRequest defines the structure for creating an inventory adjustment.
// This will result in one or more InventoryTransaction records.
type CreateInventoryAdjustmentRequest struct {
	WarehouseID     uuid.UUID                        `json:"warehouse_id" binding:"required"`
	ItemID          uuid.UUID                        `json:"item_id" binding:"required"`
	AdjustmentType  models.InventoryTransactionType `json:"adjustment_type" binding:"required"` // e.g., ADJUST_STOCK_IN, ADJUST_STOCK_OUT
	Quantity        float64                          `json:"quantity" binding:"required,gt=0"`   // Always positive, type determines effect
	TransactionDate *time.Time                       `json:"transaction_date,omitempty"`         // Defaults to Now
	Notes           string                           `json:"notes,omitempty"`
	ReferenceID     *uuid.UUID                       `json:"reference_id,omitempty"` // Optional link to a document causing adjustment
}


// --- Inventory Level DTOs ---

// InventoryLevelRequest defines parameters for querying inventory levels.
type InventoryLevelRequest struct {
	ItemID      *uuid.UUID `form:"item_id,omitempty"`      // Filter by specific item
	WarehouseID *uuid.UUID `form:"warehouse_id,omitempty"` // Filter by specific warehouse
	AsOfDate    *time.Time `form:"as_of_date,omitempty"`   // Defaults to Now
	// Page/Limit if listing many items' levels
}

// ItemStockLevelInfo represents stock level for an item in a specific warehouse.
type ItemStockLevelInfo struct {
	ItemID        uuid.UUID  `json:"item_id"`
	ItemSKU       string     `json:"item_sku"`
	ItemName      string     `json:"item_name"`
	WarehouseID   uuid.UUID  `json:"warehouse_id"`
	WarehouseCode string     `json:"warehouse_code"`
	WarehouseName string     `json:"warehouse_name"`
	Quantity      float64    `json:"quantity"`
	AsOfDate      time.Time  `json:"as_of_date"`
}

// InventoryLevelsResponse can be a list of ItemStockLevelInfo or a more complex structure.
type InventoryLevelsResponse struct {
	Levels   []ItemStockLevelInfo `json:"levels"`
	AsOfDate time.Time            `json:"as_of_date"`
	// Add pagination if applicable
}

// InventoryValuationReportRequest (Placeholder)
type InventoryValuationReportRequest struct {
    AsOfDate    time.Time  `json:"as_of_date" binding:"required"`
    WarehouseID *uuid.UUID `json:"warehouse_id,omitempty"` // Optional: filter by warehouse
    // CostingMethod string // e.g., FIFO, LIFO, Average (for future)
}

// InventoryValuationLine (Placeholder)
type InventoryValuationLine struct {
    ItemID      uuid.UUID `json:"item_id"`
    ItemSKU     string    `json:"item_sku"`
    ItemName    string    `json:"item_name"`
    Quantity    float64   `json:"quantity"`
    UnitCost    float64   `json:"unit_cost"` // This is the complex part
    TotalValue  float64   `json:"total_value"`
}

// InventoryValuationReportResponse (Placeholder)
type InventoryValuationReportResponse struct {
    ReportDate      time.Time                `json:"report_date"`
    WarehouseName   string                   `json:"warehouse_name,omitempty"` // If filtered by warehouse
    Lines           []InventoryValuationLine `json:"lines"`
    TotalValue      float64                  `json:"total_inventory_value"`
}
