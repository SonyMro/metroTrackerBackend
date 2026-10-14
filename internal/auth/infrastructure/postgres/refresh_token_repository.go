package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	authdomain "metry.uk/provext4/MetroTrackerBackend/internal/auth/domain"
	"strings"
	"time"
)

type RefreshTokenRepository struct{ db *pgxpool.Pool }

func NewRefreshTokenRepository(db *pgxpool.Pool) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Rotate(ctx context.Context, presented string, replacement authdomain.RefreshToken) (*authdomain.RefreshToken, error) {
	parts := strings.Split(presented, ".")
	if len(parts) != 2 && presented != "" {
		return nil, authdomain.ErrInvalidRefreshToken
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin refresh rotation: %w", err)
	}
	defer tx.Rollback(ctx)
	if presented == "" {
		if _, err = tx.Exec(ctx, `INSERT INTO refresh_tokens (selector, token_hash, user_id, expires_at) VALUES ($1,$2,$3,$4)`, replacement.Selector, replacement.Hash, replacement.UserID, replacement.ExpiresAt); err != nil {
			return nil, fmt.Errorf("insert refresh token: %w", err)
		}
		if err = tx.Commit(ctx); err != nil {
			return nil, err
		}
		return &replacement, nil
	}
	var old authdomain.RefreshToken
	var hash string
	var revoked *time.Time
	err = tx.QueryRow(ctx, `SELECT selector,user_id,token_hash,expires_at,revoked_at FROM refresh_tokens WHERE selector=$1 FOR UPDATE`, parts[0]).Scan(&old.Selector, &old.UserID, &hash, &old.ExpiresAt, &revoked)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, authdomain.ErrInvalidRefreshToken
	}
	if err != nil {
		return nil, fmt.Errorf("query refresh token: %w", err)
	}
	if revoked != nil {
		return nil, authdomain.ErrReusedRefreshToken
	}
	if time.Now().After(old.ExpiresAt) {
		return nil, authdomain.ErrExpiredRefreshToken
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(parts[1])) != nil {
		return nil, authdomain.ErrInvalidRefreshToken
	}
	if _, err = tx.Exec(ctx, `UPDATE refresh_tokens SET revoked_at=NOW() WHERE selector=$1 AND revoked_at IS NULL`, old.Selector); err != nil {
		return nil, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO refresh_tokens (selector, token_hash, user_id, expires_at) VALUES ($1,$2,$3,$4)`, replacement.Selector, replacement.Hash, old.UserID, replacement.ExpiresAt); err != nil {
		return nil, fmt.Errorf("insert rotated refresh token: %w", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &old, nil
}
