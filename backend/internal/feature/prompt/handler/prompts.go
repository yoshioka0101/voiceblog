package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	promptdomain "github.com/yoshioka0101/voiceblog/backend/internal/feature/prompt/domain"
	promptusecase "github.com/yoshioka0101/voiceblog/backend/internal/feature/prompt/usecase"
	"github.com/yoshioka0101/voiceblog/backend/internal/handler/middleware"
)

type Handler struct {
	useCase *promptusecase.UseCase
}

type createRequest struct {
	Name     string `json:"name" binding:"required"`
	Body     string `json:"body" binding:"required"`
	IsActive *bool  `json:"is_active"`
}

type updateRequest struct {
	Name     *string `json:"name"`
	Body     *string `json:"body"`
	IsActive *bool   `json:"is_active"`
}

type response struct {
	ID        int64  `json:"id"`
	UserID    *int64 `json:"user_id"`
	Name      string `json:"name"`
	Body      string `json:"body"`
	IsActive  bool   `json:"is_active"`
	IsDefault bool   `json:"is_default"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func NewHandler(useCase *promptusecase.UseCase) *Handler {
	return &Handler{useCase: useCase}
}

func (h *Handler) RegisterProtectedRoutes(r gin.IRoutes) {
	r.GET("/prompts", h.List)
	r.POST("/prompts", h.Create)
	r.PATCH("/prompts/:id", h.Update)
	r.DELETE("/prompts/:id", h.Delete)
}

func (h *Handler) List(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	prompts, err := h.useCase.ListVisibleByUserID(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list prompts"})
		return
	}

	responses := make([]response, 0, len(prompts))
	for _, prompt := range prompts {
		responses = append(responses, toResponse(prompt))
	}

	c.JSON(http.StatusOK, responses)
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

	name := strings.TrimSpace(req.Name)
	body := strings.TrimSpace(req.Body)
	if name == "" || body == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name and body are required"})
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	prompt, err := h.useCase.Create(c.Request.Context(), promptdomain.CreateParams{
		UserID:   user.ID,
		Name:     name,
		Body:     body,
		IsActive: isActive,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create prompt"})
		return
	}

	c.JSON(http.StatusCreated, toResponse(prompt))
}

func (h *Handler) Update(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	promptID, ok := parsePromptID(c)
	if !ok {
		return
	}

	var req updateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if req.Name == nil && req.Body == nil && req.IsActive == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one field is required"})
		return
	}

	patch := promptdomain.PatchParams{IsActive: req.IsActive}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name must not be blank"})
			return
		}
		patch.Name = &name
	}
	if req.Body != nil {
		body := strings.TrimSpace(*req.Body)
		if body == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "body must not be blank"})
			return
		}
		patch.Body = &body
	}

	prompt, err := h.useCase.Update(c.Request.Context(), user.ID, promptID, patch)
	if err != nil {
		switch {
		case errors.Is(err, promptdomain.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, promptdomain.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "prompt not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update prompt"})
		}
		return
	}

	c.JSON(http.StatusOK, toResponse(prompt))
}

func (h *Handler) Delete(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	promptID, ok := parsePromptID(c)
	if !ok {
		return
	}

	if err := h.useCase.Delete(c.Request.Context(), user.ID, promptID); err != nil {
		switch {
		case errors.Is(err, promptdomain.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, promptdomain.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "prompt not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete prompt"})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func parsePromptID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid prompt id"})
		return 0, false
	}
	return id, true
}

func toResponse(prompt *promptdomain.Prompt) response {
	return response{
		ID:        prompt.ID,
		UserID:    prompt.UserID,
		Name:      prompt.Name,
		Body:      prompt.Body,
		IsActive:  prompt.IsActive,
		IsDefault: prompt.IsDefault,
		CreatedAt: prompt.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: prompt.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func RegisterProtectedRoutes(r gin.IRoutes, useCase *promptusecase.UseCase) {
	NewHandler(useCase).RegisterProtectedRoutes(r)
}
