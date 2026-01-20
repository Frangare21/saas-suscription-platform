package middleware

import (
	"net/http"

	"github.com/google/uuid"
)

const (
	InternalUserIDHeader    = "X-Internal-User-ID"
	InternalRequestIDHeader = "X-Internal-Request-ID"
)

// InternalHeaders agrega headers internos para que los microservicios confíen en ellos
func InternalHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Obtener userID del contexto (agregado por el middleware JWT)
		if userID := r.Context().Value(UserIDKey); userID != nil {
			if userIDStr, ok := userID.(string); ok {
				r.Header.Set(InternalUserIDHeader, userIDStr)
			}
		}

		// Agregar request ID para trazabilidad
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		r.Header.Set(InternalRequestIDHeader, requestID)

		next.ServeHTTP(w, r)
	})
}
