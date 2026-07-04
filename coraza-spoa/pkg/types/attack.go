package types

// AttackInfo 攻击事件摘要 — 攻击检测时传递给回调的上下文信息
type AttackInfo struct {
	SourceIP   string
	TargetHost string
	TargetPort int
	AttackType string // "coraza" | "micro_engine" | "scan_protection"
	Severity   string
	RuleID     int
	RequestID  string
}

// AttackCallback 攻击事件回调签名，用于取证捕获、告警推送等
type AttackCallback func(info AttackInfo)

