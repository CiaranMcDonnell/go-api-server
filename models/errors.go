package models

import "errors"

var (
	ErrUnauthorized = errors.New("user is not authorized for this action")
)
