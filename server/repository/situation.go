package repository

import (
	"context"

	"github.com/mingrenya/AI-Waf/pkg/model"
	"github.com/mingrenya/AI-Waf/server/config"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// SituationRepository 态势感知数据访问接口
type SituationRepository interface {
	// 规则管理
	ListRules(ctx context.Context) ([]model.SituationRule, error)
	ListEnabledRules(ctx context.Context) ([]model.SituationRule, error)
	FindRuleByName(ctx context.Context, name string) (*model.SituationRule, error)
	GetRuleByID(ctx context.Context, id string) (*model.SituationRule, error)
	CreateRule(ctx context.Context, rule *model.SituationRule) error
	UpdateRule(ctx context.Context, id string, rule *model.SituationRule) error
	DeleteRule(ctx context.Context, id string) error

	// 攻击链
	ListChains(ctx context.Context, filter bson.M, skip, limit int64) ([]model.AttackChain, int64, error)
	GetChainByID(ctx context.Context, id string) (*model.AttackChain, error)
	UpsertChain(ctx context.Context, chain *model.AttackChain) error
	GetChainByIP(ctx context.Context, ip string) (*model.AttackChain, error)

	// 攻击者画像
	UpsertProfile(ctx context.Context, profile *model.AttackerProfile) error
	GetProfile(ctx context.Context, ip string) (*model.AttackerProfile, error)
	ListProfiles(ctx context.Context, sortBy string, skip, limit int64) ([]model.AttackerProfile, int64, error)
}

// MongoSituationRepository MongoDB 实现
type MongoSituationRepository struct {
	ruleCollection    *mongo.Collection
	chainCollection   *mongo.Collection
	profileCollection *mongo.Collection
	logger            zerolog.Logger
}

// NewSituationRepository 创建态势感知仓库
func NewSituationRepository(db *mongo.Database) SituationRepository {
	return &MongoSituationRepository{
		ruleCollection:    db.Collection("situation_rules"),
		chainCollection:   db.Collection("attack_chains"),
		profileCollection: db.Collection("attacker_profiles"),
		logger:            config.GetRepositoryLogger("situation"),
	}
}

// --- 规则管理 ---

func (r *MongoSituationRepository) ListRules(ctx context.Context) ([]model.SituationRule, error) {
	cursor, err := r.ruleCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var rules []model.SituationRule
	if err := cursor.All(ctx, &rules); err != nil {
		return nil, err
	}
	return rules, nil
}

func (r *MongoSituationRepository) ListEnabledRules(ctx context.Context) ([]model.SituationRule, error) {
	cursor, err := r.ruleCollection.Find(ctx, bson.M{"enabled": true})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var rules []model.SituationRule
	if err := cursor.All(ctx, &rules); err != nil {
		return nil, err
	}
	return rules, nil
}

func (r *MongoSituationRepository) FindRuleByName(ctx context.Context, name string) (*model.SituationRule, error) {
	var rule model.SituationRule
	err := r.ruleCollection.FindOne(ctx, bson.M{"name": name}).Decode(&rule)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *MongoSituationRepository) GetRuleByID(ctx context.Context, id string) (*model.SituationRule, error) {
	var rule model.SituationRule
	err := r.ruleCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&rule)
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *MongoSituationRepository) CreateRule(ctx context.Context, rule *model.SituationRule) error {
	_, err := r.ruleCollection.InsertOne(ctx, rule)
	return err
}

func (r *MongoSituationRepository) UpdateRule(ctx context.Context, id string, rule *model.SituationRule) error {
	_, err := r.ruleCollection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": rule})
	return err
}

func (r *MongoSituationRepository) DeleteRule(ctx context.Context, id string) error {
	_, err := r.ruleCollection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// --- 攻击链 ---

func (r *MongoSituationRepository) ListChains(ctx context.Context, filter bson.M, skip, limit int64) ([]model.AttackChain, int64, error) {
	total, err := r.chainCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	opts := options.Find().SetSkip(skip).SetLimit(limit).SetSort(bson.M{"last_seen": -1})
	cursor, err := r.chainCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)
	var chains []model.AttackChain
	if err := cursor.All(ctx, &chains); err != nil {
		return nil, 0, err
	}
	return chains, total, nil
}

func (r *MongoSituationRepository) GetChainByID(ctx context.Context, id string) (*model.AttackChain, error) {
	var chain model.AttackChain
	err := r.chainCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&chain)
	if err != nil {
		return nil, err
	}
	return &chain, nil
}

func (r *MongoSituationRepository) UpsertChain(ctx context.Context, chain *model.AttackChain) error {
	filter := bson.M{"source_ip": chain.SourceIP}
	update := bson.M{"$set": chain}
	opts := options.UpdateOne().SetUpsert(true)
	_, err := r.chainCollection.UpdateOne(ctx, filter, update, opts)
	return err
}

func (r *MongoSituationRepository) GetChainByIP(ctx context.Context, ip string) (*model.AttackChain, error) {
	var chain model.AttackChain
	err := r.chainCollection.FindOne(ctx, bson.M{"source_ip": ip}).Decode(&chain)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &chain, nil
}

// --- 攻击者画像 ---

func (r *MongoSituationRepository) UpsertProfile(ctx context.Context, profile *model.AttackerProfile) error {
	filter := bson.M{"source_ip": profile.SourceIP}
	update := bson.M{"$set": profile}
	opts := options.UpdateOne().SetUpsert(true)
	_, err := r.profileCollection.UpdateOne(ctx, filter, update, opts)
	return err
}

func (r *MongoSituationRepository) GetProfile(ctx context.Context, ip string) (*model.AttackerProfile, error) {
	var profile model.AttackerProfile
	err := r.profileCollection.FindOne(ctx, bson.M{"source_ip": ip}).Decode(&profile)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *MongoSituationRepository) ListProfiles(ctx context.Context, sortBy string, skip, limit int64) ([]model.AttackerProfile, int64, error) {
	total, err := r.profileCollection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, 0, err
	}
	sortField := "risk_score"
	if sortBy != "" {
		sortField = sortBy
	}
	opts := options.Find().SetSkip(skip).SetLimit(limit).SetSort(bson.M{sortField: -1})
	cursor, err := r.profileCollection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)
	var profiles []model.AttackerProfile
	if err := cursor.All(ctx, &profiles); err != nil {
		return nil, 0, err
	}
	return profiles, total, nil
}
