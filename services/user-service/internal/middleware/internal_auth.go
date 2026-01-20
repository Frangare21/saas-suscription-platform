package middleware

import (
	"context"
	"net/http"
)

type contextKey string

const UserIDKey contextKey = "user_id"

// InternalAuth lee el header interno X-Internal-User-ID y lo pone en el contexto
// Los microservicios confían en este header que viene del API Gateway
func InternalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-Internal-User-ID")
		if userID == "" {
			http.Error(w, "missing internal user ID", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
