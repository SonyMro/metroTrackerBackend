package application

import (
	"context"
	"fmt"

	"metry.uk/provext4/MetroTrackerBackend/internal/users/domain"
)

type UpdateUser struct {
	repository domain.UserRepository
}

func NewUpdateUser(
	repository domain.UserRepository,
) *UpdateUser {
	return &UpdateUser{
		repository: repository,
	}
}

func (uc *UpdateUser) Execute(
	ctx context.Context,
	id int32,
	nombre string,
	apepat string,
	apemat string,
	idGenero int32,
) error {
	if id <= 0 || idGenero <= 0 {
		return domain.ErrInvalidUser
	}
	user, err := uc.repository.GetByID(ctx, id)

	if err != nil {
		return fmt.Errorf("obtener usuario para actualizar: %w", err)
	}

	user.Nombre = nombre
	user.Apepat = apepat
	user.Apemat = apemat
	normalizeNames(user)
	user.IDGenero = idGenero

	if user.Nombre == "" ||
		user.Apepat == "" {
		return domain.ErrInvalidUser
	}

	if err := uc.repository.Update(ctx, user); err != nil {
		return fmt.Errorf("actualizar usuario: %w", err)
	}
	return nil
}
