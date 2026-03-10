package repository

import (
	auditDomainInterfaces "github.com/ciaranmcdonnell/go-api-server/internal/core/audit/domain/interfaces"
	auditInfraRepo "github.com/ciaranmcdonnell/go-api-server/internal/core/audit/infrastructure/repository"
	itemInterfaces "github.com/ciaranmcdonnell/go-api-server/internal/core/items/domain/interfaces"
	itemInfraRepo "github.com/ciaranmcdonnell/go-api-server/internal/core/items/infrastructure/repository"
	userrepo "github.com/ciaranmcdonnell/go-api-server/internal/core/user/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

type QueriesInterface interface {
	GetUserQueries() userrepo.UserQueriesInterface
	GetAuditQueries() auditDomainInterfaces.AuditRepository
	GetItemQueries() itemInterfaces.ItemRepository
}

type Queries struct {
	Users userrepo.UserQueriesInterface
	Audit auditDomainInterfaces.AuditRepository
	Items itemInterfaces.ItemRepository
}

func NewQueries(db *pgxpool.Pool) *Queries {
	return &Queries{
		Users: userrepo.NewUserQueries(db),
		Audit: auditInfraRepo.NewAuditRepository(db),
		Items: itemInfraRepo.NewItemRepository(db),
	}
}

func (q *Queries) GetUserQueries() userrepo.UserQueriesInterface {
	return q.Users
}

func (q *Queries) GetAuditQueries() auditDomainInterfaces.AuditRepository {
	return q.Audit
}

func (q *Queries) GetItemQueries() itemInterfaces.ItemRepository {
	return q.Items
}
