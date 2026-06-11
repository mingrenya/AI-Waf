import { get, post } from './index'
import {
  AttackEventResponse,
  AttackLogResponse,
  AttackEventQuery,
  AttackLogQuery
} from '@/types/log'

// WAF基础路径
const BASE_URL = '/log'

/**
 * WAF相关API服务
 */
export const logApi = {
  /**
   * 获取攻击事件（按客户端IP和域名聚合）
   * @param query 查询参数
   * @returns 攻击事件响应数据
   */
  getAttackEvents: (query: AttackEventQuery): Promise<AttackEventResponse> => {
    return get<AttackEventResponse>(`${BASE_URL}/event`, { params: query })
  },

  /**
   * 获取攻击日志详情
   * @param query 查询参数
   * @returns 攻击日志响应数据
   */
  getAttackLogs: (query: AttackLogQuery): Promise<AttackLogResponse> => {
    return get<AttackLogResponse>(`${BASE_URL}`, { params: query })
  },

  /**
   * Loki 引擎日志即时查询
   * @param query LogQL 查询语句
   * @param limit 返回条数上限
   * @param start 起始时间 (默认 1h)
   * @param end 结束时间 (默认 now)
   */
  lokiQuery: (query: string, limit?: number, start?: string, end?: string): Promise<LokiLogResponse> => {
    return post<LokiLogResponse>(`${BASE_URL}/loki-query`, {
      query,
      limit: limit || 100,
      start: start || '1h',
      end: end || 'now'
    })
  },

  /**
   * Loki 引擎日志范围查询
   */
  lokiRange: (query: string, start: string, end: string, limit?: number): Promise<LokiLogResponse> => {
    return post<LokiLogResponse>(`${BASE_URL}/loki-range`, {
      query,
      start,
      end,
      limit: limit || 100
    })
  }
}

/** Loki 日志条目 */
export interface LokiLogEntry {
  timestamp: string
  labels: Record<string, string>
  message: string
  level?: string
  component?: string
}

/** Loki 查询响应 */
export interface LokiLogResponse {
  results: LokiLogEntry[]
  totalHits: number
  query: string
  resultType: string
}
