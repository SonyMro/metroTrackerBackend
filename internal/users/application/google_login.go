package application

import (
	"context"
	"errors"
	"fmt"
	"strings"

	authdomain "metry.uk/provext4/MetroTrackerBackend/internal/auth/domain"
	"metry.uk/provext4/MetroTrackerBackend/internal/users/domain"
)

type GoogleLogin struct {
	verifier        domain.GoogleTokenVerifier
	identities      domain.GoogleIdentityRepository
	issuer          authdomain.SessionStore
	defaultRoleID   int32
	defaultGenderID int32
}

func NewGoogleLogin(verifier domain.GoogleTokenVerifier, identities domain.GoogleIdentityRepository, issuer authdomain.SessionStore, defaultRoleID, defaultGenderID int32) *GoogleLogin {
	return &GoogleLogin{verifier: verifier, identities: identities, issuer: issuer, defaultRoleID: defaultRoleID, defaultGenderID: defaultGenderID}
}

type GoogleLoginResult struct {
	Token string
	User  *domain.User
}

func (uc *GoogleLogin) Execute(ctx context.Context, rawToken string) (*GoogleLoginResult, error) {
	if strings.TrimSpace(rawToken) == "" {
		return nil, domain.ErrInvalidGoogleToken
	}
	claims, err := uc.verifier.Verify(ctx, rawToken)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidGoogleClaims) {
			return nil, domain.ErrInvalidGoogleClaims
		}
		return nil, fmt.Errorf("verificar token Google: %w", errors.Join(domain.ErrInvalidGoogleToken, err))
	}
	claims.Subject = strings.TrimSpace(claims.Subject)
	claims.Email = strings.ToLower(strings.TrimSpace(claims.Email))
	claims.Name = strings.TrimSpace(claims.Name)
	if claims.Subject == "" || claims.Email == "" {
		return nil, domain.ErrInvalidGoogleClaims
	}
	user, err := uc.identities.ResolveGoogleIdentity(ctx, claims, uc.defaultRoleID, uc.defaultGenderID)
	if err != nil {
		return nil, fmt.Errorf("resolver identidad Google: %w", err)
	}
	if user == nil || !user.EsActivo {
		return nil, domain.ErrInactiveUser
	}
	token, err := uc.issuer.Issue(ctx, authdomain.Identity{UserID: user.ID, RoleID: user.IDRol})
	if err != nil {
		return nil, fmt.Errorf("emitir token de aplicación: %w", err)
	}
	return &GoogleLoginResult{Token: token, User: user}, nil
}
