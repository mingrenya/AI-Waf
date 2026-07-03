package situation

import (
	"testing"

	"github.com/mingrenya/AI-Waf/pkg/model"
)

func TestCalculateRisk_LowRiskAttacker(t *testing.T) {
	profile := &model.AttackerProfile{
		TotalAttacks:      3,
		UniqueAttackTypes: 1,
		AttackPhase:       "unknown",
		IsAutomated:       false,
		IsPersistent:      false,
	}
	score := CalculateRisk(profile)
	// freq=9 + div=5 + stage=0 + persistent=0 + automated=0 = 14
	if score < 10 || score > 25 {
		t.Errorf("low-risk score expected 10~25, got %d", score)
	}
	label := RiskLabel(score)
	if label != "low" {
		t.Errorf("expected 'low', got '%s'", label)
	}
}

func TestCalculateRisk_MediumRiskAttacker(t *testing.T) {
	profile := &model.AttackerProfile{
		TotalAttacks:      20,
		UniqueAttackTypes: 3,
		AttackPhase:       "scanning",
		IsAutomated:       false,
		IsPersistent:      false,
	}
	score := CalculateRisk(profile)
	// freq=20*3=60 cap 30 + div=3*5=15 + scanning=10 = 55
	if score < 30 || score > 65 {
		t.Errorf("medium-risk score expected 30~65, got %d", score)
	}
	label := RiskLabel(score)
	if label != "medium" {
		t.Errorf("expected 'medium', got '%s'", label)
	}
}

func TestCalculateRisk_HighRiskAttacker(t *testing.T) {
	profile := &model.AttackerProfile{
		TotalAttacks:      100,
		UniqueAttackTypes: 4,
		AttackPhase:       "exploitation",
		IsAutomated:       true,
		IsPersistent:      false,
	}
	score := CalculateRisk(profile)
	// freq=30(max) + div=4*5=20(max) + exploitation=20 + automated=15 = 85
	if score < 65 || score > 95 {
		t.Errorf("high-risk score expected 65~95, got %d", score)
	}
	label := RiskLabel(score)
	if label != "high" && label != "critical" {
		t.Errorf("expected 'high' or 'critical', got '%s'", label)
	}
}

func TestCalculateRisk_CriticalAttacker(t *testing.T) {
	profile := &model.AttackerProfile{
		TotalAttacks:      500,
		UniqueAttackTypes: 6,
		AttackPhase:       "exploitation",
		IsAutomated:       true,
		IsPersistent:      true,
	}
	score := CalculateRisk(profile)
	// freq=30(max) + div=20(max) + exploitation=20 + persistent=15 + automated=15 = 100
	if score != 100 {
		t.Errorf("critical score expected 100, got %d", score)
	}
	label := RiskLabel(score)
	if label != "critical" {
		t.Errorf("expected 'critical', got '%s'", label)
	}
}

func TestCalculateRisk_CappedAt100(t *testing.T) {
	profile := &model.AttackerProfile{
		TotalAttacks:      9999,
		UniqueAttackTypes: 99,
		AttackPhase:       "exfiltration",
		IsAutomated:       true,
		IsPersistent:      true,
	}
	score := CalculateRisk(profile)
	if score > 100 {
		t.Errorf("score capped at 100, got %d", score)
	}
}

func TestCalculateRisk_ExfiltrationHighestPhase(t *testing.T) {
	profile := &model.AttackerProfile{
		TotalAttacks:      1,
		UniqueAttackTypes: 1,
		AttackPhase:       "exfiltration",
		IsAutomated:       false,
		IsPersistent:      false,
	}
	score := CalculateRisk(profile)
	// freq=3 + div=5 + exfiltration=40 = 48
	if score < 40 || score > 60 {
		t.Errorf("exfiltration phase score expected 40~60, got %d", score)
	}
}

func TestRiskLabel_Boundaries(t *testing.T) {
	tests := []struct {
		score int
		label string
	}{
		{0, "low"},
		{29, "low"},
		{30, "medium"},
		{59, "medium"},
		{60, "high"},
		{79, "high"},
		{80, "critical"},
		{100, "critical"},
	}
	for _, tc := range tests {
		got := RiskLabel(tc.score)
		if got != tc.label {
			t.Errorf("RiskLabel(%d) = %s, want %s", tc.score, got, tc.label)
		}
	}
}
