package controller

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mingrenya/AI-Waf/server/dto"
	"github.com/mingrenya/AI-Waf/server/model"
	"github.com/mingrenya/AI-Waf/server/service"
	"github.com/mingrenya/AI-Waf/server/utils/response"
)

// AlertController 告警控制器接口
type AlertController interface {
	// Channel 管理
	CreateChannel(ctx *gin.Context)
	GetChannels(ctx *gin.Context)
	GetChannelByID(ctx *gin.Context)
	UpdateChannel(ctx *gin.Context)
	DeleteChannel(ctx *gin.Context)
	TestChannel(ctx *gin.Context)

	// Rule 管理
	CreateRule(ctx *gin.Context)
	GetRules(ctx *gin.Context)
	GetRuleByID(ctx *gin.Context)
	UpdateRule(ctx *gin.Context)
	DeleteRule(ctx *gin.Context)

	// History 查询
	GetAlertHistory(ctx *gin.Context)
	AcknowledgeAlert(ctx *gin.Context)
	GetStatistics(ctx *gin.Context)
}

type alertControllerImpl struct {
	alertService service.AlertService
}

// NewAlertController 创建告警控制器
func NewAlertController(alertService service.AlertService) AlertController {
	return &alertControllerImpl{
		alertService: alertService,
	}
}

// CreateChannel 创建告警渠道
// @Summary 创建告警渠道
// @Description 创建新的告警渠道
// @Tags Alert
// @Accept json
// @Produce json
// @Param request body dto.CreateAlertChannelRequest true "创建渠道请求"
// @Success 200 {object} response.Response{data=dto.AlertChannelResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /alerts/channels [post]
func (c *alertControllerImpl) CreateChannel(ctx *gin.Context) {
	var req dto.CreateAlertChannelRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, model.ErrBadRequest(err), true)
		return
	}

	userID, _ := ctx.Get("userID")
	userIDStr, _ := userID.(string)

	result, err := c.alertService.CreateChannel(ctx.Request.Context(), &req, userIDStr)
	if err != nil {
		response.Error(ctx, model.ErrInternalServerError(err), true)
		return
	}

	response.Success(ctx, "Channel created successfully", result)
}

// GetChannels 获取所有告警渠道
// @Summary 获取告警渠道列表
// @Description 获取所有告警渠道
// @Tags Alert
// @Produce json
// @Success 200 {object} response.Response{data=[]dto.AlertChannelResponse}
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /alerts/channels [get]
func (c *alertControllerImpl) GetChannels(ctx *gin.Context) {
	channels, err := c.alertService.GetChannels(ctx.Request.Context())
	if err != nil {
		response.Error(ctx, model.ErrInternalServerError(err), true)
		return
	}

	response.Success(ctx, "Channels retrieved successfully", channels)
}

// GetChannelByID 获取告警渠道详情
// @Summary 获取告警渠道详情
// @Description 根据 ID 获取告警渠道详情
// @Tags Alert
// @Produce json
// @Param id path string true "渠道 ID"
// @Success 200 {object} response.Response{data=dto.AlertChannelResponse}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /alerts/channels/{id} [get]
func (c *alertControllerImpl) GetChannelByID(ctx *gin.Context) {
	id := ctx.Param("id")

	channel, err := c.alertService.GetChannelByID(ctx.Request.Context(), id)
	if err != nil {
		response.Error(ctx, model.ErrNotFound(err), true)
		return
	}

	response.Success(ctx, "Channel retrieved successfully", channel)
}

// UpdateChannel 更新告警渠道
// @Summary 更新告警渠道
// @Description 更新告警渠道信息
// @Tags Alert
// @Accept json
// @Produce json
// @Param id path string true "渠道 ID"
// @Param request body dto.UpdateAlertChannelRequest true "更新渠道请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /alerts/channels/{id} [put]
func (c *alertControllerImpl) UpdateChannel(ctx *gin.Context) {
	id := ctx.Param("id")

	var req dto.UpdateAlertChannelRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, model.ErrBadRequest(err), true)
		return
	}

	if err := c.alertService.UpdateChannel(ctx.Request.Context(), id, &req); err != nil {
		response.Error(ctx, model.ErrInternalServerError(err), true)
		return
	}

	response.Success(ctx, "Channel updated successfully", nil)
}

// DeleteChannel 删除告警渠道
// @Summary 删除告警渠道
// @Description 删除指定的告警渠道
// @Tags Alert
// @Produce json
// @Param id path string true "渠道 ID"
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /alerts/channels/{id} [delete]
func (c *alertControllerImpl) DeleteChannel(ctx *gin.Context) {
	id := ctx.Param("id")

	if err := c.alertService.DeleteChannel(ctx.Request.Context(), id); err != nil {
		response.Error(ctx, model.ErrInternalServerError(err), true)
		return
	}

	response.Success(ctx, "Channel deleted successfully", nil)
}

// TestChannel 测试告警渠道
// @Summary 测试告警渠道
// @Description 发送测试消息到指定渠道
// @Tags Alert
// @Accept json
// @Produce json
// @Param id path string true "渠道 ID"
// @Param request body dto.TestAlertChannelRequest true "测试请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /alerts/channels/{id}/test [post]
func (c *alertControllerImpl) TestChannel(ctx *gin.Context) {
	id := ctx.Param("id")

	var req dto.TestAlertChannelRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, model.ErrBadRequest(err), true)
		return
	}

	if err := c.alertService.TestChannel(ctx.Request.Context(), id, &req); err != nil {
		response.Error(ctx, model.ErrInternalServerError(err), true)
		return
	}

	response.Success(ctx, "Test message sent successfully", gin.H{"message": "Test message sent successfully"})
}

// CreateRule 创建告警规则
// @Summary 创建告警规则
// @Description 创建新的告警规则
// @Tags Alert
// @Accept json
// @Produce json
// @Param request body dto.CreateAlertRuleRequest true "创建规则请求"
// @Success 200 {object} response.Response{data=dto.AlertRuleResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /alerts/rules [post]
func (c *alertControllerImpl) CreateRule(ctx *gin.Context) {
	var req dto.CreateAlertRuleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, model.ErrBadRequest(err), true)
		return
	}

	userID, _ := ctx.Get("userID")
	userIDStr, _ := userID.(string)

	result, err := c.alertService.CreateRule(ctx.Request.Context(), &req, userIDStr)
	if err != nil {
		response.Error(ctx, model.ErrInternalServerError(err), true)
		return
	}

	response.Success(ctx, "Rule created successfully", result)
}

// GetRules 获取所有告警规则
// @Summary 获取告警规则列表
// @Description 获取所有告警规则
// @Tags Alert
// @Produce json
// @Success 200 {object} response.Response{data=[]dto.AlertRuleResponse}
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /alerts/rules [get]
func (c *alertControllerImpl) GetRules(ctx *gin.Context) {
	rules, err := c.alertService.GetRules(ctx.Request.Context())
	if err != nil {
		response.Error(ctx, model.ErrInternalServerError(err), true)
		return
	}

	response.Success(ctx, "Rules retrieved successfully", rules)
}

// GetRuleByID 获取告警规则详情
// @Summary 获取告警规则详情
// @Description 根据 ID 获取告警规则详情
// @Tags Alert
// @Produce json
// @Param id path string true "规则 ID"
// @Success 200 {object} response.Response{data=dto.AlertRuleResponse}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /alerts/rules/{id} [get]
func (c *alertControllerImpl) GetRuleByID(ctx *gin.Context) {
	id := ctx.Param("id")

	rule, err := c.alertService.GetRuleByID(ctx.Request.Context(), id)
	if err != nil {
		response.Error(ctx, model.ErrNotFound(err), true)
		return
	}

	response.Success(ctx, "Rule retrieved successfully", rule)
}

// UpdateRule 更新告警规则
// @Summary 更新告警规则
// @Description 更新告警规则信息
// @Tags Alert
// @Accept json
// @Produce json
// @Param id path string true "规则 ID"
// @Param request body dto.UpdateAlertRuleRequest true "更新规则请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /alerts/rules/{id} [put]
func (c *alertControllerImpl) UpdateRule(ctx *gin.Context) {
	id := ctx.Param("id")

	var req dto.UpdateAlertRuleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, model.ErrBadRequest(err), true)
		return
	}

	if err := c.alertService.UpdateRule(ctx.Request.Context(), id, &req); err != nil {
		response.Error(ctx, model.ErrInternalServerError(err), true)
		return
	}

	response.Success(ctx, "Rule updated successfully", nil)
}

// DeleteRule 删除告警规则
// @Summary 删除告警规则
// @Description 删除指定的告警规则
// @Tags Alert
// @Produce json
// @Param id path string true "规则 ID"
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /alerts/rules/{id} [delete]
func (c *alertControllerImpl) DeleteRule(ctx *gin.Context) {
	id := ctx.Param("id")

	if err := c.alertService.DeleteRule(ctx.Request.Context(), id); err != nil {
		response.Error(ctx, model.ErrInternalServerError(err), true)
		return
	}

	response.Success(ctx, "Rule deleted successfully", nil)
}

// GetAlertHistory 获取告警历史
// @Summary 获取告警历史
// @Description 分页查询告警历史记录
// @Tags Alert
// @Produce json
// @Param ruleId query string false "规则 ID"
// @Param severity query string false "严重级别"
// @Param status query string false "状态"
// @Param startTime query string false "开始时间"
// @Param endTime query string false "结束时间"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(20)
// @Success 200 {object} response.Response{data=[]dto.AlertHistoryResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /alerts/history [get]
func (c *alertControllerImpl) GetAlertHistory(ctx *gin.Context) {
	var req dto.GetAlertHistoryRequest

	req.RuleID = ctx.Query("ruleId")
	req.Severity = ctx.Query("severity")
	req.Status = ctx.Query("status")

	if startTimeStr := ctx.Query("startTime"); startTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			req.StartTime = t
		}
	}

	if endTimeStr := ctx.Query("endTime"); endTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			req.EndTime = t
		}
	}

	if pageStr := ctx.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			req.Page = page
		}
	}
	if req.Page == 0 {
		req.Page = 1
	}

	if pageSizeStr := ctx.Query("pageSize"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil {
			req.PageSize = pageSize
		}
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	histories, total, err := c.alertService.GetAlertHistory(ctx.Request.Context(), &req)
	if err != nil {
		response.Error(ctx, model.ErrInternalServerError(err), true)
		return
	}

	response.SuccessWithPagination(ctx, histories, total, req.Page, req.PageSize)
}

// AcknowledgeAlert 确认告警
// @Summary 确认告警
// @Description 确认已读告警
// @Tags Alert
// @Accept json
// @Produce json
// @Param id path string true "告警历史 ID"
// @Param request body dto.AcknowledgeAlertRequest false "确认请求"
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /alerts/history/{id}/acknowledge [post]
func (c *alertControllerImpl) AcknowledgeAlert(ctx *gin.Context) {
	id := ctx.Param("id")

	var req dto.AcknowledgeAlertRequest
	_ = ctx.ShouldBindJSON(&req)

	userID, _ := ctx.Get("userID")
	userIDStr, _ := userID.(string)

	if err := c.alertService.AcknowledgeAlert(ctx.Request.Context(), id, userIDStr, req.Comment); err != nil {
		response.Error(ctx, model.ErrInternalServerError(err), true)
		return
	}

	response.Success(ctx, "Alert acknowledged successfully", nil)
}

// GetStatistics 获取告警统计
// @Summary 获取告警统计
// @Description 获取告警统计信息
// @Tags Alert
// @Produce json
// @Param startTime query string false "开始时间"
// @Param endTime query string false "结束时间"
// @Success 200 {object} response.Response{data=dto.AlertStatisticsResponse}
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /alerts/statistics [get]
func (c *alertControllerImpl) GetStatistics(ctx *gin.Context) {
	var startTime, endTime time.Time

	if startTimeStr := ctx.Query("startTime"); startTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			startTime = t
		}
	}

	if endTimeStr := ctx.Query("endTime"); endTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			endTime = t
		}
	}

	stats, err := c.alertService.GetStatistics(ctx.Request.Context(), startTime, endTime)
	if err != nil {
		response.Error(ctx, model.ErrInternalServerError(err), true)
		return
	}

	response.Success(ctx, "Statistics retrieved successfully", stats)
}
