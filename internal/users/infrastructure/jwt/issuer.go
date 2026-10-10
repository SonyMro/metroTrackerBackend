package jwt

import (
	"fmt"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
	"metry.uk/provext4/MetroTrackerBackend/internal/users/domain"
)

type Issuer struct {
	secret     []byte
	expiration time.Duration
}

func NewIssuer(secret string, expiration time.Duration) (*Issuer, error) {
	if len(secret) < 32 || expiration <= 0 {
		return nil, domain.ErrInvalidConfiguration
	}
	return &Issuer{secret: []byte(secret), expiration: expiration}, nil
}

func (i *Issuer) Issue(user *domain.User) (string, error) {
	if user == nil || user.ID <= 0 {
		return "", domain.ErrInvalidUser
	}
	now := time.Now()
	claims := jwtv5.MapClaims{"user_id": user.ID, "role": user.IDRol, "iat": now.Unix(), "exp": now.Add(i.expiration).Unix()}
	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims)
	signed, err := token.SignedString(i.secret)
	if err != nil {
		return "", fmt.Errorf("firmar JWT: %w", err)
	}
	return signed, nil
}
