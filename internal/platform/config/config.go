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
	// Load local development values without overriding variables from the shell.
	// A missing .env is valid in production, where the environment is injected externally.
	err := godotenv.Load()
	if err != nil {
		// Es útil manejar el error por si el archivo no existe.
		// En desarrollo, podrías querer advertir al usuario.
		fmt.Printf("⚠️ Advertencia: No se pudo cargar el archivo .env. Usando variables de entorno del sistema. Error: %v\n", err)
	} else {
		fmt.Println("✅ Archivo .env cargado exitosamente.")
	}
	cfg := Config{
		AppEnv:      getEnv("APP_ENV", "development"),
		HTTPAddr:    getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
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
