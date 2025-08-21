package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"substack-auth/auth-improved/internal/handler"
	"substack-auth/auth-improved/internal/service"
	"substack-auth/pkg/config"
	"substack-auth/pkg/database"
	"substack-auth/pkg/jwt"
	"substack-auth/pkg/redis"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg := config.Load()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

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

	jwtService, err := jwt.New(cfg)
	if err != nil {
		logger.Error("Failed to initialize JWT service", "error", err)
		os.Exit(1)
	}

	userService := service.NewUserService(db, redisClient, cfg)
	authService := service.NewAuthService(userService, jwtService)
	authHandler := handler.NewAuthHandler(authService)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	r.Post("/login", authHandler.Login)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Service.AuthImprovedPort),
		Handler: r,
	}

	go func() {
		logger.Info("Starting auth-improved service", "port", cfg.Service.AuthImprovedPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server error", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	logger.Info("Server exited")
}
