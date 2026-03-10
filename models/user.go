package models

import "time"

// minimal structs for speed and allowing manual extraction when needed.
type User struct {
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	Email          string    `json:"email,omitempty"`
	Password       string    `json:"password,omitempty"` // used only for input
	HashedPassword string    `json:"-"`                  // Used in the database
	Role           string    `json:"role,omitempty"`
	CreatedAt      time.Time `json:"created_at,omitempty"`
}
