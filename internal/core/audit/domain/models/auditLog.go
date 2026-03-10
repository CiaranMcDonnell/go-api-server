package models

import "time"

type AuditLog struct {
	ID                  string    `json:"id"`
	UserID              *int      `json:"user_id"`
	Username            string    `json:"username,omitempty"`
	AttemptedIdentifier string    `json:"attempted_identifier,omitempty"`
	Action              string    `json:"action"`
	Resource            string    `json:"resource"`
	EntityID            *string   `json:"entity_id,omitempty"`
	EntityType          *string   `json:"entity_type,omitempty"`
	RequestPath         string    `json:"request_path"`
	Method              string    `json:"method"`
	StatusCode          int       `json:"status_code"`
	IPAddress           string    `json:"ip_address"`
	UserAgent           string    `json:"user_agent,omitempty"`
	RequestBody         string    `json:"request_body,omitempty"`
	Timestamp           time.Time `json:"timestamp"`
}

type CreateAuditLogDTO struct {
	UserID              *int
	AttemptedIdentifier string
	Action              string
	Resource            string
	EntityID            *string
	EntityType          *string
	RequestPath         string
	Method              string
	StatusCode          int
	IPAddress           string
	UserAgent           string
	RequestBody         string
}

type AuditLogFilter struct {
	UserID              *int
	AttemptedIdentifier *string
	StartDate           *time.Time
	EndDate             *time.Time
	Resource            *string
	Action              *string
	EntityID            *string
	EntityType          *string
	Limit               *int
	Offset              *int
}
