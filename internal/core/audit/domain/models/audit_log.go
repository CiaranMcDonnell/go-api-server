package models

import "time"

// Domain model — no JSON tags, used internally by services and repositories.
type AuditLog struct {
	ID                  string
	UserID              *int
	Username            string
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
	Timestamp           time.Time
}

// API output — safe to serialize.
type AuditLogResponse struct {
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

// API input for creating an audit log.
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

// ToResponse converts a domain AuditLog to an API response.
func (a *AuditLog) ToResponse() *AuditLogResponse {
	return &AuditLogResponse{
		ID:                  a.ID,
		UserID:              a.UserID,
		Username:            a.Username,
		AttemptedIdentifier: a.AttemptedIdentifier,
		Action:              a.Action,
		Resource:            a.Resource,
		EntityID:            a.EntityID,
		EntityType:          a.EntityType,
		RequestPath:         a.RequestPath,
		Method:              a.Method,
		StatusCode:          a.StatusCode,
		IPAddress:           a.IPAddress,
		UserAgent:           a.UserAgent,
		RequestBody:         a.RequestBody,
		Timestamp:           a.Timestamp,
	}
}
