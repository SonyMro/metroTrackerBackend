package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"metry.uk/provext4/MetroTrackerBackend/internal/auth/application"
	authdomain "metry.uk/provext4/MetroTrackerBackend/internal/auth/domain"
	"metry.uk/provext4/MetroTrackerBackend/internal/platform/httpx"
	userapplication "metry.uk/provext4/MetroTrackerBackend/internal/users/application"
	userdomain "metry.uk/provext4/MetroTrackerBackend/internal/users/domain"
)

const internalServerErrorMessage = "internal server error"

type Handler struct {
	googleLogin  GoogleLoginUseCase
	authenticate *application.AuthenticateToken
	login        *application.Login
	refresh      *application.Refresh
}

type GoogleLoginUseCase interface {
	Execute(context.Context, string) (*userapplication.GoogleLoginResult, error)
}

func NewHandler(googleLogin GoogleLoginUseCase, authenticate *application.AuthenticateToken, cases ...any) *Handler {
	h := &Handler{googleLogin: googleLogin, authenticate: authenticate}
	for _, value := range cases {
		switch v := value.(type) {
		case *application.Login:
			h.login = v
		case *application.Refresh:
			h.refresh = v
		}
	}
	return h
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/google", h.google)
	r.Post("/login", h.loginHandler)
	r.Post("/refresh", h.refreshHandler)
	r.Get("/introspect", h.introspect)
	r.Post("/introspect", h.introspect)
	r.Post("/logout", h.logout)
	return r
}

func (h *Handler) loginHandler(w http.ResponseWriter, r *http.Request) {
	if h.login == nil {
		httpx.WriteError(w, http.StatusInternalServerError, internalServerErrorMessage)
		return
	}
	var req LoginRequest
	if decodeJSON(r, &req) != nil || strings.TrimSpace(req.Email) == "" || req.Password == "" {
		httpx.WriteError(w, http.StatusBadRequest, internalServerErrorMessage)
		return
	}
	result, err := h.login.Execute(r.Context(), req.Email, req.Password)
	if err != nil {
		h.writeError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, AuthUserResponse{AccessToken: result.AccessToken, RefreshToken: result.RefreshToken, Nombre: result.Nombre, Correo: result.Correo})
}

func (h *Handler) refreshHandler(w http.ResponseWriter, r *http.Request) {
	if h.refresh == nil {
		httpx.WriteError(w, http.StatusInternalServerError, internalServerErrorMessage)
		return
	}
	var req RefreshRequest
	if decodeJSON(r, &req) != nil || strings.TrimSpace(req.RefreshToken) == "" {
		httpx.WriteError(w, http.StatusBadRequest, internalServerErrorMessage)
		return
	}
	result, err := h.refresh.Execute(r.Context(), req.RefreshToken)
	if err != nil {
		h.writeError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, TokenResponse{AccessToken: result.AccessToken, RefreshToken: result.RefreshToken})
}

func (h *Handler) google(w http.ResponseWriter, r *http.Request) {
	var request GoogleLoginRequest
	if err := decodeJSON(r, &request); err != nil || request.IDToken == "" {
		httpx.WriteError(w, http.StatusBadRequest, internalServerErrorMessage)
		return
	}
	result, err := h.googleLogin.Execute(r.Context(), request.IDToken)
	if err != nil {
		h.writeError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, AuthUserResponse{AccessToken: result.AccessToken, RefreshToken: result.RefreshToken, Nombre: result.User.Nombre, Correo: result.User.Correo})
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
	httpx.WriteJSON(w, http.StatusOK, IntrospectResponse{Active: true, UserID: identity.UserID, ExpiresAt: identity.ExpiresAt})
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
	case errors.Is(err, authdomain.ErrInvalidRefreshToken), errors.Is(err, authdomain.ErrExpiredRefreshToken), errors.Is(err, authdomain.ErrReusedRefreshToken):
		httpx.WriteError(w, http.StatusUnauthorized, "invalid refresh token")
	default:
		httpx.WriteError(w, http.StatusInternalServerError, internalServerErrorMessage)
	}
}

func bearerToken(r *http.Request) (string, bool) {
	const prefix = "Bearer "
	header := r.Header.Get("Authorization")
	if len(header) < len(prefix) || header[:len(prefix)] != prefix {
		return "", false
	}
	token := header[len(prefix):]
	return token, token != "" && strings.TrimSpace(token) == token && !strings.ContainsAny(token, " \t\r\n")
}

type GoogleLoginRequest struct {
	IDToken string `json:"idToken"`
}

// GoogleLoginResponse se conserva para compatibilidad con consumidores internos
// de la capa de transporte; el endpoint ya responde AuthUserResponse.
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
	IDGenero int32  `json:"idGenero"`
	EsActivo bool   `json:"esActivo"`
}

func decodeJSON(r *http.Request, target any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return errors.New("multiple JSON values")
	}
	return nil
}
