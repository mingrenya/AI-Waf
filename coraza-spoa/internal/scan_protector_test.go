package internal

import (
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestScanProtector_RecordHit_NewWindow(t *testing.T) {
	cfg := DefaultScanProtectorConfig()
	cfg.Enabled = true
	sp := NewScanProtector(cfg, nil, "", testLogger(t))

	shouldBlock := sp.RecordHit("10.0.0.1", 1001)
	if shouldBlock {
		t.Error("first hit should not block")
	}
}

func TestScanProtector_RecordHit_ThresholdNotReached(t *testing.T) {
	cfg := DefaultScanProtectorConfig()
	cfg.Enabled = true
	cfg.TriggerCount = 5
	cfg.TriggerTypes = 2
	sp := NewScanProtector(cfg, nil, "", testLogger(t))

	for i := 0; i < 4; i++ {
		shouldBlock := sp.RecordHit("10.0.0.2", 1000+i)
		if shouldBlock {
			t.Fatalf("hit %d should not block", i+1)
		}
	}
}

func TestScanProtector_RecordHit_BlockedAfterThreshold(t *testing.T) {
	cfg := DefaultScanProtectorConfig()
	cfg.Enabled = true
	cfg.TriggerCount = 3
	cfg.TriggerTypes = 1
	sp := NewScanProtector(cfg, nil, "", testLogger(t))

	sp.RecordHit("10.0.0.3", 2001)
	sp.RecordHit("10.0.0.3", 2001)
	shouldBlock := sp.RecordHit("10.0.0.3", 2001)

	if !shouldBlock {
		t.Fatal("should block after reaching threshold")
	}
}

func TestScanProtector_RecordHit_ResetsOnNewWindow(t *testing.T) {
	cfg := DefaultScanProtectorConfig()
	cfg.Enabled = true
	cfg.TriggerCount = 10
	cfg.TriggerTypes = 1
	sp := NewScanProtector(cfg, nil, "", testLogger(t))

	sp.RecordHit("10.0.0.4", 3001)
	sp.RecordHit("10.0.0.4", 3002)

	// 模拟窗口过期：手动修改计数器时间
	sp.mu.Lock()
	if c, ok := sp.counters["10.0.0.4"]; ok {
		c.windowStart = time.Now().Add(-time.Duration(cfg.WindowSecs+1) * time.Second)
	}
	sp.mu.Unlock()

	// 下一次 hit 应该开启新窗口，不会触发封禁
	shouldBlock := sp.RecordHit("10.0.0.4", 3003)
	if shouldBlock {
		t.Fatal("should reset counter on new window")
	}
}

func TestScanProtector_DisabledDoesNotCount(t *testing.T) {
	cfg := DefaultScanProtectorConfig()
	cfg.Enabled = false
	sp := NewScanProtector(cfg, nil, "", testLogger(t))

	for i := 0; i < 50; i++ {
		if sp.RecordHit("10.0.0.5", 4001) {
			t.Fatal("disabled protector should never block")
		}
	}
}

func TestDetectScanner_SQLMap(t *testing.T) {
	if got := DetectScanner("sqlmap/1.0#stable"); got != "SQLMap" {
		t.Errorf("expected SQLMap, got %q", got)
	}
}

func TestDetectScanner_Nuclei(t *testing.T) {
	if got := DetectScanner("Nuclei - Open-source project"); got != "Nuclei" {
		t.Errorf("expected Nuclei, got %q", got)
	}
}

func TestDetectScanner_NormalBrowser(t *testing.T) {
	if got := DetectScanner("Mozilla/5.0 Chrome/120.0"); got != "" {
		t.Errorf("expected empty for normal browser, got %q", got)
	}
}

func TestDetectScanner_EmptyUA(t *testing.T) {
	if got := DetectScanner(""); got != "" {
		t.Errorf("expected empty for empty UA, got %q", got)
	}
}

func TestScanProtector_Stats(t *testing.T) {
	cfg := DefaultScanProtectorConfig()
	cfg.Enabled = true
	sp := NewScanProtector(cfg, nil, "", testLogger(t))

	sp.RecordHit("10.0.0.6", 5001)
	sp.RecordHit("10.0.0.7", 5002)

	stats := sp.Stats()
	if stats["enabled"] != true {
		t.Error("expected enabled=true")
	}
	if stats["total_tracked"].(int) != 2 {
		t.Error("expected 2 tracked IPs")
	}
}

func testLogger(t *testing.T) zerolog.Logger {
	return zerolog.New(zerolog.NewTestWriter(t))
}
