import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { certificatesApi } from '@/api/certificate'
import { CertificateCreateRequest, CertificateUpdateRequest } from '@/types/certificate'
import { useToast } from '@/hooks/use-toast'
import { useTranslation } from 'react-i18next'

export const useCreateCertificate = () => {
    const queryClient = useQueryClient()
    const { toast } = useToast()
    const { t } = useTranslation()
    const [error, setError] = useState<string | null>(null)

    const mutation = useMutation({
        mutationFn: (data: CertificateCreateRequest) => certificatesApi.createCertificate(data),
        onSuccess: () => {
            toast({
                title: t('certificate.toast.createSuccess'),
                description: t('certificate.toast.certificateCreated'),
            })
            queryClient.invalidateQueries({ queryKey: ['certificates'] })
        },
        onError: (error: ApiError) => {
            console.error(t('certificate.toast.createFailed'), error)
            setError(error.message || t('certificate.toast.createError'))
            toast({
                title: t('certificate.toast.createFailed'),
                description: error.message || t('certificate.toast.createError'),
                variant: "destructive",
            })
        }
    })

    return {
        createCertificate: mutation.mutate,
        isLoading: mutation.isPending,
        error,
        clearError: () => setError(null),
    }
}

export const useDeleteCertificate = () => {
    const queryClient = useQueryClient()
    const { toast } = useToast()
    const { t } = useTranslation()
    const [error, setError] = useState<string | null>(null)

    const mutation = useMutation({
        mutationFn: (id: string) => certificatesApi.deleteCertificate(id),
        onSuccess: () => {
            toast({
                title: t('certificate.toast.deleteSuccess'),
                description: t('certificate.toast.certificateDeleted'),
            })
            queryClient.invalidateQueries({ queryKey: ['certificates'] })
        },
        onError: (error: ApiError) => {
            console.error(t('certificate.toast.deleteFailed'), error)
            setError(error.message || t('certificate.toast.deleteError'))
            toast({
                title: t('certificate.toast.deleteFailed'),
                description: error.message || t('certificate.toast.deleteError'),
                variant: "destructive",
            })
        }
    })

    return {
        deleteCertificate: mutation.mutate,
        isLoading: mutation.isPending,
        error,
        clearError: () => setError(null),
    }
}

export const useUpdateCertificate = () => {
    const queryClient = useQueryClient()
    const { toast } = useToast()
    const { t } = useTranslation()
    const [error, setError] = useState<string | null>(null)

    const mutation = useMutation({
        mutationFn: ({ id, data }: { id: string, data: CertificateUpdateRequest }) =>
            certificatesApi.updateCertificate(id, data),
        onSuccess: () => {
            toast({
                title: t('certificate.toast.updateSuccess'),
                description: t('certificate.toast.certificateUpdated'),
            })
            queryClient.invalidateQueries({ queryKey: ['certificates'] })
        },
        onError: (error: ApiError) => {
            console.error(t('certificate.toast.updateFailed'), error)
            setError(error.message || t('certificate.toast.updateError'))
            toast({
                title: t('certificate.toast.updateFailed'),
                description: error.message || t('certificate.toast.updateError'),
                variant: "destructive",
            })
        }
    })

    return {
        updateCertificate: mutation.mutate,
        isLoading: mutation.isPending,
        error,
        clearError: () => setError(null),
    }
} 