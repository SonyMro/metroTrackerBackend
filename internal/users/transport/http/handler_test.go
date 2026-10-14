package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	authhttp "metry.uk/provext4/MetroTrackerBackend/internal/auth/transport/http"
	"metry.uk/provext4/MetroTrackerBackend/internal/users/application"
	"metry.uk/provext4/MetroTrackerBackend/internal/users/domain"
)

type googleLoginFake struct {
	result *application.GoogleLoginResult
	err    error
	token  string
}

func (f googleLoginFake) Execute(_ context.Context, token string) (*application.GoogleLoginResult, error) {
	f.token = token
	return f.result, f.err
}

func TestHandlerRejectsInvalidJSONAndID(t *testing.T) {
	handler := NewHandler(nil, nil, nil, nil, nil, nil)
	tests := []struct{ name, method, target, body string }{
		{"campo desconocido", http.MethodPost, "/", `{"nombre":"Ana","extra":true}`},
		{"objetos concatenados", http.MethodPost, "/", `{} {}`},
		{"ID no positivo", http.MethodGet, "/0", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.target, strings.NewReader(tt.body))
			recorder := httptest.NewRecorder()
			handler.Routes().ServeHTTP(recorder, req)
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, quiere %d", recorder.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestWriteDomainError(t *testing.T) {
	tests := []struct {
		err  error
		want int
	}{
		{domain.ErrUserNotFound, http.StatusNotFound},
		{domain.ErrEmailAlreadyExists, http.StatusConflict},
		{domain.ErrInvalidUser, http.StatusBadRequest},
		{domain.ErrInvalidPassword, http.StatusBadRequest},
		{domain.ErrInvalidGoogleToken, http.StatusUnauthorized},
		{domain.ErrInvalidGoogleClaims, http.StatusBadRequest},
		{domain.ErrInactiveUser, http.StatusForbidden},
		{errors.New("infraestructura"), http.StatusInternalServerError},
	}
	for _, tt := range tests {
		recorder := httptest.NewRecorder()
		(&Handler{}).writeDomainError(recorder, tt.err)
		if recorder.Code != tt.want {
			t.Fatalf("status = %d, quiere %d", recorder.Code, tt.want)
		}
	}
}

func TestGoogleLoginEndpoint(t *testing.T) {
	user := &domain.User{ID: 4, Nombre: "Ana", Correo: "ana@example.com", EsActivo: true}
	handler := authhttp.NewHandler(googleLoginFake{result: &application.GoogleLoginResult{AccessToken: "opaque-token", RefreshToken: "refresh-token", User: user}}, nil)
	tests := []struct {
		name, body string
		wantStatus int
	}{
		{"body inválido", `{`, http.StatusBadRequest},
		{"token ausente", `{}`, http.StatusBadRequest},
		{"correcto", `{"idToken":"google"}`, http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler.Routes().ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/google", strings.NewReader(tt.body)))
			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, quiere %d", recorder.Code, tt.wantStatus)
			}
			if tt.wantStatus == http.StatusOK {
				var response authhttp.AuthUserResponse
				if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil || response.AccessToken != "opaque-token" || response.RefreshToken != "refresh-token" || response.Nombre != "Ana" || response.Correo != "ana@example.com" {
					t.Fatalf("respuesta inesperada: %s", recorder.Body.String())
				}
			}
		})
	}
}
