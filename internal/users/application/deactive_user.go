package application

import (
	"context"
	"fmt"

	"metry.uk/provext4/MetroTrackerBackend/internal/users/domain"
)

type DeactivateUser struct {
	repository domain.UserRepository
}

func NewDeactivateUser(
	repository domain.UserRepository,
) *DeactivateUser {
	return &DeactivateUser{
		repository: repository,
	}
}

func (uc *DeactivateUser) Execute(
	ctx context.Context,
	id int32,
) error {
	if id <= 0 {
		return domain.ErrUserNotFound
	}

	if err := uc.repository.Deactivate(ctx, id); err != nil {
		return fmt.Errorf("desactivar usuario: %w", err)
	}
	return nil
}
