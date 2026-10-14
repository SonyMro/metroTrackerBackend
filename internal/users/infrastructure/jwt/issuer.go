package jwt

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
	authdomain "metry.uk/provext4/MetroTrackerBackend/internal/auth/domain"
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

func (i *Issuer) Issue(_ context.Context, identity authdomain.Identity) (authdomain.IssuedJWT, error) {
	if identity.UserID <= 0 {
		return authdomain.IssuedJWT{}, domain.ErrInvalidUser
	}
	now := time.Now()
	expiresAt := now.Add(i.expiration)
	claims := jwtv5.MapClaims{"user_id": identity.UserID, "iat": now.Unix(), "exp": expiresAt.Unix()}
	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims)
	signed, err := token.SignedString(i.secret)
	if err != nil {
		return authdomain.IssuedJWT{}, fmt.Errorf("firmar JWT: %w", err)
	}
	return authdomain.IssuedJWT{Token: signed, ExpiresAt: expiresAt}, nil
}

func (i *Issuer) Verify(_ context.Context, raw string) (*authdomain.Identity, error) {
	claims := jwtv5.MapClaims{}
	token, err := jwtv5.ParseWithClaims(raw, claims, func(token *jwtv5.Token) (any, error) {
		if token.Method != jwtv5.SigningMethodHS256 {
			return nil, errors.New("unexpected JWT algorithm")
		}
		return i.secret, nil
	}, jwtv5.WithValidMethods([]string{jwtv5.SigningMethodHS256.Alg()}))
	if err != nil || token == nil || !token.Valid {
		return nil, authdomain.ErrInvalidSession
	}
	userID, ok := numericClaim(claims, "user_id")
	if !ok || userID <= 0 {
		return nil, authdomain.ErrInvalidSession
	}
	iat, ok := numericClaim(claims, "iat")
	if !ok || iat <= 0 {
		return nil, authdomain.ErrInvalidSession
	}
	exp, ok := numericClaim(claims, "exp")
	if !ok || exp <= time.Now().Unix() {
		return nil, authdomain.ErrInvalidSession
	}
	return &authdomain.Identity{UserID: int32(userID), ExpiresAt: time.Unix(exp, 0)}, nil
}

func numericClaim(claims jwtv5.MapClaims, name string) (int64, bool) {
	value := claims[name]
	switch n := value.(type) {
	case float64:
		return int64(n), n == float64(int64(n))
	case json.Number:
		parsed, err := n.Int64()
		return parsed, err == nil
	default:
		return 0, false
	}
}
