package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeStruct_TrimsWhitespace(t *testing.T) {
	type dto struct {
		Name string `json:"name"`
	}
	d := &dto{Name: "  Alice  "}
	SanitizeStruct(d)
	assert.Equal(t, "Alice", d.Name)
}

func TestSanitizeStruct_NormalizesEmail(t *testing.T) {
	type dto struct {
		Email string `json:"email" validate:"required,email"`
	}
	d := &dto{Email: "  Alice@Example.COM  "}
	SanitizeStruct(d)
	assert.Equal(t, "alice@example.com", d.Email)
}

func TestSanitizeStruct_HandlesPointerStrings(t *testing.T) {
	name := "  Bob  "
	type dto struct {
		Name *string `json:"name"`
	}
	d := &dto{Name: &name}
	SanitizeStruct(d)
	assert.Equal(t, "Bob", *d.Name)
}

func TestSanitizeStruct_HandlesNilPointers(t *testing.T) {
	type dto struct {
		Name *string `json:"name"`
	}
	d := &dto{Name: nil}
	SanitizeStruct(d) // should not panic
	assert.Nil(t, d.Name)
}

func TestSanitizeStruct_EmailByJsonTag(t *testing.T) {
	type dto struct {
		Email string `json:"email"`
	}
	d := &dto{Email: "  User@TEST.com  "}
	SanitizeStruct(d)
	assert.Equal(t, "user@test.com", d.Email)
}

func TestSanitizeStruct_MultipleFields(t *testing.T) {
	type dto struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"required,email"`
	}
	d := &dto{Name: "  Alice  ", Email: "  ALICE@EXAMPLE.COM  "}
	SanitizeStruct(d)
	assert.Equal(t, "Alice", d.Name)
	assert.Equal(t, "alice@example.com", d.Email)
}

func TestSanitizeStruct_NonPointerNoOp(t *testing.T) {
	type dto struct {
		Name string `json:"name"`
	}
	d := dto{Name: "  Alice  "}
	SanitizeStruct(d) // non-pointer, should be no-op
	assert.Equal(t, "  Alice  ", d.Name)
}
