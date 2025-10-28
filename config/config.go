package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBUrl string
	JWTSecret string
	Port string
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	cfg := &Config{
		DBUrl:     os.Getenv("DATABASE_URL"),
		JWTSecret: os.Getenv("JWT_SECRET"),
		Port:      os.Getenv("PORT"),
	}

	return cfg
}