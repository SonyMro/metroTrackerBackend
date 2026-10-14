package application

import (
	"context"
	"fmt"
	authdomain "metry.uk/provext4/MetroTrackerBackend/internal/auth/domain"
	"time"
)

type SessionService struct {
	jwt     authdomain.JWTService
	phantom authdomain.PhantomStore
	refresh authdomain.RefreshTokenRepository
	ttl     time.Duration
}

func NewSessionService(jwt authdomain.JWTService, phantom authdomain.PhantomStore, refresh authdomain.RefreshTokenRepository, ttl time.Duration) *SessionService {
	return &SessionService{jwt: jwt, phantom: phantom, refresh: refresh, ttl: ttl}
}

func (s *SessionService) Issue(ctx context.Context, identity authdomain.Identity) (string, error) {
	issued, err := s.jwt.Issue(ctx, identity)
	if err != nil {
		return "", fmt.Errorf("issue internal JWT: %w", err)
	}
	opaque, err := s.phantom.Create(ctx, issued.Token, issued.ExpiresAt)
	if err != nil {
		return "", fmt.Errorf("create Phantom session: %w", err)
	}
	return opaque, nil
}

func (s *SessionService) IssuePair(ctx context.Context, identity authdomain.Identity) (string, string, error) {
	if s.refresh == nil {
		return "", "", authdomain.ErrInvalidConfiguration
	}
	access, err := s.Issue(ctx, identity)
	if err != nil {
		return "", "", err
	}
	refresh, err := newRefreshToken(s.ttl, identity.UserID)
	if err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}
	if _, err := s.refresh.Rotate(ctx, "", refresh); err != nil {
		return "", "", fmt.Errorf("persist refresh token: %w", err)
	}
	return access, refreshPublic(refresh), nil
}
