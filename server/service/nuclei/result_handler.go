package nuclei

import (
	"github.com/mingrenya/AI-Waf/server/config"
	"github.com/projectdiscovery/nuclei/v3/pkg/output"
	"github.com/rs/zerolog"
)

// Finding 扫描发现
type Finding struct {
	TemplateID       string   `json:"template_id"`
	Name             string   `json:"name"`
	Severity         string   `json:"severity"`
	MatchedAt        string   `json:"matched_at"`
	CurlCommand      string   `json:"curl_command"`
	ExtractedResults []string `json:"extracted_results"`
	Reference        string   `json:"reference"`
}

// ResultHandler 扫描结果处理器
type ResultHandler struct {
	findings []Finding
	logger   zerolog.Logger
}

// NewResultHandler 创建结果处理器
func NewResultHandler() *ResultHandler {
	return &ResultHandler{
		findings: make([]Finding, 0),
		logger:   config.GetServiceLogger("nuclei-result"),
	}
}

// HandleResult 处理单个扫描结果（用作 GlobalResultCallback）
func (h *ResultHandler) HandleResult(event *output.ResultEvent) {
	if event == nil {
		return
	}

	finding := Finding{
		TemplateID:       event.TemplateID,
		Name:             event.Info.Name,
		Severity:         event.Info.SeverityHolder.Severity.String(),
		MatchedAt:        event.Matched,
		CurlCommand:      event.CURLCommand,
		ExtractedResults: event.ExtractedResults,
	}

	h.findings = append(h.findings, finding)
	h.logger.Debug().Str("template", event.TemplateID).Str("host", event.Host).Msg("Finding recorded")
}

// GetFindings 获取所有发现
func (h *ResultHandler) GetFindings() []Finding {
	return h.findings
}

// Reset 重置处理器
func (h *ResultHandler) Reset() {
	h.findings = make([]Finding, 0)
}
