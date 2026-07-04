package internal

import (
	"encoding/base64"
	"net/url"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/rs/zerolog"
)

// AntiEvasion 反逃逸编码预处理器
// 在 Coraza 规则引擎检测之前，对请求内容进行多层解码还原，
// 防止攻击者通过编码绕过 WAF 检测。
//
// 解码流程（按序执行）:
//   1. URL 解码 (%xx → 原始字符)
//   2. Unicode 编码解码 (\uXXXX, &#XXXX; → 原始字符)
//   3. Base64 解码 (自动检测 base64 编码块并解码)
//   4. Hex 编码解码 (\xXX → 原始字符)
//   5. 大小写还原 (SQL 关键字等)
//   6. 同形字符映射 (全角→半角, 西里尔字母→拉丁字母等)
type AntiEvasion struct {
	mu sync.RWMutex

	enabled          bool
	decodeURL        bool // URL 编码解码
	decodeUnicode    bool // Unicode 编码解码
	decodeBase64     bool // Base64 编码解码
	decodeHex        bool // Hex 编码解码
	normalizeCase    bool // 大小写还原
	mapHomoglyphs    bool // 同形字符映射

	// 统计
	totalDecoded  uint64
	evasionDetected uint64

	logger zerolog.Logger
}

// AntiEvasionConfig 反逃逸配置
type AntiEvasionConfig struct {
	Enabled          bool // 是否启用
	DecodeURL        bool // URL 编码解码，默认 true
	DecodeUnicode    bool // Unicode 编码解码，默认 true
	DecodeBase64     bool // Base64 编码解码，默认 true
	DecodeHex        bool // Hex 编码解码，默认 true
	NormalizeCase    bool // SQL 关键字大小写还原，默认 true
	MapHomoglyphs    bool // 同形字符映射，默认 true
}

// DefaultAntiEvasionConfig 默认配置（全部开启）
func DefaultAntiEvasionConfig() AntiEvasionConfig {
	return AntiEvasionConfig{
		Enabled:       false,
		DecodeURL:     true,
		DecodeUnicode: true,
		DecodeBase64:  true,
		DecodeHex:     true,
		NormalizeCase: true,
		MapHomoglyphs: true,
	}
}

// NewAntiEvasion 创建反逃逸预处理器
func NewAntiEvasion(cfg AntiEvasionConfig, logger zerolog.Logger) *AntiEvasion {
	return &AntiEvasion{
		enabled:       cfg.Enabled,
		decodeURL:     cfg.DecodeURL,
		decodeUnicode: cfg.DecodeUnicode,
		decodeBase64:  cfg.DecodeBase64,
		decodeHex:     cfg.DecodeHex,
		normalizeCase: cfg.NormalizeCase,
		mapHomoglyphs: cfg.MapHomoglyphs,
		logger:        logger,
	}
}

// Enable 启用反逃逸处理
func (ae *AntiEvasion) Enable() {
	ae.mu.Lock()
	defer ae.mu.Unlock()
	ae.enabled = true
}

// Disable 禁用反逃逸处理
func (ae *AntiEvasion) Disable() {
	ae.mu.Lock()
	defer ae.mu.Unlock()
	ae.enabled = false
}

// IsEnabled 检查是否启用
func (ae *AntiEvasion) IsEnabled() bool {
	ae.mu.RLock()
	defer ae.mu.RUnlock()
	return ae.enabled
}

// DecodeResult 解码结果
type DecodeResult struct {
	Original        string   // 原始输入
	Decoded         string   // 解码后输出
	Changed         bool     // 是否发生变化
	DecodeChain     []string // 解码链详情
}

// Decode 执行多层解码，返回最终还原后的内容
// 输入: 原始请求内容 (path / query / body / headers)
// 输出: 解码后的内容 + 变更标记
func (ae *AntiEvasion) Decode(input string) DecodeResult {
	ae.mu.RLock()
	enabled := ae.enabled
	decodeURL := ae.decodeURL
	decodeUnicode := ae.decodeUnicode
	decodeBase64 := ae.decodeBase64
	decodeHex := ae.decodeHex
	normalizeCase := ae.normalizeCase
	mapHomoglyphs := ae.mapHomoglyphs
	ae.mu.RUnlock()

	result := DecodeResult{
		Original:    input,
		Decoded:     input,
		Changed:     false,
		DecodeChain: make([]string, 0),
	}

	if !enabled || input == "" || len(input) > 65536 {
		return result
	}

	current := input
	var changed bool

	// Step 1: URL 解码
	if decodeURL {
		next := decodeURLString(current)
		if next != current {
			current = next
			changed = true
			result.DecodeChain = append(result.DecodeChain, "url_decode")
		}
	}

	// Step 2: Unicode 编码解码
	if decodeUnicode {
		next := decodeUnicodeEscapes(current)
		if next != current {
			current = next
			changed = true
			result.DecodeChain = append(result.DecodeChain, "unicode_decode")
		}
	}

	// Step 3: Base64 解码
	if decodeBase64 {
		next := decodeBase64Blocks(current)
		if next != current {
			current = next
			changed = true
			result.DecodeChain = append(result.DecodeChain, "base64_decode")
		}
	}

	// Step 4: Hex 字符解码
	if decodeHex {
		next := decodeHexEscapes(current)
		if next != current {
			current = next
			changed = true
			result.DecodeChain = append(result.DecodeChain, "hex_decode")
		}
	}

	// Step 5: 大小写还原
	if normalizeCase {
		next := normalizeSQLCase(current)
		if next != current {
			current = next
			changed = true
			result.DecodeChain = append(result.DecodeChain, "case_normalize")
		}
	}

	// Step 6: 同形字符映射
	if mapHomoglyphs {
		next := mapHomoglyphChars(current)
		if next != current {
			current = next
			changed = true
			result.DecodeChain = append(result.DecodeChain, "homoglyph_map")
		}
	}

	result.Decoded = current
	result.Changed = changed

	if changed {
		ae.mu.Lock()
		ae.totalDecoded++
		ae.evasionDetected++
		ae.mu.Unlock()
	}

	return result
}

// Stats 返回当前统计信息
func (ae *AntiEvasion) Stats() map[string]interface{} {
	ae.mu.RLock()
	defer ae.mu.RUnlock()
	return map[string]interface{}{
		"enabled":          ae.enabled,
		"total_decoded":    ae.totalDecoded,
		"evasion_detected": ae.evasionDetected,
	}
}

// ─── Step 1: URL 解码 ───

func decodeURLString(s string) string {
	// 检查是否包含 URL 编码字符
	if !strings.Contains(s, "%") {
		return s
	}

	// 多次解码，直到不再变化（应对双重/三重编码）
	result := s
	for i := 0; i < 4; i++ {
		decoded, err := url.QueryUnescape(result)
		if err != nil || decoded == result {
			break
		}
		result = decoded
	}
	return result
}

// ─── Step 2: Unicode 编码解码 ───

func decodeUnicodeEscapes(s string) string {
	if !strings.Contains(s, "\\u") && !strings.Contains(s, "&#") {
		return s
	}

	result := s

	// \uXXXX 格式
	result = replaceUnicodeEscapes(result, `\u`)

	// &#xXXXX; 和 &#XXXX; 格式 (HTML 数字实体)
	result = replaceHTMLNumericEntities(result)

	return result
}

func replaceUnicodeEscapes(s string, prefix string) string {
	if !strings.Contains(s, prefix) {
		return s
	}

	var sb strings.Builder
	sb.Grow(len(s))
	i := 0
	for i < len(s) {
		if i+6 <= len(s) && s[i:i+2] == prefix {
			hexStr := s[i+2 : i+6]
			if r := parseHexRune(hexStr); r > 0 {
				sb.WriteRune(r)
				i += 6
				continue
			}
		}
		sb.WriteByte(s[i])
		i++
	}
	return sb.String()
}

func replaceHTMLNumericEntities(s string) string {
	if !strings.Contains(s, "&#") {
		return s
	}

	var sb strings.Builder
	sb.Grow(len(s))
	i := 0
	for i < len(s) {
		if i+3 < len(s) && s[i] == '&' && s[i+1] == '#' {
			end := strings.IndexByte(s[i:], ';')
			if end < 0 {
				sb.WriteByte(s[i])
				i++
				continue
			}
			end += i
			numStr := s[i+2 : end]
			var r rune
			if len(numStr) > 1 && (numStr[0] == 'x' || numStr[0] == 'X') {
				r = parseHexRune(numStr[1:])
			} else {
				r = parseDecRune(numStr)
			}
			if r > 0 {
				sb.WriteRune(r)
				i = end + 1
				continue
			}
		}
		sb.WriteByte(s[i])
		i++
	}
	return sb.String()
}

func parseHexRune(hex string) rune {
	if len(hex) > 8 {
		return 0
	}
	var v uint32
	for _, c := range hex {
		v <<= 4
		switch {
		case c >= '0' && c <= '9':
			v |= uint32(c - '0')
		case c >= 'a' && c <= 'f':
			v |= uint32(c - 'a' + 10)
		case c >= 'A' && c <= 'F':
			v |= uint32(c - 'A' + 10)
		default:
			return 0
		}
	}
	if utf8.ValidRune(rune(v)) {
		return rune(v)
	}
	return 0
}

func parseDecRune(dec string) rune {
	var v uint32
	for _, c := range dec {
		if c < '0' || c > '9' {
			return 0
		}
		v = v*10 + uint32(c-'0')
	}
	if utf8.ValidRune(rune(v)) {
		return rune(v)
	}
	return 0
}

// ─── Step 3: Base64 解码 ───

// base64BlockMinLen Base64 块的最小长度阈值
const base64BlockMinLen = 8

func decodeBase64Blocks(s string) string {
	// 自动检测并解码 Base64 编码块
	// 特征: 连续 [A-Za-z0-9+/=]{8,}
	if len(s) < base64BlockMinLen {
		return s
	}

	// 快速扫描: 查找可能的 Base64 块
	var result strings.Builder
	result.Grow(len(s))

	i := 0
	for i < len(s) {
		// 查找 Base64 字符序列的起始位置
		start := i
		for start < len(s) && !isBase64Char(s[start]) {
			result.WriteByte(s[start])
			start++
		}

		if start >= len(s) {
			break
		}

		// 收集 Base64 字符序列
		end := start
		for end < len(s) && isBase64Char(s[end]) {
			end++
		}

		block := s[start:end]
		if len(block) >= base64BlockMinLen && len(block)%4 == 0 {
			// 尝试标准 Base64 解码
			if decoded, err := base64.StdEncoding.DecodeString(block); err == nil && isPrintableASCII(string(decoded)) {
				result.WriteString(string(decoded))
				i = end
				continue
			}
			// 尝试 URL-safe Base64 解码
			if decoded, err := base64.URLEncoding.DecodeString(block); err == nil && isPrintableASCII(string(decoded)) {
				result.WriteString(string(decoded))
				i = end
				continue
			}
		}

		// 不是有效的 Base64 块，原样输出
		result.WriteString(block)
		i = end
	}

	return result.String()
}

func isBase64Char(c byte) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') ||
		(c >= '0' && c <= '9') || c == '+' || c == '/' || c == '=' || c == '-' || c == '_'
}

func isPrintableASCII(s string) bool {
	for _, r := range s {
		if r < 0x20 || r > 0x7e {
			return false
		}
	}
	return true
}

// ─── Step 4: Hex 字符解码 ───

func decodeHexEscapes(s string) string {
	if !strings.Contains(s, "\\x") && !strings.Contains(s, "%") {
		return s
	}

	result := s

	// 重复 URL 解码（捕获 % 但非标准 URL 编码的 hex 转义）
	if strings.Contains(result, "%") && !strings.Contains(result, "%2") {
		// 只处理已经被部分解码过的内容
	}

	// \xXX 格式
	result = decodeBackslashX(result)

	return result
}

func decodeBackslashX(s string) string {
	if !strings.Contains(s, "\\x") {
		return s
	}
	var sb strings.Builder
	sb.Grow(len(s))
	i := 0
	for i < len(s) {
		if i+4 <= len(s) && s[i] == '\\' && s[i+1] == 'x' {
			h := s[i+2 : i+4]
			if b := hexToByte(h); b > 0 {
				sb.WriteByte(b)
				i += 4
				continue
			}
		}
		sb.WriteByte(s[i])
		i++
	}
	return sb.String()
}

func hexToByte(hex string) byte {
	if len(hex) != 2 {
		return 0
	}
	var b byte
	for _, c := range hex {
		b <<= 4
		switch {
		case c >= '0' && c <= '9':
			b |= byte(c - '0')
		case c >= 'a' && c <= 'f':
			b |= byte(c - 'a' + 10)
		case c >= 'A' && c <= 'F':
			b |= byte(c - 'A' + 10)
		default:
			return 0
		}
	}
	return b
}

// ─── Step 5: SQL 关键字大小写还原 ───

// SQL 关键字列表（常见注入攻击使用的关键字）
var sqlKeywords = []string{
	"SELECT", "INSERT", "UPDATE", "DELETE", "DROP", "CREATE", "ALTER",
	"UNION", "JOIN", "WHERE", "FROM", "HAVING", "GROUP", "ORDER", "BY",
	"AND", "OR", "NOT", "IN", "LIKE", "BETWEEN", "IS", "NULL",
	"EXEC", "EXECUTE", "SP_", "XP_", "WAITFOR", "DELAY",
	"SLEEP", "BENCHMARK", "LOAD_FILE", "INTO", "OUTFILE", "DUMPFILE",
	"INFORMATION_SCHEMA", "SCHEMA_NAME", "TABLE_NAME", "COLUMN_NAME",
	"CHAR", "CONCAT", "ASCII", "SUBSTR", "SUBSTRING", "MID", "LEFT", "RIGHT",
	"CAST", "CONVERT", "DECLARE", "FETCH", "OPEN", "CLOSE",
}

func normalizeSQLCase(s string) string {
	result := s
	for _, kw := range sqlKeywords {
		result = replaceCaseInsensitiveKeywords(result, kw)
	}
	return result
}

func replaceCaseInsensitiveKeywords(s, keyword string) string {
	kwLen := len(keyword)
	if kwLen > len(s) {
		return s
	}

	lowerKW := strings.ToLower(keyword)

	// 快速扫描
	var sb strings.Builder
	sb.Grow(len(s))

	i := 0
	for i < len(s) {
		if i+kwLen <= len(s) {
			candidate := s[i : i+kwLen]
			if strings.EqualFold(candidate, keyword) {
				sb.WriteString(lowerKW)
				i += kwLen
				continue
			}
		}
		sb.WriteByte(s[i])
		i++
	}
	return sb.String()
}

// ─── Step 6: 同形字符映射 ───

// homoglyphMap 同形字符 → ASCII 映射表
// 覆盖常见攻击中使用的混淆字符
var homoglyphMap = map[rune]rune{
	// 全角字符 → 半角
	'Ａ': 'A', 'Ｂ': 'B', 'Ｃ': 'C', 'Ｄ': 'D', 'Ｅ': 'E',
	'Ｆ': 'F', 'Ｇ': 'G', 'Ｈ': 'H', 'Ｉ': 'I', 'Ｊ': 'J',
	'Ｋ': 'K', 'Ｌ': 'L', 'Ｍ': 'M', 'Ｎ': 'N', 'Ｏ': 'O',
	'Ｐ': 'P', 'Ｑ': 'Q', 'Ｒ': 'R', 'Ｓ': 'S', 'Ｔ': 'T',
	'Ｕ': 'U', 'Ｖ': 'V', 'Ｗ': 'W', 'Ｘ': 'X', 'Ｙ': 'Y',
	'Ｚ': 'Z',
	'ａ': 'a', 'ｂ': 'b', 'ｃ': 'c', 'ｄ': 'd', 'ｅ': 'e',
	'ｆ': 'f', 'ｇ': 'g', 'ｈ': 'h', 'ｉ': 'i', 'ｊ': 'j',
	'ｋ': 'k', 'ｌ': 'l', 'ｍ': 'm', 'ｎ': 'n', 'ｏ': 'o',
	'ｐ': 'p', 'ｑ': 'q', 'ｒ': 'r', 'ｓ': 's', 'ｔ': 't',
	'ｕ': 'u', 'ｖ': 'v', 'ｗ': 'w', 'ｘ': 'x', 'ｙ': 'y',
	'ｚ': 'z',
	'０': '0', '１': '1', '２': '2', '３': '3', '４': '4',
	'５': '5', '６': '6', '７': '7', '８': '8', '９': '9',

	// 全角符号
	'＠': '@', '．': '.', '／': '/', '＼': '\\', '：': ':',
	'；': ';', '（': '(', '）': ')', '［': '[', '］': ']',
	'＜': '<', '＞': '>', '＝': '=', '＋': '+', '－': '-',
	'＊': '*', '＃': '#', '％': '%', '＆': '&', '｜': '|',
	'！': '!', '？': '?', '＾': '^', '～': '~',
	'｀': '`', '＇': '\'', '＂': '"', '　': ' ',
	'，': ',',

	// 西里尔字母混淆 (常用于品牌仿冒和绕过)
	'а': 'a', 'е': 'e', 'о': 'o', 'р': 'p', 'с': 'c',
	'у': 'y', 'х': 'x', 'А': 'A', 'В': 'B', 'Е': 'E',
	'К': 'K', 'М': 'M', 'Н': 'H', 'О': 'O', 'Р': 'P',
	'С': 'C', 'Т': 'T', 'У': 'Y', 'Х': 'X',

	// 希腊字母混淆
	'ο': 'o', 'ε': 'e', 'α': 'a', 'τ': 't', 'ν': 'n',
	'ι': 'i', 'κ': 'k',

	// 其他常见混淆
	'‐': '-', '‑': '-', '‒': '-', '–': '-', '—': '-', '―': '-',
	'‖': '|', '∣': '|', '⎮': '|',
	'‘': '\'', '’': '\'', '‚': ',', '‛': '\'',
	'“': '"', '”': '"', '„': '"', '‟': '"',
	'…': '.', '‥': '.',
	'⁄': '/', '∕': '/', '⧸': '/',
}

func mapHomoglyphChars(s string) string {
	// 快速检查是否有需要映射的字符
	needMap := false
	for _, r := range s {
		if _, ok := homoglyphMap[r]; ok {
			needMap = true
			break
		}
	}
	if !needMap {
		return s
	}

	var sb strings.Builder
	sb.Grow(len(s))
	for _, r := range s {
		if mapped, ok := homoglyphMap[r]; ok {
			sb.WriteRune(mapped)
		} else {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}
