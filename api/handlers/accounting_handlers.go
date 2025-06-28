package handlers

import (
	"encoding/json"
	"erp-system/internal/accounting/models"
	"erp-system/internal/accounting/service"
	acc_dto "erp-system/internal/accounting/service/dto" // Alias for accounting specific DTOs
	"erp-system/pkg/errors" // Keep this for type assertion in respondWithError if it's still specific
	// "erp-system/pkg/logger" // REMOVED - No longer directly used here
	"net/http"
	"strconv" // For parsing limit/page from query
	"time"    // Was missing, needed for date parsing in GetTrialBalance

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// AccountingHandlers wraps the accounting service to provide HTTP handlers.
type AccountingHandlers struct {
	service service.AccountingService
}

// NewAccountingHandlers creates a new AccountingHandlers instance.
func NewAccountingHandlers(serv service.AccountingService) *AccountingHandlers {
	return &AccountingHandlers{service: serv}
}

// RegisterAccountingRoutes registers all accounting module routes with the provided router.
func (h *AccountingHandlers) RegisterAccountingRoutes(r *mux.Router) {
	// Chart of Accounts Routes
	coaRouter := r.PathPrefix("/api/v1/accounting/accounts").Subrouter()
	coaRouter.HandleFunc("", h.CreateChartOfAccount).Methods("POST")
	coaRouter.HandleFunc("", h.ListChartOfAccounts).Methods("GET")
	coaRouter.HandleFunc("/{id}", h.GetChartOfAccountByID).Methods("GET")
	coaRouter.HandleFunc("/code/{code}", h.GetChartOfAccountByCode).Methods("GET") // Custom route for by code
	coaRouter.HandleFunc("/{id}", h.UpdateChartOfAccount).Methods("PUT")
	coaRouter.HandleFunc("/{id}", h.DeleteChartOfAccount).Methods("DELETE")

	// Journal Entries Routes
	journalRouter := r.PathPrefix("/api/v1/accounting/journals").Subrouter()
	journalRouter.HandleFunc("", h.CreateJournalEntry).Methods("POST")
	journalRouter.HandleFunc("", h.ListJournalEntries).Methods("GET")
	journalRouter.HandleFunc("/{id}", h.GetJournalEntryByID).Methods("GET")
	journalRouter.HandleFunc("/{id}", h.UpdateJournalEntry).Methods("PUT")
	journalRouter.HandleFunc("/{id}", h.DeleteJournalEntry).Methods("DELETE")
	journalRouter.HandleFunc("/{id}/post", h.PostJournalEntry).Methods("POST")

	// Reporting Routes
	reportRouter := r.PathPrefix("/api/v1/accounting/reports").Subrouter()
	reportRouter.HandleFunc("/trial-balance", h.GetTrialBalance).Methods("GET") // Changed to GET as it's safer for report generation
	// Add other report routes here, e.g., Balance Sheet, P&L
}

// respondWithError and respondWithJSON are now in response_utils.go (same package)

// --- Chart of Accounts Handlers ---

func (h *AccountingHandlers) CreateChartOfAccount(w http.ResponseWriter, r *http.Request) {
	var req acc_dto.CreateChartOfAccountRequest // Changed to acc_dto
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, errors.NewValidationError("Invalid request payload", err.Error()))
		return
	}
	defer r.Body.Close()

	// Default IsActive to true if not explicitly provided by client, GORM model default also does this.
	// However, explicit handling here can be clearer.
	if r.ContentLength > 0 { // Check if body was actually sent
		// If 'is_active' field is omitted in JSON, it will be `false` for bool.
		// The DTO doesn't use a pointer for IsActive on create, so it will be false if not sent.
		// The service layer should ideally handle defaulting if the DTO doesn't enforce it.
		// For now, assume service/model `BeforeCreate` handles default true for `IsActive`.
		// Or, the DTO should reflect that IsActive is optional and defaults to true.
		// Let's adjust CreateChartOfAccountRequest to have IsActive default to true if not specified.
		// This is tricky with JSON unmarshalling of bools. A pointer or explicit check is better.
		// For now, assume IsActive=false if not sent, and service will use model's default:true.
	}


	account, err := h.service.CreateChartOfAccount(r.Context(), req)
	if err != nil {
		respondWithError(w, err)
		return
	}
	respondWithJSON(w, http.StatusCreated, account)
}

func (h *AccountingHandlers) GetChartOfAccountByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		respondWithError(w, errors.NewValidationError("Missing account ID in path", "id"))
		return
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, errors.NewValidationError("Invalid account ID format", "id"))
		return
	}

	account, err := h.service.GetChartOfAccountByID(r.Context(), id)
	if err != nil {
		respondWithError(w, err)
		return
	}
	respondWithJSON(w, http.StatusOK, account)
}

func (h *AccountingHandlers) GetChartOfAccountByCode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	code, ok := vars["code"]
	if !ok || code == "" {
		respondWithError(w, errors.NewValidationError("Missing account code in path", "code"))
		return
	}

	account, err := h.service.GetChartOfAccountByCode(r.Context(), code)
	if err != nil {
		respondWithError(w, err)
		return
	}
	respondWithJSON(w, http.StatusOK, account)
}

func (h *AccountingHandlers) UpdateChartOfAccount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		respondWithError(w, errors.NewValidationError("Missing account ID in path", "id"))
		return
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, errors.NewValidationError("Invalid account ID format", "id"))
		return
	}

	var req acc_dto.UpdateChartOfAccountRequest // Changed to acc_dto
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, errors.NewValidationError("Invalid request payload", err.Error()))
		return
	}
	defer r.Body.Close()

	account, err := h.service.UpdateChartOfAccount(r.Context(), id, req)
	if err != nil {
		respondWithError(w, err)
		return
	}
	respondWithJSON(w, http.StatusOK, account)
}

func (h *AccountingHandlers) DeleteChartOfAccount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		respondWithError(w, errors.NewValidationError("Missing account ID in path", "id"))
		return
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, errors.NewValidationError("Invalid account ID format", "id"))
		return
	}

	if err := h.service.DeleteChartOfAccount(r.Context(), id); err != nil {
		respondWithError(w, err)
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Chart of account deleted successfully"})
}

func (h *AccountingHandlers) ListChartOfAccounts(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	listReq := acc_dto.ListChartOfAccountsRequest{ // Changed to acc_dto
		Page:        1,  // Default
		Limit:       20, // Default
		AccountName: queryParams.Get("account_name"),
		AccountType: models.AccountType(queryParams.Get("account_type")),
	}

	if pageStr := queryParams.Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			listReq.Page = page
		}
	}
	if limitStr := queryParams.Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			listReq.Limit = limit
		}
	}
    if isActiveStr := queryParams.Get("is_active"); isActiveStr != "" {
        isActive, err := strconv.ParseBool(isActiveStr)
        if err == nil {
            listReq.IsActive = &isActive
        } else {
            respondWithError(w, errors.NewValidationError("Invalid boolean value for 'is_active'", "is_active"))
            return
        }
    }


	accounts, total, err := h.service.ListChartOfAccounts(r.Context(), listReq)
	if err != nil {
		respondWithError(w, err)
		return
	}

	// Use shared PaginatedResponse type from handlers package
	paginatedResponse := PaginatedResponse{
		Data:  accounts,
		Page:  listReq.Page,
		Limit: listReq.Limit,
		Total: total,
	}
	respondWithJSON(w, http.StatusOK, paginatedResponse)
}


// --- Journal Entries Handlers ---

func (h *AccountingHandlers) CreateJournalEntry(w http.ResponseWriter, r *http.Request) {
	var req acc_dto.CreateJournalEntryRequest // Changed to acc_dto
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, errors.NewValidationError("Invalid request payload", err.Error()))
		return
	}
	defer r.Body.Close()

	entry, err := h.service.CreateJournalEntry(r.Context(), req)
	if err != nil {
		respondWithError(w, err)
		return
	}
	respondWithJSON(w, http.StatusCreated, entry)
}

func (h *AccountingHandlers) GetJournalEntryByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		respondWithError(w, errors.NewValidationError("Missing journal entry ID in path", "id"))
		return
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, errors.NewValidationError("Invalid journal entry ID format", "id"))
		return
	}

	entry, err := h.service.GetJournalEntryByID(r.Context(), id)
	if err != nil {
		respondWithError(w, err)
		return
	}
	respondWithJSON(w, http.StatusOK, entry)
}

func (h *AccountingHandlers) UpdateJournalEntry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		respondWithError(w, errors.NewValidationError("Missing journal entry ID in path", "id"))
		return
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, errors.NewValidationError("Invalid journal entry ID format", "id"))
		return
	}

	var req acc_dto.UpdateJournalEntryRequest // Changed to acc_dto
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, errors.NewValidationError("Invalid request payload", err.Error()))
		return
	}
	defer r.Body.Close()

	entry, err := h.service.UpdateJournalEntry(r.Context(), id, req)
	if err != nil {
		respondWithError(w, err)
		return
	}
	respondWithJSON(w, http.StatusOK, entry)
}

func (h *AccountingHandlers) DeleteJournalEntry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		respondWithError(w, errors.NewValidationError("Missing journal entry ID in path", "id"))
		return
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, errors.NewValidationError("Invalid journal entry ID format", "id"))
		return
	}

	if err := h.service.DeleteJournalEntry(r.Context(), id); err != nil {
		respondWithError(w, err)
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Journal entry deleted successfully"})
}

func (h *AccountingHandlers) ListJournalEntries(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	listReq := acc_dto.ListJournalEntriesRequest{ // Changed to acc_dto
		Page:        1,
		Limit:       20,
		Description: queryParams.Get("description"),
		Reference:   queryParams.Get("reference"),
		Status:      models.JournalStatus(queryParams.Get("status")),
	}

	if pageStr := queryParams.Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			listReq.Page = page
		}
	}
	if limitStr := queryParams.Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			listReq.Limit = limit
		}
	}
	if dateFromStr := queryParams.Get("date_from"); dateFromStr != "" {
		if t, err := time.Parse("2006-01-02", dateFromStr); err == nil {
			listReq.DateFrom = t
		} else {
			respondWithError(w, errors.NewValidationError("Invalid date_from format, use YYYY-MM-DD", "date_from"))
			return
		}
	}
	if dateToStr := queryParams.Get("date_to"); dateToStr != "" {
		if t, err := time.Parse("2006-01-02", dateToStr); err == nil {
			listReq.DateTo = t
		} else {
			respondWithError(w, errors.NewValidationError("Invalid date_to format, use YYYY-MM-DD", "date_to"))
			return
		}
	}
	if accountIDStr := queryParams.Get("account_id"); accountIDStr != "" {
		if accID, err := uuid.Parse(accountIDStr); err == nil {
			listReq.AccountID = accID
		} else {
			respondWithError(w, errors.NewValidationError("Invalid account_id format", "account_id"))
			return
		}
	}


	entries, total, err := h.service.ListJournalEntries(r.Context(), listReq)
	if err != nil {
		respondWithError(w, err)
		return
	}

	// Use shared PaginatedResponse type from handlers package
	paginatedResponse := PaginatedResponse{
		Data:  entries,
		Page:  listReq.Page,
		Limit: listReq.Limit,
		Total: total,
	}
	respondWithJSON(w, http.StatusOK, paginatedResponse)
}

func (h *AccountingHandlers) PostJournalEntry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		respondWithError(w, errors.NewValidationError("Missing journal entry ID in path", "id"))
		return
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, errors.NewValidationError("Invalid journal entry ID format", "id"))
		return
	}

	entry, err := h.service.PostJournalEntry(r.Context(), id)
	if err != nil {
		respondWithError(w, err)
		return
	}
	respondWithJSON(w, http.StatusOK, entry)
}

// --- Reporting Handlers ---

func (h *AccountingHandlers) GetTrialBalance(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	endDateStr := queryParams.Get("end_date")
	if endDateStr == "" {
		respondWithError(w, errors.NewValidationError("end_date query parameter is required", "end_date"))
		return
	}
	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		respondWithError(w, errors.NewValidationError("Invalid end_date format, use YYYY-MM-DD", "end_date"))
		return
	}

	req := acc_dto.TrialBalanceRequest{ // Changed to acc_dto
		EndDate: endDate,
	}
    if startDateStr := queryParams.Get("start_date"); startDateStr != "" {
        if t, err := time.Parse("2006-01-02", startDateStr); err == nil { // time.Parse needs "time" import
            req.StartDate = t
        } else {
            respondWithError(w, errors.NewValidationError("Invalid start_date format, use YYYY-MM-DD", "start_date"))
            return
        }
    }
	if includeZeroStr := queryParams.Get("include_zero_balance_accounts"); includeZeroStr != "" {
		includeZero, err := strconv.ParseBool(includeZeroStr)
		if err == nil {
			req.IncludeZeroBalanceAccounts = includeZero
		} else {
            respondWithError(w, errors.NewValidationError("Invalid boolean value for 'include_zero_balance_accounts'", "include_zero_balance_accounts"))
            return
        }
	}

	report, err := h.service.GetTrialBalance(r.Context(), req)
	if err != nil {
		respondWithError(w, err)
		return
	}
	respondWithJSON(w, http.StatusOK, report)
}
