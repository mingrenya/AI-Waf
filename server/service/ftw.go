package service

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/mingrenya/AI-Waf/pkg/model"
	c "github.com/mingrenya/AI-Waf/server/config"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"gopkg.in/yaml.v3"
)

// FTWTestService WAF 回归测试服务（基于 Go-FTW YAML 格式）
type FTWTestService interface {
	RunTests(ctx context.Context, targetURL string, testDir string) (*FTWReport, error)
	GetReports(ctx context.Context, limit int) ([]*FTWReport, error)
	GetTestFiles() []string
}

// FTWResult 单个测试结果
type FTWResult struct {
	TestID     string  `json:"testId" bson:"testId"`
	Title      string  `json:"title" bson:"title"`
	Passed     bool    `json:"passed" bson:"passed"`
	StatusCode int     `json:"statusCode" bson:"statusCode"`
	WAFHit     bool    `json:"wafHit" bson:"wafHit"`
	WAFRuleID  int     `json:"wafRuleId" bson:"wafRuleId"`
	WAFMessage string  `json:"wafMessage" bson:"wafMessage"`
	Category   string  `json:"category" bson:"category"`
	Error      string  `json:"error,omitempty" bson:"error,omitempty"`
	DurationMs float64 `json:"durationMs" bson:"durationMs"`
}

// FTWReport WAF 回归测试报告
type FTWReport struct {
	ID          string      `json:"id" bson:"_id,omitempty"`
	TargetURL   string      `json:"targetUrl" bson:"targetUrl"`
	TotalTests  int         `json:"totalTests" bson:"totalTests"`
	Passed      int         `json:"passed" bson:"passed"`
	Failed      int         `json:"failed" bson:"failed"`
	FalseNegs   int         `json:"falseNegs" bson:"falseNegs"`
	FalsePoss   int         `json:"falsePoss" bson:"falsePoss"`
	BlockRate   float64     `json:"blockRate" bson:"blockRate"`
	Results     []FTWResult `json:"results" bson:"results"`
	CreatedAt   time.Time   `json:"createdAt" bson:"createdAt"`
	DurationSec float64     `json:"durationSec" bson:"durationSec"`
}

func (FTWReport) GetCollectionName() string { return "ftw_test_reports" }

// FTWTestFile Go-FTW YAML 文件结构
type FTWTestFile struct {
	Meta  FTWMeta    `yaml:"meta"`
	Tests []FTWTestCase `yaml:"tests"`
}
type FTWMeta struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Enabled     bool   `yaml:"enabled"`
}
type FTWTestCase struct {
	TestID    uint   `yaml:"test_id"`
	TestTitle string `yaml:"test_title"`
	Desc      string `yaml:"desc"`
	Stages    []FTWYAMLStage `yaml:"stages"`
}
type FTWYAMLStage struct {
	Input  FTWYAMLInput  `yaml:"input"`
	Output FTWYAMLOutput `yaml:"output"`
}
type FTWYAMLInput struct {
	DestAddr string            `yaml:"dest_addr"`
	URI      string            `yaml:"uri"`
	Method   string            `yaml:"method"`
	Version  string            `yaml:"version"`
	Headers  map[string]string `yaml:"headers"`
	Data     string            `yaml:"data"`
}
type FTWYAMLOutput struct {
	ExpectResponseCode int    `yaml:"expect_response_code"`
	LogContains        string `yaml:"log_contains"`
	NoLogContains      string `yaml:"no_log_contains"`
}

type ftwTestServiceImpl struct {
	logger  zerolog.Logger
	mongoDB *mongo.Database
	testDir string
	client  *http.Client
}

// NewFTWTestService 创建 FTW 测试服务
func NewFTWTestService(db *mongo.Database) FTWTestService {
	return &ftwTestServiceImpl{
		logger:  c.GetServiceLogger("ftw-test"),
		mongoDB: db,
		testDir: "server/public/ftw-tests",
		client: &http.Client{
			Timeout: 15 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

// RunTests loads YAML tests, executes them, cross-checks waf_log, returns report
func (s *ftwTestServiceImpl) RunTests(ctx context.Context, targetURL string, testDir string) (*FTWReport, error) {
	if targetURL == "" {
		targetURL = "http://localhost:8080"
	}
	if testDir != "" {
		s.testDir = testDir
	}

	s.logger.Info().Str("target", targetURL).Str("dir", s.testDir).Msg("开始 FTW 回归测试")

	files, err := filepath.Glob(filepath.Join(s.testDir, "*.yaml"))
	if err != nil || len(files) == 0 {
		err2 := filepath.Walk(s.testDir, func(path string, info os.FileInfo, err error) error {
			if err != nil { return err }
			if strings.HasSuffix(path, ".yaml") {
				files = append(files, path)
			}
			return nil
		})
		if err2 != nil {
			return nil, fmt.Errorf("未找到 YAML 测试文件于 %s: %w", s.testDir, err)
		}
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("目录 %s 中未找到 YAML 测试文件", s.testDir)
	}

	s.logger.Info().Int("files", len(files)).Msg("已找到测试文件")

	// Load test cases
	var allCases []FTWTestCase
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			s.logger.Warn().Err(err).Str("file", f).Msg("读取测试文件失败")
			continue
		}
		var ft FTWTestFile
		if err := yaml.Unmarshal(data, &ft); err != nil {
			s.logger.Warn().Err(err).Str("file", f).Msg("解析 YAML 失败")
			continue
		}
		for i := range ft.Tests {
			s.logger.Debug().Str("id", fmt.Sprintf("%d", ft.Tests[i].TestID)).Str("title", ft.Tests[i].TestTitle).Msg("Test case loaded")
		}
		allCases = append(allCases, ft.Tests...)
	}

	if len(allCases) == 0 {
		return nil, fmt.Errorf("未找到任何测试用例")
	}

	report := &FTWReport{
		TargetURL:  targetURL,
		TotalTests: len(allCases),
		Results:    make([]FTWResult, 0, len(allCases)),
		CreatedAt:  time.Now(),
	}
	startTime := time.Now()

	// Execute tests concurrently
	sem := make(chan struct{}, 5)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, tc := range allCases {
		wg.Add(1)
		go func(test FTWTestCase) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			r := s.executeTest(ctx, targetURL, test)
			mu.Lock()
			report.Results = append(report.Results, r)
			mu.Unlock()
		}(tc)
	}
	wg.Wait()

	report.DurationSec = time.Since(startTime).Seconds()

	// Compute statistics
	for _, r := range report.Results {
		if r.Passed {
			report.Passed++
		} else {
			report.Failed++
			if !r.WAFHit {
				report.FalseNegs++
			}
		}
	}
	if report.TotalTests > 0 {
		report.BlockRate = float64(report.Passed) / float64(report.TotalTests) * 100
	}

	s.saveReport(ctx, report)

	s.logger.Info().
		Int("total", report.TotalTests).
		Int("passed", report.Passed).
		Int("failed", report.Failed).
		Float64("blockRate", report.BlockRate).
		Float64("duration", report.DurationSec).
		Msg("FTW 回归测试完成")

	return report, nil
}

// executeTest runs a single FTW test case
func (s *ftwTestServiceImpl) executeTest(ctx context.Context, targetURL string, test FTWTestCase) FTWResult {
	r := FTWResult{
		TestID:   fmt.Sprintf("%d", test.TestID),
		Title:    test.TestTitle,
		Passed:   true,
		Category: test.Desc,
	}

	if len(test.Stages) == 0 {
		return r
	}

	stage := test.Stages[0]
	method := stage.Input.Method
	if method == "" { method = "GET" }
	uri := stage.Input.URI
	if uri == "" { uri = "/" }

	fullURL := strings.TrimRight(targetURL, "/") + "/" + strings.TrimLeft(uri, "/")

	var body io.Reader
	if stage.Input.Data != "" {
		body = bytes.NewBufferString(stage.Input.Data)
	}

	startTime := time.Now()
	req, err := http.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		r.Passed = false
		r.Error = err.Error()
		r.DurationMs = float64(time.Since(startTime).Microseconds()) / 1000.0
		return r
	}

	req.Header.Set("User-Agent", "AI-WAF-FTW-Test/2.0")
	for k, v := range stage.Input.Headers {
		req.Header.Set(k, v)
	}

	resp, err := s.client.Do(req)
	r.DurationMs = float64(time.Since(startTime).Microseconds()) / 1000.0
	if err != nil {
		r.Passed = false
		r.Error = err.Error()
		return r
	}
	defer resp.Body.Close()

	r.StatusCode = resp.StatusCode

	expectedCode := stage.Output.ExpectResponseCode
	if expectedCode == 0 {
		expectedCode = 403
	}

	if resp.StatusCode != expectedCode {
		r.Passed = false
	}

	if expectedCode == 403 {
		r.WAFHit, r.WAFRuleID, r.WAFMessage = s.lookupWAFLog(ctx)
		if !r.WAFHit {
			r.Passed = false
		}
	}

	return r
}

// lookupWAFLog checks for recent WAF log entries
func (s *ftwTestServiceImpl) lookupWAFLog(ctx context.Context) (bool, int, string) {
	if s.mongoDB == nil {
		return false, 0, ""
	}
	qctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var log model.WAFLog
	err := s.mongoDB.Collection("waf_log").FindOne(
		qctx,
		bson.M{
			"createdAt": bson.M{"$gte": time.Now().Add(-30 * time.Second)},
			"severity":  bson.M{"$gt": 0},
		},
		options.FindOne().SetSort(bson.M{"createdAt": -1}),
	).Decode(&log)
	if err != nil {
		return false, 0, ""
	}
	return true, log.RuleID, log.Message
}

// GetReports returns historical test reports
func (s *ftwTestServiceImpl) GetReports(ctx context.Context, limit int) ([]*FTWReport, error) {
	if s.mongoDB == nil { return nil, nil }
	if limit <= 0 { limit = 20 }
	qctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cursor, err := s.mongoDB.Collection("ftw_test_reports").Find(
		qctx, bson.M{},
		options.Find().SetLimit(int64(limit)).SetSort(bson.M{"createdAt": -1}),
	)
	if err != nil { return nil, err }
	defer cursor.Close(qctx)

	var reports []*FTWReport
	if err := cursor.All(qctx, &reports); err != nil {
		return nil, err
	}
	return reports, nil
}

// GetTestFiles returns list of test files in testDir
func (s *ftwTestServiceImpl) GetTestFiles() []string {
	entries, err := os.ReadDir(s.testDir)
	if err != nil { return nil }
	var files []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".yaml") {
			files = append(files, e.Name())
		}
	}
	return files
}

// saveReport writes report to MongoDB
func (s *ftwTestServiceImpl) saveReport(ctx context.Context, report *FTWReport) {
	if s.mongoDB == nil { return }
	sctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_, err := s.mongoDB.Collection("ftw_test_reports").InsertOne(sctx, report)
	if err != nil {
		s.logger.Error().Err(err).Msg("保存 FTW 测试报告失败")
	}
}
