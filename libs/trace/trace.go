package trace

import (
	"context"
	"net/http"
	"strings"
)

const (
	// HeaderRequestID es el request id interno que viaja entre servicios.
	HeaderRequestID = "X-Internal-Request-ID"
	// HeaderCallStack representa la cadena de servicios por los que pasó la request.
	// Formato: "api-gateway>auth-service>user-service".
	HeaderCallStack = "X-Internal-Call-Stack"
)

type contextKey string

const (
	requestIDKey contextKey = "request_id"
	callStackKey contextKey = "call_stack"
)

func RequestIDFromContext(ctx context.Context) string {
	if v := ctx.Value(requestIDKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func CallStackFromContext(ctx context.Context) string {
	if v := ctx.Value(callStackKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	if requestID == "" {
		return ctx
	}
	return context.WithValue(ctx, requestIDKey, requestID)
}

func WithCallStack(ctx context.Context, callStack string) context.Context {
	if callStack == "" {
		return ctx
	}
	return context.WithValue(ctx, callStackKey, callStack)
}

// AppendServiceToStack agrega serviceName al final del call stack.
func AppendServiceToStack(existing, serviceName string) string {
	serviceName = strings.TrimSpace(serviceName)
	if serviceName == "" {
		return existing
	}
	if existing == "" {
		return serviceName
	}
	// Evitar duplicar si ya termina con el service actual.
	if strings.HasSuffix(existing, ">"+serviceName) || existing == serviceName {
		return existing
	}
	return existing + ">" + serviceName
}

// InjectHeaders copia request_id y call_stack del ctx a los headers del request.
func InjectHeaders(ctx context.Context, req *http.Request) {
	if req == nil {
		return
	}
	if rid := RequestIDFromContext(ctx); rid != "" {
		req.Header.Set(HeaderRequestID, rid)
	}
	if cs := CallStackFromContext(ctx); cs != "" {
		req.Header.Set(HeaderCallStack, cs)
	}
}

// ExtractAndUpdateContext lee headers de trazabilidad y actualiza el context.
// Si serviceName != "", lo agrega al call stack.
func ExtractAndUpdateContext(ctx context.Context, r *http.Request, serviceName string) context.Context {
	if r == nil {
		return ctx
	}
	requestID := r.Header.Get(HeaderRequestID)
	callStack := r.Header.Get(HeaderCallStack)
	callStack = AppendServiceToStack(callStack, serviceName)
	ctx = WithRequestID(ctx, requestID)
	ctx = WithCallStack(ctx, callStack)
	return ctx
}
