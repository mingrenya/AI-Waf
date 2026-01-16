// 自适应限流配置类型定义

export interface LearningModeConfig {
    enabled: boolean
    learningDuration: number // 秒
    sampleInterval: number // 秒
    minSamples: number
}

export interface BaselineConfig {
    calculationMethod: 'mean' | 'median' | 'percentile'
    percentile: number // 0-100
    updateInterval: number // 秒
    historyWindow: number // 秒
}

export interface AutoAdjustmentConfig {
    enabled: boolean
    anomalyThreshold: number // 倍数
    minThreshold: number
    maxThreshold: number
    adjustmentFactor: number
    cooldownPeriod: number // 秒
    gradualAdjustment: boolean
    adjustmentStepRatio: number
}

export interface ApplyToConfig {
    visitLimit: boolean
    attackLimit: boolean
    errorLimit: boolean
}

export interface AdaptiveThrottlingConfig {
    id?: string
    enabled: boolean
    createdAt?: string
    updatedAt?: string
    learningMode: LearningModeConfig
    baseline: BaselineConfig
    autoAdjustment: AutoAdjustmentConfig
    applyTo: ApplyToConfig
}

// 流量指标
export interface TrafficMetrics {
    requestRate: number // 每秒请求数
    uniqueIPs: number
    blockedCount: number
    passedCount: number
}

// 统计数据
export interface TrafficStatistics {
    mean: number
    median: number
    stdDev: number
    p95: number
    p99: number
    min: number
    max: number
}

// 流量模式记录
export interface TrafficPattern {
    id: string
    timestamp: string
    type: 'visit' | 'attack' | 'error'
    metrics: TrafficMetrics
    statistics: TrafficStatistics
}

// 基线值
export interface BaselineValue {
    id: string
    type: 'visit' | 'attack' | 'error'
    value: number
    calculatedAt: string
    sampleSize: number
    confidenceLevel: number
    updatedAt: string
}

// 限流调整日志
export interface ThrottleAdjustmentLog {
    id: string
    timestamp: string
    type: 'visit' | 'attack' | 'error'
    oldThreshold: number
    oldBaseline: number
    newThreshold: number
    newBaseline: number
    reason: string
    currentTraffic: number
    anomalyScore: number
    triggeredBy: 'auto' | 'manual'
    adjustmentRatio: number
}

// API 请求/响应类型
export interface AdaptiveThrottlingConfigResponse {
    id: string
    enabled: boolean
    createdAt: string
    updatedAt: string
    learningMode: LearningModeConfig
    baseline: BaselineConfig
    autoAdjustment: AutoAdjustmentConfig
    applyTo: ApplyToConfig
}

export interface CreateAdaptiveThrottlingConfigRequest {
    enabled: boolean
    learningMode: LearningModeConfig
    baseline: BaselineConfig
    autoAdjustment: AutoAdjustmentConfig
    applyTo: ApplyToConfig
}

export interface UpdateAdaptiveThrottlingConfigRequest {
    enabled?: boolean
    learningMode?: Partial<LearningModeConfig>
    baseline?: Partial<BaselineConfig>
    autoAdjustment?: Partial<AutoAdjustmentConfig>
    applyTo?: Partial<ApplyToConfig>
}

// 流量模式查询参数
export interface TrafficPatternQuery {
    type?: 'visit' | 'attack' | 'error'
    startTime?: string
    endTime?: string
    page?: number
    pageSize?: number
}

// 流量模式列表响应
export interface TrafficPatternListResponse {
    results: TrafficPattern[]
    totalCount: number
    pageSize: number
    currentPage: number
    totalPages: number
}

// 基线值查询参数
export interface BaselineValueQuery {
    type?: 'visit' | 'attack' | 'error'
}

// 基线值列表响应
export interface BaselineValueListResponse {
    results: BaselineValue[]
}

// 调整日志查询参数
export interface ThrottleAdjustmentLogQuery {
    type?: 'visit' | 'attack' | 'error'
    triggeredBy?: 'auto' | 'manual'
    startTime?: string
    endTime?: string
    page?: number
    pageSize?: number
}

// 调整日志列表响应
export interface ThrottleAdjustmentLogListResponse {
    results: ThrottleAdjustmentLog[]
    totalCount: number
    pageSize: number
    currentPage: number
    totalPages: number
}

// 实时统计数据
export interface AdaptiveThrottlingStats {
    currentBaseline: {
        visit: number
        attack: number
        error: number
    }
    currentThreshold: {
        visit: number
        attack: number
        error: number
    }
    recentAdjustments: number
    learningProgress: number // 0-100
    anomalyDetected: boolean
    lastUpdateTime: string
}
