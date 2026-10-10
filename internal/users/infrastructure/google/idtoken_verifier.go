package google

import (
	"context"
	"fmt"

	"google.golang.org/api/idtoken"
	"metry.uk/provext4/MetroTrackerBackend/internal/users/domain"
)

type IDTokenVerifier struct{ clientID string }

func NewIDTokenVerifier(clientID string) (*IDTokenVerifier, error) {
	if clientID == "" {
		return nil, domain.ErrInvalidConfiguration
	}
	return &IDTokenVerifier{clientID: clientID}, nil
}

func (v *IDTokenVerifier) Verify(ctx context.Context, rawToken string) (domain.GoogleClaims, error) {
	payload, err := idtoken.Validate(ctx, rawToken, v.clientID)
	if err != nil {
		return domain.GoogleClaims{}, fmt.Errorf("validar idToken: %w", err)
	}
	if payload.Subject == "" || payload.Claims["email"] == nil {
		return domain.GoogleClaims{}, domain.ErrInvalidGoogleClaims
	}
	email, ok := payload.Claims["email"].(string)
	if !ok || email == "" {
		return domain.GoogleClaims{}, domain.ErrInvalidGoogleClaims
	}
	name, _ := payload.Claims["name"].(string)
	return domain.GoogleClaims{Subject: payload.Subject, Email: email, Name: name}, nil
}
