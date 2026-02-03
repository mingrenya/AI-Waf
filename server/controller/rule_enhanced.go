package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mingrenya/AI-Waf/server/config"
	"github.com/mingrenya/AI-Waf/server/dto"
	"github.com/mingrenya/AI-Waf/server/model"
	"github.com/mingrenya/AI-Waf/server/service"
)

type RuleEnhancedController struct {
	templateService      service.RuleTemplateService
	effectivenessService service.RuleEffectivenessService
	protectionService    service.ProtectionProfileService
}

func NewRuleEnhancedController(
	templateService service.RuleTemplateService,
	effectivenessService service.RuleEffectivenessService,
	protectionService service.ProtectionProfileService,
) *RuleEnhancedController {
	return &RuleEnhancedController{
		templateService:      templateService,
		effectivenessService: effectivenessService,
		protectionService:    protectionService,
	}
}

// ============== 规则模板相关 ==============

// ListRuleTemplates 获取规则模板列表
// @Summary 获取规则模板列表
// @Description 获取OWASP Top 10规则模板列表，支持按分类和严重等级过滤
// @Tags 规则管理
// @Accept json
// @Produce json
// @Param category query string false "分类过滤"
// @Param severity query string false "严重等级过滤"
// @Success 200 {object} dto.RuleTemplateListResponse
// @Router /api/v1/rules/templates [get]
func (c *RuleEnhancedController) ListRuleTemplates(ctx *gin.Context) {
	var req dto.RuleTemplateListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	templates, err := c.templateService.ListTemplates(ctx.Request.Context(), req.Category, req.Severity)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	items := make([]dto.RuleTemplateResponse, len(templates))
	for i, t := range templates {
		items[i] = dto.ConvertRuleTemplateToResponse(&t)
	}

	response := dto.RuleTemplateListResponse{
		Total: int64(len(items)),
		Items: items,
	}

	successResp := model.NewSuccessResponse("获取规则模板列表成功", response)
	config.Logger.Info().
		Int("template_count", len(items)).
		Int64("total", response.Total).
		Interface("response", successResp).
		Msg("返回规则模板列表")

	ctx.JSON(http.StatusOK, successResp)
}

// GetRuleTemplate 获取规则模板详情
// @Summary 获取规则模板详情
// @Description 根据ID获取规则模板详情
// @Tags 规则管理
// @Accept json
// @Produce json
// @Param id path string true "模板ID"
// @Success 200 {object} dto.RuleTemplateResponse
// @Router /api/v1/rules/templates/{id} [get]
func (c *RuleEnhancedController) GetRuleTemplate(ctx *gin.Context) {
	id := ctx.Param("id")

	template, err := c.templateService.GetTemplate(ctx.Request.Context(), id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, model.NewErrorResponse(http.StatusNotFound, "模板不存在", err))
		return
	}

	ctx.JSON(http.StatusOK, model.NewSuccessResponse("获取模板详情成功", dto.ConvertRuleTemplateToResponse(template)))
}

// CreateRuleFromTemplate 从模板创建规则
// @Summary 从模板创建规则
// @Description 根据模板ID创建新的WAF规则
// @Tags 规则管理
// @Accept json
// @Produce json
// @Param request body dto.CreateRuleFromTemplateRequest true "创建请求"
// @Success 201 {object} dto.MicroRuleResponse
// @Router /api/v1/rules/templates/create-rule [post]
func (c *RuleEnhancedController) CreateRuleFromTemplate(ctx *gin.Context) {
	var req dto.CreateRuleFromTemplateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rule, err := c.templateService.CreateRuleFromTemplate(ctx.Request.Context(), req.TemplateID, req.CustomName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, model.NewErrorResponse(http.StatusInternalServerError, "创建规则失败", err))
		return
	}

	responseData := gin.H{
		"message": "规则创建成功",
		"rule_id": rule.ID.Hex(),
		"name":    rule.Name,
	}
	ctx.JSON(http.StatusCreated, model.NewSuccessResponse("规则创建成功", responseData))
}

// ============== 规则有效性评分相关 ==============

// CalculateRuleScore 计算规则评分
// @Summary 计算规则有效性评分
// @Description 根据规则的历史数据计算有效性评分
// @Tags 规则管理
// @Accept json
// @Produce json
// @Param request body dto.CalculateScoreRequest true "计算请求"
// @Success 200 {object} dto.RuleEffectivenessScoreResponse
// @Router /api/v1/rules/effectiveness/calculate [post]
func (c *RuleEnhancedController) CalculateRuleScore(ctx *gin.Context) {
	var req dto.CalculateScoreRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	score, err := c.effectivenessService.CalculateScore(ctx.Request.Context(), req.RuleID, req.Period)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, model.NewErrorResponse(http.StatusInternalServerError, "计算评分失败", err))
		return
	}

	ctx.JSON(http.StatusOK, model.NewSuccessResponse("计算评分成功", dto.ConvertRuleEffectivenessScoreToResponse(score)))
}

// GetRuleScore 获取规则评分
// @Summary 获取规则有效性评分
// @Description 获取指定规则的有效性评分
// @Tags 规则管理
// @Accept json
// @Produce json
// @Param id path string true "规则ID"
// @Success 200 {object} dto.RuleEffectivenessScoreResponse
// @Router /api/v1/rules/effectiveness/{id} [get]
func (c *RuleEnhancedController) GetRuleScore(ctx *gin.Context) {
	id := ctx.Param("id")

	score, err := c.effectivenessService.GetScore(ctx.Request.Context(), id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, model.NewErrorResponse(http.StatusNotFound, "评分记录不存在", err))
		return
	}

	ctx.JSON(http.StatusOK, model.NewSuccessResponse("获取评分成功", dto.ConvertRuleEffectivenessScoreToResponse(score)))
}

// ListRuleScores 获取所有规则评分
// @Summary 获取所有规则有效性评分
// @Description 获取所有规则的有效性评分列表，支持排序
// @Tags 规则管理
// @Accept json
// @Produce json
// @Param sortBy query string false "排序字段"
// @Param order query int false "排序方向"
// @Success 200 {object} dto.RuleEffectivenessScoreListResponse
// @Router /api/v1/rules/effectiveness [get]
func (c *RuleEnhancedController) ListRuleScores(ctx *gin.Context) {
	var req dto.ListScoresRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	scores, err := c.effectivenessService.ListScores(ctx.Request.Context(), req.SortBy, req.Order)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	items := make([]dto.RuleEffectivenessScoreResponse, len(scores))
	for i, s := range scores {
		items[i] = dto.ConvertRuleEffectivenessScoreToResponse(&s)
	}

	response := dto.RuleEffectivenessScoreListResponse{
		Total: int64(len(items)),
		Items: items,
	}
	ctx.JSON(http.StatusOK, model.NewSuccessResponse("获取评分列表成功", response))
}

// BatchCalculateScores 批量计算评分
// @Summary 批量计算规则有效性评分
// @Description 批量计算所有启用规则的有效性评分
// @Tags 规则管理
// @Accept json
// @Produce json
// @Param request body dto.BatchCalculateScoresRequest true "批量计算请求"
// @Success 200 {object} gin.H
// @Router /api/v1/rules/effectiveness/batch-calculate [post]
func (c *RuleEnhancedController) BatchCalculateScores(ctx *gin.Context) {
	var req dto.BatchCalculateScoresRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := c.effectivenessService.BatchCalculateScores(ctx.Request.Context(), req.Period)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, model.NewErrorResponse(http.StatusInternalServerError, "批量计算评分失败", err))
		return
	}

	ctx.JSON(http.StatusOK, model.NewSuccessResponse("批量计算评分成功", gin.H{"message": "批量计算评分成功"}))
}

// ============== 保护配置文件相关 ==============

// ListProtectionProfiles 获取保护配置文件列表
// @Summary 获取保护配置文件列表
// @Description 获取所有预定义的保护配置文件（基础/标准/严格）
// @Tags 规则管理
// @Accept json
// @Produce json
// @Success 200 {object} dto.ProtectionProfileListResponse
// @Router /api/v1/rules/profiles [get]
func (c *RuleEnhancedController) ListProtectionProfiles(ctx *gin.Context) {
	profiles, err := c.protectionService.ListProfiles(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	items := make([]dto.ProtectionProfileResponse, len(profiles))
	for i, p := range profiles {
		items[i] = dto.ConvertProtectionProfileToResponse(&p)
	}

	response := dto.ProtectionProfileListResponse{
		Total: int64(len(items)),
		Items: items,
	}
	ctx.JSON(http.StatusOK, model.NewSuccessResponse("获取保护配置列表成功", response))
}

// GetProtectionProfile 获取保护配置文件详情
// @Summary 获取保护配置文件详情
// @Description 根据ID获取保护配置文件详情
// @Tags 规则管理
// @Accept json
// @Produce json
// @Param id path string true "配置文件ID"
// @Success 200 {object} dto.ProtectionProfileResponse
// @Router /api/v1/rules/profiles/{id} [get]
func (c *RuleEnhancedController) GetProtectionProfile(ctx *gin.Context) {
	id := ctx.Param("id")

	profile, err := c.protectionService.GetProfile(ctx.Request.Context(), id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, model.NewErrorResponse(http.StatusNotFound, "配置文件不存在", err))
		return
	}

	ctx.JSON(http.StatusOK, model.NewSuccessResponse("获取配置详情成功", dto.ConvertProtectionProfileToResponse(profile)))
}

// ApplyProtectionProfile 应用保护配置文件
// @Summary 一键应用保护配置文件
// @Description 根据配置文件批量创建WAF规则
// @Tags 规则管理
// @Accept json
// @Produce json
// @Param request body dto.ApplyProfileRequest true "应用请求"
// @Success 200 {object} dto.ApplyProfileResponse
// @Router /api/v1/rules/profiles/apply [post]
func (c *RuleEnhancedController) ApplyProtectionProfile(ctx *gin.Context) {
	var req dto.ApplyProfileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdCount, err := c.protectionService.ApplyProfile(ctx.Request.Context(), req.ProfileID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, model.NewErrorResponse(http.StatusInternalServerError, "应用配置失败", err))
		return
	}

	response := dto.ApplyProfileResponse{
		CreatedCount: createdCount,
		Message:      "配置文件应用成功",
	}
	ctx.JSON(http.StatusOK, model.NewSuccessResponse("配置文件应用成功", response))
}
