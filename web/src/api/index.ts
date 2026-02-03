import axios, { AxiosError, AxiosRequestConfig, AxiosResponse } from 'axios'
import { useAuthStore } from '@/store/auth'
import { ENV } from '@/utils/env'

// ======================
// API 响应类型定义
// ======================

/**
 * 通用API响应接口
 */
export interface APIResponse<T = unknown> {
    code: number
    message: string
    success: boolean
    requestId: string
    timestamp: string
    data?: T
    error?: string
}

/**
 * 自定义API错误类
 */
export class ApiError extends Error {
    code: number
    requestId?: string
    errorDetail?: string

    constructor(message: string, code: number, requestId?: string, errorDetail?: string) {
        super(message)
        this.name = 'ApiError'
        this.code = code
        this.requestId = requestId
        this.errorDetail = errorDetail
    }
}

// 定义基础请求和响应接口
export type ApiRequestData = Record<string, unknown>
export type ApiResponseData = Record<string, unknown>

// ======================
// API 客户端配置
// ======================

const API_BASE_URL = ENV.API_BASE_URL || '/api/v1'
const API_TIMEOUT = Number(ENV.API_TIMEOUT || 10000)

// 创建axios实例
const apiClient = axios.create({
    baseURL: API_BASE_URL,
    timeout: API_TIMEOUT,
    headers: {
        'Content-Type': 'application/json',
    },
})

// 请求拦截器
apiClient.interceptors.request.use(
    (config) => {
        const token = useAuthStore.getState().token

        // 公开API列表（不需要token）
        const publicApis = ['/auth/login', '/auth/register', '/auth/forgot-password', '/health']
        const isPublicApi = publicApis.some(api => config.url?.includes(api))

        if (token && config.headers) {
            config.headers.Authorization = `Bearer ${token}`
        } else if (!token && !isPublicApi) {
            // 调试：只对需要认证的API记录警告
            if (import.meta.env.DEV) {
                console.warn('⚠️ API请求未携带token:', config.url)
            }
        }

        return config
    },
    (error) => {
        return Promise.reject(error)
    }
)

// 响应拦截器 - 统一处理错误和响应格式
apiClient.interceptors.response.use(
    (response) => {
        // 检查响应格式是否符合API标准格式
        const data = response.data as APIResponse

        // 特殊处理401：即使HTTP状态码是200，如果data.code是401也要跳转登录
        if (data && data.code === 401) {
            if (import.meta.env.DEV) {
                console.error('❌ 401未授权（业务状态码）:', {
                    url: response.config.url,
                    message: data.message || '请先登录',
                    httpStatus: response.status
                })
            }
            useAuthStore.getState().logout()
            window.location.href = '/login'
            throw new ApiError(
                data.message || '未授权访问',
                401,
                data.requestId,
                data.error
            )
        }

        // 如果响应中有success字段且为false，视为业务逻辑错误
        if (data && data.success === false) {
            throw new ApiError(
                data.message || '请求失败',
                data.code || response.status,
                data.requestId,
                data.error
            )
        }

        return response
    },
    (error: AxiosError<APIResponse>) => {
        const status = error.response?.status
        const errorData = error.response?.data

        // 处理401未授权错误（HTTP状态码级别）
        if (status === 401) {
            console.error('❌ 401未授权（HTTP状态码）:', {
                url: error.config?.url,
                message: errorData?.message || '请先登录',
                hasToken: !!error.config?.headers?.Authorization
            })
            useAuthStore.getState().logout()
            window.location.href = '/login'
        }

        // 转换为自定义API错误对象
        if (errorData) {
            throw new ApiError(
                errorData.message || '请求失败',
                errorData.code || status || 500,
                errorData.requestId,
                errorData.error
            )
        } else {
            // 网络错误或其他非标准错误
            throw new ApiError(
                error.message || '网络请求失败',
                status || 500
            )
        }
    }
)

// 增加请求重试功能的封装方法
const withRetry = async <T>(
    requestFn: () => Promise<T>,
    maxRetries = 3,
    delay = 1000
): Promise<T> => {
    let retries = 0

    while (retries < maxRetries) {
        try {
            return await requestFn()
        } catch (error) {
            // 使用类型断言而不是 any
            const err = error as Error
            if (error instanceof ApiError && error.code >= 500 && retries < maxRetries - 1) {
                // 只有服务器错误才重试，并且不是最后一次尝试
                retries++
                await new Promise(resolve => setTimeout(resolve, delay * retries))
                continue
            }
            throw err
        }
    }

    throw new Error('Max retries reached')
}

// ======================
// API 请求方法
// ======================

/**
 * GET请求
 * @param url 请求URL
 * @param config 请求配置
 * @returns 响应数据
 */
export const get = <T = ApiResponseData>(url: string, config?: AxiosRequestConfig): Promise<T> => {
    return apiClient.get<APIResponse<T>>(url, config)
        .then((response: AxiosResponse<APIResponse<T>>): T => {
            // 检查响应是否存在
            if (!response) {
                if (import.meta.env.DEV) {
                    console.error('GET响应对象为null:', url)
                }
                throw new Error('服务器无响应')
            }
            
            const data = response.data
            
            // 检查响应数据是否存在
            if (!data) {
                if (import.meta.env.DEV) {
                    console.error('GET响应数据为空:', url)
                }
                throw new Error('服务器无响应')
            }
            
            // 检查是否为空字符串（后端可能返回空字符串）
            if (typeof data === 'string' && (data as string).trim() === '') {
                if (import.meta.env.DEV) {
                    console.error('GET响应体为空字符串:', url, {
                    status: response.status,
                    dataType: typeof data
                })
                }
                throw new Error('后端返回空数据，请检查数据是否已初始化')
            }
            
            // 如果后端返回标准格式 {success, data, message}，提取data字段
            // 如果后端直接返回数据(如{total, items})，直接使用
            const responseData = data as { data?: T }
            if (responseData.data !== undefined) {
                return responseData.data as T
            }
            
            // 兼容直接返回数据的情况
            return data as T
        })
        .catch((error) => {
            console.error('GET请求失败:', url, error)
            throw error
        })
}

/**
 * POST请求
 * @param url 请求URL
 * @param data 请求数据
 * @param config 请求配置
 * @returns 响应数据
 */
export const post = <T = ApiResponseData>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> => {
    return apiClient.post<APIResponse<T>>(url, data, config)
        .then((response: AxiosResponse<APIResponse<T>>) => {
            // 安全处理：检查response.data是否存在
            if (!response.data) {
                if (import.meta.env.DEV) {
                    console.error('POST响应数据为空:', url)
                }
                throw new Error('服务器返回空响应')
            }
            
            // 如果后端返回标准格式，提取data字段，否则直接使用
            const responseData = response.data as { data?: T }
            if (responseData.data !== undefined) {
                return responseData.data as T
            }
            
            return response.data as T
        })
}

/**
 * PUT请求
 * @param url 请求URL
 * @param data 请求数据
 * @param config 请求配置
 * @returns 响应数据
 */
export const put = <T = ApiResponseData>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> => {
    return apiClient.put<APIResponse<T>>(url, data, config)
        .then((response: AxiosResponse<APIResponse<T>>) => {
            if (!response.data) {
                if (import.meta.env.DEV) {
                    console.error('PUT响应数据为空:', url)
                }
                throw new Error('服务器返回空响应')
            }
            const responseData = response.data as { data?: T }
            return responseData.data !== undefined ? responseData.data as T : response.data as T
        })
}

/**
 * PATCH请求
 * @param url 请求URL
 * @param data 请求数据
 * @param config 请求配置
 * @returns 响应数据
 */
export const patch = <T = ApiResponseData>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> => {
    return apiClient.patch<APIResponse<T>>(url, data, config)
        .then((response: AxiosResponse<APIResponse<T>>) => {
            return response.data.data as T
        })
}

/**
 * DELETE请求
 * @param url 请求URL
 * @param config 请求配置
 * @returns 响应数据
 */
export const del = <T = ApiResponseData>(url: string, config?: AxiosRequestConfig): Promise<T> => {
    return apiClient.delete<APIResponse<T>>(url, config)
        .then((response: AxiosResponse<APIResponse<T>>) => {
            if (!response.data) {
                if (import.meta.env.DEV) {
                    console.error('DELETE响应数据为空:', url)
                }
                throw new Error('服务器返回空响应')
            }
            const responseData = response.data as { data?: T }
            return responseData.data !== undefined ? responseData.data as T : response.data as T
        })
}

/**
 * 带重试功能的GET请求
 * @param url 请求URL
 * @param config 请求配置
 * @param retries 重试次数
 * @returns 响应数据
 */
export const getWithRetry = <T = ApiResponseData>(url: string, config?: AxiosRequestConfig, retries = 3): Promise<T> => {
    return withRetry(() => get<T>(url, config), retries)
}

/**
 * 带重试功能的POST请求
 * @param url 请求URL
 * @param data 请求数据
 * @param config 请求配置
 * @param retries 重试次数
 * @returns 响应数据
 */
export const postWithRetry = <T = ApiResponseData>(url: string, data?: unknown, config?: AxiosRequestConfig, retries = 3): Promise<T> => {
    return withRetry(() => post<T>(url, data, config), retries)
}

/**
 * 带重试功能的PUT请求
 * @param url 请求URL
 * @param data 请求数据
 * @param config 请求配置
 * @param retries 重试次数
 * @returns 响应数据
 */
export const putWithRetry = <T = ApiResponseData>(url: string, data?: unknown, config?: AxiosRequestConfig, retries = 3): Promise<T> => {
    return withRetry(() => put<T>(url, data, config), retries)
}

/**
 * 带重试功能的DELETE请求
 * @param url 请求URL
 * @param config 请求配置
 * @param retries 重试次数
 * @returns 响应数据
 */
export const delWithRetry = <T = ApiResponseData>(url: string, config?: AxiosRequestConfig, retries = 3): Promise<T> => {
    return withRetry(() => del<T>(url, config), retries)
}

/**
 * 带重试功能的PATCH请求
 * @param url 请求URL
 * @param data 请求数据
 * @param config 请求配置
 * @param retries 重试次数
 * @returns 响应数据
 */
export const patchWithRetry = <T = ApiResponseData>(url: string, data?: unknown, config?: AxiosRequestConfig, retries = 3): Promise<T> => {
    return withRetry(() => patch<T>(url, data, config), retries)
}


export default apiClient