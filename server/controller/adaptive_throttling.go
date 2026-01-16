package controller

import (
	"errors"

	"github.com/mingrenya/AI-Waf/server/config"
	"github.com/mingrenya/AI-Waf/server/dto"
	"github.com/mingrenya/AI-Waf/server/service"
	"github.com/mingrenya/AI-Waf/server/utils/response"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// AdaptiveThrottlingController 自适应限流控制器接口
type AdaptiveThrottlingController interface {
	GetConfig(ctx *gin.Context)
	UpdateConfig(ctx *gin.Context)
	DeleteConfig(ctx *gin.Context)
	GetTrafficPatterns(ctx *gin.Context)
	GetBaselines(ctx *gin.Context)
	GetAdjustmentLogs(ctx *gin.Context)
	GetStats(ctx *gin.Context)
	RecalculateBaseline(ctx *gin.Context)
	ResetLearning(ctx *gin.Context)
}

// AdaptiveThrottlingControllerImpl 自适应限流控制器实现
type AdaptiveThrottlingControllerImpl struct {
	service service.AdaptiveThrottlingService
	logger  zerolog.Logger
}

// NewAdaptiveThrottlingController 创建自适应限流控制器
func NewAdaptiveThrottlingController(service service.AdaptiveThrottlingService) AdaptiveThrottlingController {
	logger := config.GetControllerLogger("adaptive_throttling")
	return &AdaptiveThrottlingControllerImpl{
		service: service,
		logger:  logger,
	}
}

// GetConfig 获取配置
//
//	@Summary		获取自适应限流配置
//	@Description	获取当前自适应限流系统配置
//	@Tags			自适应限流
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse{data=model.AdaptiveThrottlingConfig}	"获取成功"
//	@Failure		404	{object}	model.ErrResponseDontShowError								"配置不存在"
//	@Failure		500	{object}	model.ErrResponseDontShowError								"服务器错误"
//	@Router			/api/v1/adaptive-throttling [get]
func (c *AdaptiveThrottlingControllerImpl) GetConfig(ctx *gin.Context) {
	cfg, err := c.service.GetConfig(ctx.Request.Context())
	if err != nil {
		if errors.Is(err, service.ErrAdaptiveThrottlingConfigNotFound) || errors.Is(err, mongo.ErrNoDocuments) {
			response.NotFound(ctx, errors.New("配置不存在"))
			return
		}
		c.logger.Error().Err(err).Msg("获取配置失败")
		response.InternalServerError(ctx, err, false)
		return
	}

	response.Success(ctx, "获取配置成功", cfg)
}

// UpdateConfig 更新配置
//
//	@Summary		更新自适应限流配置
//	@Description	更新自适应限流系统配置，如果配置不存在则创建
//	@Tags			自适应限流
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.AdaptiveThrottlingConfigRequest						true	"配置信息"
//	@Success		200		{object}	model.SuccessResponse{data=model.AdaptiveThrottlingConfig}	"更新成功"
//	@Failure		400		{object}	model.ErrResponseDontShowError								"请求参数错误"
//	@Failure		500		{object}	model.ErrResponseDontShowError								"服务器错误"
//	@Router			/api/v1/adaptive-throttling [put]
func (c *AdaptiveThrottlingControllerImpl) UpdateConfig(ctx *gin.Context) {
	var req dto.AdaptiveThrottlingConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, err, true)
		return
	}

	// 尝试更新，如果不存在则创建
	cfg, err := c.service.UpdateConfig(ctx.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrAdaptiveThrottlingConfigNotFound) {
			// 配置不存在，创建新配置
			cfg, err = c.service.CreateConfig(ctx.Request.Context(), &req)
			if err != nil {
				c.logger.Error().Err(err).Msg("创建配置失败")
				response.InternalServerError(ctx, err, false)
				return
			}
		} else {
			c.logger.Error().Err(err).Msg("更新配置失败")
			response.InternalServerError(ctx, err, false)
			return
		}
	}

	response.Success(ctx, "更新配置成功", cfg)
}

// DeleteConfig 删除配置
//
//	@Summary		删除自适应限流配置
//	@Description	删除自适应限流系统配置
//	@Tags			自适应限流
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse	"删除成功"
//	@Failure		500	{object}	model.ErrResponseDontShowError	"服务器错误"
//	@Router			/api/v1/adaptive-throttling [delete]
func (c *AdaptiveThrottlingControllerImpl) DeleteConfig(ctx *gin.Context) {
	if err := c.service.DeleteConfig(ctx.Request.Context()); err != nil {
		c.logger.Error().Err(err).Msg("删除配置失败")
		response.InternalServerError(ctx, err, false)
		return
	}

	response.Success(ctx, "删除配置成功", nil)
}

// GetTrafficPatterns 获取流量模式
//
//	@Summary		获取流量模式列表
//	@Description	查询流量模式历史记录
//	@Tags			自适应限流
//	@Produce		json
//	@Security		BearerAuth
//	@Param			type		query		string									false	"类型筛选 (visit/attack/error)"
//	@Param			startTime	query		string									false	"开始时间"
//	@Param			endTime		query		string									false	"结束时间"
//	@Param			page		query		int										false	"页码"
//	@Param			pageSize	query		int										false	"每页数量"
//	@Success		200			{object}	model.SuccessResponse{data=dto.TrafficPatternResponse}	"查询成功"
//	@Failure		400			{object}	model.ErrResponseDontShowError			"请求参数错误"
//	@Failure		500			{object}	model.ErrResponseDontShowError			"服务器错误"
//	@Router			/api/v1/adaptive-throttling/patterns [get]
func (c *AdaptiveThrottlingControllerImpl) GetTrafficPatterns(ctx *gin.Context) {
	var query dto.TrafficPatternQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		response.BadRequest(ctx, err, true)
		return
	}

	result, err := c.service.GetTrafficPatterns(ctx.Request.Context(), &query)
	if err != nil {
		c.logger.Error().Err(err).Msg("获取流量模式失败")
		response.InternalServerError(ctx, err, false)
		return
	}

	response.Success(ctx, "获取流量模式成功", result)
}

// GetBaselines 获取基线值
//
//	@Summary		获取基线值列表
//	@Description	查询当前基线值
//	@Tags			自适应限流
//	@Produce		json
//	@Security		BearerAuth
//	@Param			type	query		string											false	"类型筛选 (visit/attack/error)"
//	@Success		200		{object}	model.SuccessResponse{data=dto.BaselineResponse}	"查询成功"
//	@Failure		400		{object}	model.ErrResponseDontShowError					"请求参数错误"
//	@Failure		500		{object}	model.ErrResponseDontShowError					"服务器错误"
//	@Router			/api/v1/adaptive-throttling/baselines [get]
func (c *AdaptiveThrottlingControllerImpl) GetBaselines(ctx *gin.Context) {
	var query dto.BaselineQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		response.BadRequest(ctx, err, true)
		return
	}

	result, err := c.service.GetBaselines(ctx.Request.Context(), &query)
	if err != nil {
		c.logger.Error().Err(err).Msg("获取基线值失败")
		response.InternalServerError(ctx, err, false)
		return
	}

	response.Success(ctx, "获取基线值成功", result)
}

// GetAdjustmentLogs 获取调整日志
//
//	@Summary		获取调整日志列表
//	@Description	查询限流阈值调整历史记录
//	@Tags			自适应限流
//	@Produce		json
//	@Security		BearerAuth
//	@Param			type		query		string												false	"类型筛选 (visit/attack/error)"
//	@Param			startTime	query		string												false	"开始时间"
//	@Param			endTime		query		string												false	"结束时间"
//	@Param			page		query		int													false	"页码"
//	@Param			pageSize	query		int													false	"每页数量"
//	@Success		200			{object}	model.SuccessResponse{data=dto.AdjustmentLogResponse}	"查询成功"
//	@Failure		400			{object}	model.ErrResponseDontShowError						"请求参数错误"
//	@Failure		500			{object}	model.ErrResponseDontShowError						"服务器错误"
//	@Router			/api/v1/adaptive-throttling/logs [get]
func (c *AdaptiveThrottlingControllerImpl) GetAdjustmentLogs(ctx *gin.Context) {
	var query dto.AdjustmentLogQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		response.BadRequest(ctx, err, true)
		return
	}

	result, err := c.service.GetAdjustmentLogs(ctx.Request.Context(), &query)
	if err != nil {
		c.logger.Error().Err(err).Msg("获取调整日志失败")
		response.InternalServerError(ctx, err, false)
		return
	}

	response.Success(ctx, "获取调整日志成功", result)
}

// GetStats 获取统计信息
//
//	@Summary		获取统计信息
//	@Description	获取自适应限流系统实时统计信息
//	@Tags			自适应限流
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse{data=dto.AdaptiveThrottlingStatsDTO}	"查询成功"
//	@Failure		500	{object}	model.ErrResponseDontShowError								"服务器错误"
//	@Router			/api/v1/adaptive-throttling/stats [get]
func (c *AdaptiveThrottlingControllerImpl) GetStats(ctx *gin.Context) {
	stats, err := c.service.GetStats(ctx.Request.Context())
	if err != nil {
		if errors.Is(err, service.ErrAdaptiveThrottlingConfigNotFound) {
			response.NotFound(ctx, errors.New("配置不存在"))
			return
		}
		c.logger.Error().Err(err).Msg("获取统计信息失败")
		response.InternalServerError(ctx, err, false)
		return
	}

	response.Success(ctx, "获取统计信息成功", stats)
}

// RecalculateBaseline 重新计算基线
//
//	@Summary		重新计算基线
//	@Description	手动触发重新计算指定类型的基线值
//	@Tags			自适应限流
//	@Produce		json
//	@Security		BearerAuth
//	@Param			type	query		string							true	"类型 (visit/attack/error)"
//	@Success		200		{object}	model.SuccessResponse			"操作成功"
//	@Failure		400		{object}	model.ErrResponseDontShowError	"请求参数错误"
//	@Failure		500		{object}	model.ErrResponseDontShowError	"服务器错误"
//	@Router			/api/v1/adaptive-throttling/recalculate-baseline [post]
func (c *AdaptiveThrottlingControllerImpl) RecalculateBaseline(ctx *gin.Context) {
	typ := ctx.Query("type")
	if typ == "" {
		response.BadRequest(ctx, errors.New("类型参数不能为空"), true)
		return
	}

	if typ != "visit" && typ != "attack" && typ != "error" {
		response.BadRequest(ctx, errors.New("类型参数必须是 visit、attack 或 error"), true)
		return
	}

	if err := c.service.RecalculateBaseline(ctx.Request.Context(), typ); err != nil {
		c.logger.Error().Err(err).Msg("重新计算基线失败")
		response.InternalServerError(ctx, err, false)
		return
	}

	response.Success(ctx, "重新计算基线成功", nil)
}

// ResetLearning 重置学习
//
//	@Summary		重置学习
//	@Description	清空学习数据并重新开始学习
//	@Tags			自适应限流
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse			"操作成功"
//	@Failure		500	{object}	model.ErrResponseDontShowError	"服务器错误"
//	@Router			/api/v1/adaptive-throttling/reset-learning [post]
func (c *AdaptiveThrottlingControllerImpl) ResetLearning(ctx *gin.Context) {
	if err := c.service.ResetLearning(ctx.Request.Context()); err != nil {
		c.logger.Error().Err(err).Msg("重置学习失败")
		response.InternalServerError(ctx, err, false)
		return
	}

	response.Success(ctx, "重置学习成功", nil)
}
