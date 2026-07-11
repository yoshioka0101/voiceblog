package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yoshioka0101/voiceblog/backend/internal/config"
	"github.com/yoshioka0101/voiceblog/backend/internal/db"
	"github.com/yoshioka0101/voiceblog/backend/internal/logger"
	"github.com/yoshioka0101/voiceblog/backend/internal/server"
)

func main() {
	slog.SetDefault(logger.New(slog.LevelInfo))

	if err := run(); err != nil {
		slog.Error("api server exited", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	log := logger.New(cfg.LogLevel)
	slog.SetDefault(log)

	dbConn, err := db.Open(cfg.DBDSN, log)
	if err != nil {
		return fmt.Errorf("connect db: %w", err)
	}
	defer func() {
		if err := db.Close(dbConn); err != nil {
			log.Error("failed to close db", slog.String("error", err.Error()))
		}
	}()

	srv, err := server.New(dbConn, &cfg, log)
	if err != nil {
		return fmt.Errorf("create server: %w", err)
	}

	serverErr := make(chan error, 1)
	go func() {
		if err := srv.Run(":" + cfg.Port); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		return fmt.Errorf("run server: %w", err)
	case <-stop:
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown server: %w", err)
	}

	return nil
}
