package interfaces

import (
	"context"

	"github.com/ciaranmcdonnell/go-api-server/internal/core/audit/domain/models"
)

type AuditService interface {
	LogAuditEvent(ctx context.Context, dto models.CreateAuditLogDTO) error
	GetAuditLogs(ctx context.Context, filter models.AuditLogFilter) ([]*models.AuditLog, error)
}
