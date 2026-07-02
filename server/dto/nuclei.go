package dto

// ScanRequest 发起扫描请求
type ScanRequest struct {
	SiteID    string   `json:"site_id" binding:"required"`
	TargetURL string   `json:"target_url" binding:"required"`
	Templates []string `json:"templates"`
	Severity  string   `json:"severity"`
}

// ScanTaskResponse 扫描任务响应
type ScanTaskResponse struct {
	ID        string `json:"id"`
	SiteID    string `json:"site_id"`
	TargetURL string `json:"target_url"`
	Status    string `json:"status"`
	Findings  int    `json:"findings"`
	Total     int    `json:"total"`
	CreatedAt string `json:"created_at"`
}

// ScanConfigRequest 扫描配置
type ScanConfigRequest struct {
	TemplatesPath string `json:"templates_path"`
	Concurrency   int    `json:"concurrency"`
	RateLimit     int    `json:"rate_limit"`
}
