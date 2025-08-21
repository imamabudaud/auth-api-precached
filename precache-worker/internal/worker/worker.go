package worker

import (
	"context"
	"encoding/json"
	"log/slog"

	"substack-auth/pkg/config"
	"substack-auth/pkg/database"
	"substack-auth/pkg/models"
	"substack-auth/pkg/redis"
)

type Worker struct {
	db    *database.Database
	redis *redis.Redis
	cfg   *config.Config
}

func New(db *database.Database, redis *redis.Redis, cfg *config.Config) *Worker {
	return &Worker{
		db:    db,
		redis: redis,
		cfg:   cfg,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	slog.Info("Starting precache worker", "batch_size", w.cfg.Precache.BatchSize)

	lastID := int64(0)
	totalProcessed := 0

	for {
		users, err := w.loadUsersBatch(ctx, lastID, w.cfg.Precache.BatchSize)
		if err != nil {
			return err
		}

		if len(users) == 0 {
			break
		}

		if err := w.cacheUsers(ctx, users); err != nil {
			return err
		}

		// Update lastID for next iteration (cursor-based pagination)
		lastID = users[len(users)-1].ID
		totalProcessed += len(users)

		slog.Info("Processed batch", "last_id", lastID, "count", len(users), "total_processed", totalProcessed)

		if len(users) < w.cfg.Precache.BatchSize {
			break
		}
	}

	slog.Info("Precache worker completed", "total_processed", totalProcessed)
	return nil
}

func (w *Worker) loadUsersBatch(ctx context.Context, lastID int64, limit int) ([]models.User, error) {
	var users []models.User
	query := `SELECT id, username, password_hash, created_at FROM users WHERE id > ? ORDER BY id LIMIT ?`

	err := w.db.DB.SelectContext(ctx, &users, query, lastID, limit)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (w *Worker) cacheUsers(ctx context.Context, users []models.User) error {
	// Prepare batch data
	batchData := make(map[string]string)

	for _, user := range users {
		data, err := json.Marshal(user)
		if err != nil {
			slog.Error("Failed to marshal user", "username", user.Username, "error", err)
			continue
		}
		batchData[user.Username] = string(data)
	}

	// Use batch operation
	if err := w.redis.SetBatch(ctx, batchData); err != nil {
		slog.Error("Failed to cache users batch", "count", len(batchData), "error", err)
		return err
	}

	slog.Debug("Cached users batch", "count", len(batchData))
	return nil
}
