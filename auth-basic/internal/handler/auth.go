package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"substack-auth/pkg/models"
)

type AuthHandler struct {
	authService AuthService
}

type AuthService interface {
	Login(req *models.LoginRequest) (*models.LoginResponse, error)
}

func NewAuthHandler(authService AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Failed to decode request", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	response, err := h.authService.Login(&req)
	if err != nil {
		slog.Error("Login failed", "username", req.Username, "error", err)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Create secure response without password hash
	secureResponse := models.LoginResponse{
		Token: response.Token,
		User: models.UserResponse{
			ID:        response.User.ID,
			Username:  response.User.Username,
			CreatedAt: response.User.CreatedAt,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(secureResponse); err != nil {
		slog.Error("Failed to encode response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
