package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetHealth(c *gin.Context) {
	// validation

	// usecase

	// response
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

func RegisterHealthRoutes(r *gin.Engine) {
	r.GET("/health", GetHealth)
}
