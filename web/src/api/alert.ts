import { get, post, put, del } from './index'
import {
    AlertChannel,
    AlertChannelListResponse,
    CreateAlertChannelRequest,
    UpdateAlertChannelRequest,
    TestAlertChannelRequest,
    AlertRule,
    AlertRuleListResponse,
    CreateAlertRuleRequest,
    UpdateAlertRuleRequest,
    AlertHistory,
    AlertHistoryResponse,
    AcknowledgeAlertRequest,
    AlertStats
} from '@/types/alert'

// 告警API接口基础路径
const CHANNEL_URL = '/alert/channel'
const RULE_URL = '/alert/rule'
const HISTORY_URL = '/alert/history'

/**
 * 告警通道相关API服务
 */
export const alertChannelApi = {
    /**
     * 获取告警通道列表
     * @param page 页码
     * @param size 每页数量
     * @returns 告警通道列表响应数据
     */
    getChannels: (page: number = 1, size: number = 10): Promise<AlertChannelListResponse> => {
        return get<AlertChannelListResponse>(CHANNEL_URL, {
            params: { page, size }
        })
    },

    /**
     * 创建新告警通道
     * @param channel 告警通道创建请求数据
     * @returns 创建后的告警通道详情
     */
    createChannel: (channel: CreateAlertChannelRequest): Promise<AlertChannel> => {
        return post<AlertChannel>(CHANNEL_URL, channel)
    },

    /**
     * 获取单个告警通道详情
     * @param id 告警通道ID
     * @returns 告警通道详情
     */
    getChannel: (id: string): Promise<AlertChannel> => {
        return get<AlertChannel>(`${CHANNEL_URL}/${id}`)
    },

    /**
     * 更新告警通道
     * @param id 告警通道ID
     * @param channel 告警通道更新请求数据
     * @returns 更新后的告警通道详情
     */
    updateChannel: (id: string, channel: UpdateAlertChannelRequest): Promise<AlertChannel> => {
        return put<AlertChannel>(`${CHANNEL_URL}/${id}`, channel)
    },

    /**
     * 删除告警通道
     * @param id 告警通道ID
     * @returns void
     */
    deleteChannel: (id: string): Promise<void> => {
        return del<void>(`${CHANNEL_URL}/${id}`)
    },

    /**
     * 测试告警通道
     * @param id 告警通道ID
     * @param request 测试请求数据
     * @returns void
     */
    testChannel: (id: string, request: TestAlertChannelRequest): Promise<void> => {
        return post<void>(`${CHANNEL_URL}/${id}/test`, request)
    }
}

/**
 * 告警规则相关API服务
 */
export const alertRuleApi = {
    /**
     * 获取告警规则列表
     * @param page 页码
     * @param size 每页数量
     * @returns 告警规则列表响应数据
     */
    getRules: (page: number = 1, size: number = 10): Promise<AlertRuleListResponse> => {
        return get<AlertRuleListResponse>(RULE_URL, {
            params: { page, size }
        })
    },

    /**
     * 创建新告警规则
     * @param rule 告警规则创建请求数据
     * @returns 创建后的告警规则详情
     */
    createRule: (rule: CreateAlertRuleRequest): Promise<AlertRule> => {
        return post<AlertRule>(RULE_URL, rule)
    },

    /**
     * 获取单个告警规则详情
     * @param id 告警规则ID
     * @returns 告警规则详情
     */
    getRule: (id: string): Promise<AlertRule> => {
        return get<AlertRule>(`${RULE_URL}/${id}`)
    },

    /**
     * 更新告警规则
     * @param id 告警规则ID
     * @param rule 告警规则更新请求数据
     * @returns 更新后的告警规则详情
     */
    updateRule: (id: string, rule: UpdateAlertRuleRequest): Promise<AlertRule> => {
        return put<AlertRule>(`${RULE_URL}/${id}`, rule)
    },

    /**
     * 删除告警规则
     * @param id 告警规则ID
     * @returns void
     */
    deleteRule: (id: string): Promise<void> => {
        return del<void>(`${RULE_URL}/${id}`)
    }
}

/**
 * 告警历史相关API服务
 */
export const alertHistoryApi = {
    /**
     * 获取告警历史列表
     * @param page 页码
     * @param size 每页数量
     * @param ruleId 可选的规则ID过滤
     * @param severity 可选的严重等级过滤
     * @param status 可选的状态过滤
     * @returns 告警历史列表响应数据
     */
    getHistory: (
        page: number = 1,
        size: number = 10,
        ruleId?: string,
        severity?: string,
        status?: string
    ): Promise<AlertHistoryResponse> => {
        const params: Record<string, any> = { page, size }
        if (ruleId) params.ruleId = ruleId
        if (severity) params.severity = severity
        if (status) params.status = status

        return get<AlertHistoryResponse>(HISTORY_URL, { params })
    },

    /**
     * 获取单个告警历史详情
     * @param id 告警历史ID
     * @returns 告警历史详情
     */
    getHistoryDetail: (id: string): Promise<AlertHistory> => {
        return get<AlertHistory>(`${HISTORY_URL}/${id}`)
    },

    /**
     * 确认告警
     * @param id 告警历史ID
     * @param request 确认请求数据
     * @returns void
     */
    acknowledgeAlert: (id: string, request: AcknowledgeAlertRequest): Promise<void> => {
        return post<void>(`${HISTORY_URL}/${id}/acknowledge`, request)
    },

    /**
     * 获取告警统计信息
     * @returns 告警统计数据
     */
    getStats: (): Promise<AlertStats> => {
        return get<AlertStats>(`${HISTORY_URL}/stats`)
    }
}
