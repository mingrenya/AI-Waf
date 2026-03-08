package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

// Timeout 请求超时中间件。
//
// 实现说明：此中间件仅满足下游 handler 配合 context.Context 实现超时退出的却就场景。
// 超时时不再在当前 goroutine 之外写 c.Writer，避免数据竞争。
// 如需强制中断和写入超时响应体，应在同一 goroutine内结合 select 实现，
// 或使用 Nginx/HAProxy 在上游设置请求超时。
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// 将超时 context 注入请求；下游 handler 应检查 ctx.Err() 或通过 select 监听超时
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// TimeoutWithCustomHandler 带自定义处理器的超时中间件。
// handler 会在超时发生后在同一 goroutine 内调用。
func TimeoutWithCustomHandler(timeout time.Duration, handler func(*gin.Context)) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		c.Next()

		// 处理完成后检查是否因超时退出
		if ctx.Err() == context.DeadlineExceeded {
			handler(c)
		}
	}
}
