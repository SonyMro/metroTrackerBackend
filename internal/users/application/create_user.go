package application

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"metry.uk/provext4/MetroTrackerBackend/internal/users/domain"
)

type CreateUser struct {
	repository domain.UserRepository
}

func NewCreateUser(
	repository domain.UserRepository,
) *CreateUser {
	return &CreateUser{
		repository: repository,
	}
}

func (uc *CreateUser) Execute(
	ctx context.Context,
	user *domain.User,
) (*domain.User, error) {
	normalizeNames(user)
	user.Correo = strings.ToLower(
		strings.TrimSpace(user.Correo),
	)

	if user.Nombre == "" ||
		user.Apepat == "" ||
		user.Correo == "" || user.IDRol <= 0 || user.IDGenero <= 0 {
		return nil, domain.ErrInvalidUser
	}

	if len(user.Password) < 8 {
		return nil, domain.ErrInvalidPassword
	}

	existing, err := uc.repository.GetByEmail(
		ctx,
		user.Correo,
	)

	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return nil, fmt.Errorf("consultar usuario por correo: %w", err)
	}

	if existing != nil {
		return nil, domain.ErrEmailAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(user.Password),
		bcrypt.DefaultCost,
	)

	if err != nil {
		return nil, fmt.Errorf("hashear contraseña: %w", err)
	}

	user.Password = string(hashedPassword)

	user.EsActivo = true

	created, err := uc.repository.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("crear usuario: %w", err)
	}
	return created, nil
}

func normalizeNames(user *domain.User) {
	user.Nombre = strings.TrimSpace(user.Nombre)
	user.Apepat = strings.TrimSpace(user.Apepat)
	user.Apemat = strings.TrimSpace(user.Apemat)
}
