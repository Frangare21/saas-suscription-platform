package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"saas-subscription-platform/libs/trace"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserExists   = errors.New("user already exists")
	ErrServiceError = errors.New("user service error")
)

type UserClient struct {
	baseURL    string
	httpClient *http.Client
}

type CreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateUserResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

type GetUserByEmailResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	CreatedAt string `json:"created_at"`
}

func NewUserClient(baseURL string) *UserClient {
	return &UserClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// mergeHeaders aplica headers explícitos y completa con headers de trazabilidad desde ctx.
func mergeHeaders(ctx context.Context, headers map[string]string) map[string]string {
	out := make(map[string]string, len(headers)+2)
	for k, v := range headers {
		out[k] = v
	}
	if ctx != nil {
		if rid := trace.RequestIDFromContext(ctx); rid != "" {
			out[trace.HeaderRequestID] = rid
		}
		if cs := trace.CallStackFromContext(ctx); cs != "" {
			out[trace.HeaderCallStack] = cs
		}
	}
	return out
}

func (c *UserClient) CreateUser(email, password string, headers map[string]string) (CreateUserResponse, error) {
	return c.CreateUserWithContext(context.Background(), email, password, headers)
}

func (c *UserClient) CreateUserWithContext(ctx context.Context, email, password string, headers map[string]string) (CreateUserResponse, error) {
	start := time.Now()

	reqBody := CreateUserRequest{
		Email:    email,
		Password: password,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return CreateUserResponse{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/users", bytes.NewBuffer(jsonData))
	if err != nil {
		return CreateUserResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for key, value := range mergeHeaders(ctx, headers) {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("upstream_call failed service=user-service method=POST path=/users request_id=%s call_stack=%s duration_ms=%d err=%v",
			trace.RequestIDFromContext(ctx), trace.CallStackFromContext(ctx), time.Since(start).Milliseconds(), err)
		return CreateUserResponse{}, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("upstream_call failed service=user-service method=POST path=/users request_id=%s call_stack=%s duration_ms=%d err=%v",
			trace.RequestIDFromContext(ctx), trace.CallStackFromContext(ctx), time.Since(start).Milliseconds(), err)
		return CreateUserResponse{}, fmt.Errorf("failed to read response: %w", err)
	}

	log.Printf("upstream_call service=user-service method=POST path=/users request_id=%s call_stack=%s status=%d duration_ms=%d",
		trace.RequestIDFromContext(ctx), trace.CallStackFromContext(ctx), resp.StatusCode, time.Since(start).Milliseconds())

	if resp.StatusCode == http.StatusConflict {
		return CreateUserResponse{}, ErrUserExists
	}

	if resp.StatusCode != http.StatusCreated {
		return CreateUserResponse{}, fmt.Errorf("%w: status %d, body: %s", ErrServiceError, resp.StatusCode, string(body))
	}

	var userResp CreateUserResponse
	if err := json.Unmarshal(body, &userResp); err != nil {
		return CreateUserResponse{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return userResp, nil
}

func (c *UserClient) GetUserByEmail(email string, headers map[string]string) (GetUserByEmailResponse, error) {
	return c.GetUserByEmailWithContext(context.Background(), email, headers)
}

func (c *UserClient) GetUserByEmailWithContext(ctx context.Context, email string, headers map[string]string) (GetUserByEmailResponse, error) {
	start := time.Now()

	encodedEmail := url.QueryEscape(email)
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/users/email/"+encodedEmail, nil)
	if err != nil {
		return GetUserByEmailResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	for key, value := range mergeHeaders(ctx, headers) {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("upstream_call failed service=user-service method=GET path=/users/email/{email} request_id=%s call_stack=%s duration_ms=%d err=%v",
			trace.RequestIDFromContext(ctx), trace.CallStackFromContext(ctx), time.Since(start).Milliseconds(), err)
		return GetUserByEmailResponse{}, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("upstream_call failed service=user-service method=GET path=/users/email/{email} request_id=%s call_stack=%s duration_ms=%d err=%v",
			trace.RequestIDFromContext(ctx), trace.CallStackFromContext(ctx), time.Since(start).Milliseconds(), err)
		return GetUserByEmailResponse{}, fmt.Errorf("failed to read response: %w", err)
	}

	log.Printf("upstream_call service=user-service method=GET path=/users/email/{email} request_id=%s call_stack=%s status=%d duration_ms=%d",
		trace.RequestIDFromContext(ctx), trace.CallStackFromContext(ctx), resp.StatusCode, time.Since(start).Milliseconds())

	if resp.StatusCode == http.StatusNotFound {
		return GetUserByEmailResponse{}, ErrUserNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return GetUserByEmailResponse{}, fmt.Errorf("%w: status %d, body: %s", ErrServiceError, resp.StatusCode, string(body))
	}

	var userResp GetUserByEmailResponse
	if err := json.Unmarshal(body, &userResp); err != nil {
		return GetUserByEmailResponse{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return userResp, nil
}

func (c *UserClient) GetUserByID(userID string, headers map[string]string) (GetUserByEmailResponse, error) {
	return c.GetUserByIDWithContext(context.Background(), userID, headers)
}

func (c *UserClient) GetUserByIDWithContext(ctx context.Context, userID string, headers map[string]string) (GetUserByEmailResponse, error) {
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/users/"+userID, nil)
	if err != nil {
		return GetUserByEmailResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	for key, value := range mergeHeaders(ctx, headers) {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("upstream_call failed service=user-service method=GET path=/users/{id} request_id=%s call_stack=%s duration_ms=%d err=%v",
			trace.RequestIDFromContext(ctx), trace.CallStackFromContext(ctx), time.Since(start).Milliseconds(), err)
		return GetUserByEmailResponse{}, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("upstream_call failed service=user-service method=GET path=/users/{id} request_id=%s call_stack=%s duration_ms=%d err=%v",
			trace.RequestIDFromContext(ctx), trace.CallStackFromContext(ctx), time.Since(start).Milliseconds(), err)
		return GetUserByEmailResponse{}, fmt.Errorf("failed to read response: %w", err)
	}

	log.Printf("upstream_call service=user-service method=GET path=/users/{id} request_id=%s call_stack=%s status=%d duration_ms=%d",
		trace.RequestIDFromContext(ctx), trace.CallStackFromContext(ctx), resp.StatusCode, time.Since(start).Milliseconds())

	if resp.StatusCode == http.StatusNotFound {
		return GetUserByEmailResponse{}, ErrUserNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return GetUserByEmailResponse{}, fmt.Errorf("%w: status %d, body: %s", ErrServiceError, resp.StatusCode, string(body))
	}

	var userResp GetUserByEmailResponse
	if err := json.Unmarshal(body, &userResp); err != nil {
		return GetUserByEmailResponse{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return userResp, nil
}
