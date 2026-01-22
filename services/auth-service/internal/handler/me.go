package handler

import (
	"encoding/json"
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

		user, err := userClient.GetUserByID(userID, headers)
		if err != nil {
			if err == client.ErrUserNotFound {
				http.Error(w, "user not found", http.StatusNotFound)
				return
			}
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
