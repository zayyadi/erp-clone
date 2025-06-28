package repository

import (
	"context"
	"erp-system/internal/inventory/models"
	"erp-system/pkg/errors"
	"erp-system/pkg/logger"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// InventoryTransactionRepository defines the interface for database operations for InventoryTransaction.
type InventoryTransactionRepository interface {
	Create(ctx context.Context, transaction *models.InventoryTransaction) (*models.InventoryTransaction, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.InventoryTransaction, error)
	List(ctx context.Context, offset, limit int, filters map[string]interface{}) ([]*models.InventoryTransaction, int64, error)
	GetStockLevel(ctx context.Context, itemID uuid.UUID, warehouseID uuid.UUID, date time.Time) (float64, error)
	GetStockLevelsByItem(ctx context.Context, itemID uuid.UUID, date time.Time) (map[uuid.UUID]float64, error) // map[WarehouseID]StockLevel
	GetStockLevelsByWarehouse(ctx context.Context, warehouseID uuid.UUID, date time.Time) (map[uuid.UUID]float64, error) // map[ItemID]StockLevel
}

// gormInventoryTransactionRepository is an implementation of InventoryTransactionRepository using GORM.
type gormInventoryTransactionRepository struct {
	db *gorm.DB
}

// NewInventoryTransactionRepository creates a new GORM-based InventoryTransactionRepository.
func NewInventoryTransactionRepository(db *gorm.DB) InventoryTransactionRepository {
	return &gormInventoryTransactionRepository{db: db}
}

// Create adds a new inventory transaction to the database.
func (r *gormInventoryTransactionRepository) Create(ctx context.Context, transaction *models.InventoryTransaction) (*models.InventoryTransaction, error) {
	logger.InfoLogger.Printf("Repository: Attempting to create inventory transaction for item %s, type %s", transaction.ItemID, transaction.TransactionType)
	if err := r.db.WithContext(ctx).Create(transaction).Error; err != nil {
		logger.ErrorLogger.Printf("Repository: Error creating inventory transaction: %v", err)
		return nil, errors.NewInternalServerError("failed to create inventory transaction", err)
	}
	// Preload Item and Warehouse for the created transaction response if needed by service/handler
	if err := r.db.WithContext(ctx).Preload("Item").Preload("Warehouse").First(transaction, "id = ?", transaction.ID).Error; err != nil {
        logger.ErrorLogger.Printf("Repository: Error preloading Item/Warehouse for created transaction %s: %v", transaction.ID, err)
        // Non-fatal, return the transaction without preloads or handle as critical error
    }
	logger.InfoLogger.Printf("Repository: Successfully created inventory transaction with ID: %s", transaction.ID)
	return transaction, nil
}

// GetByID retrieves an inventory transaction by its ID.
func (r *gormInventoryTransactionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.InventoryTransaction, error) {
	logger.InfoLogger.Printf("Repository: Attempting to retrieve inventory transaction with ID: %s", id)
	var transaction models.InventoryTransaction
	if err := r.db.WithContext(ctx).Preload("Item").Preload("Warehouse").First(&transaction, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.WarnLogger.Printf("Repository: Inventory transaction with ID %s not found", id)
			return nil, errors.NewNotFoundError("inventory_transaction", id.String())
		}
		logger.ErrorLogger.Printf("Repository: Error retrieving inventory transaction by ID %s: %v", id, err)
		return nil, errors.NewInternalServerError(fmt.Sprintf("failed to get inventory transaction by ID %s", id), err)
	}
	logger.InfoLogger.Printf("Repository: Successfully retrieved inventory transaction with ID: %s", transaction.ID)
	return &transaction, nil
}

// List retrieves a list of inventory transactions with pagination and optional filters.
func (r *gormInventoryTransactionRepository) List(ctx context.Context, offset, limit int, filters map[string]interface{}) ([]*models.InventoryTransaction, int64, error) {
	logger.InfoLogger.Printf("Repository: Listing inventory transactions with offset: %d, limit: %d, filters: %v", offset, limit, filters)
	var transactions []*models.InventoryTransaction
	var total int64

	query := r.db.WithContext(ctx).Model(&models.InventoryTransaction{})

	// Apply filters
	if itemID, ok := filters["item_id"].(uuid.UUID); ok && itemID != uuid.Nil {
		query = query.Where("item_id = ?", itemID)
	}
	if warehouseID, ok := filters["warehouse_id"].(uuid.UUID); ok && warehouseID != uuid.Nil {
		query = query.Where("warehouse_id = ?", warehouseID)
	}
	if transactionType, ok := filters["transaction_type"].(models.InventoryTransactionType); ok && transactionType != "" {
		query = query.Where("transaction_type = ?", transactionType)
	}
	if dateFrom, ok := filters["date_from"].(time.Time); ok && !dateFrom.IsZero() {
		query = query.Where("transaction_date >= ?", dateFrom)
	}
	if dateTo, ok := filters["date_to"].(time.Time); ok && !dateTo.IsZero() {
		query = query.Where("transaction_date <= ?", dateTo)
	}
	if referenceID, ok := filters["reference_id"].(uuid.UUID); ok && referenceID != uuid.Nil {
		query = query.Where("reference_id = ?", referenceID)
	}


	if err := query.Count(&total).Error; err != nil {
		logger.ErrorLogger.Printf("Repository: Error counting inventory transactions: %v", err)
		return nil, 0, errors.NewInternalServerError("failed to count inventory transactions", err)
	}

	if limit > 0 {
		query = query.Offset(offset).Limit(limit)
	} else {
		query = query.Offset(offset)
	}

	// Preload Item and Warehouse details for the list
	query = query.Preload("Item").Preload("Warehouse").Order("transaction_date desc, created_at desc")

	if err := query.Find(&transactions).Error; err != nil {
		logger.ErrorLogger.Printf("Repository: Error listing inventory transactions: %v", err)
		return nil, 0, errors.NewInternalServerError("failed to list inventory transactions", err)
	}

	logger.InfoLogger.Printf("Repository: Successfully listed %d inventory transactions, total count: %d", len(transactions), total)
	return transactions, total, nil
}


// GetStockLevel calculates the current stock level for a specific item in a specific warehouse up to a given date.
func (r *gormInventoryTransactionRepository) GetStockLevel(ctx context.Context, itemID uuid.UUID, warehouseID uuid.UUID, date time.Time) (float64, error) {
	logger.InfoLogger.Printf("Repository: Calculating stock level for item %s in warehouse %s as of %s", itemID, warehouseID, date.Format("2006-01-02"))
	var totalStock float64

	type StockMovement struct {
		Quantity        float64
		TransactionType models.InventoryTransactionType
	}
	var movements []StockMovement

	err := r.db.WithContext(ctx).Model(&models.InventoryTransaction{}).
		Select("quantity", "transaction_type").
		Where("item_id = ? AND warehouse_id = ? AND transaction_date <= ?", itemID, warehouseID, date).
		Find(&movements).Error

	if err != nil {
		logger.ErrorLogger.Printf("Repository: Error fetching stock movements for item %s, warehouse %s: %v", itemID, warehouseID, err)
		return 0, errors.NewInternalServerError("failed to calculate stock level", err)
	}

	for _, movement := range movements {
		tempTransaction := models.InventoryTransaction{Quantity: movement.Quantity, TransactionType: movement.TransactionType}
		totalStock += tempTransaction.Quantity * float64(tempTransaction.GetEffectOnStock())
	}

	logger.InfoLogger.Printf("Repository: Calculated stock level for item %s, warehouse %s is %.3f", itemID, warehouseID, totalStock)
	return totalStock, nil
}

// GetStockLevelsByItem calculates stock levels for a given item across all warehouses up to a given date.
func (r *gormInventoryTransactionRepository) GetStockLevelsByItem(ctx context.Context, itemID uuid.UUID, date time.Time) (map[uuid.UUID]float64, error) {
	logger.InfoLogger.Printf("Repository: Calculating stock levels for item %s across warehouses as of %s", itemID, date.Format("2006-01-02"))

	type Result struct {
		WarehouseID     uuid.UUID
		TransactionType models.InventoryTransactionType
		Quantity        float64
	}
	var results []Result

	err := r.db.WithContext(ctx).Model(&models.InventoryTransaction{}).
		Select("warehouse_id, transaction_type, quantity").
		Where("item_id = ? AND transaction_date <= ?", itemID, date).
		Find(&results).Error

	if err != nil {
		logger.ErrorLogger.Printf("Repository: Error fetching stock movements for item %s: %v", itemID, err)
		return nil, errors.NewInternalServerError("failed to calculate stock levels by item", err)
	}

	stockLevels := make(map[uuid.UUID]float64)
	for _, res := range results {
		tempTransaction := models.InventoryTransaction{Quantity: res.Quantity, TransactionType: res.TransactionType}
		stockLevels[res.WarehouseID] += tempTransaction.Quantity * float64(tempTransaction.GetEffectOnStock())
	}

	logger.InfoLogger.Printf("Repository: Successfully calculated stock levels for item %s: %v", itemID, stockLevels)
	return stockLevels, nil
}


// GetStockLevelsByWarehouse calculates stock levels for all items in a given warehouse up to a given date.
func (r *gormInventoryTransactionRepository) GetStockLevelsByWarehouse(ctx context.Context, warehouseID uuid.UUID, date time.Time) (map[uuid.UUID]float64, error) {
	logger.InfoLogger.Printf("Repository: Calculating stock levels for warehouse %s across items as of %s", warehouseID, date.Format("2006-01-02"))

	type Result struct {
		ItemID          uuid.UUID
		TransactionType models.InventoryTransactionType
		Quantity        float64
	}
	var results []Result

	err := r.db.WithContext(ctx).Model(&models.InventoryTransaction{}).
		Select("item_id, transaction_type, quantity").
		Where("warehouse_id = ? AND transaction_date <= ?", warehouseID, date).
		Find(&results).Error

	if err != nil {
		logger.ErrorLogger.Printf("Repository: Error fetching stock movements for warehouse %s: %v", warehouseID, err)
		return nil, errors.NewInternalServerError("failed to calculate stock levels by warehouse", err)
	}

	stockLevels := make(map[uuid.UUID]float64)
	for _, res := range results {
		tempTransaction := models.InventoryTransaction{Quantity: res.Quantity, TransactionType: res.TransactionType}
		stockLevels[res.ItemID] += tempTransaction.Quantity * float64(tempTransaction.GetEffectOnStock())
	}

	logger.InfoLogger.Printf("Repository: Successfully calculated stock levels for warehouse %s: %v", warehouseID, stockLevels)
	return stockLevels, nil
}

// Note: Inventory transactions are typically immutable once created. Updates might involve creating reversing/correcting entries.
// A direct Update method for InventoryTransaction is usually not provided or is highly restricted.
// A Delete method is also typically not provided; transactions are reversed.
// For this reason, Update and Delete methods are omitted from this repository.
// If specific business cases require modifying certain fields of a transaction (e.g., notes, reference after creation),
// a specific, restricted Update method could be added.
