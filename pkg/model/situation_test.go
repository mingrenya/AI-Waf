package model

import (
	"testing"
)

func TestAttackStage_Order_AllConstants(t *testing.T) {
	tests := []struct {
		stage    AttackStage
		expected int
	}{
		{StageUnknown, 0},
		{StageReconnaissance, 1},
		{StageScanning, 2},
		{StageExploitation, 3},
		{StageLateralMovement, 4},
		{StageC2, 4},
		{StageExfiltration, 5},
	}
	for _, tc := range tests {
		got := tc.stage.Order()
		if got != tc.expected {
			t.Errorf("%s.Order() = %d, want %d", tc.stage, got, tc.expected)
		}
	}
}

func TestAttackStage_Order_Unknown(t *testing.T) {
	var s AttackStage = "nonexistent"
	if s.Order() != 0 {
		t.Errorf("unknown stage should have Order=0, got %d", s.Order())
	}
}

func TestStageMapping_AllKeysHaveValidStage(t *testing.T) {
	for attackType, stage := range StageMapping {
		if stage.Order() == 0 && stage != StageUnknown {
			t.Errorf("StageMapping[%s] = %s has Order=0", attackType, stage)
		}
	}
}

func TestStageMapping_ExploitationTypes(t *testing.T) {
	exploitTypes := []string{"sql_injection", "xss", "rce", "file_inclusion", "csrf", "ssrf", "command_injection"}
	for _, at := range exploitTypes {
		stage, ok := StageMapping[at]
		if !ok {
			t.Errorf("missing %s in StageMapping", at)
			continue
		}
		if stage != StageExploitation {
			t.Errorf("%s should be exploitation, got %s", at, stage)
		}
	}
}

func TestStageMapping_ScanTypes(t *testing.T) {
	scanTypes := []string{"vulnerability_scanner", "port_scan", "directory_bruteforce"}
	for _, at := range scanTypes {
		stage, ok := StageMapping[at]
		if !ok {
			t.Errorf("missing %s in StageMapping", at)
			continue
		}
		if stage != StageScanning {
			t.Errorf("%s should be scanning, got %s", at, stage)
		}
	}
}

func TestSituationRule_CollectionName(t *testing.T) {
	r := SituationRule{}
	if r.GetCollectionName() != "situation_rules" {
		t.Errorf("expected 'situation_rules', got '%s'", r.GetCollectionName())
	}
}

func TestAttackChain_CollectionName(t *testing.T) {
	c := AttackChain{}
	if c.GetCollectionName() != "attack_chains" {
		t.Errorf("expected 'attack_chains', got '%s'", c.GetCollectionName())
	}
}

func TestAttackerProfile_CollectionName(t *testing.T) {
	p := AttackerProfile{}
	if p.GetCollectionName() != "attacker_profiles" {
		t.Errorf("expected 'attacker_profiles', got '%s'", p.GetCollectionName())
	}
}

func TestSituationSnapshot_CollectionName(t *testing.T) {
	s := SituationSnapshot{}
	if s.GetCollectionName() != "situation_snapshots" {
		t.Errorf("expected 'situation_snapshots', got '%s'", s.GetCollectionName())
	}
}
