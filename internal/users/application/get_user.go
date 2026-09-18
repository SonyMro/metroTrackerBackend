package application

import (
	"context"
	"fmt"

	"metry.uk/provext4/MetroTrackerBackend/internal/users/domain"
)

type GetUser struct {
	repository domain.UserRepository
}

func NewGetUser(
	repository domain.UserRepository,
) *GetUser {
	return &GetUser{
		repository: repository,
	}
}

func (uc *GetUser) Execute(
	ctx context.Context,
	id int32,
) (*domain.User, error) {
	if id <= 0 {
		return nil, domain.ErrUserNotFound
	}

	user, err := uc.repository.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("obtener usuario: %w", err)
	}
	return user, nil
}
