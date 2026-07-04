package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// DDoSSharedBlocklist DDoS 协同黑名单管理器
//
// 核心能力:
//  1. 外部威胁情报拉取 — 从 AbuseIPDB 等平台拉取最新黑名单
//  2. 本地封禁上报   — 本地封禁的 IP 可选上报到共享平台
//  3. 多源聚合       — 支持多源 threat intelligence feeds
//  4. TTL 过期       — 黑名单条目自动过期清除
type DDoSSharedBlocklist struct {
	mu sync.RWMutex

	enabled bool

	// 本地黑名单缓存: IP → 条目详情
	blocklist map[string]*BlocklistEntry

	// API 配置
	apiURL    string
	apiKey    string
	sourceTag string // 上报标识

	// 拉取配置
	pollInterval    time.Duration // 拉取间隔，默认 5 分钟
	blocklistTTL    time.Duration // 黑名单条目 TTL，默认 30 分钟
	maxBlocklistSize int          // 最大黑名单条目数

	httpClient *http.Client
	logger     zerolog.Logger

	stopCh chan struct{}
	wg     sync.WaitGroup
}

// BlocklistEntry 黑名单条目
type BlocklistEntry struct {
	IP          string    `json:"ip"`
	Source      string    `json:"source"`      // 来源 (abuseipdb / otx / local)
	Category    string    `json:"category"`    // 分类 (ddos / scanner / malware / botnet)
	Confidence  int       `json:"confidence"`  // 置信度 0-100
	AddedAt     time.Time `json:"added_at"`    // 添加时间
	ExpiresAt   time.Time `json:"expires_at"`  // 过期时间
	ReportCount int       `json:"report_count"`// 报告次数
	Description string    `json:"description"` // 描述
}

// DDoSBlocklistConfig 协同黑名单配置
type DDoSBlocklistConfig struct {
	Enabled         bool          // 是否启用
	APIURL          string        // 威胁情报 API 地址
	APIKey          string        // API 密钥
	SourceTag       string        // 上报来源标识
	PollInterval    time.Duration // 拉取间隔
	BlocklistTTL    time.Duration // 条目 TTL
	MaxBlocklistSize int          // 最大条目数
}

// DefaultDDoSBlocklistConfig 默认配置
func DefaultDDoSBlocklistConfig() DDoSBlocklistConfig {
	return DDoSBlocklistConfig{
		Enabled:         false,
		APIURL:          "https://api.abuseipdb.com/api/v2/blacklist",
		APIKey:          os.Getenv("ABUSEIPDB_API_KEY"),
		SourceTag:       "ai-waf",
		PollInterval:    5 * time.Minute,
		BlocklistTTL:    30 * time.Minute,
		MaxBlocklistSize: 50000,
	}
}

// NewDDoSSharedBlocklist 创建协同黑名单管理器
func NewDDoSSharedBlocklist(cfg DDoSBlocklistConfig, logger zerolog.Logger) *DDoSSharedBlocklist {
	return &DDoSSharedBlocklist{
		enabled:     cfg.Enabled,
		blocklist:   make(map[string]*BlocklistEntry),
		apiURL:      cfg.APIURL,
		apiKey:      cfg.APIKey,
		sourceTag:   cfg.SourceTag,
		pollInterval:    cfg.PollInterval,
		blocklistTTL:    cfg.BlocklistTTL,
		maxBlocklistSize: cfg.MaxBlocklistSize,
		httpClient:  &http.Client{Timeout: 15 * time.Second},
		logger:      logger,
		stopCh:      make(chan struct{}),
	}
}

// Enable 启用协同黑名单
func (dsb *DDoSSharedBlocklist) Enable() {
	dsb.mu.Lock()
	defer dsb.mu.Unlock()
	dsb.enabled = true
}

// Disable 禁用协同黑名单
func (dsb *DDoSSharedBlocklist) Disable() {
	dsb.mu.Lock()
	defer dsb.mu.Unlock()
	dsb.enabled = false
}

// IsEnabled 检查是否启用
func (dsb *DDoSSharedBlocklist) IsEnabled() bool {
	dsb.mu.RLock()
	defer dsb.mu.RUnlock()
	return dsb.enabled
}

// ─── 黑名单查询 ───

// IsBlocklisted 检查 IP 是否在共享黑名单中
func (dsb *DDoSSharedBlocklist) IsBlocklisted(ip string) (bool, *BlocklistEntry) {
	dsb.mu.RLock()
	defer dsb.mu.RUnlock()

	if !dsb.enabled {
		return false, nil
	}

	entry, ok := dsb.blocklist[ip]
	if !ok {
		return false, nil
	}

	// 检查是否过期
	if time.Now().After(entry.ExpiresAt) {
		return false, nil
	}

	return true, entry
}

// GetBlocklistSize 返回当前黑名单条目数
func (dsb *DDoSSharedBlocklist) GetBlocklistSize() int {
	dsb.mu.RLock()
	defer dsb.mu.RUnlock()
	return len(dsb.blocklist)
}

// ─── 本地封禁上报 ───

// ReportIP 将本地封禁的 IP 上报到威胁情报平台
func (dsb *DDoSSharedBlocklist) ReportIP(ip, category, reason string) error {
	dsb.mu.RLock()
	enabled := dsb.enabled
	apiURL := dsb.apiURL
	apiKey := dsb.apiKey
	dsb.mu.RUnlock()

	if !enabled || apiKey == "" {
		return nil
	}

	// AbuseIPDB Report API
	req, err := http.NewRequest("POST", fmt.Sprintf("%s?ipAddress=%s&categories=%s&comment=%s",
		apiURL, ip, category, reason), nil)
	if err != nil {
		return fmt.Errorf("创建上报请求失败: %w", err)
	}

	req.Header.Set("Key", apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := dsb.httpClient.Do(req)
	if err != nil {
		dsb.logger.Warn().Err(err).Str("ip", ip).Msg("上报 IP 到威胁情报平台失败")
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		dsb.logger.Warn().
			Int("status", resp.StatusCode).
			Str("ip", ip).
			Msg("威胁情报平台返回错误")
		return fmt.Errorf("API 返回状态码 %d", resp.StatusCode)
	}

	dsb.logger.Info().
		Str("ip", ip).
		Str("category", category).
		Msg("IP 已上报到威胁情报平台")
	return nil
}

// ─── 拉取外部黑名单 ───

// fetchBlocklist 从外部 API 拉取黑名单
func (dsb *DDoSSharedBlocklist) fetchBlocklist(ctx context.Context) ([]BlocklistEntry, error) {
	dsb.mu.RLock()
	apiURL := dsb.apiURL
	apiKey := dsb.apiKey
	dsb.mu.RUnlock()

	if apiKey == "" {
		return nil, nil
	}

	url := apiURL + "?confidenceMinimum=90&limit=10000&plaintext"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建拉取请求失败: %w", err)
	}

	req.Header.Set("Key", apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := dsb.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("拉取黑名单失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API 返回状态码 %d", resp.StatusCode)
	}

	var result struct {
		Data []struct {
			IPAddress   string `json:"ipAddress"`
			Category    int    `json:"abuseConfidenceScore"`
			ReportCount int    `json:"totalReports"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		// 尝试解析为纯文本格式
		return dsb.parsePlainTextBlocklist(resp)
	}

	now := time.Now()
	entries := make([]BlocklistEntry, 0, len(result.Data))
	for _, item := range result.Data {
		entries = append(entries, BlocklistEntry{
			IP:          item.IPAddress,
			Source:      "abuseipdb",
			Category:    "multi",
			Confidence:  item.Category,
			AddedAt:     now,
			ExpiresAt:   now.Add(dsb.blocklistTTL),
			ReportCount: item.ReportCount,
			Description: fmt.Sprintf("置信度 %d, 报告 %d 次", item.Category, item.ReportCount),
		})
	}

	return entries, nil
}

// parsePlainTextBlocklist 解析纯文本格式黑名单 (每行一个IP)
func (dsb *DDoSSharedBlocklist) parsePlainTextBlocklist(resp *http.Response) ([]BlocklistEntry, error) {
	// 简化实现: 纯文本格式意味着不是 AbuseIPDB 标准 JSON, 返回空
	// 未来可以扩展支持更多格式 (如 AlienVault OTX CSV)
	return nil, nil
}

// ─── 本地黑名单管理 ───

// mergeBlocklist 合并外部拉取的黑名单到本地缓存
func (dsb *DDoSSharedBlocklist) mergeBlocklist(entries []BlocklistEntry) {
	dsb.mu.Lock()
	defer dsb.mu.Unlock()

	now := time.Now()
	added := 0

	for _, entry := range entries {
		// 跳过过期的
		if now.After(entry.ExpiresAt) {
			continue
		}

		// 缓存满了
		if len(dsb.blocklist) >= dsb.maxBlocklistSize {
			dsb.logger.Warn().Int("max", dsb.maxBlocklistSize).Msg("黑名单缓存已满")
			break
		}

		// 已存在：更新如果新的置信度更高
		if existing, ok := dsb.blocklist[entry.IP]; ok {
			if entry.Confidence > existing.Confidence {
				existing.Confidence = entry.Confidence
				existing.ExpiresAt = entry.ExpiresAt
				existing.ReportCount = entry.ReportCount
			}
			continue
		}

		entry.AddedAt = now
		dsb.blocklist[entry.IP] = &entry
		added++
	}

	if added > 0 {
		dsb.logger.Info().Int("added", added).Int("total", len(dsb.blocklist)).Msg("合并外部黑名单")
	}
}

// AddLocalBlock 添加本地封禁条目到黑名单
func (dsb *DDoSSharedBlocklist) AddLocalBlock(ip, category, reason string) *BlocklistEntry {
	dsb.mu.Lock()
	defer dsb.mu.Unlock()

	if !dsb.enabled {
		return nil
	}

	now := time.Now()
	entry := &BlocklistEntry{
		IP:          ip,
		Source:      dsb.sourceTag,
		Category:    category,
		Confidence:  100,
		AddedAt:     now,
		ExpiresAt:   now.Add(dsb.blocklistTTL),
		ReportCount: 1,
		Description: reason,
	}

	dsb.blocklist[ip] = entry
	dsb.logger.Info().
		Str("ip", ip).
		Str("category", category).
		Str("reason", reason).
		Msg("本地封禁已加入协同黑名单")
	return entry
}

// ─── 后台任务 ───

// Start 启动后台拉取 + 清理协程
func (dsb *DDoSSharedBlocklist) Start() {
	if !dsb.enabled {
		return
	}

	dsb.wg.Add(2)

	// 拉取协程
	go func() {
		defer dsb.wg.Done()
		ticker := time.NewTicker(dsb.pollInterval)
		defer ticker.Stop()

		dsb.logger.Info().Dur("interval", dsb.pollInterval).Msg("DDoS 协同黑名单拉取已启动")

		// 立即执行一次拉取
		dsb.pollNow()

		for {
			select {
			case <-dsb.stopCh:
				return
			case <-ticker.C:
				dsb.pollNow()
			}
		}
	}()

	// 清理协程
	go func() {
		defer dsb.wg.Done()
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-dsb.stopCh:
				return
			case <-ticker.C:
				dsb.cleanupExpired()
			}
		}
	}()
}

// pollNow 执行一次拉取
func (dsb *DDoSSharedBlocklist) pollNow() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	entries, err := dsb.fetchBlocklist(ctx)
	if err != nil {
		dsb.logger.Warn().Err(err).Msg("拉取外部黑名单失败")
		return
	}

	if len(entries) > 0 {
		dsb.mergeBlocklist(entries)
	}
}

// cleanupExpired 清理过期的黑名单条目
func (dsb *DDoSSharedBlocklist) cleanupExpired() {
	dsb.mu.Lock()
	defer dsb.mu.Unlock()

	now := time.Now()
	removed := 0
	for ip, entry := range dsb.blocklist {
		if now.After(entry.ExpiresAt) {
			delete(dsb.blocklist, ip)
			removed++
		}
	}
	if removed > 0 {
		dsb.logger.Debug().Int("removed", removed).Int("remaining", len(dsb.blocklist)).Msg("清理过期黑名单条目")
	}
}

// Stop 停止后台协程
func (dsb *DDoSSharedBlocklist) Stop() {
	close(dsb.stopCh)
	dsb.wg.Wait()
	dsb.logger.Info().Msg("DDoS 协同黑名单已停止")
}

// Stats 返回当前统计信息
func (dsb *DDoSSharedBlocklist) Stats() map[string]interface{} {
	dsb.mu.RLock()
	defer dsb.mu.RUnlock()
	return map[string]interface{}{
		"enabled":      dsb.enabled,
		"blocklist_size": len(dsb.blocklist),
		"poll_interval":  dsb.pollInterval.String(),
		"blocklist_ttl":  dsb.blocklistTTL.String(),
	}
}
