package prompt

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yoshioka0101/voiceblog/backend/internal/api"
	"github.com/yoshioka0101/voiceblog/backend/internal/handler/httperror"
	"github.com/yoshioka0101/voiceblog/backend/internal/handler/middleware"
	"github.com/yoshioka0101/voiceblog/backend/internal/handler/validation"
	"github.com/yoshioka0101/voiceblog/backend/internal/presenter/prompt"
	"github.com/yoshioka0101/voiceblog/backend/internal/usecase/prompt"
)

func RegisterProtectedRoutes(r gin.IRoutes, useCase *usecase.UseCase) {
	handler := &Handler{useCase: useCase}
	r.GET("/prompts", handler.GetListPrompts)
	r.POST("/prompts", handler.CreatePrompt)
	r.PATCH("/prompts/:id", handler.UpdatePrompt)
	r.DELETE("/prompts/:id", handler.DeletePrompt)
}

type Handler struct {
	useCase *usecase.UseCase
}

func (h *Handler) GetListPrompts(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		httperror.Unauthorized(c)
		return
	}

	values, err := h.useCase.ListVisiblePromptsByUserID(c.Request.Context(), user.ID)
	if err != nil {
		httperror.FromError(c, err, "failed to list prompts")
		return
	}

	c.JSON(http.StatusOK, presenter.Prompts(values))
}

func (h *Handler) CreatePrompt(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		httperror.Unauthorized(c)
		return
	}

	var req api.CreatePromptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.BadRequest(c, "invalid request")
		return
	}

	input, err := validation.CreatePromptInput(user.ID, req)
	if err != nil {
		httperror.FromError(c, err, "invalid request")
		return
	}

	value, err := h.useCase.CreatePrompt(c.Request.Context(), input)
	if err != nil {
		httperror.FromError(c, err, "failed to create prompt")
		return
	}

	c.JSON(http.StatusCreated, presenter.Prompt(value))
}

func (h *Handler) UpdatePrompt(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		httperror.Unauthorized(c)
		return
	}

	promptID, err := validation.IDParam(c.Param("id"), "prompt id")
	if err != nil {
		httperror.FromError(c, err, "invalid prompt id")
		return
	}

	var req api.UpdatePromptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.BadRequest(c, "invalid request")
		return
	}

	input, err := validation.UpdatePromptInput(user.ID, promptID, req)
	if err != nil {
		httperror.FromError(c, err, "invalid request")
		return
	}

	value, err := h.useCase.UpdatePrompt(c.Request.Context(), input)
	if err != nil {
		httperror.FromError(c, err, "failed to update prompt")
		return
	}

	c.JSON(http.StatusOK, presenter.Prompt(value))
}

func (h *Handler) DeletePrompt(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		httperror.Unauthorized(c)
		return
	}

	promptID, err := validation.IDParam(c.Param("id"), "prompt id")
	if err != nil {
		httperror.FromError(c, err, "invalid prompt id")
		return
	}

	if err := h.useCase.DeletePrompt(c.Request.Context(), user.ID, promptID); err != nil {
		httperror.FromError(c, err, "failed to delete prompt")
		return
	}

	c.Status(http.StatusNoContent)
}
