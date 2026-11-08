package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	userdomain "github.com/yoshioka0101/voiceblog/backend/internal/entity/user"
	"github.com/yoshioka0101/voiceblog/backend/internal/ratelimit"
)

func newRateLimitedEngine(counter *ratelimit.DailyCounter, limit DailyLimit) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/generate",
		func(c *gin.Context) {
			id := int64(1)
			if c.GetHeader("X-User") == "2" {
				id = 2
			}
			c.Set(currentUserKey, &userdomain.User{ID: id})
		},
		DailyRateLimit(counter, "generate", limit),
		func(c *gin.Context) { c.Status(http.StatusOK) },
	)
	return r
}

func doGenerate(r *gin.Engine, user string) int {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/generate", nil)
	req.Header.Set("X-User", user)
	r.ServeHTTP(w, req)
	return w.Code
}

func TestDailyRateLimit_PerUser(t *testing.T) {
	r := newRateLimitedEngine(ratelimit.NewDailyCounter(), DailyLimit{PerUser: 2})

	for i := range 2 {
		if code := doGenerate(r, "1"); code != http.StatusOK {
			t.Fatalf("call %d: status = %d, want 200", i+1, code)
		}
	}
	if code := doGenerate(r, "1"); code != http.StatusTooManyRequests {
		t.Fatalf("3rd call: status = %d, want 429", code)
	}
	if code := doGenerate(r, "2"); code != http.StatusOK {
		t.Fatalf("other user: status = %d, want 200", code)
	}
}

func TestDailyRateLimit_Global(t *testing.T) {
	r := newRateLimitedEngine(ratelimit.NewDailyCounter(), DailyLimit{PerUser: 10, Global: 2})

	doGenerate(r, "1")
	doGenerate(r, "2")
	if code := doGenerate(r, "2"); code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429 after global limit", code)
	}
}

func TestDailyRateLimit_RequiresUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/generate", DailyRateLimit(ratelimit.NewDailyCounter(), "generate", DailyLimit{PerUser: 1}),
		func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/generate", nil))
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", w.Code)
	}
}
