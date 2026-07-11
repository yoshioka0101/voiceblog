package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/yoshioka0101/voiceblog/backend/internal/di"
	articlehandler "github.com/yoshioka0101/voiceblog/backend/internal/handler/article"
	"github.com/yoshioka0101/voiceblog/backend/internal/handler/health"
	integrationhandler "github.com/yoshioka0101/voiceblog/backend/internal/handler/integration"
	"github.com/yoshioka0101/voiceblog/backend/internal/handler/me"
	"github.com/yoshioka0101/voiceblog/backend/internal/handler/middleware"
	prompthandler "github.com/yoshioka0101/voiceblog/backend/internal/handler/prompt"
	publishhandler "github.com/yoshioka0101/voiceblog/backend/internal/handler/publish"
	transcriptionhandler "github.com/yoshioka0101/voiceblog/backend/internal/handler/transcription"
)

func RegisterRoutes(r *gin.Engine, c *di.Container) {
	health.RegisterRoutes(r)
	auth := middleware.Auth(c.UseCases.Auth)
	me.RegisterRoutes(r, auth)

	authGroup := r.Group("/")
	authGroup.Use(auth)
	articlehandler.RegisterProtectedRoutes(authGroup, c.UseCases.Article)
	prompthandler.RegisterProtectedRoutes(authGroup, c.UseCases.Prompt)
	transcriptionhandler.RegisterProtectedRoutes(authGroup, c.UseCases.Transcription)

	if c.UseCases.Integration != nil {
		integrationhandler.RegisterProtectedRoutes(authGroup, c.UseCases.Integration)
	}
	if c.UseCases.Publish != nil {
		publishhandler.RegisterProtectedRoutes(authGroup, c.UseCases.Publish)
	}
}
