import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { siteApi } from '@/api/site'
import { CreateSiteRequest, UpdateSiteRequest } from '@/types/site'
import { useToast } from '@/hooks/use-toast'
import { useTranslation } from 'react-i18next'

type ApiError = {
    message: string
}

/**
 * 创建站点Hook
 */
export const useCreateSite = () => {
    const queryClient = useQueryClient()
    const { toast } = useToast()
    const [error, setError] = useState<string | null>(null)
    const { t } = useTranslation()

    const mutation = useMutation({
        mutationFn: (data: CreateSiteRequest) => siteApi.createSite(data),
        onSuccess: () => {
            toast({
                title: t("site.toast.createSuccess", "创建成功"),
                description: t("site.toast.siteCreated", "站点已成功创建"),
            })
            queryClient.invalidateQueries({ queryKey: ['sites'] })
        },
        onError: (error: ApiError) => {
            console.error('创建站点失败:', error)
            setError(error.message || "创建站点时出现错误")
            toast({
                title: t("site.toast.createFailed", "创建失败"),
                description: error.message || t("site.toast.createError", "创建站点时出现错误"),
                variant: "destructive",
            })
        }
    })

    return {
        createSite: mutation.mutate,
        isLoading: mutation.isPending,
        error,
        clearError: () => setError(null),
    }
}

/**
 * 删除站点Hook
 */
export const useDeleteSite = () => {
    const queryClient = useQueryClient()
    const { toast } = useToast()
    const [error, setError] = useState<string | null>(null)
    const { t } = useTranslation()

    const mutation = useMutation({
        mutationFn: (id: string) => siteApi.deleteSite(id),
        onSuccess: () => {
            toast({
                title: t("site.toast.deleteSuccess", "删除成功"),
                description: t("site.toast.siteDeleted", "站点已成功删除"),
            })
            queryClient.invalidateQueries({ queryKey: ['sites'] })
        },
        onError: (error: ApiError) => {
            console.error('删除站点失败:', error)
            setError(error.message || "删除站点时出现错误")
            toast({
                title: t("site.toast.deleteFailed", "删除失败"),
                description: error.message || t("site.toast.deleteError", "删除站点时出现错误"),
                variant: "destructive",
            })
        }
    })

    return {
        deleteSite: mutation.mutate,
        isLoading: mutation.isPending,
        error,
        clearError: () => setError(null),
    }
}

/**
 * 更新站点Hook
 */
export const useUpdateSite = () => {
    const queryClient = useQueryClient()
    const { toast } = useToast()
    const [error, setError] = useState<string | null>(null)
    const { t } = useTranslation()

    const mutation = useMutation({
        mutationFn: ({ id, data }: { id: string, data: UpdateSiteRequest }) =>
            siteApi.updateSite(id, data),
        onSuccess: () => {
            toast({
                title: t("site.toast.updateSuccess", "更新成功"),
                description: t("site.toast.siteUpdated", "站点已成功更新"),
            })
            queryClient.invalidateQueries({ queryKey: ['sites'] })
        },
        onError: (error: ApiError) => {
            console.error('更新站点失败:', error)
            setError(error.message || "更新站点时出现错误")
            toast({
                title: t("site.toast.updateFailed", "更新失败"),
                description: error.message || t("site.toast.updateError", "更新站点时出现错误"),
                variant: "destructive",
            })
        }
    })

    return {
        updateSite: mutation.mutate,
        isLoading: mutation.isPending,
        error,
        clearError: () => setError(null),
    }
}