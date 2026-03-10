package services

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/ciaranmcdonnell/go-api-server/internal/core/user/mocks"
	"github.com/ciaranmcdonnell/go-api-server/models"
	"github.com/ciaranmcdonnell/go-api-server/pkg/apperrors"
	"github.com/ciaranmcdonnell/go-api-server/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newTestConfig() *utils.Config {
	return &utils.Config{
		JWTSecret:          "test-secret-key-at-least-32-chars",
		JWTExpirationHours: 8,
	}
}

func TestRegisterUser_Success(t *testing.T) {
	userQueries := new(mocks.MockUserQueries)
	svc := NewUserService(newTestConfig(), userQueries)
	ctx := context.Background()

	req := &models.CreateUserRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "SecurePass123",
	}

	userQueries.On("CreateUser", ctx, mock.MatchedBy(func(user models.User) bool {
		return user.Name == "Test User" &&
			user.Email == "test@example.com" &&
			user.HashedPassword != "" &&
			user.HashedPassword != "SecurePass123"
	})).Return(int64(1), nil)

	id, err := svc.RegisterUser(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), id)
	userQueries.AssertExpectations(t)
}

func TestRegisterUser_DBError(t *testing.T) {
	userQueries := new(mocks.MockUserQueries)
	svc := NewUserService(newTestConfig(), userQueries)
	ctx := context.Background()

	req := &models.CreateUserRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "SecurePass123",
	}

	userQueries.On("CreateUser", ctx, mock.Anything).Return(int64(0), fmt.Errorf("duplicate key value violates unique constraint"))

	id, err := svc.RegisterUser(ctx, req)

	assert.Error(t, err)
	assert.Equal(t, int64(0), id)
	assert.True(t, errors.Is(err, apperrors.ErrConflict))
	userQueries.AssertExpectations(t)
}

func TestRegisterUser_PasswordIsHashed(t *testing.T) {
	userQueries := new(mocks.MockUserQueries)
	svc := NewUserService(newTestConfig(), userQueries)
	ctx := context.Background()

	req := &models.CreateUserRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "SecurePass123",
	}

	var capturedUser models.User
	userQueries.On("CreateUser", ctx, mock.Anything).Run(func(args mock.Arguments) {
		capturedUser = args.Get(1).(models.User)
	}).Return(int64(1), nil)

	_, _ = svc.RegisterUser(ctx, req)

	assert.Contains(t, capturedUser.HashedPassword, "$argon2id$")
	assert.True(t, utils.CheckPassword("SecurePass123", capturedUser.HashedPassword))
	assert.False(t, utils.CheckPassword("WrongPassword", capturedUser.HashedPassword))
}
