package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
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

func (c *UserClient) CreateUser(email, password string, headers map[string]string) (CreateUserResponse, error) {
	reqBody := CreateUserRequest{
		Email:    email,
		Password: password,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return CreateUserResponse{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+"/users", bytes.NewBuffer(jsonData))
	if err != nil {
		return CreateUserResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return CreateUserResponse{}, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return CreateUserResponse{}, fmt.Errorf("failed to read response: %w", err)
	}

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
	encodedEmail := url.QueryEscape(email)
	req, err := http.NewRequest("GET", c.baseURL+"/users/email/"+encodedEmail, nil)
	if err != nil {
		return GetUserByEmailResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return GetUserByEmailResponse{}, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return GetUserByEmailResponse{}, fmt.Errorf("failed to read response: %w", err)
	}

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

func (c *UserClient) GetUserByID(userID string) (GetUserByEmailResponse, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/users/"+userID, nil)
	if err != nil {
		return GetUserByEmailResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return GetUserByEmailResponse{}, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return GetUserByEmailResponse{}, fmt.Errorf("failed to read response: %w", err)
	}

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
