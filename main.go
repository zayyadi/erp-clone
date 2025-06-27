package main

import (
	// "fmt" // No longer directly used, using logger
	"erp-system/api"
	"erp-system/configs"
	"erp-system/pkg/database"
	"erp-system/pkg/logger" // Using our custom logger
	"log"                   // Standard log for fatal errors before logger might be fully up
	"net/http"
)

func main() {
	// Initialize custom logger first (it uses standard log internally but with prefixes/flags)
	logger.InfoLogger.Println("Starting ERP System...")

	// Load configuration
	// The path "." means it will look for "app.env" in the same directory as the executable.
	// For development, this is usually the project root. For deployment, ensure app.env is there.
	cfg, err := configs.LoadConfig("./configs") // Load from ./configs/app.env
	if err != nil {
		log.Fatalf("FATAL: Could not load config: %v", err) // Use standard log for critical early failure
	}
	logger.InfoLogger.Printf("Configuration loaded. Server Port: %s, DB Name: %s", cfg.ServerPort, cfg.DBName)


	// Initialize database connection
	// InitDB now takes the AppConfig struct
	db, err := database.InitDB(cfg)
	if err != nil {
		logger.ErrorLogger.Fatalf("FATAL: Could not initialize database: %v", err)
	}
	// defer database.CloseDB() // Ensure DB connection is closed when main exits
	// GORM's DB.Close() is usually called on the *sql.DB instance. database.CloseDB() handles this.
	// deferring it here means it runs when main exits.

	sqlDB, _ := db.DB() // Get the underlying sql.DB
    if sqlDB != nil {
        defer func() {
            logger.InfoLogger.Println("Closing database connection...")
            if err := sqlDB.Close(); err != nil {
                logger.ErrorLogger.Printf("Error closing database: %v", err)
            } else {
                logger.InfoLogger.Println("Database connection closed successfully.")
            }
        }()
    }


	// Initialize router
	// NewRouter now takes the *gorm.DB instance
	router := api.NewRouter(db)
	logger.InfoLogger.Println("HTTP router initialized.")

	// Start server
	// ServerPort is now from the loaded configuration
	logger.InfoLogger.Printf("Server starting on port %s", cfg.ServerPort)
	if err := http.ListenAndServe(":"+cfg.ServerPort, router); err != nil {
		logger.ErrorLogger.Fatalf("FATAL: Could not start server: %v", err)
	}
}
