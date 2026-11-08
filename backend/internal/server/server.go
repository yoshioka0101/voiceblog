package server

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yoshioka0101/voiceblog/backend/internal/config"
	"github.com/yoshioka0101/voiceblog/backend/internal/di"
	"github.com/yoshioka0101/voiceblog/backend/internal/handler"
	"github.com/yoshioka0101/voiceblog/backend/internal/handler/middleware"
)

type Server struct {
	engine *gin.Engine
	http   *http.Server
}

func New(db *sql.DB, cfg *config.Config, log *slog.Logger) (*Server, error) {
	configureGinLogging(log)
	router := gin.New()
	router.Use(requestLogger(log), recoveryLogger(log))

	container, err := di.New(db, cfg.GoogleClientID, cfg.GeminiAPIKey, cfg.GeminiModel, cfg.TokenEncryptionKey)
	if err != nil {
		return nil, err
	}
	handler.RegisterRoutes(router, container, handler.Options{
		GenerateLimit: middleware.DailyLimit{
			PerUser: cfg.GenerateDailyLimitPerUser,
			Global:  cfg.GenerateDailyLimitGlobal,
		},
	})

	return &Server{engine: router}, nil
}

func (s *Server) Run(addr string) error {
	s.http = &http.Server{
		Addr:    addr,
		Handler: s.engine,
	}
	return s.http.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.http == nil {
		return nil
	}
	return s.http.Shutdown(ctx)
}
