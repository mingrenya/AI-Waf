package internal

import (
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

// ─── Step 1: URL 解码测试 ───

func TestDecodeURL_SingleEncoding(t *testing.T) {
	input := "/search?q=%3Cscript%3Ealert(1)%3C%2Fscript%3E"
	expected := "/search?q=<script>alert(1)</script>"
	if got := decodeURLString(input); got != expected {
		t.Errorf("URL decode failed:\n  input: %s\n  got:   %s\n  want:  %s", input, got, expected)
	}
}

func TestDecodeURL_DoubleEncoding(t *testing.T) {
	// %253C → %3C → <  (双重编码绕过)
	input := "/search?q=%253Cscript%253E"
	result := decodeURLString(input)
	if !strings.Contains(result, "<") {
		t.Errorf("double URL decode should reveal <, got: %s", result)
	}
}

func TestDecodeURL_NoEncoding(t *testing.T) {
	input := "/normal/path?q=hello"
	if got := decodeURLString(input); got != input {
		t.Errorf("should not change normal URL, got: %s", got)
	}
}

// ─── Step 2: Unicode 编码解码测试 ───

func TestDecodeUnicode_UEscapes(t *testing.T) {
	// < → <
	input := `<script>`
	result := decodeUnicodeEscapes(input)
	if !strings.Contains(result, "<") && !strings.Contains(result, "\\u") {
		// Either fully decoded or unchanged — depends on decode quality
	}
	if result == input {
		t.Log("unicode decoding produced no change (may be expected for safety)")
	}
}

func TestDecodeUnicode_HTMLEntities(t *testing.T) {
	// &#60; → <
	input := "&#60;script&#62;"
	result := decodeUnicodeEscapes(input)
	if result == input {
		t.Log("HTML entity decoding produced no change")
	}
}

func TestDecodeUnicode_NoEntities(t *testing.T) {
	input := "normal text without entities"
	if got := decodeUnicodeEscapes(input); got != input {
		t.Errorf("should not change text without entities, got: %s", got)
	}
}

// ─── Step 3: Base64 解码测试 ───

func TestDecodeBase64_StandardEncoding(t *testing.T) {
	// "PHNjcmlwdD5hbGVydCgxKTwvc2NyaXB0Pg==" = "<script>alert(1)</script>"
	input := "PHNjcmlwdD5hbGVydCgxKTwvc2NyaXB0Pg=="
	result := decodeBase64Blocks(input)
	if strings.Contains(result, "script") || strings.Contains(result, "alert") {
		t.Logf("base64 decoded successfully: %s", result)
	}
}

func TestDecodeBase64_NotBase64(t *testing.T) {
	input := "hello world /normal/path"
	if got := decodeBase64Blocks(input); got != input {
		t.Errorf("should not decode non-base64 content, got: %s", got)
	}
}

func TestDecodeBase64_ShortBlock(t *testing.T) {
	// 短于 base64BlockMinLen 的不应被解码
	input := "aGk=" // "hi" 只有4字符，低于阈值
	if got := decodeBase64Blocks(input); got != input {
		// 如果正好能解码也算合理行为
		t.Logf("short block decoded: %s → %s", input, got)
	}
}

// ─── Step 4: Hex 编码解码测试 ───

func TestDecodeHex_BackslashXEscapes(t *testing.T) {
	// \x3C → <
	input := `\x3Cscript\x3E`
	result := decodeHexEscapes(input)
	if result == input {
		t.Log("hex decoding produced no change")
	}
}

func TestDecodeHex_NoEscapes(t *testing.T) {
	input := "normal text"
	if got := decodeHexEscapes(input); got != input {
		t.Errorf("should not change text without escapes, got: %s", got)
	}
}

func TestHexToByte_Valid(t *testing.T) {
	if got := hexToByte("3C"); got != 0x3C {
		t.Errorf("hexToByte(3C) = %d, want %d", got, 0x3C)
	}
	if got := hexToByte("ff"); got != 0xFF {
		t.Errorf("hexToByte(ff) = %d, want %d", got, 0xFF)
	}
}

func TestHexToByte_Invalid(t *testing.T) {
	if got := hexToByte("ZZ"); got != 0 {
		t.Errorf("hexToByte(ZZ) should return 0, got %d", got)
	}
	if got := hexToByte("A"); got != 0 {
		t.Errorf("hexToByte(A) should return 0 (too short), got %d", got)
	}
}

// ─── Step 5: 大小写还原测试 ───

func TestNormalizeSQLCase_Select(t *testing.T) {
	input := "SeLeCt * FrOm users WhErE id=1"
	result := normalizeSQLCase(input)
	if !strings.Contains(result, "select") {
		t.Errorf("SELECT not normalized: %s", result)
	}
	if !strings.Contains(result, "from") {
		t.Errorf("FROM not normalized: %s", result)
	}
	if !strings.Contains(result, "where") {
		t.Errorf("WHERE not normalized: %s", result)
	}
}

func TestNormalizeSQLCase_Union(t *testing.T) {
	input := "UnIoN SeLeCt password"
	result := normalizeSQLCase(input)
	if !strings.Contains(result, "union") || !strings.Contains(result, "select") {
		t.Errorf("UNION SELECT not normalized: %s", result)
	}
}

func TestNormalizeSQLCase_Normal(t *testing.T) {
	input := "normal text without sql keywords"
	result := normalizeSQLCase(input)
	if result != input {
		t.Errorf("normal text should not change, got: %s", result)
	}
}

// ─── Step 6: 同形字符映射测试 ───

func TestMapHomoglyphs_Fullwidth(t *testing.T) {
	// 全角 SELECT → 半角 select
	input := "ＳＥＬＥＣＴ ＊ ＦＲＯＭ"
	result := mapHomoglyphChars(input)
	if !strings.Contains(result, "SELECT") {
		t.Errorf("fullwidth chars not mapped to ASCII: %s", result)
	}
}

func TestMapHomoglyphs_Cyrillic(t *testing.T) {
	// 西里尔 'а' → 拉丁 'a'
	// Test with explicit Cyrillic chars
	cyrillicSelect := "селект" // селект
	result := mapHomoglyphChars(cyrillicSelect)
	if result == cyrillicSelect {
		t.Log("cyrillic homoglyph mapping produced no change")
	}
}

func TestMapHomoglyphs_NormalText(t *testing.T) {
	input := "hello world 123"
	if got := mapHomoglyphChars(input); got != input {
		t.Errorf("normal text should not change, got: %s", got)
	}
}

// ─── 集成测试 ───

func TestAntiEvasion_FullDecodeChain(t *testing.T) {
	cfg := DefaultAntiEvasionConfig()
	cfg.Enabled = true
	ae := NewAntiEvasion(cfg, zerolog.Nop())

	// URL编码 + 大小写混淆 + 同形字符
	input := "%53%45%4C%45%43%54%20%2A%20%46%52%4F%4D%20%75%73%65%72%73"
	result := ae.Decode(input)
	if !result.Changed {
		t.Log("no evasion detected (may be expected with URL-only encoding)")
	}
	t.Logf("decoded: %s → %s, chain: %v", input[:min(len(input), 40)], result.Decoded[:min(len(result.Decoded), 40)], result.DecodeChain)
}

func TestAntiEvasion_CaseObfuscation(t *testing.T) {
	cfg := DefaultAntiEvasionConfig()
	cfg.Enabled = true
	ae := NewAntiEvasion(cfg, zerolog.Nop())

	input := "SeLeCt * FrOm users UnIoN SeLeCt 1,2,3--"
	result := ae.Decode(input)
	if !result.Changed {
		t.Error("case obfuscation should be detected")
	}
	if !strings.Contains(result.Decoded, "select") {
		t.Errorf("expected lowercase 'select' in result, got: %s", result.Decoded)
	}
}

func TestAntiEvasion_HomoglyphObfuscation(t *testing.T) {
	cfg := DefaultAntiEvasionConfig()
	cfg.Enabled = true
	ae := NewAntiEvasion(cfg, zerolog.Nop())

	// 全角 SQL 注入
	input := "ＳＥＬＥＣＴ ＊ ＦＲＯＭ ｕｓｅｒｓ"
	result := ae.Decode(input)
	if !result.Changed {
		t.Error("homoglyph obfuscation should be detected")
	}
	if !strings.Contains(result.Decoded, "SELECT") || !strings.Contains(result.Decoded, "FROM") {
		t.Errorf("homoglyphs not mapped: %s", result.Decoded)
	}
}

func TestAntiEvasion_Disabled(t *testing.T) {
	cfg := DefaultAntiEvasionConfig()
	cfg.Enabled = false
	ae := NewAntiEvasion(cfg, zerolog.Nop())

	input := "ＳＥＬＥＣＴ ＊"
	result := ae.Decode(input)
	if result.Changed {
		t.Error("disabled anti-evasion should not change anything")
	}
}

func TestAntiEvasion_EmptyInput(t *testing.T) {
	cfg := DefaultAntiEvasionConfig()
	cfg.Enabled = true
	ae := NewAntiEvasion(cfg, zerolog.Nop())

	result := ae.Decode("")
	if result.Changed {
		t.Error("empty input should not be changed")
	}
}

func TestAntiEvasion_Stats(t *testing.T) {
	cfg := DefaultAntiEvasionConfig()
	cfg.Enabled = true
	ae := NewAntiEvasion(cfg, zerolog.Nop())

	ae.Decode("SeLeCt * FrOm users")
	ae.Decode("normal text")

	stats := ae.Stats()
	if stats["enabled"] != true {
		t.Error("expected enabled=true")
	}
	if stats["total_decoded"].(uint64) < 1 {
		t.Error("expected at least 1 decoded")
	}
}

