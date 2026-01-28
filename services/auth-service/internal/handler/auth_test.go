package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"saas-subscription-platform/services/auth-service/internal/client"
	"saas-subscription-platform/services/auth-service/internal/middleware"
	"saas-subscription-platform/services/auth-service/internal/service"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type stubUserClient struct {
	createFn    func(ctx context.Context, email, password string, headers map[string]string) (client.CreateUserResponse, error)
	getByEmail  func(ctx context.Context, email string, headers map[string]string) (client.GetUserByEmailResponse, error)
	getByIDFunc func(ctx context.Context, userID string, headers map[string]string) (client.GetUserByEmailResponse, error)
}

func (s stubUserClient) CreateUserWithContext(ctx context.Context, email, password string, headers map[string]string) (client.CreateUserResponse, error) {
	return s.createFn(ctx, email, password, headers)
}

func (s stubUserClient) GetUserByEmailWithContext(ctx context.Context, email string, headers map[string]string) (client.GetUserByEmailResponse, error) {
	return s.getByEmail(ctx, email, headers)
}

func (s stubUserClient) GetUserByIDWithContext(ctx context.Context, userID string, headers map[string]string) (client.GetUserByEmailResponse, error) {
	return s.getByIDFunc(ctx, userID, headers)
}

func newAuthHandlerWithStub(c stubUserClient) *AuthHandler {
	svc := service.NewAuthService("secret", c)
	return NewAuthHandler(svc)
}

func TestRegisterHandler(t *testing.T) {
	h := newAuthHandlerWithStub(stubUserClient{
		createFn: func(ctx context.Context, email, password string, headers map[string]string) (client.CreateUserResponse, error) {
			require.NotEmpty(t, password) // hashed
			return client.CreateUserResponse{ID: "u-1"}, nil
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(`{"email":"alice@example.com","password":"pass"}`))
	rr := httptest.NewRecorder()

	h.Register(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
}

func TestRegisterHandler_Conflict(t *testing.T) {
	h := newAuthHandlerWithStub(stubUserClient{
		createFn: func(ctx context.Context, email, password string, headers map[string]string) (client.CreateUserResponse, error) {
			return client.CreateUserResponse{}, client.ErrUserExists
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(`{"email":"dup@example.com","password":"pass"}`))
	rr := httptest.NewRecorder()

	h.Register(rr, req)

	require.Equal(t, http.StatusConflict, rr.Code)
}

func TestLoginHandler(t *testing.T) {
	hashed, _ := bcrypt.GenerateFromPassword([]byte("pass"), bcrypt.DefaultCost)
	h := newAuthHandlerWithStub(stubUserClient{
		getByEmail: func(ctx context.Context, email string, headers map[string]string) (client.GetUserByEmailResponse, error) {
			return client.GetUserByEmailResponse{ID: "u-1", Email: email, Password: string(hashed)}, nil
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(`{"email":"alice@example.com","password":"pass"}`))
	rr := httptest.NewRecorder()

	h.Login(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var resp map[string]string
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	require.NotEmpty(t, resp["access_token"])
}

func TestLoginHandler_Invalid(t *testing.T) {
	h := newAuthHandlerWithStub(stubUserClient{
		getByEmail: func(ctx context.Context, email string, headers map[string]string) (client.GetUserByEmailResponse, error) {
			return client.GetUserByEmailResponse{}, client.ErrUserNotFound
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(`{"email":"missing@example.com","password":"pass"}`))
	rr := httptest.NewRecorder()

	h.Login(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestMeHandler(t *testing.T) {
	meHandler := Me(stubUserClient{
		getByIDFunc: func(ctx context.Context, userID string, headers map[string]string) (client.GetUserByEmailResponse, error) {
			return client.GetUserByEmailResponse{ID: userID, Email: "alice@example.com"}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "u-1")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	meHandler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	require.Equal(t, "u-1", resp["id"])
	require.Equal(t, "alice@example.com", resp["email"])
}

func TestMeHandler_NotFound(t *testing.T) {
	meHandler := Me(stubUserClient{
		getByIDFunc: func(ctx context.Context, userID string, headers map[string]string) (client.GetUserByEmailResponse, error) {
			return client.GetUserByEmailResponse{}, client.ErrUserNotFound
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "u-2")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	meHandler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusNotFound, rr.Code)
}

func TestMeHandler_MissingUserID(t *testing.T) {
	meHandler := Me(stubUserClient{})

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	rr := httptest.NewRecorder()

	meHandler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
}
