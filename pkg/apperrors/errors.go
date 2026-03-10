package apperrors

import "errors"

var (
	ErrNotFound           = errors.New("not found")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrConflict           = errors.New("conflict")
)
