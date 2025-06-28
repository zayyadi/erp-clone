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

// WarehouseRepository defines the interface for database operations for Warehouse.
type WarehouseRepository interface {
	Create(ctx context.Context, warehouse *models.Warehouse) (*models.Warehouse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Warehouse, error)
	GetByCode(ctx context.Context, code string) (*models.Warehouse, error)
	Update(ctx context.Context, warehouse *models.Warehouse) (*models.Warehouse, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, offset, limit int, filters map[string]interface{}) ([]*models.Warehouse, int64, error)
}

// gormWarehouseRepository is an implementation of WarehouseRepository using GORM.
type gormWarehouseRepository struct {
	db *gorm.DB
}

// NewWarehouseRepository creates a new GORM-based WarehouseRepository.
func NewWarehouseRepository(db *gorm.DB) WarehouseRepository {
	return &gormWarehouseRepository{db: db}
}

// Create adds a new warehouse to the database.
func (r *gormWarehouseRepository) Create(ctx context.Context, warehouse *models.Warehouse) (*models.Warehouse, error) {
	logger.InfoLogger.Printf("Repository: Attempting to create warehouse with code: %s", warehouse.Code)
	if err := r.db.WithContext(ctx).Create(warehouse).Error; err != nil {
		// Handle unique constraint violation on Code, similar to Item SKU
		logger.ErrorLogger.Printf("Repository: Error creating warehouse: %v", err)
		return nil, errors.NewInternalServerError("failed to create warehouse", err)
	}
	logger.InfoLogger.Printf("Repository: Successfully created warehouse with ID: %s", warehouse.ID)
	return warehouse, nil
}

// GetByID retrieves a warehouse by its ID.
func (r *gormWarehouseRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Warehouse, error) {
	logger.InfoLogger.Printf("Repository: Attempting to retrieve warehouse with ID: %s", id)
	var warehouse models.Warehouse
	if err := r.db.WithContext(ctx).First(&warehouse, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.WarnLogger.Printf("Repository: Warehouse with ID %s not found", id)
			return nil, errors.NewNotFoundError("warehouse", id.String())
		}
		logger.ErrorLogger.Printf("Repository: Error retrieving warehouse by ID %s: %v", id, err)
		return nil, errors.NewInternalServerError(fmt.Sprintf("failed to get warehouse by ID %s", id), err)
	}
	logger.InfoLogger.Printf("Repository: Successfully retrieved warehouse with ID: %s", warehouse.ID)
	return &warehouse, nil
}

// GetByCode retrieves a warehouse by its code.
func (r *gormWarehouseRepository) GetByCode(ctx context.Context, code string) (*models.Warehouse, error) {
	logger.InfoLogger.Printf("Repository: Attempting to retrieve warehouse with code: %s", code)
	var warehouse models.Warehouse
	if err := r.db.WithContext(ctx).First(&warehouse, "code = ?", code).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.WarnLogger.Printf("Repository: Warehouse with code %s not found", code)
			return nil, errors.NewNotFoundError("warehouse_code", code)
		}
		logger.ErrorLogger.Printf("Repository: Error retrieving warehouse by code %s: %v", code, err)
		return nil, errors.NewInternalServerError(fmt.Sprintf("failed to get warehouse by code %s", code), err)
	}
	logger.InfoLogger.Printf("Repository: Successfully retrieved warehouse with ID: %s for code: %s", warehouse.ID, code)
	return &warehouse, nil
}

// Update modifies an existing warehouse in the database.
func (r *gormWarehouseRepository) Update(ctx context.Context, warehouse *models.Warehouse) (*models.Warehouse, error) {
	logger.InfoLogger.Printf("Repository: Attempting to update warehouse with ID: %s", warehouse.ID)
	// Similar to Item, if Code can be updated, uniqueness check is needed.
	if err := r.db.WithContext(ctx).Save(warehouse).Error; err != nil {
		logger.ErrorLogger.Printf("Repository: Error updating warehouse %s: %v", warehouse.ID, err)
		return nil, errors.NewInternalServerError(fmt.Sprintf("failed to update warehouse %s", warehouse.ID), err)
	}
	logger.InfoLogger.Printf("Repository: Successfully updated warehouse with ID: %s", warehouse.ID)
	return warehouse, nil
}

// Delete removes a warehouse from the database (soft delete if DeletedAt is configured).
func (r *gormWarehouseRepository) Delete(ctx context.Context, id uuid.UUID) error {
	logger.InfoLogger.Printf("Repository: Attempting to delete warehouse with ID: %s", id)
	// Business logic: check if warehouse has stock or is part of open transactions (service layer).
	if err := r.db.WithContext(ctx).Delete(&models.Warehouse{}, id).Error; err != nil {
		logger.ErrorLogger.Printf("Repository: Error deleting warehouse %s: %v", id, err)
		return errors.NewInternalServerError(fmt.Sprintf("failed to delete warehouse %s", id), err)
	}
	logger.InfoLogger.Printf("Repository: Successfully deleted warehouse with ID: %s", id)
	return nil
}

// List retrieves a list of warehouses with pagination and optional filters.
func (r *gormWarehouseRepository) List(ctx context.Context, offset, limit int, filters map[string]interface{}) ([]*models.Warehouse, int64, error) {
	logger.InfoLogger.Printf("Repository: Listing warehouses with offset: %d, limit: %d, filters: %v", offset, limit, filters)
	var warehouses []*models.Warehouse
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Warehouse{})

	// Apply filters
	if name, ok := filters["name"].(string); ok && name != "" {
		query = query.Where("name ILIKE ?", "%"+name+"%")
	}
	if code, ok := filters["code"].(string); ok && code != "" {
		query = query.Where("code ILIKE ?", "%"+code+"%")
	}
	if location, ok := filters["location"].(string); ok && location != "" {
		query = query.Where("location ILIKE ?", "%"+location+"%")
	}
	if isActive, ok := filters["is_active"].(bool); ok {
		query = query.Where("is_active = ?", isActive)
	}

	if err := query.Count(&total).Error; err != nil {
		logger.ErrorLogger.Printf("Repository: Error counting warehouses: %v", err)
		return nil, 0, errors.NewInternalServerError("failed to count warehouses", err)
	}

	if limit > 0 {
		query = query.Offset(offset).Limit(limit)
	} else {
		query = query.Offset(offset)
	}

	query = query.Order("code asc") // Default ordering

	if err := query.Find(&warehouses).Error; err != nil {
		logger.ErrorLogger.Printf("Repository: Error listing warehouses: %v", err)
		return nil, 0, errors.NewInternalServerError("failed to list warehouses", err)
	}

	logger.InfoLogger.Printf("Repository: Successfully listed %d warehouses, total count: %d", len(warehouses), total)
	return warehouses, total, nil
}
