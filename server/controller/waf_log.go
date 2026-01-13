package controller

import (
	"time"

	"github.com/HUAHUAI23/RuiQi/server/dto"
	"github.com/HUAHUAI23/RuiQi/server/service"
	"github.com/HUAHUAI23/RuiQi/server/utils/response"
	"github.com/gin-gonic/gin"
)

type WAFLogController interface {
	GetAttackEvents(ctx *gin.Context)
	GetAttackLogs(ctx *gin.Context)
}

type WAFLogControllerImpl struct {
	wafLogService service.WAFLogService
}

// NewWAFLogController 创建新的WAF日志控制器实例
func NewWAFLogController(wafLogService service.WAFLogService) WAFLogController {
	return &WAFLogControllerImpl{
		wafLogService: wafLogService,
	}
}

// GetAttackEvents godoc
//
//	@Summary		获取聚合攻击事件
//	@Description	按来源IP、目标端口和域名聚合的攻击事件统计，支持多维度筛选和分页
//	@Tags			WAF安全日志
//	@Accept			json
//	@Produce		json
//	@Param			srcIp		query		string												false	"来源IP地址，攻击者地址"
//	@Param			dstIp		query		string												false	"目标IP地址，被攻击的服务器地址"
//	@Param			domain		query		string												false	"域名，被攻击的站点域名"
//	@Param			srcPort		query		integer												false	"来源端口号，发起攻击的端口"
//	@Param			dstPort		query		integer												false	"目标端口号，被攻击的服务端口"
//	@Param			startTime	query		string												false	"查询起始时间 (ISO8601格式，如: 2024-03-17T00:00:00Z)"
//	@Param			endTime		query		string												false	"查询结束时间 (ISO8601格式，如: 2024-03-18T23:59:59Z)"
//	@Param			page		query		integer												false	"当前页码，从1开始计数 (默认: 1)"
//	@Param			pageSize	query		integer												false	"每页记录数，最大100条 (默认: 10)"
//	@Success		200			{object}	model.SuccessResponse{data=dto.AttackEventResponse}	"成功"
//	@Failure		400			{object}	model.ErrResponse									"请求参数错误"
//	@Failure		500			{object}	model.ErrResponseDontShowError						"服务器内部错误"
//	@Router			/api/v1/waf/logs/events [get]
func (c *WAFLogControllerImpl) GetAttackEvents(ctx *gin.Context) {
	var req dto.AttackEventRequset

	// 使用ShouldBindQuery绑定查询参数
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.BadRequest(ctx, err, true)
		return
	}

	// 设置默认值时使用UTC时区
	if req.StartTime.IsZero() {
		// 默认: 24小时前，使用UTC时区
		req.StartTime = time.Now().UTC().Add(-24 * time.Hour)
	}
	if req.EndTime.IsZero() {
		// 默认: 当前时间，使用UTC时区
		req.EndTime = time.Now().UTC()
	}

	// 设置默认分页参数
	page := req.Page
	if page <= 0 {
		page = 1
	}

	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	} else if pageSize > 100 {
		pageSize = 100 // 限制最大每页数量
	}

	// 调用服务
	result, err := c.wafLogService.GetAttackEvents(ctx, req, page, pageSize)
	if err != nil {
		response.InternalServerError(ctx, err, false)
		return
	}

	response.Success(ctx, "获取攻击事件成功", result)
}

// GetAttackLogs godoc
//
//	@Summary		获取详细攻击日志
//	@Description	查询详细的WAF攻击日志记录，提供多条件筛选和分页功能，支持按规则ID、IP、域名、端口和时间范围过滤
//	@Tags			WAF安全日志
//	@Accept			json
//	@Produce		json
//	@Param			ruleId		query		integer												false	"规则ID，触发攻击检测的WAF规则标识"
//	@Param			srcIp		query		string												false	"来源IP地址，攻击者地址"
//	@Param			dstIp		query		string												false	"目标IP地址，被攻击的服务器地址"
//	@Param			domain		query		string												false	"域名，被攻击的站点域名"
//	@Param			srcPort		query		integer												false	"来源端口号，发起攻击的端口"
//	@Param			dstPort		query		integer												false	"目标端口号，被攻击的服务端口"
//	@Param			requestId	query		string												false	"请求ID，唯一标识HTTP请求的ID"
//	@Param			startTime	query		string												false	"查询起始时间 (ISO8601格式，如: 2024-03-17T00:00:00Z)"
//	@Param			endTime		query		string												false	"查询结束时间 (ISO8601格式，如: 2024-03-18T23:59:59Z)"
//	@Param			page		query		integer												false	"当前页码，从1开始计数 (默认: 1)"
//	@Param			pageSize	query		integer												false	"每页记录数，最大100条 (默认: 10)"
//	@Success		200			{object}	model.SuccessResponse{data=dto.AttackLogResponse}	"成功"
//	@Failure		400			{object}	model.ErrResponse									"请求参数错误"
//	@Failure		500			{object}	model.ErrResponseDontShowError						"服务器内部错误"
//	@Router			/api/v1/waf/logs [get]
func (c *WAFLogControllerImpl) GetAttackLogs(ctx *gin.Context) {
	var req dto.AttackLogRequest

	// 使用ShouldBindQuery绑定查询参数
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.BadRequest(ctx, err, true)
		return
	}

	// 设置默认值时使用UTC时区
	if req.StartTime.IsZero() {
		// 默认: 24小时前，使用UTC时区
		req.StartTime = time.Now().UTC().Add(-24 * time.Hour)
	}
	if req.EndTime.IsZero() {
		// 默认: 当前时间，使用UTC时区
		req.EndTime = time.Now().UTC()
	}

	// 设置默认分页参数
	page := req.Page
	if page <= 0 {
		page = 1
	}

	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	} else if pageSize > 100 {
		pageSize = 100 // 限制最大每页数量
	}

	// 调用服务
	result, err := c.wafLogService.GetAttackLogs(ctx, req, page, pageSize)
	if err != nil {
		response.InternalServerError(ctx, err, false)
		return
	}

	response.Success(ctx, "获取攻击日志成功", result)
}
