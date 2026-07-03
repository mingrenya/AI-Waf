package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mingrenya/AI-Waf/pkg/model"
	"github.com/mingrenya/AI-Waf/server/dto"
	"github.com/mingrenya/AI-Waf/server/service/situation"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// --- Mock Repository ---

type mockSituationRepo struct {
	rules    []model.SituationRule
	chains   []model.AttackChain
	profiles []model.AttackerProfile
}

func (m *mockSituationRepo) ListRules(ctx context.Context) ([]model.SituationRule, error) {
	return m.rules, nil
}

func (m *mockSituationRepo) ListEnabledRules(ctx context.Context) ([]model.SituationRule, error) {
	var enabled []model.SituationRule
	for _, r := range m.rules {
		if r.Enabled {
			enabled = append(enabled, r)
		}
	}
	return enabled, nil
}

func (m *mockSituationRepo) FindRuleByName(ctx context.Context, name string) (*model.SituationRule, error) {
	for _, r := range m.rules {
		if r.Name == name {
			return &r, nil
		}
	}
	return nil, nil
}

func (m *mockSituationRepo) GetRuleByID(ctx context.Context, id string) (*model.SituationRule, error) {
	for _, r := range m.rules {
		if r.ID == id {
			return &r, nil
		}
	}
	return nil, nil
}

func (m *mockSituationRepo) CreateRule(ctx context.Context, rule *model.SituationRule) error {
	m.rules = append(m.rules, *rule)
	return nil
}

func (m *mockSituationRepo) UpdateRule(ctx context.Context, id string, rule *model.SituationRule) error {
	return nil
}

func (m *mockSituationRepo) DeleteRule(ctx context.Context, id string) error {
	return nil
}

func (m *mockSituationRepo) ListChains(ctx context.Context, filter bson.M, skip, limit int64) ([]model.AttackChain, int64, error) {
	return m.chains, int64(len(m.chains)), nil
}

func (m *mockSituationRepo) GetChainByID(ctx context.Context, id string) (*model.AttackChain, error) {
	for _, c := range m.chains {
		if c.ID == id {
			return &c, nil
		}
	}
	return nil, nil
}

func (m *mockSituationRepo) UpsertChain(ctx context.Context, chain *model.AttackChain) error {
	return nil
}

func (m *mockSituationRepo) GetChainByIP(ctx context.Context, ip string) (*model.AttackChain, error) {
	for _, c := range m.chains {
		if c.SourceIP == ip {
			return &c, nil
		}
	}
	return nil, nil
}

func (m *mockSituationRepo) UpsertProfile(ctx context.Context, profile *model.AttackerProfile) error {
	return nil
}

func (m *mockSituationRepo) GetProfile(ctx context.Context, ip string) (*model.AttackerProfile, error) {
	for _, p := range m.profiles {
		if p.SourceIP == ip {
			return &p, nil
		}
	}
	return nil, nil
}

func (m *mockSituationRepo) ListProfiles(ctx context.Context, sortBy string, skip, limit int64) ([]model.AttackerProfile, int64, error) {
	return m.profiles, int64(len(m.profiles)), nil
}

// --- Mock QuickActionService ---

type mockQuickActionSvc struct{}

func (m *mockQuickActionSvc) ExecuteQuickAction(ctx context.Context, req situation.QuickActionRequest) (*situation.QuickActionResult, error) {
	return &situation.QuickActionResult{
		Success:  true,
		SourceIP: req.SourceIP,
		Action:   req.Action,
		Blocked:  true,
		Note:     "test action executed",
	}, nil
}

// --- Helper ---

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func makeCtrl() *SituationControllerImpl {
	now := time.Now()
	return &SituationControllerImpl{
		repo: &mockSituationRepo{
			chains: []model.AttackChain{
				{
					ID:         "chain-1",
					SourceIP:   "192.168.1.100",
					GeoCountry: "CN",
					Stages: []model.ChainStage{
						{Stage: model.StageScanning, Technique: "T1595", DetectedAt: now},
						{Stage: model.StageExploitation, Technique: "T1190", DetectedAt: now},
					},
					RiskScore: 75,
					FirstSeen: now.Add(-2 * time.Hour),
					LastSeen:  now,
					Active:    true,
				},
				{
					ID:         "chain-2",
					SourceIP:   "10.0.0.55",
					GeoCountry: "US",
					Stages: []model.ChainStage{
						{Stage: model.StageReconnaissance, Technique: "T1046", DetectedAt: now},
					},
					RiskScore: 25,
					FirstSeen: now.Add(-5 * time.Hour),
					LastSeen:  now.Add(-1 * time.Hour),
					Active:    true,
				},
			},
			profiles: []model.AttackerProfile{
				{
					SourceIP:          "192.168.1.100",
					GeoCountry:        "CN",
					TotalAttacks:      150,
					UniqueAttackTypes: 4,
					TopAttackType:     "sql_injection",
					AttackPhase:       "exploitation",
					IsAutomated:       true,
					IsPersistent:      false,
					RiskScore:         75,
					RiskLabel:         "high",
					FirstSeen:         now.Add(-2 * time.Hour),
					LastSeen:          now,
				},
			},
			rules: []model.SituationRule{
				{
					ID: "rule-1", Name: "SQL注入检测", Stage: "exploitation",
					LogQL: `{attack_type="sql_injection"}`, Interval: 30, Threshold: 10,
					Severity: "critical", Enabled: true,
				},
			},
		},
		quickActionSvc: &situation.QuickActionService{},
	}
}

// --- Tests ---

func TestGetOverview_Returns200(t *testing.T) {
	ctrl := makeCtrl()
	r := setupTestRouter()
	r.GET("/overview", func(c *gin.Context) {
		c.Set("userID", "admin")
		ctrl.GetOverview(c)
	})

	req, _ := http.NewRequest("GET", "/overview", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Code int `json:"code"`
		Data dto.SituationOverviewResponse `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Data.ActiveChains != 2 {
		t.Errorf("expected 2 active chains, got %d", resp.Data.ActiveChains)
	}
	if resp.Data.TotalAttackers24h != 2 {
		t.Errorf("expected 2 attackers, got %d", resp.Data.TotalAttackers24h)
	}
}

func TestListChains_Returns200(t *testing.T) {
	ctrl := makeCtrl()
	r := setupTestRouter()
	r.GET("/chains", func(c *gin.Context) {
		c.Set("userID", "admin")
		ctrl.ListChains(c)
	})

	req, _ := http.NewRequest("GET", "/chains?page=1&page_size=10", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Data struct {
			Total int `json:"total"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Data.Total != 2 {
		t.Errorf("expected 2 chains, got %d", resp.Data.Total)
	}
}

func TestGetChainDetail_ReturnsAttackerProfile(t *testing.T) {
	ctrl := makeCtrl()
	r := setupTestRouter()
	r.GET("/chains/:id", func(c *gin.Context) {
		c.Set("userID", "admin")
		ctrl.GetChainDetail(c)
	})

	req, _ := http.NewRequest("GET", "/chains/chain-1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Data dto.ChainDetailResponse `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Data.SourceIP != "192.168.1.100" {
		t.Errorf("expected IP 192.168.1.100, got %s", resp.Data.SourceIP)
	}
	if resp.Data.AttackerProfile == nil {
		t.Error("expected attacker profile to be present")
	} else if resp.Data.AttackerProfile.GeoCountry != "CN" {
		t.Errorf("expected country CN, got %s", resp.Data.AttackerProfile.GeoCountry)
	}
}

func TestListAttackers_ReturnsSortedByRisk(t *testing.T) {
	ctrl := makeCtrl()
	r := setupTestRouter()
	r.GET("/attackers", func(c *gin.Context) {
		c.Set("userID", "admin")
		ctrl.ListAttackers(c)
	})

	req, _ := http.NewRequest("GET", "/attackers?sort_by=risk_score", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp struct {
		Data struct {
			Total int `json:"total"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Data.Total != 1 {
		t.Errorf("expected 1 attacker, got %d", resp.Data.Total)
	}
}

func TestGetAttackerProfile_ReturnsFullProfile(t *testing.T) {
	ctrl := makeCtrl()
	r := setupTestRouter()
	r.GET("/attackers/:ip", func(c *gin.Context) {
		c.Set("userID", "admin")
		ctrl.GetAttackerProfile(c)
	})

	req, _ := http.NewRequest("GET", "/attackers/192.168.1.100", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp struct {
		Data dto.AttackerProfileDetail `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Data.RiskScore != 75 {
		t.Errorf("expected risk_score 75, got %d", resp.Data.RiskScore)
	}
	if resp.Data.TopAttackType != "sql_injection" {
		t.Errorf("expected sql_injection, got %s", resp.Data.TopAttackType)
	}
}

func TestListRules_ReturnsRules(t *testing.T) {
	ctrl := makeCtrl()
	r := setupTestRouter()
	r.GET("/rules", func(c *gin.Context) {
		c.Set("userID", "admin")
		ctrl.ListRules(c)
	})

	req, _ := http.NewRequest("GET", "/rules", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestCreateRule_ValidInput(t *testing.T) {
	ctrl := makeCtrl()
	r := setupTestRouter()
	r.POST("/rules", func(c *gin.Context) {
		c.Set("userID", "admin")
		ctrl.CreateRule(c)
	})

	body := `{"name":"test","stage":"scanning","logql":"{}","interval_seconds":30,"threshold":10,"severity":"high"}`
	req, _ := http.NewRequest("POST", "/rules", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDeleteRule_ReturnsSuccess(t *testing.T) {
	ctrl := makeCtrl()
	r := setupTestRouter()
	r.DELETE("/rules/:id", func(c *gin.Context) {
		c.Set("userID", "admin")
		ctrl.DeleteRule(c)
	})

	req, _ := http.NewRequest("DELETE", "/rules/rule-1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}
