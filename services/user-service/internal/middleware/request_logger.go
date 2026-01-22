package middleware

import (
	"log"
	"net/http"
	"time"

	"saas-subscription-platform/libs/trace"
)

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// RequestLogger loguea inicio/fin de request con request_id y call_stack.
func RequestLogger(serviceName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}

			ctx := r.Context()
			log.Printf("request_start service=%s method=%s path=%s request_id=%s call_stack=%s",
				serviceName, r.Method, r.URL.Path, trace.RequestIDFromContext(ctx), trace.CallStackFromContext(ctx))

			next.ServeHTTP(sw, r)

			log.Printf("request_end service=%s method=%s path=%s status=%d duration_ms=%d request_id=%s call_stack=%s",
				serviceName, r.Method, r.URL.Path, sw.status, time.Since(start).Milliseconds(), trace.RequestIDFromContext(ctx), trace.CallStackFromContext(ctx))
		})
	}
}
