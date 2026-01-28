package router

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Route struct {
	Path         string
	TargetURL    string
	RequiresAuth bool
}

type Router struct {
	routes []Route
	client *http.Client
}

func NewRouter() *Router {
	return &Router{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		routes: []Route{
			// Auth routes (no auth required)
			{Path: "/api/auth/register", TargetURL: "", RequiresAuth: false},
			{Path: "/api/auth/login", TargetURL: "", RequiresAuth: false},

			// Protected routes (require auth)
			{Path: "/api/auth/me", TargetURL: "", RequiresAuth: true},

			// User routes (require auth)
			{Path: "/api/users", TargetURL: "", RequiresAuth: true},

			// Billing routes (require auth)
			{Path: "/api/billing", TargetURL: "", RequiresAuth: true},
		},
	}
}

func (r *Router) SetAuthServiceURL(url string) {
	for i := range r.routes {
		if strings.HasPrefix(r.routes[i].Path, "/api/auth") {
			r.routes[i].TargetURL = url
		}
	}
}

func (r *Router) SetUserServiceURL(url string) {
	for i := range r.routes {
		if strings.HasPrefix(r.routes[i].Path, "/api/users") {
			r.routes[i].TargetURL = url
		}
	}
}

func (r *Router) SetBillingServiceURL(url string) {
	for i := range r.routes {
		if strings.HasPrefix(r.routes[i].Path, "/api/billing") {
			r.routes[i].TargetURL = url
		}
	}
}

func (r *Router) FindRoute(path string) *Route {
	for _, route := range r.routes {
		if strings.HasPrefix(path, route.Path) || path == route.Path {
			return &route
		}
	}
	return nil
}

func (r *Router) Proxy(w http.ResponseWriter, req *http.Request, targetURL string) {
	// Parse target URL
	target, err := url.Parse(targetURL)
	if err != nil {
		http.Error(w, "invalid target URL", http.StatusInternalServerError)
		return
	}

	// Create new request
	proxyReq := req.Clone(req.Context())
	proxyReq.URL.Scheme = target.Scheme
	proxyReq.URL.Host = target.Host
	proxyReq.RequestURI = ""

	// Map paths: /api/auth/* -> /*, /api/users/* -> /*, /api/billing/* -> /*
	if strings.HasPrefix(req.URL.Path, "/api/auth") {
		// Remove /api/auth prefix
		proxyReq.URL.Path = strings.TrimPrefix(req.URL.Path, "/api/auth")
		if proxyReq.URL.Path == "" {
			proxyReq.URL.Path = "/"
		} else if !strings.HasPrefix(proxyReq.URL.Path, "/") {
			proxyReq.URL.Path = "/" + proxyReq.URL.Path
		}
	} else if strings.HasPrefix(req.URL.Path, "/api/users") {
		// Map /api/users/* to /users/* for the user-service
		suffix := strings.TrimPrefix(req.URL.Path, "/api/users")
		if suffix == "" {
			proxyReq.URL.Path = "/users"
		} else {
			if !strings.HasPrefix(suffix, "/") {
				suffix = "/" + suffix
			}
			proxyReq.URL.Path = "/users" + suffix
		}
	} else if strings.HasPrefix(req.URL.Path, "/api/billing") {
		// Remove /api/billing prefix, keep the rest
		proxyReq.URL.Path = strings.TrimPrefix(req.URL.Path, "/api/billing")
		if proxyReq.URL.Path == "" {
			proxyReq.URL.Path = "/"
		} else if !strings.HasPrefix(proxyReq.URL.Path, "/") {
			proxyReq.URL.Path = "/" + proxyReq.URL.Path
		}
	}

	// Copy query parameters
	proxyReq.URL.RawQuery = req.URL.RawQuery

	// Copy headers (except Host and Authorization for internal services)
	proxyReq.Header = make(http.Header)
	for key, values := range req.Header {
		if key != "Host" {
			// Remove Authorization header for internal services (they use X-Internal-User-ID)
			if key == "Authorization" && (strings.HasPrefix(req.URL.Path, "/api/users") || strings.HasPrefix(req.URL.Path, "/api/billing")) {
				continue
			}
			for _, value := range values {
				proxyReq.Header.Add(key, value)
			}
		}
	}

	// Forward the request
	resp, err := r.client.Do(proxyReq)
	if err != nil {
		http.Error(w, "proxy error: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Copy status code
	w.WriteHeader(resp.StatusCode)

	// Copy response body
	_, _ = io.Copy(w, resp.Body)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	route := r.FindRoute(req.URL.Path)
	if route == nil {
		http.Error(w, "route not found", http.StatusNotFound)
		return
	}

	if route.TargetURL == "" {
		http.Error(w, "service not configured", http.StatusInternalServerError)
		return
	}

	r.Proxy(w, req, route.TargetURL)
}
