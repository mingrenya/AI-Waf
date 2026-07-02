package situation

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mingrenya/AI-Waf/pkg/model"
	"github.com/mingrenya/AI-Waf/server/config"
	"github.com/mingrenya/AI-Waf/server/repository"
	"github.com/mingrenya/AI-Waf/server/service"
	"github.com/rs/zerolog"
)

type Profiler struct {
	loki   service.LokiLogService
	repo   repository.SituationRepository
	logger zerolog.Logger
}

func NewProfiler(loki service.LokiLogService, repo repository.SituationRepository) *Profiler {
	return &Profiler{
		loki:   loki,
		repo:   repo,
		logger: config.GetServiceLogger("profiler"),
	}
}

// BuildProfile 构建攻击者画像
func (p *Profiler) BuildProfile(ctx context.Context, ip string) (*model.AttackerProfile, error) {
	now := time.Now()
	start := now.Add(-24 * time.Hour)

	resp, err := p.loki.QueryRange(ctx, service.LokiRangeRequest{
		Query: fmt.Sprintf(`{container_name="mrya-waf",source_ip="%s"} | json`, ip),
		Start: fmt.Sprintf("%d", start.Unix()),
		End:   fmt.Sprintf("%d", now.Unix()),
		Step:  "5m",
		Limit: 500,
	})
	if err != nil {
		return nil, err
	}

	entries := service.ToLogEntries(resp)

	profile := &model.AttackerProfile{
		SourceIP:  ip,
		FirstSeen: p.findFirstTime(entries),
		LastSeen:  p.findLastTime(entries),
		UpdatedAt: now,
	}

	typeCount := make(map[string]int)
	siteCount := make(map[string]bool)
	hourCount := make([]int, 24)
	for _, e := range entries.Results {
		if at, ok := e.Labels["attack_type"]; ok && at != "" {
			typeCount[at]++
		}
		if site, ok := e.Labels["site_id"]; ok && site != "" {
			siteCount[site] = true
		}
		if ts, err := strconv.ParseInt(e.Timestamp, 10, 64); err == nil {
			h := time.Unix(0, ts).Hour()
			if h >= 0 && h < 24 {
				hourCount[h]++
			}
		}
	}
	profile.TotalAttacks = entries.TotalHits
	profile.UniqueAttackTypes = len(typeCount)
	profile.UniqueTargetSites = len(siteCount)
	profile.GeoCountry = p.extractFirstLabel(entries, "geo_country")

	maxCount := 0
	for at, c := range typeCount {
		if c > maxCount {
			maxCount = c
			profile.TopAttackType = at
		}
	}

	profile.ActiveHours = make([]int, 0)
	for h, c := range hourCount {
		if c > 0 {
			profile.ActiveHours = append(profile.ActiveHours, h)
		}
	}

	profile.IsAutomated = profile.TotalAttacks > 50 && profile.UniqueAttackTypes > 2
	profile.IsPersistent = profile.LastSeen.Sub(profile.FirstSeen) > 24*time.Hour

	chain, _ := p.repo.GetChainByIP(ctx, ip)
	if chain != nil {
		var highest model.AttackStage
		for _, s := range chain.Stages {
			if s.Stage.Order() > highest.Order() {
				highest = s.Stage
			}
		}
		profile.AttackPhase = string(highest)
	} else {
		profile.AttackPhase = string(model.StageUnknown)
	}

	profile.ToolsIdentified = p.detectTools(entries)
	return profile, nil
}

// SaveProfile 构建并持久化画像
func (p *Profiler) SaveProfile(ctx context.Context, ip string) error {
	profile, err := p.BuildProfile(ctx, ip)
	if err != nil {
		return err
	}
	if profile == nil {
		return nil
	}
	profile.UpdatedAt = time.Now()
	return p.repo.UpsertProfile(ctx, profile)
}

func (p *Profiler) findFirstTime(entries *service.LokiLogQueryResponse) time.Time {
	if len(entries.Results) > 0 {
		return time.Now().Add(-24 * time.Hour)
	}
	return time.Now()
}

func (p *Profiler) findLastTime(entries *service.LokiLogQueryResponse) time.Time {
	if len(entries.Results) > 0 {
		return time.Now()
	}
	return time.Now()
}

func (p *Profiler) extractFirstLabel(entries *service.LokiLogQueryResponse, key string) string {
	for _, e := range entries.Results {
		if v, ok := e.Labels[key]; ok && v != "" {
			return v
		}
	}
	return ""
}

func (p *Profiler) detectTools(entries *service.LokiLogQueryResponse) string {
	tools := []string{"nuclei", "nikto", "nessus", "openvas", "acunetix", "burpsuite", "sqlmap", "nmap"}
	found := make([]string, 0)
	for _, e := range entries.Results {
		msg := strings.ToLower(e.Message)
		for _, tool := range tools {
			if strings.Contains(msg, tool) {
				found = append(found, tool)
			}
		}
	}
	if len(found) > 0 {
		return strings.Join(uniqueStrings(found), ", ")
	}
	return ""
}

func uniqueStrings(slice []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, s := range slice {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}
