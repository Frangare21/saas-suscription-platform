package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"saas-subscription-platform/libs/trace"
)

func TestInternalHeaders_SetsHeaders(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(withUserID(req.Context(), "user-1"))

	var gotUserID, gotReqID, gotCallStack string
	InternalHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID = r.Header.Get(InternalUserIDHeader)
		gotReqID = r.Header.Get(InternalRequestIDHeader)
		gotCallStack = r.Header.Get(trace.HeaderCallStack)
	})).ServeHTTP(rr, req)

	if gotUserID != "user-1" {
		t.Fatalf("expected user id header set")
	}
	if gotReqID == "" {
		t.Fatalf("expected request id set")
	}
	if gotCallStack == "" || gotCallStack != trace.AppendServiceToStack("", "api-gateway") {
		t.Fatalf("unexpected call stack: %s", gotCallStack)
	}
}

func withUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}
