package database

import (
	"fmt"
	"log/slog"

	"substack-auth/pkg/config"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type Database struct {
	DB *sqlx.DB
}

func New(cfg *config.Config) (*Database, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&multiStatements=true",
		cfg.DB.User, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.Name)

	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)

	slog.Info("Connected to database", "host", cfg.DB.Host, "port", cfg.DB.Port, "database", cfg.DB.Name)

	return &Database{DB: db}, nil
}

func (d *Database) Close() error {
	return d.DB.Close()
}
