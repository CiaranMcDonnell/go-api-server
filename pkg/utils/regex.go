package utils

import "regexp"

var (
	NameRegex     = regexp.MustCompile(`^[\p{L}\p{M}' -]{2,100}$`)
	NumericRegex  = regexp.MustCompile(`^[0-9]+$`)
	PasswordRegex = regexp.MustCompile(`^.{8,128}$`)
	EmailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)
