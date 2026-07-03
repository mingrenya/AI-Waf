package situation

import (
	"strings"
	"testing"
)

func TestLogQLBuilder_NoFilters(t *testing.T) {
	b := NewLogQLBuilder()
	got := b.RawQuery()
	expected := `{container_name="mrya-waf"}`
	if got != expected {
		t.Errorf("expected '%s', got '%s'", expected, got)
	}
}

func TestLogQLBuilder_FilterIP(t *testing.T) {
	b := NewLogQLBuilder()
	got := b.FilterIP("1.2.3.4").RawQuery()
	if !strings.Contains(got, `source_ip="1.2.3.4"`) {
		t.Errorf("expected source_ip filter, got '%s'", got)
	}
}

func TestLogQLBuilder_FilterAttackType(t *testing.T) {
	b := NewLogQLBuilder()
	got := b.FilterAttackType("sql_injection").RawQuery()
	if !strings.Contains(got, `attack_type="sql_injection"`) {
		t.Errorf("expected attack_type filter, got '%s'", got)
	}
}

func TestLogQLBuilder_FilterSeverityRegex(t *testing.T) {
	b := NewLogQLBuilder()
	got := b.FilterSeverity("critical|high").RawQuery()
	if !strings.Contains(got, `severity=~"critical|high"`) {
		t.Errorf("expected severity regex filter, got '%s'", got)
	}
}

func TestLogQLBuilder_MultipleFilters(t *testing.T) {
	b := NewLogQLBuilder()
	got := b.FilterIP("1.2.3.4").FilterAttackType("xss").FilterSeverity("critical").RawQuery()
	if !strings.Contains(got, "source_ip") {
		t.Error("missing source_ip")
	}
	if !strings.Contains(got, "attack_type") {
		t.Error("missing attack_type")
	}
	if !strings.Contains(got, "severity") {
		t.Error("missing severity")
	}
	// Must start with { and end with }
	if !strings.HasPrefix(got, "{") || !strings.HasSuffix(got, "}") {
		t.Errorf("bad selector format: %s", got)
	}
}

func TestLogQLBuilder_CountOverTime(t *testing.T) {
	b := NewLogQLBuilder()
	got := b.FilterAttackType("scanner").CountOverTime("5m")
	if !strings.Contains(got, "sum by (source_ip)") {
		t.Errorf("expected sum aggregation, got '%s'", got)
	}
	if !strings.Contains(got, "[5m]") {
		t.Errorf("expected [5m] window, got '%s'", got)
	}
}

func TestLogQLBuilder_AttackChainQuery(t *testing.T) {
	b := NewLogQLBuilder()
	got := b.AttackChainQuery("10.0.0.1")
	if !strings.Contains(got, `source_ip="10.0.0.1"`) {
		t.Errorf("missing source_ip in chain query: %s", got)
	}
	if !strings.Contains(got, "attack_type") {
		t.Errorf("missing attack_type in chain query: %s", got)
	}
}

func TestLogQLBuilder_ChainedFiltersAreComposable(t *testing.T) {
	b := NewLogQLBuilder()
	got := b.FilterIP("1.2.3.4").FilterAttackType("xss").RawQuery()
	// 链式调用返回的查询应包含所有过滤条件
	if !strings.Contains(got, "source_ip") || !strings.Contains(got, "attack_type") {
		t.Errorf("chained filters not all present: %s", got)
	}
}
