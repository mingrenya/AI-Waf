package service

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"text/template"
	"time"

	"github.com/mingrenya/AI-Waf/server/config"
	"github.com/mingrenya/AI-Waf/server/dto"
	"github.com/mingrenya/AI-Waf/server/model"
	"github.com/mingrenya/AI-Waf/server/repository"
	"github.com/mingrenya/AI-Waf/server/service/alert"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// AlertService 告警服务接口
type AlertService interface {
	// Channel 管理
	CreateChannel(ctx context.Context, req *dto.CreateAlertChannelRequest, userID string) (*dto.AlertChannelResponse, error)
	GetChannels(ctx context.Context) ([]*dto.AlertChannelResponse, error)
	GetChannelByID(ctx context.Context, id string) (*dto.AlertChannelResponse, error)
	UpdateChannel(ctx context.Context, id string, req *dto.UpdateAlertChannelRequest) error
	DeleteChannel(ctx context.Context, id string) error
	TestChannel(ctx context.Context, id string, req *dto.TestAlertChannelRequest) error

	// Rule 管理
	CreateRule(ctx context.Context, req *dto.CreateAlertRuleRequest, userID string) (*dto.AlertRuleResponse, error)
	GetRules(ctx context.Context) ([]*dto.AlertRuleResponse, error)
	GetRuleByID(ctx context.Context, id string) (*dto.AlertRuleResponse, error)
	UpdateRule(ctx context.Context, id string, req *dto.UpdateAlertRuleRequest) error
	DeleteRule(ctx context.Context, id string) error

	// History 管理
	GetAlertHistory(ctx context.Context, req *dto.GetAlertHistoryRequest) ([]*dto.AlertHistoryResponse, int64, error)
	AcknowledgeAlert(ctx context.Context, id string, userID string, comment string) error
	GetStatistics(ctx context.Context, startTime, endTime time.Time) (*dto.AlertStatisticsResponse, error)

	// 核心功能
	CheckAndTriggerAlerts(ctx context.Context) error
	SendAlert(ctx context.Context, rule *model.AlertRule, data map[string]interface{}) error
}

type alertServiceImpl struct {
	channelRepo  repository.AlertChannelRepository
	ruleRepo     repository.AlertRuleRepository
	historyRepo  repository.AlertHistoryRepository
	statsService StatsService
	senders      map[string]alert.Sender
	cooldownMap  map[string]time.Time
	cooldownMu   sync.RWMutex
	logger       zerolog.Logger
}

// NewAlertService 创建告警服务
func NewAlertService(
	channelRepo repository.AlertChannelRepository,
	ruleRepo repository.AlertRuleRepository,
	historyRepo repository.AlertHistoryRepository,
	statsService StatsService,
) AlertService {
	// 初始化所有发送器
	senders := map[string]alert.Sender{
		model.AlertChannelTypeWebhook:  alert.NewWebhookSender(),
		model.AlertChannelTypeSlack:    alert.NewSlackSender(),
		model.AlertChannelTypeDiscord:  alert.NewDiscordSender(),
		model.AlertChannelTypeDingTalk: alert.NewDingTalkSender(),
		model.AlertChannelTypeWeCom:    alert.NewWeComSender(),
	}

	return &alertServiceImpl{
		channelRepo:  channelRepo,
		ruleRepo:     ruleRepo,
		historyRepo:  historyRepo,
		statsService: statsService,
		senders:      senders,
		cooldownMap:  make(map[string]time.Time),
		logger:       config.GetServiceLogger("alert"),
	}
}

// CreateChannel 创建告警渠道
func (s *alertServiceImpl) CreateChannel(ctx context.Context, req *dto.CreateAlertChannelRequest, userID string) (*dto.AlertChannelResponse, error) {
	// 验证配置
	sender, ok := s.senders[req.Type]
	if !ok {
		return nil, fmt.Errorf("unsupported channel type: %s", req.Type)
	}

	if err := sender.Validate(req.Config); err != nil {
		return nil, fmt.Errorf("invalid channel config: %w", err)
	}

	channel := &model.AlertChannel{
		Name:    req.Name,
		Type:    req.Type,
		Config:  req.Config,
		Enabled: req.Enabled,
	}

	if err := s.channelRepo.Create(ctx, channel); err != nil {
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}

	return s.toChannelResponse(channel), nil
}

// GetChannels 获取所有渠道
func (s *alertServiceImpl) GetChannels(ctx context.Context) ([]*dto.AlertChannelResponse, error) {
	channels, err := s.channelRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.AlertChannelResponse, 0, len(channels))
	for _, ch := range channels {
		responses = append(responses, s.toChannelResponse(ch))
	}

	return responses, nil
}

// GetChannelByID 获取单个渠道
func (s *alertServiceImpl) GetChannelByID(ctx context.Context, id string) (*dto.AlertChannelResponse, error) {
	channel, err := s.channelRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.toChannelResponse(channel), nil
}

// UpdateChannel 更新渠道
func (s *alertServiceImpl) UpdateChannel(ctx context.Context, id string, req *dto.UpdateAlertChannelRequest) error {
	updates := make(map[string]interface{})

	if req.Name != "" {
		updates["name"] = req.Name
	}

	if req.Config != nil {
		// 验证配置
		channel, err := s.channelRepo.GetByID(ctx, id)
		if err != nil {
			return err
		}

		sender, ok := s.senders[channel.Type]
		if !ok {
			return fmt.Errorf("unsupported channel type: %s", channel.Type)
		}

		if err := sender.Validate(req.Config); err != nil {
			return fmt.Errorf("invalid channel config: %w", err)
		}

		updates["config"] = req.Config
	}

	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}

	if len(updates) == 0 {
		return nil
	}

	return s.channelRepo.Update(ctx, id, updates)
}

// DeleteChannel 删除渠道
func (s *alertServiceImpl) DeleteChannel(ctx context.Context, id string) error {
	return s.channelRepo.Delete(ctx, id)
}

// TestChannel 测试渠道
func (s *alertServiceImpl) TestChannel(ctx context.Context, id string, req *dto.TestAlertChannelRequest) error {
	channel, err := s.channelRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	sender, ok := s.senders[channel.Type]
	if !ok {
		return fmt.Errorf("unsupported channel type: %s", channel.Type)
	}

	return sender.Send(ctx, channel, req.Message)
}

// CreateRule 创建告警规则
func (s *alertServiceImpl) CreateRule(ctx context.Context, req *dto.CreateAlertRuleRequest, userID string) (*dto.AlertRuleResponse, error) {
	rule := &model.AlertRule{
		Name:        req.Name,
		Description: req.Description,
		Conditions:  req.Conditions,
		Logic:       req.Logic,
		Channels:    req.Channels,
		Template:    req.Template,
		Cooldown:    req.Cooldown,
		Severity:    req.Severity,
		Enabled:     req.Enabled,
		CreatedBy:   userID,
	}

	if err := s.ruleRepo.Create(ctx, rule); err != nil {
		return nil, fmt.Errorf("failed to create rule: %w", err)
	}

	return s.toRuleResponse(rule), nil
}

// GetRules 获取所有规则
func (s *alertServiceImpl) GetRules(ctx context.Context) ([]*dto.AlertRuleResponse, error) {
	rules, err := s.ruleRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.AlertRuleResponse, 0, len(rules))
	for _, rule := range rules {
		responses = append(responses, s.toRuleResponse(rule))
	}

	return responses, nil
}

// GetRuleByID 获取单个规则
func (s *alertServiceImpl) GetRuleByID(ctx context.Context, id string) (*dto.AlertRuleResponse, error) {
	rule, err := s.ruleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.toRuleResponse(rule), nil
}

// UpdateRule 更新规则
func (s *alertServiceImpl) UpdateRule(ctx context.Context, id string, req *dto.UpdateAlertRuleRequest) error {
	updates := make(map[string]interface{})

	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Conditions != nil {
		updates["conditions"] = req.Conditions
	}
	if req.Logic != "" {
		updates["logic"] = req.Logic
	}
	if req.Channels != nil {
		updates["channels"] = req.Channels
	}
	if req.Template != "" {
		updates["template"] = req.Template
	}
	if req.Cooldown != nil {
		updates["cooldown"] = *req.Cooldown
	}
	if req.Severity != "" {
		updates["severity"] = req.Severity
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}

	if len(updates) == 0 {
		return nil
	}

	return s.ruleRepo.Update(ctx, id, updates)
}

// DeleteRule 删除规则
func (s *alertServiceImpl) DeleteRule(ctx context.Context, id string) error {
	return s.ruleRepo.Delete(ctx, id)
}

// GetAlertHistory 获取告警历史
func (s *alertServiceImpl) GetAlertHistory(ctx context.Context, req *dto.GetAlertHistoryRequest) ([]*dto.AlertHistoryResponse, int64, error) {
	filter := bson.M{}

	if req.RuleID != "" {
		filter["rule_id"] = req.RuleID
	}

	if req.Severity != "" {
		filter["severity"] = req.Severity
	}

	if req.Status != "" {
		filter["status"] = req.Status
	}

	if !req.StartTime.IsZero() || !req.EndTime.IsZero() {
		filter["triggered_at"] = bson.M{}
		if !req.StartTime.IsZero() {
			filter["triggered_at"].(bson.M)["$gte"] = req.StartTime
		}
		if !req.EndTime.IsZero() {
			filter["triggered_at"].(bson.M)["$lte"] = req.EndTime
		}
	}

	page := req.Page
	if page < 1 {
		page = 1
	}

	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 20
	}

	histories, total, err := s.historyRepo.Query(ctx, filter, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*dto.AlertHistoryResponse, 0, len(histories))
	for _, history := range histories {
		responses = append(responses, s.toHistoryResponse(history))
	}

	return responses, total, nil
}

// AcknowledgeAlert 确认告警
func (s *alertServiceImpl) AcknowledgeAlert(ctx context.Context, id string, userID string, comment string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":           model.AlertStatusAcknowledged,
		"acknowledged_at":  now,
		"acknowledged_by":  userID,
	}

	return s.historyRepo.Update(ctx, id, updates)
}

// GetStatistics 获取统计信息
func (s *alertServiceImpl) GetStatistics(ctx context.Context, startTime, endTime time.Time) (*dto.AlertStatisticsResponse, error) {
	stats, err := s.historyRepo.GetStatistics(ctx, startTime, endTime)
	if err != nil {
		return nil, err
	}

	response := &dto.AlertStatisticsResponse{
		TotalAlerts:      stats["total"].(int64),
		AlertsBySeverity: stats["alertsBySeverity"].(map[string]int64),
		AlertsByStatus:   stats["alertsByStatus"].(map[string]int64),
	}

	return response, nil
}

// CheckAndTriggerAlerts 检查并触发告警
func (s *alertServiceImpl) CheckAndTriggerAlerts(ctx context.Context) error {
	rules, err := s.ruleRepo.GetEnabled(ctx)
	if err != nil {
		return fmt.Errorf("failed to get enabled rules: %w", err)
	}

	for _, rule := range rules {
		// 检查冷却时间
		if !s.canTrigger(rule.ID.Hex(), rule.Cooldown) {
			continue
		}

		// 评估规则条件
		triggered, data, err := s.evaluateRule(ctx, rule)
		if err != nil {
			s.logger.Error().Err(err).Str("rule_id", rule.ID.Hex()).Msg("Failed to evaluate rule")
			continue
		}

		if triggered {
			if err := s.SendAlert(ctx, rule, data); err != nil {
				s.logger.Error().Err(err).Str("rule_id", rule.ID.Hex()).Msg("Failed to send alert")
			}
		}
	}

	return nil
}

// SendAlert 发送告警
func (s *alertServiceImpl) SendAlert(ctx context.Context, rule *model.AlertRule, data map[string]interface{}) error {
	// 渲染消息
	message, err := s.renderMessage(rule.Template, data)
	if err != nil {
		return fmt.Errorf("failed to render message: %w", err)
	}

	// 创建告警历史记录
	history := &model.AlertHistory{
		RuleID:      rule.ID.Hex(),
		RuleName:    rule.Name,
		Severity:    rule.Severity,
		Message:     message,
		Details:     data,
		Channels:    rule.Channels,
		Status:      model.AlertStatusPending,
	}

	if err := s.historyRepo.Create(ctx, history); err != nil {
		return fmt.Errorf("failed to create history: %w", err)
	}

	// 发送到各个渠道
	for _, channelID := range rule.Channels {
		channel, err := s.channelRepo.GetByID(ctx, channelID)
		if err != nil {
			s.logger.Error().Err(err).Str("channel_id", channelID).Msg("Failed to get channel")
			continue
		}

		if !channel.Enabled {
			continue
		}

		sender, ok := s.senders[channel.Type]
		if !ok {
			s.logger.Error().Str("type", channel.Type).Msg("Unsupported channel type")
			continue
		}

		if err := sender.Send(ctx, channel, message); err != nil {
			s.logger.Error().Err(err).Str("channel_id", channelID).Msg("Failed to send alert")
			// 更新失败状态
			s.historyRepo.Update(ctx, history.ID.Hex(), map[string]interface{}{
				"status":        model.AlertStatusFailed,
				"error_message": err.Error(),
			})
		}
	}

	// 更新成功状态
	now := time.Now()
	s.historyRepo.Update(ctx, history.ID.Hex(), map[string]interface{}{
		"status":  model.AlertStatusSent,
		"sent_at": now,
	})

	// 更新冷却时间
	s.updateCooldown(rule.ID.Hex())

	return nil
}

// 辅助方法

func (s *alertServiceImpl) toChannelResponse(ch *model.AlertChannel) *dto.AlertChannelResponse {
	return &dto.AlertChannelResponse{
		ID:        ch.ID.Hex(),
		Name:      ch.Name,
		Type:      ch.Type,
		Config:    ch.Config,
		Enabled:   ch.Enabled,
		CreatedAt: ch.CreatedAt,
		UpdatedAt: ch.UpdatedAt,
	}
}

func (s *alertServiceImpl) toRuleResponse(rule *model.AlertRule) *dto.AlertRuleResponse {
	return &dto.AlertRuleResponse{
		ID:          rule.ID.Hex(),
		Name:        rule.Name,
		Description: rule.Description,
		Conditions:  rule.Conditions,
		Logic:       rule.Logic,
		Channels:    rule.Channels,
		Template:    rule.Template,
		Cooldown:    rule.Cooldown,
		Severity:    rule.Severity,
		Enabled:     rule.Enabled,
		CreatedAt:   rule.CreatedAt,
		UpdatedAt:   rule.UpdatedAt,
		CreatedBy:   rule.CreatedBy,
	}
}

func (s *alertServiceImpl) toHistoryResponse(history *model.AlertHistory) *dto.AlertHistoryResponse {
	return &dto.AlertHistoryResponse{
		ID:              history.ID.Hex(),
		RuleID:          history.RuleID,
		RuleName:        history.RuleName,
		Severity:        history.Severity,
		Message:         history.Message,
		Details:         history.Details,
		Channels:        history.Channels,
		Status:          history.Status,
		ErrorMessage:    history.ErrorMessage,
		TriggeredAt:     history.TriggeredAt,
		SentAt:          history.SentAt,
		AcknowledgedAt:  history.AcknowledgedAt,
		AcknowledgedBy:  history.AcknowledgedBy,
	}
}

func (s *alertServiceImpl) canTrigger(ruleID string, cooldownMinutes int) bool {
	s.cooldownMu.RLock()
	lastTrigger, exists := s.cooldownMap[ruleID]
	s.cooldownMu.RUnlock()

	if !exists {
		return true
	}

	return time.Since(lastTrigger) > time.Duration(cooldownMinutes)*time.Minute
}

func (s *alertServiceImpl) updateCooldown(ruleID string) {
	s.cooldownMu.Lock()
	s.cooldownMap[ruleID] = time.Now()
	s.cooldownMu.Unlock()
}

func (s *alertServiceImpl) evaluateRule(ctx context.Context, rule *model.AlertRule) (bool, map[string]interface{}, error) {
	// 获取当前统计数据
	stats, err := s.statsService.GetOverviewStats(ctx, "1h")
	if err != nil {
		return false, nil, err
	}

	data := map[string]interface{}{
		"qps":          stats.MaxQPS,
		"block_rate":   float64(stats.BlockCount) / float64(stats.TotalRequests) * 100,
		"error_4xx_rate": stats.Error4xxRate,
		"error_5xx_rate": stats.Error5xxRate,
		"attack_count": stats.BlockCount,
		"traffic":      stats.InboundTraffic + stats.OutboundTraffic,
	}

	// 评估条件
	var result bool
	if rule.Logic == "AND" {
		result = true
		for _, cond := range rule.Conditions {
			if !s.evaluateCondition(cond, data) {
				result = false
				break
			}
		}
	} else { // OR
		result = false
		for _, cond := range rule.Conditions {
			if s.evaluateCondition(cond, data) {
				result = true
				break
			}
		}
	}

	return result, data, nil
}

func (s *alertServiceImpl) evaluateCondition(cond model.AlertCondition, data map[string]interface{}) bool {
	value, ok := data[cond.Metric]
	if !ok {
		return false
	}

	var numValue float64
	switch v := value.(type) {
	case int64:
		numValue = float64(v)
	case float64:
		numValue = v
	case int:
		numValue = float64(v)
	default:
		return false
	}

	var threshold float64
	switch t := cond.Threshold.(type) {
	case float64:
		threshold = t
	case int64:
		threshold = float64(t)
	case int:
		threshold = float64(t)
	default:
		return false
	}

	switch cond.Operator {
	case model.AlertOperatorGreaterThan:
		return numValue > threshold
	case model.AlertOperatorLessThan:
		return numValue < threshold
	case model.AlertOperatorGreaterThanEqual:
		return numValue >= threshold
	case model.AlertOperatorLessThanEqual:
		return numValue <= threshold
	case model.AlertOperatorEqual:
		return numValue == threshold
	case model.AlertOperatorNotEqual:
		return numValue != threshold
	default:
		return false
	}
}

func (s *alertServiceImpl) renderMessage(tmpl string, data map[string]interface{}) (string, error) {
	t, err := template.New("alert").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
