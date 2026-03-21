package user

import (
	"github.com/gin-gonic/gin"

	featurehandler "github.com/yoshioka0101/voiceblog/backend/internal/feature/user/handler"
)

func RegisterProtectedRoutes(r gin.IRoutes) {
	featurehandler.RegisterProtectedRoutes(r)
}
