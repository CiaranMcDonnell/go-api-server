package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	userrepo "github.com/ciaranmcdonnell/go-api-server/internal/core/user/repository"
	"github.com/ciaranmcdonnell/go-api-server/models"
	"github.com/ciaranmcdonnell/go-api-server/pkg/utils"
	"github.com/golang-jwt/jwt/v5"
)

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
		return nil, fmt.Errorf("authentication token required")
	}

	token, err := jwt.ParseWithClaims(tokenString, &models.Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*models.Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

func (s *AuthService) AuthenticateUser(ctx context.Context, email, password string) (*models.User, error) {
	user, err := s.userQueries.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if !utils.CheckPassword(password, user.Password) {
		return nil, fmt.Errorf("invalid credentials")
	}

	user.Password = ""

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
	user, err := s.userQueries.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting user by ID: %w", err)
	}
	return &user, nil
}

// ExtractTokenFromRequest extracts the JWT token string from Authorization header or cookie value.
func ExtractTokenFromRequest(authHeader, cookieValue string) string {
	if authHeader != "" {
		token := strings.TrimPrefix(authHeader, utils.BearerPrefix)
		if token != authHeader {
			return token
		}
	}
	return cookieValue
}
