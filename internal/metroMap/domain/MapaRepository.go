package domain

import "context"

type MapaRepository interface {
	GetPoints(ctx context.Context) ([]*PuntoMap, error)
}
