package http

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"metry.uk/provext4/MetroTrackerBackend/internal/auth/application"
	authdomain "metry.uk/provext4/MetroTrackerBackend/internal/auth/domain"
	"metry.uk/provext4/MetroTrackerBackend/internal/platform/httpx"
)

type identityKey struct{}

type AuthMiddleware struct {
	authenticate *application.AuthenticateToken
}

func NewAuthMiddleware(authenticate *application.AuthenticateToken) *AuthMiddleware {
	return &AuthMiddleware{authenticate: authenticate}
}

func IdentityFromContext(ctx context.Context) (*authdomain.Identity, bool) {
	identity, ok := ctx.Value(identityKey{}).(*authdomain.Identity)
	return identity, ok
}

func (m *AuthMiddleware) Require(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m == nil || m.authenticate == nil {
			httpx.WriteError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		header := r.Header.Get("Authorization")
		const prefix = "Bearer "
		token := strings.TrimPrefix(header, prefix)
		if !strings.HasPrefix(header, prefix) || token == "" || strings.TrimSpace(token) != token || strings.ContainsAny(token, " \t\r\n") {
			httpx.WriteError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		identity, err := m.authenticate.Execute(r.Context(), token)
		if err != nil {
			status := http.StatusUnauthorized
			if errors.Is(err, authdomain.ErrInactiveUser) {
				status = http.StatusForbidden
			}
			httpx.WriteError(w, status, "invalid authentication token")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), identityKey{}, identity)))
	})
}
