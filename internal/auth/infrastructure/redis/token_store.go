package redis

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"strconv"
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

func (s *TokenStore) Issue(ctx context.Context, identity authdomain.Identity) (string, error) {
	if identity.UserID <= 0 {
		return "", authdomain.ErrInvalidSession
	}
	if identity.ExpiresAt.IsZero() {
		identity.ExpiresAt = time.Now().Add(s.ttl)
	}
	ttl := time.Until(identity.ExpiresAt)
	if ttl <= 0 {
		return "", authdomain.ErrInvalidSession
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	key := s.key(token)
	pipe := s.client.TxPipeline()
	pipe.HSet(ctx, key, "user_id", identity.UserID, "role_id", identity.RoleID, "expires_at", identity.ExpiresAt.Unix())
	pipe.Expire(ctx, key, ttl)
	if _, err := pipe.Exec(ctx); err != nil {
		return "", err
	}
	return token, nil
}

func (s *TokenStore) Authenticate(ctx context.Context, token string) (*authdomain.Identity, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, authdomain.ErrInvalidSession
	}
	values, err := s.client.HGetAll(ctx, s.key(token)).Result()
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, authdomain.ErrSessionNotFound
	}
	userID, err := strconv.ParseInt(values["user_id"], 10, 32)
	if err != nil || userID <= 0 {
		return nil, authdomain.ErrInvalidSession
	}
	roleID, _ := strconv.ParseInt(values["role_id"], 10, 32)
	expires, err := strconv.ParseInt(values["expires_at"], 10, 64)
	if err != nil || time.Now().Unix() >= expires {
		return nil, authdomain.ErrSessionNotFound
	}
	return &authdomain.Identity{UserID: int32(userID), RoleID: int32(roleID), ExpiresAt: time.Unix(expires, 0)}, nil
}

func (s *TokenStore) Revoke(ctx context.Context, token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return authdomain.ErrInvalidSession
	}
	return s.client.Del(ctx, s.key(token)).Err()
}
