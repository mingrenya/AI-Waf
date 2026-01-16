// server/repository/ai_analyzer.go
package repository

import (
	"context"
	"errors"
	"time"

	"github.com/mingrenya/AI-Waf/pkg/model"
	"github.com/mingrenya/AI-Waf/server/config"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	ErrAttackPatternNotFound      = errors.New("攻击模式不存在")
	ErrGeneratedRuleNotFound      = errors.New("生成的规则不存在")
	ErrAIAnalyzerConfigNotFound   = errors.New("AI分析器配置不存在")
	ErrMCPConversationNotFound    = errors.New("MCP对话不存在")
)

// AttackPatternRepository 攻击模式仓库接口
type AttackPatternRepository interface {
	Create(ctx context.Context, pattern *model.AttackPattern) error
	GetByID(ctx context.Context, id bson.ObjectID) (*model.AttackPattern, error)
	List(ctx context.Context, filter bson.D, page, size int64) ([]model.AttackPattern, int64, error)
	Update(ctx context.Context, pattern *model.AttackPattern) error
	Delete(ctx context.Context, id bson.ObjectID) error
	GetBySeverity(ctx context.Context, severity string, limit int64) ([]model.AttackPattern, error)
	GetByTimeRange(ctx context.Context, start, end time.Time, page, size int64) ([]model.AttackPattern, int64, error)
	Count(ctx context.Context, filter bson.D) (int64, error)
	GetDB() *mongo.Database
}

// GeneratedRuleRepository 生成规则仓库接口
type GeneratedRuleRepository interface {
	Create(ctx context.Context, rule *model.GeneratedRule) error
	GetByID(ctx context.Context, id bson.ObjectID) (*model.GeneratedRule, error)
	List(ctx context.Context, filter bson.D, page, size int64) ([]model.GeneratedRule, int64, error)
	Update(ctx context.Context, rule *model.GeneratedRule) error
	Delete(ctx context.Context, id bson.ObjectID) error
	GetByStatus(ctx context.Context, status string, page, size int64) ([]model.GeneratedRule, int64, error)
	GetByPatternID(ctx context.Context, patternID bson.ObjectID) ([]model.GeneratedRule, error)
	GetPendingReview(ctx context.Context, page, size int64) ([]model.GeneratedRule, int64, error)
	UpdateStatus(ctx context.Context, id bson.ObjectID, status string, reviewedBy string, reviewComment string) error
	Count(ctx context.Context, filter bson.D) (int64, error)
}

// AIAnalyzerConfigRepository AI分析器配置仓库接口
type AIAnalyzerConfigRepository interface {
	Get(ctx context.Context) (*model.AIAnalyzerConfig, error)
	Update(ctx context.Context, config *model.AIAnalyzerConfig) error
	CreateDefault(ctx context.Context) error
}

// MCPConversationRepository MCP对话仓库接口
type MCPConversationRepository interface {
	Create(ctx context.Context, conversation *model.MCPConversation) error
	GetByID(ctx context.Context, id bson.ObjectID) (*model.MCPConversation, error)
	List(ctx context.Context, patternID *bson.ObjectID, page, size int64) ([]model.MCPConversation, int64, error)
	Delete(ctx context.Context, id bson.ObjectID) error
	Count(ctx context.Context, filter bson.D) (int64, error)
}

// ============================================
// MongoDB 实现
// ============================================

// MongoAttackPatternRepository MongoDB实现的攻击模式仓库
type MongoAttackPatternRepository struct {
	collection *mongo.Collection
	db         *mongo.Database
	logger     zerolog.Logger
}

// NewAttackPatternRepository 创建攻击模式仓库
func NewAttackPatternRepository(db *mongo.Database) AttackPatternRepository {
	var pattern model.AttackPattern
	collection := db.Collection(pattern.GetCollectionName())
	logger := config.GetRepositoryLogger("attack_pattern")

	// 创建索引
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 检测时间索引（降序）
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "detected_at", Value: -1}},
	})
	if err != nil {
		logger.Error().Err(err).Msg("创建detected_at索引失败")
	}

	// 严重程度索引
	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "severity", Value: 1}},
	})
	if err != nil {
		logger.Error().Err(err).Msg("创建severity索引失败")
	}

	return &MongoAttackPatternRepository{
		collection: collection,
		db:         db,
		logger:     logger,
	}
}

func (r *MongoAttackPatternRepository) Create(ctx context.Context, pattern *model.AttackPattern) error {
	result, err := r.collection.InsertOne(ctx, pattern)
	if err != nil {
		r.logger.Error().Err(err).Msg("插入攻击模式时出错")
		return err
	}
	pattern.ID = result.InsertedID.(bson.ObjectID)
	return nil
}

func (r *MongoAttackPatternRepository) GetByID(ctx context.Context, id bson.ObjectID) (*model.AttackPattern, error) {
	var pattern model.AttackPattern
	err := r.collection.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&pattern)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrAttackPatternNotFound
		}
		r.logger.Error().Err(err).Str("id", id.Hex()).Msg("查询攻击模式时出错")
		return nil, err
	}
	return &pattern, nil
}

func (r *MongoAttackPatternRepository) List(ctx context.Context, filter bson.D, page, size int64) ([]model.AttackPattern, int64, error) {
	skip := (page - 1) * size

	findOptions := options.Find().
		SetSkip(skip).
		SetLimit(size).
		SetSort(bson.D{{Key: "detected_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		r.logger.Error().Err(err).Msg("查询攻击模式列表时出错")
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var patterns []model.AttackPattern
	if err = cursor.All(ctx, &patterns); err != nil {
		r.logger.Error().Err(err).Msg("解析攻击模式列表时出错")
		return nil, 0, err
	}

	// 获取总数
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		r.logger.Error().Err(err).Msg("统计攻击模式总数时出错")
		return nil, 0, err
	}

	return patterns, total, nil
}

func (r *MongoAttackPatternRepository) Update(ctx context.Context, pattern *model.AttackPattern) error {
	filter := bson.D{{Key: "_id", Value: pattern.ID}}
	update := bson.D{{Key: "$set", Value: pattern}}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Str("id", pattern.ID.Hex()).Msg("更新攻击模式时出错")
		return err
	}

	if result.MatchedCount == 0 {
		return ErrAttackPatternNotFound
	}

	return nil
}

func (r *MongoAttackPatternRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.D{{Key: "_id", Value: id}})
	if err != nil {
		r.logger.Error().Err(err).Str("id", id.Hex()).Msg("删除攻击模式时出错")
		return err
	}

	if result.DeletedCount == 0 {
		return ErrAttackPatternNotFound
	}

	return nil
}

func (r *MongoAttackPatternRepository) GetBySeverity(ctx context.Context, severity string, limit int64) ([]model.AttackPattern, error) {
	filter := bson.D{{Key: "severity", Value: severity}}
	findOptions := options.Find().
		SetLimit(limit).
		SetSort(bson.D{{Key: "detected_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		r.logger.Error().Err(err).Str("severity", severity).Msg("按严重程度查询攻击模式时出错")
		return nil, err
	}
	defer cursor.Close(ctx)

	var patterns []model.AttackPattern
	if err = cursor.All(ctx, &patterns); err != nil {
		r.logger.Error().Err(err).Msg("解析攻击模式列表时出错")
		return nil, err
	}

	return patterns, nil
}

func (r *MongoAttackPatternRepository) GetByTimeRange(ctx context.Context, start, end time.Time, page, size int64) ([]model.AttackPattern, int64, error) {
	filter := bson.D{
		{Key: "detected_at", Value: bson.D{
			{Key: "$gte", Value: start},
			{Key: "$lte", Value: end},
		}},
	}
	return r.List(ctx, filter, page, size)
}

func (r *MongoAttackPatternRepository) Count(ctx context.Context, filter bson.D) (int64, error) {
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		r.logger.Error().Err(err).Msg("统计攻击模式数量时出错")
		return 0, err
	}
	return count, nil
}

// GetDB 获取数据库实例
func (r *MongoAttackPatternRepository) GetDB() *mongo.Database {
	return r.db
}

// MongoGeneratedRuleRepository MongoDB实现的生成规则仓库
type MongoGeneratedRuleRepository struct {
	collection *mongo.Collection
	logger     zerolog.Logger
}

// NewGeneratedRuleRepository 创建生成规则仓库
func NewGeneratedRuleRepository(db *mongo.Database) GeneratedRuleRepository {
	var rule model.GeneratedRule
	collection := db.Collection(rule.GetCollectionName())
	logger := config.GetRepositoryLogger("generated_rule")

	// 创建索引
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 状态索引
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "status", Value: 1}},
	})
	if err != nil {
		logger.Error().Err(err).Msg("创建status索引失败")
	}

	// 模式ID索引
	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "pattern_id", Value: 1}},
	})
	if err != nil {
		logger.Error().Err(err).Msg("创建pattern_id索引失败")
	}

	// 创建时间索引（降序）
	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "created_at", Value: -1}},
	})
	if err != nil {
		logger.Error().Err(err).Msg("创建created_at索引失败")
	}

	return &MongoGeneratedRuleRepository{
		collection: collection,
		logger:     logger,
	}
}

func (r *MongoGeneratedRuleRepository) Create(ctx context.Context, rule *model.GeneratedRule) error {
	result, err := r.collection.InsertOne(ctx, rule)
	if err != nil {
		r.logger.Error().Err(err).Msg("插入生成规则时出错")
		return err
	}
	rule.ID = result.InsertedID.(bson.ObjectID)
	return nil
}

func (r *MongoGeneratedRuleRepository) GetByID(ctx context.Context, id bson.ObjectID) (*model.GeneratedRule, error) {
	var rule model.GeneratedRule
	err := r.collection.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&rule)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrGeneratedRuleNotFound
		}
		r.logger.Error().Err(err).Str("id", id.Hex()).Msg("查询生成规则时出错")
		return nil, err
	}
	return &rule, nil
}

func (r *MongoGeneratedRuleRepository) List(ctx context.Context, filter bson.D, page, size int64) ([]model.GeneratedRule, int64, error) {
	skip := (page - 1) * size

	findOptions := options.Find().
		SetSkip(skip).
		SetLimit(size).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		r.logger.Error().Err(err).Msg("查询生成规则列表时出错")
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var rules []model.GeneratedRule
	if err = cursor.All(ctx, &rules); err != nil {
		r.logger.Error().Err(err).Msg("解析生成规则列表时出错")
		return nil, 0, err
	}

	// 获取总数
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		r.logger.Error().Err(err).Msg("统计生成规则总数时出错")
		return nil, 0, err
	}

	return rules, total, nil
}

func (r *MongoGeneratedRuleRepository) Update(ctx context.Context, rule *model.GeneratedRule) error {
	filter := bson.D{{Key: "_id", Value: rule.ID}}
	update := bson.D{{Key: "$set", Value: rule}}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Str("id", rule.ID.Hex()).Msg("更新生成规则时出错")
		return err
	}

	if result.MatchedCount == 0 {
		return ErrGeneratedRuleNotFound
	}

	return nil
}

func (r *MongoGeneratedRuleRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.D{{Key: "_id", Value: id}})
	if err != nil {
		r.logger.Error().Err(err).Str("id", id.Hex()).Msg("删除生成规则时出错")
		return err
	}

	if result.DeletedCount == 0 {
		return ErrGeneratedRuleNotFound
	}

	return nil
}

func (r *MongoGeneratedRuleRepository) GetByStatus(ctx context.Context, status string, page, size int64) ([]model.GeneratedRule, int64, error) {
	filter := bson.D{{Key: "status", Value: status}}
	return r.List(ctx, filter, page, size)
}

func (r *MongoGeneratedRuleRepository) GetByPatternID(ctx context.Context, patternID bson.ObjectID) ([]model.GeneratedRule, error) {
	filter := bson.D{{Key: "pattern_id", Value: patternID}}
	findOptions := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		r.logger.Error().Err(err).Str("pattern_id", patternID.Hex()).Msg("按模式ID查询生成规则时出错")
		return nil, err
	}
	defer cursor.Close(ctx)

	var rules []model.GeneratedRule
	if err = cursor.All(ctx, &rules); err != nil {
		r.logger.Error().Err(err).Msg("解析生成规则列表时出错")
		return nil, err
	}

	return rules, nil
}

func (r *MongoGeneratedRuleRepository) GetPendingReview(ctx context.Context, page, size int64) ([]model.GeneratedRule, int64, error) {
	return r.GetByStatus(ctx, "pending", page, size)
}

func (r *MongoGeneratedRuleRepository) UpdateStatus(ctx context.Context, id bson.ObjectID, status string, reviewedBy string, reviewComment string) error {
	filter := bson.D{{Key: "_id", Value: id}}
	update := bson.D{{Key: "$set", Value: bson.D{
		{Key: "status", Value: status},
		{Key: "reviewed_by", Value: reviewedBy},
		{Key: "review_comment", Value: reviewComment},
		{Key: "reviewed_at", Value: time.Now()},
	}}}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Str("id", id.Hex()).Msg("更新生成规则状态时出错")
		return err
	}

	if result.MatchedCount == 0 {
		return ErrGeneratedRuleNotFound
	}

	return nil
}

func (r *MongoGeneratedRuleRepository) Count(ctx context.Context, filter bson.D) (int64, error) {
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		r.logger.Error().Err(err).Msg("统计生成规则数量时出错")
		return 0, err
	}
	return count, nil
}

// MongoAIAnalyzerConfigRepository MongoDB实现的AI分析器配置仓库
type MongoAIAnalyzerConfigRepository struct {
	collection *mongo.Collection
	logger     zerolog.Logger
}

// NewAIAnalyzerConfigRepository 创建AI分析器配置仓库
func NewAIAnalyzerConfigRepository(db *mongo.Database) AIAnalyzerConfigRepository {
	var configModel model.AIAnalyzerConfig
	collection := db.Collection(configModel.GetCollectionName())
	logger := config.GetRepositoryLogger("ai_analyzer_config")

	return &MongoAIAnalyzerConfigRepository{
		collection: collection,
		logger:     logger,
	}
}

func (r *MongoAIAnalyzerConfigRepository) Get(ctx context.Context) (*model.AIAnalyzerConfig, error) {
	var config model.AIAnalyzerConfig
	err := r.collection.FindOne(ctx, bson.D{}).Decode(&config)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrAIAnalyzerConfigNotFound
		}
		r.logger.Error().Err(err).Msg("查询AI分析器配置时出错")
		return nil, err
	}
	return &config, nil
}

func (r *MongoAIAnalyzerConfigRepository) Update(ctx context.Context, config *model.AIAnalyzerConfig) error {
	config.UpdatedAt = time.Now()
	
	filter := bson.D{{Key: "_id", Value: config.ID}}
	update := bson.D{{Key: "$set", Value: config}}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Msg("更新AI分析器配置时出错")
		return err
	}

	if result.MatchedCount == 0 {
		return ErrAIAnalyzerConfigNotFound
	}

	return nil
}

func (r *MongoAIAnalyzerConfigRepository) CreateDefault(ctx context.Context) error {
	cfg := &model.AIAnalyzerConfig{
		ID:                bson.NewObjectID(),
		Name:              "default",
		Enabled:           false,
		AnalysisInterval:  30,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
	
	// 初始化嵌套结构
	cfg.PatternDetection.Enabled = true
	cfg.PatternDetection.MinSamples = 100
	cfg.PatternDetection.AnomalyThreshold = 2.0
	cfg.PatternDetection.ClusteringMethod = "kmeans"
	cfg.PatternDetection.TimeWindow = 24
	
	cfg.RuleGeneration.Enabled = true
	cfg.RuleGeneration.ConfidenceThreshold = 0.7
	cfg.RuleGeneration.AutoDeploy = false
	cfg.RuleGeneration.ReviewRequired = true
	cfg.RuleGeneration.DefaultAction = "block"

	_, err := r.collection.InsertOne(ctx, cfg)
	if err != nil {
		r.logger.Error().Err(err).Msg("创建默认AI分析器配置时出错")
		return err
	}

	return nil
}

// MongoMCPConversationRepository MongoDB实现的MCP对话仓库
type MongoMCPConversationRepository struct {
	collection *mongo.Collection
	logger     zerolog.Logger
}

// NewMCPConversationRepository 创建MCP对话仓库
func NewMCPConversationRepository(db *mongo.Database) MCPConversationRepository {
	var conversation model.MCPConversation
	collection := db.Collection(conversation.GetCollectionName())
	logger := config.GetRepositoryLogger("mcp_conversation")

	// 创建索引
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 模式ID索引
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "pattern_id", Value: 1}},
	})
	if err != nil {
		logger.Error().Err(err).Msg("创建pattern_id索引失败")
	}

	// 创建时间索引（降序）
	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "created_at", Value: -1}},
	})
	if err != nil {
		logger.Error().Err(err).Msg("创建created_at索引失败")
	}

	return &MongoMCPConversationRepository{
		collection: collection,
		logger:     logger,
	}
}

func (r *MongoMCPConversationRepository) Create(ctx context.Context, conversation *model.MCPConversation) error {
	result, err := r.collection.InsertOne(ctx, conversation)
	if err != nil {
		r.logger.Error().Err(err).Msg("插入MCP对话时出错")
		return err
	}
	conversation.ID = result.InsertedID.(bson.ObjectID)
	return nil
}

func (r *MongoMCPConversationRepository) GetByID(ctx context.Context, id bson.ObjectID) (*model.MCPConversation, error) {
	var conversation model.MCPConversation
	err := r.collection.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&conversation)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrMCPConversationNotFound
		}
		r.logger.Error().Err(err).Str("id", id.Hex()).Msg("查询MCP对话时出错")
		return nil, err
	}
	return &conversation, nil
}

func (r *MongoMCPConversationRepository) List(ctx context.Context, patternID *bson.ObjectID, page, size int64) ([]model.MCPConversation, int64, error) {
	filter := bson.D{}
	if patternID != nil {
		filter = bson.D{{Key: "pattern_id", Value: *patternID}}
	}

	skip := (page - 1) * size
	findOptions := options.Find().
		SetSkip(skip).
		SetLimit(size).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		r.logger.Error().Err(err).Msg("查询MCP对话列表时出错")
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var conversations []model.MCPConversation
	if err = cursor.All(ctx, &conversations); err != nil {
		r.logger.Error().Err(err).Msg("解析MCP对话列表时出错")
		return nil, 0, err
	}

	// 获取总数
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		r.logger.Error().Err(err).Msg("统计MCP对话总数时出错")
		return nil, 0, err
	}

	return conversations, total, nil
}

func (r *MongoMCPConversationRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.D{{Key: "_id", Value: id}})
	if err != nil {
		r.logger.Error().Err(err).Str("id", id.Hex()).Msg("删除MCP对话时出错")
		return err
	}

	if result.DeletedCount == 0 {
		return ErrMCPConversationNotFound
	}

	return nil
}

func (r *MongoMCPConversationRepository) Count(ctx context.Context, filter bson.D) (int64, error) {
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		r.logger.Error().Err(err).Msg("统计MCP对话数量时出错")
		return 0, err
	}
	return count, nil
}
