package server

import (
	"context"
	"log"
	"net/http"
	"saas-subscription-platform/services/user-service/internal/config"
	"saas-subscription-platform/services/user-service/internal/db"
	"saas-subscription-platform/services/user-service/internal/handler"
	"saas-subscription-platform/services/user-service/internal/middleware"
	"saas-subscription-platform/services/user-service/internal/repository"
	"saas-subscription-platform/services/user-service/internal/service"
	"time"
)

type Server struct {
	httpServer *http.Server
}

func New(cfg config.Config) *Server {
	pool, err := db.New(cfg.DBDSN)
	if err != nil {
		log.Fatalf("db connection failed: %v", err)
	}

	userRepo := repository.NewUserRepository(pool)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	internalAuthMiddleware := middleware.InternalAuth
	requestLogger := middleware.RequestLogger("user-service")

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handler.Health)

	// Protected routes - requieren header interno del API Gateway
	mux.Handle("POST /users", internalAuthMiddleware(http.HandlerFunc(userHandler.CreateUser)))
	mux.Handle("GET /users/email/{email}", internalAuthMiddleware(http.HandlerFunc(userHandler.GetUserByEmail)))
	mux.Handle("GET /users/{id}", internalAuthMiddleware(http.HandlerFunc(userHandler.GetUserByID)))
	mux.Handle("PATCH /users/{id}", internalAuthMiddleware(http.HandlerFunc(userHandler.UpdateUser)))
	mux.Handle("DELETE /users/{id}", internalAuthMiddleware(http.HandlerFunc(userHandler.DeleteUser)))

	h := requestLogger(mux)

	return &Server{
		httpServer: &http.Server{
			Addr:         cfg.HTTPAddr,
			Handler:      h,
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
