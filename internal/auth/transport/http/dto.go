package http

import "time"

type IntrospectResponse struct {
	Active    bool      `json:"active"`
	UserID    int32     `json:"userId,omitempty"`
	RoleID    int32     `json:"roleId,omitempty"`
	ExpiresAt time.Time `json:"expiresAt,omitempty"`
}
