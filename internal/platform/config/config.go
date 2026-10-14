package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config reúne la configuración de la aplicación y de sus servicios externos.
type Config struct {
	AppEnv          string
	HTTPAddr        string
	DatabaseURL     string
	RedisURL        string
	DBMaxConns      int32
	DBMinConns      int32
	GoogleClientID  string
	JWTSecret       string
	JWTExpiration   time.Duration
	AuthTokenTTL    time.Duration
	AuthTokenPrefix string
	DefaultRoleID   int32
	DefaultGenderID int32
}

func Load() (Config, error) {

	// .env es opcional; sus valores se cargan antes de consultar el entorno, que tiene precedencia.
	if err := godotenv.Load("env/.env.development", "env/.env"); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("cargar .env: %w", err)
	}
	// Los helpers aplican valores por defecto; la URL actual de PostgreSQL permanece definida en el código.
	cfg := Config{
		AppEnv:          getEnv("APP_ENV", "development"),
		HTTPAddr:        getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		RedisURL:        getEnv("REDIS_URL", "redis://localhost:6379/0"),
		DBMaxConns:      getEnvInt32("DB_MAX_CONNS", 20),
		DBMinConns:      getEnvInt32("DB_MIN_CONNS", 2),
		GoogleClientID:  os.Getenv("GOOGLE_CLIENT_ID"),
		JWTSecret:       os.Getenv("JWT_SECRET"),
		JWTExpiration:   getEnvDuration("JWT_EXPIRATION", 24*time.Hour),
		AuthTokenTTL:    getEnvDuration("AUTH_TOKEN_TTL", 24*time.Hour),
		AuthTokenPrefix: getEnv("AUTH_TOKEN_PREFIX", "phantom"),
		DefaultRoleID:   getStrictEnvInt32("DEFAULT_ROLE_ID", 1),
		DefaultGenderID: getStrictEnvInt32("DEFAULT_GENDER_ID", 1),
	}
	log.Println("El ambiente es:", cfg.AppEnv)
	// Se validan la conexión de PostgreSQL y los límites permitidos del pool.
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
	if cfg.GoogleClientID == "" {
		return Config{}, errors.New("GOOGLE_CLIENT_ID is required")
	}
	if len(cfg.JWTSecret) < 32 {
		return Config{}, errors.New("JWT_SECRET must contain at least 32 characters")
	}
	if cfg.JWTExpiration <= 0 {
		return Config{}, errors.New("JWT_EXPIRATION must be greater than zero")
	}
	if cfg.AuthTokenTTL <= 0 {
		return Config{}, errors.New("AUTH_TOKEN_TTL must be greater than zero")
	}
	if cfg.AuthTokenPrefix == "" {
		return Config{}, errors.New("AUTH_TOKEN_PREFIX must not be empty")
	}
	if cfg.DefaultRoleID <= 0 || cfg.DefaultGenderID <= 0 {
		return Config{}, errors.New("DEFAULT_ROLE_ID and DEFAULT_GENDER_ID must be greater than zero")
	}
	return cfg, nil

}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0
	}
	return parsed
}

// getEnv devuelve el valor del entorno o fallback cuando la variable está vacía.
func getEnv(key, fallback string) string {

	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value

}

// getEnvInt32 devuelve el entero del entorno o fallback si está vacío o no se puede convertir.
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

func getStrictEnvInt32(key string, fallback int32) int32 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0
	}
	return int32(parsed)
}
