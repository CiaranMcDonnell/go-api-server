package services

import (
	"context"
	"fmt"

	userrepo "github.com/ciaranmcdonnell/go-api-server/internal/core/user/repository"
	"github.com/ciaranmcdonnell/go-api-server/models"
	"github.com/ciaranmcdonnell/go-api-server/pkg/utils"
)

type UserServiceInterface interface {
	RegisterUser(ctx context.Context, req *models.CreateUserRequest) (int64, error)
}

type UserService struct {
	config      *utils.Config
	userQueries userrepo.UserQueriesInterface
}

func NewUserService(config *utils.Config, userQueries userrepo.UserQueriesInterface) *UserService {
	return &UserService{
		config:      config,
		userQueries: userQueries,
	}
}

func (s *UserService) RegisterUser(ctx context.Context, req *models.CreateUserRequest) (int64, error) {
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return 0, fmt.Errorf("error hashing password: %w", err)
	}

	user := models.User{
		Name:           req.Name,
		Email:          req.Email,
		HashedPassword: hashedPassword,
	}

	id, err := s.userQueries.CreateUser(ctx, user)
	if err != nil {
		return 0, err
	}

	return id, nil
}
