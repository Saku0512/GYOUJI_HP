package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Database DatabaseConfig
	JWT      JWTConfig
	Server   ServerConfig
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	DSN      string
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	SecretKey       string
	ExpirationHours int
	Issuer          string
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port string
	Host string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	config := &Config{}

	// Database configuration
	config.Database.Host = getEnv("DB_HOST", "localhost")
	config.Database.Port = getEnvAsInt("DB_PORT", 3306)
	config.Database.User = getEnv("DB_USER", "root")
	config.Database.Password = getEnv("DB_PASSWORD", "")
	config.Database.DBName = getEnv("DB_NAME", "tournament_db")

	// Build DSN
	config.Database.DSN = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Database.User,
		config.Database.Password,
		config.Database.Host,
		config.Database.Port,
		config.Database.DBName,
	)

	// JWT configuration
	config.JWT.SecretKey = getEnv("JWT_SECRET", "your-secret-key-change-in-production")
	config.JWT.ExpirationHours = getEnvAsInt("JWT_EXPIRATION_HOURS", 24)
	config.JWT.Issuer = getEnv("JWT_ISSUER", "tournament-backend")

	// Server configuration
	config.Server.Port = getEnv("SERVER_PORT", "8080")
	config.Server.Host = getEnv("SERVER_HOST", "0.0.0.0")

	// Validate required configuration
	if config.Database.Password == "" {
		return nil, fmt.Errorf("DB_PASSWORD environment variable is required")
	}

	if config.JWT.SecretKey == "your-secret-key-change-in-production" {
		fmt.Println("WARNING: Using default JWT secret key. Please set JWT_SECRET environment variable in production.")
	}

	return config, nil
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getEnvAsInt gets an environment variable as integer with a fallback value
func getEnvAsInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return fallback
}

// GetServerAddress returns the full server address
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%s", c.Server.Host, c.Server.Port)
}

// GetJWTExpiration returns JWT expiration duration
func (c *Config) GetJWTExpiration() time.Duration {
	return time.Duration(c.JWT.ExpirationHours) * time.Hour
}