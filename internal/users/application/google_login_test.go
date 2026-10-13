package application

import (
	"context"
	"errors"
	"testing"

	authdomain "metry.uk/provext4/MetroTrackerBackend/internal/auth/domain"
	"metry.uk/provext4/MetroTrackerBackend/internal/users/domain"
)

type googleVerifierFake struct {
	claims domain.GoogleClaims
	err    error
}

func (f googleVerifierFake) Verify(context.Context, string) (domain.GoogleClaims, error) {
	return f.claims, f.err
}

type googleIdentityFake struct {
	user  *domain.User
	err   error
	claim domain.GoogleClaims
}

func (f *googleIdentityFake) ResolveGoogleIdentity(_ context.Context, claims domain.GoogleClaims, _, _ int32) (*domain.User, error) {
	f.claim = claims
	return f.user, f.err
}

type tokenIssuerFake struct {
	token string
	err   error
}

func (f tokenIssuerFake) Issue(context.Context, authdomain.Identity) (string, error) {
	return f.token, f.err
}

func (f tokenIssuerFake) Authenticate(context.Context, string) (*authdomain.Identity, error) {
	return nil, nil
}

func (f tokenIssuerFake) Revoke(context.Context, string) error { return nil }

func TestGoogleLoginExecute(t *testing.T) {
	infraErr := errors.New("database unavailable")
	tests := []struct {
		name        string
		claims      domain.GoogleClaims
		verifyErr   error
		identity    *domain.User
		identityErr error
		issueErr    error
		wantErr     error
		wantEmail   string
	}{
		{name: "vínculo existente", claims: domain.GoogleClaims{Subject: "sub-1", Email: " USER@EXAMPLE.COM "}, identity: &domain.User{ID: 7, EsActivo: true}, wantEmail: "user@example.com"},
		{name: "token inválido", verifyErr: infraErr, wantErr: domain.ErrInvalidGoogleToken},
		{name: "datos incompletos", claims: domain.GoogleClaims{Email: "user@example.com"}, wantErr: domain.ErrInvalidGoogleClaims},
		{name: "error de persistencia", claims: domain.GoogleClaims{Subject: "sub-1", Email: "user@example.com"}, identityErr: infraErr, wantErr: infraErr},
		{name: "cuenta inactiva", claims: domain.GoogleClaims{Subject: "sub-1", Email: "user@example.com"}, identity: &domain.User{ID: 7}, wantErr: domain.ErrInactiveUser},
		{name: "error al emitir", claims: domain.GoogleClaims{Subject: "sub-1", Email: "user@example.com"}, identity: &domain.User{ID: 7, EsActivo: true}, issueErr: infraErr, wantErr: infraErr},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			identity := &googleIdentityFake{user: tt.identity, err: tt.identityErr}
			uc := NewGoogleLogin(googleVerifierFake{claims: tt.claims, err: tt.verifyErr}, identity, tokenIssuerFake{token: "app-token", err: tt.issueErr}, 2, 3)
			result, err := uc.Execute(context.Background(), "google-token")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, quiere %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				return
			}
			if result.Token != "app-token" || result.User.ID != 7 || identity.claim.Email != tt.wantEmail {
				t.Fatalf("resultado inesperado: %#v, claim=%#v", result, identity.claim)
			}
		})
	}
}
