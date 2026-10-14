package application

import (
	"context"
	"fmt"
	authdomain "metry.uk/provext4/MetroTrackerBackend/internal/auth/domain"
	userdomain "metry.uk/provext4/MetroTrackerBackend/internal/users/domain"
	"time"
)

type Refresh struct {
	users      userdomain.UserRepository
	sessions   *SessionService
	repository authdomain.RefreshTokenRepository
	ttl        time.Duration
}

func NewRefresh(users userdomain.UserRepository, sessions *SessionService, repository authdomain.RefreshTokenRepository, ttl time.Duration) *Refresh {
	return &Refresh{users: users, sessions: sessions, repository: repository, ttl: ttl}
}
func (uc *Refresh) Execute(ctx context.Context, presented string) (*LoginResult, error) {
	newToken, err := newRefreshToken(uc.ttl, 0)
	if err != nil {
		return nil, authdomain.ErrInvalidRefreshToken
	}
	old, err := uc.repository.Rotate(ctx, presented, newToken)
	if err != nil {
		return nil, err
	}
	user, err := uc.users.GetByID(ctx, old.UserID)
	if err != nil || user == nil {
		return nil, authdomain.ErrInvalidRefreshToken
	}
	if !user.EsActivo {
		return nil, authdomain.ErrInactiveUser
	}
	newToken.UserID = user.ID
	access, err := uc.sessions.Issue(ctx, authdomain.Identity{UserID: user.ID, RoleID: user.IDRol})
	if err != nil {
		return nil, fmt.Errorf("issue access token: %w", err)
	}
	return &LoginResult{AccessToken: access, RefreshToken: refreshPublic(newToken)}, nil
}
