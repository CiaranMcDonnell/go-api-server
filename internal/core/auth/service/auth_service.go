package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	userrepo "github.com/ciaranmcdonnell/go-api-server/internal/core/user/repository"
	"github.com/ciaranmcdonnell/go-api-server/internal/metrics"
	"github.com/ciaranmcdonnell/go-api-server/models"
	"github.com/ciaranmcdonnell/go-api-server/pkg/apperrors"
	"github.com/ciaranmcdonnell/go-api-server/pkg/cache"
	"github.com/ciaranmcdonnell/go-api-server/pkg/utils"
	"github.com/golang-jwt/jwt/v5"
)

var userCache = cache.New[int64, *models.User](30*time.Second, 10000)

const userCacheTTL = 5 * time.Minute

type AuthServiceInterface interface {
	AuthenticateUser(ctx context.Context, email, password string) (*models.User, error)
	GenerateAuthToken(user *models.User) (string, error)
	RefreshToken(ctx context.Context, tokenString string) (string, error)
	ValidateToken(tokenString string) (*models.Claims, error)
	GetCurrentUser(ctx context.Context, userID int64) (*models.User, error)
}

type AuthService struct {
	config      *utils.Config
	userQueries userrepo.UserQueriesInterface
}

func NewAuthService(config *utils.Config, userQueries userrepo.UserQueriesInterface) *AuthService {
	return &AuthService{
		config:      config,
		userQueries: userQueries,
	}
}

func (s *AuthService) ValidateToken(tokenString string) (*models.Claims, error) {
	if tokenString == "" {
		return nil, apperrors.ErrUnauthorized
	}

	token, err := jwt.ParseWithClaims(tokenString, &models.Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v: %w", token.Header["alg"], apperrors.ErrUnauthorized)
		}
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("%w: %w", apperrors.ErrUnauthorized, err)
	}

	claims, ok := token.Claims.(*models.Claims)
	if !ok || !token.Valid {
		return nil, apperrors.ErrUnauthorized
	}

	return claims, nil
}

func (s *AuthService) AuthenticateUser(ctx context.Context, email, password string) (*models.User, error) {
	user, err := s.userQueries.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, apperrors.ErrInvalidCredentials
	}

	if !utils.CheckPassword(password, user.HashedPassword) {
		return nil, apperrors.ErrInvalidCredentials
	}

	return &user, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, tokenString string) (string, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	userID, err := strconv.ParseInt(claims.UserID, 10, 64)
	if err != nil {
		return "", fmt.Errorf("invalid user ID in token: %w", err)
	}

	user, err := s.GetCurrentUser(ctx, userID)
	if err != nil {
		return "", err
	}

	return s.GenerateAuthToken(user)
}

func (s *AuthService) GenerateAuthToken(user *models.User) (string, error) {
	expHours := s.config.JWTExpirationHours
	if expHours <= 0 {
		expHours = 8
	}

	claims := models.Claims{
		UserID: strconv.FormatInt(user.ID, 10),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expHours) * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWTSecret))
}

func (s *AuthService) GetCurrentUser(ctx context.Context, userID int64) (*models.User, error) {
	if user, ok := userCache.Get(userID); ok {
		metrics.CacheHitsTotal.WithLabelValues("user").Inc()
		return user, nil
	}
	metrics.CacheMissesTotal.WithLabelValues("user").Inc()

	user, err := s.userQueries.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting user by ID: %w", err)
	}

	result := &user
	userCache.Set(userID, result, userCacheTTL)
	return result, nil
}

func ExtractTokenFromRequest(authHeader, cookieValue string) string {
	if authHeader != "" {
		token := strings.TrimPrefix(authHeader, utils.BearerPrefix)
		if token != authHeader {
			return token
		}
	}
	return cookieValue
}
