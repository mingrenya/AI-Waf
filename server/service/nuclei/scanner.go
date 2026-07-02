package nuclei

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mingrenya/AI-Waf/server/config"
	nuclei "github.com/projectdiscovery/nuclei/v3/lib"
	"github.com/projectdiscovery/nuclei/v3/pkg/output"
	"github.com/rs/zerolog"
)

// ScanTask 扫描任务
type ScanTask struct {
	ID          string       `json:"id"`
	SiteID      string       `json:"site_id"`
	Target      string       `json:"target"`
	Templates   []string     `json:"templates"`
	Severity    string       `json:"severity"`
	Status      string       `json:"status"` // pending/running/completed/failed/cancelled
	Progress    ScanProgress `json:"progress"`
	CreatedAt   time.Time    `json:"created_at"`
	StartedAt   *time.Time   `json:"started_at,omitempty"`
	CompletedAt *time.Time   `json:"completed_at,omitempty"`
}

// ScanProgress 扫描进度
type ScanProgress struct {
	Total     int `json:"total"`
	Completed int `json:"completed"`
	Findings  int `json:"findings"`
}

// Scanner Nuclei 扫描器接口
type Scanner interface {
	StartScan(ctx context.Context, task ScanTask, callback func(*output.ResultEvent)) error
	CancelScan(taskID string) error
	GetTask(taskID string) (*ScanTask, error)
	ListTasks() []ScanTask
}

type scannerImpl struct {
	engine *nuclei.ThreadSafeNucleiEngine
	tasks  map[string]*ScanTask
	mu     sync.RWMutex
	logger zerolog.Logger
}

// NewScanner 创建 Nuclei 扫描器
func NewScanner(ctx context.Context, templatesPath string) (Scanner, error) {
	engine, err := nuclei.NewThreadSafeNucleiEngineCtx(ctx,
		nuclei.WithTemplateFilters(nuclei.TemplateFilters{}),
		nuclei.WithTemplatesOrWorkflows(nuclei.TemplateSources{
			Templates: []string{templatesPath},
		}),
		nuclei.WithConcurrency(nuclei.Concurrency{
			TemplateConcurrency: 25,
			HostConcurrency:     10,
		}),
		nuclei.WithGlobalRateLimitCtx(ctx, 150, time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("创建 Nuclei 引擎失败: %w", err)
	}

	if err := engine.GlobalLoadAllTemplates(); err != nil {
		return nil, fmt.Errorf("加载 Nuclei 模板失败: %w", err)
	}

	return &scannerImpl{
		engine: engine,
		tasks:  make(map[string]*ScanTask),
		logger: config.GetServiceLogger("nuclei-scanner"),
	}, nil
}

// StartScan 启动扫描任务
func (s *scannerImpl) StartScan(ctx context.Context, task ScanTask, callback func(*output.ResultEvent)) error {
	s.mu.Lock()
	taskCopy := task
	taskCopy.Status = "running"
	now := time.Now()
	taskCopy.StartedAt = &now
	s.tasks[task.ID] = &taskCopy
	s.mu.Unlock()

	go func() {
		s.engine.GlobalResultCallback(callback)

		err := s.engine.ExecuteNucleiWithOptsCtx(ctx, []string{task.Target},
			nuclei.WithTemplateFilters(nuclei.TemplateFilters{
				Severity: task.Severity,
			}),
		)

		s.mu.Lock()
		defer s.mu.Unlock()
		t, ok := s.tasks[task.ID]
		if !ok {
			return
		}
		completedAt := time.Now()
		t.CompletedAt = &completedAt
		if err != nil {
			t.Status = "failed"
			s.logger.Error().Err(err).Str("task_id", task.ID).Msg("Scan failed")
		} else {
			t.Status = "completed"
		}
	}()

	return nil
}

// CancelScan 取消扫描
func (s *scannerImpl) CancelScan(taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	task, ok := s.tasks[taskID]
	if !ok {
		return fmt.Errorf("任务不存在: %s", taskID)
	}
	task.Status = "cancelled"
	now := time.Now()
	task.CompletedAt = &now
	return nil
}

// GetTask 获取任务
func (s *scannerImpl) GetTask(taskID string) (*ScanTask, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	task, ok := s.tasks[taskID]
	if !ok {
		return nil, fmt.Errorf("任务不存在: %s", taskID)
	}
	return task, nil
}

// ListTasks 列出所有任务
func (s *scannerImpl) ListTasks() []ScanTask {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tasks := make([]ScanTask, 0, len(s.tasks))
	for _, t := range s.tasks {
		tasks = append(tasks, *t)
	}
	return tasks
}
