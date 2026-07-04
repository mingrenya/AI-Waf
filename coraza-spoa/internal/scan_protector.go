package internal

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/mingrenya/AI-Waf/pkg/model"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// ScanProtector 扫描防护器
// 统计单IP在时间窗口内触发的攻击规则次数和种类，
// 当达到阈值后自动封禁该IP，实现类似云WAF的"高频扫描封禁"功能。
type ScanProtector struct {
	mu sync.RWMutex

	// IP → 计数器
	counters map[string]*scanCounter

	// 配置
	enabled        bool
	windowSecs     int   // 检测时间窗口（秒）
	triggerCount   int   // 触发规则次数阈值
	triggerTypes   int   // 触发不同规则种类阈值
	blockDuration  time.Duration // 封禁时长
	cleanupEvery   time.Duration // 清理过期计数器的间隔

	// MongoDB 集成
	mongoClient *mongo.Client
	database    string

	logger zerolog.Logger

	stopCh chan struct{}
	wg     sync.WaitGroup
}

// scanCounter 单IP的扫描计数器
type scanCounter struct {
	ip          string
	ruleIDs     map[int]bool // 触发的不同规则ID
	totalHits   int          // 总触发次数
	windowStart time.Time    // 当前窗口开始时间
	blocked     bool         // 是否已被封禁
}

// ScanProtectorConfig 扫描防护配置
type ScanProtectorConfig struct {
	Enabled       bool          // 是否启用
	WindowSecs    int           // 检测时间窗口，默认 60 秒
	TriggerCount  int           // 触发规则次数阈值，默认 20
	TriggerTypes  int           // 触发不同规则种类阈值，默认 2
	BlockDuration time.Duration // 封禁时长，默认 30 分钟
}

// DefaultScanProtectorConfig 默认配置（对标阿里云WAF默认值）
func DefaultScanProtectorConfig() ScanProtectorConfig {
	return ScanProtectorConfig{
		Enabled:       false,
		WindowSecs:    60,
		TriggerCount:  20,
		TriggerTypes:  2,
		BlockDuration: 30 * time.Minute,
	}
}

// NewScanProtector 创建扫描防护器
func NewScanProtector(cfg ScanProtectorConfig, mongoClient *mongo.Client, database string, logger zerolog.Logger) *ScanProtector {
	return &ScanProtector{
		counters:      make(map[string]*scanCounter),
		enabled:       cfg.Enabled,
		windowSecs:    cfg.WindowSecs,
		triggerCount:  cfg.TriggerCount,
		triggerTypes:  cfg.TriggerTypes,
		blockDuration: cfg.BlockDuration,
		cleanupEvery:  5 * time.Minute,
		mongoClient:   mongoClient,
		database:      database,
		logger:        logger,
		stopCh:        make(chan struct{}),
	}
}

// Enable 启用扫描防护
func (sp *ScanProtector) Enable() {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	sp.enabled = true
}

// Disable 禁用扫描防护
func (sp *ScanProtector) Disable() {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	sp.enabled = false
}

// IsEnabled 检查是否启用
func (sp *ScanProtector) IsEnabled() bool {
	sp.mu.RLock()
	defer sp.mu.RUnlock()
	return sp.enabled
}

// RecordHit 记录一次规则命中
// 返回 true 表示该 IP 应被封禁
func (sp *ScanProtector) RecordHit(ip string, ruleID int) bool {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	if !sp.enabled {
		return false
	}

	now := time.Now()
	c, exists := sp.counters[ip]

	if !exists || now.Sub(c.windowStart).Seconds() > float64(sp.windowSecs) {
		// 新窗口
		sp.counters[ip] = &scanCounter{
			ip:          ip,
			ruleIDs:     map[int]bool{ruleID: true},
			totalHits:   1,
			windowStart: now,
		}
		return false
	}

	// 已在封禁状态，跳过
	if c.blocked {
		return false
	}

	c.ruleIDs[ruleID] = true
	c.totalHits++

	// 判断是否达到阈值
	if c.totalHits >= sp.triggerCount && len(c.ruleIDs) >= sp.triggerTypes {
		c.blocked = true
		return true
	}

	return false
}

// BlockIP 封禁指定IP
func (sp *ScanProtector) BlockIP(ctx context.Context, ip string, reason string) error {
	if sp.mongoClient == nil {
		sp.logger.Warn().Str("ip", ip).Msg("无法封禁IP：MongoDB未连接")
		return nil
	}

	record := model.BlockedIPRecord{
		IP:          ip,
		Reason:      reason,
		BlockedAt:   time.Now(),
		BlockedUntil: time.Now().Add(sp.blockDuration),
	}

	// 使用简单的 upsert 逻辑
	collection := sp.mongoClient.Database(sp.database).Collection("blocked_ips")
	filter := bson.M{"ip": ip}
	update := bson.M{"$set": record}

	// 先尝试更新现有记录
	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil || result.MatchedCount == 0 {
		// 记录不存在，插入新记录
		_, err = collection.InsertOne(ctx, record)
	}

	if err != nil {
		sp.logger.Error().Err(err).Str("ip", ip).Msg("封禁IP失败")
		return err
	}

	sp.logger.Info().Str("ip", ip).Str("reason", reason).Dur("duration", sp.blockDuration).Msg("扫描防护：IP已自动封禁")
	return nil
}

// Start 启动后台清理协程
func (sp *ScanProtector) Start() {
	sp.wg.Add(1)
	go func() {
		defer sp.wg.Done()
		ticker := time.NewTicker(sp.cleanupEvery)
		defer ticker.Stop()
		for {
			select {
			case <-sp.stopCh:
				return
			case <-ticker.C:
				sp.cleanup()
			}
		}
	}()
	sp.logger.Info().Msg("扫描防护器已启动")
}

// Stop 停止后台清理协程
func (sp *ScanProtector) Stop() {
	close(sp.stopCh)
	sp.wg.Wait()
	sp.logger.Info().Msg("扫描防护器已停止")
}

// cleanup 清理过期的计数器
func (sp *ScanProtector) cleanup() {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	now := time.Now()
	maxAge := float64(sp.windowSecs * 3) // 保留3个窗口时间
	for ip, c := range sp.counters {
		if now.Sub(c.windowStart).Seconds() > maxAge {
			delete(sp.counters, ip)
		}
	}
}

// Stats 返回当前统计信息
func (sp *ScanProtector) Stats() map[string]interface{} {
	sp.mu.RLock()
	defer sp.mu.RUnlock()

	active := 0
	blocked := 0
	for _, c := range sp.counters {
		if c.blocked {
			blocked++
		} else {
			active++
		}
	}
	return map[string]interface{}{
		"total_tracked": len(sp.counters),
		"active_counters": active,
		"blocked_counters": blocked,
		"enabled": sp.enabled,
	}
}

// ─── 扫描器指纹匹配 ───

// ScannerFingerprint 扫描器指纹规则
type ScannerFingerprint struct {
	Name    string   // 扫描器名称
	Headers map[string]string // 特征 Header: SQLMap → User-Agent 包含 sqlmap
	Paths   []string // 特征路径: Nikto → 请求 /nikto-* 等
}

// knownScanners 已知扫描器指纹库
var knownScanners = []ScannerFingerprint{
	{
		Name: "SQLMap",
		Headers: map[string]string{
			"user-agent": "sqlmap",
		},
	},
	{
		Name: "Nuclei",
		Headers: map[string]string{
			"user-agent": "nuclei",
		},
	},
	{
		Name: "Nikto",
		Headers: map[string]string{
			"user-agent": "nikto",
		},
	},
	{
		Name: "Nessus",
		Headers: map[string]string{
			"user-agent": "nessus",
		},
	},
	{
		Name: "OpenVAS",
		Headers: map[string]string{
			"user-agent": "openvas",
		},
	},
	{
		Name: "Acunetix",
		Headers: map[string]string{
			"user-agent": "acunetix",
		},
	},
	{
		Name: "BurpSuite",
		Headers: map[string]string{
			"user-agent": "burp",
		},
	},
	{
		Name: "Nmap",
		Headers: map[string]string{
			"user-agent": "nmap",
		},
	},
	{
		Name: "curl",
		Headers: map[string]string{
			"user-agent": "curl",
		},
	},
	{
		Name: "wget",
		Headers: map[string]string{
			"user-agent": "wget",
		},
	},
	{
		Name: "AWVS",
		Headers: map[string]string{
			"user-agent": "acunetix",
		},
	},
	{
		Name: "Netsparker",
		Headers: map[string]string{
			"user-agent": "netsparker",
		},
	},
}

// DetectScanner 检测请求是否来自已知扫描器
// 返回扫描器名称，空字符串表示正常请求
func DetectScanner(userAgent string) string {
	if userAgent == "" {
		return ""
	}
	uaLower := strings.ToLower(userAgent)
	for _, s := range knownScanners {
		for key, value := range s.Headers {
			if key == "user-agent" && strings.Contains(uaLower, value) {
				return s.Name
			}
		}
	}
	return ""
}
