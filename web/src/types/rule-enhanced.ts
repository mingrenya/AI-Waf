// 规则模板类型定义
export interface RuleTemplate {
  id: string
  name: string
  category: string
  description: string
  severity: "low" | "medium" | "high" | "critical"
  rule_type: "whitelist" | "blacklist"
  priority: number
  tags: string[]
  condition: Record<string, unknown>
  created_at: string
  updated_at: string
}

// 规则有效性评分
export interface RuleEffectivenessScore {
  id: string
  rule_id: string
  rule_name: string
  score: number
  match_count: number
  block_count: number
  false_positive: number
  true_positive: number
  false_positive_rate: number
  true_positive_rate: number
  block_rate: number
  avg_match_time: number
  performance_impact: "low" | "medium" | "high"
  recommendation: string
  last_evaluated: string
  evaluation_period: "24h" | "7d" | "30d"
}

// 保护配置文件
export interface ProtectionProfile {
  id: string
  name: string
  level: "basic" | "standard" | "strict"
  description: string
  categories: string[]
  template_ids: string[]
  is_default: boolean
  created_at: string
  updated_at: string
}

// API 请求/响应类型
export interface RuleTemplateListResponse {
  total: number
  items: RuleTemplate[]
}

export interface RuleEffectivenessScoreListResponse {
  total: number
  items: RuleEffectivenessScore[]
}

export interface ProtectionProfileListResponse {
  total: number
  items: ProtectionProfile[]
}

export interface ApplyProfileResponse {
  created_count: number
  message: string
  rule_names?: string[]
}

// OWASP Top 10 分类映射
export const OWASP_CATEGORIES = {
  broken_access_control: "A01:2021 - 失效的访问控制",
  cryptographic_failures: "A02:2021 - 加密机制失效",
  injection: "A03:2021 - 注入",
  insecure_design: "A04:2021 - 不安全设计",
  security_misconfiguration: "A05:2021 - 安全配置错误",
  vulnerable_components: "A06:2021 - 易受攻击和过时的组件",
  authentication_failures: "A07:2021 - 识别和身份验证失败",
  data_integrity_failures: "A08:2021 - 软件和数据完整性故障",
  logging_failures: "A09:2021 - 安全日志和监控失败",
  ssrf: "A10:2021 - 服务器端请求伪造 (SSRF)"
} as const

// 严重等级映射
export const SEVERITY_LABELS = {
  low: "低",
  medium: "中",
  high: "高",
  critical: "严重"
} as const

// 保护级别映射
export const PROTECTION_LEVELS = {
  basic: "基础保护",
  standard: "标准保护",
  strict: "严格保护"
} as const
