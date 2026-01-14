import { get } from './index'
import { SecurityMetricsResponse, SecurityMetricsRequest } from '@/types/security-metrics'

/**
 * 综合安全指标 API
 */
export const securityMetricsApi = {
    /**
     * 获取综合安全指标
     */
    getSecurityMetrics: (params: SecurityMetricsRequest): Promise<SecurityMetricsResponse> => {
        return get<SecurityMetricsResponse>('/stats/security-metrics', {
            params
        })
    },
}
