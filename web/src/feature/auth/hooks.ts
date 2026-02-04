import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router'
import { useQuery, useMutation } from '@tanstack/react-query'
import { authApi } from '@/api/services'
import { useAuthActions, useAuthStore, useUser, toast } from '@/store'
import { LoginFormValues, PasswordResetFormValues } from '@/validation/auth'
import { GetUserInfoResponseData } from '@/types/auth'
import { queryKeys } from '@/lib/query-keys'
import type { ApiError } from '@/api/index'

export const useLogin = () => {
    const navigate = useNavigate()
    const { login } = useAuthActions()
    const [error, setError] = useState<string | null>(null)

    const mutation = useMutation({
        mutationFn: (values: LoginFormValues) => authApi.login(values),
        onSuccess: (data) => {
            // Store token and user data
            login(data.token, data.user)

            // Redirect based on password reset requirement
            if (data.user.needReset) {
                navigate('/reset-password')
            } else {
                navigate('/')
            }
        },
        onError: (error: ApiError) => {
            const message = error.message || '登录失败，请检查用户名和密码'
            setError(message)
            toast({ title: '登录失败', description: message, variant: 'destructive' })
        }
    })

    return {
        login: mutation.mutate,
        isLoading: mutation.isPending,
        error,
        clearError: () => setError(null),
    }
}

export const useResetPassword = () => {
    const navigate = useNavigate()
    const [error, setError] = useState<string | null>(null)
    const { logout } = useAuthActions()
    const user = useUser()

    const mutation = useMutation({
        mutationFn: (values: Omit<PasswordResetFormValues, 'confirmPassword'>) => {
            return authApi.resetPassword({
                oldPassword: values.oldPassword,
                newPassword: values.newPassword,
            })
        },
        onSuccess: () => {
            // If the user was required to reset password, log them out
            // to force a new login with the new password
            if (user?.needReset) {
                logout()
                navigate('/login', { state: { message: '密码已重置，请使用新密码登录' } })
            } else {
                navigate('/')
            }
            toast({ title: '成功', description: '密码已重置', variant: 'success' })
        },
        onError: (error: ApiError) => {
            const message = error.message || '密码重置失败，请重试'
            setError(message)
            toast({ title: '重置失败', description: message, variant: 'destructive' })
        },
    })

    return {
        resetPassword: mutation.mutate,
        isLoading: mutation.isPending,
        error,
        clearError: () => setError(null),
    }
}


export const useCurrentUser = () => {
    const { setUser } = useAuthActions()
    const user = useUser()

    const { data, isLoading, error } = useQuery<GetUserInfoResponseData, ApiError>({
        queryKey: queryKeys.auth.currentUser(),
        queryFn: () => authApi.getCurrentUser(),
        enabled: !!useAuthStore.getState().token,
    })

    // 监听数据变化，替代 onSuccess
    useEffect(() => {
        if (data) {
            setUser(data)
        }
    }, [data, setUser])

    return {
        user: data || user,
        isLoading,
        error,
    }
}