package redis

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"substack-auth/pkg/config"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
	prefix string
	ttl    time.Duration
}

func New(cfg *config.Config) (*Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	// Parse TTL duration
	ttl, err := time.ParseDuration(cfg.Redis.TTL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis TTL: %w", err)
	}

	slog.Info("Connected to Redis", "host", cfg.Redis.Host, "port", cfg.Redis.Port, "prefix", cfg.Redis.Prefix, "ttl", ttl)

	return &Redis{client: client, prefix: cfg.Redis.Prefix, ttl: ttl}, nil
}

func (r *Redis) Set(ctx context.Context, key, value string) error {
	prefixedKey := r.prefix + key
	return r.client.Set(ctx, prefixedKey, value, r.ttl).Err()
}

func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	prefixedKey := r.prefix + key
	return r.client.Get(ctx, prefixedKey).Result()
}

func (r *Redis) SetBatch(ctx context.Context, data map[string]string) error {
	if len(data) == 0 {
		return nil
	}

	pipe := r.client.Pipeline()
	for key, value := range data {
		prefixedKey := r.prefix + key
		pipe.Set(ctx, prefixedKey, value, r.ttl)
	}

	_, err := pipe.Exec(ctx)
	return err
}

func (r *Redis) Close() error {
	return r.client.Close()
}
