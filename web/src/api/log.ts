import { get } from './index'
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
  }
}
