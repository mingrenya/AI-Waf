package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/mingrenya/AI-Waf/server/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// AlertChannelRepository 告警渠道仓库接口
type AlertChannelRepository interface {
	Create(ctx context.Context, channel *model.AlertChannel) error
	GetByID(ctx context.Context, id string) (*model.AlertChannel, error)
	GetAll(ctx context.Context) ([]*model.AlertChannel, error)
	GetEnabled(ctx context.Context) ([]*model.AlertChannel, error)
	Update(ctx context.Context, id string, updates map[string]interface{}) error
	Delete(ctx context.Context, id string) error
}

type alertChannelRepository struct {
	collection *mongo.Collection
}

// NewAlertChannelRepository 创建告警渠道仓库
func NewAlertChannelRepository(db *mongo.Database) AlertChannelRepository {
	return &alertChannelRepository{
		collection: db.Collection(model.AlertChannel{}.GetCollectionName()),
	}
}

func (r *alertChannelRepository) Create(ctx context.Context, channel *model.AlertChannel) error {
	channel.ID = bson.NewObjectID()
	channel.CreatedAt = time.Now()
	channel.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, channel)
	return err
}

func (r *alertChannelRepository) GetByID(ctx context.Context, id string) (*model.AlertChannel, error) {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid channel ID: %w", err)
	}

	var channel model.AlertChannel
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&channel)
	if err != nil {
		return nil, err
	}

	return &channel, nil
}

func (r *alertChannelRepository) GetAll(ctx context.Context) ([]*model.AlertChannel, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var channels []*model.AlertChannel
	if err := cursor.All(ctx, &channels); err != nil {
		return nil, err
	}

	return channels, nil
}

func (r *alertChannelRepository) GetEnabled(ctx context.Context) ([]*model.AlertChannel, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"enabled": true})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var channels []*model.AlertChannel
	if err := cursor.All(ctx, &channels); err != nil {
		return nil, err
	}

	return channels, nil
}

func (r *alertChannelRepository) Update(ctx context.Context, id string, updates map[string]interface{}) error {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid channel ID: %w", err)
	}

	updates["updated_at"] = time.Now()

	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": updates},
	)

	return err
}

func (r *alertChannelRepository) Delete(ctx context.Context, id string) error {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid channel ID: %w", err)
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}

// AlertRuleRepository 告警规则仓库接口
type AlertRuleRepository interface {
	Create(ctx context.Context, rule *model.AlertRule) error
	GetByID(ctx context.Context, id string) (*model.AlertRule, error)
	GetAll(ctx context.Context) ([]*model.AlertRule, error)
	GetEnabled(ctx context.Context) ([]*model.AlertRule, error)
	Update(ctx context.Context, id string, updates map[string]interface{}) error
	Delete(ctx context.Context, id string) error
}

type alertRuleRepository struct {
	collection *mongo.Collection
}

// NewAlertRuleRepository 创建告警规则仓库
func NewAlertRuleRepository(db *mongo.Database) AlertRuleRepository {
	return &alertRuleRepository{
		collection: db.Collection(model.AlertRule{}.GetCollectionName()),
	}
}

func (r *alertRuleRepository) Create(ctx context.Context, rule *model.AlertRule) error {
	rule.ID = bson.NewObjectID()
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, rule)
	return err
}

func (r *alertRuleRepository) GetByID(ctx context.Context, id string) (*model.AlertRule, error) {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid rule ID: %w", err)
	}

	var rule model.AlertRule
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&rule)
	if err != nil {
		return nil, err
	}

	return &rule, nil
}

func (r *alertRuleRepository) GetAll(ctx context.Context) ([]*model.AlertRule, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var rules []*model.AlertRule
	if err := cursor.All(ctx, &rules); err != nil {
		return nil, err
	}

	return rules, nil
}

func (r *alertRuleRepository) GetEnabled(ctx context.Context) ([]*model.AlertRule, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"enabled": true})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var rules []*model.AlertRule
	if err := cursor.All(ctx, &rules); err != nil {
		return nil, err
	}

	return rules, nil
}

func (r *alertRuleRepository) Update(ctx context.Context, id string, updates map[string]interface{}) error {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid rule ID: %w", err)
	}

	updates["updated_at"] = time.Now()

	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": updates},
	)

	return err
}

func (r *alertRuleRepository) Delete(ctx context.Context, id string) error {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid rule ID: %w", err)
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}

// AlertHistoryRepository 告警历史仓库接口
type AlertHistoryRepository interface {
	Create(ctx context.Context, history *model.AlertHistory) error
	GetByID(ctx context.Context, id string) (*model.AlertHistory, error)
	Query(ctx context.Context, filter bson.M, page, pageSize int) ([]*model.AlertHistory, int64, error)
	Update(ctx context.Context, id string, updates map[string]interface{}) error
	GetStatistics(ctx context.Context, startTime, endTime time.Time) (map[string]interface{}, error)
}

type alertHistoryRepository struct {
	collection *mongo.Collection
}

// NewAlertHistoryRepository 创建告警历史仓库
func NewAlertHistoryRepository(db *mongo.Database) AlertHistoryRepository {
	return &alertHistoryRepository{
		collection: db.Collection(model.AlertHistory{}.GetCollectionName()),
	}
}

func (r *alertHistoryRepository) Create(ctx context.Context, history *model.AlertHistory) error {
	history.ID = bson.NewObjectID()
	history.TriggeredAt = time.Now()

	_, err := r.collection.InsertOne(ctx, history)
	return err
}

func (r *alertHistoryRepository) GetByID(ctx context.Context, id string) (*model.AlertHistory, error) {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid history ID: %w", err)
	}

	var history model.AlertHistory
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&history)
	if err != nil {
		return nil, err
	}

	return &history, nil
}

func (r *alertHistoryRepository) Query(ctx context.Context, filter bson.M, page, pageSize int) ([]*model.AlertHistory, int64, error) {
	// 计算总数
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// 计算分页
	skip := (page - 1) * pageSize
	opts := options.Find().
		SetSort(bson.M{"triggered_at": -1}).
		SetSkip(int64(skip)).
		SetLimit(int64(pageSize))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var histories []*model.AlertHistory
	if err := cursor.All(ctx, &histories); err != nil {
		return nil, 0, err
	}

	return histories, total, nil
}

func (r *alertHistoryRepository) Update(ctx context.Context, id string, updates map[string]interface{}) error {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid history ID: %w", err)
	}

	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": updates},
	)

	return err
}

func (r *alertHistoryRepository) GetStatistics(ctx context.Context, startTime, endTime time.Time) (map[string]interface{}, error) {
	filter := bson.M{}
	if !startTime.IsZero() || !endTime.IsZero() {
		filter["triggered_at"] = bson.M{}
		if !startTime.IsZero() {
			filter["triggered_at"].(bson.M)["$gte"] = startTime
		}
		if !endTime.IsZero() {
			filter["triggered_at"].(bson.M)["$lte"] = endTime
		}
	}

	// 总告警数
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, err
	}

	// 按严重级别统计
	severityPipeline := mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{Key: "$group", Value: bson.M{
			"_id":   "$severity",
			"count": bson.M{"$sum": 1},
		}}},
	}
	severityCursor, err := r.collection.Aggregate(ctx, severityPipeline)
	if err != nil {
		return nil, err
	}
	defer severityCursor.Close(ctx)

	severityStats := make(map[string]int64)
	for severityCursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Count int64  `bson:"count"`
		}
		if err := severityCursor.Decode(&result); err != nil {
			continue
		}
		severityStats[result.ID] = result.Count
	}

	// 按状态统计
	statusPipeline := mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{Key: "$group", Value: bson.M{
			"_id":   "$status",
			"count": bson.M{"$sum": 1},
		}}},
	}
	statusCursor, err := r.collection.Aggregate(ctx, statusPipeline)
	if err != nil {
		return nil, err
	}
	defer statusCursor.Close(ctx)

	statusStats := make(map[string]int64)
	for statusCursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Count int64  `bson:"count"`
		}
		if err := statusCursor.Decode(&result); err != nil {
			continue
		}
		statusStats[result.ID] = result.Count
	}

	return map[string]interface{}{
		"total":             total,
		"alertsBySeverity":  severityStats,
		"alertsByStatus":    statusStats,
	}, nil
}

// AlertTemplateRepository 告警模板仓库接口
type AlertTemplateRepository interface {
	Create(ctx context.Context, template *model.AlertTemplate) error
	GetByID(ctx context.Context, id string) (*model.AlertTemplate, error)
	GetAll(ctx context.Context) ([]*model.AlertTemplate, error)
	GetBuiltIn(ctx context.Context) ([]*model.AlertTemplate, error)
	Update(ctx context.Context, id string, updates map[string]interface{}) error
	Delete(ctx context.Context, id string) error
}

type alertTemplateRepository struct {
	collection *mongo.Collection
}

// NewAlertTemplateRepository 创建告警模板仓库
func NewAlertTemplateRepository(db *mongo.Database) AlertTemplateRepository {
	return &alertTemplateRepository{
		collection: db.Collection(model.AlertTemplate{}.GetCollectionName()),
	}
}

func (r *alertTemplateRepository) Create(ctx context.Context, template *model.AlertTemplate) error {
	template.ID = bson.NewObjectID()
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, template)
	return err
}

func (r *alertTemplateRepository) GetByID(ctx context.Context, id string) (*model.AlertTemplate, error) {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid template ID: %w", err)
	}

	var template model.AlertTemplate
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&template)
	if err != nil {
		return nil, err
	}

	return &template, nil
}

func (r *alertTemplateRepository) GetAll(ctx context.Context) ([]*model.AlertTemplate, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var templates []*model.AlertTemplate
	if err := cursor.All(ctx, &templates); err != nil {
		return nil, err
	}

	return templates, nil
}

func (r *alertTemplateRepository) GetBuiltIn(ctx context.Context) ([]*model.AlertTemplate, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"is_built_in": true})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var templates []*model.AlertTemplate
	if err := cursor.All(ctx, &templates); err != nil {
		return nil, err
	}

	return templates, nil
}

func (r *alertTemplateRepository) Update(ctx context.Context, id string, updates map[string]interface{}) error {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid template ID: %w", err)
	}

	updates["updated_at"] = time.Now()

	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": updates},
	)

	return err
}

func (r *alertTemplateRepository) Delete(ctx context.Context, id string) error {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid template ID: %w", err)
	}

	// 不允许删除内置模板
	var template model.AlertTemplate
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&template)
	if err != nil {
		return err
	}

	if template.IsBuiltIn {
		return fmt.Errorf("cannot delete built-in template")
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}
