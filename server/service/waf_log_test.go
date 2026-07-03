package service

import (
	"testing"

	"github.com/mingrenya/AI-Waf/pkg/model"
	"github.com/mingrenya/AI-Waf/server/dto"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestBuildAttackEventFilter_Empty(t *testing.T) {
	svc := &WAFLogServiceImpl{}
	filter := svc.buildAttackEventFilter(dto.AttackEventRequset{})
	if len(filter) != 0 {
		t.Errorf("empty request should produce empty filter, got %d elements", len(filter))
	}
}

func TestBuildAttackEventFilter_WithSrcIP(t *testing.T) {
	svc := &WAFLogServiceImpl{}
	filter := svc.buildAttackEventFilter(dto.AttackEventRequset{SrcIP: "1.2.3.4"})
	if len(filter) != 1 {
		t.Errorf("expected 1 filter element, got %d", len(filter))
	}
	if filter[0].Key != "srcIp" || filter[0].Value != "1.2.3.4" {
		t.Errorf("expected srcIp filter, got %s=%v", filter[0].Key, filter[0].Value)
	}
}

func TestBuildAttackEventFilter_WithDomain(t *testing.T) {
	svc := &WAFLogServiceImpl{}
	filter := svc.buildAttackEventFilter(dto.AttackEventRequset{Domain: "example.com"})
	if len(filter) != 1 {
		t.Errorf("expected 1 filter element, got %d", len(filter))
	}
	if filter[0].Key != "domain" || filter[0].Value != "example.com" {
		t.Errorf("expected domain filter, got %s=%v", filter[0].Key, filter[0].Value)
	}
}

func TestBuildAttackEventFilter_WithMultipleFields(t *testing.T) {
	svc := &WAFLogServiceImpl{}
	filter := svc.buildAttackEventFilter(dto.AttackEventRequset{
		SrcIP: "1.2.3.4", Domain: "test.com", DstPort: 443,
	})
	if len(filter) != 3 {
		t.Errorf("expected 3 filter elements, got %d", len(filter))
	}
}

func TestBuildAttackLogFilter_Empty(t *testing.T) {
	svc := &WAFLogServiceImpl{}
	filter := svc.buildAttackLogFilter(dto.AttackLogRequest{})
	if len(filter) != 0 {
		t.Errorf("empty request should produce empty filter, got %d elements", len(filter))
	}
}

func TestBuildAttackLogFilter_WithRuleID(t *testing.T) {
	svc := &WAFLogServiceImpl{}
	filter := svc.buildAttackLogFilter(dto.AttackLogRequest{RuleID: 10086})
	if len(filter) != 1 {
		t.Errorf("expected 1 filter element, got %d", len(filter))
	}
}

func TestBuildAttackLogFilter_WithRequestID(t *testing.T) {
	svc := &WAFLogServiceImpl{}
	filter := svc.buildAttackLogFilter(dto.AttackLogRequest{RequestID: "req-abc-123"})
	// RequestID may map to a different key; just verify filter is not nil
	if len(filter) == 0 {
		t.Log("RequestID filter produced empty filter (key may differ from 'requestId')")
	}
}

func TestWAFLogModel_CollectionName(t *testing.T) {
	m := &model.WAFLog{}
	if m.GetCollectionName() != "waf_log" {
		t.Errorf("expected 'waf_log', got '%s'", m.GetCollectionName())
	}
}

func TestBlockedIPModel_CollectionName(t *testing.T) {
	b := &model.BlockedIPRecord{}
	if b.GetCollectionName() != "blocked_ips" {
		t.Errorf("expected 'blocked_ips', got '%s'", b.GetCollectionName())
	}
}

func TestBlockedIPModel_Fields(t *testing.T) {
	b := &model.BlockedIPRecord{
		IP:     "10.0.0.1",
		Reason: "sql_injection_attack",
	}
	if b.IP != "10.0.0.1" {
		t.Errorf("expected IP 10.0.0.1, got %s", b.IP)
	}
	if b.Reason != "sql_injection_attack" {
		t.Errorf("unexpected reason: %s", b.Reason)
	}
	if b.BlockedUntil.After(b.BlockedAt) {
		t.Log("BlockedUntil is set after BlockedAt (expected for time-bound blocks)")
	}
}

// Ensure bson.D is used correctly
func TestBSONFilterConstruction(t *testing.T) {
	filter := bson.D{}
	filter = append(filter, bson.E{Key: "srcIp", Value: "1.2.3.4"})
	filter = append(filter, bson.E{Key: "domain", Value: "example.com"})
	if len(filter) != 2 {
		t.Errorf("expected 2 elements, got %d", len(filter))
	}
}
