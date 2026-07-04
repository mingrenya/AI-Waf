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

// ScanTaskDetailResponse 扫描任务详情响应(含 Findings 列表)
type ScanTaskDetailResponse struct {
	ID          string        `json:"id"`
	SiteID      string        `json:"site_id"`
	TargetURL   string        `json:"target_url"`
	Status      string        `json:"status"`
	Total       int           `json:"total"`
	Findings    []FindingItem `json:"findings"`
	CreatedAt   string        `json:"created_at"`
	StartedAt   string        `json:"started_at,omitempty"`
	CompletedAt string        `json:"completed_at,omitempty"`
}

// FindingItem 扫描发现项
type FindingItem struct {
	TemplateID       string   `json:"template_id"`
	Name             string   `json:"name"`
	Severity         string   `json:"severity"`
	MatchedAt        string   `json:"matched_at"`
	CurlCommand      string   `json:"curl_command"`
	ExtractedResults []string `json:"extracted_results"`
}

// ScanConfigRequest 扫描配置
type ScanConfigRequest struct {
	TemplatesPath string `json:"templates_path"`
	Concurrency   int    `json:"concurrency"`
	RateLimit     int    `json:"rate_limit"`
}
