package internal

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// APIDiscovery API 资产自动发现引擎
//
// 从 WAF 流量日志中自动推断 API 端点:
//  1. 流量采样 — 收集 HTTP method + path 组合
//  2. 路径模式聚合 — 将 /users/123 → /users/{id} 这种路径参数化
//  3. OpenAPI 推断 — 生成 OpenAPI 3.0 兼容的路径定义
//  4. 未授权检测 — 标记缺少认证 Header 的敏感端点
//  5. 资产存储 — 写入 MongoDB (api_assets 集合)
type APIDiscovery struct {
	mu sync.RWMutex

	enabled bool

	// 内存中的端点聚合: method:path → 统计
	endpoints map[string]*EndpointStats

	// 路径参数化用的正则
	pathParamPatterns []*regexp.Regexp

	// 敏感路径检测
	sensitivePaths []string

	// 认证 Header 名称
	authHeaders []string

	// 配置
	flushInterval   time.Duration // 刷新到 MongoDB 的间隔
	maxEndpointCache int          // 内存中最大端点缓存数

	mongoClient *mongo.Client
	database    string

	logger zerolog.Logger

	stopCh chan struct{}
	wg     sync.WaitGroup
}

// EndpointStats 端点统计
type EndpointStats struct {
	Method      string    `bson:"method" json:"method"`
	Path        string    `bson:"path" json:"path"`
	Normalized  string    `bson:"normalized" json:"normalized"` // 参数化后的路径
	HitCount    int64     `bson:"hit_count" json:"hit_count"`
	FirstSeen   time.Time `bson:"first_seen" json:"first_seen"`
	LastSeen    time.Time `bson:"last_seen" json:"last_seen"`
	AvgResponseTime float64 `bson:"avg_response_time" json:"avg_response_time"`
	Responses   map[int]int64 `bson:"responses" json:"responses"` // status → count
	HasAuth     bool      `bson:"has_auth" json:"has_auth"`
	IsSensitive bool      `bson:"is_sensitive" json:"is_sensitive"`
	Deprecated  bool      `bson:"deprecated" json:"deprecated"`
}

// APIEndpointRecord MongoDB 中的 API 资产记录
type APIEndpointRecord struct {
	ID         string     `bson:"_id" json:"id"`
	Method     string     `bson:"method" json:"method"`
	Path       string     `bson:"path" json:"path"`
	Normalized string     `bson:"normalized" json:"normalized"`
	HitCount   int64      `bson:"hit_count" json:"hit_count"`
	FirstSeen  time.Time  `bson:"first_seen" json:"first_seen"`
	LastSeen   time.Time  `bson:"last_seen" json:"last_seen"`
	HasAuth    bool       `bson:"has_auth" json:"has_auth"`
	IsSensitive bool      `bson:"is_sensitive" json:"is_sensitive"`
	Deprecated bool       `bson:"deprecated" json:"deprecated"`
	UpdatedAt  time.Time  `bson:"updated_at" json:"updated_at"`
}

// APIDiscoveryConfig API 发现配置
type APIDiscoveryConfig struct {
	Enabled          bool
	FlushInterval    time.Duration // 刷新到 MongoDB 间隔
	MaxEndpointCache int           // 最大端点缓存数
	SensitivePaths   []string      // 敏感路径前缀
	AuthHeaders      []string      // 认证 Header 名称
}

// DefaultAPIDiscoveryConfig 默认配置
func DefaultAPIDiscoveryConfig() APIDiscoveryConfig {
	return APIDiscoveryConfig{
		Enabled:          false,
		FlushInterval:    60 * time.Second,
		MaxEndpointCache: 10000,
		SensitivePaths: []string{
			"/api/admin", "/admin", "/manage", "/dashboard",
			"/swagger", "/actuator", "/metrics", "/health",
		},
		AuthHeaders: []string{"authorization", "x-api-key", "x-auth-token", "cookie"},
	}
}

// NewAPIDiscovery 创建 API 资产发现引擎
func NewAPIDiscovery(cfg APIDiscoveryConfig, mongoClient *mongo.Client, database string, logger zerolog.Logger) *APIDiscovery {
	ad := &APIDiscovery{
		enabled:        cfg.Enabled,
		endpoints:      make(map[string]*EndpointStats),
		sensitivePaths: cfg.SensitivePaths,
		authHeaders:    cfg.AuthHeaders,
		flushInterval:     cfg.FlushInterval,
		maxEndpointCache:  cfg.MaxEndpointCache,
		mongoClient:    mongoClient,
		database:       database,
		logger:         logger,
		stopCh:         make(chan struct{}),
		pathParamPatterns: []*regexp.Regexp{
			regexp.MustCompile(`/\d+`),                       // /users/123
			regexp.MustCompile(`/[0-9a-f]{8,}`),              // UUID / hash
			regexp.MustCompile(`/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`), // UUID v4
			regexp.MustCompile(`/\d{4}-\d{2}-\d{2}`),         // dates
		},
	}
	return ad
}

// Enable 启用 API 发现
func (ad *APIDiscovery) Enable() {
	ad.mu.Lock()
	defer ad.mu.Unlock()
	ad.enabled = true
}

// Disable 禁用 API 发现
func (ad *APIDiscovery) Disable() {
	ad.mu.Lock()
	defer ad.mu.Unlock()
	ad.enabled = false
}

// IsEnabled 检查是否启用
func (ad *APIDiscovery) IsEnabled() bool {
	ad.mu.RLock()
	defer ad.mu.RUnlock()
	return ad.enabled
}

// ─── 流量记录 ───

// RecordRequest 记录一个请求，用于 API 端点学习
func (ad *APIDiscovery) RecordRequest(method, path, query string, hasAuth bool, statusCode int, responseTime float64) {
	ad.mu.Lock()
	defer ad.mu.Unlock()

	if !ad.enabled {
		return
	}

	// 清理路径: 去除 query string, 规范化为小写
	cleanPath := normalizePath(path, query)

	key := method + ":" + cleanPath
	ep, exists := ad.endpoints[key]
	if !exists {
		if len(ad.endpoints) >= ad.maxEndpointCache {
			return // 缓存已满
		}
		ep = &EndpointStats{
			Method:    method,
			Path:      cleanPath,
			Normalized: normalizeParams(cleanPath, ad.pathParamPatterns),
			FirstSeen: time.Now(),
			Responses: make(map[int]int64),
		}
		ad.endpoints[key] = ep
	}

	ep.HitCount++
	ep.LastSeen = time.Now()
	ep.Responses[statusCode]++

	// 更新平均响应时间 (指数移动平均)
	if ep.AvgResponseTime == 0 {
		ep.AvgResponseTime = responseTime
	} else {
		ep.AvgResponseTime = ep.AvgResponseTime*0.9 + responseTime*0.1
	}

	// 更新认证状态 (任一请求有 auth 则标记)
	if hasAuth {
		ep.HasAuth = true
	}

	// 检查是否为敏感路径
	if !ep.IsSensitive {
		for _, prefix := range ad.sensitivePaths {
			if strings.HasPrefix(cleanPath, prefix) {
				ep.IsSensitive = true
				break
			}
		}
	}
}

// ─── 路径参数化 ───

// normalizePath 清理和规范化路径
func normalizePath(path, query string) string {
	// URL decode
	if decoded, err := url.QueryUnescape(path); err == nil {
		path = decoded
	}

	// 去除 trailing slash (但保留 root "/")
	if len(path) > 1 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}

	// 小写 (路径通常是大小写敏感的，但为了聚合可以考虑)
	// path = strings.ToLower(path)

	return path
}

// normalizeParams 将路径中的动态参数替换为 {param}
func normalizeParams(path string, patterns []*regexp.Regexp) string {
	result := path
	for _, pat := range patterns {
		result = pat.ReplaceAllString(result, "/{param}")
	}
	return result
}

// ─── MongoDB 持久化 ───

// flushToMongo 将内存中的端点统计刷新到 MongoDB
func (ad *APIDiscovery) flushToMongo(ctx context.Context) error {
	ad.mu.RLock()
	endpoints := make([]*EndpointStats, 0, len(ad.endpoints))
	for _, ep := range ad.endpoints {
		epCopy := *ep
		endpoints = append(endpoints, &epCopy)
	}
	ad.mu.RUnlock()

	if ad.mongoClient == nil {
		return nil
	}

	collection := ad.mongoClient.Database(ad.database).Collection("api_assets")

	now := time.Now()
	upserted := 0

	for _, ep := range endpoints {
		id := ep.Method + ":" + ep.Path
		filter := bson.M{"_id": id}

		update := bson.M{
			"$set": bson.M{
				"method":       ep.Method,
				"path":         ep.Path,
				"normalized":   ep.Normalized,
				"hit_count":    ep.HitCount,
				"first_seen":   ep.FirstSeen,
				"last_seen":    ep.LastSeen,
				"has_auth":     ep.HasAuth,
				"is_sensitive": ep.IsSensitive,
				"updated_at":   now,
			},
			"$setOnInsert": bson.M{
				"deprecated": false,
			},
		}

		opts := options.UpdateOne().SetUpsert(true)
		_, err := collection.UpdateOne(ctx, filter, update, opts)
		if err != nil {
			ad.logger.Warn().Err(err).Str("endpoint", id).Msg("Upsert API 资产失败")
			continue
		}
		upserted++
	}

	// 标记长期未见过的端点为 deprecated
	deprecated := ad.markDeprecated(ctx, collection)
	if upserted > 0 || deprecated > 0 {
		ad.logger.Info().
			Int("upserted", upserted).
			Int("deprecated", deprecated).
			Msg("API 资产已同步到 MongoDB")
	}

	return nil
}

// markDeprecated 标记 7 天未出现的端点为 deprecated
func (ad *APIDiscovery) markDeprecated(ctx context.Context, collection *mongo.Collection) int {
	cutoff := time.Now().Add(-7 * 24 * time.Hour)
	filter := bson.M{
		"last_seen":  bson.M{"lt": cutoff},
		"deprecated": false,
	}
	update := bson.M{
		"$set": bson.M{
			"deprecated": true,
			"updated_at": time.Now(),
		},
	}
	result, err := collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return 0
	}
	return int(result.ModifiedCount)
}

// ─── 查询 ───

// GetAPIEndpoints 从 MongoDB 查询 API 资产列表
func (ad *APIDiscovery) GetAPIEndpoints(ctx context.Context, filter map[string]interface{}, limit int) ([]APIEndpointRecord, error) {
	if ad.mongoClient == nil {
		return nil, fmt.Errorf("MongoDB 未连接")
	}

	collection := ad.mongoClient.Database(ad.database).Collection("api_assets")

	mongoFilter := bson.M{}
	for k, v := range filter {
		mongoFilter[k] = v
	}

	opts := options.Find().SetSort(bson.M{"hit_count": -1})
	if limit > 0 {
		opts = opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, mongoFilter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []APIEndpointRecord
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

// GetUnauthenticatedEndpoints 获取未授权访问的敏感端点列表
func (ad *APIDiscovery) GetUnauthenticatedEndpoints(ctx context.Context) ([]APIEndpointRecord, error) {
	return ad.GetAPIEndpoints(ctx, map[string]interface{}{
		"is_sensitive": true,
		"has_auth":    false,
		"deprecated":   false,
	}, 200)
}

// ─── 统计 ───

// DiscoveredCount 返回已发现的端点数
func (ad *APIDiscovery) DiscoveredCount() int {
	ad.mu.RLock()
	defer ad.mu.RUnlock()
	return len(ad.endpoints)
}

// ─── 后台任务 ───

// Start 启动定期刷新协程
func (ad *APIDiscovery) Start() {
	if !ad.enabled {
		return
	}

	ad.wg.Add(1)
	go func() {
		defer ad.wg.Done()
		ticker := time.NewTicker(ad.flushInterval)
		defer ticker.Stop()

		ad.logger.Info().Dur("interval", ad.flushInterval).Msg("API 资产发现已启动")

		for {
			select {
			case <-ad.stopCh:
				// 停止前最后一次刷新
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				ad.flushToMongo(ctx)
				cancel()
				return
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				ad.flushToMongo(ctx)
				cancel()
			}
		}
	}()
}

// Stop 停止背景任务
func (ad *APIDiscovery) Stop() {
	close(ad.stopCh)
	ad.wg.Wait()
	ad.logger.Info().Msg("API 资产发现已停止")
}

// Stats 返回统计信息
func (ad *APIDiscovery) Stats() map[string]interface{} {
	ad.mu.RLock()
	defer ad.mu.RUnlock()

	unauth := 0
	for _, ep := range ad.endpoints {
		if ep.IsSensitive && !ep.HasAuth {
			unauth++
		}
	}

	return map[string]interface{}{
		"enabled":        ad.enabled,
		"discovered":     len(ad.endpoints),
		"unauthorized":   unauth,
		"flush_interval": ad.flushInterval.String(),
	}
}

// TopEndpoints 返回 Top N 端点
func (ad *APIDiscovery) TopEndpoints(n int) []*EndpointStats {
	ad.mu.RLock()
	defer ad.mu.RUnlock()

	eps := make([]*EndpointStats, 0, len(ad.endpoints))
	for _, ep := range ad.endpoints {
		eps = append(eps, ep)
	}

	sort.Slice(eps, func(i, j int) bool {
		return eps[i].HitCount > eps[j].HitCount
	})

	if len(eps) > n {
		eps = eps[:n]
	}
	return eps
}
