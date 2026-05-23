package middleware

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"

	"speaknow/internal/config"
	"speaknow/pkg/response"
)

type RateLimiter struct {
	global   *rate.Limiter
	userQPS  float64
	userBurst int
	userLimiters sync.Map // userID -> *rate.Limiter
	redis    *redis.Client
	globalQPS float64
}

func NewRateLimiter(cfg config.RateLimitConfig, redisClient *redis.Client) *RateLimiter {
	return &RateLimiter{
		global:    rate.NewLimiter(rate.Limit(cfg.GlobalQPS), cfg.GlobalBurst),
		userQPS:   cfg.UserQPS,
		userBurst: cfg.UserBurst,
		redis:     redisClient,
		globalQPS: cfg.GlobalQPS,
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

		if rl.redis != nil {
			if err := rl.checkGlobalRedis(c.Request.Context()); err != nil {
				response.TooManyRequests(c, err.Error())
				c.Abort()
				return
			}
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

func (rl *RateLimiter) checkGlobalRedis(ctx context.Context) error {
	key := "ratelimit:global:" + time.Now().Format("20060102150405")
	count, err := rl.redis.Incr(ctx, key).Result()
	if err != nil {
		return nil // Redis 不可用时降级为仅本地限流
	}
	if count == 1 {
		rl.redis.Expire(ctx, key, 2*time.Second)
	}
	if float64(count) > rl.globalQPS*2 {
		return fmt.Errorf("distributed rate limit exceeded")
	}
	return nil
}

func RetryAfterMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if c.Writer.Status() == 429 {
			c.Header("Retry-After", strconv.Itoa(1))
		}
	}
}
