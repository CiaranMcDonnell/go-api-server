package repository

import (
	"context"
	"fmt"

	"github.com/ciaranmcdonnell/go-api-server/internal/core/audit/domain/interfaces"
	"github.com/ciaranmcdonnell/go-api-server/internal/core/audit/domain/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type auditRepository struct {
	db *pgxpool.Pool
}

func NewAuditRepository(db *pgxpool.Pool) interfaces.AuditRepository {
	return &auditRepository{
		db: db,
	}
}

func (r *auditRepository) Create(ctx context.Context, log *models.AuditLog) error {
	query := `INSERT INTO audit_logs (
		user_id, action, resource, request_path,
		method, status_code, ip_address, user_agent, request_body, timestamp,
		entity_id, entity_type
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	_, err := r.db.Exec(
		ctx, query,
		log.UserID, log.Action, log.Resource, log.RequestPath,
		log.Method, log.StatusCode, log.IPAddress, log.UserAgent, log.RequestBody, log.Timestamp,
		log.EntityID, log.EntityType,
	)
	return err
}

func (r *auditRepository) FindByFilter(ctx context.Context, filter models.AuditLogFilter) ([]*models.AuditLog, error) {
	var logs []*models.AuditLog

	query := `SELECT
				a.id, a.user_id, a.attempted_identifier, a.action, a.resource,
				a.entity_id, a.entity_type,
				a.request_path, a.method, a.status_code,
				a.ip_address, a.user_agent, a.request_body, a.timestamp,
				u.name as username
			 FROM audit_logs a
			 LEFT JOIN users u ON a.user_id = u.id
			 WHERE 1=1`

	args := []interface{}{}
	argIdx := 1

	if filter.UserID != nil {
		query += fmt.Sprintf(" AND a.user_id = $%d", argIdx)
		args = append(args, *filter.UserID)
		argIdx++
	}

	if filter.EntityID != nil {
		query += fmt.Sprintf(" AND a.entity_id = $%d", argIdx)
		args = append(args, *filter.EntityID)
		argIdx++
	}

	if filter.EntityType != nil {
		query += fmt.Sprintf(" AND a.entity_type = $%d", argIdx)
		args = append(args, *filter.EntityType)
		argIdx++
	}

	if filter.Resource != nil {
		query += fmt.Sprintf(" AND a.resource = $%d", argIdx)
		args = append(args, *filter.Resource)
		argIdx++
	}

	if filter.Action != nil {
		query += fmt.Sprintf(" AND a.action = $%d", argIdx)
		args = append(args, *filter.Action)
		argIdx++
	}

	if filter.StartDate != nil {
		query += fmt.Sprintf(" AND a.timestamp >= $%d", argIdx)
		args = append(args, *filter.StartDate)
		argIdx++
	}

	if filter.EndDate != nil {
		query += fmt.Sprintf(" AND a.timestamp <= $%d", argIdx)
		args = append(args, *filter.EndDate)
		argIdx++
	}

	query += " ORDER BY a.timestamp DESC"

	// Pagination
	limit := 100
	if filter.Limit != nil {
		limit = *filter.Limit
	}
	query += fmt.Sprintf(" LIMIT $%d", argIdx)
	args = append(args, limit)
	argIdx++

	if filter.Offset != nil {
		query += fmt.Sprintf(" OFFSET $%d", argIdx)
		args = append(args, *filter.Offset)
		argIdx++
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error querying audit logs: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var log models.AuditLog
		var attemptedIdentifier *string
		var username *string

		if err := rows.Scan(
			&log.ID, &log.UserID, &attemptedIdentifier, &log.Action, &log.Resource,
			&log.EntityID, &log.EntityType,
			&log.RequestPath, &log.Method, &log.StatusCode,
			&log.IPAddress, &log.UserAgent, &log.RequestBody, &log.Timestamp,
			&username,
		); err != nil {
			return nil, fmt.Errorf("error scanning audit log row: %w", err)
		}

		if attemptedIdentifier != nil {
			log.AttemptedIdentifier = *attemptedIdentifier
		}

		if username != nil {
			log.Username = *username
		}

		logs = append(logs, &log)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating audit log rows: %w", err)
	}

	return logs, nil
}
