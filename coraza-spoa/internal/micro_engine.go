package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/mingrenya/AI-Waf/pkg/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// 常量定义
const (
	// 最大递归深度限制,防止栈溢出
	MaxRecursionDepth = 10
	// 正则表达式最大长度
	MaxRegexLength = 1000
)

// 匹配方式
type MatchType string

const (
	// IP匹配方式
	MatchEqual        MatchType = "equal"
	MatchNotEqual     MatchType = "not_equal"
	MatchFuzzy        MatchType = "fuzzy"
	MatchInCIDR       MatchType = "in_cidr"
	MatchNotInCIDR    MatchType = "not_in_cidr"
	MatchInIPGroup    MatchType = "in_ipgroup"
	MatchNotInIPGroup MatchType = "not_in_ipgroup"

	// URL和Path匹配方式
	MatchInclude       MatchType = "include"
	MatchContains      MatchType = "contains"
	MatchNotContains   MatchType = "not_contains"
	MatchPrefixKeyword MatchType = "prefix_keyword"
	MatchRegex         MatchType = "regex"
)

// 匹配目标类型
type TargetType string

const (
	SourceIP   TargetType = "source_ip"
	TargetURL  TargetType = "url"
	TargetPath TargetType = "path"
)

// 逻辑操作符
type LogicalOperator string

const (
	LogicalAND LogicalOperator = "AND"
	LogicalOR  LogicalOperator = "OR"
)

// regexCacheEntry 正则表达式缓存条目
type regexCacheEntry struct {
	re       *regexp.Regexp
	lastUsed time.Time
	useCount int
}

// regexCache 线程安全的正则表达式缓存，支持 LRU 淘汰策略
type regexCache struct {
	mu      sync.RWMutex
	cache   map[string]*regexCacheEntry
	maxSize int
	ttl     time.Duration
}

// newRegexCache 创建新的正则缓存
func newRegexCache(maxSize int, ttl time.Duration) *regexCache {
	return &regexCache{
		cache:   make(map[string]*regexCacheEntry),
		maxSize: maxSize,
		ttl:     ttl,
	}
}

// get 获取缓存的正则表达式
func (rc *regexCache) get(pattern string) (*regexp.Regexp, bool) {
	rc.mu.RLock()
	entry, exists := rc.cache[pattern]
	rc.mu.RUnlock()

	if !exists {
		return nil, false
	}

	// 检查是否过期
	if rc.ttl > 0 && time.Since(entry.lastUsed) > rc.ttl {
		rc.mu.Lock()
		delete(rc.cache, pattern)
		rc.mu.Unlock()
		return nil, false
	}

	// 更新使用信息
	rc.mu.Lock()
	entry.lastUsed = time.Now()
	entry.useCount++
	rc.mu.Unlock()

	return entry.re, true
}

// set 设置缓存的正则表达式
func (rc *regexCache) set(pattern string, re *regexp.Regexp) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	// 如果缓存已满，移除最少使用的条目
	if len(rc.cache) >= rc.maxSize {
		rc.evictLRU()
	}

	rc.cache[pattern] = &regexCacheEntry{
		re:       re,
		lastUsed: time.Now(),
		useCount: 1,
	}
}

// evictLRU 淘汰最少使用的缓存条目
func (rc *regexCache) evictLRU() {
	var oldestPattern string
	var oldestTime time.Time
	var minUseCount int
	firstEntry := true

	for pattern, entry := range rc.cache {
		if firstEntry {
			oldestPattern = pattern
			oldestTime = entry.lastUsed
			minUseCount = entry.useCount
			firstEntry = false
			continue
		}

		// 优先移除使用次数少的，相同时移除最久未使用的
		if entry.useCount < minUseCount ||
			(entry.useCount == minUseCount && entry.lastUsed.Before(oldestTime)) {
			oldestPattern = pattern
			oldestTime = entry.lastUsed
			minUseCount = entry.useCount
		}
	}

	if oldestPattern != "" {
		delete(rc.cache, oldestPattern)
	}
}

// Matcher接口定义了条件匹配的方法
type Matcher interface {
	Match(eng *RuleEngine, ip, url, path string) (bool, error)
}

// 条件类型
type ConditionType string

const (
	SimpleConditionType    ConditionType = "simple"
	CompositeConditionType ConditionType = "composite"
)

// SimpleCondition 简单条件
type SimpleCondition struct {
	Type       ConditionType `json:"type" bson:"type"`
	Target     TargetType    `json:"target" bson:"target"`
	MatchType  MatchType     `json:"match_type" bson:"match_type"`
	MatchValue string        `json:"match_value" bson:"match_value"`
}

// Match 实现Matcher接口
func (c *SimpleCondition) Match(eng *RuleEngine, ip, url, path string) (bool, error) {
	switch c.Target {
	case SourceIP:
		return eng.matchIP(c, ip)
	case TargetURL:
		return eng.matchURL(c, url)
	case TargetPath:
		return eng.matchPath(c, path)
	default:
		return false, fmt.Errorf("不支持的目标类型: %s", c.Target)
	}
}

// CompositeCondition 复合条件
type CompositeCondition struct {
	Type       ConditionType   `json:"type" bson:"type"`
	Operator   LogicalOperator `json:"operator" bson:"operator"`
	Conditions []bson.Raw      `json:"conditions" bson:"conditions"`

	// 运行时字段，不用于JSON/BSON
	parsedConditions []Matcher
}

// Match 实现Matcher接口
func (c *CompositeCondition) Match(eng *RuleEngine, ip, url, path string) (bool, error) {
	if len(c.parsedConditions) == 0 {
		return false, fmt.Errorf("复合条件未初始化")
	}

	var result bool
	if c.Operator == LogicalAND {
		result = true
	} else {
		result = false
	}

	for _, condition := range c.parsedConditions {
		match, err := condition.Match(eng, ip, url, path)
		if err != nil {
			return false, err
		}

		if c.Operator == LogicalAND {
			result = result && match
			if !result {
				return false, nil // 短路
			}
		} else {
			result = result || match
			if result {
				return true, nil // 短路
			}
		}
	}

	return result, nil
}

// ConditionFactory 条件工厂
type ConditionFactory struct{}

// ParseCondition 解析条件(公共入口)
func (f *ConditionFactory) ParseCondition(data bson.Raw) (Matcher, error) {
	return f.parseConditionWithDepth(data, 0)
}

// parseConditionWithDepth 带递归深度限制的条件解析
func (f *ConditionFactory) parseConditionWithDepth(data bson.Raw, depth int) (Matcher, error) {
	// 检查递归深度限制
	if depth > MaxRecursionDepth {
		return nil, fmt.Errorf("条件嵌套深度超过限制(%d层)", MaxRecursionDepth)
	}

	// 检查数据是否为空
	if len(data) == 0 {
		return nil, fmt.Errorf("条件数据为空")
	}

	var baseCondition struct {
		Type ConditionType `json:"type" bson:"type"`
	}

	if err := bson.Unmarshal(data, &baseCondition); err != nil {
		return nil, fmt.Errorf("解析条件类型失败(深度%d): %v", depth, err)
	}

	switch baseCondition.Type {
	case SimpleConditionType:
		var condition SimpleCondition
		if err := bson.Unmarshal(data, &condition); err != nil {
			return nil, fmt.Errorf("解析简单条件失败(深度%d): %v", depth, err)
		}
		// 验证简单条件
		if err := validateSimpleCondition(&condition); err != nil {
			return nil, fmt.Errorf("简单条件验证失败(深度%d): %v", depth, err)
		}
		return &condition, nil

	case CompositeConditionType:
		var condition CompositeCondition
		if err := bson.Unmarshal(data, &condition); err != nil {
			return nil, fmt.Errorf("解析复合条件失败(深度%d): %v", depth, err)
		}

		// 验证复合条件
		if err := validateCompositeCondition(&condition); err != nil {
			return nil, fmt.Errorf("复合条件验证失败(深度%d): %v", depth, err)
		}

		condition.parsedConditions = make([]Matcher, 0, len(condition.Conditions))
		for i, rawCondition := range condition.Conditions {
			// 递归解析子条件,深度+1
			parsedCondition, err := f.parseConditionWithDepth(rawCondition, depth+1)
			if err != nil {
				return nil, fmt.Errorf("解析第%d个子条件失败: %v", i, err)
			}
			condition.parsedConditions = append(condition.parsedConditions, parsedCondition)
		}

		return &condition, nil

	default:
		return nil, fmt.Errorf("不支持的条件类型: %s (深度%d)", baseCondition.Type, depth)
	}
}

// Rule 规则 - 添加优先级和序列号字段
type Rule struct {
	// 嵌入MicroRule
	model.MicroRule `bson:",inline" json:",inline"`

	// 运行时字段，不用于JSON/BSON
	parsedCondition Matcher `bson:"-" json:"-"`
	sequence        int     `bson:"-" json:"-"`
}

// MongoDB配置
type MongoDBConfig struct {
	MongoClient       *mongo.Client
	Database          string // 数据库名称
	RuleCollection    string // 规则集合名称
	IPGroupCollection string // IP组集合名称
}

// ipGroupCache IP组缓存结构,用于优化IP查找
type ipGroupCache struct {
	exactIPs map[string]bool // 精确IP映射
	cidrNets []*net.IPNet    // CIDR网络列表
}

// RuleEngine 规则引擎
type RuleEngine struct {
	Rules         []Rule                    `json:"rules"`     // 所有规则列表
	IPGroups      map[string]*model.IPGroup `json:"ip_groups"` // IP组映射表
	ipGroupCaches map[string]*ipGroupCache  // IP组查找缓存
	regexCache    *regexCache               // 正则表达式缓存（支持 LRU）
	factory       ConditionFactory          // 条件工厂
	mongoConfig   *MongoDBConfig            // MongoDB配置
}

// NewRuleEngine 创建规则引擎
func NewRuleEngine() *RuleEngine {
	return &RuleEngine{
		Rules:         make([]Rule, 0),
		IPGroups:      make(map[string]*model.IPGroup),
		ipGroupCaches: make(map[string]*ipGroupCache),
		// 使用 LRU 缓存，最多缓存 1000 个正则表达式，1 小时过期
		regexCache: newRegexCache(1000, 1*time.Hour),
		factory:    ConditionFactory{},
	}
}

func (e *RuleEngine) InitMongoConfig(config *MongoDBConfig) error {
	e.mongoConfig = config
	return nil
}

// LoadIPGroupsFromMongoDB 从MongoDB加载IP组
func (e *RuleEngine) LoadIPGroupsFromMongoDB() error {
	if e.mongoConfig.MongoClient == nil {
		return fmt.Errorf("MongoDB客户端未初始化")
	}

	// 获取集合
	collection := e.mongoConfig.MongoClient.
		Database(e.mongoConfig.Database).
		Collection(e.mongoConfig.IPGroupCollection)

	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 检查并创建默认IP组逻辑
	defaultBlacklistCount, err := collection.CountDocuments(ctx, bson.D{{Key: "name", Value: "system_default_blacklist"}})
	if err != nil {
		return fmt.Errorf("检查默认黑名单是否存在失败: %v", err)
	}

	// 如果默认黑名单不存在，则创建
	if defaultBlacklistCount == 0 {
		// 创建默认黑名单组
		defaultBlacklist := model.IPGroup{
			Name:  "system_default_blacklist",
			Items: []string{}, // 初始为空
		}

		// 插入到数据库
		_, err := collection.InsertOne(ctx, defaultBlacklist)
		if err != nil {
			return fmt.Errorf("创建默认IP黑名单组失败: %v", err)
		}
	}

	// 查询所有IP组
	cursor, err := collection.Find(ctx, bson.D{})
	if err != nil {
		return fmt.Errorf("查询IP组失败: %v", err)
	}
	defer cursor.Close(ctx)

	// 解码IP组
	var ipGroups []model.IPGroup
	if err = cursor.All(ctx, &ipGroups); err != nil {
		return fmt.Errorf("解码IP组失败: %v", err)
	}

	// 初始化映射表
	e.IPGroups = make(map[string]*model.IPGroup)
	e.ipGroupCaches = make(map[string]*ipGroupCache)

	// 填充IP组映射
	for _, group := range ipGroups {
		for _, item := range group.Items {
			if !isValidIPOrCIDR(item) {
				return fmt.Errorf("IP组 %s 中包含无效的IP或CIDR: %s", group.Name, item)
			}
		}
		e.IPGroups[group.Name] = &group
		// 预先构建缓存
		if err := e.buildIPGroupCache(group.Name); err != nil {
			return fmt.Errorf("构建IP组 %s 的缓存失败: %v", group.Name, err)
		}
	}

	return nil
}

// LoadRulesFromMongoDB 从MongoDB加载规则
func (e *RuleEngine) LoadRulesFromMongoDB() error {
	if e.mongoConfig.MongoClient == nil {
		return fmt.Errorf("MongoDB客户端未初始化")
	}

	// 获取集合
	collection := e.mongoConfig.MongoClient.
		Database(e.mongoConfig.Database).
		Collection(e.mongoConfig.RuleCollection)

	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 检查并创建默认规则逻辑
	defaultRuleCount, err := collection.CountDocuments(ctx, bson.D{{Key: "name", Value: "system_default_ip_block"}})
	if err != nil {
		return fmt.Errorf("检查默认规则是否存在失败: %v", err)
	}

	// 如果默认规则不存在，则创建
	if defaultRuleCount == 0 {
		// 创建默认规则条件
		defaultCondition := SimpleCondition{
			Type:       "simple",
			Target:     "source_ip",
			MatchType:  "in_ipgroup",
			MatchValue: "system_default_blacklist",
		}

		// 将条件序列化为BSON
		conditionBytes, err := bson.Marshal(defaultCondition)
		if err != nil {
			return fmt.Errorf("序列化默认规则条件失败: %v", err)
		}

		// 创建默认规则
		defaultRule := model.MicroRule{
			Name:      "system_default_ip_block",
			Type:      "blacklist",
			Status:    "enabled",
			Priority:  9999, // 最高优先级
			Condition: conditionBytes,
		}

		// 插入到数据库
		_, err = collection.InsertOne(ctx, defaultRule)
		if err != nil {
			return fmt.Errorf("创建默认规则失败: %v", err)
		}
	}

	// 查询所有规则
	cursor, err := collection.Find(ctx, bson.D{})
	if err != nil {
		return fmt.Errorf("查询规则失败: %v", err)
	}
	defer cursor.Close(ctx)

	// 解码规则
	var rules []Rule
	if err = cursor.All(ctx, &rules); err != nil {
		return fmt.Errorf("解码规则失败: %v", err)
	}

	// 设置序列号
	for i := range rules {
		rules[i].sequence = i
	}

	// 解析每个规则的条件
	for i := range rules {
		rule := &rules[i]
		parsedCondition, err := e.factory.ParseCondition(rule.Condition)
		if err != nil {
			return fmt.Errorf("解析规则 %s 的条件失败: %v", rule.ID, err)
		}
		rule.parsedCondition = parsedCondition
	}

	// 按照优先级排序，优先级相同时按照原始顺序排序
	sort.Slice(rules, func(i, j int) bool {
		if rules[i].Priority != rules[j].Priority {
			return rules[i].Priority > rules[j].Priority // 优先级高的排在前面
		}
		return rules[i].sequence < rules[j].sequence // 优先级相同时按原始顺序
	})

	e.Rules = rules
	return nil
}

// LoadAllFromMongoDB 从MongoDB加载所有规则和IP组
func (e *RuleEngine) LoadAllFromMongoDB() error {
	if err := e.LoadIPGroupsFromMongoDB(); err != nil {
		return err
	}

	return e.LoadRulesFromMongoDB()
}

// AddIPGroup 添加IP组
func (e *RuleEngine) AddIPGroup(group model.IPGroup) error {
	if _, exists := e.IPGroups[group.Name]; exists {
		return fmt.Errorf("IP组 %s 已存在", group.Name)
	}

	for _, item := range group.Items {
		if !isValidIPOrCIDR(item) {
			return fmt.Errorf("IP组 %s 中包含无效的IP或CIDR: %s", group.Name, item)
		}
	}

	e.IPGroups[group.Name] = &group
	return nil
}

// LoadRulesFromJSON 从JSON加载规则 - 修改加载逻辑，增加序列号处理
// 修复: 正确处理 JSON 到 BSON 的转换
func (e *RuleEngine) LoadRulesFromJSON(data []byte) error {
	// 定义临时结构体，用于 JSON 反序列化
	type TempRule struct {
		ID        string           `json:"id,omitempty"`
		Name      string           `json:"name"`
		Type      model.RuleType   `json:"type"`
		Status    model.RuleStatus `json:"status"`
		Priority  int              `json:"priority"`
		Condition interface{}      `json:"condition"` // 使用 interface{} 接收 JSON 对象
	}

	var tempRules []TempRule
	if err := json.Unmarshal(data, &tempRules); err != nil {
		return fmt.Errorf("JSON 反序列化失败: %v", err)
	}

	// 转换为 Rule 对象
	rules := make([]Rule, 0, len(tempRules))
	for i, tempRule := range tempRules {
		// 将 condition 转换为 BSON
		conditionBytes, err := bson.Marshal(tempRule.Condition)
		if err != nil {
			return fmt.Errorf("序列化规则 %s 的条件为 BSON 失败: %v", tempRule.Name, err)
		}

		rule := Rule{
			MicroRule: model.MicroRule{
				Name:      tempRule.Name,
				Type:      tempRule.Type,
				Status:    tempRule.Status,
				Priority:  tempRule.Priority,
				Condition: conditionBytes,
			},
			sequence: i, // 设置序列号
		}

		// 解析条件
		parsedCondition, err := e.factory.ParseCondition(rule.Condition)
		if err != nil {
			return fmt.Errorf("解析规则 %s 的条件失败: %v", rule.Name, err)
		}
		rule.parsedCondition = parsedCondition

		rules = append(rules, rule)
	}

	// 按照优先级排序，优先级相同时按照原始顺序排序
	sort.Slice(rules, func(i, j int) bool {
		if rules[i].Priority != rules[j].Priority {
			return rules[i].Priority > rules[j].Priority // 优先级高的排在前面
		}
		return rules[i].sequence < rules[j].sequence // 优先级相同时按原始顺序
	})

	e.Rules = rules
	return nil
}

// AddRule 添加单个规则
func (e *RuleEngine) AddRule(rule Rule) error {
	// 解析规则条件
	parsedCondition, err := e.factory.ParseCondition(rule.Condition)
	if err != nil {
		return fmt.Errorf("解析规则 %s 的条件失败: %v", rule.ID, err)
	}
	rule.parsedCondition = parsedCondition

	// 设置规则序列号为当前规则列表长度
	rule.sequence = len(e.Rules)

	// 添加规则到列表
	e.Rules = append(e.Rules, rule)

	// 重新排序规则
	sort.Slice(e.Rules, func(i, j int) bool {
		if e.Rules[i].Priority != e.Rules[j].Priority {
			return e.Rules[i].Priority > e.Rules[j].Priority
		}
		return e.Rules[i].sequence < e.Rules[j].sequence
	})

	return nil
}

// MatchRequest 匹配请求
// 参数：
// - ip: 源IP地址
// - url: 请求URL
// - path: 请求路径
// 返回值：
// - shouldBlock: 是否应该拦截请求 (true表示拦截，false表示放行)
// - ruleType: 匹配的规则类型
// - rule: 匹配的规则
// - error: 错误信息
func (e *RuleEngine) MatchRequest(ip string, url string, path string) (shouldBlock bool, ruleType model.RuleType, rule *Rule, err error) {
	// 验证IP地址格式
	if !isValidIP(ip) {
		return false, "", nil, fmt.Errorf("无效的IP地址: %s", ip)
	}

	// 标记是否存在启用的白名单规则
	hasWhitelistRule := false

	// 遍历所有规则（已按优先级和序列号排序）
	for i := range e.Rules {
		r := &e.Rules[i]
		// 检查是否存在启用的白名单规则
		if r.Status == model.RuleEnabled && r.Type == model.WhitelistRule {
			hasWhitelistRule = true
		}

		// 跳过禁用的规则
		if r.Status == model.RuleDisabled {
			continue
		}

		// 匹配规则条件
		match, err := r.parsedCondition.Match(e, ip, url, path)
		if err != nil {
			return false, "", nil, err
		}

		// 如果规则条件匹配
		if match {
			// 根据规则类型确定是否需要拦截
			switch r.Type {
			case model.BlacklistRule:
				// 黑名单规则匹配成功 -> 返回true(拦截)
				return true, r.Type, r, nil
			case model.WhitelistRule:
				// 白名单规则匹配成功 -> 返回false(放行)
				return false, r.Type, r, nil
			default:
				return false, "", nil, fmt.Errorf("未知的规则类型: %s", r.Type)
			}
		}
	}

	// 如果没有匹配任何规则，且存在白名单规则，则拦截
	if hasWhitelistRule {
		// 存在白名单规则但未匹配任何规则 -> 拦截请求（安全默认值）
		return true, "", nil, nil
	}

	// 如果没有匹配任何规则，且不存在白名单规则，则默认放行
	return false, "", nil, nil
}

// GetRules 获取当前规则列表
func (e *RuleEngine) GetRules() []Rule {
	return e.Rules
}

// matchIP 匹配IP条件
func (e *RuleEngine) matchIP(cond *SimpleCondition, ip string) (bool, error) {
	switch cond.MatchType {
	case MatchEqual:
		return ip == cond.MatchValue, nil
	case MatchNotEqual:
		return ip != cond.MatchValue, nil
	case MatchFuzzy:
		return matchIPFuzzy(ip, cond.MatchValue)
	case MatchInCIDR:
		return isIPInCIDR(ip, cond.MatchValue)
	case MatchNotInCIDR:
		inCIDR, err := isIPInCIDR(ip, cond.MatchValue)
		return !inCIDR, err
	case MatchInIPGroup:
		return e.isIPInGroup(ip, cond.MatchValue)
	case MatchNotInIPGroup:
		inGroup, err := e.isIPInGroup(ip, cond.MatchValue)
		return !inGroup, err
	default:
		return false, fmt.Errorf("IP不支持匹配方式: %s", cond.MatchType)
	}
}

// matchURL 匹配URL条件
func (e *RuleEngine) matchURL(cond *SimpleCondition, url string) (bool, error) {
	switch cond.MatchType {
	case MatchEqual:
		return url == cond.MatchValue, nil
	case MatchNotEqual:
		return url != cond.MatchValue, nil
	case MatchInclude, MatchContains:
		return strings.Contains(url, cond.MatchValue), nil
	case MatchNotContains:
		return !strings.Contains(url, cond.MatchValue), nil
	case MatchPrefixKeyword:
		return strings.HasPrefix(url, cond.MatchValue), nil
	case MatchRegex:
		return e.matchRegex(url, cond.MatchValue)
	default:
		return false, fmt.Errorf("URL不支持匹配方式: %s", cond.MatchType)
	}
}

// matchPath 匹配Path条件
func (e *RuleEngine) matchPath(cond *SimpleCondition, path string) (bool, error) {
	return e.matchURL(cond, path)
}

// 以下是辅助函数

// isValidIP 检查IP是否有效
func isValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// isValidCIDR 检查CIDR是否有效
func isValidCIDR(cidr string) bool {
	_, _, err := net.ParseCIDR(cidr)
	return err == nil
}

// isValidIPOrCIDR 检查是否为有效的IP或CIDR
func isValidIPOrCIDR(s string) bool {
	return isValidIP(s) || isValidCIDR(s)
}

// isIPInCIDR 检查IP是否在CIDR范围内
func isIPInCIDR(ipStr, cidrStr string) (bool, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false, fmt.Errorf("无效的IP地址: %s", ipStr)
	}

	_, ipNet, err := net.ParseCIDR(cidrStr)
	if err != nil {
		return false, fmt.Errorf("无效的CIDR: %s", cidrStr)
	}

	return ipNet.Contains(ip), nil
}

// matchIPFuzzy 模糊匹配IP
func matchIPFuzzy(ip, pattern string) (bool, error) {
	ipParts := strings.Split(ip, ".")
	patternParts := strings.Split(pattern, ".")

	if len(ipParts) != 4 || len(patternParts) != 4 {
		return false, fmt.Errorf("IP格式错误")
	}

	for i := 0; i < 4; i++ {
		if patternParts[i] != "*" && ipParts[i] != patternParts[i] {
			return false, nil
		}
	}

	return true, nil
}

// isIPInGroup 检查IP是否在IP组中(优化版本:使用缓存)
func (e *RuleEngine) isIPInGroup(ip, groupName string) (bool, error) {
	// 检查IP组是否存在
	_, exists := e.IPGroups[groupName]
	if !exists {
		return false, fmt.Errorf("IP组不存在: %s", groupName)
	}

	// 检查缓存是否存在，如果不存在则构建
	cache, cacheExists := e.ipGroupCaches[groupName]
	if !cacheExists {
		if err := e.buildIPGroupCache(groupName); err != nil {
			return false, fmt.Errorf("构建IP组缓存失败: %v", err)
		}
		cache = e.ipGroupCaches[groupName]
	}

	// 1. 快速查找：精确IP匹配 O(1)
	if cache.exactIPs[ip] {
		return true, nil
	}

	// 2. CIDR匹配 O(n) - n是CIDR数量，通常远小于IP数量
	ipAddr := net.ParseIP(ip)
	if ipAddr == nil {
		return false, fmt.Errorf("无效的IP地址: %s", ip)
	}

	for _, ipNet := range cache.cidrNets {
		if ipNet.Contains(ipAddr) {
			return true, nil
		}
	}

	return false, nil
}

// matchRegex 正则表达式匹配（使用 LRU 缓存）
func (e *RuleEngine) matchRegex(s, pattern string) (bool, error) {
	// 验证正则表达式长度
	if len(pattern) > MaxRegexLength {
		return false, fmt.Errorf("正则表达式长度超过限制(%d字符)", MaxRegexLength)
	}

	// 尝试从缓存获取
	re, exists := e.regexCache.get(pattern)
	if !exists {
		// 编译正则表达式
		var err error
		re, err = regexp.Compile(pattern)
		if err != nil {
			return false, fmt.Errorf("无效的正则表达式: %s, 错误: %v", pattern, err)
		}
		// 存入缓存
		e.regexCache.set(pattern, re)
	}

	// Go 的 regexp 包使用 RE2 语义（线性时间匹配），不会出现灾难性回溯。
	// 此处额外使用 goroutine + channel 提供超时保护，防止极端构造下的意外阻塞。
	const matchTimeout = 100 * time.Millisecond
	type matchResult struct {
		matched bool
		err     error
	}
	resultCh := make(chan matchResult, 1)
	go func() {
		resultCh <- matchResult{matched: re.MatchString(s)}
	}()
	select {
	case r := <-resultCh:
		return r.matched, r.err
	case <-time.After(matchTimeout):
		return false, fmt.Errorf("正则匹配超时(>%v): pattern=%s", matchTimeout, pattern)
	}
}

// validateSimpleCondition 验证简单条件
func validateSimpleCondition(c *SimpleCondition) error {
	if c == nil {
		return fmt.Errorf("条件不能为nil")
	}
	if c.Type != SimpleConditionType {
		return fmt.Errorf("条件类型错误: %s", c.Type)
	}
	if c.Target == "" {
		return fmt.Errorf("目标不能为空")
	}
	if c.MatchType == "" {
		return fmt.Errorf("匹配类型不能为空")
	}
	// 大多数匹配类型都需要匹配值
	if c.MatchValue == "" {
		return fmt.Errorf("匹配值不能为空")
	}
	return nil
}

// validateCompositeCondition 验证复合条件
func validateCompositeCondition(c *CompositeCondition) error {
	if c == nil {
		return fmt.Errorf("条件不能为nil")
	}
	if c.Type != CompositeConditionType {
		return fmt.Errorf("条件类型错误: %s", c.Type)
	}
	if c.Operator != LogicalAND && c.Operator != LogicalOR {
		return fmt.Errorf("无效的逻辑运算符: %s", c.Operator)
	}
	if len(c.Conditions) == 0 {
		return fmt.Errorf("复合条件至少需要一个子条件")
	}
	return nil
}

// buildIPGroupCache 构建IP组的查找缓存
func (e *RuleEngine) buildIPGroupCache(groupName string) error {
	group, exists := e.IPGroups[groupName]
	if !exists {
		return fmt.Errorf("IP组不存在: %s", groupName)
	}

	cache := &ipGroupCache{
		exactIPs: make(map[string]bool),
		cidrNets: make([]*net.IPNet, 0),
	}

	for _, item := range group.Items {
		if isValidIP(item) {
			// 精确IP，存入map
			cache.exactIPs[item] = true
		} else if isValidCIDR(item) {
			// CIDR范围，解析并存入列表
			_, ipNet, err := net.ParseCIDR(item)
			if err != nil {
				return fmt.Errorf("解析CIDR失败: %s, 错误: %v", item, err)
			}
			cache.cidrNets = append(cache.cidrNets, ipNet)
		} else {
			return fmt.Errorf("IP组 %s 包含无效项: %s", groupName, item)
		}
	}

	e.ipGroupCaches[groupName] = cache
	return nil
}
