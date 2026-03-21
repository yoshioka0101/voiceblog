package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/yoshioka0101/voiceblog/backend/internal/di"
	"github.com/yoshioka0101/voiceblog/backend/internal/feature/health"
	transcriptionfeature "github.com/yoshioka0101/voiceblog/backend/internal/feature/transcription"
	userfeature "github.com/yoshioka0101/voiceblog/backend/internal/feature/user"
	"github.com/yoshioka0101/voiceblog/backend/internal/handler/middleware"
)

func RegisterRoutes(r *gin.Engine, c *di.Container) {
	health.RegisterRoutes(r)

	authGroup := r.Group("/")
	authGroup.Use(middleware.Auth(c.Auth.UseCase))
	userfeature.RegisterProtectedRoutes(authGroup)
	transcriptionfeature.RegisterProtectedRoutes(authGroup, c.Transcription)
}
