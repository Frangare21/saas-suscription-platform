package server

import (
	"context"
	"net/http"
	"saas-subscription-platform/services/auth-service/internal/client"
	"saas-subscription-platform/services/auth-service/internal/config"
	"saas-subscription-platform/services/auth-service/internal/handler"
	"saas-subscription-platform/services/auth-service/internal/middleware"
	"saas-subscription-platform/services/auth-service/internal/service"
	"time"
)

type Server struct {
	httpServer *http.Server
}

func New(cfg config.Config) *Server {
	userClient := client.NewUserClient(cfg.UserServiceURL)
	authSvc := service.NewAuthService(cfg.JWTSecret, userClient)
	authHandler := handler.NewAuthHandler(authSvc)

	internalAuthMiddleware := middleware.InternalAuth

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handler.Health)
	mux.HandleFunc("POST /register", authHandler.Register)
	mux.HandleFunc("POST /login", authHandler.Login)

	// Protected route - ahora usa internal auth en lugar de JWT
	mux.Handle("GET /me", internalAuthMiddleware(http.HandlerFunc(handler.Me(userClient))))

	return &Server{
		httpServer: &http.Server{
			Addr:         cfg.HTTPAddr,
			Handler:      mux,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
	}
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
