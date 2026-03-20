package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/yoshioka0101/voiceblog/backend/internal/di"
	"github.com/yoshioka0101/voiceblog/backend/internal/handler/health"
	"github.com/yoshioka0101/voiceblog/backend/internal/handler/me"
	"github.com/yoshioka0101/voiceblog/backend/internal/handler/middleware"
)

func RegisterRoutes(r *gin.Engine, c *di.Container) {
	r.GET("/health", health.Health)

	authGroup := r.Group("/")
	authGroup.Use(middleware.Auth(c.UseCases.Auth))
	authGroup.GET("/me", me.Me)
}
