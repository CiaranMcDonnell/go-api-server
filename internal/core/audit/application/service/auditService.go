package service

import (
	"context"
	"time"

	repoInterfaces "github.com/ciaranmcdonnell/go-api-server/internal/core/audit/domain/interfaces"
	"github.com/ciaranmcdonnell/go-api-server/internal/core/audit/domain/models"
)

type auditService struct {
	repo repoInterfaces.AuditRepository
}

func NewAuditService(repo repoInterfaces.AuditRepository) repoInterfaces.AuditService {
	return &auditService{
		repo: repo,
	}
}

func (s *auditService) LogAuditEvent(ctx context.Context, dto models.CreateAuditLogDTO) error {
	log := &models.AuditLog{
		UserID:              dto.UserID,
		Action:              dto.Action,
		Resource:            dto.Resource,
		EntityID:            dto.EntityID,
		EntityType:          dto.EntityType,
		RequestPath:         dto.RequestPath,
		Method:              dto.Method,
		StatusCode:          dto.StatusCode,
		IPAddress:           dto.IPAddress,
		UserAgent:           dto.UserAgent,
		RequestBody:         dto.RequestBody,
		AttemptedIdentifier: dto.AttemptedIdentifier,
		Timestamp:           time.Now().UTC(),
	}
	return s.repo.Create(ctx, log)
}

func (s *auditService) GetAuditLogs(ctx context.Context, filter models.AuditLogFilter) ([]*models.AuditLog, error) {
	return s.repo.FindByFilter(ctx, filter)
}
