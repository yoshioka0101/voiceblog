package me

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yoshioka0101/voiceblog/backend/internal/api"
	"github.com/yoshioka0101/voiceblog/backend/internal/handler/httperror"
	"github.com/yoshioka0101/voiceblog/backend/internal/handler/middleware"
)

func RegisterRoutes(r gin.IRoutes, auth gin.HandlerFunc) {
	r.GET("/me", auth, Me)
}

func Me(c *gin.Context) {
	u, ok := middleware.CurrentUser(c)
	if !ok {
		httperror.Unauthorized(c)
		return
	}
	c.JSON(http.StatusOK, api.UserResponse{
		Id:           u.ID,
		Email:        stringOrNil(u.Email),
		Name:         stringOrNil(u.Name),
		AuthProvider: u.AuthProvider,
	})
}

func stringOrNil(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}
