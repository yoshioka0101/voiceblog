package handler

import "github.com/gin-gonic/gin"

func RegisterAllRoutes(r *gin.Engine) {
	RegisterHealthRoutes(r)
}
