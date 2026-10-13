package application

import (
	"context"
	"strings"

	authdomain "metry.uk/provext4/MetroTrackerBackend/internal/auth/domain"
)

type UserReader interface {
	AuthUser(context.Context, int32) (roleID int32, active bool, err error)
}

type AuthenticateToken struct {
	sessions authdomain.SessionStore
	users    UserReader
}

func NewAuthenticateToken(sessions authdomain.SessionStore, users UserReader) *AuthenticateToken {
	return &AuthenticateToken{sessions: sessions, users: users}
}

func (uc *AuthenticateToken) Execute(ctx context.Context, token string) (*authdomain.Identity, error) {
	if strings.TrimSpace(token) == "" {
		return nil, authdomain.ErrInvalidSession
	}
	identity, err := uc.sessions.Authenticate(ctx, token)
	if err != nil {
		return nil, err
	}
	roleID, active, err := uc.users.AuthUser(ctx, identity.UserID)
	if err != nil {
		return nil, authdomain.ErrInvalidSession
	}
	if !active {
		return nil, authdomain.ErrInactiveUser
	}
	identity.RoleID = roleID
	return identity, nil
}

func (uc *AuthenticateToken) Revoke(ctx context.Context, token string) error {
	if strings.TrimSpace(token) == "" {
		return authdomain.ErrInvalidSession
	}
	return uc.sessions.Revoke(ctx, token)
}
