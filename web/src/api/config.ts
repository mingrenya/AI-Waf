import { get, patch } from './index'
import { ConfigResponse, ConfigPatchRequest } from '@/types/config'

const BASE_URL = '/config'
/**
 * 配置相关API服务
 */
export const configApi = {
    /**
     * 获取系统配置
     * @returns 系统配置信息
     */
    getConfig: (): Promise<ConfigResponse> => {
        return get<ConfigResponse>(BASE_URL)
    },

    /**
     * 更新系统配置
     * @param config 要更新的配置信息
     * @returns 更新后的系统配置
     */
    updateConfig: (config: ConfigPatchRequest): Promise<ConfigResponse> => {
        return patch<ConfigResponse>(BASE_URL, config)
    }
}

