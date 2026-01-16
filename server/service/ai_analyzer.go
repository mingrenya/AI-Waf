// server/service/ai_analyzer.go
package service

import (
	"context"
	"errors"
	"time"

	"github.com/mingrenya/AI-Waf/pkg/model"
	"github.com/mingrenya/AI-Waf/server/config"
	"github.com/mingrenya/AI-Waf/server/dto"
	"github.com/mingrenya/AI-Waf/server/repository"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	ErrInvalidStatus            = errors.New("无效的规则状态")
	ErrInvalidSeverity          = errors.New("无效的严重程度")
	ErrRuleNotPending           = errors.New("规则不在待审核状态")
	ErrInvalidTimeRange         = errors.New("无效的时间范围")
	ErrMCPConfigNotSet          = errors.New("MCP配置未设置")
)

// AIAnalyzerService AI分析器服务接口
type AIAnalyzerService interface {
	// 攻击模式相关
	ListAttackPatterns(ctx context.Context, req *dto.AttackPatternListRequest) (*dto.AttackPatternListResponse, error)
	GetAttackPattern(ctx context.Context, id string) (*model.AttackPattern, error)
	DeleteAttackPattern(ctx context.Context, id string) error
	GetPatternsBySeverity(ctx context.Context, severity string, limit int) ([]model.AttackPattern, error)
	GetPatternsByTimeRange(ctx context.Context, startTime, endTime time.Time, page, size int) ([]model.AttackPattern, int64, error)

	// 生成规则相关
	ListGeneratedRules(ctx context.Context, req *dto.GeneratedRuleListRequest) (*dto.GeneratedRuleListResponse, error)
	GetGeneratedRule(ctx context.Context, id string) (*model.GeneratedRule, error)
	DeleteGeneratedRule(ctx context.Context, id string) error
	ReviewRule(ctx context.Context, req *dto.ReviewRuleRequest, username string) error
	GetPendingRules(ctx context.Context, page, size int) ([]model.GeneratedRule, int64, error)
	DeployRule(ctx context.Context, id string) error

	// AI分析器配置相关
	GetAnalyzerConfig(ctx context.Context) (*model.AIAnalyzerConfig, error)
	UpdateAnalyzerConfig(ctx context.Context, req *dto.AIAnalyzerConfigRequest) (*model.AIAnalyzerConfig, error)

	// MCP对话相关
	ListMCPConversations(ctx context.Context, patternID *string, page, size int) ([]model.MCPConversation, int64, error)
	GetMCPConversation(ctx context.Context, id string) (*model.MCPConversation, error)
	DeleteMCPConversation(ctx context.Context, id string) error

	// 统计分析相关
	GetAnalyzerStats(ctx context.Context, req *dto.TriggerAnalysisRequest) (*dto.AIAnalysisStatsResponse, error)
	
	// 手动触发AI分析
	TriggerAnalysis(ctx context.Context) error
}

// AIAnalyzerServiceImpl AI分析器服务实现
type AIAnalyzerServiceImpl struct {
	patternRepo      repository.AttackPatternRepository
	ruleRepo         repository.GeneratedRuleRepository
	configRepo       repository.AIAnalyzerConfigRepository
	conversationRepo repository.MCPConversationRepository
	logger           zerolog.Logger
}

// NewAIAnalyzerService 创建AI分析器服务
func NewAIAnalyzerService(
	patternRepo repository.AttackPatternRepository,
	ruleRepo repository.GeneratedRuleRepository,
	configRepo repository.AIAnalyzerConfigRepository,
	conversationRepo repository.MCPConversationRepository,
) AIAnalyzerService {
	logger := config.GetServiceLogger("ai_analyzer")
	return &AIAnalyzerServiceImpl{
		patternRepo:      patternRepo,
		ruleRepo:         ruleRepo,
		configRepo:       configRepo,
		conversationRepo: conversationRepo,
		logger:           logger,
	}
}

// ============================================
// 攻击模式相关
// ============================================

func (s *AIAnalyzerServiceImpl) ListAttackPatterns(ctx context.Context, req *dto.AttackPatternListRequest) (*dto.AttackPatternListResponse, error) {
	// 验证分页参数
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Size < 1 || req.Size > 100 {
		req.Size = 10
	}

	// 构建查询过滤器
	filter := bson.D{}
	if req.Severity != "" {
		filter = append(filter, bson.E{Key: "severity", Value: req.Severity})
	}
	if req.PatternType != "" {
		filter = append(filter, bson.E{Key: "patternType", Value: req.PatternType})
	}
	if req.AttackType != "" {
		// AttackType与PatternType是同一个字段的别名
		filter = append(filter, bson.E{Key: "patternType", Value: req.AttackType})
	}
	if req.Status != "" {
		filter = append(filter, bson.E{Key: "status", Value: req.Status})
	}
	if req.StartTime != nil && req.EndTime != nil {
		filter = append(filter, bson.E{Key: "createdAt", Value: bson.D{
			{Key: "$gte", Value: *req.StartTime},
			{Key: "$lte", Value: *req.EndTime},
		}})
	}

	// 查询数据
	patterns, total, err := s.patternRepo.List(ctx, filter, int64(req.Page), int64(req.Size))
	if err != nil {
		s.logger.Error().Err(err).Msg("查询攻击模式列表失败")
		return nil, err
	}

	return &dto.AttackPatternListResponse{
		List:  patterns,
		Total: total,
		Page:  req.Page,
		Size:  req.Size,
	}, nil
}

func (s *AIAnalyzerServiceImpl) GetAttackPattern(ctx context.Context, id string) (*model.AttackPattern, error) {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("无效的ID格式")
	}

	pattern, err := s.patternRepo.GetByID(ctx, objectID)
	if err != nil {
		s.logger.Error().Err(err).Str("id", id).Msg("查询攻击模式失败")
		return nil, err
	}

	return pattern, nil
}

func (s *AIAnalyzerServiceImpl) DeleteAttackPattern(ctx context.Context, id string) error {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("无效的ID格式")
	}

	err = s.patternRepo.Delete(ctx, objectID)
	if err != nil {
		s.logger.Error().Err(err).Str("id", id).Msg("删除攻击模式失败")
		return err
	}

	return nil
}

func (s *AIAnalyzerServiceImpl) GetPatternsBySeverity(ctx context.Context, severity string, limit int) ([]model.AttackPattern, error) {
	if severity != "critical" && severity != "high" && severity != "medium" && severity != "low" {
		return nil, ErrInvalidSeverity
	}

	patterns, err := s.patternRepo.GetBySeverity(ctx, severity, int64(limit))
	if err != nil {
		s.logger.Error().Err(err).Str("severity", severity).Msg("按严重程度查询攻击模式失败")
		return nil, err
	}

	return patterns, nil
}

func (s *AIAnalyzerServiceImpl) GetPatternsByTimeRange(ctx context.Context, startTime, endTime time.Time, page, size int) ([]model.AttackPattern, int64, error) {
	if startTime.After(endTime) {
		return nil, 0, ErrInvalidTimeRange
	}

	patterns, total, err := s.patternRepo.GetByTimeRange(ctx, startTime, endTime, int64(page), int64(size))
	if err != nil {
		s.logger.Error().Err(err).Msg("按时间范围查询攻击模式失败")
		return nil, 0, err
	}

	return patterns, total, nil
}

// ============================================
// 生成规则相关
// ============================================

func (s *AIAnalyzerServiceImpl) ListGeneratedRules(ctx context.Context, req *dto.GeneratedRuleListRequest) (*dto.GeneratedRuleListResponse, error) {
	// 验证分页参数
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Size < 1 || req.Size > 100 {
		req.Size = 10
	}

	// 构建查询过滤器
	filter := bson.D{}
	if req.Status != "" {
		filter = append(filter, bson.E{Key: "status", Value: req.Status})
	}
	if req.RuleType != "" {
		filter = append(filter, bson.E{Key: "ruleType", Value: req.RuleType})
	}
	if req.PatternID != "" {
		// PatternID存储为字符串，不是ObjectID
		filter = append(filter, bson.E{Key: "patternId", Value: req.PatternID})
	}

	// 查询数据
	rules, total, err := s.ruleRepo.List(ctx, filter, int64(req.Page), int64(req.Size))
	if err != nil {
		s.logger.Error().Err(err).Msg("查询生成规则列表失败")
		return nil, err
	}

	return &dto.GeneratedRuleListResponse{
		List:  rules,
		Total: total,
		Page:  req.Page,
		Size:  req.Size,
	}, nil
}

func (s *AIAnalyzerServiceImpl) GetGeneratedRule(ctx context.Context, id string) (*model.GeneratedRule, error) {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("无效的ID格式")
	}

	rule, err := s.ruleRepo.GetByID(ctx, objectID)
	if err != nil {
		s.logger.Error().Err(err).Str("id", id).Msg("查询生成规则失败")
		return nil, err
	}

	return rule, nil
}

func (s *AIAnalyzerServiceImpl) DeleteGeneratedRule(ctx context.Context, id string) error {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("无效的ID格式")
	}

	err = s.ruleRepo.Delete(ctx, objectID)
	if err != nil {
		s.logger.Error().Err(err).Str("id", id).Msg("删除生成规则失败")
		return err
	}

	return nil
}

func (s *AIAnalyzerServiceImpl) ReviewRule(ctx context.Context, req *dto.ReviewRuleRequest, username string) error {
	// 验证ID
	objectID, err := bson.ObjectIDFromHex(req.RuleID)
	if err != nil {
		return errors.New("无效的规则ID格式")
	}

	// 验证状态
	if req.Action != "approve" && req.Action != "reject" {
		return ErrInvalidStatus
	}

	// 查询规则
	rule, err := s.ruleRepo.GetByID(ctx, objectID)
	if err != nil {
		return err
	}

	// 检查规则状态
	if rule.Status != "pending" {
		return ErrRuleNotPending
	}

	// 更新状态
	newStatus := "approved"
	if req.Action == "reject" {
		newStatus = "rejected"
	}

	err = s.ruleRepo.UpdateStatus(ctx, objectID, newStatus, username, req.Comment)
	if err != nil {
		s.logger.Error().Err(err).Str("rule_id", req.RuleID).Msg("更新规则状态失败")
		return err
	}

	s.logger.Info().
		Str("rule_id", req.RuleID).
		Str("action", req.Action).
		Str("user", username).
		Msg("规则审核完成")

	return nil
}

func (s *AIAnalyzerServiceImpl) GetPendingRules(ctx context.Context, page, size int) ([]model.GeneratedRule, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	rules, total, err := s.ruleRepo.GetPendingReview(ctx, int64(page), int64(size))
	if err != nil {
		s.logger.Error().Err(err).Msg("查询待审核规则失败")
		return nil, 0, err
	}

	return rules, total, nil
}

func (s *AIAnalyzerServiceImpl) DeployRule(ctx context.Context, id string) error {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("无效的ID格式")
	}

	// 查询规则
	rule, err := s.ruleRepo.GetByID(ctx, objectID)
	if err != nil {
		return err
	}

	// 检查规则状态
	if rule.Status != "approved" {
		return errors.New("只能部署已审核通过的规则")
	}

	// 更新为已部署状态
	rule.Status = "deployed"
	rule.DeployedAt = time.Now()
	err = s.ruleRepo.Update(ctx, rule)
	if err != nil {
		s.logger.Error().Err(err).Str("rule_id", id).Msg("部署规则失败")
		return err
	}

	s.logger.Info().Str("rule_id", id).Msg("规则部署成功")
	return nil
}

// ============================================
// AI分析器配置相关
// ============================================

func (s *AIAnalyzerServiceImpl) GetAnalyzerConfig(ctx context.Context) (*model.AIAnalyzerConfig, error) {
	config, err := s.configRepo.Get(ctx)
	if err != nil {
		// 如果配置不存在，创建默认配置
		if errors.Is(err, repository.ErrAIAnalyzerConfigNotFound) {
			if err := s.configRepo.CreateDefault(ctx); err != nil {
				s.logger.Error().Err(err).Msg("创建默认AI分析器配置失败")
				return nil, err
			}
			// 重新查询
			config, err = s.configRepo.Get(ctx)
			if err != nil {
				return nil, err
			}
		} else {
			s.logger.Error().Err(err).Msg("查询AI分析器配置失败")
			return nil, err
		}
	}

	return config, nil
}

func (s *AIAnalyzerServiceImpl) UpdateAnalyzerConfig(ctx context.Context, req *dto.AIAnalyzerConfigRequest) (*model.AIAnalyzerConfig, error) {
	// 获取当前配置
	config, err := s.GetAnalyzerConfig(ctx)
	if err != nil {
		return nil, err
	}

	// 更新配置字段
	config.Enabled = req.Enabled
	if req.AnalysisInterval != 0 {
		if req.AnalysisInterval < 5 || req.AnalysisInterval > 1440 {
			return nil, errors.New("分析间隔必须在5-1440分钟之间")
		}
		config.AnalysisInterval = req.AnalysisInterval
	}
	if req.PatternDetection.MinSamples != 0 {
		if req.PatternDetection.MinSamples < 10 {
			return nil, errors.New("最小日志数量不能少于10")
		}
		config.PatternDetection.MinSamples = req.PatternDetection.MinSamples
	}
	if req.PatternDetection.AnomalyThreshold != 0 {
		if req.PatternDetection.AnomalyThreshold < 1.0 || req.PatternDetection.AnomalyThreshold > 5.0 {
			return nil, errors.New("异常阈值必须在1.0-5.0之间")
		}
		config.PatternDetection.AnomalyThreshold = req.PatternDetection.AnomalyThreshold
	}
	if req.PatternDetection.ClusteringMethod != "" {
		config.PatternDetection.ClusteringMethod = req.PatternDetection.ClusteringMethod
	}
	if req.PatternDetection.TimeWindow != 0 {
		config.PatternDetection.TimeWindow = req.PatternDetection.TimeWindow
	}
	config.PatternDetection.Enabled = req.PatternDetection.Enabled
	
	// 更新规则生成配置
	if req.RuleGeneration.ConfidenceThreshold != 0 {
		if req.RuleGeneration.ConfidenceThreshold < 0.5 || req.RuleGeneration.ConfidenceThreshold > 1.0 {
			return nil, errors.New("置信度阈值必须在0.5-1.0之间")
		}
		config.RuleGeneration.ConfidenceThreshold = req.RuleGeneration.ConfidenceThreshold
	}
	if req.RuleGeneration.DefaultAction != "" {
		config.RuleGeneration.DefaultAction = req.RuleGeneration.DefaultAction
	}
	config.RuleGeneration.Enabled = req.RuleGeneration.Enabled
	config.RuleGeneration.AutoDeploy = req.RuleGeneration.AutoDeploy
	config.RuleGeneration.ReviewRequired = req.RuleGeneration.ReviewRequired

	// 保存配置
	err = s.configRepo.Update(ctx, config)
	if err != nil {
		s.logger.Error().Err(err).Msg("更新AI分析器配置失败")
		return nil, err
	}

	s.logger.Info().Msg("AI分析器配置更新成功")
	return config, nil
}

// ============================================
// MCP对话相关
// ============================================

func (s *AIAnalyzerServiceImpl) ListMCPConversations(ctx context.Context, patternID *string, page, size int) ([]model.MCPConversation, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	var patternObjectID *bson.ObjectID
	if patternID != nil && *patternID != "" {
		objectID, err := bson.ObjectIDFromHex(*patternID)
		if err != nil {
			return nil, 0, errors.New("无效的模式ID格式")
		}
		patternObjectID = &objectID
	}

	conversations, total, err := s.conversationRepo.List(ctx, patternObjectID, int64(page), int64(size))
	if err != nil {
		s.logger.Error().Err(err).Msg("查询MCP对话列表失败")
		return nil, 0, err
	}

	return conversations, total, nil
}

func (s *AIAnalyzerServiceImpl) GetMCPConversation(ctx context.Context, id string) (*model.MCPConversation, error) {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("无效的ID格式")
	}

	conversation, err := s.conversationRepo.GetByID(ctx, objectID)
	if err != nil {
		s.logger.Error().Err(err).Str("id", id).Msg("查询MCP对话失败")
		return nil, err
	}

	return conversation, nil
}

func (s *AIAnalyzerServiceImpl) DeleteMCPConversation(ctx context.Context, id string) error {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("无效的ID格式")
	}

	err = s.conversationRepo.Delete(ctx, objectID)
	if err != nil {
		s.logger.Error().Err(err).Str("id", id).Msg("删除MCP对话失败")
		return err
	}

	return nil
}

// ============================================
// 统计分析相关
// ============================================

func (s *AIAnalyzerServiceImpl) GetAnalyzerStats(ctx context.Context, req *dto.TriggerAnalysisRequest) (*dto.AIAnalysisStatsResponse, error) {
	// 获取配置
	config, err := s.GetAnalyzerConfig(ctx)
	if err != nil {
		return nil, err
	}

	// 查询攻击模式统计
	totalPatterns, err := s.patternRepo.Count(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	
	activePatterns, err := s.patternRepo.Count(ctx, bson.D{{Key: "status", Value: "active"}})
	if err != nil {
		return nil, err
	}

	// 查询生成规则统计
	totalRules, err := s.ruleRepo.Count(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	
	pendingRules, err := s.ruleRepo.Count(ctx, bson.D{{Key: "status", Value: "pending"}})
	if err != nil {
		return nil, err
	}
	
	approvedRules, err := s.ruleRepo.Count(ctx, bson.D{{Key: "status", Value: "approved"}})
	if err != nil {
		return nil, err
	}
	
	deployedRules, err := s.ruleRepo.Count(ctx, bson.D{{Key: "status", Value: "deployed"}})
	if err != nil {
		return nil, err
	}
	
	rejectedRules, err := s.ruleRepo.Count(ctx, bson.D{{Key: "status", Value: "rejected"}})
	if err != nil {
		return nil, err
	}

	return &dto.AIAnalysisStatsResponse{
		Enabled:          config.Enabled,
		LastAnalysisTime: time.Now(), // TODO: 从实际分析记录获取
		PatternStats: &dto.AttackPatternStatsResponse{
			TotalPatterns:    totalPatterns,
			ActivePatterns:   activePatterns,
			ByType:           make(map[string]int),
			BySeverity:       make(map[string]int),
			RecentDetections: 0, // TODO: 查询最近24小时
		},
		RuleStats: &dto.GeneratedRuleStatsResponse{
			TotalRules:    totalRules,
			PendingRules:  pendingRules,
			ApprovedRules: approvedRules,
			DeployedRules: deployedRules,
			RejectedRules: rejectedRules,
			ByType:        make(map[string]int),
		},
	}, nil
}

// TriggerAnalysis 手动触发AI分析
func (s *AIAnalyzerServiceImpl) TriggerAnalysis(ctx context.Context) error {
	s.logger.Info().Msg("手动触发AI分析")
	
	// 创建AI引擎实例
	engine := NewAIEngine(s.patternRepo.GetDB())
	
	// 运行攻击模式检测
	patterns, err := engine.RunAttackPatternDetection(ctx)
	if err != nil {
		s.logger.Error().Err(err).Msg("攻击模式检测失败")
		return err
	}
	
	s.logger.Info().Int("pattern_count", len(patterns)).Msg("检测到攻击模式")
	
	// 为高危模式生成规则
	var highRiskPatterns []*model.AttackPattern
	for i := range patterns {
		if patterns[i].Severity == "high" || patterns[i].Severity == "critical" {
			highRiskPatterns = append(highRiskPatterns, patterns[i])
		}
	}
	
	if len(highRiskPatterns) > 0 {
		s.logger.Info().Int("high_risk_count", len(highRiskPatterns)).Msg("为高危模式生成规则")
		_, err = engine.GenerateRulesForPatterns(ctx, highRiskPatterns, 0.7)
		if err != nil {
			s.logger.Error().Err(err).Msg("规则生成失败")
			return err
		}
	}
	
	return nil
}
