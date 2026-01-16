// AI分析器相关类型定义

export interface AttackPattern {
  id: string
  attack_type: string
  severity: "low" | "medium" | "high" | "critical"
  description: string
  detected_at: string
  sample_count: number
  features: {
    ip_addresses: string[]
    urls: string[]
    user_agents: string[]
  }
  statistical_data: {
    mean: number
    std_dev: number
    z_score: number
  }
  created_at: string
  updated_at: string
}

export interface GeneratedRule {
  id: string
  pattern_id: string
  rule_type: string
  rule_content: string
  confidence: number
  status: "pending" | "approved" | "rejected" | "deployed"
  reviewed_by?: string
  review_comment?: string
  deployed_at?: string
  created_at: string
  updated_at: string
}

export interface AIAnalyzerConfig {
  id?: string
  name?: string
  enabled: boolean
  
  // 模式检测配置
  patternDetection?: {
    enabled?: boolean
    minSamples?: number // 最小样本数
    anomalyThreshold?: number // 异常阈值
    clusteringMethod?: string // 聚类方法
    timeWindow?: number // 时间窗口(小时)
  }
  
  // 规则生成配置
  ruleGeneration?: {
    enabled?: boolean
    confidenceThreshold?: number // 置信度阈值
    autoDeploy?: boolean // 是否自动部署
    reviewRequired?: boolean // 是否需要审核
    defaultAction?: string // 默认动作
  }
  
  // 分析周期
  analysisInterval?: number // 分析间隔(分钟)
  
  createdAt?: string
  updatedAt?: string
}

export interface MCPConversation {
  id: string
  pattern_id?: string
  role: "user" | "assistant" | "system"
  content: string
  tool_calls?: any[]
  metadata?: Record<string, any>
  created_at: string
}

export interface AnalyzerStats {
  patterns_detected: number
  rules_generated: number
  rules_deployed: number
  rules_pending: number
  last_analysis?: string
}

// 请求参数类型
export interface AttackPatternListParams {
  page?: number
  size?: number
  severity?: string
  attack_type?: string
  start_date?: string
  end_date?: string
}

export interface GeneratedRuleListParams {
  page?: number
  size?: number
  status?: string
  pattern_id?: string
}

export interface MCPConversationListParams {
  page?: number
  size?: number
  pattern_id?: string
}

export interface RuleReviewRequest {
  status: "approved" | "rejected"
  comment?: string
}

// 响应类型
export interface AttackPatternListResponse {
  data: AttackPattern[]
  total: number
  page: number
  size: number
}

export interface GeneratedRuleListResponse {
  data: GeneratedRule[]
  total: number
  page: number
  size: number
}

export interface MCPConversationListResponse {
  data: MCPConversation[]
  total: number
  page: number
  size: number
}

export interface AIAnalyzerConfigResponse {
  data: AIAnalyzerConfig
}

export interface AnalyzerStatsResponse {
  data: AnalyzerStats
}
