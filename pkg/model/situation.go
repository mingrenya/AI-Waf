package model

import "time"

// AttackStage 攻击阶段
type AttackStage string

const (
	StageUnknown         AttackStage = "unknown"
	StageReconnaissance  AttackStage = "reconnaissance"
	StageScanning        AttackStage = "scanning"
	StageExploitation    AttackStage = "exploitation"
	StageLateralMovement AttackStage = "lateral_movement"
	StageC2              AttackStage = "command_and_control"
	StageExfiltration    AttackStage = "exfiltration"
)

// Order 返回攻击阶段的顺序值（数值越大越靠后）
func (s AttackStage) Order() int {
	switch s {
	case StageReconnaissance:
		return 1
	case StageScanning:
		return 2
	case StageExploitation:
		return 3
	case StageLateralMovement:
		return 4
	case StageC2:
		return 4
	case StageExfiltration:
		return 5
	default:
		return 0
	}
}

// StageMapping attack_type → AttackStage 映射
var StageMapping = map[string]AttackStage{
	"scanner":               StageReconnaissance,
	"vulnerability_scanner": StageScanning,
	"port_scan":             StageScanning,
	"directory_bruteforce":  StageScanning,
	"sql_injection":         StageExploitation,
	"xss":                   StageExploitation,
	"rce":                   StageExploitation,
	"file_inclusion":        StageExploitation,
	"csrf":                  StageExploitation,
	"ssrf":                  StageExploitation,
	"command_injection":     StageExploitation,
	"webshell":              StageLateralMovement,
	"backdoor":              StageC2,
	"data_leak":             StageExfiltration,
}

// ChainStage 攻击链阶段
type ChainStage struct {
	Stage      AttackStage `json:"stage" bson:"stage"`
	Technique  string      `json:"technique" bson:"technique"`
	DetectedAt time.Time   `json:"detected_at" bson:"detected_at"`
	Evidence   []string    `json:"evidence" bson:"evidence"`
	Confidence float64     `json:"confidence" bson:"confidence"`
}

// AttackChain 攻击链
type AttackChain struct {
	ID             string       `json:"id" bson:"_id"`
	SourceIP       string       `json:"source_ip" bson:"source_ip"`
	GeoCountry     string       `json:"geo_country" bson:"geo_country"`
	Stages         []ChainStage `json:"stages" bson:"stages"`
	CorrelationIDs []string     `json:"correlation_ids" bson:"correlation_ids"`
	RiskScore      int          `json:"risk_score" bson:"risk_score"`
	FirstSeen      time.Time    `json:"first_seen" bson:"first_seen"`
	LastSeen       time.Time    `json:"last_seen" bson:"last_seen"`
	Active         bool         `json:"active" bson:"active"`
	CreatedAt      time.Time    `json:"created_at" bson:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at" bson:"updated_at"`
}

func (AttackChain) GetCollectionName() string { return "attack_chains" }

// AttackerProfile 攻击者画像
type AttackerProfile struct {
	ID                string    `json:"id" bson:"_id"`
	SourceIP          string    `json:"source_ip" bson:"source_ip"`
	GeoCountry        string    `json:"geo_country" bson:"geo_country"`
	GeoCity           string    `json:"geo_city,omitempty" bson:"geo_city,omitempty"`
	TotalAttacks      int       `json:"total_attacks" bson:"total_attacks"`
	UniqueAttackTypes int       `json:"unique_attack_types" bson:"unique_attack_types"`
	TopAttackType     string    `json:"top_attack_type" bson:"top_attack_type"`
	UniqueTargetSites int       `json:"unique_target_sites" bson:"unique_target_sites"`
	ActiveHours       []int     `json:"active_hours" bson:"active_hours"`
	BurstIntervals    []string  `json:"burst_intervals" bson:"burst_intervals"`
	AttackPhase       string    `json:"attack_phase" bson:"attack_phase"`
	ToolsIdentified   string    `json:"tools_identified" bson:"tools_identified"`
	IsAutomated       bool      `json:"is_automated" bson:"is_automated"`
	IsPersistent      bool      `json:"is_persistent" bson:"is_persistent"`
	RiskScore         int       `json:"risk_score" bson:"risk_score"`
	RiskLabel         string    `json:"risk_label" bson:"risk_label"`
	LastSeen          time.Time `json:"last_seen" bson:"last_seen"`
	FirstSeen         time.Time `json:"first_seen" bson:"first_seen"`
	UpdatedAt         time.Time `json:"updated_at" bson:"updated_at"`
}

func (AttackerProfile) GetCollectionName() string { return "attacker_profiles" }

// SituationRule 态势检测规则
type SituationRule struct {
	ID             string    `json:"id" bson:"_id"`
	Name           string    `json:"name" bson:"name"`
	Stage          string    `json:"stage" bson:"stage"`
	LogQL          string    `json:"logql" bson:"logql"`
	Interval       int       `json:"interval_seconds" bson:"interval_seconds"`
	Threshold      int       `json:"threshold" bson:"threshold"`
	Severity       string    `json:"severity" bson:"severity"`
	MITRETactic    string    `json:"mitre_tactic" bson:"mitre_tactic"`
	MITRETechnique string    `json:"mitre_technique" bson:"mitre_technique"`
	Enabled        bool      `json:"enabled" bson:"enabled"`
	CreatedAt      time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" bson:"updated_at"`
}

func (SituationRule) GetCollectionName() string { return "situation_rules" }

// SituationSnapshot 态势快照
type SituationSnapshot struct {
	ID               string         `json:"id" bson:"_id"`
	Timestamp        time.Time      `json:"timestamp" bson:"timestamp"`
	TotalChains      int            `json:"total_chains" bson:"total_chains"`
	ActiveChains     int            `json:"active_chains" bson:"active_chains"`
	ByAttackType     map[string]int `json:"by_attack_type" bson:"by_attack_type"`
	ByCountry        map[string]int `json:"by_country" bson:"by_country"`
	ByStage          map[string]int `json:"by_stage" bson:"by_stage"`
	OverallRiskScore float64        `json:"overall_risk_score" bson:"overall_risk_score"`
}

func (SituationSnapshot) GetCollectionName() string { return "situation_snapshots" }
