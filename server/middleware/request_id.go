package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestID 请求追踪中间件 - 为每个请求生成唯一ID
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试从请求头获取已有的Request ID
		requestID := c.GetHeader("X-Request-ID")

		// 如果没有，生成新的UUID
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// 设置到context中，供后续处理使用
		c.Set("requestID", requestID)

		// 设置响应头，方便客户端追踪
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// TraceID 链路追踪中间件 - 支持分布式追踪
func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试从请求头获取Trace ID（兼容多种追踪系统）
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = c.GetHeader("X-B3-TraceId") // Zipkin格式
		}
		if traceID == "" {
			traceID = c.GetHeader("traceparent") // W3C格式
		}

		// 如果没有，生成新的
		if traceID == "" {
			traceID = uuid.New().String()
		}

		// 生成Span ID
		spanID := uuid.New().String()[:16]

		// 设置到context
		c.Set("traceID", traceID)
		c.Set("spanID", spanID)

		// 设置响应头
		c.Header("X-Trace-ID", traceID)
		c.Header("X-Span-ID", spanID)

		c.Next()
	}
}
