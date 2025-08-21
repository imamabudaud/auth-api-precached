package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"substack-auth/pkg/config"
	"substack-auth/pkg/database"
	"substack-auth/pkg/redis"
	"substack-auth/precache-worker/internal/worker"

	"github.com/robfig/cron/v3"
)

func main() {
	cfg := config.Load()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	if !cfg.Features.PrecacheEnabled {
		logger.Info("Precache worker is disabled, exiting")
		return
	}

	db, err := database.New(cfg)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	redisClient, err := redis.New(cfg)
	if err != nil {
		logger.Error("Failed to connect to Redis", "error", err)
		os.Exit(1)
	}
	defer redisClient.Close()

	precacheWorker := worker.New(db, redisClient, cfg)

	// Run immediately first
	logger.Info("Running precache worker immediately...")
	if err := precacheWorker.Run(context.Background()); err != nil {
		logger.Error("Initial precache worker run failed", "error", err)
	} else {
		logger.Info("Initial precache worker run completed successfully")
	}

	// Then schedule to run every 5 minutes
	c := cron.New(cron.WithSeconds())
	entryID := c.Schedule(cron.Every(5*time.Minute), cron.FuncJob(func() {
		if err := precacheWorker.Run(context.Background()); err != nil {
			logger.Error("Precache worker failed", "error", err)
		}
	}))

	logger.Info("Precache worker scheduled to run every 5 minutes", "entry_id", entryID)

	c.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down precache worker...")
	c.Stop()
	logger.Info("Precache worker stopped")
}
