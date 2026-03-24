package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/yoshioka0101/voiceblog/backend/internal/di"
	articlehandler "github.com/yoshioka0101/voiceblog/backend/internal/handler/article"
	"github.com/yoshioka0101/voiceblog/backend/internal/handler/health"
	"github.com/yoshioka0101/voiceblog/backend/internal/handler/me"
	"github.com/yoshioka0101/voiceblog/backend/internal/handler/middleware"
	prompthandler "github.com/yoshioka0101/voiceblog/backend/internal/handler/prompt"
	promptrunjobhandler "github.com/yoshioka0101/voiceblog/backend/internal/handler/promptrunjob"
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
	promptrunjobhandler.RegisterProtectedRoutes(authGroup, c.UseCases.PromptRunJob)
	transcriptionhandler.RegisterProtectedRoutes(authGroup, c.UseCases.Transcription)
}
