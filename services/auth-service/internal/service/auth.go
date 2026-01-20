package service

import (
	"errors"
	"fmt"
	"saas-subscription-platform/services/auth-service/internal/client"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = client.ErrUserExists
)

type AuthService struct {
	jwtSecret  []byte
	userClient *client.UserClient
}

func NewAuthService(secret string, userClient *client.UserClient) *AuthService {
	return &AuthService{
		jwtSecret:  []byte(secret),
		userClient: userClient,
	}
}

func (s *AuthService) Register(email, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	headers := map[string]string{
		"X-Internal-User-ID": "auth-service",
	}

	_, err = s.userClient.CreateUser(email, string(hash), headers)
	if err == client.ErrUserExists {
		return err
	}
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (s *AuthService) Login(email, password string) (string, error) {
	headers := map[string]string{
		"X-Internal-User-ID": "auth-service",
	}

	user, err := s.userClient.GetUserByEmail(email, headers)
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
