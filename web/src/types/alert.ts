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
    GreaterThan = ">",
    GreaterThanOrEqual = ">=",
    LessThan = "<",
    LessThanOrEqual = "<=",
    Equal = "==",
    NotEqual = "!=",
}

/**
 * 告警通道配置基础
 */
export interface BaseAlertChannelConfig {
    // 所有通道共有的字段
    [key: string]: string | string[] | boolean | Record<string, string> | undefined
}

/**
 * Webhook 通道配置
 */
export interface WebhookChannelConfig extends BaseAlertChannelConfig {
    url: string
    method?: string
    headers?: Record<string, string>
}

/**
 * Slack 通道配置
 */
export interface SlackChannelConfig extends BaseAlertChannelConfig {
    token: string
    channel: string
}

/**
 * Discord 通道配置
 */
export interface DiscordChannelConfig extends BaseAlertChannelConfig {
    webhookUrl: string
    username?: string
    avatarUrl?: string
}

/**
 * 钉钉通道配置
 */
export interface DingTalkChannelConfig extends BaseAlertChannelConfig {
    accessToken: string
    secret?: string
    atMobiles?: string[]
    isAtAll?: boolean
}

/**
 * 企业微信通道配置
 */
export interface WeComChannelConfig extends BaseAlertChannelConfig {
    webhookKey: string
    mentionedList?: string[]
    mentionedMobileList?: string[]
}

/**
 * 告警通道配置（统一类型）
 */
export type AlertChannelConfig = 
    | WebhookChannelConfig 
    | SlackChannelConfig 
    | DiscordChannelConfig 
    | DingTalkChannelConfig 
    | WeComChannelConfig
    | Record<string, string | string[] | boolean | Record<string, string> | undefined>

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
    duration: number // 持续时间(分钟)
}

/**
 * 告警规则
 */
export interface AlertRule {
    id: string
    name: string
    description?: string
    conditions: AlertCondition[]
    logic: "AND" | "OR"
    channels: string[]
    template: string
    cooldown: number // 冷却时间（分钟）
    severity: AlertSeverity
    enabled: boolean
    createdAt: string
    updatedAt: string
    createdBy?: string
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
    details: Record<string, unknown>
    channels: string[]
    status: AlertHistoryStatus
    errorMessage?: string
    acknowledgedBy?: string
    acknowledgedAt?: string
    triggeredAt: string
    sentAt?: string
}

/**
 * 告警历史分页响应
 */
export interface AlertHistoryListResponse {
    items: AlertHistory[]
    total: number
    page: number
    pageSize: number
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
    conditions: AlertCondition[]
    logic: "AND" | "OR"
    channels: string[]
    template: string
    cooldown: number
    severity: AlertSeverity
    enabled: boolean
}

/**
 * 更新告警规则请求
 */
export interface UpdateAlertRuleRequest {
    name?: string
    description?: string
    conditions?: AlertCondition[]
    logic?: "AND" | "OR"
    channels?: string[]
    template?: string
    cooldown?: number
    severity?: AlertSeverity
    enabled?: boolean
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
    comment?: string
}

/**
 * 告警统计
 */
export interface AlertStatisticsResponse {
    totalAlerts: number
    alertsBySeverity: Record<string, number>
    alertsByStatus: Record<string, number>
    // 这里只定义前端当前用到的字段；TopAlertRules 和 RecentAlerts 如需使用可再扩展
}
