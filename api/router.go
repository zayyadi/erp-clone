package api

import (
	acc_handlers "erp-system/api/handlers" // Alias for accounting handlers
	inv_handlers "erp-system/api/handlers" // Alias for inventory handlers (will be distinct type)
	// "erp-system/api/middleware"
	acc_repo "erp-system/internal/accounting/repository" // Alias for accounting repo
	acc_service "erp-system/internal/accounting/service" // Alias for accounting service
	inv_repo "erp-system/internal/inventory/repository" // Alias for inventory repo
	inv_service "erp-system/internal/inventory/service" // Alias for inventory service
	"erp-system/pkg/logger"
	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

// NewRouter creates and configures a new Gorilla Mux router.
// It initializes dependencies for handlers.
func NewRouter(db *gorm.DB) *mux.Router {
	logger.InfoLogger.Println("Initializing application router...")

	r := mux.NewRouter()

	// Health check endpoint
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","message":"ERP system is healthy"}`))
	}).Methods("GET")

	// --- Initialize Accounting Dependencies ---
	accountingCoaRepo := acc_repo.NewChartOfAccountRepository(db)
	accountingJournalRepo := acc_repo.NewJournalEntryRepository(db)
	accountingService := acc_service.NewAccountingService(accountingCoaRepo, accountingJournalRepo)
	accountingAPIHandlers := acc_handlers.NewAccountingHandlers(accountingService)

	// --- Initialize Inventory Dependencies ---
	itemRepo := inv_repo.NewItemRepository(db)
	warehouseRepo := inv_repo.NewWarehouseRepository(db)
	inventoryTransactionRepo := inv_repo.NewInventoryTransactionRepository(db)
	inventoryService := inv_service.NewInventoryService(itemRepo, warehouseRepo, inventoryTransactionRepo)
	// Correctly use inv_handlers for NewInventoryHandlers
	inventoryAPIHandlers := inv_handlers.NewInventoryHandlers(inventoryService)


	// Apply global middleware (e.g., logging, CORS, authentication if globally applied)
	// r.Use(middleware.LoggingMiddleware)
	// r.Use(middleware.CORSMiddleware)
	// apiV1Router := r.PathPrefix("/api/v1").Subrouter()
	// apiV1Router.Use(middleware.Authenticate) // Apply auth to all /api/v1 routes


	// Register routes for different modules
	// The handlers themselves define full paths starting with /api/v1/...
	// So, we register them directly on the main router `r`.

	accountingAPIHandlers.RegisterAccountingRoutes(r)
	inventoryAPIHandlers.RegisterInventoryRoutes(r)
	// Add more module route registrations here as they are implemented

	logger.InfoLogger.Println("Router initialization complete.")
	return r
}

// Example placeholder for LoggingMiddleware (can be moved to middleware package)
/*
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        logger.InfoLogger.Printf("Request: %s %s", r.Method, r.RequestURI)
        // Potentially use a more structured logger here if you have one
        // And wrap the ResponseWriter to capture status code, response size etc.
        next.ServeHTTP(w, r)
        // Log after request is handled, e.g. response status
    })
}
*/

// Example placeholder for CORSMiddleware (can be moved to middleware package)
/*
func CORSMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*") // Adjust for production
        w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
        w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

        // Handle preflight requests
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        next.ServeHTTP(w, r)
    })
}
*/

// Note: The actual `middleware.Authenticate` is a placeholder.
// It needs to be fully implemented with JWT validation or other auth mechanisms.
// The `db *gorm.DB` is passed to initialize repositories which are then passed to services,
// and services to handlers. This sets up the dependency injection chain.
