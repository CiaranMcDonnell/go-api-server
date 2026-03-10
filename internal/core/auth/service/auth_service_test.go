package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ciaranmcdonnell/go-api-server/internal/core/user/mocks"
	"github.com/ciaranmcdonnell/go-api-server/models"
	"github.com/ciaranmcdonnell/go-api-server/pkg/apperrors"
	"github.com/ciaranmcdonnell/go-api-server/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func newTestConfig() *utils.Config {
	return &utils.Config{
		JWTSecret:          "test-secret-key-that-is-at-least-32-chars-long",
		JWTExpirationHours: 8,
	}
}

func newTestUser() models.User {
	hash, _ := utils.HashPassword("CorrectPassword123")
	return models.User{
		ID:             1,
		Name:           "Test User",
		Email:          "test@example.com",
		HashedPassword: hash,
		Role:           "user",
		CreatedAt:      time.Now(),
	}
}

func TestAuthenticateUser_Success(t *testing.T) {
	userQueries := new(mocks.MockUserQueries)
	svc := NewAuthService(newTestConfig(), userQueries)
	ctx := context.Background()

	testUser := newTestUser()
	userQueries.On("GetUserByEmail", ctx, "test@example.com").Return(testUser, nil)

	user, err := svc.AuthenticateUser(ctx, "test@example.com", "CorrectPassword123")

	assert.NoError(t, err)
	assert.Equal(t, int64(1), user.ID)
	assert.Equal(t, "test@example.com", user.Email)
	userQueries.AssertExpectations(t)
}

func TestAuthenticateUser_WrongPassword(t *testing.T) {
	userQueries := new(mocks.MockUserQueries)
	svc := NewAuthService(newTestConfig(), userQueries)
	ctx := context.Background()

	testUser := newTestUser()
	userQueries.On("GetUserByEmail", ctx, "test@example.com").Return(testUser, nil)

	user, err := svc.AuthenticateUser(ctx, "test@example.com", "WrongPassword")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.True(t, errors.Is(err, apperrors.ErrInvalidCredentials))
}

func TestAuthenticateUser_UserNotFound(t *testing.T) {
	userQueries := new(mocks.MockUserQueries)
	svc := NewAuthService(newTestConfig(), userQueries)
	ctx := context.Background()

	userQueries.On("GetUserByEmail", ctx, "nobody@example.com").Return(models.User{}, fmt.Errorf("no rows"))

	user, err := svc.AuthenticateUser(ctx, "nobody@example.com", "password")

	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestGenerateAndValidateToken(t *testing.T) {
	svc := NewAuthService(newTestConfig(), nil)

	user := &models.User{ID: 42}
	token, err := svc.GenerateAuthToken(user)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := svc.ValidateToken(token)

	assert.NoError(t, err)
	assert.Equal(t, "42", claims.UserID)
}

func TestValidateToken_InvalidToken(t *testing.T) {
	svc := NewAuthService(newTestConfig(), nil)

	claims, err := svc.ValidateToken("garbage.token.value")

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestValidateToken_EmptyToken(t *testing.T) {
	svc := NewAuthService(newTestConfig(), nil)

	claims, err := svc.ValidateToken("")

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestValidateToken_WrongSecret(t *testing.T) {
	svc1 := NewAuthService(&utils.Config{JWTSecret: "secret-one-that-is-long-enough-32", JWTExpirationHours: 8}, nil)
	svc2 := NewAuthService(&utils.Config{JWTSecret: "secret-two-that-is-long-enough-32", JWTExpirationHours: 8}, nil)

	token, _ := svc1.GenerateAuthToken(&models.User{ID: 1})
	claims, err := svc2.ValidateToken(token)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestGetCurrentUser_Success(t *testing.T) {
	userQueries := new(mocks.MockUserQueries)
	svc := NewAuthService(newTestConfig(), userQueries)
	ctx := context.Background()

	userCache.Delete(int64(1))

	testUser := newTestUser()
	userQueries.On("GetUserByID", ctx, int64(1)).Return(testUser, nil)

	user, err := svc.GetCurrentUser(ctx, 1)

	assert.NoError(t, err)
	assert.Equal(t, "Test User", user.Name)
	userQueries.AssertExpectations(t)
}

func TestGetCurrentUser_Caches(t *testing.T) {
	userQueries := new(mocks.MockUserQueries)
	svc := NewAuthService(newTestConfig(), userQueries)
	ctx := context.Background()

	userCache.Delete(int64(2))

	testUser := models.User{
		ID:    2,
		Name:  "Cached User",
		Email: "cached@example.com",
	}
	userQueries.On("GetUserByID", ctx, int64(2)).Return(testUser, nil).Once()

	user1, err1 := svc.GetCurrentUser(ctx, 2)
	user2, err2 := svc.GetCurrentUser(ctx, 2)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Equal(t, user1.Name, user2.Name)
	userQueries.AssertNumberOfCalls(t, "GetUserByID", 1)
}

func TestGetCurrentUser_NotFound(t *testing.T) {
	userQueries := new(mocks.MockUserQueries)
	svc := NewAuthService(newTestConfig(), userQueries)
	ctx := context.Background()

	userCache.Delete(int64(999))

	userQueries.On("GetUserByID", ctx, int64(999)).Return(models.User{}, fmt.Errorf("no rows"))

	user, err := svc.GetCurrentUser(ctx, 999)

	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestExtractTokenFromRequest_BearerHeader(t *testing.T) {
	token := ExtractTokenFromRequest("Bearer my-jwt-token", "")
	assert.Equal(t, "my-jwt-token", token)
}

func TestExtractTokenFromRequest_CookieFallback(t *testing.T) {
	token := ExtractTokenFromRequest("", "cookie-token")
	assert.Equal(t, "cookie-token", token)
}

func TestExtractTokenFromRequest_HeaderPriority(t *testing.T) {
	token := ExtractTokenFromRequest("Bearer header-token", "cookie-token")
	assert.Equal(t, "header-token", token)
}

func TestExtractTokenFromRequest_NoBearer(t *testing.T) {
	token := ExtractTokenFromRequest("Basic abc123", "cookie-token")
	assert.Equal(t, "cookie-token", token)
}
