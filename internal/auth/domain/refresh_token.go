package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrExpiredRefreshToken = errors.New("expired refresh token")
	ErrReusedRefreshToken  = errors.New("reused refresh token")
)

type RefreshToken struct {
	Selector  string
	UserID    int32
	Hash      string
	ExpiresAt time.Time
	RevokedAt *time.Time
	Secret    string // transient; never persisted
}

type RefreshTokenRepository interface {
	Rotate(context.Context, string, RefreshToken) (*RefreshToken, error)
}
