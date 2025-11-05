package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Environment string
	Server      ServerConfig
	Database    DatabaseConfig
	JWT         JWTConfig
	Cookie      CookieConfig
	Security    SecurityConfig
	Redis       RedisConfig
}

type ServerConfig struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

type DatabaseConfig struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

type JWTConfig struct {
	Secret          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type CookieConfig struct {
	RefreshCookieName string
	Domain            string
	Secure            bool
	SameSite          string
}

type SecurityConfig struct {
	MaxLoginAttempts    int
	AccountLockDuration time.Duration
	AllowedOrigins      []string
	RateLimitPerMinute  int
}

type RedisConfig struct {
	Enabled  bool
	Host     string
	Port     string
	Password string
	DB       int
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	return &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Server: ServerConfig{
			Port:            getEnv("PORT", "9001"),
			ReadTimeout:     parseDuration(getEnv("SERVER_READ_TIMEOUT", "10s")),
			WriteTimeout:    parseDuration(getEnv("SERVER_WRITE_TIMEOUT", "10s")),
			ShutdownTimeout: parseDuration(getEnv("SERVER_SHUTDOWN_TIMEOUT", "5s")),
		},
		Database: DatabaseConfig{
			URL:             getEnv("DATABASE_URL", ""),
			MaxOpenConns:    parseInt(getEnv("DB_MAX_OPEN_CONNS", "25")),
			MaxIdleConns:    parseInt(getEnv("DB_MAX_IDLE_CONNS", "5")),
			ConnMaxLifetime: parseDuration(getEnv("DB_CONN_MAX_LIFETIME", "5m")),
			ConnMaxIdleTime: parseDuration(getEnv("DB_CONN_MAX_IDLE_TIME", "5m")),
		},
		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", ""),
			AccessTokenTTL:  parseDuration(getEnv("ACCESS_TOKEN_TTL", "15m")),
			RefreshTokenTTL: parseDuration(getEnv("REFRESH_TOKEN_TTL", "720h")),
		},
		Cookie: CookieConfig{
			RefreshCookieName: getEnv("REFRESH_COOKIE_NAME", "refresh_token"),
			Domain:            getEnv("COOKIE_DOMAIN", "localhost"),
			Secure:            parseBool(getEnv("COOKIE_SECURE", "false")),
			SameSite:          getEnv("COOKIE_SAMESITE", "Lax"),
		},
		Security: SecurityConfig{
			MaxLoginAttempts:    parseInt(getEnv("MAX_LOGIN_ATTEMPTS", "5")),
			AccountLockDuration: parseDuration(getEnv("ACCOUNT_LOCK_DURATION", "15m")),
			AllowedOrigins:      parseStringSlice(getEnv("ALLOWED_ORIGINS", "http://localhost:3000")),
			RateLimitPerMinute:  parseInt(getEnv("RATE_LIMIT_PER_MINUTE", "60")),
		},
		Redis: RedisConfig{
			Enabled:  parseBool(getEnv("REDIS_ENABLED", "false")),
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       parseInt(getEnv("REDIS_DB", "0")),
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func parseBool(s string) bool {
	b, _ := strconv.ParseBool(s)
	return b
}

func parseDuration(s string) time.Duration {
	d, _ := time.ParseDuration(s)
	return d
}

func parseStringSlice(s string) []string {
	if s == "" {
		return []string{}
	}
	return []string{s}
}

func (c *Config) Validate() error {
	if c.Database.URL == "" {
		return ErrMissingDatabaseURL
	}
	if c.JWT.Secret == "" {
		return ErrMissingJWTSecret
	}
	return nil
}

var (
	ErrMissingDatabaseURL = &ConfigError{"DATABASE_URL is required"}
	ErrMissingJWTSecret   = &ConfigError{"JWT_SECRET is required"}
)

type ConfigError struct {
	Message string
}

func (e *ConfigError) Error() string {
	return e.Message
}
