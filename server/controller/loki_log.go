package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/mingrenya/AI-Waf/server/config"
	"github.com/mingrenya/AI-Waf/server/service"
	"github.com/mingrenya/AI-Waf/server/utils/response"
	"github.com/rs/zerolog"
)

type LokiLogController interface {
	QueryLogs(ctx *gin.Context)
	QueryRange(ctx *gin.Context)
}

type lokiLogControllerImpl struct {
	service service.LokiLogService
	logger  zerolog.Logger
}

func NewLokiLogController(service service.LokiLogService) LokiLogController {
	return &lokiLogControllerImpl{
		service: service,
		logger:  config.GetControllerLogger("loki-log"),
	}
}

// QueryLogs 即时日志查询
// POST /api/v1/log/loki-query
func (c *lokiLogControllerImpl) QueryLogs(ctx *gin.Context) {
	var req service.LokiQueryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, err, true)
		return
	}

	result, err := c.service.QueryLogs(ctx.Request.Context(), req)
	if err != nil {
		c.logger.Error().Err(err).Msg("Loki 查询失败")
		response.InternalServerError(ctx, err, false)
		return
	}

	entries := service.ToLogEntries(result)
	entries.Query = req.Query
	response.Success(ctx, "查询成功", entries)
}

// QueryRange 范围日志查询
// POST /api/v1/log/loki-range
func (c *lokiLogControllerImpl) QueryRange(ctx *gin.Context) {
	var req service.LokiRangeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, err, true)
		return
	}

	result, err := c.service.QueryRange(ctx.Request.Context(), req)
	if err != nil {
		c.logger.Error().Err(err).Msg("Loki 范围查询失败")
		response.InternalServerError(ctx, err, false)
		return
	}

	entries := service.ToLogEntries(result)
	entries.Query = req.Query
	response.Success(ctx, "查询成功", entries)
}
