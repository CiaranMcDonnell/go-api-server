package service

import (
	auditsvc "github.com/ciaranmcdonnell/go-api-server/internal/core/audit/application/service"
	auditinterfaces "github.com/ciaranmcdonnell/go-api-server/internal/core/audit/domain/interfaces"
	authsvc "github.com/ciaranmcdonnell/go-api-server/internal/core/auth/service"
	serviceReg "github.com/ciaranmcdonnell/go-api-server/internal/core/common/repository"
	itemsvc "github.com/ciaranmcdonnell/go-api-server/internal/core/items/application/service"
	iteminterfaces "github.com/ciaranmcdonnell/go-api-server/internal/core/items/domain/interfaces"
	usersvc "github.com/ciaranmcdonnell/go-api-server/internal/core/user/service"
	"github.com/ciaranmcdonnell/go-api-server/internal/database"
	"github.com/ciaranmcdonnell/go-api-server/pkg/utils"
)

type ServicesInterface interface {
	GetAuthService() authsvc.AuthServiceInterface
	GetUserService() usersvc.UserServiceInterface
	GetAuditService() auditinterfaces.AuditService
	GetItemService() iteminterfaces.ItemService
	GetQueriesManager() serviceReg.QueriesInterface
	GetConfig() *utils.Config
}

type Services struct {
	Auth           authsvc.AuthServiceInterface
	User           usersvc.UserServiceInterface
	Audit          auditinterfaces.AuditService
	Items          iteminterfaces.ItemService
	queriesManager serviceReg.QueriesInterface
	config         *utils.Config
}

func NewServices(config *utils.Config, queriesManager serviceReg.QueriesInterface, txMgr database.TxManager) ServicesInterface {
	auditService := auditsvc.NewAuditService(queriesManager.GetAuditQueries())

	return &Services{
		Auth: authsvc.NewAuthService(
			config,
			queriesManager.GetUserQueries(),
		),
		User: usersvc.NewUserService(
			config,
			queriesManager.GetUserQueries(),
		),
		Audit:          auditService,
		Items:          itemsvc.NewItemService(queriesManager.GetItemQueries(), txMgr),
		queriesManager: queriesManager,
		config:         config,
	}
}

func (s *Services) GetAuthService() authsvc.AuthServiceInterface {
	return s.Auth
}

func (s *Services) GetUserService() usersvc.UserServiceInterface {
	return s.User
}

func (s *Services) GetAuditService() auditinterfaces.AuditService {
	return s.Audit
}

func (s *Services) GetQueriesManager() serviceReg.QueriesInterface {
	return s.queriesManager
}

func (s *Services) GetItemService() iteminterfaces.ItemService {
	return s.Items
}

func (s *Services) GetConfig() *utils.Config {
	return s.config
}
