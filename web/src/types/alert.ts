// 告警相关类型定义

/**
 * 告警通道类型
 */
export enum AlertChannelType {
    Webhook = "webhook",
    Slack = "slack",
    Discord = "discord",
    DingTalk = "dingtalk",
    WeCom = "wecom"
}

/**
 * 告警严重等级
 */
export enum AlertSeverity {
    Low = "low",
    Medium = "medium",
    High = "high",
    Critical = "critical"
}

/**
 * 条件运算符
 */
export enum ConditionOperator {
    GreaterThan = "gt",
    GreaterThanOrEqual = "gte",
    LessThan = "lt",
    LessThanOrEqual = "lte",
    Equal = "eq",
    NotEqual = "ne",
    Contains = "contains"
}

/**
 * 告警通道配置
 */
export interface AlertChannelConfig {
    // Webhook配置
    url?: string
    method?: string
    headers?: Record<string, string>

    // Slack配置
    token?: string
    channel?: string

    // Discord配置
    webhookUrl?: string
    username?: string
    avatarUrl?: string

    // 钉钉配置
    accessToken?: string
    secret?: string
    atMobiles?: string[]
    isAtAll?: boolean

    // 企业微信配置
    webhookKey?: string
    mentionedList?: string[]
    mentionedMobileList?: string[]
}

/**
 * 告警通道
 */
export interface AlertChannel {
    id: string
    name: string
    type: AlertChannelType
    config: AlertChannelConfig
    enabled: boolean
    createdAt: string
    updatedAt: string
}

/**
 * 告警条件
 */
export interface AlertCondition {
    metric: string
    operator: ConditionOperator
    threshold: number
    duration?: number // 持续时间(秒)
}

/**
 * 告警规则
 */
export interface AlertRule {
    id: string
    name: string
    description?: string
    enabled: boolean
    severity: AlertSeverity
    conditions: AlertCondition[]
    channelIds: string[]
    cooldown: number // 冷却时间(秒)
    template?: string
    createdAt: string
    updatedAt: string
}

/**
 * 告警历史状态
 */
export enum AlertHistoryStatus {
    Pending = "pending",
    Sent = "sent",
    Failed = "failed",
    Acknowledged = "acknowledged"
}

/**
 * 告警历史
 */
export interface AlertHistory {
    id: string
    ruleId: string
    ruleName: string
    severity: AlertSeverity
    message: string
    channelIds: string[]
    status: AlertHistoryStatus
    errorMessage?: string
    acknowledgedBy?: string
    acknowledgedAt?: string
    triggeredAt: string
    createdAt: string
}

/**
 * 告警历史响应(带分页)
 */
export interface AlertHistoryResponse {
    items: AlertHistory[]
    total: number
    page: number
    size: number
}

/**
 * 创建告警通道请求
 */
export interface CreateAlertChannelRequest {
    name: string
    type: AlertChannelType
    config: AlertChannelConfig
    enabled: boolean
}

/**
 * 更新告警通道请求
 */
export interface UpdateAlertChannelRequest {
    name?: string
    config?: AlertChannelConfig
    enabled?: boolean
}

/**
 * 创建告警规则请求
 */
export interface CreateAlertRuleRequest {
    name: string
    description?: string
    enabled: boolean
    severity: AlertSeverity
    conditions: AlertCondition[]
    channelIds: string[]
    cooldown: number
    template?: string
}

/**
 * 更新告警规则请求
 */
export interface UpdateAlertRuleRequest {
    name?: string
    description?: string
    enabled?: boolean
    severity?: AlertSeverity
    conditions?: AlertCondition[]
    channelIds?: string[]
    cooldown?: number
    template?: string
}

/**
 * 告警通道列表响应
 */
export interface AlertChannelListResponse {
    items: AlertChannel[]
    total: number
}

/**
 * 告警规则列表响应
 */
export interface AlertRuleListResponse {
    items: AlertRule[]
    total: number
}

/**
 * 测试告警通道请求
 */
export interface TestAlertChannelRequest {
    message: string
}

/**
 * 确认告警请求
 */
export interface AcknowledgeAlertRequest {
    note?: string
}

/**
 * 告警统计
 */
export interface AlertStats {
    total: number
    pending: number
    sent: number
    failed: number
    acknowledged: number
    bySeverity: {
        low: number
        medium: number
        high: number
        critical: number
    }
}
