package integration

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yoshioka0101/voiceblog/backend/internal/handler/httperror"
	"github.com/yoshioka0101/voiceblog/backend/internal/handler/middleware"
	usecase "github.com/yoshioka0101/voiceblog/backend/internal/usecase/integration"
)

func RegisterProtectedRoutes(r gin.IRoutes, uc *usecase.UseCase) {
	h := &Handler{uc: uc}
	r.GET("/integrations", h.GetIntegrations)
	r.PUT("/integrations/:provider/token", h.StoreToken)
	r.DELETE("/integrations/:provider/token", h.DeleteToken)
}

type Handler struct {
	uc *usecase.UseCase
}

type integrationResponse struct {
	Provider  string `json:"provider"`
	Connected bool   `json:"connected"`
}

type storeTokenRequest struct {
	Token string `json:"token" binding:"required"`
}

func (h *Handler) GetIntegrations(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		httperror.Unauthorized(c)
		return
	}

	statuses, err := h.uc.GetIntegrations(c.Request.Context(), user.ID)
	if err != nil {
		httperror.FromError(c, err, "failed to get integrations")
		return
	}

	resp := make([]integrationResponse, 0, len(statuses))
	for _, s := range statuses {
		resp = append(resp, integrationResponse{
			Provider:  s.Provider,
			Connected: s.Connected,
		})
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) StoreToken(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		httperror.Unauthorized(c)
		return
	}

	provider := c.Param("provider")

	var req storeTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.BadRequest(c, "invalid request")
		return
	}

	if err := h.uc.StoreToken(c.Request.Context(), user.ID, provider, req.Token); err != nil {
		httperror.FromError(c, err, "failed to store token")
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) DeleteToken(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		httperror.Unauthorized(c)
		return
	}

	provider := c.Param("provider")

	if err := h.uc.DeleteToken(c.Request.Context(), user.ID, provider); err != nil {
		httperror.FromError(c, err, "failed to delete token")
		return
	}

	c.Status(http.StatusNoContent)
}
