package application

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	authdomain "metry.uk/provext4/MetroTrackerBackend/internal/auth/domain"
	userdomain "metry.uk/provext4/MetroTrackerBackend/internal/users/domain"
	"strings"
	"time"
)

type LoginResult struct {
	AccessToken  string
	RefreshToken string
	Nombre       string
	Correo       string
}
type Login struct {
	users      userdomain.UserRepository
	sessions   *SessionService
	refresh    authdomain.RefreshTokenRepository
	refreshTTL time.Duration
}

func NewLogin(users userdomain.UserRepository, sessions *SessionService, refresh authdomain.RefreshTokenRepository, refreshTTL time.Duration) *Login {
	return &Login{users: users, sessions: sessions, refresh: refresh, refreshTTL: refreshTTL}
}

func (uc *Login) Execute(ctx context.Context, email, password string) (*LoginResult, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	user, err := uc.users.GetByEmail(ctx, email)
	if err != nil || user == nil || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		return nil, authdomain.ErrInvalidSession
	}
	if !user.EsActivo {
		return nil, authdomain.ErrInactiveUser
	}
	return uc.issue(ctx, user)
}

func (uc *Login) issue(ctx context.Context, user *userdomain.User) (*LoginResult, error) {
	access, err := uc.sessions.Issue(ctx, authdomain.Identity{UserID: user.ID, RoleID: user.IDRol})
	if err != nil {
		return nil, err
	}
	refresh, err := newRefreshToken(uc.refreshTTL, user.ID)
	if err != nil {
		return nil, err
	}
	if _, err = uc.refresh.Rotate(ctx, "", refresh); err != nil {
		return nil, fmt.Errorf("persist refresh token: %w", err)
	}
	return &LoginResult{AccessToken: access, RefreshToken: refreshPublic(refresh), Nombre: user.Nombre, Correo: user.Correo}, nil
}

func newRefreshToken(ttl time.Duration, userID int32) (authdomain.RefreshToken, error) {
	if ttl <= 0 {
		return authdomain.RefreshToken{}, authdomain.ErrInvalidConfiguration
	}
	selector, secret, err := randomPair()
	if err != nil {
		return authdomain.RefreshToken{}, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
	if err != nil {
		return authdomain.RefreshToken{}, err
	}
	return authdomain.RefreshToken{Selector: selector, UserID: userID, Hash: string(hash), Secret: secret, ExpiresAt: time.Now().Add(ttl)}, nil
}
func randomPair() (string, string, error) {
	a := make([]byte, 32)
	b := make([]byte, 32)
	if _, err := rand.Read(a); err != nil {
		return "", "", err
	}
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	return base64.RawURLEncoding.EncodeToString(a), base64.RawURLEncoding.EncodeToString(b), nil
}
func refreshPublic(t authdomain.RefreshToken) string { return t.Selector + "." + t.Secret }
