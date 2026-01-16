import { get, post, put, del } from './index'
import type {
    AdaptiveThrottlingConfigResponse,
    CreateAdaptiveThrottlingConfigRequest,
    UpdateAdaptiveThrottlingConfigRequest,
    TrafficPatternQuery,
    TrafficPatternListResponse,
    BaselineValueQuery,
    BaselineValueListResponse,
    ThrottleAdjustmentLogQuery,
    ThrottleAdjustmentLogListResponse,
    AdaptiveThrottlingStats
} from '@/types/adaptive-throttling'

// 自适应限流API基础路径
const BASE_URL = '/adaptive-throttling'

/**
 * 自适应限流相关API服务
 */
export const adaptiveThrottlingApi = {
    /**
     * 获取自适应限流配置
     * @returns 自适应限流配置
     */
    getConfig: (): Promise<AdaptiveThrottlingConfigResponse> => {
        return get<AdaptiveThrottlingConfigResponse>(BASE_URL)
    },

    /**
     * 创建自适应限流配置
     * @param config 配置数据
     * @returns 创建的配置
     */
    createConfig: (config: CreateAdaptiveThrottlingConfigRequest): Promise<AdaptiveThrottlingConfigResponse> => {
        return post<AdaptiveThrottlingConfigResponse>(BASE_URL, config)
    },

    /**
     * 更新自适应限流配置
     * @param config 配置数据
     * @returns 更新后的配置
     */
    updateConfig: (config: UpdateAdaptiveThrottlingConfigRequest): Promise<AdaptiveThrottlingConfigResponse> => {
        return put<AdaptiveThrottlingConfigResponse>(BASE_URL, config)
    },

    /**
     * 删除自适应限流配置
     * @returns void
     */
    deleteConfig: (): Promise<void> => {
        return del<void>(BASE_URL)
    },

    /**
     * 获取流量模式历史数据
     * @param query 查询参数
     * @returns 流量模式列表
     */
    getTrafficPatterns: (query: TrafficPatternQuery): Promise<TrafficPatternListResponse> => {
        return get<TrafficPatternListResponse>(`${BASE_URL}/patterns`, { params: query })
    },

    /**
     * 获取当前基线值
     * @param query 查询参数
     * @returns 基线值列表
     */
    getBaselines: (query: BaselineValueQuery): Promise<BaselineValueListResponse> => {
        return get<BaselineValueListResponse>(`${BASE_URL}/baselines`, { params: query })
    },

    /**
     * 获取调整日志
     * @param query 查询参数
     * @returns 调整日志列表
     */
    getAdjustmentLogs: (query: ThrottleAdjustmentLogQuery): Promise<ThrottleAdjustmentLogListResponse> => {
        return get<ThrottleAdjustmentLogListResponse>(`${BASE_URL}/logs`, { params: query })
    },

    /**
     * 获取实时统计数据
     * @returns 统计数据
     */
    getStats: (): Promise<AdaptiveThrottlingStats> => {
        return get<AdaptiveThrottlingStats>(`${BASE_URL}/stats`)
    },

    /**
     * 手动触发基线重新计算
     * @returns void
     */
    recalculateBaseline: (): Promise<void> => {
        return post<void>(`${BASE_URL}/recalculate`)
    },

    /**
     * 重置学习数据
     * @returns void
     */
    resetLearning: (): Promise<void> => {
        return post<void>(`${BASE_URL}/reset`)
    }
}
