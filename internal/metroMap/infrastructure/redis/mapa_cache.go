// Package redis contiene los adaptadores Redis del módulo de mapas.
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"metry.uk/provext4/MetroTrackerBackend/internal/metroMap/domain"
)

const mapaCacheKey = "metro-map:points:v1"
const mapaCacheTTL = 1 * time.Hour

// MapaCacheRepository implementa cache-aside delante del repositorio principal.
// Redis se consulta primero; PostgreSQL sólo se usa cuando no existe un valor
// válido en caché, y el resultado de PostgreSQL se almacena nuevamente en Redis.
type MapaCacheRepository struct {
	primary  domain.MapaRepository
	client   *goredis.Client
	cacheKey string
}

// NewMapaCacheRepository crea el decorador de caché para los puntos del mapa.
func NewMapaCacheRepository(primary domain.MapaRepository, client *goredis.Client) *MapaCacheRepository {
	return &MapaCacheRepository{
		primary:  primary,
		client:   client,
		cacheKey: mapaCacheKey,
	}
}

// GetPoints aplica la estrategia cache-aside para obtener los puntos del mapa.
func (r *MapaCacheRepository) GetPoints(ctx context.Context) ([]*domain.PuntoMap, error) {
	// Un error de Redis, incluido un cache miss, no impide atender la petición:
	// la fuente de verdad sigue siendo el repositorio primario.
	cached, err := r.client.Get(ctx, r.cacheKey).Bytes()
	if err == nil {
		points := make([]*domain.PuntoMap, 0)
		if err := json.Unmarshal(cached, &points); err == nil {
			return points, nil
		}
	}

	points, err := r.primary.GetPoints(ctx)
	if err != nil {
		return nil, fmt.Errorf("obtener puntos desde el repositorio principal: %w", err)
	}

	// El valor se mantiene una hora; después Redis fuerza una nueva lectura
	// desde PostgreSQL cuando vuelva a solicitarse el mapa.
	payload, err := json.Marshal(points)
	if err != nil {
		return points, fmt.Errorf("serializar puntos para Redis: %w", err)
	}
	if err := r.client.Set(ctx, r.cacheKey, payload, mapaCacheTTL).Err(); err != nil {
		// No se convierte en error de negocio: la consulta primaria ya fue exitosa.
		return points, nil
	}

	return points, nil
}
