package promptrunjob

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yoshioka0101/voiceblog/backend/internal/api"
	"github.com/yoshioka0101/voiceblog/backend/internal/handler/httperror"
	"github.com/yoshioka0101/voiceblog/backend/internal/handler/middleware"
	"github.com/yoshioka0101/voiceblog/backend/internal/handler/validation"
	"github.com/yoshioka0101/voiceblog/backend/internal/presenter/promptrunjob"
	"github.com/yoshioka0101/voiceblog/backend/internal/usecase/promptrunjob"
)

func RegisterProtectedRoutes(r gin.IRoutes, useCase *usecase.UseCase) {
	handler := &Handler{useCase: useCase}
	r.POST("/prompt-run-jobs", handler.CreatePromptRunJob)
	r.GET("/prompt-run-jobs/:id", handler.GetPromptRunJob)
}

type Handler struct {
	useCase *usecase.UseCase
}

func (h *Handler) CreatePromptRunJob(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		httperror.Unauthorized(c)
		return
	}

	var req api.CreatePromptRunJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.BadRequest(c, "invalid request")
		return
	}

	input, err := validation.CreatePromptRunJobInput(user.ID, req)
	if err != nil {
		httperror.FromError(c, err, "invalid request")
		return
	}

	value, err := h.useCase.CreatePromptRunJob(c.Request.Context(), input)
	if err != nil {
		httperror.FromError(c, err, "failed to create prompt run job")
		return
	}

	c.JSON(http.StatusCreated, presenter.PromptRunJob(value))
}

func (h *Handler) GetPromptRunJob(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		httperror.Unauthorized(c)
		return
	}

	jobID, err := validation.IDParam(c.Param("id"), "prompt run job id")
	if err != nil {
		httperror.FromError(c, err, "invalid prompt run job id")
		return
	}

	value, err := h.useCase.GetPromptRunJob(c.Request.Context(), user.ID, jobID)
	if err != nil {
		httperror.FromError(c, err, "failed to get prompt run job")
		return
	}

	c.JSON(http.StatusOK, presenter.PromptRunJob(value))
}
