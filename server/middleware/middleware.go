package middleware

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mingrenya/AI-Waf/server/config"
	"github.com/mingrenya/AI-Waf/server/model"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// 辅助函数：获取环境变量或默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// 辅助函数：分割并去除空格
func splitAndTrim(s string, sep string) []string {
	parts := strings.Split(s, sep)
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// 辅助函数：检查slice是否包含某元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Logger middleware logs the request/response details
func Logger() gin.HandlerFunc {
	log := config.Logger
	isProduction := config.Global.IsProduction

	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Process request
		c.Next()

		// Calculate request time
		elapsed := time.Since(start)
		statusCode := c.Writer.Status()

		// 在生产环境中，只记录错误、警告或较慢的请求
		if !isProduction ||
			statusCode >= 400 ||
			elapsed > 500*time.Millisecond {

			// 根据状态码选择日志级别
			event := log.Info()
			if statusCode >= 400 && statusCode < 500 {
				event = log.Warn()
			} else if statusCode >= 500 {
				event = log.Error()
			}

			event.Str("method", method).
				Str("path", path).
				Int("status", statusCode).
				// Dur("latency", elapsed). //单位是毫秒
				Str("latency", elapsed.String()).
				Msg("HTTP Request")
		}
	}
}

// Cors middleware handles CORS requests with configurable origins
func Cors() gin.HandlerFunc {
	// 从环境变量获取允许的源，默认为localhost开发环境
	allowedOrigins := getEnvOrDefault("CORS_ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:3000,http://127.0.0.1:5173")
	origins := splitAndTrim(allowedOrigins, ",")

	return func(c *gin.Context) {
		requestOrigin := c.Request.Header.Get("Origin")

		// 检查请求源是否在允许列表中
		allowed := false
		for _, origin := range origins {
			if origin == requestOrigin || origin == "*" {
				allowed = true
				break
			}
		}

		if allowed {
			if requestOrigin != "" {
				c.Writer.Header().Set("Access-Control-Allow-Origin", requestOrigin)
			} else if contains(origins, "*") {
				c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			}
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
			c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, X-Request-ID")
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Max-Age", "43200") // 12小时
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// RequestID middleware generates and attaches a unique ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.New().String()
		c.Set("RequestID", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// CustomErrorHandler 简化版的错误处理函数
func CustomErrorHandler(c *gin.Context, err any) {
	requestID, _ := c.Get("RequestID")
	requestIDStr, _ := requestID.(string)

	// 记录错误日志
	config.Logger.Error().
		Interface("error", err).
		Str("request", c.Request.URL.Path).
		Str("requestId", requestIDStr).
		Msg("Recovery from panic")

	// 创建标准错误响应
	errorResp := model.NewErrorResponse(
		http.StatusInternalServerError,
		"服务器内部错误",
		nil,
		// fmt.Errorf("%v", err),
	)

	// 添加请求ID
	errorResp.RequestID = requestIDStr

	// 返回标准错误响应
	c.JSON(http.StatusInternalServerError, errorResp)
}
