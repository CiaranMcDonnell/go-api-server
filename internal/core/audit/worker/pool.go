package worker

import (
	"context"
	"log/slog"
	"sync"

	"github.com/ciaranmcdonnell/go-api-server/internal/core/audit/domain/interfaces"
	"github.com/ciaranmcdonnell/go-api-server/internal/core/audit/domain/models"
)

type Pool struct {
	workers int
	queue   chan models.CreateAuditLogDTO
	service interfaces.AuditService
	wg      sync.WaitGroup
}

func NewPool(workers, queueSize int, service interfaces.AuditService) *Pool {
	return &Pool{
		workers: workers,
		queue:   make(chan models.CreateAuditLogDTO, queueSize),
		service: service,
	}
}

func (p *Pool) Start(ctx context.Context) {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go func(id int) {
			defer p.wg.Done()
			slog.Debug("Audit worker started", "worker", id)
			for dto := range p.queue {
				if err := p.service.LogAuditEvent(ctx, dto); err != nil {
					slog.Error("Audit worker failed to log event", "worker", id, "error", err)
				}
			}
			slog.Debug("Audit worker stopped", "worker", id)
		}(i)
	}
	slog.Info("Audit worker pool started", "workers", p.workers, "queue_size", cap(p.queue))
}

func (p *Pool) Submit(dto models.CreateAuditLogDTO) {
	select {
	case p.queue <- dto:
	default:
		slog.Warn("Audit queue full, dropping event", "path", dto.RequestPath)
	}
}

func (p *Pool) Stop() {
	close(p.queue)
	p.wg.Wait()
	slog.Info("Audit worker pool stopped")
}
