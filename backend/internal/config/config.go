package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server
	Port        string
	Environment string
	FrontendURL string

	// Database
	DatabaseURL string

	// Redis (optional)
	RedisURL string

	// JWT
	JWTSecret            string
	JWTAccessExpireMin   int
	JWTRefreshExpireDays int

	// Encryption
	EncryptionKey string

	// Slack OAuth
	SlackClientID     string
	SlackClientSecret string
	SlackRedirectURI  string

	// Rate Limiting
	RateLimitRequests      int
	RateLimitWindowMinutes int

	// Cleanup
	CleanupIntervalHours int
	ChatRetentionDays    int
}

func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	cfg := &Config{
		Port:                   getEnv("PORT", "8080"),
		Environment:            getEnv("ENV", "development"),
		FrontendURL:            getEnv("FRONTEND_URL", "http://localhost:5173"),
		DatabaseURL:            getEnv("DATABASE_URL", ""),
		RedisURL:               getEnv("REDIS_URL", ""),
		JWTSecret:              getEnv("JWT_SECRET", ""),
		JWTAccessExpireMin:     getEnvAsInt("JWT_ACCESS_EXPIRE_MINUTES", 60),
		JWTRefreshExpireDays:   getEnvAsInt("JWT_REFRESH_EXPIRE_DAYS", 30),
		EncryptionKey:          getEnv("ENCRYPTION_KEY", ""),
		SlackClientID:          getEnv("SLACK_CLIENT_ID", ""),
		SlackClientSecret:      getEnv("SLACK_CLIENT_SECRET", ""),
		SlackRedirectURI:       getEnv("SLACK_REDIRECT_URI", ""),
		RateLimitRequests:      getEnvAsInt("RATE_LIMIT_REQUESTS", 100),
		RateLimitWindowMinutes: getEnvAsInt("RATE_LIMIT_WINDOW_MINUTES", 15),
		CleanupIntervalHours:   getEnvAsInt("CLEANUP_INTERVAL_HOURS", 24),
		ChatRetentionDays:      getEnvAsInt("CHAT_RETENTION_DAYS", 60),
	}

	// Validate required fields
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if len(cfg.JWTSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	if cfg.EncryptionKey == "" {
		return nil, fmt.Errorf("ENCRYPTION_KEY is required")
	}
	if len(cfg.EncryptionKey) != 32 {
		return nil, fmt.Errorf("ENCRYPTION_KEY must be exactly 32 bytes")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
