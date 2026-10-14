package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidConfiguration = errors.New("invalid authentication configuration")
	ErrInvalidSession       = errors.New("invalid session")
	ErrSessionNotFound      = errors.New("session not found")
	ErrInactiveUser         = errors.New("user account is inactive")
)

type Identity struct {
	UserID    int32
	RoleID    int32
	ExpiresAt time.Time
}

type IssuedJWT struct {
	Token     string
	ExpiresAt time.Time
}

type JWTService interface {
	Issue(context.Context, Identity) (IssuedJWT, error)
	Verify(context.Context, string) (*Identity, error)
}

type PhantomStore interface {
	Create(context.Context, string, time.Time) (string, error)
	Resolve(context.Context, string) (string, error)
	Revoke(context.Context, string) error
}

type SessionStore interface {
	Issue(context.Context, Identity) (string, error)
}

type SessionPairIssuer interface {
	IssuePair(context.Context, Identity) (accessToken, refreshToken string, err error)
}
