package service

import (
	"context"
	"fmt"
	"time"

	"github.com/mingrenya/AI-Waf/pkg/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// ProtectionProfileService 保护配置文件服务接口
type ProtectionProfileService interface {
	// 获取所有配置文件
	ListProfiles(ctx context.Context) ([]model.ProtectionProfile, error)
	// 获取配置文件详情
	GetProfile(ctx context.Context, id string) (*model.ProtectionProfile, error)
	// 应用配置文件（批量创建规则）
	ApplyProfile(ctx context.Context, profileID string) (int, error)
	// 初始化默认配置文件
	InitializeDefaultProfiles(ctx context.Context) error
}

type protectionProfileServiceImpl struct {
	profileCollection  *mongo.Collection
	templateCollection *mongo.Collection
	ruleCollection     *mongo.Collection
	templateService    RuleTemplateService
}

// NewProtectionProfileService 创建保护配置文件服务
func NewProtectionProfileService(db *mongo.Database, templateService RuleTemplateService) ProtectionProfileService {
	return &protectionProfileServiceImpl{
		profileCollection:  db.Collection("protection_profile"),
		templateCollection: db.Collection("rule_template"),
		ruleCollection:     db.Collection("micro_rule"),
		templateService:    templateService,
	}
}

// ListProfiles 获取所有配置文件
func (s *protectionProfileServiceImpl) ListProfiles(ctx context.Context) ([]model.ProtectionProfile, error) {
	opts := options.Find().SetSort(bson.D{{Key: "level", Value: 1}})
	cursor, err := s.profileCollection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("查询配置文件失败: %w", err)
	}
	defer cursor.Close(ctx)

	var profiles []model.ProtectionProfile
	if err = cursor.All(ctx, &profiles); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return profiles, nil
}

// GetProfile 获取配置文件详情
func (s *protectionProfileServiceImpl) GetProfile(ctx context.Context, id string) (*model.ProtectionProfile, error) {
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("无效的配置文件ID: %w", err)
	}

	var profile model.ProtectionProfile
	err = s.profileCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&profile)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("配置文件不存在")
		}
		return nil, fmt.Errorf("查询配置文件失败: %w", err)
	}

	return &profile, nil
}

// ApplyProfile 应用配置文件
func (s *protectionProfileServiceImpl) ApplyProfile(ctx context.Context, profileID string) (int, error) {
	profile, err := s.GetProfile(ctx, profileID)
	if err != nil {
		return 0, err
	}

	createdCount := 0

	// 遍历配置文件中的所有模板ID
	for _, templateID := range profile.TemplateIDs {
		// 获取模板
		var template model.RuleTemplate
		err := s.templateCollection.FindOne(ctx, bson.M{"_id": templateID}).Decode(&template)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				continue // 跳过不存在的模板
			}
			return createdCount, fmt.Errorf("查询模板失败: %w", err)
		}

		// 检查是否已存在相同名称的规则
		existCount, err := s.ruleCollection.CountDocuments(ctx, bson.M{"name": template.Name})
		if err != nil {
			return createdCount, fmt.Errorf("检查规则存在失败: %w", err)
		}

		if existCount > 0 {
			// 规则已存在，跳过
			continue
		}

		// 从模板创建规则
		rule := &model.MicroRule{
			Name:      template.Name,
			Type:      template.RuleType,
			Status:    model.RuleEnabled,
			Priority:  template.Priority,
			Condition: template.Condition,
		}

		_, err = s.ruleCollection.InsertOne(ctx, rule)
		if err != nil {
			return createdCount, fmt.Errorf("创建规则失败: %w", err)
		}

		createdCount++
	}

	return createdCount, nil
}

// InitializeDefaultProfiles 初始化默认配置文件
func (s *protectionProfileServiceImpl) InitializeDefaultProfiles(ctx context.Context) error {
	// 检查是否已经初始化
	count, err := s.profileCollection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("检查配置文件失败: %w", err)
	}

	if count > 0 {
		// 已经有配置文件，跳过初始化
		return nil
	}

	// 获取所有模板用于创建配置文件
	templates, err := s.templateService.ListTemplates(ctx, "", "")
	if err != nil {
		return fmt.Errorf("获取模板失败: %w", err)
	}

	// 按照分类和严重等级组织模板
	templatesByCategory := make(map[string][]model.RuleTemplate)
	templatesBySeverity := make(map[string][]model.RuleTemplate)

	for _, t := range templates {
		templatesByCategory[t.Category] = append(templatesByCategory[t.Category], t)
		templatesBySeverity[t.Severity] = append(templatesBySeverity[t.Severity], t)
	}

	profiles := []model.ProtectionProfile{}

	// 1. 基础保护配置 - 只包含critical级别的规则
	basicTemplateIDs := []bson.ObjectID{}
	for _, t := range templatesBySeverity[model.SeverityCritical] {
		basicTemplateIDs = append(basicTemplateIDs, t.ID)
	}

	profiles = append(profiles, model.ProtectionProfile{
		Name:        "基础保护",
		Level:       model.ProtectionLevelBasic,
		Description: "包含最关键的安全规则，适合对性能要求高的场景",
		Categories:  []string{model.CategoryInjection},
		TemplateIDs: basicTemplateIDs,
		IsDefault:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	})

	// 2. 标准保护配置 - 包含critical和high级别的规则
	standardTemplateIDs := []bson.ObjectID{}
	for _, severity := range []string{model.SeverityCritical, model.SeverityHigh} {
		for _, t := range templatesBySeverity[severity] {
			standardTemplateIDs = append(standardTemplateIDs, t.ID)
		}
	}

	profiles = append(profiles, model.ProtectionProfile{
		Name:        "标准保护",
		Level:       model.ProtectionLevelStandard,
		Description: "平衡安全性和性能，适合大多数Web应用",
		Categories: []string{
			model.CategoryInjection,
			model.CategoryBrokenAccessControl,
			model.CategorySecurityMisconfiguration,
			model.CategorySSRF,
		},
		TemplateIDs: standardTemplateIDs,
		IsDefault:   true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	})

	// 3. 严格保护配置 - 包含所有级别的规则
	strictTemplateIDs := []bson.ObjectID{}
	for _, t := range templates {
		strictTemplateIDs = append(strictTemplateIDs, t.ID)
	}

	profiles = append(profiles, model.ProtectionProfile{
		Name:        "严格保护",
		Level:       model.ProtectionLevelStrict,
		Description: "最全面的安全保护，包含所有OWASP Top 10防护规则",
		Categories: []string{
			model.CategoryBrokenAccessControl,
			model.CategoryCryptographicFailures,
			model.CategoryInjection,
			model.CategoryInsecureDesign,
			model.CategorySecurityMisconfiguration,
			model.CategoryVulnerableComponents,
			model.CategoryAuthenticationFailures,
			model.CategoryDataIntegrityFailures,
			model.CategoryLoggingFailures,
			model.CategorySSRF,
		},
		TemplateIDs: strictTemplateIDs,
		IsDefault:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	})

	// 插入配置文件
	var docs []interface{}
	for i := range profiles {
		docs = append(docs, profiles[i])
	}

	_, err = s.profileCollection.InsertMany(ctx, docs)
	if err != nil {
		return fmt.Errorf("初始化配置文件失败: %w", err)
	}

	return nil
}
