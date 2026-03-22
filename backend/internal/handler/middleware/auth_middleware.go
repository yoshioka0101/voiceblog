package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	userdomain "github.com/yoshioka0101/voiceblog/backend/internal/entity/user"
	authusecase "github.com/yoshioka0101/voiceblog/backend/internal/usecase/auth"
)

const currentUserKey = "currentUser"

func Auth(authUC *authusecase.UseCase) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}
		tokenStr := strings.TrimPrefix(header, "Bearer ")

		u, err := authUC.Authenticate(c.Request.Context(), tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Set(currentUserKey, u)
		c.Next()
	}
}

func CurrentUser(c *gin.Context) (*userdomain.User, bool) {
	v, ok := c.Get(currentUserKey)
	if !ok {
		return nil, false
	}
	u, ok := v.(*userdomain.User)
	return u, ok
}
