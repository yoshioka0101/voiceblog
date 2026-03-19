package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/yoshioka0101/voiceblog/backend/internal/health"
)

type Server struct {
	engine *gin.Engine
	http   *http.Server
}

func New(log *slog.Logger) *Server {
	router := gin.New()
	router.Use(requestLogger(log), gin.Recovery())

	health.RegisterRoutes(router)

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

func requestLogger(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		log.Info("request",
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.Int("status", status),
			slog.Duration("latency", latency),
		)
	}
}
