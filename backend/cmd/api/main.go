package main

import (
	"context"
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

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	log := logger.New(cfg.LogLevel)
	slog.SetDefault(log)
	addr := ":" + cfg.Port

	dbConn, err := db.Open(cfg.DBDSN, log)
	if err != nil {
		log.Error("failed to connect db", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(dbConn); err != nil {
			log.Error("failed to close db", slog.String("error", err.Error()))
		}
	}()

	srv := server.New(dbConn, &cfg, log)
	go func() {
		if err := srv.Run(addr); err != nil && err != http.ErrServerClosed {
			log.Error("server stopped", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("failed to shutdown server", slog.String("error", err.Error()))
	}
}
