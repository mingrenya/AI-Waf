package situation

import (
	"context"
	"fmt"
	"time"

	"github.com/mingrenya/AI-Waf/pkg/model"
	"github.com/mingrenya/AI-Waf/server/config"
	"github.com/mingrenya/AI-Waf/server/repository"
	"github.com/mingrenya/AI-Waf/server/service"
	"github.com/rs/zerolog"
)

// DefaultRules 内置默认检测规则
var DefaultRules = []model.SituationRule{
	{
		Name:           "高频端口扫描检测",
		Stage:          string(model.StageScanning),
		LogQL:          `sum by (source_ip)(count_over_time({container_name="mrya-waf",attack_type="scanner"}[5m])) > 100`,
		Interval:       30,
		Threshold:      100,
		Severity:       "medium",
		MITRETactic:    "TA0043",
		MITRETechnique: "T1595",
		Enabled:        true,
	},
	{
		Name:           "SQL注入高频利用",
		Stage:          string(model.StageExploitation),
		LogQL:          `sum by (source_ip)(count_over_time({container_name="mrya-waf",attack_type="sql_injection",severity=~"critical|high"}[5m])) > 10`,
		Interval:       30,
		Threshold:      10,
		Severity:       "critical",
		MITRETactic:    "TA0001",
		MITRETechnique: "T1190",
		Enabled:        true,
	},
	{
		Name:           "多攻击面探测",
		Stage:          string(model.StageScanning),
		LogQL:          `count by (source_ip)(count_over_time({container_name="mrya-waf",attack_type!=""}[10m])) > 50`,
		Interval:       60,
		Threshold:      3,
		Severity:       "high",
		MITRETactic:    "TA0043",
		MITRETechnique: "T1046",
		Enabled:        true,
	},
	{
		Name:           "漏洞扫描器指纹识别",
		Stage:          string(model.StageScanning),
		LogQL:          `{container_name="mrya-waf"} |~ "(?i)(nuclei|nikto|nessus|openvas|acunetix|burpsuite)"`,
		Interval:       30,
		Threshold:      1,
		Severity:       "high",
		MITRETactic:    "TA0043",
		MITRETechnique: "T1595",
		Enabled:        true,
	},
	{
		Name:           "RCE攻击检测",
		Stage:          string(model.StageExploitation),
		LogQL:          `sum by (source_ip)(count_over_time({container_name="mrya-waf",attack_type=~"rce|command_injection"}[5m])) > 5`,
		Interval:       30,
		Threshold:      5,
		Severity:       "critical",
		MITRETactic:    "TA0002",
		MITRETechnique: "T1059",
		Enabled:        true,
	},
}

// RuleEngine LogQL 规则评估引擎
type RuleEngine struct {
	loki   service.LokiLogService
	repo   repository.SituationRepository
	logger zerolog.Logger
}

// NewRuleEngine 创建规则引擎
func NewRuleEngine(loki service.LokiLogService, repo repository.SituationRepository) *RuleEngine {
	return &RuleEngine{
		loki:   loki,
		repo:   repo,
		logger: config.GetServiceLogger("situation-rule"),
	}
}

// InitializeDefaults 写入默认规则（已存在的跳过，按 name 去重）
func (e *RuleEngine) InitializeDefaults(ctx context.Context) error {
	for _, rule := range DefaultRules {
		existing, err := e.repo.FindRuleByName(ctx, rule.Name)
		if err != nil {
			return fmt.Errorf("初始化规则 '%s' 失败: %w", rule.Name, err)
		}
		if existing != nil {
			continue
		}
		if err := e.repo.CreateRule(ctx, &rule); err != nil {
			return fmt.Errorf("创建规则 '%s' 失败: %w", rule.Name, err)
		}
	}
	e.logger.Info().Int("count", len(DefaultRules)).Msg("Default rules initialised")
	return nil
}

// Evaluate 执行单条规则评估
func (e *RuleEngine) Evaluate(ctx context.Context, rule model.SituationRule) (*RuleResult, error) {
	resp, err := e.loki.QueryLogs(ctx, service.LokiQueryRequest{
		Query: rule.LogQL,
		Limit: 100,
		Start: fmt.Sprintf("%dm", rule.Interval/60+1),
	})
	if err != nil {
		e.logger.Warn().Err(err).Str("rule_id", rule.ID).Msg("LogQL evaluation failed")
		return nil, err
	}

	result := &RuleResult{
		Rule:      rule,
		HitCount:  len(resp.Data.Result),
		HitIPs:    extractSourceIPs(resp),
		Timestamp: time.Now(),
	}
	result.Triggered = result.HitCount > rule.Threshold
	return result, nil
}

// EvaluateAll 评估所有启用的规则
func (e *RuleEngine) EvaluateAll(ctx context.Context) ([]RuleResult, error) {
	rules, err := e.repo.ListEnabledRules(ctx)
	if err != nil {
		return nil, err
	}

	results := make([]RuleResult, 0)
	for _, rule := range rules {
		result, err := e.Evaluate(ctx, rule)
		if err != nil {
			e.logger.Warn().Err(err).Str("rule_id", rule.ID).Msg("Skipping failed rule")
			continue
		}
		if result.Triggered {
			results = append(results, *result)
		}
	}
	return results, nil
}

// RuleResult 单条规则评估结果
type RuleResult struct {
	Rule      model.SituationRule `json:"rule"`
	HitCount  int                 `json:"hit_count"`
	HitIPs    []string            `json:"hit_ips"`
	Triggered bool                `json:"triggered"`
	Timestamp time.Time           `json:"timestamp"`
}

func extractSourceIPs(resp *service.LokiQueryResponse) []string {
	seen := make(map[string]bool)
	var ips []string
	for _, stream := range resp.Data.Result {
		if ip, ok := stream.Stream["source_ip"]; ok && ip != "" && !seen[ip] {
			seen[ip] = true
			ips = append(ips, ip)
		}
	}
	return ips
}
