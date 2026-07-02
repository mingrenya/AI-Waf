package redis

import (
	"context"
	"os"
	"sync"

	goredis "github.com/redis/go-redis/v9"
)

var (
	client  *goredis.Client
	once    sync.Once
	initErr error
)

// GetClient 获取 Redis 客户端单例。
// REDIS_ADDR 未配置时返回 nil，系统降级运行。
func GetClient(ctx context.Context) (*goredis.Client, error) {
	once.Do(func() {
		addr := os.Getenv("REDIS_ADDR")
		if addr == "" {
			return
		}
		client = goredis.NewClient(&goredis.Options{
			Addr:     addr,
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       0,
		})
		if err := client.Ping(ctx).Err(); err != nil {
			initErr = err
			client.Close()
			client = nil
		}
	})
	return client, initErr
}

// IsAvailable 检查 Redis 是否可用
func IsAvailable(ctx context.Context) bool {
	c, err := GetClient(ctx)
	return err == nil && c != nil
}
