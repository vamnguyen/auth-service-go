package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DBUrl     string
	JWTSecret string
	Port      string

	AccessTokenTTL    time.Duration
	RefreshTokenTTL   time.Duration
	RefreshCookieName string
	CookieDomain      string
	CookieSecure      bool
	CookieSameSite    string
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	cfg := &Config{
		DBUrl:     os.Getenv("DATABASE_URL"),
		JWTSecret: os.Getenv("JWT_SECRET"),
		Port:      os.Getenv("PORT"),

		AccessTokenTTL:    parseDurationWithDefault(os.Getenv("ACCESS_TOKEN_TTL"), 15*time.Minute),
		RefreshTokenTTL:   parseDurationWithDefault(os.Getenv("REFRESH_TOKEN_TTL"), 720*time.Hour), // 30 ngày
		RefreshCookieName: firstNonEmpty(os.Getenv("REFRESH_COOKIE_NAME"), "refresh_token"),
		CookieDomain:      os.Getenv("COOKIE_DOMAIN"),
		CookieSecure:      parseBoolWithDefault(os.Getenv("COOKIE_SECURE"), false),
		CookieSameSite:    firstNonEmpty(os.Getenv("COOKIE_SAMESITE"), "Lax"),
	}

	return cfg
}

func parseDurationWithDefault(s string, defaultDur time.Duration) time.Duration {
	if d, err := time.ParseDuration(s); err == nil {
		return d
	}
	return defaultDur
}

func parseBoolWithDefault(s string, def bool) bool {
	if s == "" {
		return def
	}
	b, err := strconv.ParseBool(s)
	if err != nil {
		return def
	}
	return b
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
