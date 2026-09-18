package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv      string
	HTTPAddr    string
	DatabaseURL string
	RedisURL    string
	DBMaxConns  int32
	DBMinConns  int32
}

func Load() (Config, error) {
	// El entorno tiene precedencia; un .env ausente es normal fuera de desarrollo.
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("cargar .env: %w", err)
	}
	cfg := Config{
		AppEnv:      getEnv("APP_ENV", "development"),
		HTTPAddr:    getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL: "postgres://user_metry:erw34$23Ert@localhost:5432/dbmetry?sslmode=disable", //os.Getenv("DATABASE_URL"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379/0"),
		DBMaxConns:  getEnvInt32("DB_MAX_CONNS", 20),
		DBMinConns:  getEnvInt32("DB_MIN_CONNS", 2),
	}
	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	if cfg.DBMinConns < 0 {
		return Config{}, errors.New("DB_MIN_CONNS must be >= 0")
	}
	if cfg.DBMaxConns <= 0 {
		return Config{}, errors.New("DB_MAX_CONNS must be > 0")
	}
	if cfg.DBMinConns > cfg.DBMaxConns {
		return Config{}, errors.New(
			"DB_MIN_CONNS cannot be greater than DB_MAX_CONNS",
		)
	}
	return cfg, nil

}

func getEnv(key, fallback string) string {

	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value

}

func getEnvInt32(key string, fallback int32) int32 {

	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return fallback
	}
	return int32(parsed)

}
