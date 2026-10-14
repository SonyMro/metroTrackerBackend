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

	IDGenero int32 `json:"idGenero"`
}

type UpdateUserRequest struct {
	Nombre string `json:"nombre"`

	Apepat string `json:"apepat"`

	Apemat string `json:"apemat"`

	IDGenero int32 `json:"idGenero"`
}

type CreateUserResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Nombre       string `json:"nombre"`
	Correo       string `json:"correo"`
}

type UserResponse struct {
	ID int32 `json:"id"`

	Nombre string `json:"nombre"`

	Apepat string `json:"apepat"`

	Apemat string `json:"apemat"`

	Correo string `json:"correo"`

	FechaRegistro time.Time `json:"fechaRegistro"`

	FechaAct time.Time `json:"fechaAct"`

	IDGenero int32 `json:"idGenero"`

	Genero *GenderResponse `json:"genero,omitempty"`

	EsActivo bool `json:"esActivo"`
}

type GenderResponse struct {
	ID int32 `json:"id"`

	Tipo string `json:"tipo"`
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
		IDGenero:      user.IDGenero,
		EsActivo:      user.EsActivo,
	}

	if user.Genero != nil {
		response.Genero = &GenderResponse{
			ID:   user.Genero.ID,
			Tipo: user.Genero.Tipo,
		}
	}

	return response
}
