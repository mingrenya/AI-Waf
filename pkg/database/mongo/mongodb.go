// mongodb/client.go
package mongodb

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	client     *mongo.Client
	clientOnce sync.Once
	clientErr  error
)

// Connect 根据连接字符串连接到MongoDB，使用推荐的最佳实践
// 返回的客户端是单例，后续调用会返回相同实例
func Connect(uri string) (*mongo.Client, error) {
	// 懒加载模式
	clientOnce.Do(func() {
		// 设置默认选项
		serverAPI := options.ServerAPI(options.ServerAPIVersion1)
		opts := options.Client().
			ApplyURI(uri).
			SetServerAPIOptions(serverAPI).
			SetRetryWrites(true).
			SetRetryReads(true)

		// 创建客户端并连接
		c, err := mongo.Connect(opts)
		if err != nil {
			clientErr = fmt.Errorf("连接MongoDB失败: %w", err)
			return
		}

		// 验证连接
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := c.Database("admin").RunCommand(
			ctx,
			bson.D{{Key: "ping", Value: 1}},
		).Err(); err != nil {
			// 断开连接以清理资源
			_ = c.Disconnect(context.Background())
			clientErr = fmt.Errorf("验证MongoDB连接失败: %w", err)
			return
		}

		client = c
	})

	if clientErr != nil {
		return nil, clientErr
	}

	return client, nil
}

// Disconnect 断开MongoDB连接
func Disconnect() error {
	if client == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.Disconnect(ctx)
	if err != nil {
		return fmt.Errorf("断开MongoDB连接失败: %w", err)
	}

	// 重置单例状态，允许重新连接
	client = nil
	clientOnce = sync.Once{}

	return nil
}

// GetDatabase 获取指定名称的数据库实例
func GetDatabase(dbName string) (*mongo.Database, error) {
	if client == nil {
		return nil, fmt.Errorf("MongoDB未连接，请先调用Connect")
	}
	return client.Database(dbName), nil
}

// GetCollection 获取指定数据库中的集合
func GetCollection(dbName, collName string) (*mongo.Collection, error) {
	if client == nil {
		return nil, fmt.Errorf("MongoDB未连接，请先调用Connect")
	}
	return client.Database(dbName).Collection(collName), nil
}

// Ping 检查MongoDB连接是否活跃
func Ping() error {
	if client == nil {
		return fmt.Errorf("MongoDB未连接，请先调用Connect")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return client.Database("admin").RunCommand(
		ctx,
		bson.D{{Key: "ping", Value: 1}},
	).Err()
}
