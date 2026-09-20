package application

import (
	"context"
	"fmt"

	"metry.uk/provext4/MetroTrackerBackend/internal/users/domain"
)

type ListGenders struct {
	repository domain.UserRepository
}

func NewListGenders(repository domain.UserRepository) *ListGenders {
	return &ListGenders{repository: repository}
}

func (uc *ListGenders) Execute(ctx context.Context) ([]*domain.Genero, error) {
	items, err := uc.repository.ListGeneros(ctx)
	if err != nil {
		return nil, fmt.Errorf("Lista generos %w", err)
	}
	return items, nil
}
