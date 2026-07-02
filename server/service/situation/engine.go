package situation

import (
	"context"
	"time"

	"github.com/mingrenya/AI-Waf/server/config"
	"github.com/rs/zerolog"
)

type Engine struct {
	ruleEngine   *RuleEngine
	chainBuilder *AttackChainBuilder
	profiler     *Profiler
	publisher    *Publisher
	logger       zerolog.Logger
	running      bool
	stopCh       chan struct{}
}

func NewEngine(
	ruleEngine *RuleEngine,
	chainBuilder *AttackChainBuilder,
	profiler *Profiler,
	publisher *Publisher,
) *Engine {
	return &Engine{
		ruleEngine:   ruleEngine,
		chainBuilder: chainBuilder,
		profiler:     profiler,
		publisher:    publisher,
		logger:       config.GetServiceLogger("situation-engine"),
		stopCh:       make(chan struct{}),
	}
}

func (e *Engine) Start(ctx context.Context, interval time.Duration) {
	if e.running {
		return
	}
	e.running = true
	e.logger.Info().Dur("interval", interval).Msg("Situation engine started")

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-e.stopCh:
				e.logger.Info().Msg("Situation engine stopped")
				return
			case <-ticker.C:
				e.runCycle(ctx)
			}
		}
	}()
}

func (e *Engine) Stop() {
	if e.running {
		e.stopCh <- struct{}{}
		e.running = false
	}
}

func (e *Engine) runCycle(ctx context.Context) {
	results, err := e.ruleEngine.EvaluateAll(ctx)
	if err != nil {
		e.logger.Error().Err(err).Msg("Rule evaluation failed")
		return
	}
	if len(results) == 0 {
		return
	}
	e.logger.Info().Int("triggered_rules", len(results)).Msg("Rules triggered")

	for _, result := range results {
		if err := e.chainBuilder.UpdateChainFromRuleResult(ctx, result); err != nil {
			e.logger.Warn().Err(err).Str("rule", result.Rule.Name).Msg("Chain update failed")
		}
	}

	processedIPs := make(map[string]bool)
	for _, result := range results {
		for _, ip := range result.HitIPs {
			if processedIPs[ip] {
				continue
			}
			processedIPs[ip] = true

			if err := e.profiler.SaveProfile(ctx, ip); err != nil {
				e.logger.Warn().Err(err).Str("ip", ip).Msg("Profile build failed")
				continue
			}

			profile, err := e.profiler.BuildProfile(ctx, ip)
			if err != nil || profile == nil {
				continue
			}

			score := CalculateRisk(profile)
			profile.RiskScore = score
			profile.RiskLabel = RiskLabel(score)

			if score >= 60 {
				e.publisher.PublishAlert(map[string]interface{}{
					"ip": ip, "score": score, "label": profile.RiskLabel,
					"phase": profile.AttackPhase, "country": profile.GeoCountry,
					"attack_type": profile.TopAttackType,
				})
			}
			e.publisher.PublishUpdate(map[string]interface{}{"profile": profile})
		}
	}
}
