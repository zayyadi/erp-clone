package repository_test

import (
	"context"
	"erp-system/configs"
	accModels "erp-system/internal/accounting/models" // Accounting models for full schema if needed by FKs from inventory
	"erp-system/internal/inventory/models"           // Inventory models
	"erp-system/pkg/database"
	"erp-system/pkg/logger"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/gorm"
)

var (
	dbInstance_inventory *gorm.DB // Suffix to distinguish from accounting's global if they were in same scope
	testConfig_inventory configs.AppConfig
)

// setupInventoryTestDB sets up a PostgreSQL test container for inventory tests.
func setupInventoryTestDB(ctx context.Context) (*gorm.DB, configs.AppConfig, func(), error) {
	logger.InfoLogger.Println("Setting up INVENTORY test database container...")

	dbName := "erp_inv_test_" + strings.ReplaceAll(uuid.New().String(), "-", "")[:8]
	dbUser := "testinvuser"
	dbPassword := "testinvpassword"

	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(1*time.Minute)),
	)
	if err != nil {
		return nil, configs.AppConfig{}, nil, fmt.Errorf("failed to start PostgreSQL container for inventory: %w", err)
	}

	host, err := pgContainer.Host(ctx)
	if err != nil { pgContainer.Terminate(ctx); return nil, configs.AppConfig{}, nil, fmt.Errorf("failed to get inv container host: %w", err) }
	port, err := pgContainer.MappedPort(ctx, "5432/tcp")
	if err != nil { pgContainer.Terminate(ctx); return nil, configs.AppConfig{}, nil, fmt.Errorf("failed to get inv container port: %w", err) }

	logger.InfoLogger.Printf("Inventory Test container started. Host: %s, Port: %s, DBName: %s", host, port.Port(), dbName)

	cfg := configs.AppConfig{
		DBHost: host, DBPort: port.Port(), DBUser: dbUser, DBPassword: dbPassword, DBName: dbName, SSLMode: "disable",
	}

	gormDB, err := database.InitTestDB(cfg) // Using the test DB initializer
	if err != nil {
		pgContainer.Terminate(ctx)
		return nil, configs.AppConfig{}, nil, fmt.Errorf("failed to connect to inventory test database: %w", err)
	}

	logger.InfoLogger.Println("Connected to inventory test database. Running migrations for all modules...")

	// AutoMigrate all known schemas: Accounting and Inventory
	// This ensures foreign key constraints between modules can be satisfied if any.
	err = gormDB.AutoMigrate(
		// Accounting Models (needed if inventory tables have FKs to them, or just for complete schema)
		&accModels.ChartOfAccount{},
		&accModels.JournalEntry{},
		&accModels.JournalLine{},
		// Inventory Models
		&models.Item{},
		&models.Warehouse{},
		&models.InventoryTransaction{},
	)
	if err != nil {
		sqlDB, _ := gormDB.DB(); if sqlDB != nil { sqlDB.Close() }; pgContainer.Terminate(ctx)
		return nil, configs.AppConfig{}, nil, fmt.Errorf("failed to run GORM auto-migrations for inventory test DB: %w", err)
	}
	logger.InfoLogger.Println("GORM Auto-migrations completed for inventory test DB.")

	cleanupFunc := func() {
		logger.InfoLogger.Println("Terminating INVENTORY test database container...")
		sqlDB, _ := gormDB.DB(); if sqlDB != nil { sqlDB.Close() }
		if err := pgContainer.Terminate(ctx); err != nil {
			log.Fatalf("Failed to terminate inventory test container: %v", err)
		}
		logger.InfoLogger.Println("Inventory test database container terminated.")
	}

	return gormDB, cfg, cleanupFunc, nil
}

// TestMain for inventory repository tests.
func TestMain(m *testing.M) {
	// Note: testing.Short() check moved to individual TestXxx suite functions
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var err error
	var cleanup func()

	dbInstance_inventory, testConfig_inventory, cleanup, err = setupInventoryTestDB(ctx)
	if err != nil {
		log.Fatalf("Failed to set up test database for package inventory_repository_test: %v", err)
	}

	exitCode := m.Run()

	if cleanup != nil {
		cleanup()
	}
	os.Exit(exitCode)
}

// resetInventoryTables truncates tables relevant to inventory tests.
// It should also include any tables from other modules if there are FK dependencies.
func resetInventoryTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	// Order: transactions first, then master data.
	err := db.Exec("TRUNCATE TABLE inventory_transactions CASCADE").Error
	assert.NoError(t, err, "Failed to truncate inventory_transactions")

	err = db.Exec("TRUNCATE TABLE items CASCADE").Error
	assert.NoError(t, err, "Failed to truncate items")

	err = db.Exec("TRUNCATE TABLE warehouses CASCADE").Error
	assert.NoError(t, err, "Failed to truncate warehouses")

	// If accounting tables were also used directly in these inventory tests (e.g. for FKs)
	// and not just for schema setup, they might need resetting too.
	// For now, assuming inventory tests primarily focus on inventory tables after initial schema setup.
	// If `integration_test_setup_test.go` from accounting package was used, its `resetTables` would be more general.
	// Since this is a new setup file, this reset is specific.
	logger.InfoLogger.Printf("Inventory tables truncated for test: %s", t.Name())
}
