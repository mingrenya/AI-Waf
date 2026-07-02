package nuclei

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mingrenya/AI-Waf/server/config"
	"github.com/rs/zerolog"
)

// TemplateInfo 模板信息
type TemplateInfo struct {
	Path string `json:"path"`
	Name string `json:"name"`
}

// TemplateManager 模板管理器
type TemplateManager struct {
	path   string
	logger zerolog.Logger
}

// NewTemplateManager 创建模板管理器
func NewTemplateManager(path string) *TemplateManager {
	return &TemplateManager{
		path:   path,
		logger: config.GetServiceLogger("nuclei-template"),
	}
}

// ListTemplates 列出可用模板
func (m *TemplateManager) ListTemplates(_ context.Context) ([]TemplateInfo, error) {
	var templates []TemplateInfo

	err := filepath.Walk(m.path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".yaml") {
			return nil
		}

		relPath, _ := filepath.Rel(m.path, path)
		templates = append(templates, TemplateInfo{
			Path: relPath,
			Name: strings.TrimSuffix(info.Name(), ".yaml"),
		})
		return nil
	})

	return templates, err
}

// GetPath 返回模板目录路径
func (m *TemplateManager) GetPath() string {
	return m.path
}

// EnsureTemplatesDir 确保模板目录存在
func (m *TemplateManager) EnsureTemplatesDir() error {
	if _, err := os.Stat(m.path); os.IsNotExist(err) {
		return fmt.Errorf("模板目录不存在: %s（请先执行 nuclei -ut 下载模板或挂载 nuclei-templates 目录）", m.path)
	}
	return nil
}
