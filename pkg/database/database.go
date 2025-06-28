package database

import (
	"erp-system/configs"
	"fmt"
	"log"
	"os" // <-- ADDED
	"sync"
	"time" // <-- ADDED

	"gorm.io/driver/postgres" // Or any other driver you plan to use
	"gorm.io/gorm"
	gorm_logger "gorm.io/gorm/logger" // <-- ADDED (aliased or direct)
)

var (
	DB   *gorm.DB
	once sync.Once
	err  error
	// testDBLock sync.Mutex // Potentially needed if InitTestDB modifies shared state with InitDB
)

// InitDB initializes the database connection using the global AppConfig for the main application.
// It's designed to be called once.
func InitDB(appCfg configs.AppConfig) (*gorm.DB, error) {
	once.Do(func() {
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
			appCfg.DBHost, appCfg.DBUser, appCfg.DBPassword, appCfg.DBName, appCfg.DBPort, appCfg.SSLMode)

		// Configure GORM logger for the main application DB instance
		gormLogger := NewGormLogger(log.New(os.Stdout, "\r\n", log.LstdFlags), LoggerConfig{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  gorm_logger.Info, // Adjust log level (Silent, Error, Warn, Info)
			IgnoreRecordNotFoundError: false,
			Colorful:                  true,
		})

		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: gormLogger,
		})
		if err != nil {
			log.Printf("Failed to connect to database: %v\n", err)
			return // err is captured by the outer scope
		}

		log.Println("Database connection established successfully.")

		// Optional: Configure connection pool
		sqlDB, sqlErr := DB.DB()
		if sqlErr != nil {
			log.Printf("Failed to get generic database object: %v\n", sqlErr)
			err = sqlErr // Propagate this error
			return
		}
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		// sqlDB.SetConnMaxLifetime(time.Hour) // Example
	})

	if err != nil {
		return nil, fmt.Errorf("error initializing database: %w", err)
	}
	if DB == nil && err == nil {
		return nil, fmt.Errorf("database not initialized and no error was reported")
	}
	return DB, nil
}

// InitTestDB initializes a new database connection for testing purposes.
// It does not use the singleton 'once' and returns a new *gorm.DB instance each time.
func InitTestDB(cfg configs.AppConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.SSLMode)

	// Configure GORM logger for test DB instance (e.g., quieter or different output)
	gormTestLogger := NewGormLogger(log.New(os.Stdout, "[TEST_DB] ", log.LstdFlags), LoggerConfig{
		SlowThreshold:             500 * time.Millisecond,
		LogLevel:                  gorm_logger.Warn, // Less verbose for tests
		IgnoreRecordNotFoundError: true,           // Often expected in tests
		Colorful:                  false,
	})

	testDb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormTestLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to test database: %w", err)
	}

	// Optional: Configure connection pool for test DB if needed, though often not necessary for tests
	sqlDB, sqlErr := testDb.DB()
	if sqlErr != nil {
		return nil, fmt.Errorf("failed to get generic database object for test DB: %w", sqlErr)
	}
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(10)

	log.Printf("Test database connection established successfully to %s:%s/%s", cfg.DBHost, cfg.DBPort, cfg.DBName)
	return testDb, nil
}


// GetDB returns the current database instance.
func GetDB() *gorm.DB {
	if DB == nil {
		log.Println("Warning: GetDB called before InitDB or after a failed InitDB.")
		// Depending on strictness, you might panic or return nil
		// For now, returning nil, which will cause a panic if used without checking.
	}
	return DB
}

// CloseDB closes the main application's database connection.
func CloseDB() {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			log.Printf("Error getting DB instance for closing: %v\n", err)
			return
		}
		err = sqlDB.Close()
		if err != nil {
			log.Printf("Error closing database connection: %v\n", err)
		} else {
			log.Println("Database connection closed.")
		}
	}
}
