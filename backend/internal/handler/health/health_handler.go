package health

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yoshioka0101/voiceblog/backend/internal/api"
)

func RegisterRoutes(r gin.IRoutes) {
	r.GET("/health", Health)
}

func Health(c *gin.Context) {
	c.JSON(http.StatusOK, api.HealthResponse{Status: "ok"})
}
