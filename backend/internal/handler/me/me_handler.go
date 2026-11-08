package me

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yoshioka0101/voiceblog/backend/internal/handler/httperror"
	"github.com/yoshioka0101/voiceblog/backend/internal/handler/middleware"
	presenter "github.com/yoshioka0101/voiceblog/backend/internal/presenter/user"
)

// AccountDeleter はユーザーと関連データを削除する usecase の契約。
type AccountDeleter interface {
	DeleteAccount(ctx context.Context, userID int64) error
}

type Handler struct {
	deleter AccountDeleter
}

func RegisterRoutes(r gin.IRoutes, auth gin.HandlerFunc, deleter AccountDeleter) {
	h := &Handler{deleter: deleter}
	r.GET("/me", auth, Me)
	r.DELETE("/me", auth, h.DeleteMe)
}

func Me(c *gin.Context) {
	u, ok := middleware.CurrentUser(c)
	if !ok {
		httperror.Unauthorized(c)
		return
	}
	c.JSON(http.StatusOK, presenter.User(u))
}

// DeleteMe は current user と関連データをすべて削除する。
func (h *Handler) DeleteMe(c *gin.Context) {
	u, ok := middleware.CurrentUser(c)
	if !ok {
		httperror.Unauthorized(c)
		return
	}

	if err := h.deleter.DeleteAccount(c.Request.Context(), u.ID); err != nil {
		httperror.FromError(c, err, "アカウントの削除に失敗しました")
		return
	}

	c.Status(http.StatusNoContent)
}
