package middleware

import (
	"net/http"

	"github.com/google/uuid"

	"saas-subscription-platform/libs/trace"
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

		// Call stack: inicializar si no viene (o si viene, conservar)
		callStack := r.Header.Get(trace.HeaderCallStack)
		if callStack == "" {
			callStack = trace.AppendServiceToStack("", "api-gateway")
		}
		r.Header.Set(trace.HeaderCallStack, callStack)

		next.ServeHTTP(w, r)
	})
}
