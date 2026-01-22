package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"saas-subscription-platform/services/auth-service/internal/client"
	"saas-subscription-platform/services/auth-service/internal/middleware"
)

func Me(userClient *client.UserClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDValue := r.Context().Value(middleware.UserIDKey)
		if userIDValue == nil {
			http.Error(w, "user ID not found", http.StatusUnauthorized)
			return
		}

		userID, ok := userIDValue.(string)
		if !ok || userID == "" {
			http.Error(w, "invalid user ID", http.StatusUnauthorized)
			return
		}

		headers := map[string]string{
			"X-Internal-User-ID": "auth-service",
		}

		user, err := userClient.GetUserByIDWithContext(r.Context(), userID, headers)
		if err != nil {
			if err == client.ErrUserNotFound {
				http.Error(w, "user not found", http.StatusNotFound)
				return
			}
			log.Printf("auth_me_failed err=%v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
		})
	}
}
