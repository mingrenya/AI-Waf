import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import {
    Form,
    FormControl,
    FormField,
    FormItem,
    FormLabel,
    FormMessage,
    FormDescription
} from '@/components/ui/form'
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { AlertCircle, Loader2 } from 'lucide-react'
import { AlertChannelType, CreateAlertChannelRequest, UpdateAlertChannelRequest } from '@/types/alert'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { alertChannelApi } from '@/api/alert'
import { useToast } from '@/hooks/use-toast'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

// 表单验证Schema
const channelFormSchema = z.object({
    name: z.string().min(1, 'Channel name is required'),
    type: z.nativeEnum(AlertChannelType),
    enabled: z.boolean(),
    config: z.record(z.any())
})

interface ChannelFormProps {
    mode: 'create' | 'update'
    channelId?: string
    onSuccess?: () => void
    defaultValues?: Partial<CreateAlertChannelRequest>
}

export function ChannelForm({
    mode = 'create',
    channelId,
    onSuccess,
    defaultValues = {
        name: '',
        type: AlertChannelType.Webhook,
        enabled: true,
        config: {}
    }
}: ChannelFormProps) {
    const { t } = useTranslation()
    const { toast } = useToast()
    const queryClient = useQueryClient()
    const [selectedType, setSelectedType] = useState<AlertChannelType>(
        defaultValues.type || AlertChannelType.Webhook
    )

    const form = useForm<CreateAlertChannelRequest>({
        resolver: zodResolver(channelFormSchema),
        defaultValues,
    })

    // 创建mutation
    const createMutation = useMutation({
        mutationFn: (data: CreateAlertChannelRequest) => alertChannelApi.createChannel(data),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['alertChannels'] })
            toast({ title: 'Success', description: t('alert.createSuccess') })
            onSuccess?.()
        },
        onError: (error: any) => {
            toast({ title: 'Error', description: error?.response?.data?.message || t('alert.createFailed'), variant: 'destructive' })
        }
    })

    // 更新mutation
    const updateMutation = useMutation({
        mutationFn: (data: { id: string, payload: UpdateAlertChannelRequest }) =>
            alertChannelApi.updateChannel(data.id, data.payload),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['alertChannels'] })
            toast({ title: 'Success', description: t('alert.updateSuccess') })
            onSuccess?.()
        },
        onError: (error: any) => {
            toast({ title: 'Error', description: error?.response?.data?.message || t('alert.updateFailed'), variant: 'destructive' })
        }
    })

    const isLoading = createMutation.isPending || updateMutation.isPending
    const error = createMutation.error || updateMutation.error

    const onSubmit = (data: CreateAlertChannelRequest) => {
        if (mode === 'create') {
            createMutation.mutate(data)
        } else if (channelId) {
            updateMutation.mutate({ id: channelId, payload: data })
        }
    }

    // 根据类型渲染配置字段
    const renderConfigFields = () => {
        switch (selectedType) {
            case AlertChannelType.Webhook:
                return (
                    <>
                        <FormField
                            control={form.control}
                            name="config.url"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>Webhook URL *</FormLabel>
                                    <FormControl>
                                        <Input placeholder="https://example.com/webhook" {...field} />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />
                        <FormField
                            control={form.control}
                            name="config.method"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>HTTP Method</FormLabel>
                                    <Select onValueChange={field.onChange} defaultValue={field.value || 'POST'}>
                                        <FormControl>
                                            <SelectTrigger>
                                                <SelectValue />
                                            </SelectTrigger>
                                        </FormControl>
                                        <SelectContent>
                                            <SelectItem value="POST">POST</SelectItem>
                                            <SelectItem value="PUT">PUT</SelectItem>
                                        </SelectContent>
                                    </Select>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />
                    </>
                )

            case AlertChannelType.Slack:
                return (
                    <>
                        <FormField
                            control={form.control}
                            name="config.token"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>Bot Token *</FormLabel>
                                    <FormControl>
                                        <Input type="password" placeholder="xoxb-..." {...field} />
                                    </FormControl>
                                    <FormDescription>Slack Bot OAuth Token</FormDescription>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />
                        <FormField
                            control={form.control}
                            name="config.channel"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>Channel *</FormLabel>
                                    <FormControl>
                                        <Input placeholder="#alerts" {...field} />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />
                    </>
                )

            case AlertChannelType.Discord:
                return (
                    <>
                        <FormField
                            control={form.control}
                            name="config.webhookUrl"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>Webhook URL *</FormLabel>
                                    <FormControl>
                                        <Input placeholder="https://discord.com/api/webhooks/..." {...field} />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />
                        <FormField
                            control={form.control}
                            name="config.username"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>Bot Username</FormLabel>
                                    <FormControl>
                                        <Input placeholder="WAF Alert Bot" {...field} />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />
                    </>
                )

            case AlertChannelType.DingTalk:
                return (
                    <>
                        <FormField
                            control={form.control}
                            name="config.accessToken"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>Access Token *</FormLabel>
                                    <FormControl>
                                        <Input type="password" placeholder="..." {...field} />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />
                        <FormField
                            control={form.control}
                            name="config.secret"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>Secret Key</FormLabel>
                                    <FormControl>
                                        <Input type="password" placeholder="..." {...field} />
                                    </FormControl>
                                    <FormDescription>Optional, for signature verification</FormDescription>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />
                    </>
                )

            case AlertChannelType.WeCom:
                return (
                    <FormField
                        control={form.control}
                        name="config.webhookKey"
                        render={({ field }) => (
                            <FormItem>
                                <FormLabel>Webhook Key *</FormLabel>
                                <FormControl>
                                    <Input placeholder="..." {...field} />
                                </FormControl>
                                <FormDescription>Enterprise WeChat webhook key</FormDescription>
                                <FormMessage />
                            </FormItem>
                        )}
                    />
                )

            default:
                return null
        }
    }

    return (
        <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
                {error && (
                    <Alert variant="destructive">
                        <AlertCircle className="h-4 w-4" />
                        <AlertDescription>{String(error)}</AlertDescription>
                    </Alert>
                )}

                <div className="space-y-4">
                    <h3 className="text-lg font-medium">Basic Information</h3>

                    <FormField
                        control={form.control}
                        name="name"
                        render={({ field }) => (
                            <FormItem>
                                <FormLabel>{t('alert.form.channelName')} *</FormLabel>
                                <FormControl>
                                    <Input placeholder="My Alert Channel" {...field} />
                                </FormControl>
                                <FormMessage />
                            </FormItem>
                        )}
                    />

                    <FormField
                        control={form.control}
                        name="type"
                        render={({ field }) => (
                            <FormItem>
                                <FormLabel>{t('alert.form.channelTypeLabel')}</FormLabel>
                                <Select
                                    onValueChange={(value) => {
                                        field.onChange(value)
                                        setSelectedType(value as AlertChannelType)
                                    }}
                                    defaultValue={field.value}
                                    disabled={mode === 'update'}
                                >
                                    <FormControl>
                                        <SelectTrigger>
                                            <SelectValue />
                                        </SelectTrigger>
                                    </FormControl>
                                    <SelectContent>
                                        <SelectItem value={AlertChannelType.Webhook}>Webhook</SelectItem>
                                        <SelectItem value={AlertChannelType.Slack}>Slack</SelectItem>
                                        <SelectItem value={AlertChannelType.Discord}>Discord</SelectItem>
                                        <SelectItem value={AlertChannelType.DingTalk}>DingTalk</SelectItem>
                                        <SelectItem value={AlertChannelType.WeCom}>Enterprise WeChat</SelectItem>
                                    </SelectContent>
                                </Select>
                                {mode === 'update' && (
                                    <FormDescription>{t('alert.form.channelTypeDisabled')}</FormDescription>
                                )}
                                <FormMessage />
                            </FormItem>
                        )}
                    />

                    <FormField
                        control={form.control}
                        name="enabled"
                        render={({ field }) => (
                            <FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
                                <div className="space-y-0.5">
                                    <FormLabel className="text-base">{t('alert.form.enableChannel')}</FormLabel>
                                    <FormDescription>
                                        {t('alert.form.enableChannelDescription')}
                                    </FormDescription>
                                </div>
                                <FormControl>
                                    <Switch
                                        checked={field.value}
                                        onCheckedChange={field.onChange}
                                    />
                                </FormControl>
                            </FormItem>
                        )}
                    />
                </div>

                <div className="space-y-4">
                    <h3 className="text-lg font-medium">{t('alert.form.channelConfiguration')}</h3>
                    {renderConfigFields()}
                </div>

                <div className="flex justify-end gap-2 pt-4">
                    <Button type="submit" disabled={isLoading}>
                        {isLoading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                        {mode === 'create' ? t('common.create') : t('common.save')}
                    </Button>
                </div>
            </form>
        </Form>
    )
}
