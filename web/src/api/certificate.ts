import { get, post, put, del } from './index'
import {
    Certificate,
    CertificateListResponse,
    CertificateCreateRequest,
    CertificateUpdateRequest,
} from '@/types/certificate'

// 证书API接口基础路径
const BASE_URL = '/certificate'

/**
 * 证书相关API服务
 */
export const certificatesApi = {
    /**
     * 获取证书列表
     * @param page 页码
     * @param size 每页数量
     * @returns 证书列表响应数据
     */
    getCertificates: (page: number = 1, size: number = 10): Promise<CertificateListResponse> => {
        return get<CertificateListResponse>(BASE_URL, {
            params: { page, size }
        })
    },

    /**
     * 创建新证书
     * @param certificate 证书创建请求数据
     * @returns 创建后的证书详情
     */
    createCertificate: (certificate: CertificateCreateRequest): Promise<Certificate> => {
        return post<Certificate>(BASE_URL, certificate)
    },

    /**
     * 获取单个证书详情
     * @param id 证书ID
     * @returns 证书详情
     */
    getCertificate: (id: string): Promise<Certificate> => {
        return get<Certificate>(`${BASE_URL}/${id}`)
    },

    /**
     * 更新证书
     * @param id 证书ID
     * @param certificate 证书更新请求数据
     * @returns 更新后的证书详情
     */
    updateCertificate: (id: string, certificate: CertificateUpdateRequest): Promise<Certificate> => {
        return put<Certificate>(`${BASE_URL}/${id}`, certificate)
    },

    /**
     * 删除证书
     * @param id 证书ID
     * @returns void
     */
    deleteCertificate: (id: string): Promise<void> => {
        return del<void>(`${BASE_URL}/${id}`)
    }
}