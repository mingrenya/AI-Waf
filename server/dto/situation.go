package dto

import "time"

// === 态势概览 ===

type SituationOverviewResponse struct {
	ActiveChains      int         `json:"active_chains"`
	TotalChains24h    int         `json:"total_chains_24h"`
	TotalAttackers24h int         `json:"total_attackers_24h"`
	OverallRiskScore  float64     `json:"overall_risk_score"`
	RiskTrend         string      `json:"risk_trend"`
	TopAttackTypes    []CountItem `json:"top_attack_types"`
	TopAttackerIPs    []CountItem `json:"top_attacker_ips"`
	TopTargetSites    []CountItem `json:"top_target_sites"`
	ByCountry         []CountItem `json:"by_country"`
}

type CountItem struct {
	Label string `json:"label"`
	Count int    `json:"count"`
}

// === 攻击链 ===

type ChainListRequest struct {
	SourceIP string `form:"source_ip"`
	Stage    string `form:"stage"`
	Active   *bool  `form:"active"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

type ChainListResponse struct {
	Chains   []ChainSummary `json:"chains"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
}

type ChainSummary struct {
	ID         string    `json:"id"`
	SourceIP   string    `json:"source_ip"`
	GeoCountry string    `json:"geo_country"`
	Stages     []string  `json:"stages"`
	RiskScore  int       `json:"risk_score"`
	FirstSeen  time.Time `json:"first_seen"`
	LastSeen   time.Time `json:"last_seen"`
	Active     bool      `json:"active"`
}

type ChainDetailResponse struct {
	ID              string                 `json:"id"`
	SourceIP        string                 `json:"source_ip"`
	GeoCountry      string                 `json:"geo_country"`
	Stages          []ChainStageItem       `json:"stages"`
	CorrelationIDs  []string               `json:"correlation_ids"`
	RiskScore       int                    `json:"risk_score"`
	RiskLabel       string                 `json:"risk_label"`
	FirstSeen       time.Time              `json:"first_seen"`
	LastSeen        time.Time              `json:"last_seen"`
	Active          bool                   `json:"active"`
	AttackerProfile *AttackerProfileDetail `json:"attacker_profile,omitempty"`
}

type ChainStageItem struct {
	Stage      string    `json:"stage"`
	Technique  string    `json:"technique"`
	DetectedAt time.Time `json:"detected_at"`
	Confidence float64   `json:"confidence"`
	Evidence   []string  `json:"evidence"`
}

// === 攻击者 ===

type AttackerListRequest struct {
	Page      int    `form:"page"`
	PageSize  int    `form:"page_size"`
	SortBy    string `form:"sort_by"`
	RiskLabel string `form:"risk_label"`
}

type AttackerListResponse struct {
	Attackers []AttackerProfileSummary `json:"attackers"`
	Total     int64                    `json:"total"`
}

type AttackerProfileSummary struct {
	SourceIP          string    `json:"source_ip"`
	GeoCountry        string    `json:"geo_country"`
	TotalAttacks      int       `json:"total_attacks"`
	UniqueAttackTypes int       `json:"unique_attack_types"`
	TopAttackType     string    `json:"top_attack_type"`
	AttackPhase       string    `json:"attack_phase"`
	RiskScore         int       `json:"risk_score"`
	RiskLabel         string    `json:"risk_label"`
	LastSeen          time.Time `json:"last_seen"`
}

type AttackerProfileDetail struct {
	SourceIP          string         `json:"source_ip"`
	GeoCountry        string         `json:"geo_country"`
	GeoCity           string         `json:"geo_city,omitempty"`
	TotalAttacks      int            `json:"total_attacks"`
	UniqueAttackTypes int            `json:"unique_attack_types"`
	TopAttackType     string         `json:"top_attack_type"`
	UniqueTargetSites int            `json:"unique_target_sites"`
	ActiveHours       []int          `json:"active_hours"`
	BurstIntervals    []string       `json:"burst_intervals"`
	AttackPhase       string         `json:"attack_phase"`
	ToolsIdentified   string         `json:"tools_identified"`
	IsAutomated       bool           `json:"is_automated"`
	IsPersistent      bool           `json:"is_persistent"`
	RiskScore         int            `json:"risk_score"`
	RiskLabel         string         `json:"risk_label"`
	FirstSeen         time.Time      `json:"first_seen"`
	LastSeen          time.Time      `json:"last_seen"`
	RecentEvents      []LogEventItem `json:"recent_events"`
}

type LogEventItem struct {
	ID            string    `json:"id"`
	AttackType    string    `json:"attack_type"`
	Severity      string    `json:"severity"`
	Action        string    `json:"action"`
	SiteDomain    string    `json:"site_domain"`
	CorrelationID string    `json:"correlation_id"`
	Timestamp     time.Time `json:"timestamp"`
}

// === 趋势 ===

type TrendRequest struct {
	Duration string `form:"duration"`
}

type TrendResponse struct {
	Timeline        []TrendPoint `json:"timeline"`
	FrequentTypes   []CountItem  `json:"frequent_types"`
	ActiveAttackers int          `json:"active_attackers"`
	NewChains24h    int          `json:"new_chains_24h"`
}

type TrendPoint struct {
	Timestamp    int64 `json:"timestamp"`
	TotalEvents  int   `json:"total_events"`
	BlockedCount int   `json:"blocked_count"`
	DetectCount  int   `json:"detect_count"`
	UniqueIPs    int   `json:"unique_ips"`
}

// === 规则管理 ===

type SituationRuleRequest struct {
	Name           string `json:"name" binding:"required"`
	Stage          string `json:"stage" binding:"required"`
	LogQL          string `json:"logql" binding:"required"`
	Interval       int    `json:"interval_seconds" binding:"required"`
	Threshold      int    `json:"threshold" binding:"required"`
	Severity       string `json:"severity" binding:"required"`
	MITRETactic    string `json:"mitre_tactic"`
	MITRETechnique string `json:"mitre_technique"`
	Enabled        bool   `json:"enabled"`
}

// === 快速处置 ===

type QuickActionRequest struct {
	SourceIP      string `json:"source_ip" binding:"required"`
	Action        string `json:"action" binding:"required"`
	DurationHours int    `json:"duration_hours"`
	Reason        string `json:"reason" binding:"required"`
	CorrelationID string `json:"correlation_id"`
}

type QuickActionResponse struct {
	Success     bool   `json:"success"`
	SourceIP    string `json:"source_ip"`
	Action      string `json:"action"`
	Blocked     bool   `json:"blocked"`
	Blacklisted bool   `json:"blacklisted"`
	Note        string `json:"note"`
}
