package repository_test

import (
	"context"
	"erp-system/configs"
	accModels "erp-system/internal/accounting/models" // Accounting models
	// No need to import inventory models here if accounting repo tests don't directly depend on them for FKs
	// If they do, then inventory models should be imported and migrated.
	// For strict separation, only migrate tables relevant to the package being tested.
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
	// Suffix _acc_repo to distinguish if it were ever in a truly global scope with other test DBs
	dbInstance_acc_repo *gorm.DB
	testConfig_acc_repo configs.AppConfig
)

// setupAccRepoTestDB sets up a PostgreSQL test container for accounting repository tests.
func setupAccRepoTestDB(ctx context.Context) (*gorm.DB, configs.AppConfig, func(), error) {
	logger.InfoLogger.Println("Setting up ACCOUNTING REPOSITORY test database container...")

	dbName := "erp_accrepo_test_" + strings.ReplaceAll(uuid.New().String(), "-", "")[:6]
	dbUser := "testaccrepouser"
	dbPassword := "testaccrepopass"

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
		return nil, configs.AppConfig{}, nil, fmt.Errorf("pg container start for acc repo: %w", err)
	}

	host, _ := pgContainer.Host(ctx)
	port, _ := pgContainer.MappedPort(ctx, "5432/tcp")
	cfg := configs.AppConfig{DBHost: host, DBPort: port.Port(), DBUser: dbUser, DBPassword: dbPassword, DBName: dbName, SSLMode: "disable"}

	gormDB, err := database.InitTestDB(cfg)
	if err != nil {
		pgContainer.Terminate(ctx)
		return nil, configs.AppConfig{}, nil, fmt.Errorf("gorm init for acc repo: %w", err)
	}

	logger.InfoLogger.Println("Connected to acc repo test database. Running migrations...")

	// Only migrate accounting models for accounting repository tests
	err = gormDB.AutoMigrate(
		&accModels.ChartOfAccount{},
		&accModels.JournalEntry{},
		&accModels.JournalLine{},
	)
	if err != nil {
		sqlDB, _ := gormDB.DB(); if sqlDB != nil { sqlDB.Close() }; pgContainer.Terminate(ctx)
		return nil, configs.AppConfig{}, nil, fmt.Errorf("automigrate for acc repo: %w", err)
	}
	logger.InfoLogger.Println("GORM Auto-migrations completed for acc repo test DB.")

	cleanupFunc := func() {
		logger.InfoLogger.Println("Terminating ACCOUNTING REPOSITORY test database container...")
		sqlDB, _ := gormDB.DB(); if sqlDB != nil { sqlDB.Close() }
		if err := pgContainer.Terminate(ctx); err != nil {
			log.Fatalf("Failed to terminate acc repo test container: %v", err)
		}
		logger.InfoLogger.Println("Acc repo test database container terminated.")
	}

	return gormDB, cfg, cleanupFunc, nil
}

// TestMain for accounting repository tests.
func TestMain(m *testing.M) {
	// REMOVED: if testing.Short() check, as it causes panic if flags not parsed.
	// The check should be in individual TestXxx functions that run suites.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute) // Shorter timeout for module specific
	defer cancel()

	var cleanup func()
	var err error
	dbInstance_acc_repo, testConfig_acc_repo, cleanup, err = setupAccRepoTestDB(ctx)
	if err != nil {
		log.Fatalf("FATAL: Failed to set up test environment for accounting repository tests: %v", err)
	}

	exitCode := m.Run()
	if cleanup != nil { cleanup() }
	os.Exit(exitCode)
}

// resetAccRepoTables truncates tables relevant to accounting repository tests.
func resetAccRepoTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	tables := []string{"journal_lines", "journal_entries", "chart_of_accounts"}
	// Truncate in reverse order of creation or consider FKs
	for _, table := range tables {
		err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)).Error
		assert.NoError(t, err, "Failed to truncate table %s for acc repo tests", table)
	}
	logger.InfoLogger.Printf("Accounting repository tables truncated for test: %s", t.Name())
}
