package repository

import (
	"context"
	"erp-system/internal/accounting/models"
	"erp-system/pkg/errors"
	"erp-system/pkg/logger"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ChartOfAccountRepository defines the interface for database operations for ChartOfAccount.
type ChartOfAccountRepository interface {
	Create(ctx context.Context, account *models.ChartOfAccount) (*models.ChartOfAccount, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.ChartOfAccount, error)
	GetByCode(ctx context.Context, code string) (*models.ChartOfAccount, error)
	Update(ctx context.Context, account *models.ChartOfAccount) (*models.ChartOfAccount, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, offset, limit int, filters map[string]interface{}) ([]*models.ChartOfAccount, int64, error)
}

// gormChartOfAccountRepository is an implementation of ChartOfAccountRepository using GORM.
type gormChartOfAccountRepository struct {
	db *gorm.DB
}

// NewChartOfAccountRepository creates a new GORM-based ChartOfAccountRepository.
func NewChartOfAccountRepository(db *gorm.DB) ChartOfAccountRepository {
	return &gormChartOfAccountRepository{db: db}
}

// Create adds a new chart of account to the database.
func (r *gormChartOfAccountRepository) Create(ctx context.Context, account *models.ChartOfAccount) (*models.ChartOfAccount, error) {
	logger.InfoLogger.Printf("Repository: Attempting to create chart of account with code: %s", account.AccountCode)
	if err := r.db.WithContext(ctx).Create(account).Error; err != nil {
		logger.ErrorLogger.Printf("Repository: Error creating chart of account: %v", err)
		return nil, errors.NewInternalServerError("failed to create chart of account", err)
	}
	logger.InfoLogger.Printf("Repository: Successfully created chart of account with ID: %s", account.ID)
	return account, nil
}

// GetByID retrieves a chart of account by its ID.
func (r *gormChartOfAccountRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.ChartOfAccount, error) {
	logger.InfoLogger.Printf("Repository: Attempting to retrieve chart of account with ID: %s", id)
	var account models.ChartOfAccount
	if err := r.db.WithContext(ctx).First(&account, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.WarnLogger.Printf("Repository: Chart of account with ID %s not found", id)
			return nil, errors.NewNotFoundError("chart_of_account", id.String())
		}
		logger.ErrorLogger.Printf("Repository: Error retrieving chart of account by ID %s: %v", id, err)
		return nil, errors.NewInternalServerError(fmt.Sprintf("failed to get chart of account by ID %s", id), err)
	}
	logger.InfoLogger.Printf("Repository: Successfully retrieved chart of account with ID: %s", account.ID)
	return &account, nil
}

// GetByCode retrieves a chart of account by its account code.
func (r *gormChartOfAccountRepository) GetByCode(ctx context.Context, code string) (*models.ChartOfAccount, error) {
	logger.InfoLogger.Printf("Repository: Attempting to retrieve chart of account with code: %s", code)
	var account models.ChartOfAccount
	if err := r.db.WithContext(ctx).First(&account, "account_code = ?", code).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.WarnLogger.Printf("Repository: Chart of account with code %s not found", code)
			return nil, errors.NewNotFoundError("chart_of_account_code", code)
		}
		logger.ErrorLogger.Printf("Repository: Error retrieving chart of account by code %s: %v", code, err)
		return nil, errors.NewInternalServerError(fmt.Sprintf("failed to get chart of account by code %s", code), err)
	}
	logger.InfoLogger.Printf("Repository: Successfully retrieved chart of account with ID: %s for code: %s", account.ID, code)
	return &account, nil
}

// Update modifies an existing chart of account in the database.
func (r *gormChartOfAccountRepository) Update(ctx context.Context, account *models.ChartOfAccount) (*models.ChartOfAccount, error) {
	logger.InfoLogger.Printf("Repository: Attempting to update chart of account with ID: %s", account.ID)
	if err := r.db.WithContext(ctx).Save(account).Error; err != nil {
		logger.ErrorLogger.Printf("Repository: Error updating chart of account %s: %v", account.ID, err)
		return nil, errors.NewInternalServerError(fmt.Sprintf("failed to update chart of account %s", account.ID), err)
	}
	logger.InfoLogger.Printf("Repository: Successfully updated chart of account with ID: %s", account.ID)
	return account, nil
}

// Delete removes a chart of account from the database (soft delete if DeletedAt is configured).
func (r *gormChartOfAccountRepository) Delete(ctx context.Context, id uuid.UUID) error {
	logger.InfoLogger.Printf("Repository: Attempting to delete chart of account with ID: %s", id)
	if err := r.db.WithContext(ctx).Delete(&models.ChartOfAccount{}, id).Error; err != nil {
		logger.ErrorLogger.Printf("Repository: Error deleting chart of account %s: %v", id, err)
		return errors.NewInternalServerError(fmt.Sprintf("failed to delete chart of account %s", id), err)
	}
	logger.InfoLogger.Printf("Repository: Successfully deleted chart of account with ID: %s", id)
	return nil
}

// List retrieves a list of chart of accounts with pagination and optional filters.
func (r *gormChartOfAccountRepository) List(ctx context.Context, offset, limit int, filters map[string]interface{}) ([]*models.ChartOfAccount, int64, error) {
	logger.InfoLogger.Printf("Repository: Listing chart of accounts with offset: %d, limit: %d, filters: %v", offset, limit, filters)
	var accounts []*models.ChartOfAccount
	var total int64

	query := r.db.WithContext(ctx).Model(&models.ChartOfAccount{})

	// Apply filters
	if name, ok := filters["account_name"].(string); ok && name != "" {
		query = query.Where("account_name ILIKE ?", "%"+name+"%")
	}
	if accType, ok := filters["account_type"].(models.AccountType); ok && accType != "" {
		query = query.Where("account_type = ?", accType)
	}
    if isActive, ok := filters["is_active"].(bool); ok {
        query = query.Where("is_active = ?", isActive)
    }


	if err := query.Count(&total).Error; err != nil {
		logger.ErrorLogger.Printf("Repository: Error counting chart of accounts: %v", err)
		return nil, 0, errors.NewInternalServerError("failed to count chart of accounts", err)
	}

	if limit > 0 {
		query = query.Offset(offset).Limit(limit)
	} else { // if limit is 0 or negative, retrieve all matching records (up to a reasonable max if necessary)
		query = query.Offset(offset)
	}

	query = query.Order("account_code asc") // Default ordering

	if err := query.Find(&accounts).Error; err != nil {
		logger.ErrorLogger.Printf("Repository: Error listing chart of accounts: %v", err)
		return nil, 0, errors.NewInternalServerError("failed to list chart of accounts", err)
	}

	logger.InfoLogger.Printf("Repository: Successfully listed %d chart of accounts, total count: %d", len(accounts), total)
	return accounts, total, nil
}
