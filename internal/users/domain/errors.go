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

	ErrInvalidGoogleToken      = errors.New("invalid Google token")
	ErrInvalidGoogleClaims     = errors.New("incomplete Google claims")
	ErrInvalidConfiguration    = errors.New("invalid authentication configuration")
	ErrGoogleLinkAlreadyExists = errors.New("Google identity is already linked")
	ErrInactiveUser            = errors.New("user account is inactive")
)
