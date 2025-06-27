package configs

import (
	"log"
	"os"
	"strconv"

	"github.com/spf13/viper"
)

// AppConfig holds all configuration for the application.
type AppConfig struct {
	ServerPort string `mapstructure:"SERVER_PORT"`
	DBHost     string `mapstructure:"DB_HOST"`
	DBPort     string `mapstructure:"DB_PORT"`
	DBUser     string `mapstructure:"DB_USER"`
	DBPassword string `mapstructure:"DB_PASSWORD"`
	DBName     string `mapstructure:"DB_NAME"`
	SSLMode    string `mapstructure:"DB_SSLMODE"`
	// Add other configurations here, e.g., JWT secret, API keys, etc.
}

var GlobalConfig AppConfig

// LoadConfig loads configuration from file and environment variables.
func LoadConfig(path string) (AppConfig, error) {
	viper.AddConfigPath(path) // Path to look for the config file in
	viper.SetConfigName("app") // Name of config file (without extension)
	viper.SetConfigType("env") // Config file type (e.g., .env, .yaml, .json)

	viper.AutomaticEnv() // Read in environment variables that match

	// Set default values
	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_USER", "postgres")
	viper.SetDefault("DB_PASSWORD", "password")
	viper.SetDefault("DB_NAME", "erp_dev")
	viper.SetDefault("DB_SSLMODE", "disable")

	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if we have defaults/env vars
			log.Println("Config file not found, using defaults and environment variables.")
		} else {
			// Config file was found but another error was produced
			return AppConfig{}, err
		}
	}

	var config AppConfig
	err = viper.Unmarshal(&config)
	if err != nil {
		return AppConfig{}, err
	}

	// Override with environment variables if they are set,
	// as viper.AutomaticEnv() might not always override file values for struct unmarshalling
	// depending on how keys are cased or structured.
	// This ensures env vars take precedence.
	overrideWithEnvVar("SERVER_PORT", &config.ServerPort)
	overrideWithEnvVar("DB_HOST", &config.DBHost)
	overrideWithEnvVar("DB_PORT", &config.DBPort)
	overrideWithEnvVar("DB_USER", &config.DBUser)
	overrideWithEnvVar("DB_PASSWORD", &config.DBPassword)
	overrideWithEnvVar("DB_NAME", &config.DBName)
	overrideWithEnvVar("DB_SSLMODE", &config.SSLMode)

	GlobalConfig = config
	log.Println("Configuration loaded successfully.")
	return config, nil
}

func overrideWithEnvVar(envVar string, value *string) {
	if val, exists := os.LookupEnv(envVar); exists {
		*value = val
	}
}

// GetConfig returns the loaded application configuration.
func GetConfig() AppConfig {
	return GlobalConfig
}

// Helper function to get string from viper or default
func getString(key string, defaultValue string) string {
	if viper.IsSet(key) {
		return viper.GetString(key)
	}
	return defaultValue
}

// Helper function to get int from viper or default
func getInt(key string, defaultValue int) int {
	if viper.IsSet(key) {
		val, err := strconv.Atoi(viper.GetString(key))
		if err == nil {
			return val
		}
		log.Printf("Warning: Could not parse %s as int, using default. Value: %s", key, viper.GetString(key))
	}
	return defaultValue
}
