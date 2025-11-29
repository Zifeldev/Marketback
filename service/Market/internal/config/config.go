package config

import (
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
	Addr     string
	Password string
	DB       int
}

type Config struct {
	Strict   bool
	Database DatabaseConfig
	HTTP     HTTPConfig
	Logger   LoggerConfig
	JWT      JWTConfig
	Redis    RedisConfig
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsInt32(key string, defaultValue int32) int32 {
	valueStr := os.Getenv(key)
	if value, err := strconv.ParseInt(valueStr, 10, 32); err == nil {
		return int32(value)
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		Strict: getEnv("STRICT_MODE", "false") == "true",
		Database: DatabaseConfig{
			Host:              getEnv("DB_HOST", "localhost"),
			Port:              getEnvAsInt("DB_PORT", 5434),
			User:              getEnv("DB_USER", "market_user"),
			Password:          getEnv("DB_PASSWORD", "market_password"),
			Name:              getEnv("DB_NAME", "market_db"),
			SSLMode:           getEnv("DB_SSLMODE", "disable"),
			MaxConns:          getEnvAsInt32("DB_MAX_CONNS", 10),
			MinConns:          getEnvAsInt32("DB_MIN_CONNS", 2),
			MaxConnLifetime:   getEnvAsDuration("DB_MAX_CONN_LIFETIME", 1*time.Hour),
			MaxConnIdleTime:   getEnvAsDuration("DB_MAX_CONN_IDLE_TIME", 30*time.Minute),
			HealthCheckPeriod: getEnvAsDuration("DB_HEALTH_CHECK_PERIOD", 1*time.Minute),
			QueryTimeout:      getEnvAsDuration("DB_QUERY_TIMEOUT", 30*time.Second),
		},
		HTTP: HTTPConfig{
			Host:            getEnv("HTTP_HOST", ":8080"),
			ShutdownTimeout: getEnvAsDuration("HTTP_SHUTDOWN_TIMEOUT", 10*time.Second),
			RequestTimeout:  getEnvAsDuration("HTTP_REQUEST_TIMEOUT", 30*time.Second),
		},
		Logger: LoggerConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
		JWT: JWTConfig{
			AccessSecret: os.Getenv("JWT_ACCESS_SECRET"),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
	}

	if cfg.JWT.AccessSecret == "" {
		return nil, fmt.Errorf("JWT_ACCESS_SECRET is required")
	}

	return cfg, nil
}
