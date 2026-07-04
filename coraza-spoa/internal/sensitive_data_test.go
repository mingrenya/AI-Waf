package internal

import (
	"testing"
)

func TestMaskIDCard(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"110101199001011234", "11************1234"},
		{"33010619850615278X", "33************278X"},
		{"12345", "*****"},
	}
	for _, tt := range tests {
		got := maskIDCard(tt.input)
		if got != tt.expected {
			t.Errorf("maskIDCard(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestMaskPhone(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"13812345678", "138****5678"},
		{"15900001111", "159****1111"},
		{"12345", "*****"},
	}
	for _, tt := range tests {
		got := maskPhone(tt.input)
		if got != tt.expected {
			t.Errorf("maskPhone(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestMaskBankCard(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"6222021234567890123", "6222***********0123"},
		{"1234567890123456", "1234********3456"},
		{"12345", "*****"},
	}
	for _, tt := range tests {
		got := maskBankCard(tt.input)
		if got != tt.expected {
			t.Errorf("maskBankCard(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestSensitiveDataFilter_FilterResponse(t *testing.T) {
	f := NewSensitiveDataFilter()
	f.Enable()

	body := []byte(`{"user":{"name":"测试","id_card":"110101199001011234","phone":"13812345678","bank":"6222021234567890123"}}`)
	masked, matches := f.Filter(body)

	if len(matches) == 0 {
		t.Fatal("expected at least one sensitive match")
	}

	maskedStr := string(masked)
	if maskedStr == string(body) {
		t.Error("response body was not masked")
	}

	// 确认原始敏感数据已被脱敏
	if contains(maskedStr, "110101199001011234") {
		t.Error("id_card was not masked")
	}
	if contains(maskedStr, "13812345678") {
		t.Error("phone was not masked")
	}
	if contains(maskedStr, "6222021234567890123") {
		t.Error("bank_card was not masked")
	}
}

func TestSensitiveDataFilter_Disabled(t *testing.T) {
	f := NewSensitiveDataFilter()
	// 不调用 Enable
	body := []byte(`id_card=110101199001011234`)
	masked, matches := f.Filter(body)

	if string(masked) != string(body) {
		t.Error("disabled filter should not mask data")
	}
	if matches != nil {
		t.Error("disabled filter should return nil matches")
	}
}

func TestSensitiveDataFilter_HasSensitiveData(t *testing.T) {
	f := NewSensitiveDataFilter()
	f.Enable()

	if f.HasSensitiveData([]byte(`{"phone":"13812345678"}`)) == false {
		t.Error("should detect phone number")
	}
	if f.HasSensitiveData([]byte(`{"id_card":"110101199001011234"}`)) == false {
		t.Error("should detect id card")
	}
	if f.HasSensitiveData([]byte(`{"name":"test"}`)) == true {
		t.Error("should not detect plain text")
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
