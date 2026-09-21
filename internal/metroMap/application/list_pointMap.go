package application

import (
	"context"
	"fmt"

	"metry.uk/provext4/MetroTrackerBackend/internal/metroMap/domain"
)

type ListPoint struct {
	repository domain.MapaRepository
}

func NewListPoint(repository domain.MapaRepository) *ListPoint {
	return &ListPoint{repository: repository}
}

func (uc *ListPoint) Execute(ctx context.Context) ([]*domain.PuntoMap, error) {
	items, err := uc.repository.GetPoints(ctx)
	if err != nil {
		return nil, fmt.Errorf("Lista de puntos %w", err)
	}
	return items, nil

}
