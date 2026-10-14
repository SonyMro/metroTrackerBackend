package http

import "time"

type IntrospectResponse struct {
	Active    bool      `json:"active"`
	UserID    int32     `json:"userId,omitempty"`
	ExpiresAt time.Time `json:"expiresAt,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type AuthUserResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Nombre       string `json:"nombre"`
	Correo       string `json:"correo"`
}
