package internal

import (
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// BotCategory 爬虫/自动化工具分类
type BotCategory string

const (
	BotUnknown    BotCategory = "unknown"    // 未知来源
	BotHuman      BotCategory = "human"      // 正常人类用户
	BotSearch     BotCategory = "search"     // 搜索引擎爬虫
	BotSocial     BotCategory = "social"     // 社交媒体爬虫
	BotMonitoring BotCategory = "monitoring" // 监控/健康检查
	BotAutomation BotCategory = "automation" // Headless 浏览器 / 自动化工具
	BotScanner    BotCategory = "scanner"    // 已知漏洞扫描器
	BotMalicious  BotCategory = "malicious"  // 已知恶意爬虫/僵尸网络
)

// BotDecision 处置决策
type BotDecision string

const (
	BotAllow  BotDecision = "allow"  // 放行
	BotTag    BotDecision = "tag"    // 打标签但放行
	BotBlock  BotDecision = "block"  // 拦截
)

// BotDetectionResult Bot 检测结果
type BotDetectionResult struct {
	IP        string       // 客户端 IP
	UserAgent string       // 原始 UA
	Category  BotCategory  // 分类
	BotName   string       // 识别出的 Bot 名称（空：非 Bot）
	Decision  BotDecision  // 处置决策
	Reason    string       // 决策原因
	Score     int          // 综合评分 0~100，越高越像 Bot
	RateRPS   float64      // 当前请求速率（请求/秒）
}

// ─── UA 指纹库 ───

// knownBotFingerprint 已知 Bot 的 UA 特征
type knownBotFingerprint struct {
	Category BotCategory
	Name     string
	Keywords []string // UA 中含任一关键词则匹配
}

// knownBots 已知 Bot 指纹库
// 参考 HTTP Archive + OWASP ModSecurity 维护的 Bot UA 列表
var knownBots = []knownBotFingerprint{
	// ── 搜索引擎（allow）──
	{Category: BotSearch, Name: "Googlebot", Keywords: []string{"googlebot", "google-image", "google-web"}},
	{Category: BotSearch, Name: "Bingbot", Keywords: []string{"bingbot", "msnbot", "bingpreview"}},
	{Category: BotSearch, Name: "Baiduspider", Keywords: []string{"baiduspider"}},
	{Category: BotSearch, Name: "YandexBot", Keywords: []string{"yandexbot", "yandeximages"}},
	{Category: BotSearch, Name: "DuckDuckBot", Keywords: []string{"duckduckbot"}},
	{Category: BotSearch, Name: "Slurp", Keywords: []string{"slurp", "yahoo"}},
	{Category: BotSearch, Name: "Sogou", Keywords: []string{"sogou", "sogou web spider"}},
	{Category: BotSearch, Name: "360Spider", Keywords: []string{"360spider", "hao.360.cn"}},
	{Category: BotSearch, Name: "Bytespider", Keywords: []string{"bytespider", "bytedance"}},

	// ── 社交媒体（allow）──
	{Category: BotSocial, Name: "Facebook", Keywords: []string{"facebookexternalhit", "facebookcatalog"}},
	{Category: BotSocial, Name: "Twitterbot", Keywords: []string{"twitterbot"}},
	{Category: BotSocial, Name: "LinkedInBot", Keywords: []string{"linkedinbot"}},
	{Category: BotSocial, Name: "TelegramBot", Keywords: []string{"telegrambot"}},
	{Category: BotSocial, Name: "WhatsApp", Keywords: []string{"whatsapp"}},
	{Category: BotSocial, Name: "Slackbot", Keywords: []string{"slackbot", "slack-imgproxy"}},
	{Category: BotSocial, Name: "Discordbot", Keywords: []string{"discordbot"}},

	// ── 监控/健康检查（allow）──
	{Category: BotMonitoring, Name: "UptimeRobot", Keywords: []string{"uptimerobot"}},
	{Category: BotMonitoring, Name: "Pingdom", Keywords: []string{"pingdom"}},
	{Category: BotMonitoring, Name: "Datadog", Keywords: []string{"datadog-agent", "datadog-synthetics"}},
	{Category: BotMonitoring, Name: "NewRelic", Keywords: []string{"newrelic", "new relic"}},
	{Category: BotMonitoring, Name: "StatusCake", Keywords: []string{"statuscake"}},
	{Category: BotMonitoring, Name: "AhrefsBot", Keywords: []string{"ahrefsbot"}},
	{Category: BotMonitoring, Name: "SemrushBot", Keywords: []string{"semrushbot"}},

	// ── Headless 浏览器 / 自动化工具（block）──
	{Category: BotAutomation, Name: "HeadlessChrome", Keywords: []string{"headlesschrome", "headless chrome"}},
	{Category: BotAutomation, Name: "Puppeteer", Keywords: []string{"puppeteer"}},
	{Category: BotAutomation, Name: "Playwright", Keywords: []string{"playwright"}},
	{Category: BotAutomation, Name: "Selenium", Keywords: []string{"selenium", "webdriver"}},
	{Category: BotAutomation, Name: "PhantomJS", Keywords: []string{"phantomjs", "phantom.js"}},
	{Category: BotAutomation, Name: "CasperJS", Keywords: []string{"casperjs"}},

	// ── 已知扫描器（block，也被 ScanProtector 覆盖）──
	{Category: BotScanner, Name: "SQLMap", Keywords: []string{"sqlmap"}},
	{Category: BotScanner, Name: "Nuclei", Keywords: []string{"nuclei"}},
	{Category: BotScanner, Name: "Nikto", Keywords: []string{"nikto"}},
	{Category: BotScanner, Name: "Nessus", Keywords: []string{"nessus"}},
	{Category: BotScanner, Name: "OpenVAS", Keywords: []string{"openvas"}},
	{Category: BotScanner, Name: "Acunetix", Keywords: []string{"acunetix"}},
	{Category: BotScanner, Name: "BurpSuite", Keywords: []string{"burpsuite", "burp suite"}},
	{Category: BotScanner, Name: "Nmap", Keywords: []string{"nmap scripting engine", "nmap"}},
	{Category: BotScanner, Name: "ZAP", Keywords: []string{"zap", "owasp zap"}},
	{Category: BotScanner, Name: "WPScan", Keywords: []string{"wpscan"}},
	{Category: BotScanner, Name: "DirBuster", Keywords: []string{"dirbuster", "dirb"}},
	{Category: BotScanner, Name: "Gobuster", Keywords: []string{"gobuster"}},
	{Category: BotScanner, Name: "Hydra", Keywords: []string{"hydra"}},
	{Category: BotScanner, Name: "FFUF", Keywords: []string{"ffuf"}},
	{Category: BotScanner, Name: "Masscan", Keywords: []string{"masscan"}},

	// ── 恶意爬虫 / 内容抓取（block）──
	{Category: BotMalicious, Name: "MJ12bot", Keywords: []string{"mj12bot"}},
	{Category: BotMalicious, Name: "DotBot", Keywords: []string{"dotbot"}},
	{Category: BotMalicious, Name: "BLEXBot", Keywords: []string{"blexbot"}},
	{Category: BotMalicious, Name: "PetalBot", Keywords: []string{"petalbot"}},
	{Category: BotMalicious, Name: "Claudebot", Keywords: []string{"claudebot", "anthropic-ai", "claude-web"}},
	{Category: BotMalicious, Name: "GPTBot", Keywords: []string{"gptbot", "chatgpt-user", "openai"}},
	{Category: BotMalicious, Name: "CCBot", Keywords: []string{"ccbot", "commoncrawl"}},
	{Category: BotMalicious, Name: "PerplexityBot", Keywords: []string{"perplexitybot"}},
	{Category: BotMalicious, Name: "Copilot", Keywords: []string{"bingcopilot", "copilot"}},
}

// ─── Bot 检测器 ───

// BotDetector Bot检测器
// 结合 UA 指纹库 + 行为频率分析，识别爬虫/自动化工具并做出处置决策。
type BotDetector struct {
	mu sync.RWMutex

	// 配置
	enabled       bool
	windowSecs    int           // 行为分析时间窗口（秒）
	rateThreshold float64       // 高频率阈值（请求/秒），超过则标记为 Bot
	blockAutomation bool        // 是否直接拦截自动化工具
	blockScanners   bool        // 是否直接拦截扫描器
	blockMalicious   bool       // 是否直接拦截恶意爬虫

	// IP → 行为统计
	visits map[string]*botVisitCounter

	// 日志
	logger zerolog.Logger

	stopCh chan struct{}
	wg     sync.WaitGroup
}

// botVisitCounter 单 IP 访问计数器
type botVisitCounter struct {
	ip          string
	requestCount int       // 时间窗口内请求数
	windowStart  time.Time // 当前窗口起始时间
	uaList       []string  // 最近使用过的 UA（最多保留 10 个）
}

// BotDetectorConfig Bot检测器配置
type BotDetectorConfig struct {
	Enabled         bool    // 是否启用
	WindowSecs      int     // 行为分析时间窗口，默认 10 秒
	RateThreshold   float64 // 高频阈值（请求/秒），超过则标记为自动化行为，默认 20.0
	BlockAutomation bool    // 是否拦截自动化工具，默认 true
	BlockScanners   bool    // 是否拦截扫描器，默认 true
	BlockMalicious  bool    // 是否拦截恶意爬虫，默认 true
}

// DefaultBotDetectorConfig 默认配置
func DefaultBotDetectorConfig() BotDetectorConfig {
	return BotDetectorConfig{
		Enabled:         false,
		WindowSecs:      10,
		RateThreshold:   20.0,
		BlockAutomation: true,
		BlockScanners:   true,
		BlockMalicious:  true,
	}
}

// NewBotDetector 创建 Bot 检测器
func NewBotDetector(cfg BotDetectorConfig, logger zerolog.Logger) *BotDetector {
	return &BotDetector{
		enabled:         cfg.Enabled,
		windowSecs:      cfg.WindowSecs,
		rateThreshold:   cfg.RateThreshold,
		blockAutomation: cfg.BlockAutomation,
		blockScanners:   cfg.BlockScanners,
		blockMalicious:  cfg.BlockMalicious,
		visits:          make(map[string]*botVisitCounter),
		logger:          logger,
		stopCh:          make(chan struct{}),
	}
}

// Enable 启用检测器
func (bd *BotDetector) Enable() {
	bd.mu.Lock()
	defer bd.mu.Unlock()
	bd.enabled = true
}

// Disable 禁用检测器
func (bd *BotDetector) Disable() {
	bd.mu.Lock()
	defer bd.mu.Unlock()
	bd.enabled = false
}

// IsEnabled 检查是否启用
func (bd *BotDetector) IsEnabled() bool {
	bd.mu.RLock()
	defer bd.mu.RUnlock()
	return bd.enabled
}

// DetectBot 检测一个请求是否为爬虫/自动化工具
// 返回 BotDetectionResult，其中 Decision 指示应采取的处置。
func (bd *BotDetector) DetectBot(ip, userAgent string) BotDetectionResult {
	bd.mu.Lock()
	defer bd.mu.Unlock()

	result := BotDetectionResult{
		IP:        ip,
		UserAgent: userAgent,
		Category:  BotUnknown,
		Decision:  BotAllow,
	}

	// Disabled: skip all detection and always allow
	if !bd.enabled {
		return result
	}

	// Step 1: UA 指纹匹配
	result.Category, result.BotName = classifyUA(userAgent)
	switch result.Category {
	case BotScanner:
		result.Score += 40
		if bd.blockScanners {
			result.Decision = BotBlock
			result.Reason = "已识别扫描器UA: " + result.BotName
		} else {
			result.Decision = BotTag
			result.Reason = "扫描器UA标记: " + result.BotName
		}
	case BotAutomation:
		result.Score += 35
		if bd.blockAutomation {
			result.Decision = BotBlock
			result.Reason = "已识别自动化工具UA: " + result.BotName
		} else {
			result.Decision = BotTag
			result.Reason = "自动化工具UA标记: " + result.BotName
		}
	case BotMalicious:
		result.Score += 50
		if bd.blockMalicious {
			result.Decision = BotBlock
			result.Reason = "已识别恶意爬虫UA: " + result.BotName
		} else {
			result.Decision = BotTag
			result.Reason = "恶意爬虫UA标记: " + result.BotName
		}
	case BotSearch, BotSocial, BotMonitoring:
		result.Decision = BotAllow
		result.Reason = "合法爬虫/监控服务: " + result.BotName
	default:
		// 非已知 Bot → 行为频率分析
	}

	// Step 2: 行为频率分析（已放过 UA 级 Block 的请求仍需检查行为）
	if result.Decision != BotBlock && bd.enabled {
		now := time.Now()
		v, exists := bd.visits[ip]
		if !exists || now.Sub(v.windowStart).Seconds() > float64(bd.windowSecs) {
			bd.visits[ip] = &botVisitCounter{
				ip:          ip,
				requestCount: 1,
				windowStart:  now,
				uaList:       []string{userAgent},
			}
		} else {
			v.requestCount++
			if len(v.uaList) < 10 {
				v.uaList = append(v.uaList, userAgent)
			}

			elapsed := now.Sub(v.windowStart).Seconds()
			if elapsed > 0 {
				result.RateRPS = float64(v.requestCount) / elapsed
			}

			// 高频检测
			if result.RateRPS >= bd.rateThreshold {
				result.Score += 30
				if result.Decision == BotAllow {
					result.Decision = BotTag
					result.Reason = "高频请求行为异常"
				}
			}

			// UA 频繁切换检测
			if len(v.uaList) >= 5 && hasManyUniqueUAs(v.uaList, 3) {
				result.Score += 25
				result.Decision = BotBlock
				result.Reason = "频繁切换User-Agent疑似自动化脚本"
				result.Category = BotAutomation
			}
		}
	}

	return result
}

// classifyUA 根据 UA 字符串分类 Bot
func classifyUA(userAgent string) (BotCategory, string) {
	if userAgent == "" {
		return BotHuman, ""
	}

	uaLower := toLowerASCII(userAgent)
	for _, bot := range knownBots {
		for _, kw := range bot.Keywords {
			if containsASCII(uaLower, kw) {
				return bot.Category, bot.Name
			}
		}
	}

	return BotHuman, ""
}

// ─── 常用 UA 特征：判断是否为"类浏览器"请求 ───

// LooksHumanUA 判断 UA 是否为典型浏览器特征
func LooksHumanUA(userAgent string) bool {
	if userAgent == "" {
		return false
	}
	uaLower := toLowerASCII(userAgent)
	// 包含 Mozilla + (Chrome/Firefox/Safari/Edge) → 很可能是真实浏览器
	if containsASCII(uaLower, "mozilla") {
		if containsASCII(uaLower, "chrome") || containsASCII(uaLower, "firefox") ||
			containsASCII(uaLower, "safari") || containsASCII(uaLower, "edge") {
			return true
		}
	}
	return false
}

// ─── 后台清理 ───

// Start 启动后台清理协程
func (bd *BotDetector) Start() {
	bd.wg.Add(1)
	go func() {
		defer bd.wg.Done()
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-bd.stopCh:
				return
			case <-ticker.C:
				bd.cleanup()
			}
		}
	}()
}

// Stop 停止后台清理协程
func (bd *BotDetector) Stop() {
	close(bd.stopCh)
	bd.wg.Wait()
	bd.logger.Info().Msg("Bot 检测器已停止")
}

// cleanup 清理过期的访问计数器
func (bd *BotDetector) cleanup() {
	bd.mu.Lock()
	defer bd.mu.Unlock()

	now := time.Now()
	maxAge := float64(bd.windowSecs * 3)
	for ip, v := range bd.visits {
		if now.Sub(v.windowStart).Seconds() > maxAge {
			delete(bd.visits, ip)
		}
	}
}

// Stats 返回当前统计信息
func (bd *BotDetector) Stats() map[string]interface{} {
	bd.mu.RLock()
	defer bd.mu.RUnlock()

	tracked := len(bd.visits)
	highRate := 0
	for _, v := range bd.visits {
		elapsed := time.Since(v.windowStart).Seconds()
		if elapsed > 0 && float64(v.requestCount)/elapsed >= bd.rateThreshold {
			highRate++
		}
	}
	return map[string]interface{}{
		"enabled":           bd.enabled,
		"tracked_ips":       tracked,
		"high_rate_ips":     highRate,
		"rate_threshold":    bd.rateThreshold,
	}
}

// ─── ASCII 字符串工具（避免引入 strings 包的部分开销）───

// toLowerASCII 将 ASCII 字符串转为小写
func toLowerASCII(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		b[i] = c
	}
	return string(b)
}

// containsASCII 检查 s 中是否包含 substr（ASCII only）
func containsASCII(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// hasManyUniqueUAs 检查 UA 列表中是否包含多个不同的值
func hasManyUniqueUAs(uas []string, threshold int) bool {
	seen := make(map[string]struct{})
	for _, ua := range uas {
		seen[ua] = struct{}{}
		if len(seen) >= threshold {
			return true
		}
	}
	return false
}
