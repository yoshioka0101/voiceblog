package me

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yoshioka0101/voiceblog/backend/internal/handler/middleware"
)

func RegisterRoutes(r gin.IRoutes, auth gin.HandlerFunc) {
	r.GET("/me", auth, Me)
}

func Me(c *gin.Context) {
	u, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":            u.ID,
		"email":         u.Email,
		"name":          u.Name,
		"auth_provider": u.AuthProvider,
	})
}
