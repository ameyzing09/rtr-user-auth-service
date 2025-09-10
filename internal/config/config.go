package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Database DatabaseConfig
	JWT      JWTConfig
	Server   ServerConfig
	CORS     CORSConfig
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret           string
	AccessExpiry     time.Duration
	RefreshExpiry    time.Duration
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port    string
	GinMode string
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins []string
}

// GetDSN returns the database connection string
func (dc *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dc.User, dc.Password, dc.Host, dc.Port, dc.Name)
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// It's okay if .env file doesn't exist
		fmt.Println("No .env file found, using environment variables")
	}

	cfg := &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "3306"),
			Name:     getEnv("DB_NAME", "recrutr_auth"),
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", "password"),
		},
		JWT: JWTConfig{
			Secret:           getEnv("JWT_SECRET", "your-super-secret-jwt-key-here"),
			AccessExpiry:     parseDuration(getEnv("JWT_EXPIRY", "24h")),
			RefreshExpiry:    parseDuration(getEnv("JWT_REFRESH_EXPIRY", "168h")),
		},
		Server: ServerConfig{
			Port:    getEnv("PORT", "8080"),
			GinMode: getEnv("GIN_MODE", "debug"),
		},
		CORS: CORSConfig{
			AllowedOrigins: parseStringSlice(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:8080")),
		},
	}

	return cfg, cfg.validate()
}

// validate validates the configuration
func (c *Config) validate() error {
	if c.JWT.Secret == "" || c.JWT.Secret == "your-super-secret-jwt-key-here" {
		return fmt.Errorf("JWT_SECRET must be set to a secure value")
	}
	if len(c.JWT.Secret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters long")
	}
	if c.Database.Host == "" {
		return fmt.Errorf("DB_HOST must be set")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("DB_NAME must be set")
	}
	return nil
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// parseDuration parses duration string with fallback
func parseDuration(str string) time.Duration {
	duration, err := time.ParseDuration(str)
	if err != nil {
		return 24 * time.Hour // Default to 24 hours
	}
	return duration
}

// parseStringSlice parses comma-separated string into slice
func parseStringSlice(str string) []string {
	if str == "" {
		return []string{}
	}
	var result []string
	for _, s := range splitAndTrim(str, ",") {
		if s != "" {
			result = append(result, s)
		}
	}
	return result
}

// splitAndTrim splits string by separator and trims each element
func splitAndTrim(str, sep string) []string {
	var result []string
	for _, s := range []string{str} {
		if sep == "," {
			// Simple comma splitting
			parts := []string{}
			current := ""
			for _, char := range s {
				if char == ',' {
					parts = append(parts, current)
					current = ""
				} else {
					current += string(char)
				}
			}
			parts = append(parts, current)
			
			for _, part := range parts {
				trimmed := ""
				// Manual trim
				start := 0
				end := len(part)
				for start < len(part) && (part[start] == ' ' || part[start] == '\t' || part[start] == '\n') {
					start++
				}
				for end > start && (part[end-1] == ' ' || part[end-1] == '\t' || part[end-1] == '\n') {
					end--
				}
				if start < end {
					trimmed = part[start:end]
				}
				if trimmed != "" {
					result = append(result, trimmed)
				}
			}
		}
	}
	return result
}