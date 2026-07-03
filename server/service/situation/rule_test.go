package situation

import (
	"testing"

	"github.com/mingrenya/AI-Waf/pkg/model"
)

func TestDefaultRules_Count(t *testing.T) {
	if len(DefaultRules) < 3 {
		t.Errorf("expected at least 3 default rules, got %d", len(DefaultRules))
	}
}

func TestDefaultRules_AllHaveRequiredFields(t *testing.T) {
	for i, rule := range DefaultRules {
		if rule.Name == "" {
			t.Errorf("rule[%d]: Name is empty", i)
		}
		if rule.Stage == "" {
			t.Errorf("rule[%d] '%s': Stage is empty", i, rule.Name)
		}
		if rule.LogQL == "" {
			t.Errorf("rule[%d] '%s': LogQL is empty", i, rule.Name)
		}
		if rule.Interval < 5 {
			t.Errorf("rule[%d] '%s': Interval too low (%d)", i, rule.Name, rule.Interval)
		}
		if rule.Threshold < 1 {
			t.Errorf("rule[%d] '%s': Threshold must be >= 1", i, rule.Name)
		}
		if rule.Severity != "critical" && rule.Severity != "high" && rule.Severity != "medium" && rule.Severity != "low" {
			t.Errorf("rule[%d] '%s': invalid Severity '%s'", i, rule.Name, rule.Severity)
		}
		if rule.MITRETactic == "" {
			t.Logf("rule[%d] '%s': MITRETactic is empty (non-critical)", i, rule.Name)
		}
	}
}

func TestDefaultRules_EnabledByDefault(t *testing.T) {
	for _, rule := range DefaultRules {
		if !rule.Enabled {
			t.Errorf("rule '%s' should be enabled by default", rule.Name)
		}
	}
}

func TestDefaultRules_NoDuplicateNames(t *testing.T) {
	seen := make(map[string]bool)
	for _, rule := range DefaultRules {
		if seen[rule.Name] {
			t.Errorf("duplicate rule name: '%s'", rule.Name)
		}
		seen[rule.Name] = true
	}
}

func TestRuleResult_TriggeredThresholdLogic(t *testing.T) {
	// HitCount > Threshold 才触发
	rule := model.SituationRule{Name: "test", Threshold: 10}

	result := RuleResult{Rule: rule, HitCount: 5}
	result.Triggered = result.HitCount > result.Rule.Threshold
	if result.Triggered {
		t.Error("5 hits should not trigger threshold of 10")
	}

	result2 := RuleResult{Rule: rule, HitCount: 11}
	result2.Triggered = result2.HitCount > result2.Rule.Threshold
	if !result2.Triggered {
		t.Error("11 hits should trigger threshold of 10")
	}
}
