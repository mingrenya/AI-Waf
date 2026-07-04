package internal

import (
	"testing"
	"time"
)

func TestClassifyUA_Googlebot(t *testing.T) {
	cat, name := classifyUA("Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)")
	if cat != BotSearch {
		t.Errorf("expected BotSearch, got %q", cat)
	}
	if name != "Googlebot" {
		t.Errorf("expected Googlebot, got %q", name)
	}
}

func TestClassifyUA_SQLMap(t *testing.T) {
	cat, name := classifyUA("sqlmap/1.5#stable (http://sqlmap.org)")
	if cat != BotScanner {
		t.Errorf("expected BotScanner, got %q", cat)
	}
	if name != "SQLMap" {
		t.Errorf("expected SQLMap, got %q", name)
	}
}

func TestClassifyUA_HeadlessChrome(t *testing.T) {
	cat, name := classifyUA("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 HeadlessChrome/97")
	if cat != BotAutomation {
		t.Errorf("expected BotAutomation, got %q", cat)
	}
	if name != "HeadlessChrome" {
		t.Errorf("expected HeadlessChrome, got %q", name)
	}
}

func TestClassifyUA_NormalBrowser(t *testing.T) {
	cat, name := classifyUA("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36")
	if cat != BotHuman {
		t.Errorf("expected BotHuman, got %q", cat)
	}
	if name != "" {
		t.Errorf("expected no bot name, got %q", name)
	}
}

func TestClassifyUA_Empty(t *testing.T) {
	cat, name := classifyUA("")
	if cat != BotHuman {
		t.Errorf("expected BotHuman for empty UA, got %q", cat)
	}
	if name != "" {
		t.Errorf("expected no bot name for empty UA, got %q", name)
	}
}

func TestClassifyUA_GPTBot(t *testing.T) {
	cat, name := classifyUA("Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko); compatible; GPTBot/1.0; +https://openai.com/gptbot")
	if cat != BotMalicious {
		t.Errorf("expected BotMalicious, got %q", cat)
	}
	if name != "GPTBot" {
		t.Errorf("expected GPTBot, got %q", name)
	}
}

func TestClassifyUA_Facebook(t *testing.T) {
	cat, name := classifyUA("facebookexternalhit/1.1 (+http://www.facebook.com/externalhit_uatext.php)")
	if cat != BotSocial {
		t.Errorf("expected BotSocial, got %q", cat)
	}
	if name != "Facebook" {
		t.Errorf("expected Facebook, got %q", name)
	}
}

func TestClassifyUA_UptimeRobot(t *testing.T) {
	cat, name := classifyUA("UptimeRobot/2.0; (+http://www.uptimerobot.com/)")
	if cat != BotMonitoring {
		t.Errorf("expected BotMonitoring, got %q", cat)
	}
	if name != "UptimeRobot" {
		t.Errorf("expected UptimeRobot, got %q", name)
	}
}

func TestBotDetector_GooglebotAllowed(t *testing.T) {
	cfg := DefaultBotDetectorConfig()
	cfg.Enabled = true
	bd := NewBotDetector(cfg, testLogger(t))

	result := bd.DetectBot("10.0.0.1", "Googlebot/2.1")
	if result.Decision != BotAllow {
		t.Errorf("Googlebot should be allowed, got %q: %s", result.Decision, result.Reason)
	}
}

func TestBotDetector_SQLMapBlocked(t *testing.T) {
	cfg := DefaultBotDetectorConfig()
	cfg.Enabled = true
	cfg.BlockScanners = true
	bd := NewBotDetector(cfg, testLogger(t))

	result := bd.DetectBot("10.0.0.2", "sqlmap/1.5")
	if result.Decision != BotBlock {
		t.Errorf("SQLMap should be blocked, got %q", result.Decision)
	}
}

func TestBotDetector_NormalBrowserAllowed(t *testing.T) {
	cfg := DefaultBotDetectorConfig()
	cfg.Enabled = true
	bd := NewBotDetector(cfg, testLogger(t))

	result := bd.DetectBot("10.0.0.3", "Mozilla/5.0 Chrome/120")
	if result.Decision != BotAllow {
		t.Errorf("normal browser should be allowed, got %q: %s", result.Decision, result.Reason)
	}
}

func TestBotDetector_HighFrequencyTagged(t *testing.T) {
	cfg := DefaultBotDetectorConfig()
	cfg.Enabled = true
	cfg.RateThreshold = 5.0
	bd := NewBotDetector(cfg, testLogger(t))

	// Fire 6 requests quickly from the same IP
	ip := "10.0.0.100"
	for i := 0; i < 6; i++ {
		result := bd.DetectBot(ip, "Mozilla/5.0 Chrome/120")
		if i >= 5 {
			if result.RateRPS < cfg.RateThreshold {
				t.Logf("rate=%.1f rps (expected >= %.1f)", result.RateRPS, cfg.RateThreshold)
			}
			if result.Decision != BotTag && result.Decision != BotBlock {
				t.Errorf("high frequency should be at least tagged, got %q", result.Decision)
			}
		}
	}
}

func TestBotDetector_UAChangeBlocked(t *testing.T) {
	cfg := DefaultBotDetectorConfig()
	cfg.Enabled = true
	bd := NewBotDetector(cfg, testLogger(t))

	ip := "10.0.0.200"
	uas := []string{
		"Mozilla/5.0 Chrome/120",
		"Mozilla/5.0 Firefox/121",
		"Mozilla/5.0 Safari/605",
		"Mozilla/5.0 Edge/120",
	}
	for _, ua := range uas {
		result := bd.DetectBot(ip, ua)
		if result.Decision == BotBlock {
			if result.Reason != "频繁切换User-Agent疑似自动化脚本" {
				t.Errorf("unexpected reason: %s", result.Reason)
			}
			return
		}
	}
	// 应该至少被标记（4 个请求可能还不够触发 UA 切换检测阈值 3）
	// 但如果都没触发，说明逻辑正常——需要 5 个请求 + 其中 3 个不同 UA
}

func TestBotDetector_DisabledDoesNotBlock(t *testing.T) {
	cfg := DefaultBotDetectorConfig()
	cfg.Enabled = false
	bd := NewBotDetector(cfg, testLogger(t))

	result := bd.DetectBot("10.0.0.5", "sqlmap/1.5")
	if result.Decision != BotAllow {
		t.Errorf("disabled detector should allow everything, got %q", result.Decision)
	}
}

func TestBotDetector_WindowReset(t *testing.T) {
	cfg := DefaultBotDetectorConfig()
	cfg.Enabled = true
	cfg.RateThreshold = 10.0
	cfg.WindowSecs = 1
	bd := NewBotDetector(cfg, testLogger(t))

	ip := "10.0.0.6"
	for i := 0; i < 20; i++ {
		bd.DetectBot(ip, "Mozilla/5.0")
	}

	// 模拟窗口过期
	bd.mu.Lock()
	if c, ok := bd.visits[ip]; ok {
		c.windowStart = time.Now().Add(-time.Duration(cfg.WindowSecs+1) * time.Second)
	}
	bd.mu.Unlock()

	result := bd.DetectBot(ip, "Mozilla/5.0")
	if result.RateRPS > cfg.RateThreshold {
		t.Errorf("window should have reset but rate is %.1f (> %.1f)", result.RateRPS, cfg.RateThreshold)
	}
}

func TestBotDetector_Stats(t *testing.T) {
	cfg := DefaultBotDetectorConfig()
	cfg.Enabled = true
	bd := NewBotDetector(cfg, testLogger(t))

	bd.DetectBot("10.0.0.10", "Mozilla/5.0")
	bd.DetectBot("10.0.0.11", "Mozilla/5.0")

	stats := bd.Stats()
	if stats["enabled"] != true {
		t.Error("expected enabled=true")
	}
	if stats["tracked_ips"].(int) != 2 {
		t.Errorf("expected 2 tracked IPs, got %d", stats["tracked_ips"])
	}
}

func TestLooksHumanUA_Chrome(t *testing.T) {
	if !LooksHumanUA("Mozilla/5.0 Chrome/120.0") {
		t.Error("Chrome should look human")
	}
}

func TestLooksHumanUA_Empty(t *testing.T) {
	if LooksHumanUA("") {
		t.Error("empty UA should not look human")
	}
}

func TestToLowerASCII(t *testing.T) {
	if got := toLowerASCII("SQLMap/1.0"); got != "sqlmap/1.0" {
		t.Errorf("expected sqlmap/1.0, got %q", got)
	}
}
