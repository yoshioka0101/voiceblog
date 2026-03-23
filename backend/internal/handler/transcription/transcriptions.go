package transcription

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/yoshioka0101/voiceblog/backend/internal/api"
	"github.com/yoshioka0101/voiceblog/backend/internal/handler/httperror"
	"github.com/yoshioka0101/voiceblog/backend/internal/handler/middleware"
	"github.com/yoshioka0101/voiceblog/backend/internal/presenter/transcription"
	"github.com/yoshioka0101/voiceblog/backend/internal/usecase/transcription"
)

func RegisterProtectedRoutes(r gin.IRoutes, useCase *usecase.UseCase) {
	handler := &Handler{useCase: useCase}
	r.POST("/transcriptions", handler.CreateTranscription)
	r.GET("/transcriptions", handler.ListTranscriptions)
	r.GET("/transcriptions/:id", handler.GetTranscription)
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

	segmentsJSON, err := json.Marshal(req.SegmentsJson)
	if err != nil {
		httperror.BadRequest(c, "segments_json must be valid JSON")
		return
	}

	value, err := h.useCase.CreateTranscription(c.Request.Context(), usecase.CreateTranscriptionInput{
		UserID:       user.ID,
		FullText:     req.FullText,
		SegmentsJSON: segmentsJSON,
	})
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

func (h *Handler) ListTranscriptions(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		httperror.Unauthorized(c)
		return
	}

	values, err := h.useCase.ListTranscriptionsByUserID(c.Request.Context(), user.ID)
	if err != nil {
		httperror.FromError(c, err, "failed to list transcriptions")
		return
	}

	response, err := presenter.Transcriptions(values)
	if err != nil {
		httperror.FromError(c, err, "failed to present transcriptions")
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) GetTranscription(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		httperror.Unauthorized(c)
		return
	}

	transcriptionID, err := parseIDParam(c.Param("id"))
	if err != nil {
		httperror.BadRequest(c, "invalid transcription id")
		return
	}

	value, err := h.useCase.GetTranscription(c.Request.Context(), user.ID, transcriptionID)
	if err != nil {
		httperror.FromError(c, err, "failed to get transcription")
		return
	}

	response, err := presenter.Transcription(value)
	if err != nil {
		httperror.FromError(c, err, "failed to present transcription")
		return
	}

	c.JSON(http.StatusOK, response)
}

func parseIDParam(raw string) (int64, error) {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, err
	}
	if id <= 0 {
		return 0, errors.New("id must be positive")
	}
	return id, nil
}
