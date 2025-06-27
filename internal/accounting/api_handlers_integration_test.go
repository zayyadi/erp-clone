package accounting_test

// Note: This file is in package accounting_test, same as the setup and repo tests.

import (
	"bytes"
	"encoding/json"
	"erp-system/api" // For NewRouter
	"erp-system/internal/accounting/models"
	acc_repo "erp-system/internal/accounting/repository" // Alias to avoid conflict if any
	"erp-system/internal/accounting/service"
	dto "erp-system/internal/accounting/service/dto"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// APIHandlersIntegrationTestSuite defines the suite for API handler integration tests.
type APIHandlersIntegrationTestSuite struct {
	suite.Suite
	db     *gorm.DB
	router *mux.Router

	// Services and Repos (can be initialized here if needed, or rely on NewRouter to do it)
	// coaRepo acc_repo.ChartOfAccountRepository
	// journalRepo acc_repo.JournalEntryRepository
	// accService service.AccountingService
}

// SetupSuite runs once before all tests in the suite.
func (s *APIHandlersIntegrationTestSuite) SetupSuite() {
	s.T().Log("Setting up suite for API Handlers integration tests...")
	s.db = dbInstance // From integration_test_setup_test.go

	// Initialize the main application router. NewRouter sets up repos, services, and handlers.
	s.router = api.NewRouter(s.db)
	s.T().Log("API Handlers suite setup complete.")
}

// SetupTest runs before each test in the suite.
func (s *APIHandlersIntegrationTestSuite) SetupTest() {
	s.T().Logf("Setting up test: %s", s.T().Name())
	resetTables(s.T(), s.db) // Reset database tables
	s.T().Logf("Test setup complete for: %s", s.T().Name())
}

// TestAPIHandlersIntegration runs the entire suite.
func TestAPIHandlersIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode.")
	}
	t.Log("Starting APIHandlersIntegration Test Suite...")
	suite.Run(t, new(APIHandlersIntegrationTestSuite))
	t.Log("APIHandlersIntegration Test Suite finished.")
}

// Helper to make HTTP requests
func (s *APIHandlersIntegrationTestSuite) performRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		b, err := json.Marshal(body)
		s.Require().NoError(err, "Failed to marshal request body")
		reqBody = bytes.NewBuffer(b)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(method, path, reqBody)
	s.Require().NoError(err, "Failed to create HTTP request")
	req.Header.Set("Content-Type", "application/json")
	// Add Auth header if needed: req.Header.Set("Authorization", "Bearer test_token")

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

// --- Chart of Accounts API Tests ---

func (s *APIHandlersIntegrationTestSuite) TestCreateChartOfAccountAPI() {
	s.T().Log("Running TestCreateChartOfAccountAPI")
	payload := dto.CreateChartOfAccountRequest{
		AccountCode: "API1010",
		AccountName: "Cash via API",
		AccountType: models.Asset,
		IsActive:    true,
	}

	rr := s.performRequest("POST", "/api/v1/accounting/accounts", payload)
	s.Equal(http.StatusCreated, rr.Code, "Status code should be 201 Created")

	var response dto.SuccessResponse
	var createdAccount models.ChartOfAccount
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	s.NoError(err, "Failed to unmarshal success response")
	s.Equal("success", response.Status)

	// Convert response.Data (map[string]interface{}) to models.ChartOfAccount
	dataBytes, _ := json.Marshal(response.Data)
	err = json.Unmarshal(dataBytes, &createdAccount)
	s.NoError(err, "Failed to unmarshal account data from response")

	s.Equal(payload.AccountCode, createdAccount.AccountCode)
	s.NotEqual(uuid.Nil, createdAccount.ID)

	// Verify in DB
	var dbAccount models.ChartOfAccount
	err = s.db.First(&dbAccount, "account_code = ?", payload.AccountCode).Error
	s.NoError(err, "Account not found in DB after API creation")
	s.Equal(payload.AccountName, dbAccount.AccountName)
}

func (s *APIHandlersIntegrationTestSuite) TestGetChartOfAccountByIDAPI() {
	s.T().Log("Running TestGetChartOfAccountByIDAPI")
	// Seed an account directly or via API (if Create API test passed)
	seedAcc := models.ChartOfAccount{AccountCode: "API1020", AccountName: "AR API", AccountType: models.Asset, IsActive: true}
	result := s.db.Create(&seedAcc) // GORM Create directly
	s.Require().NoError(result.Error)
	s.Require().NotEqual(uuid.Nil, seedAcc.ID)

	path := fmt.Sprintf("/api/v1/accounting/accounts/%s", seedAcc.ID.String())
	rr := s.performRequest("GET", path, nil)
	s.Equal(http.StatusOK, rr.Code)

	var response dto.SuccessResponse
	var fetchedAccount models.ChartOfAccount
	json.Unmarshal(rr.Body.Bytes(), &response)
	dataBytes, _ := json.Marshal(response.Data)
	json.Unmarshal(dataBytes, &fetchedAccount)

	s.Equal(seedAcc.AccountCode, fetchedAccount.AccountCode)
	s.Equal(seedAcc.ID, fetchedAccount.ID)

	// Test Not Found
	nonExistentID := uuid.New()
	pathNotFound := fmt.Sprintf("/api/v1/accounting/accounts/%s", nonExistentID.String())
	rrNotFound := s.performRequest("GET", pathNotFound, nil)
	s.Equal(http.StatusNotFound, rrNotFound.Code)
}


func (s *APIHandlersIntegrationTestSuite) TestListChartOfAccountsAPI_WithFilters() {
	s.T().Log("Running TestListChartOfAccountsAPI_WithFilters")
	// Seed accounts
	s.db.Create(&models.ChartOfAccount{AccountCode: "API_L1", AccountName: "Asset One", AccountType: models.Asset, IsActive: true})
	s.db.Create(&models.ChartOfAccount{AccountCode: "API_L2", AccountName: "Liability One", AccountType: models.Liability, IsActive: true})
	s.db.Create(&models.ChartOfAccount{AccountCode: "API_L3", AccountName: "Asset Two Inactive", AccountType: models.Asset, IsActive: false})

	// Test with filter: type=ASSET, is_active=true
	rr := s.performRequest("GET", "/api/v1/accounting/accounts?account_type=ASSET&is_active=true&limit=5", nil)
	s.Equal(http.StatusOK, rr.Code)

	var response dto.SuccessResponse
	// The actual data structure for paginated list is nested
	var paginatedData struct {
		Data  []models.ChartOfAccount `json:"data"`
		Page  int                     `json:"page"`
		Limit int                     `json:"limit"`
		Total int64                   `json:"total"`
	}

	err := json.Unmarshal(rr.Body.Bytes(), &response)
	s.Require().NoError(err)
	s.Equal("success", response.Status)

	dataBytes, _ := json.Marshal(response.Data) // response.Data is the paginatedData struct
	err = json.Unmarshal(dataBytes, &paginatedData)
	s.Require().NoError(err)

	s.Len(paginatedData.Data, 1, "Should find 1 active asset account")
	s.Equal(int64(1), paginatedData.Total)
	if len(paginatedData.Data) == 1 {
		s.Equal("API_L1", paginatedData.Data[0].AccountCode)
	}
}


// --- Journal Entries API Tests ---

func (s *APIHandlersIntegrationTestSuite) TestCreateJournalEntryAPI_ValidAndBalanced() {
	s.T().Log("Running TestCreateJournalEntryAPI_ValidAndBalanced")
	// Need accounts to exist
	cashAcc := models.ChartOfAccount{AccountCode: "JEAPI1000", AccountName: "Cash JE API", AccountType: models.Asset, IsActive: true}
	revAcc := models.ChartOfAccount{AccountCode: "JEAPI4000", AccountName: "Revenue JE API", AccountType: models.Revenue, IsActive: true}
	s.db.Create(&cashAcc); s.db.Create(&revAcc)

	payload := dto.CreateJournalEntryRequest{
		EntryDate:   time.Now(),
		Description: "API Journal Sale",
		Lines: []dto.JournalLineRequest{
			{AccountID: cashAcc.ID, Amount: 250.50, IsDebit: true},
			{AccountID: revAcc.ID, Amount: 250.50, IsDebit: false},
		},
	}

	rr := s.performRequest("POST", "/api/v1/accounting/journals", payload)
	s.Equal(http.StatusCreated, rr.Code)

	var response dto.SuccessResponse
	var createdEntry models.JournalEntry
	json.Unmarshal(rr.Body.Bytes(), &response)
	dataBytes, _ := json.Marshal(response.Data)
	json.Unmarshal(dataBytes, &createdEntry)

	s.Equal(payload.Description, createdEntry.Description)
	s.Len(createdEntry.JournalLines, 2)
	s.Equal(models.StatusDraft, createdEntry.Status) // Default for new entries

	// Verify in DB
	var dbEntry models.JournalEntry
	s.db.Preload("JournalLines").First(&dbEntry, "id = ?", createdEntry.ID)
	s.Len(dbEntry.JournalLines, 2)
}

func (s *APIHandlersIntegrationTestSuite) TestCreateJournalEntryAPI_Unbalanced() {
	s.T().Log("Running TestCreateJournalEntryAPI_Unbalanced")
	cashAcc := models.ChartOfAccount{AccountCode: "JEAPI1001", AccountName: "Cash JE API Unb", AccountType: models.Asset, IsActive: true}
	revAcc := models.ChartOfAccount{AccountCode: "JEAPI4001", AccountName: "Revenue JE API Unb", AccountType: models.Revenue, IsActive: true}
	s.db.Create(&cashAcc); s.db.Create(&revAcc)

	payload := dto.CreateJournalEntryRequest{
		Description: "API Unbalanced Journal",
		Lines: []dto.JournalLineRequest{
			{AccountID: cashAcc.ID, Amount: 100.00, IsDebit: true},
			{AccountID: revAcc.ID, Amount: 99.00, IsDebit: false}, // Unbalanced
		},
	}

	rr := s.performRequest("POST", "/api/v1/accounting/journals", payload)
	s.Equal(http.StatusBadRequest, rr.Code, "Should return 400 for unbalanced entry")

	var errorResponse dto.ErrorResponse
	json.Unmarshal(rr.Body.Bytes(), &errorResponse)
	s.Equal("error", errorResponse.Status)
	s.Contains(errorResponse.Message, "debits (100.00) must equal credits (99.00)")
}


func (s *APIHandlersIntegrationTestSuite) TestPostJournalEntryAPI() {
	s.T().Log("Running TestPostJournalEntryAPI")
	// 1. Create a DRAFT journal entry
	cashAcc := models.ChartOfAccount{AccountCode: "JEAPIPost1", AccountName: "Cash JE API Post", AccountType: models.Asset, IsActive: true}
	revAcc := models.ChartOfAccount{AccountCode: "JEAPIPost2", AccountName: "Revenue JE API Post", AccountType: models.Revenue, IsActive: true}
	s.db.Create(&cashAcc); s.db.Create(&revAcc)

	draftEntry := models.JournalEntry{
		EntryDate: time.Now(), Description: "Draft for API Post", Status: models.StatusDraft,
		JournalLines: []models.JournalLine{
			{AccountID: cashAcc.ID, Amount: 75.00, IsDebit: true},
			{AccountID: revAcc.ID, Amount: 75.00, IsDebit: false},
		},
	}
	// Use GORM to create directly for test setup
	result := s.db.Create(&draftEntry)
	s.Require().NoError(result.Error)
	// Manually create lines as GORM Create doesn't cascade by default on struct like this unless associations are set up differently
	for i := range draftEntry.JournalLines {
		draftEntry.JournalLines[i].JournalID = draftEntry.ID
		s.db.Create(&draftEntry.JournalLines[i])
	}


	// 2. Call the POST endpoint
	path := fmt.Sprintf("/api/v1/accounting/journals/%s/post", draftEntry.ID.String())
	rr := s.performRequest("POST", path, nil)
	s.Equal(http.StatusOK, rr.Code, "Posting should return 200 OK")

	var response dto.SuccessResponse
	var postedEntry models.JournalEntry
	json.Unmarshal(rr.Body.Bytes(), &response)
	dataBytes, _ := json.Marshal(response.Data)
	json.Unmarshal(dataBytes, &postedEntry)

	s.Equal(models.StatusPosted, postedEntry.Status, "Entry status should be POSTED")

	// Verify in DB
	var dbEntry models.JournalEntry
	s.db.First(&dbEntry, "id = ?", draftEntry.ID)
	s.Equal(models.StatusPosted, dbEntry.Status)
}

// --- Reporting API Tests ---

func (s *APIHandlersIntegrationTestSuite) TestGetTrialBalanceAPI() {
	s.T().Log("Running TestGetTrialBalanceAPI")
	// Seed accounts
	cash := models.ChartOfAccount{AccountCode: "TB1000", AccountName: "Cash TB", AccountType: models.Asset, IsActive: true}
	revenue := models.ChartOfAccount{AccountCode: "TB4000", AccountName: "Revenue TB", AccountType: models.Revenue, IsActive: true}
	expense := models.ChartOfAccount{AccountCode: "TB5000", AccountName: "Expense TB", AccountType: models.Expense, IsActive: true}
	s.db.Create(&cash); s.db.Create(&revenue); s.db.Create(&expense)

	// Seed a posted journal entry
	entryTime := time.Date(2023, 10, 15, 0,0,0,0, time.UTC)
	postedEntry := models.JournalEntry{
		EntryDate: entryTime, Description: "TB API Sale", Status: models.StatusPosted,
		JournalLines: []models.JournalLine{
			{AccountID: cash.ID, Amount: 500.00, IsDebit: true},
			{AccountID: revenue.ID, Amount: 500.00, IsDebit: false},
		},
	}
	s.db.Create(&postedEntry)
	for i := range postedEntry.JournalLines {
		postedEntry.JournalLines[i].JournalID = postedEntry.ID
		s.db.Create(&postedEntry.JournalLines[i])
	}
	// Another entry
	postedEntry2 := models.JournalEntry{
		EntryDate: entryTime.Add(24 * time.Hour), Description: "TB API Expense", Status: models.StatusPosted,
		JournalLines: []models.JournalLine{
			{AccountID: expense.ID, Amount: 100.00, IsDebit: true},
			{AccountID: cash.ID, Amount: 100.00, IsDebit: false},
		},
	}
	s.db.Create(&postedEntry2)
	for i := range postedEntry2.JournalLines {
		postedEntry2.JournalLines[i].JournalID = postedEntry2.ID
		s.db.Create(&postedEntry2.JournalLines[i])
	}


	endDate := "2023-10-30"
	path := fmt.Sprintf("/api/v1/accounting/reports/trial-balance?end_date=%s&include_zero_balance_accounts=true", endDate)
	rr := s.performRequest("GET", path, nil)

	s.T().Logf("Response Body for Trial Balance: %s", rr.Body.String())
	s.Equal(http.StatusOK, rr.Code, "Trial Balance API should return 200 OK")

	var response dto.SuccessResponse
	var tbResponse dto.TrialBalanceResponse
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	s.Require().NoError(err, "Failed to unmarshal success wrapper")

	dataBytes, _ := json.Marshal(response.Data)
	err = json.Unmarshal(dataBytes, &tbResponse)
	s.Require().NoError(err, "Failed to unmarshal TrialBalanceResponse data")

	s.Len(tbResponse.Lines, 3, "Trial balance should have 3 lines (Cash, Revenue, Expense)")
	s.InDelta(500.00, tbResponse.TotalDebits, 0.001)  // Cash 400 DR, Expense 100 DR
	s.InDelta(500.00, tbResponse.TotalCredits, 0.001) // Revenue 500 CR

	foundCash := false
	for _, line := range tbResponse.Lines {
		if line.AccountCode == "TB1000" { // Cash
			s.InDelta(400.00, line.Debit, 0.001) // 500 DR - 100 CR
			s.InDelta(0.00, line.Credit, 0.001)
			foundCash = true
		}
	}
	s.True(foundCash, "Cash account not found or incorrect balance in TB API response")
}
