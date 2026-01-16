package repository

import (
	"context"
	"time"

	"github.com/mingrenya/AI-Waf/pkg/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	CollectionAdaptiveThrottlingConfig = "adaptive_throttling_config"
	CollectionTrafficPatterns          = "traffic_patterns"
	CollectionBaselineValues           = "baseline_values"
	CollectionThrottleAdjustmentLogs   = "throttle_adjustment_logs"
)

// AdaptiveThrottlingRepository 自适应限流仓储接口
type AdaptiveThrottlingRepository interface {
	// 配置管理
	GetConfig(ctx context.Context) (*model.AdaptiveThrottlingConfig, error)
	CreateConfig(ctx context.Context, config *model.AdaptiveThrottlingConfig) error
	UpdateConfig(ctx context.Context, config *model.AdaptiveThrottlingConfig) error
	DeleteConfig(ctx context.Context) error

	// 流量模式
	GetTrafficPatterns(ctx context.Context, filter bson.M, skip, limit int64) ([]*model.TrafficPattern, int64, error)
	CreateTrafficPattern(ctx context.Context, pattern *model.TrafficPattern) error

	// 基线值
	GetBaselines(ctx context.Context, filter bson.M) ([]*model.BaselineValue, error)
	GetBaselineByType(ctx context.Context, typ string) (*model.BaselineValue, error)
	UpsertBaseline(ctx context.Context, baseline *model.BaselineValue) error

	// 调整日志
	GetAdjustmentLogs(ctx context.Context, filter bson.M, skip, limit int64) ([]*model.ThrottleAdjustmentLog, int64, error)
	CreateAdjustmentLog(ctx context.Context, log *model.ThrottleAdjustmentLog) error
	GetRecentAdjustmentCount(ctx context.Context, since time.Time) (int64, error)
}

type adaptiveThrottlingRepo struct {
	db *mongo.Database
}

// NewAdaptiveThrottlingRepository 创建自适应限流仓储实例
func NewAdaptiveThrottlingRepository(db *mongo.Database) AdaptiveThrottlingRepository {
	return &adaptiveThrottlingRepo{db: db}
}

// GetConfig 获取配置
func (r *adaptiveThrottlingRepo) GetConfig(ctx context.Context) (*model.AdaptiveThrottlingConfig, error) {
	collection := r.db.Collection(CollectionAdaptiveThrottlingConfig)
	var config model.AdaptiveThrottlingConfig
	err := collection.FindOne(ctx, bson.M{}).Decode(&config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// CreateConfig 创建配置
func (r *adaptiveThrottlingRepo) CreateConfig(ctx context.Context, config *model.AdaptiveThrottlingConfig) error {
	collection := r.db.Collection(CollectionAdaptiveThrottlingConfig)
	config.ID = bson.NewObjectID().Hex()
	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()
	_, err := collection.InsertOne(ctx, config)
	return err
}

// UpdateConfig 更新配置
func (r *adaptiveThrottlingRepo) UpdateConfig(ctx context.Context, config *model.AdaptiveThrottlingConfig) error {
	collection := r.db.Collection(CollectionAdaptiveThrottlingConfig)
	config.UpdatedAt = time.Now()
	
	filter := bson.M{"_id": config.ID}
	update := bson.M{"$set": config}
	
	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	
	return nil
}

// DeleteConfig 删除配置
func (r *adaptiveThrottlingRepo) DeleteConfig(ctx context.Context) error {
	collection := r.db.Collection(CollectionAdaptiveThrottlingConfig)
	_, err := collection.DeleteMany(ctx, bson.M{})
	return err
}

// GetTrafficPatterns 获取流量模式列表
func (r *adaptiveThrottlingRepo) GetTrafficPatterns(ctx context.Context, filter bson.M, skip, limit int64) ([]*model.TrafficPattern, int64, error) {
	collection := r.db.Collection(CollectionTrafficPatterns)
	
	// 获取总数
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	
	// 查询数据
	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.M{"timestamp": -1})
	
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)
	
	var patterns []*model.TrafficPattern
	if err = cursor.All(ctx, &patterns); err != nil {
		return nil, 0, err
	}
	
	return patterns, total, nil
}

// CreateTrafficPattern 创建流量模式记录
func (r *adaptiveThrottlingRepo) CreateTrafficPattern(ctx context.Context, pattern *model.TrafficPattern) error {
	collection := r.db.Collection(CollectionTrafficPatterns)
	pattern.ID = bson.NewObjectID().Hex()
	_, err := collection.InsertOne(ctx, pattern)
	return err
}

// GetBaselines 获取基线值列表
func (r *adaptiveThrottlingRepo) GetBaselines(ctx context.Context, filter bson.M) ([]*model.BaselineValue, error) {
	collection := r.db.Collection(CollectionBaselineValues)
	
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var baselines []*model.BaselineValue
	if err = cursor.All(ctx, &baselines); err != nil {
		return nil, err
	}
	
	return baselines, nil
}

// GetBaselineByType 根据类型获取基线值
func (r *adaptiveThrottlingRepo) GetBaselineByType(ctx context.Context, typ string) (*model.BaselineValue, error) {
	collection := r.db.Collection(CollectionBaselineValues)
	
	var baseline model.BaselineValue
	err := collection.FindOne(ctx, bson.M{"type": typ}).Decode(&baseline)
	if err != nil {
		return nil, err
	}
	
	return &baseline, nil
}

// UpsertBaseline 插入或更新基线值
func (r *adaptiveThrottlingRepo) UpsertBaseline(ctx context.Context, baseline *model.BaselineValue) error {
	collection := r.db.Collection(CollectionBaselineValues)
	
	baseline.UpdatedAt = time.Now()
	
	filter := bson.M{"type": baseline.Type}
	update := bson.M{
		"$set": baseline,
		"$setOnInsert": bson.M{
			"_id":         bson.NewObjectID().Hex(),
			"calculatedAt": time.Now(),
		},
	}
	
	opts := options.UpdateOne().SetUpsert(true)
	_, err := collection.UpdateOne(ctx, filter, update, opts)
	
	return err
}

// GetAdjustmentLogs 获取调整日志列表
func (r *adaptiveThrottlingRepo) GetAdjustmentLogs(ctx context.Context, filter bson.M, skip, limit int64) ([]*model.ThrottleAdjustmentLog, int64, error) {
	collection := r.db.Collection(CollectionThrottleAdjustmentLogs)
	
	// 获取总数
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	
	// 查询数据
	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.M{"timestamp": -1})
	
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)
	
	var logs []*model.ThrottleAdjustmentLog
	if err = cursor.All(ctx, &logs); err != nil {
		return nil, 0, err
	}
	
	return logs, total, nil
}

// CreateAdjustmentLog 创建调整日志
func (r *adaptiveThrottlingRepo) CreateAdjustmentLog(ctx context.Context, log *model.ThrottleAdjustmentLog) error {
	collection := r.db.Collection(CollectionThrottleAdjustmentLogs)
	log.ID = bson.NewObjectID().Hex()
	log.Timestamp = time.Now()
	_, err := collection.InsertOne(ctx, log)
	return err
}

// GetRecentAdjustmentCount 获取最近调整次数
func (r *adaptiveThrottlingRepo) GetRecentAdjustmentCount(ctx context.Context, since time.Time) (int64, error) {
	collection := r.db.Collection(CollectionThrottleAdjustmentLogs)
	filter := bson.M{"timestamp": bson.M{"$gte": since}}
	return collection.CountDocuments(ctx, filter)
}
