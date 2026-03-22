package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r gin.IRoutes) {
	r.GET("/health", Health)
}

func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
