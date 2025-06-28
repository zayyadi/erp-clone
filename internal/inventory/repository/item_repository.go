package repository

import (
	"context"
	"erp-system/internal/inventory/models"
	"erp-system/pkg/errors"
	"erp-system/pkg/logger"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ItemRepository defines the interface for database operations for Item.
type ItemRepository interface {
	Create(ctx context.Context, item *models.Item) (*models.Item, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Item, error)
	GetBySKU(ctx context.Context, sku string) (*models.Item, error)
	Update(ctx context.Context, item *models.Item) (*models.Item, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, offset, limit int, filters map[string]interface{}) ([]*models.Item, int64, error)
	// Add other specific methods if needed, e.g., FindByName, FindByType
}

// gormItemRepository is an implementation of ItemRepository using GORM.
type gormItemRepository struct {
	db *gorm.DB
}

// NewItemRepository creates a new GORM-based ItemRepository.
func NewItemRepository(db *gorm.DB) ItemRepository {
	return &gormItemRepository{db: db}
}

// Create adds a new item to the database.
func (r *gormItemRepository) Create(ctx context.Context, item *models.Item) (*models.Item, error) {
	logger.InfoLogger.Printf("Repository: Attempting to create item with SKU: %s", item.SKU)
	if err := r.db.WithContext(ctx).Create(item).Error; err != nil {
		// Check for unique constraint violation on SKU (driver specific error)
		// For PostgreSQL, unique violation error code is 23505
		// This check is a bit fragile as it depends on error message strings or codes.
		// A more robust way is to attempt GetBySKU first if performance allows, or handle specific DB errors.
		// if strings.Contains(err.Error(), "duplicate key value violates unique constraint") && strings.Contains(err.Error(), "items_sku_key") {
		// 	logger.WarnLogger.Printf("Repository: Item with SKU %s already exists.", item.SKU)
		// 	return nil, errors.NewConflictError(fmt.Sprintf("item with SKU %s already exists", item.SKU))
		// }
		logger.ErrorLogger.Printf("Repository: Error creating item: %v", err)
		return nil, errors.NewInternalServerError("failed to create item", err)
	}
	logger.InfoLogger.Printf("Repository: Successfully created item with ID: %s", item.ID)
	return item, nil
}

// GetByID retrieves an item by its ID.
func (r *gormItemRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Item, error) {
	logger.InfoLogger.Printf("Repository: Attempting to retrieve item with ID: %s", id)
	var item models.Item
	if err := r.db.WithContext(ctx).First(&item, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.WarnLogger.Printf("Repository: Item with ID %s not found", id)
			return nil, errors.NewNotFoundError("item", id.String())
		}
		logger.ErrorLogger.Printf("Repository: Error retrieving item by ID %s: %v", id, err)
		return nil, errors.NewInternalServerError(fmt.Sprintf("failed to get item by ID %s", id), err)
	}
	logger.InfoLogger.Printf("Repository: Successfully retrieved item with ID: %s", item.ID)
	return &item, nil
}

// GetBySKU retrieves an item by its SKU.
func (r *gormItemRepository) GetBySKU(ctx context.Context, sku string) (*models.Item, error) {
	logger.InfoLogger.Printf("Repository: Attempting to retrieve item with SKU: %s", sku)
	var item models.Item
	if err := r.db.WithContext(ctx).First(&item, "sku = ?", sku).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.WarnLogger.Printf("Repository: Item with SKU %s not found", sku)
			return nil, errors.NewNotFoundError("item_sku", sku)
		}
		logger.ErrorLogger.Printf("Repository: Error retrieving item by SKU %s: %v", sku, err)
		return nil, errors.NewInternalServerError(fmt.Sprintf("failed to get item by SKU %s", sku), err)
	}
	logger.InfoLogger.Printf("Repository: Successfully retrieved item with ID: %s for SKU: %s", item.ID, sku)
	return &item, nil
}

// Update modifies an existing item in the database.
func (r *gormItemRepository) Update(ctx context.Context, item *models.Item) (*models.Item, error) {
	logger.InfoLogger.Printf("Repository: Attempting to update item with ID: %s", item.ID)
	// Ensure SKU uniqueness if SKU is being changed (though SKU is often immutable post-creation).
	// If SKU can be changed, a check similar to Create is needed:
	// var existing models.Item
	// if err := r.db.WithContext(ctx).Where("sku = ? AND id != ?", item.SKU, item.ID).First(&existing).Error; err == nil {
	//    logger.WarnLogger.Printf("Repository: Another item with SKU %s already exists.", item.SKU)
	//	  return nil, errors.NewConflictError(fmt.Sprintf("another item with SKU %s already exists", item.SKU))
	// } else if err != gorm.ErrRecordNotFound {
	//    logger.ErrorLogger.Printf("Repository: Error checking SKU uniqueness during update for item %s: %v", item.ID, err)
	//    return nil, errors.NewInternalServerError("failed to check SKU uniqueness during update", err)
	// }

	if err := r.db.WithContext(ctx).Save(item).Error; err != nil {
		// Handle potential unique constraint violation on SKU if it was changed and conflicts
		logger.ErrorLogger.Printf("Repository: Error updating item %s: %v", item.ID, err)
		return nil, errors.NewInternalServerError(fmt.Sprintf("failed to update item %s", item.ID), err)
	}
	logger.InfoLogger.Printf("Repository: Successfully updated item with ID: %s", item.ID)
	return item, nil
}

// Delete removes an item from the database (soft delete if DeletedAt is configured).
func (r *gormItemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	logger.InfoLogger.Printf("Repository: Attempting to delete item with ID: %s", id)
	// Business logic: before deleting an item, check if it has any stock or is part of open transactions.
	// This should ideally be in the service layer.
	// For now, repository just performs the delete action.
	if err := r.db.WithContext(ctx).Delete(&models.Item{}, id).Error; err != nil {
		logger.ErrorLogger.Printf("Repository: Error deleting item %s: %v", id, err)
		return errors.NewInternalServerError(fmt.Sprintf("failed to delete item %s", id), err)
	}
	logger.InfoLogger.Printf("Repository: Successfully deleted item with ID: %s", id)
	return nil
}

// List retrieves a list of items with pagination and optional filters.
func (r *gormItemRepository) List(ctx context.Context, offset, limit int, filters map[string]interface{}) ([]*models.Item, int64, error) {
	logger.InfoLogger.Printf("Repository: Listing items with offset: %d, limit: %d, filters: %v", offset, limit, filters)
	var items []*models.Item
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Item{})

	// Apply filters
	if name, ok := filters["name"].(string); ok && name != "" {
		query = query.Where("name ILIKE ?", "%"+name+"%")
	}
	if sku, ok := filters["sku"].(string); ok && sku != "" {
		query = query.Where("sku ILIKE ?", "%"+sku+"%")
	}
	if itemType, ok := filters["item_type"].(models.ItemType); ok && itemType != "" {
		query = query.Where("item_type = ?", itemType)
	}
	if isActive, ok := filters["is_active"].(bool); ok { // Direct bool check
		query = query.Where("is_active = ?", isActive)
	}
    // If isActive is a *bool in filters to distinguish between not set, true, false:
    // if isActivePtr, ok := filters["is_active"].(*bool); ok && isActivePtr != nil {
	// 	query = query.Where("is_active = ?", *isActivePtr)
	// }


	if err := query.Count(&total).Error; err != nil {
		logger.ErrorLogger.Printf("Repository: Error counting items: %v", err)
		return nil, 0, errors.NewInternalServerError("failed to count items", err)
	}

	if limit > 0 {
		query = query.Offset(offset).Limit(limit)
	} else { // if limit is 0 or negative, retrieve all matching records
		query = query.Offset(offset) // Potentially add a hard cap if needed
	}

	query = query.Order("sku asc") // Default ordering

	if err := query.Find(&items).Error; err != nil {
		logger.ErrorLogger.Printf("Repository: Error listing items: %v", err)
		return nil, 0, errors.NewInternalServerError("failed to list items", err)
	}

	logger.InfoLogger.Printf("Repository: Successfully listed %d items, total count: %d", len(items), total)
	return items, total, nil
}
