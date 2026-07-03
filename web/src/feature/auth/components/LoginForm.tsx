import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { LockKeyhole, User, LogIn } from 'lucide-react'
import { Form, FormControl, FormField, FormItem, FormMessage } from '@/components/ui/form'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { useLogin } from '../hooks'
import { loginSchema, LoginFormValues } from '@/validation/auth'
import { useTranslation } from 'react-i18next'

export function LoginForm() {
    const { t } = useTranslation()
    const { login, isLoading, error, clearError } = useLogin()

    const form = useForm<LoginFormValues>({
        resolver: zodResolver(loginSchema),
        defaultValues: {
            username: '',
            password: '',
        },
    })

    const onSubmit = (values: LoginFormValues) => {
        clearError()
        login(values)
    }

    return (
        <div className="space-y-5">
            {error && (
                <Alert variant="destructive" className="mb-4 bg-red-500/10 border-red-500/30 backdrop-blur-sm animate-fade-in text-white">
                    <AlertDescription className="text-red-300">{error}</AlertDescription>
                </Alert>
            )}

            <Form {...form}>
                <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
                    <FormField
                        control={form.control}
                        name="username"
                        render={({ field }) => (
                            <FormItem className="animate-fade-in">
                                <FormControl>
                                    <div className="relative">
                                        <User className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-white/40" />
                                        <input
                                            placeholder={t('auth.usernamePlaceholder')}
                                            autoComplete="username"
                                            className="glass-input w-full pl-10 pr-4 py-3 text-sm outline-none text-white placeholder:text-white/30"
                                            {...field}
                                            onChange={(e) => {
                                                clearError()
                                                field.onChange(e)
                                            }}
                                        />
                                    </div>
                                </FormControl>
                                <FormMessage />
                            </FormItem>
                        )}
                    />

                    <FormField
                        control={form.control}
                        name="password"
                        render={({ field }) => (
                            <FormItem className="animate-fade-in [animation-delay:150ms]">
                                <FormControl>
                                    <div className="relative">
                                        <LockKeyhole className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-white/40" />
                                        <input
                                            type="password"
                                            placeholder={t('auth.passwordPlaceholder')}
                                            autoComplete="current-password"
                                            className="glass-input w-full pl-10 pr-4 py-3 text-sm outline-none text-white placeholder:text-white/30"
                                            {...field}
                                            onChange={(e) => {
                                                clearError()
                                                field.onChange(e)
                                            }}
                                        />
                                    </div>
                                </FormControl>
                                <FormMessage />
                            </FormItem>
                        )}
                    />

                    <button
                        type="submit"
                        disabled={isLoading}
                        className="glass-btn ripple-effect w-full py-3 flex items-center justify-center gap-2 text-sm animate-fade-in [animation-delay:300ms]"
                    >
                        {isLoading ? (
                            <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                        ) : (
                            <LogIn className="w-4 h-4" />
                        )}
                        {isLoading ? t('auth.loggingIn') : t('auth.login')}
                    </button>
                </form>
            </Form>

            <p className="text-center text-xs text-white/40 pt-2 border-t border-white/10">
                {t('auth.passwordRequirement')}
            </p>
        </div>
    )
}
