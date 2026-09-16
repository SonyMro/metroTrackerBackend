package domain

import "errors"

var (
	ErrUserNotFound = errors.New(
		"user not found",
	)

	ErrEmailAlreadyExists = errors.New(
		"email already exists",
	)

	ErrInvalidUser = errors.New(
		"invalid user",
	)

	ErrInvalidPassword = errors.New(
		"invalid password",
	)
)
