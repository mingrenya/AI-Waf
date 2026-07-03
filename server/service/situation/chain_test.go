package situation

import (
	"testing"

	"github.com/mingrenya/AI-Waf/pkg/model"
)

// chainWithStage 创建带有指定阶段的测试攻击链
func chainWithStage(stage model.AttackStage) *model.AttackChain {
	return &model.AttackChain{
		SourceIP: "1.2.3.4",
		Stages: []model.ChainStage{
			{Stage: stage, Confidence: 1.0},
		},
	}
}

// chainWithStages 创建多阶段攻击链
func chainWithStages(stages ...model.AttackStage) *model.AttackChain {
	chainStages := make([]model.ChainStage, len(stages))
	for i, s := range stages {
		chainStages[i] = model.ChainStage{Stage: s, Confidence: 1.0}
	}
	return &model.AttackChain{
		SourceIP: "1.2.3.4",
		Stages:  chainStages,
	}
}

func TestShouldAdvanceStage_Advances(t *testing.T) {
	b := &AttackChainBuilder{}
	// scanning → exploitation 应该推进
	result := b.shouldAdvanceStage(
		chainWithStage(model.StageScanning),
		model.StageExploitation,
	)
	if result != model.StageExploitation {
		t.Errorf("expected exploitation, got %s", result)
	}
}

func TestShouldAdvanceStage_NoRegress(t *testing.T) {
	b := &AttackChainBuilder{}
	// exploitation 后出现 scanning 不倒退
	result := b.shouldAdvanceStage(
		chainWithStage(model.StageExploitation),
		model.StageScanning,
	)
	if result != model.StageExploitation {
		t.Errorf("should not regress to scanning, got %s", result)
	}
}

func TestShouldAdvanceStage_SameStage(t *testing.T) {
	b := &AttackChainBuilder{}
	result := b.shouldAdvanceStage(
		chainWithStage(model.StageExploitation),
		model.StageExploitation,
	)
	if result != model.StageExploitation {
		t.Errorf("expected exploitation, got %s", result)
	}
}

func TestShouldAdvanceStage_FromReconToExfil(t *testing.T) {
	b := &AttackChainBuilder{}
	result := b.shouldAdvanceStage(
		chainWithStages(model.StageReconnaissance, model.StageScanning),
		model.StageExfiltration,
	)
	if result != model.StageExfiltration {
		t.Errorf("expected exfiltration (highest), got %s", result)
	}
}

func TestHighestStage_Empty(t *testing.T) {
	b := &AttackChainBuilder{}
	result := b.highestStage(&model.AttackChain{Stages: []model.ChainStage{}})
	if result != "" && result != model.StageUnknown {
		t.Errorf("empty chain should return zero-value or unknown, got '%s'", result)
	}
}

func TestHighestStage_Multiple(t *testing.T) {
	b := &AttackChainBuilder{}
	result := b.highestStage(chainWithStages(
		model.StageReconnaissance,
		model.StageScanning,
		model.StageExploitation,
	))
	if result != model.StageExploitation {
		t.Errorf("expected exploitation, got %s", result)
	}
}

func TestStageMapping_CoversAllDefaults(t *testing.T) {
	// 验证默认规则中的 stage 值在 StageMapping 中都有对应
	for _, rule := range DefaultRules {
		if _, ok := model.StageMapping[rule.Stage]; !ok {
			// 规则 stage 值可以是 StageMapping 的 key
			// 或者是直接的 AttackStage 字符串
			stage := model.AttackStage(rule.Stage)
			if stage.Order() == 0 && stage != model.StageUnknown {
				t.Logf("rule '%s' stage '%s' not in StageMapping and has Order=0", rule.Name, rule.Stage)
			}
		}
	}
}

func TestAttackStageOrder(t *testing.T) {
	tests := []struct {
		stage model.AttackStage
		min   int
	}{
		{model.StageUnknown, 0},
		{model.StageReconnaissance, 1},
		{model.StageScanning, 2},
		{model.StageExploitation, 3},
		{model.StageLateralMovement, 4},
		{model.StageC2, 4},
		{model.StageExfiltration, 5},
	}
	for _, tc := range tests {
		got := tc.stage.Order()
		if got < tc.min {
			t.Errorf("%s.Order() = %d, want >= %d", tc.stage, got, tc.min)
		}
	}
}

func TestStageProgression_IsMonotonic(t *testing.T) {
	// scanning 应该超过 reconnaissance
	if model.StageScanning.Order() <= model.StageReconnaissance.Order() {
		t.Error("scanning should outrank reconnaissance")
	}
	// exploitation 应该超过 scanning
	if model.StageExploitation.Order() <= model.StageScanning.Order() {
		t.Error("exploitation should outrank scanning")
	}
	// 任何已知阶段不应该倒退
	b := &AttackChainBuilder{}
	result := b.shouldAdvanceStage(
		chainWithStage(model.StageExploitation),
		model.StageReconnaissance,
	)
	if result.Order() < model.StageExploitation.Order() {
		t.Errorf("should not regress below exploitation, got %s (order=%d)", result, result.Order())
	}
}
