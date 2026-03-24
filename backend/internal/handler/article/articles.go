package article

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/yoshioka0101/voiceblog/backend/internal/api"
	"github.com/yoshioka0101/voiceblog/backend/internal/handler/httperror"
	"github.com/yoshioka0101/voiceblog/backend/internal/handler/middleware"
	"github.com/yoshioka0101/voiceblog/backend/internal/presenter/article"
	"github.com/yoshioka0101/voiceblog/backend/internal/usecase/article"
)

func RegisterProtectedRoutes(r gin.IRoutes, useCase *usecase.UseCase) {
	handler := &Handler{useCase: useCase}
	r.GET("/articles", handler.GetListArticles)
	r.POST("/articles", handler.CreateArticle)
	r.GET("/articles/:id", handler.GetArticle)
	r.PATCH("/articles/:id", handler.UpdateArticle)
	r.DELETE("/articles/:id", handler.DeleteArticle)
}

type Handler struct {
	useCase *usecase.UseCase
}

func (h *Handler) GetListArticles(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		httperror.Unauthorized(c)
		return
	}

	values, err := h.useCase.ListArticlesByUserID(c.Request.Context(), user.ID)
	if err != nil {
		httperror.FromError(c, err, "failed to list articles")
		return
	}

	c.JSON(http.StatusOK, presenter.Articles(values))
}

func (h *Handler) CreateArticle(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		httperror.Unauthorized(c)
		return
	}

	var req api.CreateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.BadRequest(c, "invalid request")
		return
	}

	value, err := h.useCase.CreateArticle(c.Request.Context(), usecase.CreateArticleInput{
		UserID:         user.ID,
		PromptRunJobID: req.PromptRunJobId,
		Title:          req.Title,
		Content:        req.Content,
	})
	if err != nil {
		httperror.FromError(c, err, "failed to create article")
		return
	}

	c.JSON(http.StatusCreated, presenter.Article(value))
}

func (h *Handler) GetArticle(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		httperror.Unauthorized(c)
		return
	}

	articleID, err := parseIDParam(c.Param("id"))
	if err != nil {
		httperror.BadRequest(c, "invalid article id")
		return
	}

	value, err := h.useCase.GetArticle(c.Request.Context(), user.ID, articleID)
	if err != nil {
		httperror.FromError(c, err, "failed to get article")
		return
	}

	c.JSON(http.StatusOK, presenter.Article(value))
}

func (h *Handler) UpdateArticle(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		httperror.Unauthorized(c)
		return
	}

	articleID, err := parseIDParam(c.Param("id"))
	if err != nil {
		httperror.BadRequest(c, "invalid article id")
		return
	}

	var req api.UpdateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.BadRequest(c, "invalid request")
		return
	}

	value, err := h.useCase.UpdateArticle(c.Request.Context(), usecase.UpdateArticleInput{
		UserID:    user.ID,
		ArticleID: articleID,
		Title:     req.Title,
		Content:   req.Content,
	})
	if err != nil {
		httperror.FromError(c, err, "failed to update article")
		return
	}

	c.JSON(http.StatusOK, presenter.Article(value))
}

func (h *Handler) DeleteArticle(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		httperror.Unauthorized(c)
		return
	}

	articleID, err := parseIDParam(c.Param("id"))
	if err != nil {
		httperror.BadRequest(c, "invalid article id")
		return
	}

	if err := h.useCase.DeleteArticle(c.Request.Context(), user.ID, articleID); err != nil {
		httperror.FromError(c, err, "failed to delete article")
		return
	}

	c.Status(http.StatusNoContent)
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
