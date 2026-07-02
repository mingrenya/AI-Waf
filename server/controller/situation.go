package controller

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/mingrenya/AI-Waf/pkg/model"
	"github.com/mingrenya/AI-Waf/server/config"
	"github.com/mingrenya/AI-Waf/server/dto"
	"github.com/mingrenya/AI-Waf/server/repository"
	"github.com/mingrenya/AI-Waf/server/service/situation"
	"github.com/mingrenya/AI-Waf/server/utils/response"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// SituationController 态势感知控制器接口
type SituationController interface {
	GetOverview(ctx *gin.Context)
	ListChains(ctx *gin.Context)
	GetChainDetail(ctx *gin.Context)
	ListAttackers(ctx *gin.Context)
	GetAttackerProfile(ctx *gin.Context)
	GetTrends(ctx *gin.Context)
	ListRules(ctx *gin.Context)
	CreateRule(ctx *gin.Context)
	UpdateRule(ctx *gin.Context)
	DeleteRule(ctx *gin.Context)
	QuickAction(ctx *gin.Context)
}

// SituationControllerImpl 态势感知控制器实现
type SituationControllerImpl struct {
	repo           repository.SituationRepository
	quickActionSvc *situation.QuickActionService
	logger         zerolog.Logger
}

// NewSituationController 创建态势感知控制器
func NewSituationController(repo repository.SituationRepository, quickActionSvc *situation.QuickActionService) SituationController {
	return &SituationControllerImpl{
		repo:           repo,
		quickActionSvc: quickActionSvc,
		logger:         config.GetControllerLogger("situation"),
	}
}

// GetOverview 获取态势概览
func (c *SituationControllerImpl) GetOverview(ctx *gin.Context) {
	chains, total, err := c.repo.ListChains(ctx, bson.M{"active": true}, 0, 5)
	if err != nil {
		response.InternalServerError(ctx, err, false)
		return
	}
	// 聚合攻击类型 Top 5（简化实现：从活跃链中统计）
	typeCount := make(map[string]int)
	ipCount := make(map[string]int)
	countryCount := make(map[string]int)
	for _, chain := range chains {
		ipCount[chain.SourceIP]++
		if chain.GeoCountry != "" && chain.GeoCountry != "unknown" {
			countryCount[chain.GeoCountry]++
		}
		for _, s := range chain.Stages {
			typeCount[string(s.Stage)]++
		}
	}
	// 转 CountItem
	topTypes := topN(typeCount, 5)
	topIPs := topN(ipCount, 5)
	topCountries := topN(countryCount, 5)

	overview := dto.SituationOverviewResponse{
		ActiveChains:      len(chains),
		TotalChains24h:    int(total),
		TotalAttackers24h: len(ipCount),
		OverallRiskScore:  0,
		RiskTrend:         "stable",
		TopAttackTypes:    topTypes,
		TopAttackerIPs:    topIPs,
		TopTargetSites:    []dto.CountItem{},
		ByCountry:         topCountries,
	}
	response.Success(ctx, "获取态势概览成功", overview)
}

// ListChains 获取攻击链列表
func (c *SituationControllerImpl) ListChains(ctx *gin.Context) {
	var req dto.ChainListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.BadRequest(ctx, err, true)
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	filter := bson.M{}
	if req.SourceIP != "" {
		filter["source_ip"] = req.SourceIP
	}
	if req.Stage != "" {
		filter["stages.stage"] = req.Stage
	}
	if req.Active != nil {
		filter["active"] = *req.Active
	}
	skip := int64((req.Page - 1) * req.PageSize)
	limit := int64(req.PageSize)

	chains, total, err := c.repo.ListChains(ctx, filter, skip, limit)
	if err != nil {
		response.InternalServerError(ctx, err, false)
		return
	}
	summaries := make([]dto.ChainSummary, 0, len(chains))
	for _, ch := range chains {
		stages := make([]string, 0, len(ch.Stages))
		for _, s := range ch.Stages {
			stages = append(stages, string(s.Stage))
		}
		summaries = append(summaries, dto.ChainSummary{
			ID: ch.ID, SourceIP: ch.SourceIP, GeoCountry: ch.GeoCountry,
			Stages: stages, RiskScore: ch.RiskScore,
			FirstSeen: ch.FirstSeen, LastSeen: ch.LastSeen, Active: ch.Active,
		})
	}
	response.Success(ctx, "获取攻击链列表成功", dto.ChainListResponse{
		Chains: summaries, Total: total, Page: req.Page, PageSize: req.PageSize,
	})
}

// GetChainDetail 获取攻击链详情
func (c *SituationControllerImpl) GetChainDetail(ctx *gin.Context) {
	id := ctx.Param("id")
	chain, err := c.repo.GetChainByID(ctx, id)
	if err != nil {
		response.InternalServerError(ctx, err, false)
		return
	}
	if chain == nil {
		response.NotFound(ctx, errors.New("攻击链不存在"))
		return
	}
	stages := make([]dto.ChainStageItem, 0, len(chain.Stages))
	for _, s := range chain.Stages {
		stages = append(stages, dto.ChainStageItem{
			Stage: string(s.Stage), Technique: s.Technique,
			DetectedAt: s.DetectedAt, Confidence: s.Confidence, Evidence: s.Evidence,
		})
	}
	detail := dto.ChainDetailResponse{
		ID: chain.ID, SourceIP: chain.SourceIP, GeoCountry: chain.GeoCountry,
		Stages: stages, CorrelationIDs: chain.CorrelationIDs,
		RiskScore: chain.RiskScore, RiskLabel: situation.RiskLabel(chain.RiskScore),
		FirstSeen: chain.FirstSeen, LastSeen: chain.LastSeen, Active: chain.Active,
	}
	// 尝试加载画像
	profile, _ := c.repo.GetProfile(ctx, chain.SourceIP)
	if profile != nil {
		detail.AttackerProfile = &dto.AttackerProfileDetail{
			SourceIP: profile.SourceIP, GeoCountry: profile.GeoCountry, GeoCity: profile.GeoCity,
			TotalAttacks: profile.TotalAttacks, UniqueAttackTypes: profile.UniqueAttackTypes,
			TopAttackType: profile.TopAttackType, UniqueTargetSites: profile.UniqueTargetSites,
			ActiveHours: profile.ActiveHours, AttackPhase: profile.AttackPhase,
			ToolsIdentified: profile.ToolsIdentified, IsAutomated: profile.IsAutomated,
			IsPersistent: profile.IsPersistent, RiskScore: profile.RiskScore,
			RiskLabel: profile.RiskLabel, FirstSeen: profile.FirstSeen, LastSeen: profile.LastSeen,
		}
	}
	response.Success(ctx, "获取攻击链详情成功", detail)
}

// ListAttackers 获取攻击者列表
func (c *SituationControllerImpl) ListAttackers(ctx *gin.Context) {
	var req dto.AttackerListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.BadRequest(ctx, err, true)
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	skip := int64((req.Page - 1) * req.PageSize)
	limit := int64(req.PageSize)

	profiles, total, err := c.repo.ListProfiles(ctx, req.SortBy, skip, limit)
	if err != nil {
		response.InternalServerError(ctx, err, false)
		return
	}
	summaries := make([]dto.AttackerProfileSummary, 0, len(profiles))
	for _, p := range profiles {
		summaries = append(summaries, dto.AttackerProfileSummary{
			SourceIP: p.SourceIP, GeoCountry: p.GeoCountry,
			TotalAttacks: p.TotalAttacks, UniqueAttackTypes: p.UniqueAttackTypes,
			TopAttackType: p.TopAttackType, AttackPhase: p.AttackPhase,
			RiskScore: p.RiskScore, RiskLabel: p.RiskLabel, LastSeen: p.LastSeen,
		})
	}
	response.Success(ctx, "获取攻击者列表成功", dto.AttackerListResponse{
		Attackers: summaries, Total: total,
	})
}

// GetAttackerProfile 获取攻击者详情
func (c *SituationControllerImpl) GetAttackerProfile(ctx *gin.Context) {
	ip := ctx.Param("ip")
	profile, err := c.repo.GetProfile(ctx, ip)
	if err != nil {
		response.InternalServerError(ctx, err, false)
		return
	}
	if profile == nil {
		response.NotFound(ctx, errors.New("无此攻击者记录"))
		return
	}
	detail := dto.AttackerProfileDetail{
		SourceIP: profile.SourceIP, GeoCountry: profile.GeoCountry, GeoCity: profile.GeoCity,
		TotalAttacks: profile.TotalAttacks, UniqueAttackTypes: profile.UniqueAttackTypes,
		TopAttackType: profile.TopAttackType, UniqueTargetSites: profile.UniqueTargetSites,
		ActiveHours: profile.ActiveHours, AttackPhase: profile.AttackPhase,
		ToolsIdentified: profile.ToolsIdentified, IsAutomated: profile.IsAutomated,
		IsPersistent: profile.IsPersistent, RiskScore: profile.RiskScore,
		RiskLabel: profile.RiskLabel, FirstSeen: profile.FirstSeen, LastSeen: profile.LastSeen,
	}
	response.Success(ctx, "获取攻击者画像成功", detail)
}

// GetTrends 获取趋势数据
func (c *SituationControllerImpl) GetTrends(ctx *gin.Context) {
	duration := ctx.DefaultQuery("duration", "24h")
	chains, total, _ := c.repo.ListChains(ctx, bson.M{"active": true}, 0, 100)

	_ = duration
	_ = total

	// 简化的趋势统计
	response.Success(ctx, "获取趋势数据成功", dto.TrendResponse{
		Timeline:        []dto.TrendPoint{},
		FrequentTypes:   []dto.CountItem{},
		ActiveAttackers: len(chains),
		NewChains24h:    int(total),
	})
}

// ListRules 获取规则列表
func (c *SituationControllerImpl) ListRules(ctx *gin.Context) {
	rules, err := c.repo.ListRules(ctx)
	if err != nil {
		response.InternalServerError(ctx, err, false)
		return
	}
	response.Success(ctx, "获取规则列表成功", rules)
}

// CreateRule 创建规则
func (c *SituationControllerImpl) CreateRule(ctx *gin.Context) {
	var req dto.SituationRuleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, err, true)
		return
	}
	rule := &model.SituationRule{
		Name: req.Name, Stage: req.Stage, LogQL: req.LogQL,
		Interval: req.Interval, Threshold: req.Threshold,
		Severity: req.Severity, MITRETactic: req.MITRETactic,
		MITRETechnique: req.MITRETechnique, Enabled: req.Enabled,
	}
	if err := c.repo.CreateRule(ctx, rule); err != nil {
		response.InternalServerError(ctx, err, false)
		return
	}
	response.Success(ctx, "创建规则成功", rule)
}

// UpdateRule 更新规则
func (c *SituationControllerImpl) UpdateRule(ctx *gin.Context) {
	id := ctx.Param("id")
	var req dto.SituationRuleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, err, true)
		return
	}
	rule := &model.SituationRule{
		Name: req.Name, Stage: req.Stage, LogQL: req.LogQL,
		Interval: req.Interval, Threshold: req.Threshold,
		Severity: req.Severity, MITRETactic: req.MITRETactic,
		MITRETechnique: req.MITRETechnique, Enabled: req.Enabled,
	}
	if err := c.repo.UpdateRule(ctx, id, rule); err != nil {
		response.InternalServerError(ctx, err, false)
		return
	}
	response.Success(ctx, "更新规则成功", rule)
}

// DeleteRule 删除规则
func (c *SituationControllerImpl) DeleteRule(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := c.repo.DeleteRule(ctx, id); err != nil {
		response.InternalServerError(ctx, err, false)
		return
	}
	response.Success(ctx, "删除规则成功", nil)
}

// QuickAction 一键处置
func (c *SituationControllerImpl) QuickAction(ctx *gin.Context) {
	var req dto.QuickActionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, err, true)
		return
	}
	result, err := c.quickActionSvc.ExecuteQuickAction(ctx, situation.QuickActionRequest{
		SourceIP: req.SourceIP, Action: req.Action,
		DurationHours: req.DurationHours, Reason: req.Reason, CorrelationID: req.CorrelationID,
	})
	if err != nil {
		response.InternalServerError(ctx, err, false)
		return
	}
	response.Success(ctx, "处置成功", dto.QuickActionResponse{
		Success: result.Success, SourceIP: result.SourceIP, Action: result.Action,
		Blocked: result.Blocked, Blacklisted: result.Blacklisted, Note: result.Note,
	})
}

func topN(m map[string]int, n int) []dto.CountItem {
	items := make([]dto.CountItem, 0, len(m))
	for k, v := range m {
		items = append(items, dto.CountItem{Label: k, Count: v})
	}
	for i := 0; i < len(items) && i < n; i++ {
		for j := i + 1; j < len(items); j++ {
			if items[j].Count > items[i].Count {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
	if len(items) > n {
		items = items[:n]
	}
	return items
}
