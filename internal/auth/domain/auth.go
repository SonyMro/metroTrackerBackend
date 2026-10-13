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

type SessionStore interface {
	Issue(context.Context, Identity) (string, error)
	Authenticate(context.Context, string) (*Identity, error)
	Revoke(context.Context, string) error
}
