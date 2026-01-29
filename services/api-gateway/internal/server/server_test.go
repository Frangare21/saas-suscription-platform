package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"saas-subscription-platform/services/api-gateway/internal/config"
	"saas-subscription-platform/services/api-gateway/internal/middleware"

	"saas-subscription-platform/libs/trace"

	"github.com/golang-jwt/jwt/v5"
)

func makeToken(t *testing.T, secret, userID string) string {
	t.Helper()
	claims := jwt.MapClaims{"sub": userID, "exp": time.Now().Add(time.Hour).Unix()}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	return signed
}

func TestServer_PublicRoute_ProxiesAuth(t *testing.T) {
	authHits := make(chan *http.Request, 1)
	authBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHits <- r
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("ok"))
	}))
	defer authBackend.Close()

	cfg := config.Config{
		JWTSecret:         "secret",
		AuthServiceURL:    authBackend.URL,
		UserServiceURL:    "http://localhost", // unused in this test
		BillingServiceURL: "http://localhost",
	}

	srv := New(cfg)
	ts := httptest.NewServer(srv.httpServer.Handler)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/api/auth/register", "application/json", strings.NewReader(`{"email":"a"}`))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	select {
	case r := <-authHits:
		if r.URL.Path != "/register" {
			t.Fatalf("expected /register, got %s", r.URL.Path)
		}
	case <-time.After(time.Second):
		t.Fatalf("proxy did not hit auth backend")
	}
}

func TestServer_ProtectedRoute_AddsInternalHeaders(t *testing.T) {
	userHits := make(chan *http.Request, 1)
	userBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userHits <- r
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer userBackend.Close()

	cfg := config.Config{
		JWTSecret:         "secret",
		AuthServiceURL:    "http://localhost",
		UserServiceURL:    userBackend.URL,
		BillingServiceURL: "http://localhost",
	}

	srv := New(cfg)
	ts := httptest.NewServer(srv.httpServer.Handler)
	defer ts.Close()

	token := makeToken(t, "secret", "user-1")

	req, err := http.NewRequest(http.MethodGet, ts.URL+"/api/users/123?status=active", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d, body=%s", resp.StatusCode, string(body))
	}

	select {
	case r := <-userHits:
		if r.URL.Path != "/users/123" {
			t.Fatalf("expected /users/123, got %s", r.URL.Path)
		}
		if r.URL.RawQuery != "status=active" {
			t.Fatalf("expected query preserved, got %s", r.URL.RawQuery)
		}
		if auth := r.Header.Get("Authorization"); auth != "" {
			t.Fatalf("expected Authorization to be stripped, got %q", auth)
		}
		if userID := r.Header.Get(middleware.InternalUserIDHeader); userID != "user-1" {
			t.Fatalf("expected internal user id, got %s", userID)
		}
		if reqID := r.Header.Get(middleware.InternalRequestIDHeader); reqID == "" {
			t.Fatalf("expected internal request id header")
		}
		if callStack := r.Header.Get(trace.HeaderCallStack); !strings.Contains(callStack, "api-gateway") {
			t.Fatalf("expected call stack to include api-gateway, got %s", callStack)
		}
	case <-time.After(time.Second):
		t.Fatalf("proxy did not hit user backend")
	}
}
