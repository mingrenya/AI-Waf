// server/controller/ai_analyzer.go
package controller

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mingrenya/AI-Waf/server/config"
	"github.com/mingrenya/AI-Waf/server/dto"
	"github.com/mingrenya/AI-Waf/server/model"
	"github.com/mingrenya/AI-Waf/server/service"
	"github.com/mingrenya/AI-Waf/server/utils/response"
	"github.com/rs/zerolog"
)

// AIAnalyzerController AI分析器控制器接口
type AIAnalyzerController interface {
	// 攻击模式相关
	ListAttackPatterns(ctx *gin.Context)
	GetAttackPattern(ctx *gin.Context)
	DeleteAttackPattern(ctx *gin.Context)

	// 生成规则相关
	ListGeneratedRules(ctx *gin.Context)
	GetGeneratedRule(ctx *gin.Context)
	DeleteGeneratedRule(ctx *gin.Context)
	ReviewRule(ctx *gin.Context)
	GetPendingRules(ctx *gin.Context)
	DeployRule(ctx *gin.Context)

	// AI分析器配置相关
	GetAnalyzerConfig(ctx *gin.Context)
	UpdateAnalyzerConfig(ctx *gin.Context)

	// MCP对话相关
	ListMCPConversations(ctx *gin.Context)
	GetMCPConversation(ctx *gin.Context)
	DeleteMCPConversation(ctx *gin.Context)

	// 统计分析相关
	GetAnalyzerStats(ctx *gin.Context)
	
	// 手动触发AI分析
	TriggerAnalysis(ctx *gin.Context)
}

// AIAnalyzerControllerImpl AI分析器控制器实现
type AIAnalyzerControllerImpl struct {
	service service.AIAnalyzerService
	logger  zerolog.Logger
}

// NewAIAnalyzerController 创建AI分析器控制器
func NewAIAnalyzerController(service service.AIAnalyzerService) AIAnalyzerController {
	logger := config.GetControllerLogger("ai_analyzer")
	return &AIAnalyzerControllerImpl{
		service: service,
		logger:  logger,
	}
}

// ============================================
// 攻击模式相关
// ============================================

// ListAttackPatterns 获取攻击模式列表
// @Summary 获取攻击模式列表
// @Tags AI分析器
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(10)
// @Param severity query string false "严重程度" Enums(critical, high, medium, low)
// @Param attackType query string false "攻击类型"
// @Param startTime query string false "开始时间(RFC3339)"
// @Param endTime query string false "结束时间(RFC3339)"
// @Success 200 {object} dto.AttackPatternListResponse
// @Router /api/v1/ai-analyzer/patterns [get]
func (c *AIAnalyzerControllerImpl) ListAttackPatterns(ctx *gin.Context) {
	var req dto.AttackPatternListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.Error(ctx, model.NewAPIError(http.StatusBadRequest, "请求参数错误", err), false)
		return
	}

	resp, err := c.service.ListAttackPatterns(ctx.Request.Context(), &req)
	if err != nil {
		c.logger.Error().Err(err).Msg("查询攻击模式列表失败")
		response.Error(ctx, model.NewAPIError(http.StatusInternalServerError, "查询攻击模式列表失败", err), false)
		return
	}

	response.Success(ctx, "查询成功", resp)
}

// GetAttackPattern 获取攻击模式详情
// @Summary 获取攻击模式详情
// @Tags AI分析器
// @Accept json
// @Produce json
// @Param id path string true "模式ID"
// @Success 200 {object} model.AttackPattern
// @Router /api/v1/ai-analyzer/patterns/{id} [get]
func (c *AIAnalyzerControllerImpl) GetAttackPattern(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		response.Error(ctx, model.NewAPIError(http.StatusBadRequest, "模式ID不能为空", nil), false)
		return
	}

	pattern, err := c.service.GetAttackPattern(ctx.Request.Context(), id)
	if err != nil {
		c.logger.Error().Err(err).Str("id", id).Msg("查询攻击模式失败")
		response.Error(ctx, model.NewAPIError(http.StatusInternalServerError, "查询攻击模式失败", err), false)
		return
	}

	response.Success(ctx, "查询成功", pattern)
}

// DeleteAttackPattern 删除攻击模式
// @Summary 删除攻击模式
// @Tags AI分析器
// @Accept json
// @Produce json
// @Param id path string true "模式ID"
// @Success 200 {object} response.Response
// @Router /api/v1/ai-analyzer/patterns/{id} [delete]
func (c *AIAnalyzerControllerImpl) DeleteAttackPattern(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		response.Error(ctx, model.NewAPIError(http.StatusBadRequest, "模式ID不能为空", nil), false)
		return
	}

	err := c.service.DeleteAttackPattern(ctx.Request.Context(), id)
	if err != nil {
		c.logger.Error().Err(err).Str("id", id).Msg("删除攻击模式失败")
		response.Error(ctx, model.NewAPIError(http.StatusInternalServerError, "删除攻击模式失败", err), false)
		return
	}

	response.Success(ctx, "删除成功", nil)
}

// ============================================
// 生成规则相关
// ============================================

// ListGeneratedRules 获取生成规则列表
// @Summary 获取生成规则列表
// @Tags AI分析器
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(10)
// @Param status query string false "规则状态" Enums(pending, approved, rejected, deployed)
// @Param ruleType query string false "规则类型" Enums(modsecurity, microrule)
// @Param patternId query string false "关联的攻击模式ID"
// @Success 200 {object} dto.GeneratedRuleListResponse
// @Router /api/v1/ai-analyzer/rules [get]
func (c *AIAnalyzerControllerImpl) ListGeneratedRules(ctx *gin.Context) {
	var req dto.GeneratedRuleListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.Error(ctx, model.NewAPIError(http.StatusBadRequest, "请求参数错误", err), false)
		return
	}

	resp, err := c.service.ListGeneratedRules(ctx.Request.Context(), &req)
	if err != nil {
		c.logger.Error().Err(err).Msg("查询生成规则列表失败")
		response.Error(ctx, model.NewAPIError(http.StatusInternalServerError, "查询生成规则列表失败", err), false)
		return
	}

	response.Success(ctx, "查询成功", resp)
}

// GetGeneratedRule 获取生成规则详情
// @Summary 获取生成规则详情
// @Tags AI分析器
// @Accept json
// @Produce json
// @Param id path string true "规则ID"
// @Success 200 {object} model.GeneratedRule
// @Router /api/v1/ai-analyzer/rules/{id} [get]
func (c *AIAnalyzerControllerImpl) GetGeneratedRule(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		response.Error(ctx, model.NewAPIError(http.StatusBadRequest, "规则ID不能为空", nil), false)
		return
	}

	rule, err := c.service.GetGeneratedRule(ctx.Request.Context(), id)
	if err != nil {
		c.logger.Error().Err(err).Str("id", id).Msg("查询生成规则失败")
		response.Error(ctx, model.NewAPIError(http.StatusInternalServerError, "查询生成规则失败", err), false)
		return
	}

	response.Success(ctx, "查询成功", rule)
}

// DeleteGeneratedRule 删除生成规则
// @Summary 删除生成规则
// @Tags AI分析器
// @Accept json
// @Produce json
// @Param id path string true "规则ID"
// @Success 200 {object} response.Response
// @Router /api/v1/ai-analyzer/rules/{id} [delete]
func (c *AIAnalyzerControllerImpl) DeleteGeneratedRule(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		response.Error(ctx, model.NewAPIError(http.StatusBadRequest, "规则ID不能为空", nil), false)
		return
	}

	err := c.service.DeleteGeneratedRule(ctx.Request.Context(), id)
	if err != nil {
		c.logger.Error().Err(err).Str("id", id).Msg("删除生成规则失败")
		response.Error(ctx, model.NewAPIError(http.StatusInternalServerError, "删除生成规则失败", err), false)
		return
	}

	response.Success(ctx, "删除成功", nil)
}

// ReviewRule 审核规则
// @Summary 审核规则
// @Tags AI分析器
// @Accept json
// @Produce json
// @Param request body dto.ReviewRuleRequest true "审核请求"
// @Success 200 {object} response.Response
// @Router /api/v1/ai-analyzer/rules/review [post]
func (c *AIAnalyzerControllerImpl) ReviewRule(ctx *gin.Context) {
	var req dto.ReviewRuleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, model.NewAPIError(http.StatusBadRequest, "请求参数错误", err), false)
		return
	}

	// 从上下文获取用户信息
	username, exists := ctx.Get("username")
	if !exists {
		response.Error(ctx, model.NewAPIError(http.StatusUnauthorized, "未授权", nil), false)
		return
	}

	err := c.service.ReviewRule(ctx.Request.Context(), &req, username.(string))
	if err != nil {
		c.logger.Error().Err(err).Str("rule_id", req.RuleID).Msg("审核规则失败")
		response.Error(ctx, model.NewAPIError(http.StatusInternalServerError, "审核规则失败", err), false)
		return
	}

	response.Success(ctx, "审核成功", nil)
}

// GetPendingRules 获取待审核规则列表
// @Summary 获取待审核规则列表
// @Tags AI分析器
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(10)
// @Success 200 {object} response.Response
// @Router /api/v1/ai-analyzer/rules/pending [get]
func (c *AIAnalyzerControllerImpl) GetPendingRules(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(ctx.DefaultQuery("size", "10"))

	rules, total, err := c.service.GetPendingRules(ctx.Request.Context(), page, size)
	if err != nil {
		c.logger.Error().Err(err).Msg("查询待审核规则失败")
		response.Error(ctx, model.NewAPIError(http.StatusInternalServerError, "查询待审核规则失败", err), false)
		return
	}

	response.Success(ctx, "查询成功", gin.H{
		"list":  rules,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

// DeployRule 部署规则
// @Summary 部署规则
// @Tags AI分析器
// @Accept json
// @Produce json
// @Param id path string true "规则ID"
// @Success 200 {object} response.Response
// @Router /api/v1/ai-analyzer/rules/{id}/deploy [post]
func (c *AIAnalyzerControllerImpl) DeployRule(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		response.Error(ctx, model.NewAPIError(http.StatusBadRequest, "规则ID不能为空", nil), false)
		return
	}

	err := c.service.DeployRule(ctx.Request.Context(), id)
	if err != nil {
		c.logger.Error().Err(err).Str("id", id).Msg("部署规则失败")
		response.Error(ctx, model.NewAPIError(http.StatusInternalServerError, "部署规则失败", err), false)
		return
	}

	response.Success(ctx, "部署成功", nil)
}

// ============================================
// AI分析器配置相关
// ============================================

// GetAnalyzerConfig 获取AI分析器配置
// @Summary 获取AI分析器配置
// @Tags AI分析器
// @Accept json
// @Produce json
// @Success 200 {object} model.AIAnalyzerConfig
// @Router /api/v1/ai-analyzer/config [get]
func (c *AIAnalyzerControllerImpl) GetAnalyzerConfig(ctx *gin.Context) {
	config, err := c.service.GetAnalyzerConfig(ctx.Request.Context())
	if err != nil {
		c.logger.Error().Err(err).Msg("查询AI分析器配置失败")
		response.Error(ctx, model.NewAPIError(http.StatusInternalServerError, "查询AI分析器配置失败", err), false)
		return
	}

	response.Success(ctx, "查询成功", config)
}

// UpdateAnalyzerConfig 更新AI分析器配置
// @Summary 更新AI分析器配置
// @Tags AI分析器
// @Accept json
// @Produce json
// @Param request body dto.AIAnalyzerConfigRequest true "配置请求"
// @Success 200 {object} model.AIAnalyzerConfig
// @Router /api/v1/ai-analyzer/config [put]
func (c *AIAnalyzerControllerImpl) UpdateAnalyzerConfig(ctx *gin.Context) {
	var req dto.AIAnalyzerConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, model.NewAPIError(http.StatusBadRequest, "请求参数错误", err), false)
		return
	}

	config, err := c.service.UpdateAnalyzerConfig(ctx.Request.Context(), &req)
	if err != nil {
		c.logger.Error().Err(err).Msg("更新AI分析器配置失败")
		response.Error(ctx, model.NewAPIError(http.StatusInternalServerError, "更新AI分析器配置失败", err), false)
		return
	}

	response.Success(ctx, "更新成功", config)
}

// ============================================
// MCP对话相关
// ============================================

// ListMCPConversations 获取MCP对话列表
// @Summary 获取MCP对话列表
// @Tags AI分析器
// @Accept json
// @Produce json
// @Param patternId query string false "关联的攻击模式ID"
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(10)
// @Success 200 {object} response.Response
// @Router /api/v1/ai-analyzer/conversations [get]
func (c *AIAnalyzerControllerImpl) ListMCPConversations(ctx *gin.Context) {
	patternID := ctx.Query("patternId")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(ctx.DefaultQuery("size", "10"))

	var patternIDPtr *string
	if patternID != "" {
		patternIDPtr = &patternID
	}

	conversations, total, err := c.service.ListMCPConversations(ctx.Request.Context(), patternIDPtr, page, size)
	if err != nil {
		c.logger.Error().Err(err).Msg("查询MCP对话列表失败")
		response.Error(ctx, model.NewAPIError(http.StatusInternalServerError, "查询MCP对话列表失败", err), false)
		return
	}

	response.Success(ctx, "查询成功", gin.H{
		"list":  conversations,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

// GetMCPConversation 获取MCP对话详情
// @Summary 获取MCP对话详情
// @Tags AI分析器
// @Accept json
// @Produce json
// @Param id path string true "对话ID"
// @Success 200 {object} model.MCPConversation
// @Router /api/v1/ai-analyzer/conversations/{id} [get]
func (c *AIAnalyzerControllerImpl) GetMCPConversation(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		response.Error(ctx, model.NewAPIError(http.StatusBadRequest, "对话ID不能为空", nil), false)
		return
	}

	conversation, err := c.service.GetMCPConversation(ctx.Request.Context(), id)
	if err != nil {
		c.logger.Error().Err(err).Str("id", id).Msg("查询MCP对话失败")
		response.Error(ctx, model.NewAPIError(http.StatusInternalServerError, "查询MCP对话失败", err), false)
		return
	}

	response.Success(ctx, "查询成功", conversation)
}

// DeleteMCPConversation 删除MCP对话
// @Summary 删除MCP对话
// @Tags AI分析器
// @Accept json
// @Produce json
// @Param id path string true "对话ID"
// @Success 200 {object} response.Response
// @Router /api/v1/ai-analyzer/conversations/{id} [delete]
func (c *AIAnalyzerControllerImpl) DeleteMCPConversation(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		response.Error(ctx, model.NewAPIError(http.StatusBadRequest, "对话ID不能为空", nil), false)
		return
	}

	err := c.service.DeleteMCPConversation(ctx.Request.Context(), id)
	if err != nil {
		c.logger.Error().Err(err).Str("id", id).Msg("删除MCP对话失败")
		response.Error(ctx, model.NewAPIError(http.StatusInternalServerError, "删除MCP对话失败", err), false)
		return
	}

	response.Success(ctx, "删除成功", nil)
}

// ============================================
// 统计分析相关
// ============================================

// GetAnalyzerStats 获取AI分析器统计信息
// @Summary 获取AI分析器统计信息
// @Tags AI分析器
// @Accept json
// @Produce json
// @Param startTime query string false "开始时间(RFC3339)"
// @Param endTime query string false "结束时间(RFC3339)"
// @Success 200 {object} dto.AIAnalysisStatsResponse
// @Router /api/v1/ai-analyzer/stats [get]
func (c *AIAnalyzerControllerImpl) GetAnalyzerStats(ctx *gin.Context) {
	var req dto.TriggerAnalysisRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.Error(ctx, model.NewAPIError(http.StatusBadRequest, "请求参数错误", err), false)
		return
	}

	stats, err := c.service.GetAnalyzerStats(ctx.Request.Context(), &req)
	if err != nil {
		c.logger.Error().Err(err).Msg("查询AI分析器统计信息失败")
		response.Error(ctx, model.NewAPIError(http.StatusInternalServerError, "查询统计信息失败", err), false)
		return
	}

	response.Success(ctx, "查询成功", stats)
}

// TriggerAnalysis 手动触发AI分析
// @Summary 手动触发AI分析
// @Description 立即运行攻击模式检测和规则生成
// @Tags AI分析器
// @Accept json
// @Produce json
// @Success 200 {object} model.APIResponse
// @Failure 500 {object} model.APIError
// @Router /api/v1/ai-analyzer/trigger [post]
func (c *AIAnalyzerControllerImpl) TriggerAnalysis(ctx *gin.Context) {
	c.logger.Info().Msg("手动触发AI分析")
	
	// 异步执行分析任务
	go func() {
		bgCtx := context.Background()
		if err := c.service.TriggerAnalysis(bgCtx); err != nil {
			c.logger.Error().Err(err).Msg("AI分析执行失败")
		}
	}()
	
	response.Success(ctx, "AI分析已触发，请稍后查看结果", nil)
}
