package service

import (
	auditsvc "github.com/ciaranmcdonnell/go-api-server/internal/core/audit/application/service"
	auditinterfaces "github.com/ciaranmcdonnell/go-api-server/internal/core/audit/domain/interfaces"
	authsvc "github.com/ciaranmcdonnell/go-api-server/internal/core/auth/service"
	serviceReg "github.com/ciaranmcdonnell/go-api-server/internal/core/common/repository"
	usersvc "github.com/ciaranmcdonnell/go-api-server/internal/core/user/service"
	"github.com/ciaranmcdonnell/go-api-server/pkg/utils"
)

type ServicesInterface interface {
	GetAuthService() authsvc.AuthServiceInterface
	GetUserService() usersvc.UserServiceInterface
	GetAuditService() auditinterfaces.AuditService
	GetQueriesManager() serviceReg.QueriesInterface
	GetConfig() *utils.Config
}

type Services struct {
	Auth           authsvc.AuthServiceInterface
	User           usersvc.UserServiceInterface
	Audit          auditinterfaces.AuditService
	queriesManager serviceReg.QueriesInterface
	config         *utils.Config
}

func NewServices(config *utils.Config, queriesManager serviceReg.QueriesInterface) ServicesInterface {
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

func (s *Services) GetConfig() *utils.Config {
	return s.config
}
