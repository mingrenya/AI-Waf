package response

import (
	"net/http"

	"github.com/HUAHUAI23/RuiQi/server/config"
	"github.com/HUAHUAI23/RuiQi/server/model"
	"github.com/gin-gonic/gin"
)

// WithRequestID 添加请求ID到响应
func WithRequestID(c *gin.Context, resp model.APIResponse) model.APIResponse {
	requestID, exists := c.Get("RequestID")
	if exists {
		resp.RequestID = requestID.(string)
	}
	return resp
}

// Success 返回成功响应
func Success(c *gin.Context, message string, data interface{}) {
	resp := model.NewSuccessResponse(message, data)
	resp = WithRequestID(c, resp)
	c.JSON(http.StatusOK, resp)
}

// Error 返回错误响应
func Error(c *gin.Context, apiErr *model.APIError, showErr bool) {
	// 记录错误日志
	logger := config.Logger
	if apiErr.Err != nil {
		logger.Error().Err(apiErr.Err).Str("message", apiErr.Message).Int("code", apiErr.Code).Send()
	} else {
		logger.Error().Str("message", apiErr.Message).Int("code", apiErr.Code).Send()
	}

	// 创建错误响应
	// err type zero value is nil
	var errToShow error
	if showErr {
		errToShow = apiErr.Err
	}
	resp := model.NewErrorResponse(apiErr.Code, apiErr.Message, errToShow)
	resp = WithRequestID(c, resp)

	// 返回错误响应
	c.JSON(apiErr.Code, resp)
	c.Abort()
}

// BadRequest 返回400错误
func BadRequest(c *gin.Context, err error, showErr bool) {
	Error(c, model.ErrBadRequest(err), showErr)
}

// Unauthorized 返回401错误
func Unauthorized(c *gin.Context, err error) {
	Error(c, model.ErrUnauthorized(err), false)
}

// Forbidden 返回403错误
func Forbidden(c *gin.Context, err error) {
	Error(c, model.ErrForbidden(err), false)
}

// NotFound 返回404错误
func NotFound(c *gin.Context, err error) {
	Error(c, model.ErrNotFound(err), false)
}

// InternalServerError 返回500错误
func InternalServerError(c *gin.Context, err error, showErr bool) {
	Error(c, model.ErrInternalServerError(err), showErr)
}
