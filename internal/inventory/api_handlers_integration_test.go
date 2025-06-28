package inventory_test // Belongs to a package specific to inventory integration tests

import (
	"bytes"
	"context"
	"encoding/json"
	"erp-system/api" // For NewRouter
	"erp-system/configs"
	accModels "erp-system/internal/accounting/models"      // For full schema setup
	invModels "erp-system/internal/inventory/models"       // Actual models being tested
	invRepo "erp-system/internal/inventory/repository"     // For direct seeding if needed
	invServiceDTO "erp-system/internal/inventory/service/dto" // DTOs for requests/responses
	"erp-system/pkg/database"
	"erp-system/pkg/logger"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert" // Keep assert for direct assertions if needed outside suite
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/gorm"
)

var (
	dbInstance_inventory_api *gorm.DB
	testConfig_inventory_api configs.AppConfig
	globalRouter_inventory   *mux.Router // Global router for inventory API tests
)

// setupInventoryAPITestDB is similar to other setups but used for API handler tests for inventory.
func setupInventoryAPITestDB(ctx context.Context) (*gorm.DB, configs.AppConfig, *mux.Router, func(), error) {
	logger.InfoLogger.Println("Setting up INVENTORY API test database container...")
	dbName := "erp_inv_api_test_" + strings.ReplaceAll(uuid.New().String(), "-", "")[:6]
	dbUser, dbPassword := "testinvapiuser", "testinvapipass"

	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgres.WithDatabase(dbName), postgres.WithUsername(dbUser), postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(1*time.Minute)),
	)
	if err != nil { return nil, configs.AppConfig{}, nil, nil, fmt.Errorf("pg container start: %w", err) }

	host, _ := pgContainer.Host(ctx)
	port, _ := pgContainer.MappedPort(ctx, "5432/tcp")
	cfg := configs.AppConfig{DBHost: host, DBPort: port.Port(), DBUser: dbUser, DBPassword: dbPassword, DBName: dbName, SSLMode: "disable"}

	gormDB, err := database.InitTestDB(cfg)
	if err != nil { pgContainer.Terminate(ctx); return nil, configs.AppConfig{}, nil, nil, fmt.Errorf("gorm init: %w", err) }

	// Migrate all known schemas
	err = gormDB.AutoMigrate(
		&accModels.ChartOfAccount{}, &accModels.JournalEntry{}, &accModels.JournalLine{},
		&invModels.Item{}, &invModels.Warehouse{}, &invModels.InventoryTransaction{},
	)
	if err != nil { sqlDB, _ := gormDB.DB(); sqlDB.Close(); pgContainer.Terminate(ctx); return nil, configs.AppConfig{}, nil, nil, fmt.Errorf("automigrate: %w", err)}

	router := api.NewRouter(gormDB) // Initialize router with this specific DB for inventory API tests

	cleanupFunc := func() {
		logger.InfoLogger.Println("Terminating INVENTORY API test database container...")
		sqlDB, _ := gormDB.DB(); if sqlDB != nil { sqlDB.Close() }
		pgContainer.Terminate(ctx)
	}
	return gormDB, cfg, router, cleanupFunc, nil
}

// TestMain for inventory API integration tests.
func TestMain(m *testing.M) {
	// testing.Short() check moved to TestInventoryAPIHandlersIntegration
	// This TestMain will manage the lifecycle for tests in this specific file/package.
	// if testing.Short() { // This was the problematic line, already removed in creation.
	// 	log.Println("Skipping API integration tests in short mode.")
	// 	os.Exit(0)
	// }
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var cleanup func()
	var err error
	dbInstance_inventory_api, testConfig_inventory_api, globalRouter_inventory, cleanup, err = setupInventoryAPITestDB(ctx)
	if err != nil {
		log.Fatalf("FATAL: Failed to set up test environment for inventory API tests: %v", err)
	}

	exitCode := m.Run()
	if cleanup != nil { cleanup() }
	os.Exit(exitCode)
}


// InventoryAPIHandlersIntegrationTestSuite defines the suite for API handler integration tests.
type InventoryAPIHandlersIntegrationTestSuite struct {
	suite.Suite
	db     *gorm.DB
	router *mux.Router
	itemRepo invRepo.ItemRepository
	whRepo   invRepo.WarehouseRepository
}

func (s *InventoryAPIHandlersIntegrationTestSuite) SetupSuite() {
	s.T().Log("Setting up suite for Inventory API Handlers integration tests...")
	s.db = dbInstance_inventory_api
	s.router = globalRouter_inventory
	s.itemRepo = invRepo.NewItemRepository(s.db)
	s.whRepo = invRepo.NewWarehouseRepository(s.db)
	s.T().Log("Inventory API Handlers suite setup complete.")
}

func (s *InventoryAPIHandlersIntegrationTestSuite) SetupTest() {
	s.T().Logf("Setting up test: %s", s.T().Name())
	tables := []string{"inventory_transactions", "items", "warehouses", "journal_lines", "journal_entries", "chart_of_accounts"}
	for _, table := range tables {
		err := s.db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)).Error
		s.Require().NoError(err, "Failed to truncate table %s", table)
	}
	s.T().Logf("Test setup complete for: %s (tables truncated)", s.T().Name())
}

func TestInventoryAPIHandlersIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping inventory API integration tests in short mode.")
	}
	suite.Run(t, new(InventoryAPIHandlersIntegrationTestSuite))
}

func (s *InventoryAPIHandlersIntegrationTestSuite) performRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer = bytes.NewBuffer(nil)
	if body != nil {
		b, err := json.Marshal(body)
		s.Require().NoError(err)
		reqBody = bytes.NewBuffer(b)
	}
	req, err := http.NewRequest(method, path, reqBody)
	s.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

// --- Item API Tests ---
func (s *InventoryAPIHandlersIntegrationTestSuite) TestCreateItemAPI() {
	payload := invServiceDTO.CreateItemRequest{
		SKU: "API_ITEM001", Name: "API Test Item", UnitOfMeasure: "PCS", ItemType: invModels.FinishedGood, IsActive: true,
	}
	rr := s.performRequest("POST", "/api/v1/inventory/items", payload)
	s.Equal(http.StatusCreated, rr.Code)

	var response invServiceDTO.SuccessResponse
	var createdItem invModels.Item
	err := json.Unmarshal(rr.Body.Bytes(), &response)
    s.Require().NoError(err)
	dataBytes, err := json.Marshal(response.Data)
    s.Require().NoError(err)
	err = json.Unmarshal(dataBytes, &createdItem)
    s.Require().NoError(err)

	s.Equal(payload.SKU, createdItem.SKU)
	s.NotEmpty(createdItem.ID)
}

func (s *InventoryAPIHandlersIntegrationTestSuite) TestGetItemBySKUAPI() {
	seededItem, err := s.itemRepo.Create(context.Background(), &invModels.Item{
		SKU: "API_SKU001", Name: "Fetch By SKU", UnitOfMeasure: "EA", ItemType: invModels.RawMaterial, IsActive: true,
	})
    s.Require().NoError(err)

	rr := s.performRequest("GET", fmt.Sprintf("/api/v1/inventory/items/sku/%s", seededItem.SKU), nil)
	s.Equal(http.StatusOK, rr.Code)

	var response invServiceDTO.SuccessResponse
	var fetchedItem invModels.Item
	err = json.Unmarshal(rr.Body.Bytes(), &response)
    s.Require().NoError(err)
	dataBytes, err := json.Marshal(response.Data)
    s.Require().NoError(err)
	err = json.Unmarshal(dataBytes, &fetchedItem)
    s.Require().NoError(err)

	s.Equal(seededItem.ID, fetchedItem.ID)
	s.Equal(seededItem.Name, fetchedItem.Name)
}


// --- Warehouse API Tests ---
func (s *InventoryAPIHandlersIntegrationTestSuite) TestCreateWarehouseAPI() {
    payload := invServiceDTO.CreateWarehouseRequest{
        Code: "API_WH001", Name: "API Main Warehouse", Location: "API Test Location", IsActive: true,
    }
    rr := s.performRequest("POST", "/api/v1/inventory/warehouses", payload)
    s.Equal(http.StatusCreated, rr.Code)

    var response invServiceDTO.SuccessResponse
    var createdWarehouse invModels.Warehouse
    err := json.Unmarshal(rr.Body.Bytes(), &response)
    s.Require().NoError(err)
    dataBytes, err := json.Marshal(response.Data)
    s.Require().NoError(err)
    err = json.Unmarshal(dataBytes, &createdWarehouse)
    s.Require().NoError(err)

    s.Equal(payload.Code, createdWarehouse.Code)
    s.NotEmpty(createdWarehouse.ID)
}


// --- Inventory Adjustment API Tests ---
func (s *InventoryAPIHandlersIntegrationTestSuite) TestCreateInventoryAdjustmentAPI() {
    item, errItem := s.itemRepo.Create(context.Background(), &invModels.Item{SKU: "ADJ_ITEM_API", Name: "Adj Item API", UnitOfMeasure: "PCS", ItemType: invModels.FinishedGood, IsActive: true})
    s.Require().NoError(errItem)
    wh, errWh := s.whRepo.Create(context.Background(), &invModels.Warehouse{Code: "ADJ_WH_API", Name: "Adj WH API", IsActive: true})
    s.Require().NoError(errWh)

    payload := invServiceDTO.CreateInventoryAdjustmentRequest{
        ItemID: item.ID, WarehouseID: wh.ID, Quantity: 10, AdjustmentType: invModels.AdjustStockIn, Notes: "API Adjustment In",
    }
    rr := s.performRequest("POST", "/api/v1/inventory/adjustments", payload)
    s.Equal(http.StatusCreated, rr.Code)

    var response invServiceDTO.SuccessResponse
    var createdTxn invModels.InventoryTransaction
    err := json.Unmarshal(rr.Body.Bytes(), &response)
    s.Require().NoError(err)
    dataBytes, err := json.Marshal(response.Data)
    s.Require().NoError(err)
    err = json.Unmarshal(dataBytes, &createdTxn)
    s.Require().NoError(err)

    s.Equal(item.ID, createdTxn.ItemID)
    s.Equal(wh.ID, createdTxn.WarehouseID)
    s.Equal(float64(10), createdTxn.Quantity)
    s.Equal(invModels.AdjustStockIn, createdTxn.TransactionType)
    s.NotEmpty(createdTxn.ID)
}

// --- Inventory Level API Tests ---
func (s *InventoryAPIHandlersIntegrationTestSuite) TestGetInventoryLevelsAPI_SpecificItemWarehouse() {
    item, _ := s.itemRepo.Create(context.Background(), &invModels.Item{SKU: "LVL_ITEM_API", Name: "Level Item API", UnitOfMeasure: "PCS", ItemType: invModels.FinishedGood, IsActive: true})
    wh, _ := s.whRepo.Create(context.Background(), &invModels.Warehouse{Code: "LVL_WH_API", Name: "Level WH API", IsActive: true})

    invTxnRepo := invRepo.NewInventoryTransactionRepository(s.db)
    _, errTxn := invTxnRepo.Create(context.Background(), &invModels.InventoryTransaction{
        ItemID: item.ID, WarehouseID: wh.ID, Quantity: 25, TransactionType: invModels.ReceiveStock, TransactionDate: time.Now(),
    })
    s.Require().NoError(errTxn)


    path := fmt.Sprintf("/api/v1/inventory/levels?item_id=%s&warehouse_id=%s", item.ID, wh.ID)
    rr := s.performRequest("GET", path, nil)
    s.Equal(http.StatusOK, rr.Code)

    var response invServiceDTO.SuccessResponse
    var levelsResponse invServiceDTO.InventoryLevelsResponse
    err := json.Unmarshal(rr.Body.Bytes(), &response)
    s.Require().NoError(err)
    dataBytes, err := json.Marshal(response.Data)
    s.Require().NoError(err)
    err = json.Unmarshal(dataBytes, &levelsResponse)
    s.Require().NoError(err)

    s.Require().Len(levelsResponse.Levels, 1)
    s.Equal(item.ID, levelsResponse.Levels[0].ItemID)
    s.Equal(wh.ID, levelsResponse.Levels[0].WarehouseID)
    s.Equal(float64(25), levelsResponse.Levels[0].Quantity)
}

func (s *InventoryAPIHandlersIntegrationTestSuite) TestGetSpecificItemStockLevelAPI() {
    item, _ := s.itemRepo.Create(context.Background(), &invModels.Item{SKU: "SPEC_ITEM_API", Name: "Specific Level Item API", UnitOfMeasure: "PCS", ItemType: invModels.FinishedGood, IsActive: true})
    wh, _ := s.whRepo.Create(context.Background(), &invModels.Warehouse{Code: "SPEC_WH_API", Name: "Specific Level WH API", IsActive: true})

    invTxnRepo := invRepo.NewInventoryTransactionRepository(s.db)
    _, errTxn := invTxnRepo.Create(context.Background(), &invModels.InventoryTransaction{
        ItemID: item.ID, WarehouseID: wh.ID, Quantity: 33, TransactionType: invModels.AdjustStockIn, TransactionDate: time.Now(),
    })
    s.Require().NoError(errTxn)

    path := fmt.Sprintf("/api/v1/inventory/levels/item/%s/warehouse/%s", item.ID, wh.ID)
    rr := s.performRequest("GET", path, nil)
    s.Equal(http.StatusOK, rr.Code)

    var response invServiceDTO.SuccessResponse
    var levelInfo invServiceDTO.ItemStockLevelInfo
    err := json.Unmarshal(rr.Body.Bytes(), &response)
    s.Require().NoError(err)
    dataBytes, err := json.Marshal(response.Data)
    s.Require().NoError(err)
    err = json.Unmarshal(dataBytes, &levelInfo)
    s.Require().NoError(err)

    s.Equal(item.ID, levelInfo.ItemID)
    s.Equal(wh.ID, levelInfo.WarehouseID)
    s.Equal(float64(33), levelInfo.Quantity)
    s.Equal(item.SKU, levelInfo.ItemSKU)
    s.Equal(wh.Code, levelInfo.WarehouseCode)
}
