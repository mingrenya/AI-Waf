package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/mingrenya/AI-Waf/server/config"
	"github.com/mingrenya/AI-Waf/server/model"
	"github.com/mingrenya/AI-Waf/server/utils/response"
)

// Recovery 错误恢复中间件 - 捕获panic并返回友好错误
func Recovery() gin.HandlerFunc {
	log := config.Logger

	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 获取堆栈信息
				stack := string(debug.Stack())

				// 获取请求信息
				requestID := c.GetString("requestID")
				method := c.Request.Method
				path := c.Request.URL.Path
				clientIP := c.ClientIP()

				// 记录错误日志
				log.Error().
					Str("request_id", requestID).
					Str("method", method).
					Str("path", path).
					Str("client_ip", clientIP).
					Interface("error", err).
					Str("stack", stack).
					Msg("Panic recovered")

				// 检查是否是已知的错误类型
				errMsg := fmt.Sprintf("%v", err)

				// 返回错误响应
				response.InternalServerError(c, fmt.Errorf("internal server error: %s", errMsg), false)

				// 中止后续处理
				c.Abort()
			}
		}()

		c.Next()
	}
}

// ErrorHandler 统一错误处理中间件
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 检查是否有错误
		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			// 如果响应已经写入，不再处理
			if c.Writer.Written() {
				return
			}

			// 根据错误类型返回不同的响应
			switch err.Type {
			case gin.ErrorTypeBind:
				response.BadRequest(c, err.Err, true)
			case gin.ErrorTypePublic:
				response.Error(c, &model.APIError{
					Code:    http.StatusBadRequest,
					Message: err.Err.Error(),
				}, true)
			default:
				response.InternalServerError(c, err.Err, false)
			}
		}
	}
}
