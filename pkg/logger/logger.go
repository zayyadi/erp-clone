package logger

import (
	"log"
	"os"
)

var (
	InfoLogger  *log.Logger
	WarnLogger  *log.Logger
	ErrorLogger *log.Logger
)

func init() {
	// You can customize the output (e.g., a file) and flags (e.g., include timestamp, file name)
	InfoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarnLogger = log.New(os.Stdout, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

// Example of how to use the loggers:
// logger.InfoLogger.Println("This is an info message.")
// logger.WarnLogger.Println("This is a warning message.")
// logger.ErrorLogger.Println("This is an error message.")

// For more advanced logging, consider libraries like:
// - logrus: https://github.com/sirupsen/logrus
// - zap: https://github.com/uber-go/zap
// These offer structured logging, levels, hooks, and more.
// For now, the standard library's log package is used for simplicity.

// If you want to integrate with the config (e.g., log level, log file path):
/*
import "erp-system/configs"

func InitLogger(cfg configs.AppConfig) {
    // Example: Set log level based on config
    logLevel := cfg.LogLevel // Assuming LogLevel is part of AppConfig

    infoHandle := os.Stdout
    warnHandle := os.Stdout
    errorHandle := os.Stderr

    // Example: Log to file if specified in config
    // if cfg.LogFilePath != "" {
    //     file, err := os.OpenFile(cfg.LogFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
    //     if err == nil {
    //         infoHandle = file
    //         warnHandle = file
    //         errorHandle = file
    //     } else {
    //         log.Printf("Failed to open log file: %s, defaulting to stdout/stderr", cfg.LogFilePath)
    //     }
    // }

    InfoLogger = log.New(infoHandle, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
    WarnLogger = log.New(warnHandle, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile)
    ErrorLogger = log.New(errorHandle, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

    InfoLogger.Printf("Logger initialized with level: %s", logLevel)
}
*/
