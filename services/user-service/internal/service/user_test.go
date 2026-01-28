package service

import (
	"testing"
	"time"

	"saas-subscription-platform/services/user-service/internal/model"
	"saas-subscription-platform/services/user-service/internal/repository"
	"saas-subscription-platform/services/user-service/internal/service/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestUserService_DelegatesToStore(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := mocks.NewMockUserStore(ctrl)
	svc := NewUserService(store)

	createdAt := time.Now()
	expectedUser := model.User{ID: "u-1", Email: "alice@example.com", Password: "hash", CreatedAt: createdAt}

	store.EXPECT().Create("alice@example.com", "hash").Return(expectedUser, nil)
	user, err := svc.CreateUser("alice@example.com", "hash")
	require.NoError(t, err)
	require.Equal(t, expectedUser, user)

	store.EXPECT().GetByEmail("alice@example.com").Return(expectedUser, nil)
	user, err = svc.GetUserByEmail("alice@example.com")
	require.NoError(t, err)
	require.Equal(t, expectedUser, user)

	store.EXPECT().GetByID("u-1").Return(expectedUser, nil)
	user, err = svc.GetUserByID("u-1")
	require.NoError(t, err)
	require.Equal(t, expectedUser, user)

	newEmail := "new@example.com"
	store.EXPECT().UpdateFields("u-1", &newEmail, (*string)(nil)).Return(repository.ErrUserExists)
	err = svc.UpdateUser("u-1", &newEmail, nil)
	require.ErrorIs(t, err, repository.ErrUserExists)

	store.EXPECT().Delete("u-1").Return(nil)
	require.NoError(t, svc.DeleteUser("u-1"))
}
