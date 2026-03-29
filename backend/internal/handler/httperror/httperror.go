package httperror

import (
	"errors"
	"log/slog"
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
		if appErr.Status >= http.StatusInternalServerError {
			slog.Error("app error",
				"error", err.Error(),
				"status", appErr.Status,
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
			)
		}
		JSON(c, appErr.Status, appErr.Message, appErr.Code)
		return
	}

	slog.Error("internal server error",
		"error", err.Error(),
		"method", c.Request.Method,
		"path", c.Request.URL.Path,
		"message", fallbackMessage,
	)
	JSON(c, http.StatusInternalServerError, fallbackMessage, "internal_error")
}

func JSON(c *gin.Context, status int, message string, codes ...string) {
	resp := gin.H{"error": message}
	if len(codes) > 0 && codes[0] != "" {
		resp["code"] = codes[0]
	}
	c.JSON(status, resp)
}

func Unauthorized(c *gin.Context) {
	JSON(c, http.StatusUnauthorized, "unauthorized", "unauthorized")
}

func BadRequest(c *gin.Context, message string) {
	JSON(c, http.StatusBadRequest, message, "bad_request")
}
