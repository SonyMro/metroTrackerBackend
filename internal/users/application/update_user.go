package application

import (
	"context"
	"strings"

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
	idRol int32,
	idGenero int32,
) error {
	user, err := uc.repository.GetById(
		ctx,
		id,
	)

	if err != nil {
		return err
	}

	user.Nombre = strings.TrimSpace(nombre)
	user.Apepat = strings.TrimSpace(apepat)
	user.Apemat = strings.TrimSpace(apemat)
	user.IDRol = idRol
	user.IDGenero = idGenero

	if user.Nombre == "" ||
		user.Apepat == "" {
		return domain.ErrInvalidUser
	}

	if idRol <= 0 || idGenero <= 0 {
		return domain.ErrInvalidUser
	}

	return uc.repository.Update(
		ctx,
		user,
	)
}
