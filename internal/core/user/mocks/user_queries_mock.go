package mocks

import (
	"context"

	"github.com/ciaranmcdonnell/go-api-server/models"
	"github.com/stretchr/testify/mock"
)

type MockUserQueries struct {
	mock.Mock
}

func (m *MockUserQueries) CreateUser(ctx context.Context, user models.User) (int64, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserQueries) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *MockUserQueries) GetUserByID(ctx context.Context, id int64) (models.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(models.User), args.Error(1)
}
