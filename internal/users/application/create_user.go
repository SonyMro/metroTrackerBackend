package application

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
	authdomain "metry.uk/provext4/MetroTrackerBackend/internal/auth/domain"
	"metry.uk/provext4/MetroTrackerBackend/internal/users/domain"
)

type CreateUser struct {
	repository domain.UserRepository
	sessions   authdomain.SessionStore
}

type CreateUserResult struct {
	User         *domain.User
	Token        string
	RefreshToken string
	AccessToken  string
}

func NewCreateUser(repository domain.UserRepository, sessions ...authdomain.SessionStore) *CreateUser {
	var sessionStore authdomain.SessionStore
	if len(sessions) > 0 {
		sessionStore = sessions[0]
	}
	return &CreateUser{
		repository: repository,
		sessions:   sessionStore,
	}
}

func (uc *CreateUser) Execute(
	ctx context.Context,
	user *domain.User,
) (*CreateUserResult, error) {
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
	result := &CreateUserResult{User: created}
	if uc.sessions != nil {
		identity := authdomain.Identity{UserID: created.ID, RoleID: created.IDRol}
		if pairIssuer, ok := uc.sessions.(authdomain.SessionPairIssuer); ok {
			result.AccessToken, result.RefreshToken, err = pairIssuer.IssuePair(ctx, identity)
			result.Token = result.AccessToken
		} else {
			result.Token, err = uc.sessions.Issue(ctx, identity)
			result.AccessToken = result.Token
		}
		if err != nil {
			return nil, fmt.Errorf("emitir sesión de usuario: %w", err)
		}
	}
	return result, nil
}

func normalizeNames(user *domain.User) {
	user.Nombre = strings.TrimSpace(user.Nombre)
	user.Apepat = strings.TrimSpace(user.Apepat)
	user.Apemat = strings.TrimSpace(user.Apemat)
}
