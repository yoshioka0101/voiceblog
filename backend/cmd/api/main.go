package main

import (
	"log/slog"
	"os"

	"github.com/yoshioka0101/voiceblog/backend/internal/config"
	"github.com/yoshioka0101/voiceblog/backend/internal/logger"
	"github.com/yoshioka0101/voiceblog/backend/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	log := logger.New(cfg.LogLevel)
	addr := ":" + cfg.Port

	if err := server.New(log).Run(addr); err != nil {
		log.Error("server stopped", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
