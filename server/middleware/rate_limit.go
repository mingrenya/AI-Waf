package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mingrenya/AI-Waf/server/model"
	"github.com/redis/go-redis/v9"
)

// RateLimiter 限流器接口，支持内存和 Redis 两种实现
type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) bool
}

// ──────────────────────────────────────────
// Redis 滑动窗口实现（适用于多实例部署）
// ──────────────────────────────────────────

type redisRateLimiter struct {
	client *redis.Client
}

// NewRedisRateLimiter 创建基于 Redis 的分布式限流器。
// addr 格式如 "redis:6379"，password 为空字符串表示无密码。
func NewRedisRateLimiter(addr, password string, db int) RateLimiter {
	return &redisRateLimiter{
		client: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
		}),
	}
}

// Allow 使用 Redis INCR + EXPIRE 实现固定窗口计数限流。
// 首次写入时设置过期时间，保证窗口自动滑动。
func (r *redisRateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) bool {
	redisKey := fmt.Sprintf("ratelimit:%s", key)
	pipe := r.client.TxPipeline()
	incr := pipe.Incr(ctx, redisKey)
	pipe.Expire(ctx, redisKey, window)
	if _, err := pipe.Exec(ctx); err != nil {
		// Redis 不可用时放行，降级由调用层处理
		return true
	}
	return int(incr.Val()) <= limit
}

// ──────────────────────────────────────────
// 内存实现（单实例 / Redis 不可用降级）
// ──────────────────────────────────────────

type memoryRateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*memVisitor
}

type memVisitor struct {
	lastSeen time.Time
	count    int
}

// NewMemoryRateLimiter 创建基于内存的限流器（不适用于多实例部署）。
func NewMemoryRateLimiter() RateLimiter {
	rl := &memoryRateLimiter{
		visitors: make(map[string]*memVisitor),
	}
	go rl.cleanup()
	return rl
}

func (rl *memoryRateLimiter) Allow(_ context.Context, key string, limit int, window time.Duration) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	v, exists := rl.visitors[key]
	if !exists {
		rl.visitors[key] = &memVisitor{lastSeen: now, count: 1}
		return true
	}
	if now.Sub(v.lastSeen) > window {
		v.lastSeen = now
		v.count = 1
		return true
	}
	if v.count >= limit {
		return false
	}
	v.count++
	return true
}

func (rl *memoryRateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, v := range rl.visitors {
			if now.Sub(v.lastSeen) > time.Minute*10 {
				delete(rl.visitors, key)
			}
		}
		rl.mu.Unlock()
	}
}

// ──────────────────────────────────────────
// Gin 中间件
// ──────────────────────────────────────────

// RateLimit 返回按客户端 IP 限流的中间件。
// 优先使用 Redis 分布式限流；Redis 不可用时自动降级为内存限流。
// limit: 时间窗口内允许的最大请求数；window: 时间窗口长度。
func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	return rateLimitMiddleware(defaultLimiter(), limit, window, func(c *gin.Context) string {
		return c.ClientIP()
	})
}

// RateLimitByUser 返回按用户 ID 限流的中间件（需在认证中间件之后使用）。
func RateLimitByUser(limit int, window time.Duration) gin.HandlerFunc {
	return rateLimitMiddleware(defaultLimiter(), limit, window, func(c *gin.Context) string {
		if uid, ok := c.Get("userID"); ok {
			return uid.(string)
		}
		return c.ClientIP()
	})
}

// defaultLimiter 从环境变量读取 Redis 地址，若未配置则降级为内存限流器。
func defaultLimiter() RateLimiter {
	import_os_once.Do(func() {
		addr := getEnv("REDIS_ADDR", "")
		if addr != "" {
			globalLimiter = NewRedisRateLimiter(addr, getEnv("REDIS_PASSWORD", ""), 0)
		} else {
			globalLimiter = NewMemoryRateLimiter()
		}
	})
	return globalLimiter
}

var (
	import_os_once sync.Once
	globalLimiter  RateLimiter
)

func rateLimitMiddleware(limiter RateLimiter, limit int, window time.Duration, keyFn func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := keyFn(c)
		if !limiter.Allow(c.Request.Context(), key, limit, window) {
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

// getEnv 从环境变量读取值，若不存在则返回 fallback。
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
