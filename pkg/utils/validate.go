package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

var nameRegex = regexp.MustCompile(`^[\p{L}\p{M}' -]+$`)

func init() {
	validate = validator.New()

	validate.RegisterValidation("name", func(fl validator.FieldLevel) bool {
		return nameRegex.MatchString(fl.Field().String())
	})
}

// BindJSON parses a JSON body into the target struct and runs struct validation.
// Returns nil on success, or a user-friendly error string.
func BindJSON(body io.Reader, obj interface{}) error {
	data, err := io.ReadAll(body)
	if err != nil {
		if err.Error() == "http: request body too large" {
			return fmt.Errorf("Request body too large")
		}
		return fmt.Errorf("Failed to read request body")
	}

	if err := json.Unmarshal(data, obj); err != nil {
		return fmt.Errorf("Invalid JSON format")
	}

	SanitizeStruct(obj)

	if err := validate.Struct(obj); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok {
			return fmt.Errorf("%s", formatValidationErrors(errs))
		}
		return fmt.Errorf("Validation failed")
	}

	return nil
}

func formatValidationErrors(errs validator.ValidationErrors) string {
	msgs := make([]string, 0, len(errs))
	for _, e := range errs {
		field := strings.ToLower(e.Field())
		switch e.Tag() {
		case "required":
			msgs = append(msgs, fmt.Sprintf("%s is required", field))
		case "email":
			msgs = append(msgs, fmt.Sprintf("%s must be a valid email", field))
		case "min":
			msgs = append(msgs, fmt.Sprintf("%s must be at least %s characters", field, e.Param()))
		case "max":
			msgs = append(msgs, fmt.Sprintf("%s must be at most %s characters", field, e.Param()))
		case "name":
			msgs = append(msgs, fmt.Sprintf("%s contains invalid characters", field))
		default:
			msgs = append(msgs, fmt.Sprintf("%s is invalid", field))
		}
	}
	return strings.Join(msgs, "; ")
}
