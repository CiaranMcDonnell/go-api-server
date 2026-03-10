package mocks

import (
	"context"

	"github.com/ciaranmcdonnell/go-api-server/models"
	"github.com/stretchr/testify/mock"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) AuthenticateUser(ctx context.Context, email, password string) (*models.User, error) {
	args := m.Called(ctx, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockAuthService) GenerateAuthToken(user *models.User) (string, error) {
	args := m.Called(user)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) RefreshToken(ctx context.Context, tokenString string) (string, error) {
	args := m.Called(ctx, tokenString)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) ValidateToken(tokenString string) (*models.Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Claims), args.Error(1)
}

func (m *MockAuthService) GetCurrentUser(ctx context.Context, userID int64) (*models.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}
