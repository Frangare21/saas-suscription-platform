package service

import (
	"context"
	"testing"

	"saas-subscription-platform/services/auth-service/internal/client"
	"saas-subscription-platform/services/auth-service/internal/service/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthService_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockUser := mocks.NewMockUserClient(ctrl)
	svc := NewAuthService("secret", mockUser)

	mockUser.EXPECT().CreateUserWithContext(gomock.Any(), "alice@example.com", gomock.Any(), gomock.Any()).Return(client.CreateUserResponse{ID: "u-1"}, nil)
	err := svc.Register("alice@example.com", "pass")
	require.NoError(t, err)

	mockUser.EXPECT().CreateUserWithContext(gomock.Any(), "dup@example.com", gomock.Any(), gomock.Any()).Return(client.CreateUserResponse{}, client.ErrUserExists)
	err = svc.Register("dup@example.com", "pass")
	require.ErrorIs(t, err, ErrUserExists)
}

func TestAuthService_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockUser := mocks.NewMockUserClient(ctrl)
	svc := NewAuthService("secret", mockUser)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("pass"), bcrypt.DefaultCost)

	mockUser.EXPECT().GetUserByEmailWithContext(gomock.Any(), "alice@example.com", gomock.Any()).Return(client.GetUserByEmailResponse{
		ID:       "u-1",
		Email:    "alice@example.com",
		Password: string(hashed),
	}, nil)

	token, err := svc.Login("alice@example.com", "pass")
	require.NoError(t, err)
	require.NotEmpty(t, token)

	mockUser.EXPECT().GetUserByEmailWithContext(gomock.Any(), "missing@example.com", gomock.Any()).Return(client.GetUserByEmailResponse{}, client.ErrUserNotFound)
	_, err = svc.Login("missing@example.com", "pass")
	require.ErrorIs(t, err, ErrInvalidCredentials)

	mockUser.EXPECT().GetUserByEmailWithContext(gomock.Any(), "alice@example.com", gomock.Any()).Return(client.GetUserByEmailResponse{
		ID:       "u-1",
		Email:    "alice@example.com",
		Password: string(hashed),
	}, nil)
	_, err = svc.LoginWithContext(context.Background(), "alice@example.com", "wrong")
	require.ErrorIs(t, err, ErrInvalidCredentials)
}
