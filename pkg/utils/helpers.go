package utils

import (
	"strconv"
)

// ParseAndValidateLimit converts a string limit to an integer and validates it against allowed values.
// Returns the default limit if conversion fails or the limit is not allowed.
func ParseAndValidateLimit(limitArg string) int {
	limit, err := strconv.Atoi(limitArg)
	if err != nil || !AllowedPaginationLimits[limit] {
		return DefaultPaginationLimit
	}
	return limit
}
