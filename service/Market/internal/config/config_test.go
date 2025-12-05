package config

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetEnv_Default(t *testing.T) {
	result := getEnv("NON_EXISTENT_VAR_12345", "default_value")
	assert.Equal(t, "default_value", result)
}

func TestGetEnv_ExistingVar(t *testing.T) {
	os.Setenv("TEST_VAR_CONFIG", "test_value")
	defer os.Unsetenv("TEST_VAR_CONFIG")

	result := getEnv("TEST_VAR_CONFIG", "default")
	assert.Equal(t, "test_value", result)
}

func TestGetEnv_EmptyDefault(t *testing.T) {
	result := getEnv("NON_EXISTENT_VAR_67890", "")
	assert.Equal(t, "", result)
}

func TestLoad_DefaultConfig(t *testing.T) {
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("JWT_ACCESS_SECRET", "testsecret")
	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("JWT_ACCESS_SECRET")
	}()

	cfg, err := Load(context.Background())
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "testuser", cfg.Database.User)
	assert.Equal(t, "testdb", cfg.Database.Name)
}

func TestLoad_StrictMode(t *testing.T) {
	os.Setenv("STRICT_MODE", "true")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("JWT_ACCESS_SECRET", "testsecret")
	defer func() {
		os.Unsetenv("STRICT_MODE")
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("JWT_ACCESS_SECRET")
	}()

	cfg, err := Load(context.Background())
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.True(t, cfg.Strict)
}

func TestDatabaseConfig_Defaults(t *testing.T) {
	dbConfig := DatabaseConfig{
		Host:              "localhost",
		Port:              5432,
		User:              "user",
		Password:          "password",
		Name:              "database",
		SSLMode:           "disable",
		MaxConns:          10,
		MinConns:          2,
		MaxConnLifetime:   time.Hour,
		MaxConnIdleTime:   30 * time.Minute,
		HealthCheckPeriod: time.Minute,
		QueryTimeout:      5 * time.Second,
	}

	assert.Equal(t, "localhost", dbConfig.Host)
	assert.Equal(t, 5432, dbConfig.Port)
	assert.Equal(t, int32(10), dbConfig.MaxConns)
	assert.Equal(t, int32(2), dbConfig.MinConns)
}

func TestHTTPConfig_Defaults(t *testing.T) {
	httpConfig := HTTPConfig{
		Host:            ":8080",
		ShutdownTimeout: 30 * time.Second,
		RequestTimeout:  10 * time.Second,
	}

	assert.Equal(t, ":8080", httpConfig.Host)
	assert.Equal(t, 30*time.Second, httpConfig.ShutdownTimeout)
	assert.Equal(t, 10*time.Second, httpConfig.RequestTimeout)
}

func TestLoggerConfig_Levels(t *testing.T) {
	levels := []string{"debug", "info", "warn", "error"}

	for _, level := range levels {
		loggerConfig := LoggerConfig{Level: level}
		assert.Equal(t, level, loggerConfig.Level)
	}
}

func TestJWTConfig_Secret(t *testing.T) {
	jwtConfig := JWTConfig{AccessSecret: "my-super-secret-key"}
	assert.Equal(t, "my-super-secret-key", jwtConfig.AccessSecret)
}

func TestRedisConfig_Enabled(t *testing.T) {
	redisConfig := RedisConfig{
		Enabled:  true,
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}

	assert.True(t, redisConfig.Enabled)
	assert.Equal(t, "localhost:6379", redisConfig.Addr)
	assert.Equal(t, 0, redisConfig.DB)
}

func TestRedisConfig_Disabled(t *testing.T) {
	redisConfig := RedisConfig{Enabled: false}
	assert.False(t, redisConfig.Enabled)
}

func TestRateLimitConfig_Enabled(t *testing.T) {
	rateLimitConfig := RateLimitConfig{
		Enabled:  true,
		Max:      100,
		Interval: time.Minute,
	}

	assert.True(t, rateLimitConfig.Enabled)
	assert.Equal(t, 100, rateLimitConfig.Max)
	assert.Equal(t, time.Minute, rateLimitConfig.Interval)
}

func TestRateLimitConfig_Disabled(t *testing.T) {
	rateLimitConfig := RateLimitConfig{Enabled: false}
	assert.False(t, rateLimitConfig.Enabled)
}

func TestConfig_FullStruct(t *testing.T) {
	cfg := Config{
		Strict: false,
		Database: DatabaseConfig{
			Host: "db.example.com",
			Port: 5432,
		},
		HTTP: HTTPConfig{
			Host: ":8080",
		},
		Logger: LoggerConfig{
			Level: "info",
		},
		JWT: JWTConfig{
			AccessSecret: "secret",
		},
		Redis: RedisConfig{
			Enabled: true,
			Addr:    "redis:6379",
		},
		RateLimit: RateLimitConfig{
			Enabled:  true,
			Max:      100,
			Interval: time.Second,
		},
		UploadDir: "/uploads",
		BaseURL:   "http://localhost:8080",
	}

	assert.Equal(t, "db.example.com", cfg.Database.Host)
	assert.Equal(t, ":8080", cfg.HTTP.Host)
	assert.Equal(t, "info", cfg.Logger.Level)
	assert.Equal(t, "/uploads", cfg.UploadDir)
	assert.Equal(t, "http://localhost:8080", cfg.BaseURL)
}
