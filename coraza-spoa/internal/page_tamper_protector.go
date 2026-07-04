package internal

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// PageTamperProtector 网页防篡改保护器
//
// 功能:
//  1. 快照采集 — 启动时对受保护的静态页面计算 SHA256 哈希
//  2. 定时巡检 — 周期对比当前内容与快照，检测被篡改的页面
//  3. 响应防护 — 检测到篡改后自动从备份恢复原始内容
//  4. 告警记录 — 篡改事件记录到 MongoDB + 触发攻击回调
type PageTamperProtector struct {
	mu sync.RWMutex

	enabled     bool
	snapshotDir string                     // 快照存储目录
	protectPaths []string                  // 受保护的页面路径前缀
	checkInterval time.Duration            // 巡检间隔
	maxSnapshots  int                      // 最大快照数

	// 内存中的快照缓存: 文件路径 → SHA256 哈希 (hex)
	cache map[string]string

	// 备份存储: 文件路径 → 原始文件内容
	backups map[string][]byte

	mongoClient *mongo.Client
	database    string

	logger zerolog.Logger

	stopCh chan struct{}
	wg     sync.WaitGroup
}

// PageTamperConfig 网页防篡改配置
type PageTamperConfig struct {
	Enabled       bool
	SnapshotDir   string          // 快照/备份存储目录
	ProtectPaths  []string        // 受保护的页面路径前缀
	CheckInterval time.Duration   // 巡检间隔，默认 30 秒
	MaxSnapshots  int             // 最大快照数，默认 10000
}

// DefaultPageTamperConfig 默认配置
func DefaultPageTamperConfig() PageTamperConfig {
	return PageTamperConfig{
		Enabled:       false,
		SnapshotDir:   "/tmp/ai-waf/tamper-snapshots",
		ProtectPaths:  []string{"/", "/index.html"},
		CheckInterval: 30 * time.Second,
		MaxSnapshots:  10000,
	}
}

// NewPageTamperProtector 创建网页防篡改保护器
func NewPageTamperProtector(cfg PageTamperConfig, mongoClient *mongo.Client, database string, logger zerolog.Logger) *PageTamperProtector {
	return &PageTamperProtector{
		enabled:       cfg.Enabled,
		snapshotDir:   cfg.SnapshotDir,
		protectPaths:  cfg.ProtectPaths,
		checkInterval: cfg.CheckInterval,
		maxSnapshots:  cfg.MaxSnapshots,
		cache:         make(map[string]string),
		backups:       make(map[string][]byte),
		mongoClient:   mongoClient,
		database:      database,
		logger:        logger,
		stopCh:        make(chan struct{}),
	}
}

// Enable 启用篡改保护
func (ptp *PageTamperProtector) Enable() {
	ptp.mu.Lock()
	defer ptp.mu.Unlock()
	ptp.enabled = true
}

// Disable 禁用篡改保护
func (ptp *PageTamperProtector) Disable() {
	ptp.mu.Lock()
	defer ptp.mu.Unlock()
	ptp.enabled = false
}

// IsEnabled 检查是否启用
func (ptp *PageTamperProtector) IsEnabled() bool {
	ptp.mu.RLock()
	defer ptp.mu.RUnlock()
	return ptp.enabled
}

// TakeSnapshot 对指定的文件内容计算快照并缓存
// 返回: SHA256 + 是否为新快照
func (ptp *PageTamperProtector) TakeSnapshot(path string, content []byte) (string, bool) {
	if !ptp.enabled {
		return "", false
	}

	hash := sha256Hex(content)

	ptp.mu.Lock()
	defer ptp.mu.Unlock()

	if len(ptp.cache) >= ptp.maxSnapshots {
		ptp.logger.Warn().Int("max", ptp.maxSnapshots).Msg("快照缓存已满")
		return hash, false
	}

	oldHash, exists := ptp.cache[path]
	if exists && oldHash == hash {
		return hash, false // 未变化
	}

	// 首次快照：缓存 + 备份
	ptp.cache[path] = hash
	ptp.backups[path] = append([]byte(nil), content...)

	// 保存到磁盘
	ptp.saveSnapshotToDisk(path, hash, content)

	return hash, !exists // 首次快照返回 true
}

// CheckPage 检查页面是否被篡改
// 返回: 是否被篡改 + 原始内容 + 当前哈希 + 快照哈希
func (ptp *PageTamperProtector) CheckPage(path string, currentContent []byte) (tampered bool, originalContent []byte, currentHash, snapshotHash string) {
	ptp.mu.RLock()
	defer ptp.mu.RUnlock()

	currentHash = sha256Hex(currentContent)

	snapshotHash, hasSnapshot := ptp.cache[path]
	if !hasSnapshot {
		return false, nil, currentHash, ""
	}

	if currentHash != snapshotHash {
		originalContent = ptp.backups[path]
		return true, originalContent, currentHash, snapshotHash
	}

	return false, nil, currentHash, snapshotHash
}

// GetBackup 获取页面的备份内容（用于恢复）
func (ptp *PageTamperProtector) GetBackup(path string) []byte {
	ptp.mu.RLock()
	defer ptp.mu.RUnlock()
	return ptp.backups[path]
}

// ─── 快照持久化 ───

// saveSnapshotToDisk 保存快照到磁盘文件
func (ptp *PageTamperProtector) saveSnapshotToDisk(path, hash string, content []byte) {
	if err := os.MkdirAll(ptp.snapshotDir, 0755); err != nil {
		ptp.logger.Error().Err(err).Msg("创建快照目录失败")
		return
	}

	// 文件名：sha256_前8位 + 原始路径的 hash
	safeName := hash[:8] + "_" + hex.EncodeToString([]byte(path))[:16]
	snapshotFile := filepath.Join(ptp.snapshotDir, safeName)

	if err := os.WriteFile(snapshotFile, content, 0644); err != nil {
		ptp.logger.Error().Err(err).Str("file", snapshotFile).Msg("保存快照文件失败")
		return
	}
}

// ─── 定时巡检 ───

// StartPatrol 启动定时巡检 goroutine
func (ptp *PageTamperProtector) StartPatrol(onTamper func(path string, originalContent, currentContent []byte)) {
	ptp.wg.Add(1)
	go func() {
		defer ptp.wg.Done()
		ticker := time.NewTicker(ptp.checkInterval)
		defer ticker.Stop()

		ptp.logger.Info().Dur("interval", ptp.checkInterval).Msg("网页防篡改巡检已启动")

		for {
			select {
			case <-ptp.stopCh:
				return
			case <-ticker.C:
				ptp.patrolCheck(onTamper)
			}
		}
	}()
}

// patrolCheck 执行一次巡检
func (ptp *PageTamperProtector) patrolCheck(onTamper func(path string, originalContent, currentContent []byte)) {
	ptp.mu.RLock()
	paths := make([]string, 0, len(ptp.cache))
	for p := range ptp.cache {
		paths = append(paths, p)
	}
	ptp.mu.RUnlock()

	for _, path := range paths {
		// 读取当前文件内容
		currentContent, err := os.ReadFile(path)
		if err != nil {
			// 文件可能已被删除（也是篡改的一种形式）
			ptp.logger.Warn().Err(err).Str("path", path).Msg("巡检无法读取页面文件（可能被删除）")
			if onTamper != nil {
				ptp.mu.RLock()
				original := ptp.backups[path]
				ptp.mu.RUnlock()
				if original != nil {
					onTamper(path, original, nil)
				}
			}
			continue
		}

		tampered, original, currentHash, snapshotHash := ptp.CheckPage(path, currentContent)
		if tampered {
			ptp.logger.Warn().
				Str("path", path).
				Str("current_hash", currentHash).
				Str("snapshot_hash", snapshotHash).
				Msg("检测到网页被篡改")

			// 记录到 MongoDB
			ptp.logTamperEvent(path, currentHash, snapshotHash)

			if onTamper != nil && original != nil {
				onTamper(path, original, currentContent)
			}
		}
	}
}

// ─── MongoDB 事件记录 ───

// TamperEvent 篡改事件
type TamperEvent struct {
	Path         string    `bson:"path" json:"path"`
	CurrentHash  string    `bson:"current_hash" json:"current_hash"`
	SnapshotHash string    `bson:"snapshot_hash" json:"snapshot_hash"`
	DetectedAt   time.Time `bson:"detected_at" json:"detected_at"`
	Restored     bool      `bson:"restored" json:"restored"`
}

func (ptp *PageTamperProtector) logTamperEvent(path, currentHash, snapshotHash string) {
	if ptp.mongoClient == nil {
		return
	}

	collection := ptp.mongoClient.Database(ptp.database).Collection("page_tamper_events")
	event := TamperEvent{
		Path:         path,
		CurrentHash:  currentHash,
		SnapshotHash: snapshotHash,
		DetectedAt:   time.Now(),
		Restored:     false,
	}

	// 使用context.Background()因为这是后台goroutine
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, event)
	if err != nil {
		ptp.logger.Error().Err(err).Str("path", path).Msg("记录篡改事件失败")
	}
}

// ─── Stop ───

// Stop 停止巡检
func (ptp *PageTamperProtector) Stop() {
	close(ptp.stopCh)
	ptp.wg.Wait()
	ptp.logger.Info().Msg("网页防篡改巡检已停止")
}

// Stats 返回当前统计信息
func (ptp *PageTamperProtector) Stats() map[string]interface{} {
	ptp.mu.RLock()
	defer ptp.mu.RUnlock()
	return map[string]interface{}{
		"enabled":        ptp.enabled,
		"snapshot_count": len(ptp.cache),
		"backup_count":   len(ptp.backups),
		"snapshot_dir":   ptp.snapshotDir,
		"check_interval": ptp.checkInterval.String(),
	}
}

// ─── 工具函数 ───

// sha256Hex 计算 SHA256 并返回 hex 字符串
func sha256Hex(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// AddProtectPath 动态添加受保护路径
func (ptp *PageTamperProtector) AddProtectPath(path string, content []byte) {
	ptp.TakeSnapshot(path, content)
}

// RemoveProtectPath 移除受保护路径
func (ptp *PageTamperProtector) RemoveProtectPath(path string) {
	ptp.mu.Lock()
	defer ptp.mu.Unlock()
	delete(ptp.cache, path)
	delete(ptp.backups, path)
}

// CheckResponseBody 检查响应体是否匹配快照
// 应用于 HandleResponse 阶段，检查 Web 服务器返回的内容是否被篡改
func (ptp *PageTamperProtector) CheckResponseBody(path string, responseBody []byte) (tampered bool, originalBody []byte) {
	if !ptp.enabled {
		return false, nil
	}

	// 快速过滤：只检查受保护的路径
	protected := false
	ptp.mu.RLock()
	for _, prefix := range ptp.protectPaths {
		if strings.HasPrefix(path, prefix) || strings.HasSuffix(path, ".html") || strings.HasSuffix(path, ".htm") {
			protected = true
			break
		}
	}
	ptp.mu.RUnlock()

	if !protected {
		return false, nil
	}

	tampered, original, _, _ := ptp.CheckPage(path, responseBody)
	if tampered {
		// 自动恢复：用快照内容替换被篡改的内容
		if original != nil {
			return true, original
		}
	}
	return false, nil
}
