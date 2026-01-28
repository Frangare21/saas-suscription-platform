package service

import (
	"context"
	"errors"
	"fmt"
	"saas-subscription-platform/services/auth-service/internal/client"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// UserClient define las operaciones del cliente de usuarios que la capa de servicio necesita.
type UserClient interface {
	CreateUserWithContext(ctx context.Context, email, password string, headers map[string]string) (client.CreateUserResponse, error)
	GetUserByEmailWithContext(ctx context.Context, email string, headers map[string]string) (client.GetUserByEmailResponse, error)
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = client.ErrUserExists
)

type AuthService struct {
	jwtSecret  []byte
	userClient UserClient
}

func NewAuthService(secret string, userClient UserClient) *AuthService {
	return &AuthService{
		jwtSecret:  []byte(secret),
		userClient: userClient,
	}
}

// Register mantiene compatibilidad, pero usa context.Background().
func (s *AuthService) Register(email, password string) error {
	return s.RegisterWithContext(context.Background(), email, password)
}

func (s *AuthService) RegisterWithContext(ctx context.Context, email, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	headers := map[string]string{
		"X-Internal-User-ID": "auth-service",
	}

	_, err = s.userClient.CreateUserWithContext(ctx, email, string(hash), headers)
	if err == client.ErrUserExists {
		return err
	}
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// Login mantiene compatibilidad, pero usa context.Background().
func (s *AuthService) Login(email, password string) (string, error) {
	return s.LoginWithContext(context.Background(), email, password)
}

func (s *AuthService) LoginWithContext(ctx context.Context, email, password string) (string, error) {
	headers := map[string]string{
		"X-Internal-User-ID": "auth-service",
	}

	user, err := s.userClient.GetUserByEmailWithContext(ctx, email, headers)
	if err != nil {
		return "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(15 * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(s.jwtSecret)
}
