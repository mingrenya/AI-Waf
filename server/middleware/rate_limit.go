package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mingrenya/AI-Waf/server/model"
)

// rateLimiter 简单的内存限流器
type rateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	limit    int           // 最大请求数
	window   time.Duration // 时间窗口
}

type visitor struct {
	lastSeen time.Time
	count    int
}

// newRateLimiter 创建新的限流器
func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		visitors: make(map[string]*visitor),
		limit:    limit,
		window:   window,
	}

	// 启动清理goroutine，定期清理过期的访客记录
	go rl.cleanupVisitors()

	return rl
}

// allow 检查是否允许访问
func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	v, exists := rl.visitors[key]

	if !exists {
		// 首次访问
		rl.visitors[key] = &visitor{
			lastSeen: now,
			count:    1,
		}
		return true
	}

	// 检查时间窗口是否已过
	if now.Sub(v.lastSeen) > rl.window {
		// 时间窗口已过，重置计数
		v.lastSeen = now
		v.count = 1
		return true
	}

	// 在时间窗口内
	if v.count >= rl.limit {
		// 超过限制
		return false
	}

	// 增加计数
	v.count++
	return true
}

// cleanupVisitors 定期清理过期的访客记录
func (rl *rateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, v := range rl.visitors {
			if now.Sub(v.lastSeen) > rl.window {
				delete(rl.visitors, key)
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimit 返回限流中间件
// limit: 时间窗口内允许的最大请求数
// window: 时间窗口
// 示例: RateLimit(5, time.Minute) 表示每分钟最多5次请求
func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	limiter := newRateLimiter(limit, window)

	return func(c *gin.Context) {
		// 使用IP作为限流key
		key := c.ClientIP()

		if !limiter.allow(key) {
			// 超过限制，返回429错误
			c.JSON(http.StatusTooManyRequests, model.NewErrorResponse(
				http.StatusTooManyRequests,
				"请求过于频繁，请稍后再试",
				nil,
			))
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitByUser 基于用户ID的限流中间件
// 需要在认证中间件之后使用
func RateLimitByUser(limit int, window time.Duration) gin.HandlerFunc {
	limiter := newRateLimiter(limit, window)

	return func(c *gin.Context) {
		// 尝试从context获取用户ID
		userID, exists := c.Get("userID")
		if !exists {
			// 如果没有用户ID，使用IP
			userID = c.ClientIP()
		}

		key := userID.(string)

		if !limiter.allow(key) {
			c.JSON(http.StatusTooManyRequests, model.NewErrorResponse(
				http.StatusTooManyRequests,
				"请求过于频繁，请稍后再试",
				nil,
			))
			c.Abort()
			return
		}

		c.Next()
	}
}
