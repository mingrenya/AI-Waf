package repository

import (
	"context"
	"time"
	
	pkgModel "github.com/mingrenya/AI-Waf/pkg/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MCPRepository struct {
	db *mongo.Database
}

func NewMCPRepository(db *mongo.Database) *MCPRepository {
	return &MCPRepository{db: db}
}

// GetLastToolCall 获取最后一次工具调用记录
func (r *MCPRepository) GetLastToolCall(ctx context.Context) (*pkgModel.MCPToolCall, error) {
	var call pkgModel.MCPToolCall
	collection := r.db.Collection(call.GetCollectionName())

	opts := options.FindOne().SetSort(bson.D{{Key: "timestamp", Value: -1}})
	err := collection.FindOne(ctx, bson.M{}, opts).Decode(&call)
	if err != nil {
		return nil, err
	}

	return &call, nil
}

// GetToolCallHistory 获取工具调用历史
func (r *MCPRepository) GetToolCallHistory(ctx context.Context, limit, offset int) ([]pkgModel.MCPToolCall, int64, error) {
	var call pkgModel.MCPToolCall
	collection := r.db.Collection(call.GetCollectionName())

	// 计数
	total, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, 0, err
	}

	// 查询
	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var calls []pkgModel.MCPToolCall
	if err := cursor.All(ctx, &calls); err != nil {
		return nil, 0, err
	}

	return calls, total, nil
}

// RecordToolCall 记录工具调用
func (r *MCPRepository) RecordToolCall(ctx context.Context, toolName string, duration int64, success bool, errorMsg string) error {
	var callModel pkgModel.MCPToolCall
	collection := r.db.Collection(callModel.GetCollectionName())

	call := pkgModel.MCPToolCall{
		ToolName:  toolName,
		Timestamp: time.Now(),
		Duration:  duration,
		Success:   success,
		Error:     errorMsg,
	}

	_, err := collection.InsertOne(ctx, call)
	return err
}
