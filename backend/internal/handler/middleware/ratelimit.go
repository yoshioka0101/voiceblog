package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/yoshioka0101/voiceblog/backend/internal/handler/httperror"
	"github.com/yoshioka0101/voiceblog/backend/internal/ratelimit"
)

// DailyLimit は 1 日あたりの上限回数。0 以下は無制限。
type DailyLimit struct {
	PerUser int
	Global  int
}

// DailyRateLimit は認証済みユーザー単位と全体で 1 日あたりの呼び出し回数を制限する。
// Auth ミドルウェアの後に置くこと。name は同じ counter を複数エンドポイントで共有する時のキー接頭辞。
func DailyRateLimit(counter *ratelimit.DailyCounter, name string, limit DailyLimit) gin.HandlerFunc {
	return func(c *gin.Context) {
		u, ok := CurrentUser(c)
		if !ok {
			httperror.Unauthorized(c)
			c.Abort()
			return
		}

		allowed := counter.Allow(
			ratelimit.Rule{Key: name + ":user:" + strconv.FormatInt(u.ID, 10), Limit: limit.PerUser},
			ratelimit.Rule{Key: name + ":global", Limit: limit.Global},
		)
		if !allowed {
			httperror.JSON(c, http.StatusTooManyRequests, "本日の利用上限に達しました。明日もう一度お試しください。", "rate_limited")
			c.Abort()
			return
		}

		c.Next()
	}
}
