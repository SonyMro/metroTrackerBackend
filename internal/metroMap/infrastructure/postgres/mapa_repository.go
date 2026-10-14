package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"metry.uk/provext4/MetroTrackerBackend/internal/metroMap/domain"
)

type MapaRepository struct{ db *pgxpool.Pool }

func NewMapaRepository(db *pgxpool.Pool) *MapaRepository { return &MapaRepository{db: db} }

func scanListPoint(row pgx.Row) (*domain.PuntoMap, error) {
	punto := &domain.PuntoMap{}
	err := row.Scan(&punto.Linea, &punto.Estacion, &punto.Habilitada, &punto.Longitud, &punto.Latitud, &punto.Color)
	return punto, err
}

func (r *MapaRepository) GetPoints(ctx context.Context) ([]*domain.PuntoMap, error) {
	rows, err := r.db.Query(ctx, `SELECT l.nombre AS linea, e.nombre AS estacion, e.habilitada, e.longitud, e.latitud, l.color  FROM estacion AS e
		INNER JOIN  estacionlinea AS el ON el.idestacion = e.id
		INNER JOIN linea AS l ON l.id= el.idlinea
		ORDER BY l.numero, el.numero;`)

	if err != nil {
		return nil, fmt.Errorf("consultar lista usuarios: %w", err)
	}
	defer rows.Close()
	puntos := make([]*domain.PuntoMap, 0)
	for rows.Next() {

		punto, err := scanListPoint(rows)
		if err != nil {
			return nil, fmt.Errorf("leer puntos listados: %w", err)

		}
		puntos = append(puntos, punto)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterar lista de puntos: %w", err)
	}
	return puntos, nil
}
