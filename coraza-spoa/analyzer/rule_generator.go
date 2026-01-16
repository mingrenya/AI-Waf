package analyzer

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mingrenya/AI-Waf/pkg/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// Logger 日志接口
type Logger interface {
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Debugf(format string, args ...interface{})
}

// RuleGenerator 规则生成器 - 基于攻击模式生成ModSecurity和MicroRule规则
type RuleGenerator struct {
	db     *mongo.Database
	logger Logger
	
	// 规则ID计数器
	nextRuleID int
}

// NewRuleGenerator 创建规则生成器
func NewRuleGenerator(db *mongo.Database, logger Logger) *RuleGenerator {
	return &RuleGenerator{
		db:         db,
		logger:     logger,
		nextRuleID: 90000, // AI生成的规则ID从90000开始
	}
}

// GenerateRules 生成防护规则
func (rg *RuleGenerator) GenerateRules(patterns []*model.AttackPattern, confidenceThreshold float64) ([]*model.GeneratedRule, error) {
	rg.logger.Infof("开始生成规则, 模式数量: %d, 置信度阈值: %.2f", len(patterns), confidenceThreshold)
	
	rules := make([]*model.GeneratedRule, 0)
	
	for _, pattern := range patterns {
		// 只为高置信度模式生成规则
		if pattern.Confidence < confidenceThreshold {
			rg.logger.Debugf("跳过低置信度模式: %s (%.2f)", pattern.Name, pattern.Confidence)
			continue
		}
		
		// 生成ModSecurity规则
		if modSecRule := rg.generateModSecurityRule(pattern); modSecRule != nil {
			rules = append(rules, modSecRule)
		}
		
		// 生成MicroRule
		if microRule := rg.generateMicroRule(pattern); microRule != nil {
			rules = append(rules, microRule)
		}
	}
	
	rg.logger.Infof("规则生成完成, 共生成 %d 条规则", len(rules))
	return rules, nil
}

// generateModSecurityRule 生成ModSecurity规则
func (rg *RuleGenerator) generateModSecurityRule(pattern *model.AttackPattern) *model.GeneratedRule {
	ruleID := rg.getNextRuleID()
	
	// 构建SecLang指令
	var directive strings.Builder
	directive.WriteString(fmt.Sprintf("SecRule REQUEST_URI|ARGS|REQUEST_HEADERS "))
	
	// 根据模式类型选择匹配方式
	switch pattern.PatternType {
	case "sql_injection":
		directive.WriteString(`"@rx (?i)(union|select|insert|update|delete|drop|create|alter|exec).*from" `)
	case "xss":
		directive.WriteString(`"@rx (?i)(<script|<iframe|javascript:|onerror=|onload=|<img[^>]+src)" `)
	case "path_traversal":
		directive.WriteString(`"@rx (\\.\\./|\\.\\.\\\\\\\\\|%2e%2e%2f)" `)
	case "command_injection":
		directive.WriteString(`"@rx (?i)(;\\s*(ls|cat|wget|curl|bash|sh)|\\||&&|\\$\\(|\\x60)" `)
	default:
		if pattern.PayloadRegex != "" {
			directive.WriteString(fmt.Sprintf(`"@rx %s" `, pattern.PayloadRegex))
		} else {
			return nil
		}
	}
	
	// 添加动作和元数据
	action := "deny"
	if pattern.Severity == "low" || pattern.Severity == "medium" {
		action = "log"
	}
	
	directive.WriteString(fmt.Sprintf(`"id:%d,phase:2,deny,status:403,`, ruleID))
	directive.WriteString(fmt.Sprintf(`msg:'AI检测: %s',`, pattern.Name))
	directive.WriteString(fmt.Sprintf(`severity:'%s',`, strings.ToUpper(pattern.Severity)))
	directive.WriteString(fmt.Sprintf(`tag:'ai-generated',tag:'%s',`, pattern.PatternType))
	directive.WriteString(fmt.Sprintf(`logdata:'Matched Pattern: %s'"`, pattern.ID.Hex()))
	
	rule := &model.GeneratedRule{
		Name:             fmt.Sprintf("AI规则 - %s", pattern.Name),
		Description:      fmt.Sprintf("基于模式'%s'生成的ModSecurity规则\\n%s", pattern.Name, pattern.Description),
		RuleType:         "modsecurity",
		SecLangDirective: directive.String(),
		PatternID:        pattern.ID.Hex(),
		PatternName:      pattern.Name,
		Confidence:       pattern.Confidence,
		Severity:         pattern.Severity,
		Action:           action,
		Status:           "pending",
		ReviewRequired:   true,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	
	return rule
}

// generateMicroRule 生成MicroRule
func (rg *RuleGenerator) generateMicroRule(pattern *model.AttackPattern) *model.GeneratedRule {
	// 构建MicroRule条件
	conditions := make([]interface{}, 0)
	
	// IP条件
	if pattern.IPPattern != "" && pattern.IPPattern != "0.0.0.0/0" {
		ipCondition := map[string]interface{}{
			"type":        "simple",
			"target":      "source_ip",
			"match_type":  "in_cidr",
			"match_value": pattern.IPPattern,
		}
		conditions = append(conditions, ipCondition)
	}
	
	// Path条件
	if pattern.PathPattern != "" {
		pathCondition := map[string]interface{}{
			"type":        "simple",
			"target":      "path",
			"match_type":  "regex",
			"match_value": pattern.PathPattern,
		}
		conditions = append(conditions, pathCondition)
	}
	
	// URL条件
	if pattern.URLPattern != "" && pattern.URLPattern != pattern.PathPattern {
		urlCondition := map[string]interface{}{
			"type":        "simple",
			"target":      "url",
			"match_type":  "regex",
			"match_value": pattern.URLPattern,
		}
		conditions = append(conditions, urlCondition)
	}
	
	// 如果没有任何条件，跳过
	if len(conditions) == 0 {
		return nil
	}
	
	// 构建最终条件
	var condition interface{}
	if len(conditions) == 1 {
		condition = conditions[0]
	} else {
		condition = map[string]interface{}{
			"type":       "composite",
			"operator":   "AND",
			"conditions": conditions,
		}
	}
	
	// 转换为BSON
	conditionBSON, err := bson.Marshal(condition)
	if err != nil {
		rg.logger.Errorf("MicroRule条件序列化失败: %v", err)
		return nil
	}
	
	rule := &model.GeneratedRule{
		Name:               fmt.Sprintf("AI微规则 - %s", pattern.Name),
		Description:        fmt.Sprintf("基于模式'%s'生成的MicroRule规则\\n%s", pattern.Name, pattern.Description),
		RuleType:           "micro_rule",
		MicroRuleCondition: conditionBSON,
		PatternID:          pattern.ID.Hex(),
		PatternName:        pattern.Name,
		Confidence:         pattern.Confidence,
		Severity:           pattern.Severity,
		Action:             "block",
		Status:             "pending",
		ReviewRequired:     true,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
	
	return rule
}

// SaveGeneratedRule 保存生成的规则
func (rg *RuleGenerator) SaveGeneratedRule(rule *model.GeneratedRule) error {
	collection := rg.db.Collection("generated_rules")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	_, err := collection.InsertOne(ctx, rule)
	if err != nil {
		return fmt.Errorf("保存规则失败: %w", err)
	}
	
	rg.logger.Infof("保存生成的规则: %s (类型: %s)", rule.Name, rule.RuleType)
	return nil
}

// ApproveRule 批准规则
func (rg *RuleGenerator) ApproveRule(ruleID string, reviewedBy string, comment string) error {
	collection := rg.db.Collection("generated_rules")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	objID, err := bson.ObjectIDFromHex(ruleID)
	if err != nil {
		return fmt.Errorf("无效的规则ID: %w", err)
	}
	
	update := bson.M{
		"$set": bson.M{
			"status":        "approved",
			"reviewedBy":    reviewedBy,
			"reviewedAt":    time.Now(),
			"reviewComment": comment,
			"updatedAt":     time.Now(),
		},
	}
	
	result, err := collection.UpdateByID(ctx, objID, update)
	if err != nil {
		return fmt.Errorf("批准规则失败: %w", err)
	}
	
	if result.MatchedCount == 0 {
		return fmt.Errorf("规则不存在")
	}
	
	rg.logger.Infof("规则已批准: %s, 审核人: %s", ruleID, reviewedBy)
	return nil
}

// RejectRule 拒绝规则
func (rg *RuleGenerator) RejectRule(ruleID string, reviewedBy string, reason string) error {
	collection := rg.db.Collection("generated_rules")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	objID, err := bson.ObjectIDFromHex(ruleID)
	if err != nil {
		return fmt.Errorf("无效的规则ID: %w", err)
	}
	
	update := bson.M{
		"$set": bson.M{
			"status":        "rejected",
			"reviewedBy":    reviewedBy,
			"reviewedAt":    time.Now(),
			"reviewComment": reason,
			"updatedAt":     time.Now(),
		},
	}
	
	result, err := collection.UpdateByID(ctx, objID, update)
	if err != nil {
		return fmt.Errorf("拒绝规则失败: %w", err)
	}
	
	if result.MatchedCount == 0 {
		return fmt.Errorf("规则不存在")
	}
	
	rg.logger.Infof("规则已拒绝: %s, 审核人: %s, 原因: %s", ruleID, reviewedBy, reason)
	return nil
}

// DeployRule 部署规则
func (rg *RuleGenerator) DeployRule(ruleID string) error {
	collection := rg.db.Collection("generated_rules")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	objID, err := bson.ObjectIDFromHex(ruleID)
	if err != nil {
		return fmt.Errorf("无效的规则ID: %w", err)
	}
	
	// 获取规则
	var rule model.GeneratedRule
	err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&rule)
	if err != nil {
		return fmt.Errorf("规则不存在: %w", err)
	}
	
	// 检查状态
	if rule.Status != "approved" {
		return fmt.Errorf("只能部署已批准的规则, 当前状态: %s", rule.Status)
	}
	
	// 根据规则类型部署
	var deployedRuleID string
	switch rule.RuleType {
	case "modsecurity":
		// TODO: 部署到ModSecurity
		deployedRuleID = fmt.Sprintf("modsec_%s", objID.Hex())
		rg.logger.Infof("部署ModSecurity规则: %s", rule.SecLangDirective)
		
	case "micro_rule":
		// TODO: 部署到MicroRule
		deployedRuleID = fmt.Sprintf("micro_%s", objID.Hex())
		rg.logger.Infof("部署MicroRule规则: %s", rule.Name)
		
	default:
		return fmt.Errorf("不支持的规则类型: %s", rule.RuleType)
	}
	
	// 更新状态
	update := bson.M{
		"$set": bson.M{
			"status":         "deployed",
			"deployedAt":     time.Now(),
			"deployedRuleId": deployedRuleID,
			"updatedAt":      time.Now(),
		},
	}
	
	_, err = collection.UpdateByID(ctx, objID, update)
	if err != nil {
		return fmt.Errorf("更新部署状态失败: %w", err)
	}
	
	rg.logger.Infof("规则部署成功: %s -> %s", ruleID, deployedRuleID)
	return nil
}

// GetPendingRules 获取待审核规则
func (rg *RuleGenerator) GetPendingRules() ([]*model.GeneratedRule, error) {
	collection := rg.db.Collection("generated_rules")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	filter := bson.M{"status": "pending"}
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var rules []*model.GeneratedRule
	if err := cursor.All(ctx, &rules); err != nil {
		return nil, err
	}
	
	return rules, nil
}

// GetRuleStats 获取规则统计
func (rg *RuleGenerator) GetRuleStats() (map[string]interface{}, error) {
	collection := rg.db.Collection("generated_rules")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// 按状态统计
	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.M{
			"_id":   "$status",
			"count": bson.M{"$sum": 1},
		}}},
	}
	
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	statusStats := make(map[string]int)
	for cursor.Next(ctx) {
		var result struct {
			Status string `bson:"_id"`
			Count  int    `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			continue
		}
		statusStats[result.Status] = result.Count
	}
	
	return map[string]interface{}{
		"byStatus": statusStats,
	}, nil
}

// getNextRuleID 获取下一个规则ID
func (rg *RuleGenerator) getNextRuleID() int {
	rg.nextRuleID++
	return rg.nextRuleID
}
