/**
 * 综合安全指标类型定义
 */

// 规则触发统计
export interface RuleTriggerStats {
    ruleId: number
    ruleName: string
    count: number
    percentage: number
}

// 严重等级统计
export interface SeverityStats {
    level: number
    levelName: string
    count: number
    percentage: number
}

// 攻击类型统计
export interface AttackTypeStats {
    category: string
    count: number
    percentage: number
}

// 地理位置统计
export interface GeoLocationStats {
    country: string
    countryCode: string
    city: string
    count: number
    percentage: number
}

// 规则引擎统计
export interface RuleEngineStats {
    totalRules: number
    enabledRules: number
    disabledRules: number
    whitelistRules: number
    blacklistRules: number
    avgMatchTime: number
    ruleEfficiency: number
}

// 封禁IP统计
export interface BlockedIPStats {
    totalBlocked: number
    activeBlocked: number
    expiredBlocked: number
    highFrequencyVisit: number
    highFrequencyAttack: number
    highFrequencyError: number
}

// 威胁等级分布
export interface ThreatLevelDistribution {
    critical: number
    high: number
    medium: number
    low: number
}

// 响应时间统计
export interface ResponseTimeStats {
    avgResponseTime: number
    maxResponseTime: number
    minResponseTime: number
    p50ResponseTime: number
    p95ResponseTime: number
    p99ResponseTime: number
}

// 时间序列数据点
export interface TimeSeriesDataPoint {
    timestamp: string
    value: number
}

// 流量数据点
export interface TrafficDataPoint {
    timestamp: string
    inboundTraffic: number
    outboundTraffic: number
}

// 时间序列响应
export interface TimeSeriesResponse {
    metric: string
    timeRange: string
    data: TimeSeriesDataPoint[]
}

// 流量时间序列响应
export interface TrafficTimeSeriesResponse {
    timeRange: string
    data: TrafficDataPoint[]
}

// 概览统计（重用现有的）
export interface OverviewStats {
    timeRange: string
    totalRequests: number
    inboundTraffic: number
    outboundTraffic: number
    maxQPS: number
    error4xx: number
    error4xxRate: number
    error5xx: number
    error5xxRate: number
    blockCount: number
    attackIPCount: number
}

// 综合安全指标响应
export interface SecurityMetricsResponse {
    timeRange: string
    overview: OverviewStats
    ruleEngine: RuleEngineStats
    topTriggeredRules: RuleTriggerStats[]
    severityDistribution: SeverityStats[]
    attackTypeDistribution: AttackTypeStats[]
    topAttackSources: GeoLocationStats[]
    blockedIPMetrics: BlockedIPStats
    threatLevel: ThreatLevelDistribution
    responseTime: ResponseTimeStats
    requestTrend: TimeSeriesResponse
    blockTrend: TimeSeriesResponse
    trafficTrend: TrafficTimeSeriesResponse
}

// 安全指标请求
export interface SecurityMetricsRequest {
    timeRange: '24h' | '7d' | '30d'
}
