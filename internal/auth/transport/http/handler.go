package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"metry.uk/provext4/MetroTrackerBackend/internal/auth/application"
	authdomain "metry.uk/provext4/MetroTrackerBackend/internal/auth/domain"
	"metry.uk/provext4/MetroTrackerBackend/internal/platform/httpx"
	userapplication "metry.uk/provext4/MetroTrackerBackend/internal/users/application"
	userdomain "metry.uk/provext4/MetroTrackerBackend/internal/users/domain"
)

type Handler struct {
	googleLogin  GoogleLoginUseCase
	authenticate *application.AuthenticateToken
}

type GoogleLoginUseCase interface {
	Execute(context.Context, string) (*userapplication.GoogleLoginResult, error)
}

func NewHandler(googleLogin GoogleLoginUseCase, authenticate *application.AuthenticateToken) *Handler {
	return &Handler{googleLogin: googleLogin, authenticate: authenticate}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/google", h.google)
	r.Get("/introspect", h.introspect)
	r.Post("/introspect", h.introspect)
	r.Post("/logout", h.logout)
	return r
}

func (h *Handler) google(w http.ResponseWriter, r *http.Request) {
	var request GoogleLoginRequest
	if err := decodeJSON(r, &request); err != nil || request.IDToken == "" {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	result, err := h.googleLogin.Execute(r.Context(), request.IDToken)
	if err != nil {
		h.writeError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, GoogleLoginResponse{Token: result.Token, User: toUserResponse(result.User)})
}

func (h *Handler) introspect(w http.ResponseWriter, r *http.Request) {
	token, ok := bearerToken(r)
	if !ok || h.authenticate == nil {
		httpx.WriteJSON(w, http.StatusOK, IntrospectResponse{Active: false})
		return
	}
	identity, err := h.authenticate.Execute(r.Context(), token)
	if err != nil {
		httpx.WriteJSON(w, http.StatusOK, IntrospectResponse{Active: false})
		return
	}
	httpx.WriteJSON(w, http.StatusOK, IntrospectResponse{Active: true, UserID: identity.UserID, RoleID: identity.RoleID, ExpiresAt: identity.ExpiresAt})
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	token, ok := bearerToken(r)
	if !ok || h.authenticate == nil {
		httpx.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	if err := h.authenticate.Revoke(r.Context(), token); err != nil {
		h.writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, userdomain.ErrInvalidGoogleToken):
		httpx.WriteError(w, http.StatusUnauthorized, "invalid Google token")
	case errors.Is(err, userdomain.ErrInvalidGoogleClaims):
		httpx.WriteError(w, http.StatusBadRequest, "invalid Google identity")
	case errors.Is(err, userdomain.ErrInactiveUser), errors.Is(err, authdomain.ErrInactiveUser):
		httpx.WriteError(w, http.StatusForbidden, "user account is inactive")
	case errors.Is(err, authdomain.ErrInvalidSession), errors.Is(err, authdomain.ErrSessionNotFound):
		httpx.WriteError(w, http.StatusUnauthorized, "invalid authentication token")
	default:
		httpx.WriteError(w, http.StatusInternalServerError, "internal server error")
	}
}

func bearerToken(r *http.Request) (string, bool) {
	const prefix = "Bearer "
	header := r.Header.Get("Authorization")
	if len(header) < len(prefix) || header[:len(prefix)] != prefix {
		return "", false
	}
	token := header[len(prefix):]
	return token, token != "" && token[0] != ' '
}

type GoogleLoginRequest struct {
	IDToken string `json:"idToken"`
}
type GoogleLoginResponse struct {
	Token string        `json:"token"`
	User  *UserResponse `json:"user"`
}
type UserResponse struct {
	ID       int32  `json:"id"`
	Nombre   string `json:"nombre"`
	Apepat   string `json:"apepat"`
	Apemat   string `json:"apemat"`
	Correo   string `json:"correo"`
	IDRol    int32  `json:"idRol"`
	IDGenero int32  `json:"idGenero"`
	EsActivo bool   `json:"esActivo"`
}

func toUserResponse(user *userdomain.User) *UserResponse {
	if user == nil {
		return nil
	}
	return &UserResponse{ID: user.ID, Nombre: user.Nombre, Apepat: user.Apepat, Apemat: user.Apemat, Correo: user.Correo, IDRol: user.IDRol, IDGenero: user.IDGenero, EsActivo: user.EsActivo}
}

func decodeJSON(r *http.Request, target any) error {
	return json.NewDecoder(r.Body).Decode(target)
}
