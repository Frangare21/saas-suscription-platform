package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"saas-subscription-platform/services/user-service/internal/model"
	"saas-subscription-platform/services/user-service/internal/repository"
	"saas-subscription-platform/services/user-service/internal/service"

	"github.com/stretchr/testify/require"
)

type stubUserStore struct {
	createFn       func(email, password string) (model.User, error)
	getByEmailFn   func(email string) (model.User, error)
	getByIDFn      func(userID string) (model.User, error)
	updateFieldsFn func(userID string, email, password *string) error
	deleteFn       func(userID string) error
}

func (s stubUserStore) Create(email, password string) (model.User, error) {
	return s.createFn(email, password)
}

func (s stubUserStore) GetByEmail(email string) (model.User, error) {
	return s.getByEmailFn(email)
}

func (s stubUserStore) GetByID(userID string) (model.User, error) {
	return s.getByIDFn(userID)
}

func (s stubUserStore) UpdateFields(userID string, email, password *string) error {
	return s.updateFieldsFn(userID, email, password)
}

func (s stubUserStore) Delete(userID string) error {
	return s.deleteFn(userID)
}

func newHandlerWithStore(store service.UserStore) *UserHandler {
	svc := service.NewUserService(store)
	return NewUserHandler(svc)
}

func TestCreateUserHandler(t *testing.T) {
	createdAt := time.Date(2024, 12, 1, 10, 0, 0, 0, time.UTC)
	h := newHandlerWithStore(stubUserStore{
		createFn: func(email, password string) (model.User, error) {
			return model.User{ID: "u-1", Email: email, CreatedAt: createdAt}, nil
		},
	})

	body := bytes.NewBufferString(`{"email":"alice@example.com","password":"secret"}`)
	req := httptest.NewRequest(http.MethodPost, "/users", body)
	rr := httptest.NewRecorder()

	h.CreateUser(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
	var resp UserResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	require.Equal(t, "u-1", resp.ID)
	require.Equal(t, "alice@example.com", resp.Email)
	require.Equal(t, "2024-12-01T10:00:00Z", resp.CreatedAt)
}

func TestCreateUserHandler_Conflict(t *testing.T) {
	h := newHandlerWithStore(stubUserStore{
		createFn: func(email, password string) (model.User, error) {
			return model.User{}, repository.ErrUserExists
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(`{"email":"dup@example.com","password":"x"}`))
	rr := httptest.NewRecorder()

	h.CreateUser(rr, req)

	require.Equal(t, http.StatusConflict, rr.Code)
}

func TestGetUserByEmailHandler(t *testing.T) {
	createdAt := time.Date(2024, 12, 2, 9, 0, 0, 0, time.UTC)
	h := newHandlerWithStore(stubUserStore{
		getByEmailFn: func(email string) (model.User, error) {
			return model.User{ID: "u-1", Email: email, Password: "hash", CreatedAt: createdAt}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/users/email/alice@example.com", nil)
	req.SetPathValue("email", "alice@example.com")
	rr := httptest.NewRecorder()

	h.GetUserByEmail(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var resp UserResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	require.Equal(t, "alice@example.com", resp.Email)
	require.Equal(t, "hash", resp.Password)
	require.Equal(t, "2024-12-02T09:00:00Z", resp.CreatedAt)
}

func TestGetUserByEmailHandler_NotFound(t *testing.T) {
	h := newHandlerWithStore(stubUserStore{
		getByEmailFn: func(email string) (model.User, error) {
			return model.User{}, repository.ErrUserNotFound
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/users/email/missing@example.com", nil)
	req.SetPathValue("email", "missing@example.com")
	rr := httptest.NewRecorder()

	h.GetUserByEmail(rr, req)

	require.Equal(t, http.StatusNotFound, rr.Code)
}

func TestUpdateUserHandler(t *testing.T) {
	var capturedEmail *string
	h := newHandlerWithStore(stubUserStore{
		updateFieldsFn: func(userID string, email, password *string) error {
			capturedEmail = email
			return nil
		},
	})

	req := httptest.NewRequest(http.MethodPatch, "/users/u-1", bytes.NewBufferString(`{"email":"new@example.com"}`))
	req.SetPathValue("id", "u-1")
	rr := httptest.NewRecorder()

	h.UpdateUser(rr, req)

	require.Equal(t, http.StatusNoContent, rr.Code)
	require.NotNil(t, capturedEmail)
	require.Equal(t, "new@example.com", *capturedEmail)
}

func TestUpdateUserHandler_Conflict(t *testing.T) {
	h := newHandlerWithStore(stubUserStore{
		updateFieldsFn: func(userID string, email, password *string) error {
			return repository.ErrUserExists
		},
	})

	req := httptest.NewRequest(http.MethodPatch, "/users/u-1", bytes.NewBufferString(`{"email":"dup@example.com"}`))
	req.SetPathValue("id", "u-1")
	rr := httptest.NewRecorder()

	h.UpdateUser(rr, req)

	require.Equal(t, http.StatusConflict, rr.Code)
}

func TestDeleteUserHandler(t *testing.T) {
	h := newHandlerWithStore(stubUserStore{
		deleteFn: func(userID string) error { return nil },
	})

	req := httptest.NewRequest(http.MethodDelete, "/users/u-1", nil)
	req.SetPathValue("id", "u-1")
	rr := httptest.NewRecorder()

	h.DeleteUser(rr, req)

	require.Equal(t, http.StatusNoContent, rr.Code)
}

func TestDeleteUserHandler_NotFound(t *testing.T) {
	h := newHandlerWithStore(stubUserStore{
		deleteFn: func(userID string) error { return repository.ErrUserNotFound },
	})

	req := httptest.NewRequest(http.MethodDelete, "/users/u-2", nil)
	req.SetPathValue("id", "u-2")
	rr := httptest.NewRecorder()

	h.DeleteUser(rr, req)

	require.Equal(t, http.StatusNotFound, rr.Code)
}

func TestUpdateUserHandler_NoFields(t *testing.T) {
	h := newHandlerWithStore(stubUserStore{})

	req := httptest.NewRequest(http.MethodPatch, "/users/u-1", bytes.NewBufferString(`{}`))
	req.SetPathValue("id", "u-1")
	rr := httptest.NewRecorder()

	h.UpdateUser(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}
