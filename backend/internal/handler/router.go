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
	"github.com/yoshioka0101/voiceblog/backend/internal/ratelimit"
)

// Options はルーティングに渡す設定。
type Options struct {
	// GenerateLimit は AI 記事生成 API の 1 日あたりの上限。
	GenerateLimit middleware.DailyLimit
}

func RegisterRoutes(r *gin.Engine, c *di.Container, opts Options) {
	health.RegisterRoutes(r)
	auth := middleware.Auth(c.UseCases.Auth)
	me.RegisterRoutes(r, auth, c.UseCases.User)

	authGroup := r.Group("/")
	authGroup.Use(auth)
	generateLimit := middleware.DailyRateLimit(ratelimit.NewDailyCounter(), "generate", opts.GenerateLimit)
	articlehandler.RegisterProtectedRoutes(authGroup, c.UseCases.Article, generateLimit)
	prompthandler.RegisterProtectedRoutes(authGroup, c.UseCases.Prompt)
	transcriptionhandler.RegisterProtectedRoutes(authGroup, c.UseCases.Transcription)

	if c.UseCases.Integration != nil {
		integrationhandler.RegisterProtectedRoutes(authGroup, c.UseCases.Integration)
	}
	if c.UseCases.Publish != nil {
		publishhandler.RegisterProtectedRoutes(authGroup, c.UseCases.Publish)
	}
}
