package situation

import (
	"github.com/mingrenya/AI-Waf/pkg/model"
)

var attackStageWeight = map[string]int{
	string(model.StageReconnaissance):   5,
	string(model.StageScanning):        10,
	string(model.StageExploitation):    20,
	string(model.StageLateralMovement): 30,
	string(model.StageC2):              30,
	string(model.StageExfiltration):    40,
}

// CalculateRisk 计算风险评分 0-100
func CalculateRisk(profile *model.AttackerProfile) int {
	score := 0
	freqScore := profile.TotalAttacks * 3
	if freqScore > 30 {
		freqScore = 30
	}
	score += freqScore
	divScore := profile.UniqueAttackTypes * 5
	if divScore > 20 {
		divScore = 20
	}
	score += divScore
	if w, ok := attackStageWeight[profile.AttackPhase]; ok {
		score += w
	}
	if profile.IsPersistent {
		score += 15
	}
	if profile.IsAutomated {
		score += 15
	}
	if score > 100 {
		score = 100
	}
	return score
}

// RiskLabel 根据评分返回风险标签
func RiskLabel(score int) string {
	switch {
	case score >= 80:
		return "critical"
	case score >= 60:
		return "high"
	case score >= 30:
		return "medium"
	default:
		return "low"
	}
}
