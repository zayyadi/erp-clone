package database

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	gorm_logger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

// LoggerConfig defines the configuration for the GORM logger.
type LoggerConfig struct {
	SlowThreshold             time.Duration
	LogLevel                  gorm_logger.LogLevel
	IgnoreRecordNotFoundError bool
	Colorful                  bool
}

// gormLogger is a custom GORM logger implementation.
type gormLogger struct {
	writer *log.Logger
	config LoggerConfig
}

// NewGormLogger creates a new custom GORM logger.
func NewGormLogger(writer *log.Logger, config LoggerConfig) gorm_logger.Interface {
	return &gormLogger{
		writer: writer,
		config: config,
	}
}

// LogMode sets the log level.
func (l *gormLogger) LogMode(level gorm_logger.LogLevel) gorm_logger.Interface {
	newLogger := *l
	newLogger.config.LogLevel = level
	return &newLogger
}

// Info prints info messages.
func (l *gormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.config.LogLevel >= gorm_logger.Info {
		l.writer.Printf(msg, data...)
	}
}

// Warn prints warning messages.
func (l *gormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.config.LogLevel >= gorm_logger.Warn {
		l.writer.Printf(msg, data...)
	}
}

// Error prints error messages.
func (l *gormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.config.LogLevel >= gorm_logger.Error {
		l.writer.Printf(msg, data...)
	}
}

// Trace prints SQL queries and execution details.
func (l *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.config.LogLevel <= gorm_logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()
	sqlSource := utils.FileWithLineNum() // Gets the file and line number of the GORM call

	logMsg := fmt.Sprintf("[GORM] Source: %s | Elapsed: %.3fms | Rows: %d | SQL: %s", sqlSource, float64(elapsed.Milliseconds()), rows, sql)

	switch {
	case err != nil && l.config.LogLevel >= gorm_logger.Error && (!errors.Is(err, gorm_logger.ErrRecordNotFound) || !l.config.IgnoreRecordNotFoundError):
		// Error (excluding ErrRecordNotFound if ignored)
		l.writer.Printf("%s | Error: %v", logMsg, err)
	case elapsed > l.config.SlowThreshold && l.config.SlowThreshold != 0 && l.config.LogLevel >= gorm_logger.Warn:
		// Slow query
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.config.SlowThreshold)
		l.writer.Printf("%s | %s", logMsg, slowLog)
	case l.config.LogLevel >= gorm_logger.Info:
		// Normal info log for queries
		l.writer.Printf("%s", logMsg)
	}
}

/*
// Example usage if you were to set it directly in GORM config:
// (This is now handled by InitDB and InitTestDB)

import (
	"log"
	"os"
	"time"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func main() {
	// Standard library logger
	stdLibLogger := log.New(os.Stdout, "\r\n", log.LstdFlags)

	// Custom GORM logger configuration
	customLoggerConfig := LoggerConfig{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  gormlogger.Info,
		IgnoreRecordNotFoundError: true,
		Colorful:                  true, // Colorful usually works best in terminal, might be messy in files
	}

	loggerInstance := NewGormLogger(stdLibLogger, customLoggerConfig)

	dsn := "host=localhost user=youruser password=yourpassword dbname=yourdb port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: loggerInstance,
	})

	if err != nil {
		panic("Failed to connect to database")
	}

	// Use db instance...
}

*/
