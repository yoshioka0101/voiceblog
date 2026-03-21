package transcription

import (
	"github.com/gin-gonic/gin"

	featurehandler "github.com/yoshioka0101/voiceblog/backend/internal/feature/transcription/handler"
)

func RegisterProtectedRoutes(r gin.IRoutes, feature *Feature) {
	featurehandler.RegisterProtectedRoutes(r, feature.UseCase)
}
