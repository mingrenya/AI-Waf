package service

import (
	"testing"
)

func TestRuleErrorSentinelValues(t *testing.T) {
	errors := []struct {
		name string
		err  error
	}{
		{"ErrMicroRuleNotFound", ErrMicroRuleNotFound},
		{"ErrMicroRuleNameExists", ErrMicroRuleNameExists},
		{"ErrSystemRuleNoMod", ErrSystemRuleNoMod},
		{"ErrSystemRuleNoDelete", ErrSystemRuleNoDelete},
	}
	for _, e := range errors {
		if e.err == nil {
			t.Errorf("%s is nil", e.name)
		}
	}
}

func TestSystemDefaultIPBlockRuleConstant(t *testing.T) {
	if SystemDefaultIPBlockRule != "system_default_ip_block" {
		t.Errorf("expected 'system_default_ip_block', got '%s'", SystemDefaultIPBlockRule)
	}
}

func TestBlockedIPErrorSentinelValues(t *testing.T) {
	if ErrBlockedIPNotFound == nil {
		t.Error("ErrBlockedIPNotFound is nil")
	}
	if ErrInvalidPageSize == nil {
		t.Error("ErrInvalidPageSize is nil")
	}
}
