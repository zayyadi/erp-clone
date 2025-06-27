package api

import (
	"erp-system/api/handlers"
	"erp-system/api/middleware"
	"erp-system/internal/accounting/repository"
	"erp-system/internal/accounting/service"
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

	// Initialize repositories
	coaRepo := repository.NewChartOfAccountRepository(db)
	journalRepo := repository.NewJournalEntryRepository(db)
	// Add other repositories here as modules are implemented

	// Initialize services
	accountingService := service.NewAccountingService(coaRepo, journalRepo)
	// Add other services here

	// Initialize handlers
	accountingHandlers := handlers.NewAccountingHandlers(accountingService)
	// Add other handlers here

	// Register routes for different modules
	// All routes can be prefixed with /api/v1 if desired, handled by subrouters or PathPrefix
	// Example: apiV1 := r.PathPrefix("/api/v1").Subrouter()

	// Apply global middleware (e.g., logging, CORS, authentication if globally applied)
	// r.Use(middleware.LoggingMiddleware) // Example global logging middleware
	// r.Use(middleware.CORSMiddleware)    // Example CORS middleware

	// For authenticated routes:
	// authRouter := r.PathPrefix("/api/v1").Subrouter() // Or apply to specific subrouters
	// authRouter.Use(middleware.Authenticate) // Apply authentication middleware

	// Register Accounting Module Routes (these are already prefixed with /api/v1/accounting in the handler)
	// If Authenticate middleware should apply to these, it needs to be on `r` or a parent subrouter.
	// For now, let's assume accounting routes might need authentication.
	// We can create a subrouter for authenticated API endpoints.

	// Example: Group API routes that need authentication
	// This assumes that all routes registered by accountingHandlers.RegisterAccountingRoutes
	// are intended to be under /api/v1 and require authentication.
	// The handler itself defines /api/v1/accounting/*, so we can attach middleware to `r` if all are auth'd
	// or be more granular.
	// Let's make a general API subrouter and apply auth there.

	apiRouter := r.PathPrefix("/api/v1").Subrouter()
	// To enable authentication for all /api/v1 routes:
	// apiRouter.Use(middleware.Authenticate) // Uncomment when JWT/auth is fully set up

	// Pass the main router `r` to RegisterAccountingRoutes, as it internally defines paths starting with /api/v1
	// OR pass `apiRouter` if RegisterAccountingRoutes defines paths relative to it (e.g., "/accounting/journals")
	// Given the current RegisterAccountingRoutes, it expects the main router `r`.
	accountingHandlers.RegisterAccountingRoutes(r) // Registers routes like /api/v1/accounting/...

	// If we wanted all routes from accountingHandlers to be under an `apiRouter` that already has `/api/v1`
	// and potentially middleware like `Authenticate`, then RegisterAccountingRoutes would need to
	// define its paths relative to that subrouter (e.g. `coaRouter := r.PathPrefix("/accounts").Subrouter()`).
	// For now, the current setup is fine, assuming `middleware.Authenticate` would be added to `r` or `apiRouter` as needed.

	// Example of how to use the Authenticate middleware on a specific group if not all routes need it:
	// authenticatedAccountingRouter := r.PathPrefix("/api/v1/accounting").Subrouter()
	// authenticatedAccountingRouter.Use(middleware.Authenticate)
	// accountingHandlers.RegisterAccountingRoutes(authenticatedAccountingRouter) // This would require handler paths to be relative

	// Add more module route registrations here:
	// e.g., inventoryHandlers.RegisterInventoryRoutes(apiRouter) // if apiRouter is /api/v1

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
