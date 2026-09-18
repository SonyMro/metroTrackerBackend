package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"metry.uk/provext4/MetroTrackerBackend/internal/users/domain"
)

func TestHandlerRejectsInvalidJSONAndID(t *testing.T) {
	handler := NewHandler(nil, nil, nil, nil, nil)
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
