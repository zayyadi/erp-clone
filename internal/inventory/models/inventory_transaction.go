package models

import (
	"time"
	// "github.com/shopspring/decimal" // For precise quantity if needed
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// InventoryTransactionType represents the nature of an inventory movement.
type InventoryTransactionType string

const (
	ReceiveStock      InventoryTransactionType = "RECEIVE_STOCK"       // From Purchase Order, Transfer In, Initial Stock
	IssueStock        InventoryTransactionType = "ISSUE_STOCK"         // For Sales Order, Transfer Out, Production Consumption
	AdjustStockIn     InventoryTransactionType = "ADJUST_STOCK_IN"     // Positive adjustment (e.g., found stock, cycle count up)
	AdjustStockOut    InventoryTransactionType = "ADJUST_STOCK_OUT"    // Negative adjustment (e.g., damaged stock, cycle count down, scrap)
	TransferOut       InventoryTransactionType = "TRANSFER_OUT"       // Stock moving from this warehouse to another (source)
	TransferIn        InventoryTransactionType = "TRANSFER_IN"         // Stock moving into this warehouse from another (destination)
	ProductionOutput  InventoryTransactionType = "PRODUCTION_OUTPUT"   // Finished goods from production
	ProductionConsume InventoryTransactionType = "PRODUCTION_CONSUME" // Raw materials consumed by production
	SalesReturn       InventoryTransactionType = "SALES_RETURN"       // Customer returns stock
	PurchaseReturn    InventoryTransactionType = "PURCHASE_RETURN"     // Returning stock to vendor
)

// InventoryTransaction records movements of items in and out of warehouses.
type InventoryTransaction struct {
	ID               uuid.UUID                `gorm:"type:uuid;primary_key;" json:"id"`
	ItemID           uuid.UUID                `gorm:"type:uuid;not null;index" json:"item_id"`     // Foreign key to Item
	WarehouseID      uuid.UUID                `gorm:"type:uuid;not null;index" json:"warehouse_id"` // Foreign key to Warehouse
	Quantity         float64                  `gorm:"type:numeric(10,3);not null" json:"quantity"` // Quantity of the transaction (positive for IN, can be positive for OUT and type defines direction)
	TransactionType  InventoryTransactionType `gorm:"type:varchar(30);not null;index" json:"transaction_type"`
	ReferenceID      *uuid.UUID               `gorm:"type:uuid;index" json:"reference_id,omitempty"`      // Optional: links to PO, SO, Adjustment ID, Transfer ID etc.
	TransactionDate  time.Time                `gorm:"not null;index" json:"transaction_date"`          // Actual date of the physical transaction
	Notes            string                   `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt        time.Time                `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time                `gorm:"autoUpdateTime" json:"updated_at"`
	// DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"` // Usually inventory transactions are not soft-deleted, but voided/reversed by counter-transactions.

	// Associations
	Item      *Item      `gorm:"foreignKey:ItemID;references:ID" json:"item,omitempty"`
	Warehouse *Warehouse `gorm:"foreignKey:WarehouseID;references:ID" json:"warehouse,omitempty"`

	// Potential future fields:
	// Cost                *decimal.Decimal `gorm:"type:numeric(15,2)" json:"cost,omitempty"` // Cost of goods for this transaction (relevant for valuation)
	// LotNumber           string           `gorm:"type:varchar(50);index" json:"lot_number,omitempty"`
	// SerialNumber        string           `gorm:"type:varchar(100);index" json:"serial_number,omitempty"` // If item is serialized
	// ExpiryDate          *time.Time       `json:"expiry_date,omitempty"`
	// RelatedTransactionID *uuid.UUID      `gorm:"type:uuid;index" json:"related_transaction_id,omitempty"` // e.g. links TransferOut to TransferIn
	// UserID              *uuid.UUID      `gorm:"type:uuid;index" json:"user_id,omitempty"` // User who performed transaction
}

// TableName specifies the table name for InventoryTransaction model.
func (InventoryTransaction) TableName() string {
	return "inventory_transactions"
}

// BeforeCreate will set a UUID for the new inventory transaction.
func (it *InventoryTransaction) BeforeCreate(tx *gorm.DB) (err error) {
	if it.ID == uuid.Nil {
		it.ID = uuid.New()
	}
	if it.TransactionDate.IsZero() {
		it.TransactionDate = time.Now()
	}
	if it.Quantity <= 0 { // Basic validation, quantity should be positive. Type indicates direction.
		return gorm.ErrInvalidData // Or custom error: "transaction quantity must be positive"
	}

	// Validate TransactionType
	switch it.TransactionType {
	case ReceiveStock, IssueStock, AdjustStockIn, AdjustStockOut, TransferOut, TransferIn, ProductionOutput, ProductionConsume, SalesReturn, PurchaseReturn:
		// valid type
	default:
		if string(it.TransactionType) == "" {
			return gorm.ErrInvalidData // Or custom error: "transaction type is required"
		}
		// Invalid type string provided
		return gorm.ErrInvalidData // Or custom error: fmt.Sprintf("invalid transaction type: %s", it.TransactionType)
	}
	return
}

// GetEffectOnStock returns 1 if the transaction increases stock, -1 if it decreases, 0 if neutral.
// This is a simplified view; some types might be more complex (e.g., transfers if not split into two records).
func (it *InventoryTransaction) GetEffectOnStock() int {
	switch it.TransactionType {
	case ReceiveStock, AdjustStockIn, TransferIn, ProductionOutput, SalesReturn:
		return 1 // Increases stock
	case IssueStock, AdjustStockOut, TransferOut, ProductionConsume, PurchaseReturn:
		return -1 // Decreases stock
	default:
		return 0 // Unknown or neutral type
	}
}
