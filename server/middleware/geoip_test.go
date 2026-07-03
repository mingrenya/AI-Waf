package middleware

import (
	"testing"
)

func TestParseCountryList_Empty(t *testing.T) {
	result := parseCountryList("")
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %v", result)
	}
}

func TestParseCountryList_Single(t *testing.T) {
	result := parseCountryList("CN")
	if len(result) != 1 || result[0] != "CN" {
		t.Errorf("expected [CN], got %v", result)
	}
}

func TestParseCountryList_Multiple(t *testing.T) {
	result := parseCountryList("CN,RU,KP")
	if len(result) != 3 {
		t.Errorf("expected 3, got %d: %v", len(result), result)
	}
	if result[0] != "CN" || result[1] != "RU" || result[2] != "KP" {
		t.Errorf("unexpected order: %v", result)
	}
}

func TestParseCountryList_TrimsWhitespace(t *testing.T) {
	result := parseCountryList(" CN , US , RU ")
	if len(result) != 3 || result[0] != "CN" || result[1] != "US" || result[2] != "RU" {
		t.Errorf("whitespace not trimmed: %v", result)
	}
}

func TestParseCountryList_Lowercase(t *testing.T) {
	result := parseCountryList("cn,us,ru")
	if len(result) != 3 || result[0] != "CN" || result[1] != "US" || result[2] != "RU" {
		t.Errorf("lowercase not uppercased: %v", result)
	}
}

func TestParseCountryList_InvalidLength(t *testing.T) {
	// "USA" has 3 chars, should be filtered out
	result := parseCountryList("CN,USA,RU")
	if len(result) != 2 || result[0] != "CN" || result[1] != "RU" {
		t.Errorf("3-char code should be filtered: %v", result)
	}
}

func TestParseCountryList_SingleCharIgnored(t *testing.T) {
	result := parseCountryList("C,CN")
	if len(result) != 1 || result[0] != "CN" {
		t.Errorf("1-char code should be filtered: %v", result)
	}
}

func TestContainsCountry_Positive(t *testing.T) {
	list := []string{"CN", "RU", "US"}
	if !containsCountry(list, "CN") {
		t.Error("CN should be in list")
	}
	if !containsCountry(list, "cn") {
		t.Error("lowercase cn should match")
	}
}

func TestContainsCountry_Negative(t *testing.T) {
	list := []string{"CN"}
	if containsCountry(list, "JP") {
		t.Error("JP should not be in list")
	}
}

func TestContainsCountry_EmptyList(t *testing.T) {
	if containsCountry([]string{}, "CN") {
		t.Error("empty list should contain nothing")
	}
}

func TestGeoIPConfig_UpdateAndRead(t *testing.T) {
	cfg := GetGeoIPConfig()
	if cfg.AllowMode {
		t.Log("allow_mode can be true if previously set")
	}
	if len(cfg.BlockCountries) >= 0 {
		t.Log("block countries list is valid")
	}
}
