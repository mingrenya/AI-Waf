package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mingrenya/AI-Waf/server/model"
	"github.com/mingrenya/AI-Waf/server/utils/response"
)

// ValidateContentType 验证Content-Type中间件
func ValidateContentType(allowedTypes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// GET和DELETE请求不需要验证Content-Type
		if c.Request.Method == "GET" || c.Request.Method == "DELETE" {
			c.Next()
			return
		}

		contentType := c.GetHeader("Content-Type")
		if contentType == "" {
			response.BadRequest(c, fmt.Errorf("Content-Type header is required"), true)
			c.Abort()
			return
		}

		// 检查是否在允许的类型中
		valid := false
		for _, allowed := range allowedTypes {
			if strings.HasPrefix(contentType, allowed) {
				valid = true
				break
			}
		}

		if !valid {
			response.Error(c, &model.APIError{
				Code:    http.StatusUnsupportedMediaType,
				Message: fmt.Sprintf("Content-Type must be one of: %s", strings.Join(allowedTypes, ", ")),
			}, true)
			c.Abort()
			return
		}

		c.Next()
	}
}

// ValidateJSONContentType 验证JSON Content-Type的快捷中间件
func ValidateJSONContentType() gin.HandlerFunc {
	return ValidateContentType("application/json")
}

// ValidatePathParam 验证路径参数中间件
func ValidatePathParam(paramName string, validator func(string) bool, errorMsg string) gin.HandlerFunc {
	return func(c *gin.Context) {
		value := c.Param(paramName)

		if !validator(value) {
			if errorMsg == "" {
				errorMsg = fmt.Sprintf("Invalid path parameter: %s", paramName)
			}
			response.BadRequest(c, errors.New(errorMsg), true)
			c.Abort()
			return
		}

		c.Next()
	}
}

// ValidateMongoID 验证MongoDB ObjectID格式的路径参数
func ValidateMongoID(paramName string) gin.HandlerFunc {
	hexPattern := regexp.MustCompile("^[0-9a-fA-F]{24}$")

	return ValidatePathParam(paramName, func(value string) bool {
		return hexPattern.MatchString(value)
	}, fmt.Sprintf("Invalid MongoDB ObjectID format for parameter: %s", paramName))
}

// ValidateUUID 验证UUID格式的路径参数
func ValidateUUID(paramName string) gin.HandlerFunc {
	uuidPattern := regexp.MustCompile("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")

	return ValidatePathParam(paramName, func(value string) bool {
		return uuidPattern.MatchString(value)
	}, fmt.Sprintf("Invalid UUID format for parameter: %s", paramName))
}

// ValidateQueryParam 验证查询参数中间件
func ValidateQueryParam(paramName string, required bool, validator func(string) bool, errorMsg string) gin.HandlerFunc {
	return func(c *gin.Context) {
		value := c.Query(paramName)

		if value == "" && required {
			response.BadRequest(c, fmt.Errorf("Required query parameter missing: %s", paramName), true)
			c.Abort()
			return
		}

		if value != "" && !validator(value) {
			if errorMsg == "" {
				errorMsg = fmt.Sprintf("Invalid query parameter: %s", paramName)
			}
			response.BadRequest(c, errors.New(errorMsg), true)
			c.Abort()
			return
		}

		c.Next()
	}
}

// ValidatePagination 验证分页参数中间件
func ValidatePagination() gin.HandlerFunc {
	return func(c *gin.Context) {
		page := c.DefaultQuery("page", "1")
		size := c.DefaultQuery("size", "10")

		// 验证page
		if !isPositiveInt(page) {
			response.BadRequest(c, fmt.Errorf("page must be a positive integer"), true)
			c.Abort()
			return
		}

		// 验证size
		if !isPositiveInt(size) || !isValidPageSize(size) {
			response.BadRequest(c, fmt.Errorf("size must be between 1 and 100"), true)
			c.Abort()
			return
		}

		c.Next()
	}
}

// isPositiveInt 检查字符串是否为正整数
func isPositiveInt(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// isValidPageSize 检查页面大小是否在有效范围内
func isValidPageSize(s string) bool {
	var size int
	fmt.Sscanf(s, "%d", &size)
	return size >= 1 && size <= 100
}

// SecurityHeaders 设置安全响应头中间件
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 防止点击劫持
		c.Header("X-Frame-Options", "DENY")

		// 启用XSS保护
		c.Header("X-XSS-Protection", "1; mode=block")

		// 禁止MIME类型嗅探
		c.Header("X-Content-Type-Options", "nosniff")

		// 严格的传输安全（仅HTTPS）
		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		// Content Security Policy
		c.Header("Content-Security-Policy", "default-src 'self'")

		// 引用策略
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// 权限策略
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		c.Next()
	}
}

// NoCache 禁用缓存中间件（用于敏感接口）
func NoCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate, private")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.Next()
	}
}
