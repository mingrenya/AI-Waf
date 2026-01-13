package middleware

import (
	"net/http"
	"time"

	"github.com/HUAHUAI23/RuiQi/server/config"
	"github.com/HUAHUAI23/RuiQi/server/model"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

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

// Cors middleware handles CORS requests
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

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
