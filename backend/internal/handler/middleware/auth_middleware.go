package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	domainUser "github.com/yoshioka0101/voiceblog/backend/internal/domain/user"
	authUseCase "github.com/yoshioka0101/voiceblog/backend/internal/usecase/auth"
)

const currentUserKey = "currentUser"

func Auth(authUC *authUseCase.UseCase) gin.HandlerFunc {
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

func CurrentUser(c *gin.Context) (*domainUser.User, bool) {
	v, ok := c.Get(currentUserKey)
	if !ok {
		return nil, false
	}
	u, ok := v.(*domainUser.User)
	return u, ok
}
