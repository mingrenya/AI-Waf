package service

import (
	"context"
	"fmt"
	"time"

	"github.com/mingrenya/AI-Waf/pkg/model"
	"github.com/mingrenya/AI-Waf/server/config"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// RuleTemplateService 规则模板服务接口
type RuleTemplateService interface {
	// 获取所有模板
	ListTemplates(ctx context.Context, category string, severity string) ([]model.RuleTemplate, error)
	// 获取模板详情
	GetTemplate(ctx context.Context, id string) (*model.RuleTemplate, error)
	// 创建模板
	CreateTemplate(ctx context.Context, template *model.RuleTemplate) error
	// 从模板创建规则
	CreateRuleFromTemplate(ctx context.Context, templateID string, customName string) (*model.MicroRule, error)
	// 初始化OWASP Top 10模板
	InitializeOWASPTemplates(ctx context.Context) error
}

type ruleTemplateServiceImpl struct {
	templateCollection *mongo.Collection
	ruleCollection     *mongo.Collection
}

// NewRuleTemplateService 创建规则模板服务
func NewRuleTemplateService(db *mongo.Database) RuleTemplateService {
	return &ruleTemplateServiceImpl{
		templateCollection: db.Collection("rule_template"),
		ruleCollection:     db.Collection("micro_rule"),
	}
}

// ListTemplates 获取模板列表
func (s *ruleTemplateServiceImpl) ListTemplates(ctx context.Context, category string, severity string) ([]model.RuleTemplate, error) {
	filter := bson.M{}
	if category != "" {
		filter["category"] = category
	}
	if severity != "" {
		filter["severity"] = severity
	}

	opts := options.Find().SetSort(bson.D{{Key: "priority", Value: -1}})
	cursor, err := s.templateCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("查询模板失败: %w", err)
	}
	defer cursor.Close(ctx)

	var templates []model.RuleTemplate
	if err = cursor.All(ctx, &templates); err != nil {
		return nil, fmt.Errorf("解析模板失败: %w", err)
	}

	return templates, nil
}

// GetTemplate 获取模板详情
func (s *ruleTemplateServiceImpl) GetTemplate(ctx context.Context, id string) (*model.RuleTemplate, error) {
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("无效的模板ID: %w", err)
	}

	var template model.RuleTemplate
	err = s.templateCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&template)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("模板不存在")
		}
		return nil, fmt.Errorf("查询模板失败: %w", err)
	}

	return &template, nil
}

// CreateTemplate 创建模板
func (s *ruleTemplateServiceImpl) CreateTemplate(ctx context.Context, template *model.RuleTemplate) error {
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()

	_, err := s.templateCollection.InsertOne(ctx, template)
	if err != nil {
		return fmt.Errorf("创建模板失败: %w", err)
	}

	return nil
}

// CreateRuleFromTemplate 从模板创建规则
func (s *ruleTemplateServiceImpl) CreateRuleFromTemplate(ctx context.Context, templateID string, customName string) (*model.MicroRule, error) {
	template, err := s.GetTemplate(ctx, templateID)
	if err != nil {
		return nil, err
	}

	ruleName := customName
	if ruleName == "" {
		ruleName = template.Name
	}

	rule := &model.MicroRule{
		Name:      ruleName,
		Type:      template.RuleType,
		Status:    model.RuleEnabled,
		Priority:  template.Priority,
		Condition: template.Condition,
	}

	_, err = s.ruleCollection.InsertOne(ctx, rule)
	if err != nil {
		return nil, fmt.Errorf("创建规则失败: %w", err)
	}

	return rule, nil
}

// InitializeOWASPTemplates 初始化OWASP Top 10模板
func (s *ruleTemplateServiceImpl) InitializeOWASPTemplates(ctx context.Context) error {
	// 检查是否已经初始化
	count, err := s.templateCollection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("检查模板失败: %w", err)
	}

	if count > 0 {
		// 已经有模板，记录日志并跳过初始化
		config.Logger.Info().Int64("count", count).Msg("OWASP模板已存在，跳过初始化")
		return nil
	}

	config.Logger.Info().Msg("开始初始化OWASP Top 10模板...")
	templates := s.getOWASPTemplates()

	var docs []interface{}
	for i := range templates {
		templates[i].CreatedAt = time.Now()
		templates[i].UpdatedAt = time.Now()
		docs = append(docs, templates[i])
	}

	result, err := s.templateCollection.InsertMany(ctx, docs)
	if err != nil {
		return fmt.Errorf("初始化OWASP模板失败: %w", err)
	}

	config.Logger.Info().Int("count", len(result.InsertedIDs)).Msg("OWASP模板初始化成功")
	return nil
}

// getOWASPTemplates 获取OWASP Top 10预定义模板
func (s *ruleTemplateServiceImpl) getOWASPTemplates() []model.RuleTemplate {
	return []model.RuleTemplate{
		// A01:2021 - 失效的访问控制
		{
			Name:        "防止未授权访问管理路径",
			Category:    model.CategoryBrokenAccessControl,
			Description: "阻止对管理路径的未授权访问，如 /admin、/管理等",
			Severity:    model.SeverityCritical,
			RuleType:    model.BlacklistRule,
			Priority:    900,
			Tags:        []string{"owasp", "a01", "access-control", "admin"},
			Condition: mustMarshal(map[string]interface{}{
				"type":     "composite",
				"operator": "AND",
				"conditions": []interface{}{
					map[string]interface{}{
						"type":        "simple",
						"target":      "path",
						"match_type":  "regex",
						"match_value": "^/(admin|管理|后台|manage|backend|console)",
					},
				},
			}),
		},
		// A03:2021 - 注入
		{
			Name:        "SQL注入防护",
			Category:    model.CategoryInjection,
			Description: "检测并阻止常见的SQL注入攻击模式",
			Severity:    model.SeverityCritical,
			RuleType:    model.BlacklistRule,
			Priority:    950,
			Tags:        []string{"owasp", "a03", "injection", "sql"},
			Condition: mustMarshal(map[string]interface{}{
				"type":     "composite",
				"operator": "OR",
				"conditions": []interface{}{
					map[string]interface{}{
						"type":        "simple",
						"target":      "url",
						"match_type":  "regex",
						"match_value": "(?i)(union.*select|select.*from|insert.*into|delete.*from|update.*set|drop.*table)",
					},
					map[string]interface{}{
						"type":        "simple",
						"target":      "url",
						"match_type":  "regex",
						"match_value": "(?i)('|\")(\\s)*(or|and)(\\s)*('|\")?(\\s)*=",
					},
				},
			}),
		},
		{
			Name:        "XSS跨站脚本防护",
			Category:    model.CategoryInjection,
			Description: "检测并阻止跨站脚本(XSS)攻击",
			Severity:    model.SeverityHigh,
			RuleType:    model.BlacklistRule,
			Priority:    850,
			Tags:        []string{"owasp", "a03", "injection", "xss"},
			Condition: mustMarshal(map[string]interface{}{
				"type":     "composite",
				"operator": "OR",
				"conditions": []interface{}{
					map[string]interface{}{
						"type":        "simple",
						"target":      "url",
						"match_type":  "regex",
						"match_value": "(?i)<script[^>]*>.*?</script>",
					},
					map[string]interface{}{
						"type":        "simple",
						"target":      "url",
						"match_type":  "regex",
						"match_value": "(?i)(javascript:|onerror=|onload=|onclick=)",
					},
				},
			}),
		},
		{
			Name:        "命令注入防护",
			Category:    model.CategoryInjection,
			Description: "检测并阻止操作系统命令注入攻击",
			Severity:    model.SeverityCritical,
			RuleType:    model.BlacklistRule,
			Priority:    900,
			Tags:        []string{"owasp", "a03", "injection", "command"},
			Condition: mustMarshal(map[string]interface{}{
				"type":     "composite",
				"operator": "OR",
				"conditions": []interface{}{
					map[string]interface{}{
						"type":        "simple",
						"target":      "url",
						"match_type":  "regex",
						"match_value": "(?i)(;|\\||&|`|\\$\\(|\\$\\{)",
					},
					map[string]interface{}{
						"type":        "simple",
						"target":      "url",
						"match_type":  "regex",
						"match_value": "(?i)(cat|ls|wget|curl|chmod|chown|rm|mv|cp)\\s",
					},
				},
			}),
		},
		// A05:2021 - 安全配置错误
		{
			Name:        "敏感文件访问防护",
			Category:    model.CategorySecurityMisconfiguration,
			Description: "阻止访问敏感配置文件和备份文件",
			Severity:    model.SeverityHigh,
			RuleType:    model.BlacklistRule,
			Priority:    800,
			Tags:        []string{"owasp", "a05", "config", "sensitive"},
			Condition: mustMarshal(map[string]interface{}{
				"type":     "composite",
				"operator": "OR",
				"conditions": []interface{}{
					map[string]interface{}{
						"type":        "simple",
						"target":      "path",
						"match_type":  "regex",
						"match_value": "\\.(env|config|ini|sql|bak|backup|old|swp|~)$",
					},
					map[string]interface{}{
						"type":        "simple",
						"target":      "path",
						"match_type":  "regex",
						"match_value": "/(\\.|_)(git|svn|hg|bzr)",
					},
				},
			}),
		},
		// A07:2021 - 识别和身份验证失败
		{
			Name:        "暴力破解防护",
			Category:    model.CategoryAuthenticationFailures,
			Description: "防止对登录和认证端点的暴力破解攻击",
			Severity:    model.SeverityHigh,
			RuleType:    model.BlacklistRule,
			Priority:    750,
			Tags:        []string{"owasp", "a07", "auth", "brute-force"},
			Condition: mustMarshal(map[string]interface{}{
				"type":        "simple",
				"target":      "path",
				"match_type":  "regex",
				"match_value": "/(login|signin|auth|authenticate)",
			}),
		},
		// A10:2021 - 服务器端请求伪造
		{
			Name:        "SSRF防护",
			Category:    model.CategorySSRF,
			Description: "检测并阻止服务器端请求伪造(SSRF)攻击",
			Severity:    model.SeverityHigh,
			RuleType:    model.BlacklistRule,
			Priority:    850,
			Tags:        []string{"owasp", "a10", "ssrf"},
			Condition: mustMarshal(map[string]interface{}{
				"type":     "composite",
				"operator": "OR",
				"conditions": []interface{}{
					map[string]interface{}{
						"type":        "simple",
						"target":      "url",
						"match_type":  "regex",
						"match_value": "(?i)(localhost|127\\.0\\.0\\.1|0\\.0\\.0\\.0|\\[::1\\])",
					},
					map[string]interface{}{
						"type":        "simple",
						"target":      "url",
						"match_type":  "regex",
						"match_value": "(?i)(file://|dict://|gopher://|ftp://)",
					},
				},
			}),
		},
		// 路径遍历防护
		{
			Name:        "路径遍历防护",
			Category:    model.CategoryBrokenAccessControl,
			Description: "防止目录遍历攻击",
			Severity:    model.SeverityHigh,
			RuleType:    model.BlacklistRule,
			Priority:    850,
			Tags:        []string{"owasp", "a01", "path-traversal"},
			Condition: mustMarshal(map[string]interface{}{
				"type":     "composite",
				"operator": "OR",
				"conditions": []interface{}{
					map[string]interface{}{
						"type":        "simple",
						"target":      "path",
						"match_type":  "regex",
						"match_value": "\\.\\./",
					},
					map[string]interface{}{
						"type":        "simple",
						"target":      "url",
						"match_type":  "contains",
						"match_value": "%2e%2e/",
					},
				},
			}),
		},
		// 恶意User-Agent检测
		{
			Name:        "恶意扫描器和爬虫检测",
			Category:    model.CategoryVulnerableComponents,
			Description: "检测并阻止已知的恶意扫描器和爬虫",
			Severity:    model.SeverityMedium,
			RuleType:    model.BlacklistRule,
			Priority:    600,
			Tags:        []string{"owasp", "a06", "scanner", "bot"},
			Condition: mustMarshal(map[string]interface{}{
				"type":        "simple",
				"target":      "url",
				"match_type":  "regex",
				"match_value": "(?i)(nmap|sqlmap|nikto|nessus|masscan|metasploit|burp)",
			}),
		},
	}
}

// mustMarshal 辅助函数，将数据序列化为BSON
// 如果序列化失败，记录错误并返回空BSON
func mustMarshal(v interface{}) bson.Raw {
	data, err := bson.Marshal(v)
	if err != nil {
		config.Logger.Error().Err(err).Msg("Failed to marshal rule template condition")
		// 返回空BSON而不是panic
		emptyData, _ := bson.Marshal(bson.M{})
		return emptyData
	}
	return data
}
