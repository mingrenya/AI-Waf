package dto

// StartCaptureRequest 发起流量捕获请求
type StartCaptureRequest struct {
	Interface    string `json:"interface"`
	BPFFilter    string `json:"bpf_filter"`
	MaxPackets   int    `json:"max_packets"`
	DurationSecs int    `json:"duration_secs"`
	Description  string `json:"description"`
}

// CaptureSessionResponse 捕获会话响应
type CaptureSessionResponse struct {
	ID           string `json:"id"`
	Interface    string `json:"interface"`
	BPFFilter    string `json:"bpf_filter"`
	Status       string `json:"status"`
	MaxPackets   int    `json:"max_packets"`
	DurationSecs int    `json:"duration_secs"`
	Description  string `json:"description"`
	PacketCount  int    `json:"packet_count"`
	FileSize     int64  `json:"file_size"`
	FilePath     string `json:"file_path"`
	CreatedAt    string `json:"created_at"`
	StartedAt    string `json:"started_at,omitempty"`
	StoppedAt    string `json:"stopped_at,omitempty"`
	ErrorMsg     string `json:"error_msg,omitempty"`
}

// CaptureListResponse 捕获会话列表响应
type CaptureListResponse struct {
	Sessions []CaptureSessionResponse `json:"sessions"`
	Total    int64                    `json:"total"`
}
