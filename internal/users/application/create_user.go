package application

import (
	"context"
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
	user.Nombre = strings.TrimSpace(user.Nombre)
	user.Apepat = strings.TrimSpace(user.Apepat)
	user.Apemat = strings.TrimSpace(user.Apemat)
	user.Correo = strings.ToLower(
		strings.TrimSpace(user.Correo),
	)

	if user.Nombre == "" ||
		user.Apepat == "" ||
		user.Correo == "" {
		return nil, domain.ErrInvalidUser
	}

	if len(user.Password) < 8 {
		return nil, domain.ErrInvalidPassword
	}

	existing, err := uc.repository.GetByEmail(
		ctx,
		user.Correo,
	)

	if err != nil &&
		err != domain.ErrUserNotFound {
		return nil, err
	}

	if existing != nil {
		return nil, domain.ErrEmailAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(user.Password),
		bcrypt.DefaultCost,
	)

	if err != nil {
		return nil, err
	}

	user.Password = string(hashedPassword)

	user.EsActivo = true

	return uc.repository.Create(ctx, user)
}
