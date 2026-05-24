package middleware

import (
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"speaknow/internal/config"
	"speaknow/pkg/response"
)

type RateLimiter struct {
	global       *rate.Limiter
	userQPS      float64
	userBurst    int
	userLimiters sync.Map // userID -> *rate.Limiter
}

func NewRateLimiter(cfg config.RateLimitConfig) *RateLimiter {
	return &RateLimiter{
		global:    rate.NewLimiter(rate.Limit(cfg.GlobalQPS), cfg.GlobalBurst),
		userQPS:   cfg.UserQPS,
		userBurst: cfg.UserBurst,
	}
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !rl.global.Allow() {
			response.TooManyRequests(c, "global rate limit exceeded")
			c.Abort()
			return
		}

		userID := c.GetString("user_id")
		if userID == "" {
			userID = "anonymous"
		}

		limiter := rl.getUserLimiter(userID)
		if !limiter.Allow() {
			response.TooManyRequests(c, "user rate limit exceeded")
			c.Abort()
			return
		}

		c.Next()
	}
}

func (rl *RateLimiter) getUserLimiter(userID string) *rate.Limiter {
	if v, ok := rl.userLimiters.Load(userID); ok {
		return v.(*rate.Limiter)
	}
	limiter := rate.NewLimiter(rate.Limit(rl.userQPS), rl.userBurst)
	actual, _ := rl.userLimiters.LoadOrStore(userID, limiter)
	return actual.(*rate.Limiter)
}

func RetryAfterMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if c.Writer.Status() == 429 {
			c.Header("Retry-After", strconv.Itoa(1))
		}
	}
}
