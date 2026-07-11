package transcription

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yoshioka0101/voiceblog/backend/internal/api"
	"github.com/yoshioka0101/voiceblog/backend/internal/handler/httperror"
	"github.com/yoshioka0101/voiceblog/backend/internal/handler/middleware"
	"github.com/yoshioka0101/voiceblog/backend/internal/handler/validation"
	"github.com/yoshioka0101/voiceblog/backend/internal/presenter/transcription"
	"github.com/yoshioka0101/voiceblog/backend/internal/usecase/transcription"
)

func RegisterProtectedRoutes(r gin.IRoutes, useCase *usecase.UseCase) {
	handler := &Handler{useCase: useCase}
	r.POST("/transcriptions", handler.CreateTranscription)
}

type Handler struct {
	useCase *usecase.UseCase
}

func (h *Handler) CreateTranscription(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		httperror.Unauthorized(c)
		return
	}

	var req api.CreateTranscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.BadRequest(c, "invalid request")
		return
	}

	input, err := validation.CreateTranscriptionInput(user.ID, req)
	if err != nil {
		httperror.FromError(c, err, "invalid request")
		return
	}

	value, err := h.useCase.CreateTranscription(c.Request.Context(), input)
	if err != nil {
		httperror.FromError(c, err, "failed to create transcription")
		return
	}

	response, err := presenter.Transcription(value)
	if err != nil {
		httperror.FromError(c, err, "failed to present transcription")
		return
	}

	c.JSON(http.StatusCreated, response)
}
