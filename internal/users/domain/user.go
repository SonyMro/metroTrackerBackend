package domain

import "time"

type User struct {
	ID            int32
	Nombre        string
	Apepat        string
	Apemat        string
	Foto          []byte
	Correo        string
	GoogleID      string
	Password      string
	FechaRegistro time.Time
	FechaAct      time.Time
	IDRol         int32
	IDGenero      int32
	EsActivo      bool
	Rol           *Role
	Genero        *Genero
}
