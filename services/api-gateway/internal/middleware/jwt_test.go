package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func makeToken(t *testing.T, secret, userID string, addExp bool) string {
	claims := jwt.MapClaims{"sub": userID}
	if addExp {
		claims["exp"] = time.Now().Add(time.Hour).Unix()
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	return signed
}

func TestJWT_MissingHeader(t *testing.T) {
	mw := JWT("secret")
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("handler should not be called")
	})).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestJWT_ValidTokenSetsContext(t *testing.T) {
	mw := JWT("secret")
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	token := makeToken(t, "secret", "user-1", true)
	req.Header.Set("Authorization", "Bearer "+token)

	var gotUser string
	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := r.Context().Value(UserIDKey)
		if v == nil {
			t.Fatalf("expected user id in context")
		}
		gotUser = v.(string)
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if gotUser != "user-1" {
		t.Fatalf("expected user-1, got %s", gotUser)
	}
}
