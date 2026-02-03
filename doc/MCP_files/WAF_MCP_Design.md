# WAF MCP服务器完整设计文档

## 一、系统架构概览

```
┌─────────────────────────────────────────────────────────────────┐
│                     MCP Client (Claude/App)                      │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│              WAF MCP Server (Transport: HTTP/stdio)              │
├─────────────────────────────────────────────────────────────────┤
│  Tool Handler & Router Layer                                     │
├─────────────────────────────────────────────────────────────────┤
│  Core Service Layer                                              │
│  ├── IP黑名单服务                                               │
│  ├── 规则管理服务                                               │
│  ├── 攻击分析服务                                               │
│  └── 系统监控服务                                               │
├─────────────────────────────────────────────────────────────────┤
│  API Client & Authentication Layer                              │
├─────────────────────────────────────────────────────────────────┤
│  Error Handling & Response Formatting                           │
└─────────────────────────────────────────────────────────────────┘
                             │
                             ▼
                  ┌──────────────────────┐
                  │   WAF Backend API    │
                  │   (REST Endpoints)   │
                  └──────────────────────┘
```

---

## 二、核心工具列表（按功能分类）

### 2.1 IP黑名单管理工具

#### 2.1.1 `waf_block_ip` - 添加单个IP到黑名单
- **描述**: 将恶意IP地址添加到黑名单，可设置永久或临时封禁
- **参数**:
  - `ip_address` (string): IP地址，支持IPv4/IPv6
  - `reason` (string): 封禁原因（必需）
  - `duration_seconds` (integer, optional): 封禁时长(秒)，0表示永久，默认3600
  - `severity` (enum): 严重级别 [critical, high, medium, low]
  - `tags` (array[string], optional): 标签，便于后续查询
- **返回值**: BlockIPResponse
- **权限**: 写权限
- **幂等性**: 否
- **销毁性**: 是

#### 2.1.2 `waf_batch_block_ips` - 批量添加IP到黑名单
- **描述**: 批量添加多个IP到黑名单
- **参数**:
  - `ips` (array[string]): IP地址列表，最多1000条
  - `reason` (string): 统一的封禁原因
  - `duration_seconds` (integer, optional): 统一的封禁时长
  - `severity` (enum): 统一的严重级别
- **返回值**: BatchBlockResponse
- **权限**: 写权限
- **幂等性**: 否
- **销毁性**: 是

#### 2.1.3 `waf_unblock_ip` - 从黑名单移除IP
- **描述**: 将IP地址从黑名单中移除
- **参数**:
  - `ip_address` (string): 要移除的IP地址
  - `reason` (string, optional): 移除原因，便于审计
- **返回值**: UnblockIPResponse
- **权限**: 写权限
- **幂等性**: 否
- **销毁性**: 是

#### 2.1.4 `waf_batch_unblock_ips` - 批量移除IP
- **描述**: 批量从黑名单中移除IP地址
- **参数**:
  - `ips` (array[string]): IP地址列表
  - `reason` (string, optional): 统一的移除原因
- **返回值**: BatchUnblockResponse
- **权限**: 写权限
- **幂等性**: 否
- **销毁性**: 是

#### 2.1.5 `waf_list_blocked_ips` - 查询黑名单
- **描述**: 分页查询当前黑名单中的所有IP
- **参数**:
  - `page` (integer, optional): 页码，从1开始，默认1
  - `page_size` (integer, optional): 每页数量，默认20，最大100
  - `filter` (object, optional):
    - `severity`: 按严重级别过滤
    - `tags`: 按标签过滤
    - `reason_contains`: 按原因关键词过滤
    - `created_after`: 按创建时间过滤
  - `sort_by` (enum, optional): 排序字段 [created_time, updated_time, severity]
  - `sort_order` (enum, optional): 排序顺序 [asc, desc]
- **返回值**: BlockedIPListResponse (包含分页信息)
- **权限**: 读权限
- **幂等性**: 是

#### 2.1.6 `waf_get_blocked_ip_stats` - 黑名单统计
- **描述**: 获取黑名单的统计信息
- **参数**: 无
- **返回值**: BlockedIPStatsResponse
  - `total_blocked_count`: 历史封禁总数
  - `active_blocks_count`: 当前活跃封禁数
  - `stats_by_severity`: 按严重级别的统计
  - `stats_by_reason`: 按原因的统计
  - `recent_blocks`: 最近30条封禁记录
- **权限**: 读权限
- **幂等性**: 是

---

### 2.2 规则管理工具

#### 2.2.1 `waf_create_rule` - 创建单条防护规则
- **描述**: 创建新的MicroRule防护规则
- **参数**:
  - `rule_name` (string): 规则名称，唯一标识
  - `rule_type` (enum): 规则类型 [blacklist, whitelist, pattern, rate_limit, geo_block]
  - `action` (enum): 规则动作 [block, allow, log, challenge]
  - `conditions` (object): 规则条件，支持以下字段:
    - `method` (array[string]): HTTP方法 [GET, POST, PUT, DELETE, PATCH]
    - `path` (string): 请求路径，支持通配符
    - `source_ip` (string): 源IP或IP段
    - `user_agent` (string): User-Agent匹配规则
    - `request_body_contains` (string): 请求体包含的内容
    - `response_code` (array[integer]): HTTP响应码
    - `country_code` (array[string]): 国家代码
    - `rate_limit` (object): 速率限制
      - `requests_per_minute` (integer)
      - `window_seconds` (integer)
  - `priority` (integer): 优先级，1-10000，数字越大优先级越高
  - `enabled` (boolean, optional): 是否启用，默认true
  - `description` (string, optional): 规则描述
  - `tags` (array[string], optional): 标签
- **返回值**: RuleResponse
- **权限**: 写权限
- **幂等性**: 否
- **销毁性**: 否

#### 2.2.2 `waf_batch_create_rules` - 批量创建规则
- **描述**: 一次性创建多条规则
- **参数**:
  - `rules` (array[RuleCreateRequest]): 规则列表，最多100条
- **返回值**: BatchRuleResponse
- **权限**: 写权限
- **幂等性**: 否
- **销毁性**: 否

#### 2.2.3 `waf_update_rule` - 更新规则
- **描述**: 更新已存在的规则
- **参数**:
  - `rule_id` (string): 规则ID
  - `rule_name` (string, optional): 新规则名称
  - `action` (enum, optional): 新动作
  - `conditions` (object, optional): 新条件
  - `priority` (integer, optional): 新优先级
  - `enabled` (boolean, optional): 启用/禁用
  - `description` (string, optional): 新描述
- **返回值**: RuleResponse
- **权限**: 写权限
- **幂等性**: 否
- **销毁性**: 否

#### 2.2.4 `waf_delete_rule` - 删除规则
- **描述**: 删除指定的规则
- **参数**:
  - `rule_id` (string): 规则ID
  - `force` (boolean, optional): 是否强制删除，默认false
- **返回值**: DeleteRuleResponse
- **权限**: 写权限
- **幂等性**: 否
- **销毁性**: 是

#### 2.2.5 `waf_batch_delete_rules` - 批量删除规则
- **描述**: 批量删除规则
- **参数**:
  - `rule_ids` (array[string]): 规则ID列表
  - `force` (boolean, optional): 是否强制删除
- **返回值**: BatchDeleteRuleResponse
- **权限**: 写权限
- **幂等性**: 否
- **销毁性**: 是

#### 2.2.6 `waf_list_rules` - 查询规则列表
- **描述**: 查询所有规则，支持过滤和排序
- **参数**:
  - `page` (integer, optional): 页码
  - `page_size` (integer, optional): 每页数量
  - `filter` (object, optional):
    - `rule_type`: 按类型过滤
    - `action`: 按动作过滤
    - `enabled`: 按启用状态过滤
    - `tags`: 按标签过滤
    - `name_contains`: 按名称过滤
  - `sort_by` (enum, optional): [created_time, priority, name]
  - `sort_order` (enum, optional): [asc, desc]
- **返回值**: RuleListResponse
- **权限**: 读权限
- **幂等性**: 是

#### 2.2.7 `waf_get_rule_details` - 获取规则详情
- **描述**: 获取单条规则的完整详情
- **参数**:
  - `rule_id` (string): 规则ID
- **返回值**: RuleDetailResponse
- **权限**: 读权限
- **幂等性**: 是

#### 2.2.8 `waf_export_rules` - 导出规则
- **描述**: 导出规则为JSON/YAML格式
- **参数**:
  - `format` (enum): [json, yaml]
  - `filter` (object, optional): 导出过滤条件
  - `include_disabled` (boolean, optional): 是否包含禁用的规则
- **返回值**: ExportRulesResponse (包含规则文本)
- **权限**: 读权限
- **幂等性**: 是

#### 2.2.9 `waf_import_rules` - 导入规则
- **描述**: 从JSON/YAML导入规则
- **参数**:
  - `rules_content` (string): 规则文本内容
  - `format` (enum): [json, yaml]
  - `merge_mode` (enum, optional): [replace, merge, skip_duplicate]
- **返回值**: ImportRulesResponse
- **权限**: 写权限
- **幂等性**: 否
- **销毁性**: 是

---

### 2.3 攻击分析与检测工具

#### 2.3.1 `waf_list_attack_logs` - 查询攻击日志
- **描述**: 查询WAF检测到的攻击日志
- **参数**:
  - `page` (integer, optional): 页码
  - `page_size` (integer, optional): 每页数量
  - `time_range` (object, optional):
    - `start_time` (datetime): 开始时间
    - `end_time` (datetime): 结束时间
  - `filter` (object, optional):
    - `attack_type`: [sql_injection, xss, path_traversal, cmd_injection, etc]
    - `severity`: [critical, high, medium, low]
    - `source_ip`: 源IP
    - `target_path`: 目标路径
    - `blocked`: 是否已阻止
  - `sort_by` (enum, optional): [timestamp, severity]
- **返回值**: AttackLogListResponse
- **权限**: 读权限
- **幂等性**: 是

#### 2.3.2 `waf_analyze_attack_patterns` - 分析攻击模式
- **描述**: 基于历史日志分析攻击模式，支持AI自动分析
- **参数**:
  - `time_range` (object):
    - `hours`: 分析最近N小时的数据
  - `clustering_method` (enum, optional): [kmeans, dbscan] 默认kmeans
  - `min_samples` (integer, optional): 最小样本数，默认10
  - `anomaly_threshold` (number, optional): 异常检测阈值，默认2.0
- **返回值**: AttackPatternAnalysisResponse
  - `patterns` (array): 检测到的攻击模式
  - `anomalies` (array): 异常行为
  - `statistics` (object): 统计信息
  - `recommendations` (array): 建议的防护规则
- **权限**: 读权限
- **幂等性**: 是

#### 2.3.3 `waf_generate_rule_from_pattern` - 根据攻击模式生成规则
- **描述**: 根据检测到的攻击模式自动生成防护规则
- **参数**:
  - `pattern_id` (string): 攻击模式ID
  - `action` (enum, optional): [block, log] 默认block
  - `priority` (integer, optional): 优先级，默认100
  - `auto_review` (boolean, optional): 是否自动审核，默认false
  - `rule_type` (enum, optional): [micro_rule, modsecurity] 默认micro_rule
- **返回值**: GeneratedRuleResponse
- **权限**: 写权限
- **幂等性**: 否
- **销毁性**: 否

#### 2.3.4 `waf_review_generated_rule` - 审核AI生成的规则
- **描述**: 批准或拒绝AI自动生成的规则
- **参数**:
  - `rule_id` (string): 待审核的规则ID
  - `action` (enum): [approve, reject]
  - `comment` (string, optional): 审核意见
  - `suggested_modifications` (object, optional): 建议的修改
- **返回值**: ReviewRuleResponse
- **权限**: 写权限
- **幂等性**: 否
- **销毁性**: 否

#### 2.3.5 `waf_deploy_generated_rule` - 部署已审核的规则
- **描述**: 将已批准的规则部署到生产环境
- **参数**:
  - `rule_id` (string): 规则ID
  - `test_first` (boolean, optional): 是否先在测试环境部署，默认true
  - `deployment_strategy` (enum, optional): [immediate, gradual, scheduled]
  - `schedule_time` (datetime, optional): 计划部署时间
- **返回值**: DeploymentResponse
- **权限**: 写权限
- **幂等性**: 否
- **销毁性**: 否

---

### 2.4 系统监控与统计工具

#### 2.4.1 `waf_get_stats_overview` - 获取统计概览
- **描述**: 获取WAF的实时统计数据概览
- **参数**:
  - `time_range` (enum, optional): [1h, 6h, 24h, 7d, 30d] 默认24h
- **返回值**: StatsOverviewResponse
  - `total_requests`: 总请求数
  - `blocked_requests`: 被阻止的请求数
  - `block_rate` (percent): 阻止率
  - `error_rate` (percent): 错误率
  - `top_attack_types` (array): 排名前5的攻击类型
  - `top_blocked_ips` (array): 排名前5的被阻止IP
- **权限**: 读权限
- **幂等性**: 是

#### 2.4.2 `waf_get_time_series_data` - 获取时间序列数据
- **描述**: 获取指定时间范围内的时间序列监控数据
- **参数**:
  - `metric_type` (enum): [requests, errors, response_time, blocked_rate]
  - `time_range` (enum): [1h, 6h, 24h, 7d, 30d]
  - `interval` (enum, optional): [1m, 5m, 1h, 1d] 自动推荐
- **返回值**: TimeSeriesResponse
  - `data_points` (array): 时间序列数据点
  - `statistics` (object): 统计信息（平均值、最大值、最小值）
- **权限**: 读权限
- **幂等性**: 是

#### 2.4.3 `waf_get_security_metrics` - 获取安全指标
- **描述**: 获取安全相关的指标和评分
- **参数**:
  - `time_range` (enum, optional): [24h, 7d, 30d]
- **返回值**: SecurityMetricsResponse
  - `threat_level`: 威胁等级 [critical, high, medium, low]
  - `detected_threats`: 检测到的威胁数量
  - `prevented_breaches`: 阻止的可能突破次数
  - `vulnerability_score`: 漏洞评分（0-100）
  - `rule_effectiveness`: 规则有效性
- **权限**: 读权限
- **幂等性**: 是

#### 2.4.4 `waf_get_realtime_qps` - 实时QPS监控
- **描述**: 获取实时的QPS（每秒请求数）数据
- **参数**:
  - `limit` (integer, optional): 返回最近N个数据点，默认30，最大60
- **返回值**: RealtimeQPSResponse
  - `current_qps`: 当前QPS
  - `peak_qps`: 峰值QPS
  - `average_qps`: 平均QPS
  - `data_points` (array): 实时数据点列表
- **权限**: 读权限
- **幂等性**: 是

#### 2.4.5 `waf_compare_rules` - 对比两条规则的效果
- **描述**: 对比两条规则的性能和安全效果
- **参数**:
  - `rule_id_1` (string): 规则1 ID
  - `rule_id_2` (string): 规则2 ID
  - `time_range` (enum, optional): [24h, 7d, 30d]
- **返回值**: RuleComparisonResponse
  - `rule_1_stats`: 规则1的统计数据
  - `rule_2_stats`: 规则2的统计数据
  - `comparison_metrics`: 对比指标
  - `recommendation`: 推荐结果
- **权限**: 读权限
- **幂等性**: 是

---

### 2.5 规则优化与调试工具

#### 2.5.1 `waf_evaluate_rule_effectiveness` - 评估规则有效性
- **描述**: 评估指定规则的有效性和影响
- **参数**:
  - `rule_id` (string): 规则ID
  - `time_range` (enum, optional): [1h, 6h, 24h, 7d] 默认24h
- **返回值**: RuleEffectivenessResponse
  - `false_positive_rate`: 误报率
  - `true_positive_rate`: 真正率
  - `blocking_rate`: 阻止率
  - `affected_requests`: 影响的请求数
  - `improvement_suggestions`: 改进建议
- **权限**: 读权限
- **幂等性**: 是

#### 2.5.2 `waf_optimize_rule` - 优化规则
- **描述**: 根据历史数据自动优化规则参数
- **参数**:
  - `rule_id` (string): 规则ID
  - `optimize_for` (enum, optional): [accuracy, performance, both] 默认both
  - `keep_history` (boolean, optional): 是否保留优化历史，默认true
- **返回值**: RuleOptimizationResponse
  - `optimized_rule`: 优化后的规则
  - `changes_made`: 做出的改变
  - `performance_improvement`: 性能提升百分比
  - `accuracy_improvement`: 准确性提升百分比
- **权限**: 写权限
- **幂等性**: 否
- **销毁性**: 否

#### 2.5.3 `waf_test_rule` - 测试规则
- **描述**: 在测试环境中测试规则效果，不实际应用
- **参数**:
  - `rule_id` (string): 规则ID（或rule_config用于新规则）
  - `test_payload` (string): 测试负载/请求
  - `expected_action` (enum): [block, allow]
  - `test_count` (integer, optional): 重复测试次数
- **返回值**: RuleTestResponse
  - `passed` (boolean): 测试是否通过
  - `actual_action`: 实际动作
  - `execution_time`: 执行时间（ms）
  - `detailed_results` (array): 详细结果
- **权限**: 读权限
- **幂等性**: 是

#### 2.5.4 `waf_get_rule_logs` - 获取规则执行日志
- **描述**: 查询指定规则的执行日志
- **参数**:
  - `rule_id` (string): 规则ID
  - `time_range` (object): 时间范围
  - `filter` (object, optional):
    - `action`: [triggered, passed, error]
    - `request_ip`: 源IP
  - `limit` (integer, optional): 最大返回数量，默认100
- **返回值**: RuleLogListResponse
- **权限**: 读权限
- **幂等性**: 是

---

### 2.6 配置管理工具

#### 2.6.1 `waf_get_config` - 获取WAF配置
- **描述**: 获取WAF系统的完整配置
- **参数**: 无
- **返回值**: WAFConfigResponse
- **权限**: 读权限
- **幂等性**: 是

#### 2.6.2 `waf_update_config` - 更新WAF配置
- **描述**: 更新WAF系统配置（部分更新）
- **参数**:
  - `enable_ai_analysis` (boolean, optional): 启用AI分析
  - `enable_rate_limit` (boolean, optional): 启用速率限制
  - `enable_geo_ip` (boolean, optional): 启用GeoIP
  - `rate_limit_rpm` (integer, optional): 每分钟请求限制
  - `max_request_body_size` (integer, optional): 最大请求体大小（字节）
  - `block_duration` (integer, optional): 默认封禁时长（秒）
  - `log_level` (enum, optional): [debug, info, warn, error]
- **返回值**: WAFConfigResponse
- **权限**: 写权限
- **幂等性**: 否
- **销毁性**: 否

#### 2.6.3 `waf_get_system_health` - 获取系统健康状态
- **描述**: 获取WAF系统的健康状态
- **参数**: 无
- **返回值**: SystemHealthResponse
  - `status`: [healthy, degraded, unhealthy]
  - `uptime`: 运行时长
  - `cpu_usage` (percent)
  - `memory_usage` (percent)
  - `disk_usage` (percent)
  - `services_status` (array): 各个服务的状态
  - `last_error` (optional): 最后一个错误
- **权限**: 读权限
- **幂等性**: 是

---

## 三、数据结构定义

### 3.1 IP黑名单相关数据结构

```typescript
// 单条IP记录
interface BlockedIP {
  id: string;
  ip_address: string;
  reason: string;
  severity: 'critical' | 'high' | 'medium' | 'low';
  duration_seconds: number; // 0表示永久
  created_at: datetime;
  updated_at: datetime;
  expires_at?: datetime;
  tags: string[];
  blocked_requests_count: number;
  last_attack_timestamp?: datetime;
}

// 黑名单响应
interface BlockedIPListResponse {
  items: BlockedIP[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };
  summary: {
    total_blocked_count: number;
    active_blocks_count: number;
  };
}

// 黑名单统计
interface BlockedIPStatsResponse {
  total_blocked_count: number;
  active_blocks_count: number;
  stats_by_severity: {
    critical: number;
    high: number;
    medium: number;
    low: number;
  };
  stats_by_reason: Map<string, number>;
  recent_blocks: BlockedIP[];
}
```

### 3.2 规则相关数据结构

```typescript
// 规则条件
interface RuleConditions {
  method?: string[];
  path?: string;
  source_ip?: string;
  user_agent?: string;
  request_body_contains?: string;
  response_code?: number[];
  country_code?: string[];
  rate_limit?: {
    requests_per_minute: number;
    window_seconds: number;
  };
  [key: string]: any; // 支持自定义条件
}

// 规则定义
interface WafRule {
  id: string;
  rule_name: string;
  rule_type: 'blacklist' | 'whitelist' | 'pattern' | 'rate_limit' | 'geo_block';
  action: 'block' | 'allow' | 'log' | 'challenge';
  conditions: RuleConditions;
  priority: number;
  enabled: boolean;
  description?: string;
  tags: string[];
  created_at: datetime;
  updated_at: datetime;
  created_by: string;
  updated_by: string;
  is_system_rule: boolean;
  effectiveness_stats?: {
    true_positive_rate: number;
    false_positive_rate: number;
    block_rate: number;
  };
}

// 规则列表响应
interface RuleListResponse {
  items: WafRule[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };
}
```

### 3.3 攻击日志数据结构

```typescript
interface AttackLog {
  id: string;
  attack_type: string;
  severity: 'critical' | 'high' | 'medium' | 'low';
  source_ip: string;
  target_path: string;
  target_domain: string;
  request_method: string;
  request_body?: string;
  response_code: number;
  matched_rule_id?: string;
  matched_rule_name?: string;
  action_taken: 'blocked' | 'logged' | 'allowed';
  timestamp: datetime;
  user_agent?: string;
  referer?: string;
  request_id: string;
  geo_location?: {
    country: string;
    city: string;
    latitude: number;
    longitude: number;
  };
}
```

### 3.4 攻击模式分析结果

```typescript
interface AttackPattern {
  pattern_id: string;
  attack_type: string;
  frequency: number;
  affected_paths: string[];
  affected_ips: string[];
  severity_score: number; // 0-100
  first_detected: datetime;
  last_detected: datetime;
  characteristics: {
    [key: string]: any;
  };
  recommended_action: string;
  confidence_score: number; // 0-100
}

interface AttackPatternAnalysisResponse {
  analysis_period: {
    start_time: datetime;
    end_time: datetime;
  };
  patterns: AttackPattern[];
  anomalies: any[];
  statistics: {
    total_attacks: number;
    unique_attackers: number;
    attack_types: Map<string, number>;
  };
  recommendations: string[];
  suggested_rules: WafRule[];
}
```

---

## 四、API错误处理与标准化响应

### 4.1 标准响应格式

```typescript
// 成功响应
interface SuccessResponse<T> {
  success: true;
  data: T;
  timestamp: datetime;
  request_id: string;
  message?: string;
}

// 错误响应
interface ErrorResponse {
  success: false;
  error: {
    code: string;
    message: string;
    details?: {
      [key: string]: any;
    };
  };
  timestamp: datetime;
  request_id: string;
}
```

### 4.2 错误代码映射

```
WAF_BLOCK_IP_FAILED          - IP封禁失败
WAF_RULE_CREATE_FAILED        - 规则创建失败
WAF_RULE_UPDATE_FAILED        - 规则更新失败
WAF_INVALID_PARAMETER         - 参数验证失败
WAF_RULE_NOT_FOUND            - 规则不存在
WAF_IP_NOT_FOUND              - IP不存在
WAF_PERMISSION_DENIED         - 权限不足
WAF_OPERATION_TIMEOUT         - 操作超时
WAF_BACKEND_ERROR             - 后端服务错误
WAF_CONFLICT                  - 冲突（如IP已存在）
```

---

## 五、认证与授权

### 5.1 认证方式
- API Key认证
- Bearer Token认证
- 基于角色的访问控制（RBAC）

### 5.2 权限模型
- `waf:read` - 读权限（查询数据）
- `waf:write` - 写权限（修改数据）
- `waf:admin` - 管理权限（系统配置）
- `waf:audit` - 审计权限（查看审计日志）

---

## 六、服务层实现要点

### 6.1 IP黑名单服务（IPBlacklistService）

```python
class IPBlacklistService:
    """IP黑名单管理服务"""
    
    def block_ip(self, ip_address: str, reason: str, 
                 duration_seconds: int = 3600, 
                 severity: str = 'medium',
                 tags: List[str] = None) -> BlockedIP:
        """添加IP到黑名单"""
        # 1. 参数验证 - IP格式、重复检查
        # 2. 前置检查 - IP是否已在黑名单中
        # 3. 数据库操作 - 插入新记录
        # 4. 缓存更新 - 更新内存黑名单缓存
        # 5. WAF引擎更新 - 通知WAF引擎
        # 6. 审计日志 - 记录操作
        # 7. 返回结果
        
    def unblock_ip(self, ip_address: str, reason: str = None) -> bool:
        """从黑名单移除IP"""
        # 1. 检查IP是否存在
        # 2. 删除记录
        # 3. 更新缓存
        # 4. 通知WAF引擎
        # 5. 审计日志
        
    def list_blocked_ips(self, page: int = 1, page_size: int = 20,
                        filter: Dict = None) -> BlockedIPListResponse:
        """查询黑名单"""
        # 1. 构建查询条件
        # 2. 数据库查询
        # 3. 计算分页信息
        # 4. 返回结果
        
    def get_stats(self) -> BlockedIPStatsResponse:
        """获取黑名单统计"""
        # 1. 计算各项统计指标
        # 2. 按严重级别分类
        # 3. 按原因分类
        # 4. 返回统计结果
```

### 6.2 规则管理服务（RuleManagementService）

```python
class RuleManagementService:
    """规则管理服务"""
    
    def create_rule(self, rule_name: str, rule_type: str, 
                   action: str, conditions: Dict,
                   priority: int, enabled: bool = True) -> WafRule:
        """创建新规则"""
        # 1. 参数验证
        # 2. 检查重复 - 同名规则
        # 3. 优先级冲突检查
        # 4. 条件语法校验
        # 5. 数据库保存
        # 6. 缓存更新
        # 7. 审计日志
        # 8. 返回结果
        
    def update_rule(self, rule_id: str, 
                   updates: Dict) -> WafRule:
        """更新规则"""
        # 1. 检查规则存在性
        # 2. 验证更新内容
        # 3. 版本控制 - 保存旧版本
        # 4. 更新数据库
        # 5. 缓存更新
        # 6. 自动测试 - 运行单元测试
        # 7. 审计日志
        
    def delete_rule(self, rule_id: str, force: bool = False) -> bool:
        """删除规则"""
        # 1. 检查规则是否被依赖
        # 2. 如果force=False，需要确认
        # 3. 从数据库删除
        # 4. 缓存更新
        # 5. 审计日志
        
    def list_rules(self, page: int = 1, page_size: int = 20,
                  filter: Dict = None) -> RuleListResponse:
        """查询规则列表"""
        
    def test_rule(self, rule_id: str, test_payload: str,
                 expected_action: str) -> RuleTestResponse:
        """测试规则"""
        # 1. 加载规则
        # 2. 在隔离环境中执行规则
        # 3. 检查是否匹配预期结果
        # 4. 返回测试结果
```

### 6.3 攻击分析服务（AttackAnalysisService）

```python
class AttackAnalysisService:
    """攻击分析和检测服务"""
    
    def analyze_patterns(self, time_range_hours: int,
                        clustering_method: str = 'kmeans') -> AttackPatternAnalysisResponse:
        """分析攻击模式"""
        # 1. 获取时间范围内的所有攻击日志
        # 2. 数据预处理 - 特征提取
        # 3. 聚类分析 - 使用指定算法
        # 4. 异常检测 - Isolation Forest或其他算法
        # 5. 模式识别 - 识别常见攻击模式
        # 6. 生成建议规则
        # 7. 返回分析结果
        
    def generate_rule_from_pattern(self, pattern_id: str,
                                  action: str = 'block') -> GeneratedRuleResponse:
        """根据攻击模式生成防护规则"""
        # 1. 获取模式详情
        # 2. 分析模式特征
        # 3. 生成规则条件
        # 4. 评估规则准确性
        # 5. 返回生成的规则
        
    def suggest_improvements(self, rule_id: str) -> List[str]:
        """建议规则改进"""
        # 基于误报率、漏报率等指标提建议
```

### 6.4 监控统计服务（MonitoringService）

```python
class MonitoringService:
    """监控和统计服务"""
    
    def get_stats_overview(self, time_range: str = '24h') -> StatsOverviewResponse:
        """获取统计概览"""
        # 1. 查询指定时间范围内的统计数据
        # 2. 计算各项指标
        # 3. 排序TOP数据
        # 4. 返回概览
        
    def get_time_series_data(self, metric_type: str,
                            time_range: str,
                            interval: str = None) -> TimeSeriesResponse:
        """获取时间序列数据"""
        # 1. 智能推荐时间间隔
        # 2. 从数据库查询数据
        # 3. 聚合数据（如果需要）
        # 4. 计算统计信息
        
    def get_realtime_qps(self, limit: int = 30) -> RealtimeQPSResponse:
        """获取实时QPS"""
        # 从实时监控系统获取数据
```

---

## 七、数据库设计

### 7.1 表结构

```sql
-- IP黑名单表
CREATE TABLE blocked_ips (
    id VARCHAR(36) PRIMARY KEY,
    ip_address VARCHAR(45) NOT NULL UNIQUE,
    reason VARCHAR(255),
    severity ENUM('critical', 'high', 'medium', 'low'),
    duration_seconds INT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    expires_at TIMESTAMP,
    tags JSON,
    blocked_requests_count INT DEFAULT 0,
    last_attack_timestamp TIMESTAMP,
    created_by VARCHAR(255),
    INDEX idx_created_at (created_at),
    INDEX idx_severity (severity)
);

-- 防护规则表
CREATE TABLE waf_rules (
    id VARCHAR(36) PRIMARY KEY,
    rule_name VARCHAR(255) NOT NULL UNIQUE,
    rule_type VARCHAR(50),
    action VARCHAR(20),
    conditions JSON,
    priority INT,
    enabled BOOLEAN,
    description TEXT,
    tags JSON,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    is_system_rule BOOLEAN DEFAULT false,
    version INT DEFAULT 1,
    INDEX idx_priority (priority),
    INDEX idx_enabled (enabled),
    INDEX idx_rule_type (rule_type)
);

-- 攻击日志表
CREATE TABLE attack_logs (
    id VARCHAR(36) PRIMARY KEY,
    attack_type VARCHAR(100),
    severity VARCHAR(20),
    source_ip VARCHAR(45),
    target_path VARCHAR(500),
    target_domain VARCHAR(255),
    request_method VARCHAR(10),
    request_body LONGTEXT,
    response_code INT,
    matched_rule_id VARCHAR(36),
    matched_rule_name VARCHAR(255),
    action_taken VARCHAR(20),
    timestamp TIMESTAMP,
    user_agent VARCHAR(500),
    request_id VARCHAR(36),
    geo_location JSON,
    INDEX idx_timestamp (timestamp),
    INDEX idx_source_ip (source_ip),
    INDEX idx_attack_type (attack_type),
    FOREIGN KEY (matched_rule_id) REFERENCES waf_rules(id)
);

-- 审计日志表
CREATE TABLE audit_logs (
    id VARCHAR(36) PRIMARY KEY,
    operation_type VARCHAR(50),
    resource_type VARCHAR(50),
    resource_id VARCHAR(36),
    operator_id VARCHAR(255),
    changes JSON,
    timestamp TIMESTAMP,
    status VARCHAR(20),
    INDEX idx_timestamp (timestamp),
    INDEX idx_resource_type (resource_type)
);
```

---

## 八、实现优先级

### Phase 1：核心功能（第一阶段）
- [x] IP黑名单基础管理（block_ip, unblock_ip）
- [x] 规则基础管理（create_rule, update_rule, delete_rule）
- [x] 统计概览（get_stats_overview）
- [x] 攻击日志查询（list_attack_logs）

### Phase 2：增强功能（第二阶段）
- [ ] 批量操作（batch_block_ips, batch_create_rules等）
- [ ] 规则优化（optimize_rule, evaluate_rule_effectiveness）
- [ ] 攻击模式分析（analyze_attack_patterns）
- [ ] 实时监控（get_realtime_qps, get_time_series_data）

### Phase 3：AI与自动化（第三阶段）
- [ ] AI自动规则生成（generate_rule_from_pattern）
- [ ] 规则审核流程（review_generated_rule）
- [ ] 智能建议（improvement suggestions）
- [ ] 自动部署（deploy_with_strategy）

### Phase 4：高级功能（第四阶段）
- [ ] 规则导入导出（import_rules, export_rules）
- [ ] 规则版本控制
- [ ] A/B测试支持
- [ ] 高级分析报告

---

## 九、与现有WAF API的映射

### 现有API问题分析
1. ❌ `batch_block_ips` 返回错误 → 需要修复后端实现
2. ❌ `create_micro_rule` 参数验证过于严格 → 需要改进参数定义
3. ❌ `list_micro_rules` 数据格式错误 → 需要修复序列化逻辑

### 改进方案
1. 建立中间层适配器，转换新旧API
2. 添加重试机制和错误恢复
3. 实现本地缓存，降低依赖性
4. 添加详细的错误诊断和日志

---

## 十、部署与集成

### 10.1 MCP Server启动

```bash
# Python方式
pip install fastmcp
python waf_mcp_server.py

# TypeScript方式
npm install
npm run build
npm start
```

### 10.2 客户端集成

```python
# Claude SDK中集成
from anthropic import Anthropic

client = Anthropic()

# 配置MCP服务器
response = client.messages.create(
    model="claude-opus-4.5",
    max_tokens=4096,
    tools=[
        # 工具定义会自动从MCP服务器加载
    ],
    messages=[...],
)
```

---

## 十一、测试策略

### 单元测试覆盖
- IP黑名单CRUD操作
- 规则验证逻辑
- 条件匹配引擎

### 集成测试覆盖
- 端到端API调用
- 数据库操作
- 缓存同步

### E2E测试场景
- 完整的攻击检测和防护流程
- 规则生成和部署流程
- 大规模数据处理性能

---

## 十二、安全考虑

- [ ] API认证与授权
- [ ] 输入参数验证和清理
- [ ] SQL注入防护
- [ ] 审计日志记录
- [ ] 敏感数据脱敏
- [ ] 速率限制

---

## 十三、监控与告警

- [ ] API调用延迟告警
- [ ] 规则匹配失败告警
- [ ] 异常访问模式告警
- [ ] 系统资源告警
- [ ] 规则冲突告警
