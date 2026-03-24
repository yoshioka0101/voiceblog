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
)

type Server struct {
	engine *gin.Engine
	http   *http.Server
}

func New(db *sql.DB, cfg *config.Config, log *slog.Logger) *Server {
	configureGinLogging(log)
	router := gin.New()
	router.Use(requestLogger(log), recoveryLogger(log))

	container := di.New(db, cfg.GoogleClientID, cfg.GeminiAPIKey)
	handler.RegisterRoutes(router, container)

	return &Server{engine: router}
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
