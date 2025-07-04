package accounting_test // Reverted package declaration

import (
	"context"
	"erp-system/configs"
	"erp-system/internal/accounting/models" // For GORM auto-migration
	inventoryModels "erp-system/internal/inventory/models"
	"erp-system/pkg/database"
	"erp-system/pkg/logger"
	"github.com/stretchr/testify/assert"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/gorm"
)

var (
	dbInstance *gorm.DB
	testConfig configs.AppConfig
)

// setupTestDB sets up a PostgreSQL test container and initializes a GORM connection.
// It also runs migrations.
func setupTestDB(ctx context.Context) (*gorm.DB, configs.AppConfig, func(), error) {
	logger.InfoLogger.Println("Setting up test database container...")

	// Define a unique name for the test database to avoid conflicts if tests run in parallel (though usually not an issue with containers)
	dbName := "erp_test_" + strings.ReplaceAll(uuid.New().String(), "-", "")[:8]
	dbUser := "testuser"
	dbPassword := "testpassword"

	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		// postgres.WithInitScripts(filepath.Join("..", "..", "migrations", "0001_create_accounting_tables.up.sql")), // This runs scripts *inside* the container, relative to its FS or specified path.
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2). // Wait for the second occurrence for robustness
				WithStartupTimeout(1*time.Minute)),
	)
	if err != nil {
		return nil, configs.AppConfig{}, nil, fmt.Errorf("failed to start PostgreSQL container: %w", err)
	}

	// Get connection details
	host, err := pgContainer.Host(ctx)
	if err != nil {
		pgContainer.Terminate(ctx)
		return nil, configs.AppConfig{}, nil, fmt.Errorf("failed to get container host: %w", err)
	}
	port, err := pgContainer.MappedPort(ctx, "5432/tcp")
	if err != nil {
		pgContainer.Terminate(ctx)
		return nil, configs.AppConfig{}, nil, fmt.Errorf("failed to get container port: %w", err)
	}

	logger.InfoLogger.Printf("Test container started. Host: %s, Port: %s, DBName: %s", host, port.Port(), dbName)

	// Create a test configuration
	cfg := configs.AppConfig{
		DBHost:     host,
		DBPort:     port.Port(),
		DBUser:     dbUser,
		DBPassword: dbPassword,
		DBName:     dbName,
		SSLMode:    "disable", // Typically disable SSL for local test containers
		ServerPort: "0",       // Not used by repository tests directly
	}

	// Initialize GORM with the test database
	// Need to reset the singleton 'once' in database package for testing, or make InitDB more flexible.
	// For now, we'll create a new GORM instance directly for test isolation, bypassing the package's singleton.
	// dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
	// 	cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.SSLMode)
	// testDb, err := gorm.Open(gorm_postgres.Open(dsn), &gorm.Config{})

	// Let's try to use our existing database.InitDB but this is tricky with singletons.
	// A better approach is to have InitDB return a new instance always, or allow re-init for tests.
	// For now, using the existing InitDB from pkg/database.
	// This requires careful management if tests run in parallel or if InitDB has global state.
	// The 'once' in database.InitDB will prevent re-initialization.
	// So, we will need a way to get a fresh DB instance for tests.
	// Let's modify InitDB to be callable multiple times if a certain test flag is set, or make it return a new instance.
	// Simpler: create a new GORM instance directly for tests, bypassing the global `database.DB`.

	gormDB, err := database.InitTestDB(cfg) // We will create InitTestDB in pkg/database
	if err != nil {
		pgContainer.Terminate(ctx)
		return nil, configs.AppConfig{}, nil, fmt.Errorf("failed to connect to test database: %w", err)
	}

	logger.InfoLogger.Println("Connected to test database. Running migrations...")

	// Run migrations using GORM AutoMigrate (if SQL files are not used by testcontainers)
	// Or execute SQL migration files directly.
	// AutoMigrate is simpler for tests if models are the source of truth for schema.

	// Get all models from accounting and inventory (and future modules)
	allModels := []interface{}{
		&models.ChartOfAccount{},
		&models.JournalEntry{},
		&models.JournalLine{},
		// Add Inventory Models here once they are created in a shared place or imported
		// For now, assuming they are in a path accessible like:
		// &inventoryModels.Item{},
		// &inventoryModels.Warehouse{},
		// &inventoryModels.InventoryTransaction{},
	}

	// Dynamically add inventory models if they exist to avoid import cycle if they are in sub-packages of this test package
	// This is a bit of a workaround. Better to have models in distinct packages.
	// For this structure where tests are in `internal/accounting` but need `internal/inventory/models`:
	// We will explicitly add them.
	// This requires `inventoryModels "erp-system/internal/inventory/models"` import.
	// This will be added when the actual inventory test files are created in their respective package.
	// For now, this setup file is in `accounting_test` package.
	// The ideal setup would be a shared test helper package or running AutoMigrate within each module's test suite setup.

	// For now, let's assume this setup is generic enough.
	// When inventory tests run, their TestMain or SuiteSetup can add their models to this list or run their own AutoMigrate.
	// However, since dbInstance is global here, we should try to migrate all known schemas.
	// Let's assume we will add inventory models when we write inventory integration tests.
	// For now, this only migrates accounting.
	//
	// **Update:** Since we are now in the inventory module phase, let's add them.
	// This file is `internal/accounting/integration_test_setup_test.go`.
	// To access inventory models, we'd need `inventoryModels "erp-system/internal/inventory/models"`.
	// This creates a slight awkwardness as accounting tests might not need inventory models.
	// A shared test setup or module-specific setups are cleaner.
	//
	// Let's modify this to be a more general setup and include inventory models.
	// The `resetTables` function will also need to be updated.

	// inventoryModels "erp-system/internal/inventory/models" // This line is removed as it's now in the main import block


	err = gormDB.AutoMigrate(
		&models.ChartOfAccount{}, // Accounting model
		&models.JournalEntry{},   // Accounting model
		&models.JournalLine{},    // Accounting model
		&inventoryModels.Item{},
		&inventoryModels.Warehouse{},
		&inventoryModels.InventoryTransaction{},
	)
	if err != nil {
		sqlDB, _ := gormDB.DB()
		sqlDB.Close()
		pgContainer.Terminate(ctx)
		return nil, configs.AppConfig{}, nil, fmt.Errorf("failed to run GORM auto-migrations: %w", err)
	}
    logger.InfoLogger.Println("GORM Auto-migrations completed.")

	// Alternative: Execute SQL migration files
	// This requires locating the migration files relative to the test execution path.
	// _, b, _, _ := runtime.Caller(0)
    // projectRoot := filepath.Join(filepath.Dir(b), "..", "..", "..") // Adjust based on test file location
    // migrationFilePath := filepath.Join(projectRoot, "migrations", "0001_create_accounting_tables.up.sql")
	// logger.InfoLogger.Printf("Looking for migration file at: %s", migrationFilePath)

	// migrationSQL, err := os.ReadFile(migrationFilePath)
	// if err != nil {
	// 	sqlDB, _ := gormDB.DB()
	// 	if sqlDB != nil { sqlDB.Close() }
	// 	pgContainer.Terminate(ctx)
	// 	return nil, configs.AppConfig{}, nil, fmt.Errorf("failed to read migration file %s: %w", migrationFilePath, err)
	// }

	// if err := gormDB.Exec(string(migrationSQL)).Error; err != nil {
	// 	sqlDB, _ := gormDB.DB()
	// 	if sqlDB != nil { sqlDB.Close() }
	// 	pgContainer.Terminate(ctx)
	// 	return nil, configs.AppConfig{}, nil, fmt.Errorf("failed to execute migration SQL: %w", err)
	// }
	// logger.InfoLogger.Println("SQL Migrations executed successfully.")


	cleanupFunc := func() {
		logger.InfoLogger.Println("Terminating test database container...")
		sqlDB, _ := gormDB.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
		if err := pgContainer.Terminate(ctx); err != nil {
			log.Fatalf("Failed to terminate test container: %v", err)
		}
		logger.InfoLogger.Println("Test database container terminated.")
	}

	return gormDB, cfg, cleanupFunc, nil
}


// TestMain sets up the database for all tests in this package.
func TestMain(m *testing.M) {
	// Set up a global context for TestMain
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute) // Increased timeout for container setup
	defer cancel()

	var err error
	var cleanup func()

	// Setup test DB only once for the package
	// Note: This means tests within this package share the same DB instance state unless reset.
	// For fully isolated tests, setup/teardown should be per test or per suite.
	dbInstance, testConfig, cleanup, err = setupTestDB(ctx)
	if err != nil {
		log.Fatalf("Failed to set up test database for package accounting_test: %v", err)
	}

	// Run tests
	exitCode := m.Run()

	// Teardown
	if cleanup != nil {
		cleanup()
	}

	os.Exit(exitCode)
}

// Helper to reset database tables between tests if needed.
// This is crucial if tests within the same package share the dbInstance from TestMain.
func resetTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	// Order matters due to foreign key constraints.
	// Inventory transactions might reference items and warehouses.
	// Journal lines reference chart of accounts.

	// Inventory Module Tables
	err := db.Exec("TRUNCATE TABLE inventory_transactions CASCADE").Error
	assert.NoError(t, err, "Failed to truncate inventory_transactions")

	err = db.Exec("TRUNCATE TABLE items CASCADE").Error
	assert.NoError(t, err, "Failed to truncate items")

	err = db.Exec("TRUNCATE TABLE warehouses CASCADE").Error
	assert.NoError(t, err, "Failed to truncate warehouses")

	// Accounting Module Tables
	err = db.Exec("TRUNCATE TABLE journal_lines CASCADE").Error
	assert.NoError(t, err, "Failed to truncate journal_lines")

	err = db.Exec("TRUNCATE TABLE journal_entries CASCADE").Error
	assert.NoError(t, err, "Failed to truncate journal_entries")

	err = db.Exec("TRUNCATE TABLE chart_of_accounts CASCADE").Error
	assert.NoError(t, err, "Failed to truncate chart_of_accounts")

	// If using sequences that need resetting (e.g. for serial IDs, not UUIDs):
	// db.Exec("ALTER SEQUENCE chart_of_accounts_id_seq RESTART WITH 1")
	// etc. for other tables. Not needed for UUIDs.
	logger.InfoLogger.Printf("Tables truncated for test: %s", t.Name())
}

// findProjectRoot searches for the project root directory by looking for go.mod.
func findProjectRoot() (string, error) {
	_, b, _, _ := runtime.Caller(0)
	dir := filepath.Dir(b)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}
