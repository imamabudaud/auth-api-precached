package service

import (
	"database/sql"
	"fmt"
	"log/slog"

	"substack-auth/pkg/database"
	"substack-auth/pkg/models"
)

type UserService struct {
	db *database.Database
}

func NewUserService(db *database.Database) *UserService {
	return &UserService{db: db}
}

func (s *UserService) GetByUsername(username string) (*models.User, error) {
	var user models.User
	query := `SELECT id, username, password_hash, created_at FROM users WHERE username = ?`

	err := s.db.DB.Get(&user, query, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (s *UserService) Create(username, passwordHash string) error {
	query := `INSERT INTO users (username, password_hash) VALUES (?, ?)`

	_, err := s.db.DB.Exec(query, username, passwordHash)
	if err != nil {
		slog.Error("Failed to create user", "username", username, "error", err)
		return fmt.Errorf("failed to create user: %w", err)
	}

	slog.Info("User created", "username", username)
	return nil
}
