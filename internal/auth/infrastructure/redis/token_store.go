package redis

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"
	authdomain "metry.uk/provext4/MetroTrackerBackend/internal/auth/domain"
)

type TokenStore struct {
	client *goredis.Client
	prefix string
	ttl    time.Duration
}

func NewTokenStore(client *goredis.Client, prefix string, ttl time.Duration) (*TokenStore, error) {
	if client == nil || strings.TrimSpace(prefix) == "" || ttl <= 0 {
		return nil, authdomain.ErrInvalidConfiguration
	}
	return &TokenStore{client: client, prefix: strings.TrimSuffix(prefix, ":"), ttl: ttl}, nil
}

func (s *TokenStore) key(token string) string { return s.prefix + ":" + token }

func (s *TokenStore) Create(ctx context.Context, signedJWT string, expiresAt time.Time) (string, error) {
	if strings.TrimSpace(signedJWT) == "" {
		return "", authdomain.ErrInvalidSession
	}
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return "", authdomain.ErrInvalidSession
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	key := s.key(token)
	if err := s.client.Set(ctx, key, signedJWT, ttl).Err(); err != nil {
		return "", err
	}
	return token, nil
}

func (s *TokenStore) Resolve(ctx context.Context, token string) (string, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return "", authdomain.ErrInvalidSession
	}
	value, err := s.client.Get(ctx, s.key(token)).Result()
	if err == goredis.Nil {
		return "", authdomain.ErrSessionNotFound
	}
	if err != nil {
		return "", err
	}
	return value, nil
}

func (s *TokenStore) Revoke(ctx context.Context, token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return authdomain.ErrInvalidSession
	}
	return s.client.Del(ctx, s.key(token)).Err()
}
