package publish

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/yoshioka0101/voiceblog/backend/internal/handler/httperror"
	"github.com/yoshioka0101/voiceblog/backend/internal/handler/middleware"
	usecase "github.com/yoshioka0101/voiceblog/backend/internal/usecase/publish"
)

func RegisterProtectedRoutes(r gin.IRoutes, uc *usecase.UseCase) {
	h := &Handler{uc: uc}
	r.POST("/articles/:id/publish", h.PublishArticle)
	r.GET("/articles/:id/share-targets", h.GetShareTargets)
}

type Handler struct {
	uc *usecase.UseCase
}

type publishRequest struct {
	Provider string `json:"provider" binding:"required"`
}

type shareTargetResponse struct {
	ID          int64      `json:"id"`
	ArticleID   int64      `json:"article_id"`
	Provider    string     `json:"provider"`
	ExternalID  string     `json:"external_id"`
	ExternalURL string     `json:"external_url"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (h *Handler) PublishArticle(c *gin.Context) {
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

	var req publishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.BadRequest(c, "invalid request")
		return
	}

	target, err := h.uc.PublishArticle(c.Request.Context(), user.ID, articleID, req.Provider)
	if err != nil {
		httperror.FromError(c, err, "failed to publish article")
		return
	}

	c.JSON(http.StatusCreated, shareTargetResponse{
		ID:          target.ID,
		ArticleID:   target.ArticleID,
		Provider:    target.Provider,
		ExternalID:  target.ExternalID,
		ExternalURL: target.ExternalURL,
		PublishedAt: target.PublishedAt,
		CreatedAt:   target.CreatedAt,
		UpdatedAt:   target.UpdatedAt,
	})
}

func (h *Handler) GetShareTargets(c *gin.Context) {
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

	targets, err := h.uc.GetShareTargets(c.Request.Context(), user.ID, articleID)
	if err != nil {
		httperror.FromError(c, err, "failed to get share targets")
		return
	}

	resp := make([]shareTargetResponse, 0, len(targets))
	for _, t := range targets {
		resp = append(resp, shareTargetResponse{
			ID:          t.ID,
			ArticleID:   t.ArticleID,
			Provider:    t.Provider,
			ExternalID:  t.ExternalID,
			ExternalURL: t.ExternalURL,
			PublishedAt: t.PublishedAt,
			CreatedAt:   t.CreatedAt,
			UpdatedAt:   t.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, resp)
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
