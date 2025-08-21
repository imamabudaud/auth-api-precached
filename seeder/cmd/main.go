package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"substack-auth/pkg/config"
	"substack-auth/pkg/database"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	var numUsers int
	var username string
	var password string

	flag.IntVar(&numUsers, "n", 1000, "Number of users to generate")
	flag.StringVar(&username, "username", "", "Single username to insert")
	flag.StringVar(&password, "password", "", "Password for single user")
	flag.Parse()

	startTime := time.Now()

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

	// Check if single user mode
	if username != "" && password != "" {
		if err := insertSingleUser(db, username, password); err != nil {
			logger.Error("Failed to insert single user", "username", username, "error", err)
			os.Exit(1)
		}
		logger.Info("Single user inserted successfully", "username", username, "password", password)
		return
	}

	// Generate shared password hash once for bulk mode
	sharedPassword := "testpassword123"
	sharedPasswordHash, err := bcrypt.GenerateFromPassword([]byte(sharedPassword), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("Failed to generate shared password hash", "error", err)
		os.Exit(1)
	}

	// Get the next available sequence number
	nextSequence, err := getNextSequence(db)
	if err != nil {
		logger.Error("Failed to get next sequence", "error", err)
		os.Exit(1)
	}

	logger.Info("Starting bulk user seeder", "count", numUsers, "shared_password", sharedPassword, "starting_from", nextSequence)

	var wg sync.WaitGroup
	var mu sync.Mutex
	semaphore := make(chan struct{}, 10) // Limit to 10 concurrent goroutines
	batchSize := 1000
	totalProcessed := 0

	// Start progress updater goroutine
	progressChan := make(chan int, 100)
	progressDone := make(chan struct{})

	go func() {
		for {
			select {
			case processed := <-progressChan:
				mu.Lock()
				totalProcessed += processed
				progress := float64(totalProcessed) / float64(numUsers) * 100
				bar := createProgressBar(progress, 40)
				elapsed := time.Since(startTime)
				rate := float64(totalProcessed) / elapsed.Seconds()
				fmt.Printf("\r%s %.1f%% (%d/%d) %.0f users/sec", bar, progress, totalProcessed, numUsers, rate)
				os.Stdout.Sync()
				mu.Unlock()
			case <-progressDone:
				return
			}
		}
	}()

	for i := 0; i < numUsers; i += batchSize {
		end := i + batchSize
		if end > numUsers {
			end = numUsers
		}

		wg.Add(1)
		go func(start, batchEnd int) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			if err := seedBatch(db, start, batchEnd, string(sharedPasswordHash), nextSequence); err != nil {
				logger.Error("Failed to seed batch", "start", start, "end", batchEnd, "error", err)
				return
			}

			// Send progress update to single updater
			progressChan <- (batchEnd - start)
		}(i, end)
	}

	wg.Wait()
	close(progressDone)

	fmt.Println()

	duration := time.Since(startTime)
	rate := float64(numUsers) / duration.Seconds()
	logger.Info("User seeder completed",
		"total_users", numUsers,
		"duration", duration,
		"rate", fmt.Sprintf("%.2f", rate),
		"unit", "users_per_second",
		"batch_size", batchSize)
}

func insertSingleUser(db *database.Database, username, password string) error {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	query := `INSERT INTO users (username, password_hash) VALUES (?, ?)`
	_, err = db.DB.Exec(query, username, passwordHash)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}

	return nil
}

func getNextSequence(db *database.Database) (int, error) {
	// Get the highest sequence number from existing usernames
	var maxSequence int
	query := `SELECT COALESCE(MAX(CAST(SUBSTRING_INDEX(username, '@', 1) AS UNSIGNED)), 0) FROM users`
	err := db.DB.Get(&maxSequence, query)
	if err != nil {
		return 0, fmt.Errorf("failed to get max sequence: %w", err)
	}
	return maxSequence + 1, nil
}

func seedBatch(db *database.Database, start, end int, sharedPasswordHash string, startSequence int) error {
	batchSize := end - start

	// Build bulk insert query with shared password hash
	placeholders := make([]string, batchSize)
	values := make([]interface{}, 0, batchSize*2)

	for i := 0; i < batchSize; i++ {
		placeholders[i] = "(?, ?)"
		username := generateUsername(startSequence + start + i) // startSequence is the base, start is batch offset, i is position in batch
		values = append(values, username, sharedPasswordHash)
	}

	query := `INSERT INTO users (username, password_hash) VALUES ` + strings.Join(placeholders, ", ")

	tx, err := db.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(query, values...)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func generateUsername(sequence int) string {
	return fmt.Sprintf("%08d@katakode.com", sequence)
}

func createProgressBar(percentage float64, width int) string {
	filled := int(percentage / 100 * float64(width))
	bar := "["

	for i := 0; i < width; i++ {
		if i < filled {
			bar += "="
		} else if i == filled {
			bar += ">"
		} else {
			bar += " "
		}
	}

	bar += "]"
	return bar
}
