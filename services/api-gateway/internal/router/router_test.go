package router

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// stubClient lets us intercept requests without hitting the network.
type stubClient struct {
	doFn func(req *http.Request) (*http.Response, error)
}

func (s stubClient) Do(req *http.Request) (*http.Response, error) {
	return s.doFn(req)
}

func TestFindRoute(t *testing.T) {
	r := NewRouter()
	if r.FindRoute("/api/unknown") != nil {
		t.Fatalf("expected nil for unknown route")
	}
	if r.FindRoute("/api/auth/login") == nil {
		t.Fatalf("expected auth route present")
	}
}

func TestServeHTTP_ServiceNotConfigured(t *testing.T) {
	r := NewRouter()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/auth/login", nil)

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
}

func TestProxy_RewritesAuthPath(t *testing.T) {
	r := NewRouterWithClient(stubClient{doFn: func(req *http.Request) (*http.Response, error) {
		if req.URL.Path != "/login" {
			t.Fatalf("expected /login, got %s", req.URL.Path)
		}
		body := io.NopCloser(strings.NewReader(`{"ok":true}`))
		return &http.Response{StatusCode: http.StatusOK, Body: body, Header: http.Header{}}, nil
	}})
	r.SetAuthServiceURL("http://auth.test")

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestProxy_RewritesUserPathAndDropsAuthHeader(t *testing.T) {
	var forwardedAuth string
	r := NewRouterWithClient(stubClient{doFn: func(req *http.Request) (*http.Response, error) {
		if req.URL.Path != "/users/123" {
			t.Fatalf("expected /users/123, got %s", req.URL.Path)
		}
		forwardedAuth = req.Header.Get("Authorization")
		body := io.NopCloser(strings.NewReader(`{"id":"123"}`))
		return &http.Response{StatusCode: http.StatusOK, Body: body, Header: http.Header{"Content-Type": []string{"application/json"}}}, nil
	}})
	r.SetUserServiceURL("http://user.test")

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/users/123", nil)
	req.Header.Set("Authorization", "Bearer token")

	r.ServeHTTP(rr, req)

	if forwardedAuth != "" {
		t.Fatalf("expected Authorization to be stripped, got %q", forwardedAuth)
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestProxy_RewritesBillingPathAndKeepsQuery(t *testing.T) {
	r := NewRouterWithClient(stubClient{doFn: func(req *http.Request) (*http.Response, error) {
		if req.URL.Path != "/invoices" {
			t.Fatalf("expected /invoices, got %s", req.URL.Path)
		}
		if req.URL.RawQuery != "status=pending" {
			t.Fatalf("expected query preserved, got %s", req.URL.RawQuery)
		}
		body := io.NopCloser(strings.NewReader(`[]`))
		return &http.Response{StatusCode: http.StatusOK, Body: body, Header: http.Header{}}, nil
	}})
	r.SetBillingServiceURL("http://billing.test")

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/billing/invoices?status=pending", nil)

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}
