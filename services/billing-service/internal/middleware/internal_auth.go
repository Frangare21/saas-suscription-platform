package middleware

import (
	"net/http"

	"github.com/gorilla/mux"
)

// InternalAuthMux es equivalente al middleware InternalAuth del user-service,
// pero adaptado a gorilla/mux (mux.MiddlewareFunc).
//
// Exige el header interno X-Internal-User-ID (agregado por el API Gateway).
func InternalAuthMux(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-Internal-User-ID")
		if userID == "" {
			http.Error(w, "missing internal user ID", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

var _ mux.MiddlewareFunc = InternalAuthMux
