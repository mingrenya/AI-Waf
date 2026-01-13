import { get, post, put, del } from './index'
import {
    Site,
    SiteListResponse,
    CreateSiteRequest,
    UpdateSiteRequest
} from '@/types/site'

// 站点API接口基础路径
const BASE_URL = '/site'

/**
 * 站点相关API服务
 */
export const siteApi = {
    /**
     * 获取站点列表
     * @param page 页码
     * @param size 每页数量
     * @returns 站点列表响应数据
     */
    getSites: (page: number = 1, size: number = 10): Promise<SiteListResponse> => {
        return get<SiteListResponse>(BASE_URL, {
            params: { page, size }
        })
    },

    /**
     * 创建新站点
     * @param site 站点创建请求数据
     * @returns 创建后的站点详情
     */
    createSite: (site: CreateSiteRequest): Promise<Site> => {
        return post<Site>(BASE_URL, site)
    },

    /**
     * 获取单个站点详情
     * @param id 站点ID
     * @returns 站点详情
     */
    getSite: (id: string): Promise<Site> => {
        return get<Site>(`${BASE_URL}/${id}`)
    },

    /**
     * 更新站点
     * @param id 站点ID
     * @param site 站点更新请求数据
     * @returns 更新后的站点详情
     */
    updateSite: (id: string, site: UpdateSiteRequest): Promise<Site> => {
        return put<Site>(`${BASE_URL}/${id}`, site)
    },

    /**
     * 删除站点
     * @param id 站点ID
     * @returns void
     */
    deleteSite: (id: string): Promise<void> => {
        return del<void>(`${BASE_URL}/${id}`)
    }
}