package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	transcriptiondomain "github.com/yoshioka0101/voiceblog/backend/internal/feature/transcription/domain"
	transcriptionusecase "github.com/yoshioka0101/voiceblog/backend/internal/feature/transcription/usecase"
	"github.com/yoshioka0101/voiceblog/backend/internal/handler/middleware"
)

type Handler struct {
	useCase *transcriptionusecase.UseCase
}

type createRequest struct {
	FullText     string          `json:"full_text" binding:"required"`
	SegmentsJSON json.RawMessage `json:"segments_json"`
}

type response struct {
	ID           int64           `json:"id"`
	UserID       int64           `json:"user_id"`
	FullText     string          `json:"full_text"`
	SegmentsJSON json.RawMessage `json:"segments_json"`
	CreatedAt    string          `json:"created_at"`
	UpdatedAt    string          `json:"updated_at"`
}

func NewHandler(useCase *transcriptionusecase.UseCase) *Handler {
	return &Handler{useCase: useCase}
}

func (h *Handler) RegisterProtectedRoutes(r gin.IRoutes) {
	r.POST("/transcriptions", h.Create)
	r.GET("/transcriptions", h.List)
	r.GET("/transcriptions/:id", h.Get)
}

func (h *Handler) Create(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if len(req.SegmentsJSON) == 0 || !json.Valid(req.SegmentsJSON) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "segments_json must be valid JSON"})
		return
	}

	transcription, err := h.useCase.Create(c.Request.Context(), transcriptiondomain.CreateParams{
		UserID:       user.ID,
		FullText:     req.FullText,
		SegmentsJSON: req.SegmentsJSON,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create transcription"})
		return
	}

	c.JSON(http.StatusCreated, toResponse(transcription))
}

func (h *Handler) List(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	transcriptions, err := h.useCase.ListByUserID(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list transcriptions"})
		return
	}

	responses := make([]response, 0, len(transcriptions))
	for _, transcription := range transcriptions {
		responses = append(responses, toResponse(transcription))
	}

	c.JSON(http.StatusOK, responses)
}

func (h *Handler) Get(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transcription id"})
		return
	}

	transcription, err := h.useCase.FindByID(c.Request.Context(), user.ID, id)
	if err != nil {
		switch {
		case errors.Is(err, transcriptiondomain.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, transcriptiondomain.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "transcription not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get transcription"})
		}
		return
	}

	c.JSON(http.StatusOK, toResponse(transcription))
}

func toResponse(transcription *transcriptiondomain.Transcription) response {
	return response{
		ID:           transcription.ID,
		UserID:       transcription.UserID,
		FullText:     transcription.FullText,
		SegmentsJSON: transcription.SegmentsJSON,
		CreatedAt:    transcription.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:    transcription.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func RegisterProtectedRoutes(r gin.IRoutes, useCase *transcriptionusecase.UseCase) {
	NewHandler(useCase).RegisterProtectedRoutes(r)
}
