import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Lock } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { useResetPassword } from '../hooks'
import { passwordResetSchema, PasswordResetFormValues } from '@/validation/auth'
import useAuthStore from '@/store/auth'
import { useTranslation } from 'react-i18next'

export function PasswordResetForm() {
    const { t } = useTranslation()
    const { resetPassword, isLoading, error, clearError } = useResetPassword()
    const { user } = useAuthStore()

    const form = useForm<PasswordResetFormValues>({
        resolver: zodResolver(passwordResetSchema),
        defaultValues: {
            oldPassword: '',
            newPassword: '',
            confirmPassword: '',
        },
    })

    const onSubmit = (values: PasswordResetFormValues) => {
        clearError()
        resetPassword({
            oldPassword: values.oldPassword,
            newPassword: values.newPassword,
        })
    }

    return (
        <Card className="w-full max-w-md mx-auto surface-card border-0 transition-all hover:shadow-2xl duration-300" style={{ backdropFilter: 'blur(18px)', WebkitBackdropFilter: 'blur(18px)' }}>
            <CardHeader className="space-y-1 pb-2">
                <CardTitle className="text-2xl font-bold text-center" style={{ color: 'var(--text-primary)' }}>{t('auth.resetPassword')}</CardTitle>
                <CardDescription className="text-center" style={{ color: 'var(--text-secondary)' }}>
                    {user?.needReset
                        ? t('auth.firstLoginReset')
                        : t('auth.enterOldAndNewPassword')}
                </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4 pt-4">
                {error && (
                    <Alert variant="destructive" className="mb-4 bg-red-50 dark:bg-red-950/30 border-red-200 dark:border-red-800/50 animate-fade-in-up">
                        <AlertDescription className="text-red-700 dark:text-red-300">{error}</AlertDescription>
                    </Alert>
                )}

                <Form {...form}>
                    <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
                        <FormField
                            control={form.control}
                            name="oldPassword"
                            render={({ field }) => (
                                <FormItem className="transition-all duration-300 animate-fade-in-up">
                                    <FormLabel style={{ color: 'var(--text-secondary)' }}>{t('auth.currentPassword')}</FormLabel>
                                    <FormControl>
                                        <div className="relative group">
                                            <Lock className="absolute left-3 top-3 h-5 w-5 text-muted-foreground group-hover:text-purple-500 transition-colors duration-300" />
                                            <Input
                                                type="password"
                                                placeholder={t('auth.enterCurrentPassword')}
                                                className="pl-10 py-6 bg-black/5 dark:bg-white/5 border-transparent focus:bg-white/10 transition-all group-hover:border-purple-300/50"
                                                style={{ color: 'var(--text-primary)' }}
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
                            name="newPassword"
                            render={({ field }) => (
                                <FormItem className="transition-all duration-300 animate-fade-in-up [animation-delay:150ms]">
                                    <FormLabel style={{ color: 'var(--text-secondary)' }}>{t('auth.newPassword')}</FormLabel>
                                    <FormControl>
                                        <div className="relative group">
                                            <Lock className="absolute left-3 top-3 h-5 w-5 text-muted-foreground group-hover:text-purple-500 transition-colors duration-300" />
                                            <Input
                                                type="password"
                                                placeholder={t('auth.enterNewPassword')}
                                                className="pl-10 py-6 bg-black/5 dark:bg-white/5 border-transparent focus:bg-white/10 transition-all group-hover:border-purple-300/50"
                                                style={{ color: 'var(--text-primary)' }}
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
                            name="confirmPassword"
                            render={({ field }) => (
                                <FormItem className="transition-all duration-300 animate-fade-in-up [animation-delay:300ms]">
                                    <FormLabel style={{ color: 'var(--text-secondary)' }}>{t('auth.confirmNewPassword')}</FormLabel>
                                    <FormControl>
                                        <div className="relative group">
                                            <Lock className="absolute left-3 top-3 h-5 w-5 text-muted-foreground group-hover:text-purple-500 transition-colors duration-300" />
                                            <Input
                                                type="password"
                                                placeholder={t('auth.enterNewPasswordAgain')}
                                                className="pl-10 py-6 bg-black/5 dark:bg-white/5 border-transparent focus:bg-white/10 transition-all group-hover:border-purple-300/50"
                                                style={{ color: 'var(--text-primary)' }}
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

                        <Button
                            type="submit"
                            className="w-full mt-6 py-6 !text-white bg-gradient-to-r from-purple-600 to-indigo-600 hover:from-purple-700 hover:to-indigo-700 transition-all shadow-md hover:shadow-lg animate-fade-in-up [animation-delay:450ms] hover:translate-y-[-2px]"
                            disabled={isLoading}
                        >
                            {isLoading ? t('auth.submitting') : t('auth.resetPassword')}
                        </Button>
                    </form>
                </Form>
            </CardContent>
            <CardFooter className="flex justify-center pt-4" style={{ borderTop: '1px solid var(--surface-card-border)' }}>
                <p className="text-sm" style={{ color: 'var(--text-muted)' }}>
                    {t('auth.passwordRequirement')}
                </p>
            </CardFooter>
        </Card>
    )
}