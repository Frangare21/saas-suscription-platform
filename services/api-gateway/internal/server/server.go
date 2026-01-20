package server

import (
	"context"
	"net/http"
	"saas-subscription-platform/services/api-gateway/internal/config"
	"saas-subscription-platform/services/api-gateway/internal/middleware"
	"saas-subscription-platform/services/api-gateway/internal/router"
	"time"
)

type Server struct {
	httpServer *http.Server
}

func New(cfg config.Config) *Server {
	gatewayRouter := router.NewRouter()
	gatewayRouter.SetAuthServiceURL(cfg.AuthServiceURL)
	gatewayRouter.SetUserServiceURL(cfg.UserServiceURL)
	gatewayRouter.SetBillingServiceURL(cfg.BillingServiceURL)

	jwtMiddleware := middleware.JWT(cfg.JWTSecret)
	internalHeadersMiddleware := middleware.InternalHeaders

	mux := http.NewServeMux()

	// Health check (no auth required)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// Public routes (no auth)
	mux.HandleFunc("POST /api/auth/register", gatewayRouter.ServeHTTP)
	mux.HandleFunc("POST /api/auth/login", gatewayRouter.ServeHTTP)

	// Protected routes (require auth)
	mux.Handle("/api/auth/me", jwtMiddleware(internalHeadersMiddleware(gatewayRouter)))
	mux.Handle("/api/users/", jwtMiddleware(internalHeadersMiddleware(gatewayRouter)))
	mux.Handle("/api/billing/", jwtMiddleware(internalHeadersMiddleware(gatewayRouter)))

	return &Server{
		httpServer: &http.Server{
			Addr:         cfg.HTTPAddr,
			Handler:      mux,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
	}
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
