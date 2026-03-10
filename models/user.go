package models

import "time"

// Domain model — no JSON tags, used internally by services and repositories.
type User struct {
	ID             int64
	Name           string
	Email          string
	HashedPassword string
	Role           string
	CreatedAt      time.Time
}

// API input for user registration.
type CreateUserRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100,name"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=128"`
}

// API input for user login.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// API output — safe to serialize, excludes sensitive fields.
type UserResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// ToResponse converts a domain User to an API response.
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
	}
}
