package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const pingTimeout = 5 * time.Second

func Open(dsn string, log *slog.Logger) (*sql.DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("DB_DSN is empty")
	}

	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()

	if err := conn.PingContext(ctx); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	log.Info("db connected")
	return conn, nil
}

func Close(conn *sql.DB) error {
	if conn == nil {
		return nil
	}

	if err := conn.Close(); err != nil {
		return fmt.Errorf("close db: %w", err)
	}

	return nil
}
