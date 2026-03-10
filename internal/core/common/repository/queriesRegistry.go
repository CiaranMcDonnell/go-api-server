package repository

import (
	auditDomainInterfaces "github.com/ciaranmcdonnell/go-api-server/internal/core/audit/domain/interfaces"
	auditInfraRepo "github.com/ciaranmcdonnell/go-api-server/internal/core/audit/infrastructure/repository"
	userrepo "github.com/ciaranmcdonnell/go-api-server/internal/core/user/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

type QueriesInterface interface {
	GetUserQueries() userrepo.UserQueriesInterface
	GetAuditQueries() auditDomainInterfaces.AuditRepository
}

type Queries struct {
	Users userrepo.UserQueriesInterface
	Audit auditDomainInterfaces.AuditRepository
}

func NewQueries(db *pgxpool.Pool) *Queries {
	return &Queries{
		Users: userrepo.NewUserQueries(db),
		Audit: auditInfraRepo.NewAuditRepository(db),
	}
}

func (q *Queries) GetUserQueries() userrepo.UserQueriesInterface {
	return q.Users
}

func (q *Queries) GetAuditQueries() auditDomainInterfaces.AuditRepository {
	return q.Audit
}
