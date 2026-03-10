package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testDTO struct {
	Name  string `json:"name" validate:"required,min=2,max=50"`
	Email string `json:"email" validate:"required,email"`
}

type optionalDTO struct {
	Name *string `json:"name" validate:"omitempty,min=1"`
}

func TestBindJSON_ValidInput(t *testing.T) {
	body := strings.NewReader(`{"name":"Alice","email":"alice@example.com"}`)
	var dto testDTO

	err := BindJSON(body, &dto)

	assert.NoError(t, err)
	assert.Equal(t, "Alice", dto.Name)
	assert.Equal(t, "alice@example.com", dto.Email)
}

func TestBindJSON_MissingRequired(t *testing.T) {
	body := strings.NewReader(`{"name":"Alice"}`)
	var dto testDTO

	err := BindJSON(body, &dto)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email is required")
}

func TestBindJSON_MultipleErrors(t *testing.T) {
	body := strings.NewReader(`{}`)
	var dto testDTO

	err := BindJSON(body, &dto)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
	assert.Contains(t, err.Error(), "email is required")
}

func TestBindJSON_InvalidEmail(t *testing.T) {
	body := strings.NewReader(`{"name":"Alice","email":"not-an-email"}`)
	var dto testDTO

	err := BindJSON(body, &dto)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email must be a valid email")
}

func TestBindJSON_TooShort(t *testing.T) {
	body := strings.NewReader(`{"name":"A","email":"a@b.com"}`)
	var dto testDTO

	err := BindJSON(body, &dto)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name must be at least 2 characters")
}

func TestBindJSON_TooLong(t *testing.T) {
	longName := strings.Repeat("A", 51)
	body := strings.NewReader(`{"name":"` + longName + `","email":"a@b.com"}`)
	var dto testDTO

	err := BindJSON(body, &dto)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name must be at most 50 characters")
}

func TestBindJSON_InvalidJSON(t *testing.T) {
	body := strings.NewReader(`{broken json`)
	var dto testDTO

	err := BindJSON(body, &dto)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid JSON format")
}

func TestBindJSON_EmptyBody(t *testing.T) {
	body := strings.NewReader(``)
	var dto testDTO

	err := BindJSON(body, &dto)

	assert.Error(t, err)
}

func TestBindJSON_OptionalFields(t *testing.T) {
	body := strings.NewReader(`{}`)
	var dto optionalDTO

	err := BindJSON(body, &dto)

	assert.NoError(t, err)
	assert.Nil(t, dto.Name)
}

// --- Custom name validator ---

type nameDTO struct {
	Name string `json:"name" validate:"required,name"`
}

func TestBindJSON_NameValidator_ValidNames(t *testing.T) {
	valid := []string{
		"Alice",
		"O'Brien",
		"Jean-Pierre",
		"María García",
		"Müller",
	}

	for _, name := range valid {
		body := strings.NewReader(`{"name":"` + name + `"}`)
		var dto nameDTO
		err := BindJSON(body, &dto)
		assert.NoError(t, err, "expected valid name: %s", name)
	}
}

func TestBindJSON_NameValidator_InvalidNames(t *testing.T) {
	invalid := []string{
		"<script>",
		"test@user",
		"user;DROP",
		"test123",
	}

	for _, name := range invalid {
		body := strings.NewReader(`{"name":"` + name + `"}`)
		var dto nameDTO
		err := BindJSON(body, &dto)
		assert.Error(t, err, "expected invalid name: %s", name)
		assert.Contains(t, err.Error(), "contains invalid characters")
	}
}
