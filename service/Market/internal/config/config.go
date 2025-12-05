package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

type DatabaseConfig struct {
	Host              string
	Port              int
	User              string
	Password          string
	Name              string
	SSLMode           string
	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
	QueryTimeout      time.Duration
}

type HTTPConfig struct {
	Host            string
	ShutdownTimeout time.Duration
	RequestTimeout  time.Duration
}

type LoggerConfig struct {
	Level string
}

type JWTConfig struct {
	AccessSecret string
}

type RedisConfig struct {
	Enabled  bool
	Addr     string
	Password string
	DB       int
}

type RateLimitConfig struct {
	Enabled  bool
	Max      int
	Interval time.Duration
}

type Config struct {
	Strict    bool
	Database  DatabaseConfig
	HTTP      HTTPConfig
	Logger    LoggerConfig
	JWT       JWTConfig
	Redis     RedisConfig
	RateLimit RateLimitConfig
	UploadDir string
	BaseURL   string
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func Load(ctx context.Context) (*Config, error) {
	cfg := &Config{}

	// Strict mode
	cfg.Strict = getEnv("STRICT_MODE", "false") == "true"

	// Database
	port, err := strconv.Atoi(getEnv("DB_PORT", "5434"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
	}

	maxConns, err := strconv.ParseInt(getEnv("DB_MAX_CONNS", "10"), 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid DB_MAX_CONNS: %w", err)
	}

	minConns, err := strconv.ParseInt(getEnv("DB_MIN_CONNS", "2"), 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid DB_MIN_CONNS: %w", err)
	}

	queryTimeout, err := time.ParseDuration(getEnv("DB_QUERY_TIMEOUT", "30s"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_QUERY_TIMEOUT: %w", err)
	}

	cfg.Database = DatabaseConfig{
		Host:              getEnv("DB_HOST", "localhost"),
		Port:              port,
		User:              getEnv("DB_USER", "market_user"),
		Password:          getEnv("DB_PASSWORD", "market_password"),
		Name:              getEnv("DB_NAME", "market_db"),
		SSLMode:           getEnv("DB_SSLMODE", "disable"),
		MaxConns:          int32(maxConns),
		MinConns:          int32(minConns),
		MaxConnLifetime:   time.Hour,
		MaxConnIdleTime:   30 * time.Minute,
		HealthCheckPeriod: time.Minute,
		QueryTimeout:      queryTimeout,
	}

	// HTTP
	shutdownTimeout, err := time.ParseDuration(getEnv("HTTP_SHUTDOWN_TIMEOUT", "10s"))
	if err != nil {
		return nil, fmt.Errorf("invalid HTTP_SHUTDOWN_TIMEOUT: %w", err)
	}

	requestTimeout, err := time.ParseDuration(getEnv("HTTP_REQUEST_TIMEOUT", "30s"))
	if err != nil {
		return nil, fmt.Errorf("invalid HTTP_REQUEST_TIMEOUT: %w", err)
	}

	cfg.HTTP = HTTPConfig{
		Host:            getEnv("HTTP_HOST", ":8080"),
		ShutdownTimeout: shutdownTimeout,
		RequestTimeout:  requestTimeout,
	}

	// Logger
	cfg.Logger = LoggerConfig{
		Level: getEnv("LOG_LEVEL", "info"),
	}

	// JWT
	accessSecret := getEnv("JWT_ACCESS_SECRET", "")
	if accessSecret == "" {
		return nil, errors.New("JWT_ACCESS_SECRET is required")
	}

	cfg.JWT = JWTConfig{
		AccessSecret: accessSecret,
	}

	// Redis
	redisDB, err := strconv.Atoi(getEnv("REDIS_DB", "0"))
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_DB: %w", err)
	}

	cfg.Redis = RedisConfig{
		Enabled:  getEnv("REDIS_ENABLED", "true") == "true",
		Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       redisDB,
	}

	// Rate Limit
	rateLimitInterval, err := time.ParseDuration(getEnv("RATE_LIMIT_INTERVAL", "1m"))
	if err != nil {
		return nil, fmt.Errorf("invalid RATE_LIMIT_INTERVAL: %w", err)
	}

	rateLimitMax, err := strconv.Atoi(getEnv("RATE_LIMIT_MAX", "100"))
	if err != nil {
		return nil, fmt.Errorf("invalid RATE_LIMIT_MAX: %w", err)
	}

	cfg.RateLimit = RateLimitConfig{
		Enabled:  getEnv("RATE_LIMIT_ENABLED", "true") == "true",
		Max:      rateLimitMax,
		Interval: rateLimitInterval,
	}

	// Upload settings
	cfg.UploadDir = getEnv("UPLOAD_DIR", "./uploads")
	cfg.BaseURL = getEnv("BASE_URL", "http://localhost:8080")

	return cfg, nil
}

// LoadConfig is an alias for Load for backward compatibility
func LoadConfig() (*Config, error) {
	return Load(context.Background())
}
