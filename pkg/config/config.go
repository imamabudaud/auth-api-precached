package config

import (
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DB struct {
		Host     string
		Port     int
		User     string
		Password string
		Name     string
	}
	Redis struct {
		Host     string
		Port     int
		Password string
		DB       int
		Prefix   string
		TTL      string
	}
	JWT struct {
		PrivateKeyPath string
		PublicKeyPath  string
		Expiration     time.Duration
	}
	Features struct {
		CacheEnabled    bool
		PrecacheEnabled bool
	}
	Precache struct {
		BatchSize    int
		CronSchedule string
	}
	Log struct {
		Level string
	}
	Service struct {
		AuthBasicPort    int
		AuthImprovedPort int
	}
}

func Load() *Config {
	// Try multiple locations for .env file
	envFiles := []string{
		".env",          // Current directory
		"../.env",       // Parent directory
		"../../.env",    // Grandparent directory
		"../../../.env", // Great-grandparent directory
	}

	var loaded bool
	for _, envFile := range envFiles {
		if err := godotenv.Load(envFile); err == nil {
			slog.Info("Loaded environment from", "file", envFile)
			loaded = true
			break
		}
	}

	if !loaded {
		slog.Warn("No .env file found, using default values")
	}

	cfg := &Config{}

	cfg.DB.Host = getEnv("DB_HOST", "localhost")
	cfg.DB.Port = getEnvAsInt("DB_PORT", 3306)
	cfg.DB.User = getEnv("DB_USER", "root")
	cfg.DB.Password = getEnv("DB_PASSWORD", "root")
	cfg.DB.Name = getEnv("DB_NAME", "auth")

	cfg.Redis.Host = getEnv("REDIS_HOST", "localhost")
	cfg.Redis.Port = getEnvAsInt("REDIS_PORT", 6379)
	cfg.Redis.Password = getEnv("REDIS_PASSWORD", "")
	cfg.Redis.DB = getEnvAsInt("REDIS_DB", 0)
	cfg.Redis.Prefix = getEnv("REDIS_PREFIX", "auth:")
	cfg.Redis.TTL = getEnv("REDIS_TTL", "1h")

	cfg.JWT.PrivateKeyPath = getEnv("JWT_PRIVATE_KEY_PATH", "./keys/private.pem")
	cfg.JWT.PublicKeyPath = getEnv("JWT_PUBLIC_KEY_PATH", "./keys/public.pem")
	cfg.JWT.Expiration = getEnvAsDuration("JWT_EXPIRATION", time.Hour)

	cfg.Features.CacheEnabled = getEnvAsBool("CACHE_ENABLED", true)
	cfg.Features.PrecacheEnabled = getEnvAsBool("PRECACHE_ENABLED", true)

	cfg.Precache.BatchSize = getEnvAsInt("BATCH_SIZE", 10000)
	cfg.Precache.CronSchedule = getEnv("CRON_SCHEDULE", "* * * * *")

	cfg.Log.Level = getEnv("LOG_LEVEL", "debug")

	cfg.Service.AuthBasicPort = getEnvAsInt("AUTH_BASIC_PORT", 8080)
	cfg.Service.AuthImprovedPort = getEnvAsInt("AUTH_IMPROVED_PORT", 8081)

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
