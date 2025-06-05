package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

// Config holds the configuration for the Job Distributor
type Config struct {
	Port          int    `json:"port" yaml:"port"`
	LogLevel      string `json:"log_level" yaml:"log_level"`
	EnableMetrics bool   `json:"enable_metrics" yaml:"enable_metrics"`
	MetricsPort   int    `json:"metrics_port" yaml:"metrics_port"`
	ConfigFile    string `json:"config_file" yaml:"config_file"`
}

// Load loads configuration from environment variables and optionally from a config file
func Load() *Config {
	cfg := &Config{
		Port:          getEnvInt("JD_PORT", 50051),
		LogLevel:      getEnvString("JD_LOG_LEVEL", "info"),
		EnableMetrics: getEnvBool("JD_ENABLE_METRICS", false),
		MetricsPort:   getEnvInt("JD_METRICS_PORT", 8080),
		ConfigFile:    getEnvString("JD_CONFIG_FILE", ""),
	}

	// Load from config file if specified
	if cfg.ConfigFile != "" {
		if err := cfg.loadFromFile(); err != nil {
			fmt.Printf("⚠️ Warning: Failed to load config file: %v\n", err)
		}
	}

	return cfg
}

// loadFromFile loads configuration from a JSON config file
func (c *Config) loadFromFile() error {
	data, err := os.ReadFile(c.ConfigFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := json.Unmarshal(data, c); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	return nil
}

// Helper functions for environment variable parsing
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
