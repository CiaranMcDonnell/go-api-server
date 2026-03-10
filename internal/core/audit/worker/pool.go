package worker

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/ciaranmcdonnell/go-api-server/internal/core/audit/domain/interfaces"
	"github.com/ciaranmcdonnell/go-api-server/internal/core/audit/domain/models"
	"github.com/ciaranmcdonnell/go-api-server/internal/metrics"
)

const (
	batchSize     = 50
	flushInterval = 100 * time.Millisecond
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
			p.runBatchLoop(ctx, id)
			slog.Debug("Audit worker stopped", "worker", id)
		}(i)
	}
	metrics.AuditQueueCapacity.Set(float64(cap(p.queue)))
	slog.Info("Audit worker pool started", "workers", p.workers, "queue_size", cap(p.queue), "batch_size", batchSize)
}

func (p *Pool) runBatchLoop(ctx context.Context, workerID int) {
	batch := make([]models.CreateAuditLogDTO, 0, batchSize)
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		if err := p.service.LogAuditEventBatch(ctx, batch); err != nil {
			slog.Error("Audit batch insert failed", "worker", workerID, "count", len(batch), "error", err)
		}
		batch = batch[:0]
		metrics.AuditQueueLength.Set(float64(len(p.queue)))
	}

	for {
		select {
		case dto, ok := <-p.queue:
			if !ok {
				flush()
				return
			}
			batch = append(batch, dto)
			if len(batch) >= batchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}

func (p *Pool) Submit(dto models.CreateAuditLogDTO) {
	select {
	case p.queue <- dto:
		metrics.AuditQueueLength.Set(float64(len(p.queue)))
	default:
		slog.Warn("Audit queue full, dropping event", "path", dto.RequestPath)
	}
}

func (p *Pool) Stop() {
	close(p.queue)
	p.wg.Wait()
	slog.Info("Audit worker pool stopped")
}
