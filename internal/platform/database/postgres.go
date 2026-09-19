package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Config contiene los parámetros usados para configurar el pool PostgreSQL.
type Config struct {
	URL      string
	MaxConns int32
	MinConns int32
}

func NewPostgres(
	ctx context.Context,
	cfg Config,
) (*pgxpool.Pool, error) {
	// Se interpreta la URL, se aplican los límites y tiempos del pool, y se crea la conexión.
	poolConfig, err := pgxpool.ParseConfig(cfg.URL)

	if err != nil {
		return nil, fmt.Errorf("interpretar configuración PostgreSQL: %w", err)
	}

	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns

	// La renovación, el límite de inactividad y la comprobación periódica evitan conexiones obsoletas.
	poolConfig.MaxConnLifetime = 30 * time.Minute
	poolConfig.MaxConnIdleTime = 5 * time.Minute
	poolConfig.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(
		ctx,
		poolConfig,
	)

	if err != nil {
		return nil, fmt.Errorf("crear pool PostgreSQL: %w", err)
	}

	// Ping valida la conectividad inicial; si falla, el pool se cierra antes de devolver el error.
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("verificar pool PostgreSQL: %w", err)
	}

	return pool, nil
}
