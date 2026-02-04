package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mingrenya/AI-Waf/server/model"
	"github.com/mingrenya/AI-Waf/server/utils/response"
)

// Timeout 请求超时中间件
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 创建带超时的context
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// 替换请求的context
		c.Request = c.Request.WithContext(ctx)

		// 使用channel来监听请求完成
		finished := make(chan struct{})
		panicChan := make(chan interface{}, 1)

		go func() {
			defer func() {
				if p := recover(); p != nil {
					panicChan <- p
				}
			}()
			c.Next()
			finished <- struct{}{}
		}()

		select {
		case <-finished:
			// 请求正常完成
			return
		case p := <-panicChan:
			// 捕获panic
			panic(p)
		case <-ctx.Done():
			// 请求超时
			c.Header("Connection", "close")
			response.Error(c, &model.APIError{
				Code:    http.StatusGatewayTimeout,
				Message: fmt.Sprintf("Request timeout after %v", timeout),
			}, true)
			c.Abort()
			return
		}
	}
}

// TimeoutWithCustomHandler 带自定义处理器的超时中间件
func TimeoutWithCustomHandler(timeout time.Duration, handler func(*gin.Context)) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		finished := make(chan struct{})
		panicChan := make(chan interface{}, 1)

		go func() {
			defer func() {
				if p := recover(); p != nil {
					panicChan <- p
				}
			}()
			c.Next()
			finished <- struct{}{}
		}()

		select {
		case <-finished:
			return
		case p := <-panicChan:
			panic(p)
		case <-ctx.Done():
			handler(c)
			c.Abort()
			return
		}
	}
}
