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
	phantom authdomain.PhantomStore
	jwt     authdomain.JWTService
	users   UserReader
}

func NewAuthenticateToken(phantom authdomain.PhantomStore, jwt authdomain.JWTService, users UserReader) *AuthenticateToken {
	return &AuthenticateToken{phantom: phantom, jwt: jwt, users: users}
}

func (uc *AuthenticateToken) Execute(ctx context.Context, token string) (*authdomain.Identity, error) {
	if strings.TrimSpace(token) == "" || uc == nil || uc.phantom == nil || uc.jwt == nil || uc.users == nil {
		return nil, authdomain.ErrInvalidSession
	}
	rawJWT, err := uc.phantom.Resolve(ctx, token)
	if err != nil {
		return nil, authdomain.ErrInvalidSession
	}
	identity, err := uc.jwt.Verify(ctx, rawJWT)
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
	return uc.phantom.Revoke(ctx, token)
}
