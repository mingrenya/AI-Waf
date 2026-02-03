// src/feature/global-setting/hooks/useConfig.ts
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { configApi } from '@/api/config'
import { ConfigPatchRequest } from '@/types/config'
import { useToast } from '@/hooks/use-toast'
import { ApiError } from '@/api/index'
import { ConstantCategory, getConstant } from '@/constant'
import { useTranslation } from 'react-i18next'

// 获取配置查询hook
export const useConfigQuery = () => {
    const query = useQuery({
        queryKey: ['config'],
        queryFn: configApi.getConfig,
        // 添加重试策略，提高数据加载可靠性
        retry: 2,
        // 减少staleTime，确保页面切换时能获取最新数据
        staleTime: 1000 * 60, // 1分钟
        // 窗口重新聚焦时重新获取数据
        refetchOnWindowFocus: true,
        // 确保在挂载时总是获取数据
        refetchOnMount: 'always'
    })

    return {
        config: query.data,
        isLoading: query.isPending,
        error: query.error,
        refetch: query.refetch
    }
}

// 更新配置mutation hook
export const useUpdateConfig = () => {
    const queryClient = useQueryClient()
    const { toast } = useToast()
    const [error, setError] = useState<ApiError | null>(null)
    const { t } = useTranslation()

    const mutation = useMutation({
        mutationFn: (data: ConfigPatchRequest) => configApi.updateConfig(data),
        onSuccess: () => {
            toast({
                title: t("globalSetting.config.updateSuccess", "更新成功"),
                description: t("globalSetting.config.systemConfigUpdated", "系统配置已成功更新"),
                duration: getConstant(ConstantCategory.FEATURE, 'TOAST_DURATION', 3000),
            })
            queryClient.invalidateQueries({ queryKey: ['config'] })
        },
        onError: (error: ApiError) => {
            if (import.meta.env.DEV) {
                console.error('更新配置失败:', error)
            }
            setError(error)
            toast({
                title: t("globalSetting.config.updateFailed", "更新失败"),
                description: error.message || t("globalSetting.config.updateError", "更新配置时出现错误"),
                variant: "destructive",
                duration: getConstant(ConstantCategory.FEATURE, 'TOAST_DURATION', 3000),
            })
        }
    })

    return {
        updateConfig: mutation.mutate,
        isLoading: mutation.isPending,
        error,
        clearError: () => setError(null),
    }
}

// 获取引擎名称常量
export const getEngineNameConstant = (): string => {
    return getConstant(ConstantCategory.CONFIG, 'ENGINE_NAME', 'coraza')
}

// 更新配置中指定引擎的指令
export const updateEngineDirectives = (config: ConfigPatchRequest, directives: string): ConfigPatchRequest => {
    const engineName = getEngineNameConstant()

    // 如果没有引擎配置或appConfig，直接返回原始配置
    if (!config.engine || !config.engine.appConfig) {
        return config
    }

    // 查找特定引擎的配置项
    const engineAppConfig = config.engine.appConfig.find(app => app.name === engineName)

    // 只有找到才更新，否则不做任何处理
    if (engineAppConfig) {
        engineAppConfig.directives = directives
    }

    return config
}