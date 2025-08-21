package service

import (
	"fmt"
	"log/slog"

	"substack-auth/pkg/jwt"
	"substack-auth/pkg/models"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userService *UserService
	jwtService  *jwt.JWT
}

func NewAuthService(userService *UserService, jwtService *jwt.JWT) *AuthService {
	return &AuthService{
		userService: userService,
		jwtService:  jwtService,
	}
}

func (s *AuthService) Login(req *models.LoginRequest) (*models.LoginResponse, error) {
	user, err := s.userService.GetByUsername(req.Username)
	if err != nil {
		slog.Error("User not found", "username", req.Username)
		return nil, fmt.Errorf("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		slog.Error("Invalid password", "username", req.Username)
		return nil, fmt.Errorf("invalid credentials")
	}

	token, err := s.jwtService.GenerateToken(user.Username)
	if err != nil {
		slog.Error("Failed to generate token", "username", req.Username, "error", err)
		return nil, fmt.Errorf("failed to generate token")
	}

	slog.Info("User logged in successfully", "username", req.Username)

	return &models.LoginResponse{
		Token: token,
		User: models.UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			CreatedAt: user.CreatedAt,
		},
	}, nil
}
