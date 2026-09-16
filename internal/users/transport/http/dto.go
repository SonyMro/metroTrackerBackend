package http

import (
	"time"

	"metry.uk/provext4/MetroTrackerBackend/internal/users/domain"
)

type CreateUserRequest struct {
	Nombre string `json:"nombre"`

	Apepat string `json:"apepat"`

	Apemat string `json:"apemat"`

	Correo string `json:"correo"`

	Password string `json:"password"`

	IDRol int32 `json:"idRol"`

	IDGenero int32 `json:"idGenero"`
}

type UpdateUserRequest struct {
	Nombre string `json:"nombre"`

	Apepat string `json:"apepat"`

	Apemat string `json:"apemat"`

	IDRol int32 `json:"idRol"`

	IDGenero int32 `json:"idGenero"`
}

type UserResponse struct {
	ID int32 `json:"id"`

	Nombre string `json:"nombre"`

	Apepat string `json:"apepat"`

	Apemat string `json:"apemat"`

	Correo string `json:"correo"`

	FechaRegistro time.Time `json:"fechaRegistro"`

	FechaAct time.Time `json:"fechaAct"`

	IDRol int32 `json:"idRol"`

	Rol *RoleResponse `json:"rol,omitempty"`

	IDGenero int32 `json:"idGenero"`

	Genero *GenderResponse `json:"genero,omitempty"`

	EsActivo bool `json:"esActivo"`
}

type RoleResponse struct {
	ID int32 `json:"id"`

	Nombre string `json:"nombre"`
}

type GenderResponse struct {
	ID int32 `json:"id"`

	Tipo string `json:"tipo"`
}

type UserListResponse struct {
	Items []*UserResponse `json:"items"`

	Page int `json:"page"`

	PageSize int `json:"pageSize"`

	Total int `json:"total"`

	TotalPages int `json:"totalPages"`
}

func toUserResponse(
	user *domain.User,
) *UserResponse {
	response := &UserResponse{
		ID:            user.ID,
		Nombre:        user.Nombre,
		Apepat:        user.Apepat,
		Apemat:        user.Apemat,
		Correo:        user.Correo,
		FechaRegistro: user.FechaRegistro,
		FechaAct:      user.FechaAct,
		IDRol:         user.IDRol,
		IDGenero:      user.IDGenero,
		EsActivo:      user.EsActivo,
	}

	if user.Rol != nil {
		response.Rol = &RoleResponse{
			ID:     user.Rol.ID,
			Nombre: user.Rol.Nombre,
		}
	}

	if user.Genero != nil {
		response.Genero = &GenderResponse{
			ID:   user.Genero.ID,
			Tipo: user.Genero.Tipo,
		}
	}

	return response
}
