package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"

	"substack-auth/pkg/config"
	"substack-auth/pkg/database"
	"substack-auth/pkg/models"
	"substack-auth/pkg/redis"
)

type UserService struct {
	db    *database.Database
	redis *redis.Redis
	cfg   *config.Config
}

func NewUserService(db *database.Database, redis *redis.Redis, cfg *config.Config) *UserService {
	return &UserService{
		db:    db,
		redis: redis,
		cfg:   cfg,
	}
}

func (s *UserService) GetByUsername(username string) (*models.User, error) {
	if s.cfg.Features.CacheEnabled {
		if user, err := s.getFromCache(username); err == nil {
			return user, nil
		}
	}

	slog.Debug("Cache miss: ", "username", username)

	user, err := s.getFromDatabase(username)
	if err != nil {
		return nil, err
	}

	if s.cfg.Features.CacheEnabled {
		s.cacheUser(username, user)
	}

	return user, nil
}

func (s *UserService) getFromCache(username string) (*models.User, error) {
	ctx := context.Background()
	data, err := s.redis.Get(ctx, username)
	if err != nil {
		return nil, err
	}

	var user models.User
	if err := json.Unmarshal([]byte(data), &user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *UserService) getFromDatabase(username string) (*models.User, error) {
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

func (s *UserService) cacheUser(username string, user *models.User) {
	ctx := context.Background()
	data, err := json.Marshal(user)
	if err != nil {
		slog.Error("Failed to marshal user for cache", "username", username, "error", err)
		return
	}

	if err := s.redis.Set(ctx, username, string(data)); err != nil {
		slog.Error("Failed to cache user", "username", username, "error", err)
	}
}
