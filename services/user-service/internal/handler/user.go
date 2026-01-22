package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"saas-subscription-platform/services/user-service/internal/repository"
	"saas-subscription-platform/services/user-service/internal/service"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

type CreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateUserRequest struct {
	Email    *string `json:"email"`
	Password *string `json:"password"`
}

type UserResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Password  string `json:"password,omitempty"`
	CreatedAt string `json:"created_at"`
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	user, err := h.userService.CreateUser(req.Email, req.Password)
	if err != nil {
		if err == repository.ErrUserExists {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		// Log the actual error for debugging
		log.Printf("Error creating user: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

func (h *UserHandler) GetUserByEmail(w http.ResponseWriter, r *http.Request) {
	email := r.PathValue("email")
	if email == "" {
		http.Error(w, "email parameter required", http.StatusBadRequest)
		return
	}

	user, err := h.userService.GetUserByEmail(email)
	if err != nil {
		if err == repository.ErrUserNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		// Log the actual error for debugging
		log.Printf("Error getting user by email: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Password:  user.Password,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	if userID == "" {
		http.Error(w, "id parameter required", http.StatusBadRequest)
		return
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		if err == repository.ErrUserNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		// Log the actual error for debugging
		log.Printf("Error getting user by id: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	if userID == "" {
		http.Error(w, "id parameter required", http.StatusBadRequest)
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	if req.Email == nil && req.Password == nil {
		http.Error(w, "no fields to update", http.StatusBadRequest)
		return
	}

	err := h.userService.UpdateUser(userID, req.Email, req.Password)
	if err != nil {
		if err == repository.ErrUserNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if err == repository.ErrUserExists {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		// Log the actual error for debugging
		log.Printf("Error updating user: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	if userID == "" {
		http.Error(w, "id parameter required", http.StatusBadRequest)
		return
	}

	err := h.userService.DeleteUser(userID)
	if err != nil {
		if err == repository.ErrUserNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		// Log the actual error for debugging
		log.Printf("Error deleting user: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
