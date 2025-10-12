package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Auth     AuthConfig
	Logging  LoggingConfig
	Slug     SlugConfig
	Platform PlatformConfig
	Cookie   CookieConfig
}

// ServerConfig contains server-related settings
type ServerConfig struct {
	Port    string
	GinMode string
	Env     string
}

// DatabaseConfig contains database connection settings
type DatabaseConfig struct {
	User            string
	Password        string
	Host            string
	Port            string
	Name            string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// JWTConfig contains JWT token settings
type JWTConfig struct {
	Secret      string
	DefaultTTL  time.Duration
	DevFallback string
}

// AuthConfig contains authentication settings
type AuthConfig struct {
	SuperadminDevToken string
	AllowDevSuperadmin bool
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level string
}

// SlugConfig contains slug generation settings
type SlugConfig struct {
	MinLength int
	MaxLength int
	Suffixes  []string
}

// PlatformConfig contains platform branding settings
type PlatformConfig struct {
	BrandName          string
	BrandLogoURL       string
	BrandPrimaryColor  string
	BrandAccentColor   string
	BrandNavbarTitle   string
	BrandSidebarTitle  string
	BrandSidebarLinks  string // JSON string
	ParsedSidebarLinks []PlatformNavItem
}

// PlatformNavItem represents a navigation item in the platform sidebar
type PlatformNavItem struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Path  string `json:"path"`
}

// CookieConfig contains HTTP cookie settings for session management
type CookieConfig struct {
	Domain   string        // Cookie domain (empty for current domain)
	Secure   bool          // Only send over HTTPS
	SameSite string        // SameSite attribute: "Lax", "Strict", or "None"
	MaxAge   time.Duration // Cookie expiration duration
}

const (
	// Default values
	defaultServerPort        = "8082"
	defaultJWTTTL            = 24 * time.Hour
	defaultJWTSecret         = "dev-secret"
	defaultDevSuperadminTok  = "dev-superadmin"
	defaultDBMaxOpenConns    = 25
	defaultDBMaxIdleConns    = 5
	defaultDBConnMaxLifetime = 5 * time.Minute
	defaultSlugMinLength     = 3
	defaultSlugMaxLength     = 30
	defaultCookieMaxAge      = 24 * time.Hour
	defaultCookieSameSite    = "Lax"
)

var defaultSlugSuffixes = []string{"-hq", "-io", "-team", "-app", "-co"}

var globalConfig *Config

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Attempt to load .env file; ignore if not present
	_ = godotenv.Load()

	cfg := &Config{
		Server:   loadServerConfig(),
		Database: loadDatabaseConfig(),
		JWT:      loadJWTConfig(),
		Auth:     loadAuthConfig(),
		Logging:  loadLoggingConfig(),
		Slug:     loadSlugConfig(),
		Platform: loadPlatformConfig(),
		Cookie:   loadCookieConfig(),
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	globalConfig = cfg
	return cfg, nil
}

// Get returns the global config instance
func Get() *Config {
	return globalConfig
}

func loadServerConfig() ServerConfig {
	return ServerConfig{
		Port:    getEnvOrDefault("SERVER_PORT", defaultServerPort),
		GinMode: getEnvOrDefault("GIN_MODE", ""),
		Env:     strings.ToLower(getEnvOrDefault("ENV", "local")),
	}
}

func loadDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		User:            os.Getenv("DB_USER"),
		Password:        os.Getenv("DB_PASSWORD"),
		Host:            os.Getenv("DB_HOST"),
		Port:            os.Getenv("DB_PORT"),
		Name:            os.Getenv("DB_NAME"),
		MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", defaultDBMaxOpenConns),
		MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", defaultDBMaxIdleConns),
		ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", defaultDBConnMaxLifetime),
	}
}

func loadJWTConfig() JWTConfig {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = defaultJWTSecret
	}

	return JWTConfig{
		Secret:      secret,
		DefaultTTL:  getEnvAsDuration("JWT_TTL", defaultJWTTTL),
		DevFallback: getEnvOrDefault("JWT_DEV_SECRET", defaultJWTSecret),
	}
}

func loadAuthConfig() AuthConfig {
	env := strings.ToLower(os.Getenv("ENV"))
	allowDevSuperadmin := env == "local" || env == "dev" || env == ""

	return AuthConfig{
		SuperadminDevToken: getEnvOrDefault("SUPERADMIN_DEV_TOKEN", defaultDevSuperadminTok),
		AllowDevSuperadmin: allowDevSuperadmin,
	}
}

func loadLoggingConfig() LoggingConfig {
	return LoggingConfig{
		Level: strings.ToLower(getEnvOrDefault("LOG_LEVEL", "")),
	}
}

func loadSlugConfig() SlugConfig {
	suffixes := defaultSlugSuffixes
	if envSuffixes := os.Getenv("SLUG_SUGGESTION_SUFFIXES"); envSuffixes != "" {
		parsed := parseSlugSuffixes(envSuffixes)
		if len(parsed) > 0 {
			suffixes = parsed
		}
	}

	return SlugConfig{
		MinLength: getEnvAsInt("SLUG_MIN_LENGTH", defaultSlugMinLength),
		MaxLength: getEnvAsInt("SLUG_MAX_LENGTH", defaultSlugMaxLength),
		Suffixes:  suffixes,
	}
}

func (c *Config) validate() error {
	if c.Database.User == "" || c.Database.Password == "" ||
		c.Database.Host == "" || c.Database.Port == "" ||
		c.Database.Name == "" {
		return fmt.Errorf("database configuration incomplete: ensure DB_USER, DB_PASSWORD, DB_HOST, DB_PORT, and DB_NAME are set")
	}
	return nil
}

// DSN returns the MySQL connection string with UTC timezone enforcement
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true&loc=UTC",
		c.User, c.Password, c.Host, c.Port, c.Name)
}

// Helper functions

func getEnvOrDefault(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if duration, err := time.ParseDuration(val); err == nil {
			return duration
		}
	}
	return defaultValue
}

func parseSlugSuffixes(value string) []string {
	parts := strings.Split(value, ",")
	suffixes := make([]string, 0, len(parts))

	for _, part := range parts {
		suffix := strings.TrimSpace(part)
		if suffix == "" {
			continue
		}
		if !strings.HasPrefix(suffix, "-") {
			suffix = "-" + suffix
		}
		suffixes = append(suffixes, suffix)
	}

	return suffixes
}

func loadPlatformConfig() PlatformConfig {
	sidebarLinksJSON := strings.TrimSpace(os.Getenv("PLATFORM_BRAND_SIDEBAR_LINKS"))
	var parsedLinks []PlatformNavItem

	if sidebarLinksJSON != "" {
		// Parse JSON, ignore errors and use defaults
		_ = parseJSONLinks(sidebarLinksJSON, &parsedLinks)
	}

	// Default links if none provided or parsing failed
	if len(parsedLinks) == 0 {
		parsedLinks = []PlatformNavItem{
			{Key: "overview", Label: "Overview", Path: "/admin"},
			{Key: "tenants", Label: "Tenants", Path: "/admin/tenants"},
			{Key: "provisioning", Label: "Provisioning", Path: "/admin/provisioning"},
		}
	}

	return PlatformConfig{
		BrandName:          strings.TrimSpace(os.Getenv("PLATFORM_BRAND_NAME")),
		BrandLogoURL:       strings.TrimSpace(os.Getenv("PLATFORM_BRAND_LOGO_URL")),
		BrandPrimaryColor:  strings.TrimSpace(os.Getenv("PLATFORM_BRAND_PRIMARY_COLOR")),
		BrandAccentColor:   strings.TrimSpace(os.Getenv("PLATFORM_BRAND_ACCENT_COLOR")),
		BrandNavbarTitle:   strings.TrimSpace(os.Getenv("PLATFORM_BRAND_NAVBAR_TITLE")),
		BrandSidebarTitle:  strings.TrimSpace(os.Getenv("PLATFORM_BRAND_SIDEBAR_TITLE")),
		BrandSidebarLinks:  sidebarLinksJSON,
		ParsedSidebarLinks: parsedLinks,
	}
}

func parseJSONLinks(jsonStr string, target *[]PlatformNavItem) error {
	return json.Unmarshal([]byte(jsonStr), target)
}

func loadCookieConfig() CookieConfig {
	env := strings.ToLower(getEnvOrDefault("ENV", "local"))

	// In local/dev, don't require HTTPS; in prod, require it
	secure := env != "local" && env != "dev"

	return CookieConfig{
		Domain:   getEnvOrDefault("COOKIE_DOMAIN", ""),
		Secure:   getEnvAsBool("COOKIE_SECURE", secure),
		SameSite: getEnvOrDefault("COOKIE_SAMESITE", defaultCookieSameSite),
		MaxAge:   getEnvAsDuration("COOKIE_MAX_AGE", defaultCookieMaxAge),
	}
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if val := os.Getenv(key); val != "" {
		return val == "true" || val == "1" || val == "yes"
	}
	return defaultValue
}
