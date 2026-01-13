// src/feature/global-setting/hooks/useRunner.ts
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { runnerApi } from '@/api/runner'
import { RunnerAction } from '@/types/runner'
import { useToast } from '@/hooks/use-toast'
import { ApiError } from '@/api/index'
import { getConstant } from '@/constant'
import { ConstantCategory } from '@/constant'
import { useTranslation } from 'react-i18next'

// 获取运行器状态查询hook
export const useRunnerStatusQuery = () => {
    const query = useQuery({
        queryKey: ['runner-status'],
        queryFn: runnerApi.getStatus,
        // refetchInterval: 5000 // 每5秒自动刷新一次
    })

    return {
        status: query.data,
        isLoading: query.isPending,
        error: query.error,
        refetch: query.refetch
    }
}

// 控制运行器mutation hook
export const useRunnerControl = () => {
    const queryClient = useQueryClient()
    const { toast } = useToast()
    const [error, setError] = useState<string | null>(null)
    const { t } = useTranslation()

    const mutation = useMutation({
        mutationFn: (action: RunnerAction) => runnerApi.control({ action }),
        onSuccess: (data) => {
            toast({
                title: t("globalSetting.engine.operationSuccess", "操作成功"),
                description: data.message || t("globalSetting.engine.operationCompleted", "运行器控制操作已成功执行"),
                duration: getConstant(ConstantCategory.FEATURE, 'TOAST_DURATION', 3000),
            })
            queryClient.invalidateQueries({ queryKey: ['runner-status'] })
        },
        onError: (error: ApiError) => {
            console.error('控制运行器失败:', error)
            setError(error.message || "控制运行器时出现错误")
            toast({
                title: t("globalSetting.engine.operationFailed", "操作失败"),
                description: error.message || t("globalSetting.engine.operationError", "控制运行器时出现错误"),
                variant: "destructive",
                duration: getConstant(ConstantCategory.FEATURE, 'TOAST_DURATION', 3000),
            })
        }
    })

    return {
        controlRunner: mutation.mutate,
        isLoading: mutation.isPending,
        error,
        clearError: () => setError(null),
    }
}