// src/api/runner.ts
import { get, post } from './index'
import { RunnerStatusResponse, RunnerControlRequest, RunnerControlResponse } from '@/types/runner'

const BASE_URL = '/runner'
/**
 * 运行器相关API服务
 */
export const runnerApi = {
    /**
     * 获取运行器状态
     * @returns 运行器状态信息
     */
    getStatus: (): Promise<RunnerStatusResponse> => {
        return get<RunnerStatusResponse>(`${BASE_URL}/status`)
    },

    /**
     * 控制运行器（启动、停止、重启等）
     * @param request 控制请求
     * @returns 控制操作结果
     */
    control: (request: RunnerControlRequest): Promise<RunnerControlResponse> => {
        return post<RunnerControlResponse>(`${BASE_URL}/control`, request)
    }
}