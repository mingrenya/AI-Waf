package situation

import (
	"context"
	"fmt"
	"time"

	"github.com/mingrenya/AI-Waf/pkg/model"
	"github.com/mingrenya/AI-Waf/server/config"
	"github.com/mingrenya/AI-Waf/server/repository"
	"github.com/mingrenya/AI-Waf/server/service"
	"github.com/rs/zerolog"
)

type AttackChainBuilder struct {
	loki   service.LokiLogService
	repo   repository.SituationRepository
	logger zerolog.Logger
}

func NewAttackChainBuilder(loki service.LokiLogService, repo repository.SituationRepository) *AttackChainBuilder {
	return &AttackChainBuilder{
		loki:   loki,
		repo:   repo,
		logger: config.GetServiceLogger("attack-chain"),
	}
}

// BuildChain 构建指定 IP 的完整攻击链
func (b *AttackChainBuilder) BuildChain(ctx context.Context, ip string) (*model.AttackChain, error) {
	resp, err := b.loki.QueryRange(ctx, service.LokiRangeRequest{
		Query: fmt.Sprintf(`{container_name="mrya-waf",source_ip="%s"} | json`, ip),
		Start: fmt.Sprintf("%d", time.Now().Add(-24*time.Hour).Unix()),
		End:   fmt.Sprintf("%d", time.Now().Unix()),
		Step:  "1m",
		Limit: 500,
	})
	if err != nil {
		return nil, fmt.Errorf("Loki查询失败: %w", err)
	}

	entries := service.ToLogEntries(resp)
	if entries.TotalHits == 0 {
		return nil, nil
	}

	stages := b.buildStages(entries)
	correlationIDs := extractCorrelationIDs(entries)

	chain := &model.AttackChain{
		SourceIP:       ip,
		GeoCountry:     extractFirstGeo(entries),
		Stages:         stages,
		CorrelationIDs: correlationIDs,
		FirstSeen:      parseFirstTimestamp(entries),
		LastSeen:       parseLastTimestamp(entries),
		Active:         true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	return chain, nil
}

// UpdateChainFromRuleResult 根据规则命中更新/创建攻击链
func (b *AttackChainBuilder) UpdateChainFromRuleResult(ctx context.Context, result RuleResult) error {
	rule := result.Rule
	stage := model.AttackStage(rule.Stage)
	if mapped, ok := model.StageMapping[rule.Stage]; ok {
		stage = mapped
	}

	for _, ip := range result.HitIPs {
		existing, _ := b.repo.GetChainByIP(ctx, ip)

		if existing == nil {
			chain := &model.AttackChain{
				SourceIP: ip,
				Stages: []model.ChainStage{{
					Stage:      stage,
					Technique:  rule.MITRETechnique,
					DetectedAt: result.Timestamp,
					Confidence: float64(result.HitCount) / float64(rule.Threshold+1),
				}},
				FirstSeen: result.Timestamp,
				LastSeen:  result.Timestamp,
				Active:    true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			if err := b.repo.UpsertChain(ctx, chain); err != nil {
				b.logger.Error().Err(err).Str("ip", ip).Msg("创建攻击链失败")
			}
		} else {
			newStage := b.shouldAdvanceStage(existing, stage)
			existing.Stages = append(existing.Stages, model.ChainStage{
				Stage:      newStage,
				Technique:  rule.MITRETechnique,
				DetectedAt: result.Timestamp,
				Confidence: float64(result.HitCount) / float64(rule.Threshold+1),
			})
			existing.LastSeen = result.Timestamp
			existing.UpdatedAt = time.Now()
			if err := b.repo.UpsertChain(ctx, existing); err != nil {
				b.logger.Error().Err(err).Str("ip", ip).Msg("更新攻击链失败")
			}
		}
	}
	return nil
}

func (b *AttackChainBuilder) shouldAdvanceStage(existing *model.AttackChain, detectedStage model.AttackStage) model.AttackStage {
	highest := b.highestStage(existing)
	if detectedStage.Order() > highest.Order() {
		return detectedStage
	}
	return highest
}

func (b *AttackChainBuilder) highestStage(chain *model.AttackChain) model.AttackStage {
	var highest model.AttackStage
	for _, s := range chain.Stages {
		if s.Stage.Order() > highest.Order() {
			highest = s.Stage
		}
	}
	return highest
}

func (b *AttackChainBuilder) buildStages(entries *service.LokiLogQueryResponse) []model.ChainStage {
	stageMap := make(map[model.AttackStage]*model.ChainStage)
	for _, entry := range entries.Results {
		attackType := entry.Labels["attack_type"]
		stage, ok := model.StageMapping[attackType]
		if !ok {
			stage = model.StageUnknown
		}
		if existing, ok := stageMap[stage]; ok {
			existing.Confidence += 0.1
		} else {
			stageMap[stage] = &model.ChainStage{
				Stage:      stage,
				Confidence: 0.5,
			}
		}
	}
	stages := make([]model.ChainStage, 0, len(stageMap))
	for _, s := range stageMap {
		if s.Confidence > 1.0 {
			s.Confidence = 1.0
		}
		stages = append(stages, *s)
	}
	return stages
}

func extractCorrelationIDs(entries *service.LokiLogQueryResponse) []string {
	seen := make(map[string]bool)
	var ids []string
	for _, e := range entries.Results {
		if cid, ok := e.Labels["correlation_id"]; ok && cid != "" && !seen[cid] {
			seen[cid] = true
			ids = append(ids, cid)
		}
	}
	return ids
}

func extractFirstGeo(entries *service.LokiLogQueryResponse) string {
	for _, e := range entries.Results {
		if geo, ok := e.Labels["geo_country"]; ok && geo != "" {
			return geo
		}
	}
	return "unknown"
}

func parseFirstTimestamp(entries *service.LokiLogQueryResponse) time.Time {
	if len(entries.Results) > 0 {
		return time.Now().Add(-24 * time.Hour)
	}
	return time.Now()
}

func parseLastTimestamp(entries *service.LokiLogQueryResponse) time.Time {
	if len(entries.Results) > 0 {
		return time.Now()
	}
	return time.Now()
}
