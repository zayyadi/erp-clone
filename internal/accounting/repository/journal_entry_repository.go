package repository

import (
	"context"
	"erp-system/internal/accounting/models"
	"erp-system/pkg/errors"
	"erp-system/pkg/logger"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	// "gorm.io/gorm/clause" // Not used in the recreated version below, but was in previous attempt
)

// JournalEntryRepository defines the interface for database operations for JournalEntry.
type JournalEntryRepository interface {
	Create(ctx context.Context, entry *models.JournalEntry) (*models.JournalEntry, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.JournalEntry, error)
	Update(ctx context.Context, entry *models.JournalEntry) (*models.JournalEntry, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, offset, limit int, filters map[string]interface{}) ([]*models.JournalEntry, int64, error)
	UpdateJournalEntryStatus(ctx context.Context, id uuid.UUID, newStatus models.JournalStatus) error
	GetJournalLinesByEntryID(ctx context.Context, journalID uuid.UUID) ([]models.JournalLine, error)
	AddJournalLine(ctx context.Context, journalID uuid.UUID, line *models.JournalLine) (*models.JournalLine, error)
	RemoveJournalLine(ctx context.Context, lineID uuid.UUID) error
	UpdateJournalLine(ctx context.Context, line *models.JournalLine) (*models.JournalLine, error)
	GetJournalEntriesForTrialBalance(ctx context.Context, startDate, endDate time.Time) ([]models.JournalEntry, error)
	GetJournalEntriesByAccountID(ctx context.Context, accountID uuid.UUID, offset, limit int, startDate, endDate time.Time) ([]*models.JournalEntry, int64, error)
}

// gormJournalEntryRepository is an implementation of JournalEntryRepository using GORM.
type gormJournalEntryRepository struct {
	db *gorm.DB
}

// NewJournalEntryRepository creates a new GORM-based JournalEntryRepository.
func NewJournalEntryRepository(db *gorm.DB) JournalEntryRepository {
	return &gormJournalEntryRepository{db: db}
}

func (r *gormJournalEntryRepository) Create(ctx context.Context, entry *models.JournalEntry) (*models.JournalEntry, error) {
	logger.InfoLogger.Printf("Repository: Attempting to create journal entry with description: %s", entry.Description)
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(entry).Error; err != nil { // This creates header and lines if associations are set up
			logger.ErrorLogger.Printf("Repository: Error creating journal entry (and lines): %v", err)
			return err
		}
		return nil
	})

	if err != nil {
		logger.ErrorLogger.Printf("Repository: Transaction failed for creating journal entry: %v", err)
		return nil, errors.NewInternalServerError("failed to create journal entry in transaction", err)
	}

	// Reload to ensure all data (like preloaded lines with their own DB-generated fields) is fresh.
	if err := r.db.WithContext(ctx).Preload("JournalLines").First(entry, "id = ?", entry.ID).Error; err != nil {
		logger.ErrorLogger.Printf("Repository: Error reloading journal entry with lines after creation: %v", err)
		return nil, errors.NewInternalServerError("failed to reload journal entry after creation", err)
	}

	logger.InfoLogger.Printf("Repository: Successfully created journal entry ID: %s with %d lines", entry.ID, len(entry.JournalLines))
	return entry, nil
}

func (r *gormJournalEntryRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.JournalEntry, error) {
	logger.InfoLogger.Printf("Repository: Attempting to retrieve journal entry with ID: %s", id)
	var entry models.JournalEntry
	if err := r.db.WithContext(ctx).Preload("JournalLines").Preload("JournalLines.ChartOfAccount").First(&entry, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.WarnLogger.Printf("Repository: Journal entry with ID %s not found", id)
			return nil, errors.NewNotFoundError("journal_entry", id.String())
		}
		logger.ErrorLogger.Printf("Repository: Error retrieving journal entry by ID %s: %v", id, err)
		return nil, errors.NewInternalServerError(fmt.Sprintf("failed to get journal entry by ID %s", id), err)
	}
	logger.InfoLogger.Printf("Repository: Successfully retrieved journal entry with ID: %s", entry.ID)
	return &entry, nil
}

func (r *gormJournalEntryRepository) Update(ctx context.Context, entry *models.JournalEntry) (*models.JournalEntry, error) {
	logger.InfoLogger.Printf("Repository: Attempting to update journal entry with ID: %s", entry.ID)
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Save the main entry fields. Using Select("*") to ensure all fields are updated, including zero values if intended.
		// Or, use .Updates() with a map for partial updates if only specific fields should change.
		// For full replacement including associations, GORM's Save is powerful.
		if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).Save(entry).Error; err != nil {
			logger.ErrorLogger.Printf("Repository: Error saving journal entry (and lines) %s: %v", entry.ID, err)
			return err
		}
		return nil
	})

	if err != nil {
		logger.ErrorLogger.Printf("Repository: Transaction failed for updating journal entry %s: %v", entry.ID, err)
		return nil, errors.NewInternalServerError(fmt.Sprintf("failed to update journal entry %s in transaction", entry.ID), err)
	}
	// Reload to get fresh data post-update
	return r.GetByID(ctx, entry.ID)
}

func (r *gormJournalEntryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	logger.InfoLogger.Printf("Repository: Attempting to (soft) delete journal entry with ID: %s", id)
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Manually "delete" lines if entry is soft-deleted and cascade doesn't handle it for soft deletes.
		// As JournalLine has no DeletedAt, this means hard delete.
		if err := tx.Where("journal_id = ?", id).Delete(&models.JournalLine{}).Error; err != nil {
			logger.ErrorLogger.Printf("Repository: Error deleting journal lines for entry %s during soft delete: %v", id, err)
			return err
		}
		// Soft delete the journal entry itself
		if err := tx.Delete(&models.JournalEntry{}, id).Error; err != nil {
			logger.ErrorLogger.Printf("Repository: Error soft-deleting journal entry %s: %v", id, err)
			return err
		}
		return nil
	})

	if err != nil {
		logger.ErrorLogger.Printf("Repository: Transaction failed for deleting journal entry %s: %v", id, err)
		return errors.NewInternalServerError(fmt.Sprintf("failed to delete journal entry %s", id), err)
	}
	logger.InfoLogger.Printf("Repository: Successfully (soft) deleted journal entry with ID: %s and its lines", id)
	return nil
}

func (r *gormJournalEntryRepository) List(ctx context.Context, offset, limit int, filters map[string]interface{}) ([]*models.JournalEntry, int64, error) {
	var entries []*models.JournalEntry
	var total int64
	query := r.db.WithContext(ctx).Model(&models.JournalEntry{})

	if desc, ok := filters["description"].(string); ok && desc != "" { query = query.Where("description ILIKE ?", "%"+desc+"%") }
	if ref, ok := filters["reference"].(string); ok && ref != "" { query = query.Where("reference ILIKE ?", "%"+ref+"%") }
	if status, ok := filters["status"].(models.JournalStatus); ok && status != "" { query = query.Where("status = ?", status) }
	if dateFrom, ok := filters["date_from"].(time.Time); ok && !dateFrom.IsZero() { query = query.Where("entry_date >= ?", dateFrom) }
	if dateTo, ok := filters["date_to"].(time.Time); ok && !dateTo.IsZero() { query = query.Where("entry_date <= ?", dateTo) }

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.NewInternalServerError("failed to count journal entries", err)
	}
	if limit > 0 { query = query.Offset(offset).Limit(limit) } else { query = query.Offset(offset) }

	err := query.Preload("JournalLines").Preload("JournalLines.ChartOfAccount").Order("entry_date desc, created_at desc").Find(&entries).Error
	if err != nil {
		return nil, 0, errors.NewInternalServerError("failed to list journal entries", err)
	}
	return entries, total, nil
}

func (r *gormJournalEntryRepository) UpdateJournalEntryStatus(ctx context.Context, id uuid.UUID, newStatus models.JournalStatus) error {
	result := r.db.WithContext(ctx).Model(&models.JournalEntry{}).Where("id = ?", id).Update("status", newStatus)
	if result.Error != nil {
		return errors.NewInternalServerError(fmt.Sprintf("failed to update status for journal entry %s", id), result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.NewNotFoundError("journal_entry", id.String())
	}
	return nil
}

func (r *gormJournalEntryRepository) GetJournalLinesByEntryID(ctx context.Context, journalID uuid.UUID) ([]models.JournalLine, error) {
	var lines []models.JournalLine
	err := r.db.WithContext(ctx).Where("journal_id = ?", journalID).Order("created_at asc").Find(&lines).Error
	if err != nil {
		return nil, errors.NewInternalServerError(fmt.Sprintf("failed to get lines for journal %s", journalID), err)
	}
	return lines, nil
}

func (r *gormJournalEntryRepository) AddJournalLine(ctx context.Context, journalID uuid.UUID, line *models.JournalLine) (*models.JournalLine, error) {
	line.JournalID = journalID
	if err := r.db.WithContext(ctx).Create(line).Error; err != nil {
		return nil, errors.NewInternalServerError("failed to add journal line", err)
	}
	return line, nil
}

func (r *gormJournalEntryRepository) RemoveJournalLine(ctx context.Context, lineID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&models.JournalLine{}, lineID).Error; err != nil {
		return errors.NewInternalServerError(fmt.Sprintf("failed to remove journal line %s", lineID), err)
	}
	return nil
}

func (r *gormJournalEntryRepository) UpdateJournalLine(ctx context.Context, line *models.JournalLine) (*models.JournalLine, error) {
	// Using Save for full update, ensure line.ID is set.
	if err := r.db.WithContext(ctx).Save(line).Error; err != nil {
		return nil, errors.NewInternalServerError(fmt.Sprintf("failed to update journal line %s", line.ID), err)
	}
	return line, nil
}

func (r *gormJournalEntryRepository) GetJournalEntriesForTrialBalance(ctx context.Context, startDate, endDate time.Time) ([]models.JournalEntry, error) {
	var entries []models.JournalEntry
	err := r.db.WithContext(ctx).
		Preload("JournalLines").
		Preload("JournalLines.ChartOfAccount").
		Where("status = ? AND entry_date BETWEEN ? AND ?", models.StatusPosted, startDate, endDate).
		Order("entry_date asc").
		Find(&entries).Error
	if err != nil {
		return nil, errors.NewInternalServerError("failed to fetch entries for trial balance", err)
	}
	return entries, nil
}

func (r *gormJournalEntryRepository) GetJournalEntriesByAccountID(ctx context.Context, accountID uuid.UUID, offset, limit int, startDate, endDate time.Time) ([]*models.JournalEntry, int64, error) {
	var entries []*models.JournalEntry
	var total int64

	// Subquery to find journal_ids that have a line with the specified account_id
	subQuery := r.db.Model(&models.JournalLine{}).Select("journal_id").Where("account_id = ?", accountID)

	query := r.db.WithContext(ctx).Model(&models.JournalEntry{}).Where("id IN (?)", subQuery)
	if !startDate.IsZero() { query = query.Where("entry_date >= ?", startDate) }
	if !endDate.IsZero() { query = query.Where("entry_date <= ?", endDate) }

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.NewInternalServerError("failed to count journal entries for account", err)
	}
	if limit > 0 { query = query.Offset(offset).Limit(limit) } else { query = query.Offset(offset) }

	err := query.Preload("JournalLines").Preload("JournalLines.ChartOfAccount").Order("entry_date desc, created_at desc").Find(&entries).Error
	if err != nil {
		return nil, 0, errors.NewInternalServerError("failed to list journal entries for account", err)
	}
	return entries, total, nil
}
