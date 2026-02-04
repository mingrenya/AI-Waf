import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { ruleApi } from '@/api/rule'
import { MicroRuleCreateRequest, MicroRuleUpdateRequest } from '@/types/rule'
import { toast } from '@/store'
import { useTranslation } from 'react-i18next'
import { queryKeys } from '@/lib/query-keys'

export const useCreateMicroRule = () => {
    const queryClient = useQueryClient()
    const { t } = useTranslation()
    const [error, setError] = useState<string | null>(null)

    const mutation = useMutation({
        mutationFn: (data: MicroRuleCreateRequest) => ruleApi.createMicroRule(data),
        onSuccess: () => {
            toast({
                title: t('microRule.toast.createSuccess'),
                description: t('microRule.toast.microRuleCreated'),
                variant: 'success',
            })
            queryClient.invalidateQueries({ queryKey: queryKeys.microRule.lists() })
        },
        onError: (error: ApiError) => {
            if (import.meta.env.DEV) {
                console.error(t('microRule.toast.createFailed'), error)
            }
            setError(error.message || t('microRule.toast.createError'))
            toast({
                title: t('microRule.toast.createFailed'),
                description: error.message || t('microRule.toast.createError'),
                variant: "destructive",
            })
        }
    })

    return {
        createMicroRule: mutation.mutate,
        isLoading: mutation.isPending,
        error,
        clearError: () => setError(null),
    }
}

export const useDeleteMicroRule = () => {
    const queryClient = useQueryClient()
    const { t } = useTranslation()
    const [error, setError] = useState<string | null>(null)

    const mutation = useMutation({
        mutationFn: (id: string) => ruleApi.deleteMicroRule(id),
        onSuccess: () => {
            toast({
                title: t('microRule.toast.deleteSuccess'),
                description: t('microRule.toast.microRuleDeleted'),
                variant: 'success',
            })
            queryClient.invalidateQueries({ queryKey: queryKeys.microRule.lists() })
        },
        onError: (error: ApiError) => {
            if (import.meta.env.DEV) {
                console.error(t('microRule.toast.deleteFailed'), error)
            }
            setError(error.message || t('microRule.toast.deleteError'))
            toast({
                title: t('microRule.toast.deleteFailed'),
                description: error.message || t('microRule.toast.deleteError'),
                variant: "destructive",
            })
        }
    })

    return {
        deleteMicroRule: mutation.mutate,
        isLoading: mutation.isPending,
        error,
        clearError: () => setError(null),
    }
}

export const useUpdateMicroRule = () => {
    const queryClient = useQueryClient()
    const { t } = useTranslation()
    const [error, setError] = useState<string | null>(null)

    const mutation = useMutation({
        mutationFn: ({ id, data }: { id: string, data: MicroRuleUpdateRequest }) =>
            ruleApi.updateMicroRule(id, data),
        onSuccess: () => {
            toast({
                title: t('microRule.toast.updateSuccess'),
                description: t('microRule.toast.microRuleUpdated'),
                variant: 'success',
            })
            queryClient.invalidateQueries({ queryKey: queryKeys.microRule.lists() })
        },
        onError: (error: ApiError) => {
            if (import.meta.env.DEV) {
                console.error(t('microRule.toast.updateFailed'), error)
            }
            setError(error.message || t('microRule.toast.updateError'))
            toast({
                title: t('microRule.toast.updateFailed'),
                description: error.message || t('microRule.toast.updateError'),
                variant: "destructive",
            })
        }
    })

    return {
        updateMicroRule: mutation.mutate,
        isLoading: mutation.isPending,
        error,
        clearError: () => setError(null),
    }
}