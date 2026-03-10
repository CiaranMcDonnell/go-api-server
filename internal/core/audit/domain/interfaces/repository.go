package interfaces

import (
	"context"

	"github.com/ciaranmcdonnell/go-api-server/internal/core/audit/domain/models"
)

type AuditRepository interface {
	Create(ctx context.Context, log *models.AuditLog) error
	CreateBatch(ctx context.Context, logs []*models.AuditLog) error
	FindByFilter(ctx context.Context, filter models.AuditLogFilter) ([]*models.AuditLog, error)
}
