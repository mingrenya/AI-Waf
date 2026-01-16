import { useState, useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'
import { Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Loader2, Save, RefreshCw } from 'lucide-react'
import { adaptiveThrottlingApi } from '@/api/adaptive-throttling'
import { useToast } from '@/hooks/use-toast'

const configSchema = z.object({
    enabled: z.boolean(),
    learningMode: z.object({
        enabled: z.boolean(),
        learningDuration: z.number().min(3600).max(604800),
        sampleInterval: z.number().min(10).max(3600),
        minSamples: z.number().min(10).max(10000)
    }),
    baseline: z.object({
        calculationMethod: z.enum(['mean', 'median', 'percentile']),
        percentile: z.number().min(0).max(100),
        updateInterval: z.number().min(60).max(86400),
        historyWindow: z.number().min(86400).max(2592000)
    }),
    autoAdjustment: z.object({
        enabled: z.boolean(),
        anomalyThreshold: z.number().min(1).max(10),
        minThreshold: z.number().min(1).max(1000),
        maxThreshold: z.number().min(100).max(100000),
        adjustmentFactor: z.number().min(1).max(5),
        cooldownPeriod: z.number().min(60).max(3600),
        gradualAdjustment: z.boolean(),
        adjustmentStepRatio: z.number().min(0.01).max(1)
    }),
    applyTo: z.object({
        visitLimit: z.boolean(),
        attackLimit: z.boolean(),
        errorLimit: z.boolean()
    })
})

type ConfigFormValues = z.infer<typeof configSchema>

export function AdaptiveConfigForm() {
    const { t } = useTranslation()
    const { toast } = useToast()
    const [loading, setLoading] = useState(false)
    const [fetching, setFetching] = useState(true)

    const form = useForm<ConfigFormValues>({
        resolver: zodResolver(configSchema),
        defaultValues: {
            enabled: false,
            learningMode: {
                enabled: true,
                learningDuration: 86400,
                sampleInterval: 60,
                minSamples: 100
            },
            baseline: {
                calculationMethod: 'percentile',
                percentile: 95,
                updateInterval: 3600,
                historyWindow: 604800
            },
            autoAdjustment: {
                enabled: true,
                anomalyThreshold: 2.0,
                minThreshold: 10,
                maxThreshold: 10000,
                adjustmentFactor: 1.5,
                cooldownPeriod: 300,
                gradualAdjustment: true,
                adjustmentStepRatio: 0.1
            },
            applyTo: {
                visitLimit: true,
                attackLimit: true,
                errorLimit: false
            }
        }
    })

    useEffect(() => {
        fetchConfig()
    }, [])

    const fetchConfig = async () => {
        try {
            setFetching(true)
            const config = await adaptiveThrottlingApi.getConfig()
            form.reset(config)
        } catch (error) {
            // 配置不存在时使用默认值
            console.log('No existing config, using defaults')
        } finally {
            setFetching(false)
        }
    }

    const onSubmit = async (data: ConfigFormValues) => {
        try {
            setLoading(true)
            await adaptiveThrottlingApi.updateConfig(data)
            toast({
                title: t('adaptiveThrottling.config.saveSuccess', '配置保存成功')
            })
        } catch (error) {
            toast({
                title: t('adaptiveThrottling.config.saveError', '配置保存失败'),
                variant: 'destructive'
            })
        } finally {
            setLoading(false)
        }
    }

    if (fetching) {
        return (
            <div className="flex items-center justify-center py-12">
                <Loader2 className="h-8 w-8 animate-spin text-primary" />
            </div>
        )
    }

    return (
        <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
                {/* 启用开关 */}
                <FormField
                    control={form.control}
                    name="enabled"
                    render={({ field }) => (
                        <FormItem className="flex items-center justify-between rounded-lg border p-4">
                            <div className="space-y-0.5">
                                <FormLabel className="text-base">
                                    {t('adaptiveThrottling.config.enabled', '启用自适应限流')}
                                </FormLabel>
                                <FormDescription>
                                    {t('adaptiveThrottling.config.enabledDesc', '开启后系统将自动学习流量模式并调整限流阈值')}
                                </FormDescription>
                            </div>
                            <FormControl>
                                <Switch checked={field.value} onCheckedChange={field.onChange} />
                            </FormControl>
                        </FormItem>
                    )}
                />

                {/* 学习模式配置 */}
                <div className="space-y-4">
                    <h3 className="text-lg font-medium">{t('adaptiveThrottling.config.learningMode', '学习模式')}</h3>
                    
                    <FormField
                        control={form.control}
                        name="learningMode.enabled"
                        render={({ field }) => (
                            <FormItem className="flex items-center justify-between rounded-lg border p-4">
                                <div className="space-y-0.5">
                                    <FormLabel>{t('adaptiveThrottling.config.learningEnabled', '启用学习模式')}</FormLabel>
                                </div>
                                <FormControl>
                                    <Switch checked={field.value} onCheckedChange={field.onChange} />
                                </FormControl>
                            </FormItem>
                        )}
                    />

                    <div className="grid grid-cols-3 gap-4">
                        <FormField
                            control={form.control}
                            name="learningMode.learningDuration"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>{t('adaptiveThrottling.config.learningDuration', '学习周期(秒)')}</FormLabel>
                                    <FormControl>
                                        <Input type="number" {...field} onChange={e => field.onChange(Number(e.target.value))} />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        <FormField
                            control={form.control}
                            name="learningMode.sampleInterval"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>{t('adaptiveThrottling.config.sampleInterval', '采样间隔(秒)')}</FormLabel>
                                    <FormControl>
                                        <Input type="number" {...field} onChange={e => field.onChange(Number(e.target.value))} />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        <FormField
                            control={form.control}
                            name="learningMode.minSamples"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>{t('adaptiveThrottling.config.minSamples', '最小样本数')}</FormLabel>
                                    <FormControl>
                                        <Input type="number" {...field} onChange={e => field.onChange(Number(e.target.value))} />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />
                    </div>
                </div>

                {/* 基线配置 */}
                <div className="space-y-4">
                    <h3 className="text-lg font-medium">{t('adaptiveThrottling.config.baseline', '基线配置')}</h3>
                    
                    <div className="grid grid-cols-2 gap-4">
                        <FormField
                            control={form.control}
                            name="baseline.calculationMethod"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>{t('adaptiveThrottling.config.calculationMethod', '计算方法')}</FormLabel>
                                    <Select onValueChange={field.onChange} defaultValue={field.value}>
                                        <FormControl>
                                            <SelectTrigger>
                                                <SelectValue />
                                            </SelectTrigger>
                                        </FormControl>
                                        <SelectContent>
                                            <SelectItem value="mean">{t('adaptiveThrottling.method.mean', '均值')}</SelectItem>
                                            <SelectItem value="median">{t('adaptiveThrottling.method.median', '中位数')}</SelectItem>
                                            <SelectItem value="percentile">{t('adaptiveThrottling.method.percentile', '百分位数')}</SelectItem>
                                        </SelectContent>
                                    </Select>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        <FormField
                            control={form.control}
                            name="baseline.percentile"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>{t('adaptiveThrottling.config.percentile', '百分位数')}</FormLabel>
                                    <FormControl>
                                        <Input type="number" {...field} onChange={e => field.onChange(Number(e.target.value))} />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        <FormField
                            control={form.control}
                            name="baseline.updateInterval"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>{t('adaptiveThrottling.config.updateInterval', '更新间隔(秒)')}</FormLabel>
                                    <FormControl>
                                        <Input type="number" {...field} onChange={e => field.onChange(Number(e.target.value))} />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        <FormField
                            control={form.control}
                            name="baseline.historyWindow"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>{t('adaptiveThrottling.config.historyWindow', '历史窗口(秒)')}</FormLabel>
                                    <FormControl>
                                        <Input type="number" {...field} onChange={e => field.onChange(Number(e.target.value))} />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />
                    </div>
                </div>

                {/* 自动调整策略 */}
                <div className="space-y-4">
                    <h3 className="text-lg font-medium">{t('adaptiveThrottling.config.autoAdjustment', '自动调整策略')}</h3>
                    
                    <FormField
                        control={form.control}
                        name="autoAdjustment.enabled"
                        render={({ field }) => (
                            <FormItem className="flex items-center justify-between rounded-lg border p-4">
                                <div className="space-y-0.5">
                                    <FormLabel>{t('adaptiveThrottling.config.autoAdjustmentEnabled', '启用自动调整')}</FormLabel>
                                </div>
                                <FormControl>
                                    <Switch checked={field.value} onCheckedChange={field.onChange} />
                                </FormControl>
                            </FormItem>
                        )}
                    />

                    <div className="grid grid-cols-2 gap-4">
                        <FormField
                            control={form.control}
                            name="autoAdjustment.anomalyThreshold"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>{t('adaptiveThrottling.config.anomalyThreshold', '异常阈值(倍数)')}</FormLabel>
                                    <FormControl>
                                        <Input type="number" step="0.1" {...field} onChange={e => field.onChange(Number(e.target.value))} />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        <FormField
                            control={form.control}
                            name="autoAdjustment.adjustmentFactor"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>{t('adaptiveThrottling.config.adjustmentFactor', '调整因子')}</FormLabel>
                                    <FormControl>
                                        <Input type="number" step="0.1" {...field} onChange={e => field.onChange(Number(e.target.value))} />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        <FormField
                            control={form.control}
                            name="autoAdjustment.minThreshold"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>{t('adaptiveThrottling.config.minThreshold', '最小阈值')}</FormLabel>
                                    <FormControl>
                                        <Input type="number" {...field} onChange={e => field.onChange(Number(e.target.value))} />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        <FormField
                            control={form.control}
                            name="autoAdjustment.maxThreshold"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>{t('adaptiveThrottling.config.maxThreshold', '最大阈值')}</FormLabel>
                                    <FormControl>
                                        <Input type="number" {...field} onChange={e => field.onChange(Number(e.target.value))} />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        <FormField
                            control={form.control}
                            name="autoAdjustment.cooldownPeriod"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>{t('adaptiveThrottling.config.cooldownPeriod', '冷却期(秒)')}</FormLabel>
                                    <FormControl>
                                        <Input type="number" {...field} onChange={e => field.onChange(Number(e.target.value))} />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        <FormField
                            control={form.control}
                            name="autoAdjustment.adjustmentStepRatio"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>{t('adaptiveThrottling.config.adjustmentStepRatio', '调整步长比例')}</FormLabel>
                                    <FormControl>
                                        <Input type="number" step="0.01" {...field} onChange={e => field.onChange(Number(e.target.value))} />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />
                    </div>

                    <FormField
                        control={form.control}
                        name="autoAdjustment.gradualAdjustment"
                        render={({ field }) => (
                            <FormItem className="flex items-center justify-between rounded-lg border p-4">
                                <div className="space-y-0.5">
                                    <FormLabel>{t('adaptiveThrottling.config.gradualAdjustment', '渐进式调整')}</FormLabel>
                                    <FormDescription>
                                        {t('adaptiveThrottling.config.gradualAdjustmentDesc', '启用后将分步调整阈值，避免剧烈波动')}
                                    </FormDescription>
                                </div>
                                <FormControl>
                                    <Switch checked={field.value} onCheckedChange={field.onChange} />
                                </FormControl>
                            </FormItem>
                        )}
                    />
                </div>

                {/* 应用范围 */}
                <div className="space-y-4">
                    <h3 className="text-lg font-medium">{t('adaptiveThrottling.config.applyTo', '应用范围')}</h3>
                    
                    <div className="grid grid-cols-3 gap-4">
                        <FormField
                            control={form.control}
                            name="applyTo.visitLimit"
                            render={({ field }) => (
                                <FormItem className="flex items-center justify-between rounded-lg border p-4">
                                    <div className="space-y-0.5">
                                        <FormLabel>{t('adaptiveThrottling.config.visitLimit', '访问限流')}</FormLabel>
                                    </div>
                                    <FormControl>
                                        <Switch checked={field.value} onCheckedChange={field.onChange} />
                                    </FormControl>
                                </FormItem>
                            )}
                        />

                        <FormField
                            control={form.control}
                            name="applyTo.attackLimit"
                            render={({ field }) => (
                                <FormItem className="flex items-center justify-between rounded-lg border p-4">
                                    <div className="space-y-0.5">
                                        <FormLabel>{t('adaptiveThrottling.config.attackLimit', '攻击限流')}</FormLabel>
                                    </div>
                                    <FormControl>
                                        <Switch checked={field.value} onCheckedChange={field.onChange} />
                                    </FormControl>
                                </FormItem>
                            )}
                        />

                        <FormField
                            control={form.control}
                            name="applyTo.errorLimit"
                            render={({ field }) => (
                                <FormItem className="flex items-center justify-between rounded-lg border p-4">
                                    <div className="space-y-0.5">
                                        <FormLabel>{t('adaptiveThrottling.config.errorLimit', '错误限流')}</FormLabel>
                                    </div>
                                    <FormControl>
                                        <Switch checked={field.value} onCheckedChange={field.onChange} />
                                    </FormControl>
                                </FormItem>
                            )}
                        />
                    </div>
                </div>

                {/* 操作按钮 */}
                <div className="flex justify-end gap-4">
                    <Button type="button" variant="outline" onClick={fetchConfig} disabled={loading}>
                        <RefreshCw className="mr-2 h-4 w-4" />
                        {t('common.reset', '重置')}
                    </Button>
                    <Button type="submit" disabled={loading}>
                        {loading ? (
                            <>
                                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                                {t('common.saving', '保存中...')}
                            </>
                        ) : (
                            <>
                                <Save className="mr-2 h-4 w-4" />
                                {t('common.save', '保存')}
                            </>
                        )}
                    </Button>
                </div>
            </form>
        </Form>
    )
}
