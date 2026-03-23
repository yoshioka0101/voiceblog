package httperror

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yoshioka0101/voiceblog/backend/internal/apperr"
)

type Rule struct {
	Target  error
	Status  int
	Message string
}

func Handle(c *gin.Context, err error, fallback Rule, rules ...Rule) {
	for _, rule := range rules {
		if errors.Is(err, rule.Target) {
			JSON(c, rule.Status, rule.Message)
			return
		}
	}

	JSON(c, fallback.Status, fallback.Message)
}

func FromError(c *gin.Context, err error, fallbackMessage string) {
	var appErr *apperr.AppError
	if errors.As(err, &appErr) {
		JSON(c, appErr.Status, appErr.Message)
		return
	}

	JSON(c, http.StatusInternalServerError, fallbackMessage)
}

func JSON(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message})
}

func Unauthorized(c *gin.Context) {
	JSON(c, http.StatusUnauthorized, "unauthorized")
}

func BadRequest(c *gin.Context, message string) {
	JSON(c, http.StatusBadRequest, message)
}
