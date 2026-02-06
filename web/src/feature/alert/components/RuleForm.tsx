import { useForm, useFieldArray } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import {
    Form,
    FormControl,
    FormField,
    FormItem,
    FormLabel,
    FormMessage,
    FormDescription,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { AlertCircle, Loader2, PlusCircle, Trash2 } from 'lucide-react'
import {
    AlertSeverity,
    ConditionOperator,
    type CreateAlertRuleRequest,
    type UpdateAlertRuleRequest,
} from '@/types/alert'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { alertRuleApi } from '@/api/alert'
import { toast } from '@/store'
import { useTranslation } from 'react-i18next'
import { queryKeys } from '@/lib/query-keys'

// 表单内部使用的值类型（数值以字符串形式输入，提交时再转换）
export interface RuleFormValues {
    name: string
    description?: string
    enabled: boolean
    severity: AlertSeverity
    logic: 'AND' | 'OR'
    cooldown: string
    template: string
    channels: string
    conditions: {
        metric: string
        operator: ConditionOperator
        threshold: string
        duration?: string
    }[]
}

const conditionSchema = z.object({
    metric: z.string().min(1, 'Metric is required'),
    operator: z.nativeEnum(ConditionOperator),
    threshold: z
        .string()
        .min(1, 'Threshold is required')
        .refine((val) => !Number.isNaN(Number(val)), 'Threshold must be a number'),
    duration: z
        .string()
        .optional()
        .refine(
            (val) => !val || !Number.isNaN(Number(val)),
            'Duration must be a number',
        ),
})

const ruleFormSchema = z.object({
    name: z.string().min(1, 'Rule name is required'),
    description: z.string().optional(),
    enabled: z.boolean(),
    severity: z.nativeEnum(AlertSeverity),
    logic: z.enum(['AND', 'OR']),
    cooldown: z
        .string()
        .min(1, 'Cooldown is required')
        .refine((val) => !Number.isNaN(Number(val)), 'Cooldown must be a number'),
    template: z.string().min(1, 'Template is required'),
    channels: z
        .string()
        .min(1, 'At least one channel id is required'),
    conditions: z
        .array(conditionSchema)
        .min(1, 'At least one condition is required'),
})

interface RuleFormProps {
    mode: 'create' | 'update'
    ruleId?: string
    onSuccess?: () => void
    defaultValues?: RuleFormValues
}

export function RuleForm({
    mode,
    ruleId,
    onSuccess,
    defaultValues,
}: RuleFormProps) {
    const { t } = useTranslation()
    const queryClient = useQueryClient()

    const form = useForm<RuleFormValues>({
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        resolver: zodResolver(ruleFormSchema) as any,
        defaultValues:
            defaultValues ??
            ({
                name: '',
                description: '',
                enabled: true,
                logic: 'AND',
                severity: AlertSeverity.Medium,
                cooldown: '60',
                template: '',
                channels: '',
                conditions: [
                    {
                        metric: 'attack_rate',
                        operator: ConditionOperator.GreaterThanOrEqual,
                        threshold: '100',
                        duration: '1',
                    },
                ],
            } satisfies RuleFormValues),
    })

    const { fields, append, remove } = useFieldArray({
        control: form.control,
        name: 'conditions',
    })

    const createMutation = useMutation({
        mutationFn: (payload: CreateAlertRuleRequest) =>
            alertRuleApi.createRule(payload),
        onSuccess: () => {
            queryClient.invalidateQueries({
                queryKey: queryKeys.alert.rules.lists(),
            })
            toast({
                title: t('common.success'),
                description: t('alert.createRuleSuccess', {
                    defaultValue: 'Alert rule created successfully',
                }),
                variant: 'success',
            })
            onSuccess?.()
        },
        onError: (error: Error & { response?: { data?: { message?: string } } }) => {
            toast({
                title: t('common.error'),
                description:
                    error?.response?.data?.message ??
                    t('alert.createRuleFailed', {
                        defaultValue: 'Failed to create alert rule',
                    }),
                variant: 'destructive',
            })
        },
    })

    const updateMutation = useMutation({
        mutationFn: (data: { id: string; payload: UpdateAlertRuleRequest }) =>
            alertRuleApi.updateRule(data.id, data.payload),
        onSuccess: () => {
            queryClient.invalidateQueries({
                queryKey: queryKeys.alert.rules.lists(),
            })
            toast({
                title: t('common.success'),
                description: t('alert.updateSuccess'),
                variant: 'success',
            })
            onSuccess?.()
        },
        onError: (error: Error & { response?: { data?: { message?: string } } }) => {
            toast({
                title: t('common.error'),
                description:
                    error?.response?.data?.message ??
                    t('alert.updateFailed'),
                variant: 'destructive',
            })
        },
    })

    const isSubmitting = createMutation.isPending || updateMutation.isPending
    const submitError = createMutation.error || updateMutation.error

    const buildPayload = (values: RuleFormValues): CreateAlertRuleRequest => {
        const cooldown = Number(values.cooldown) || 0
        const conditions = values.conditions.map((c) => ({
            metric: c.metric,
            operator: c.operator,
            threshold: Number(c.threshold) || 0,
            duration: Number(c.duration) || 0,
        }))
        const channels = values.channels
            .split(',')
            .map((id) => id.trim())
            .filter(Boolean)

        return {
            name: values.name,
            description: values.description || undefined,
            conditions,
            logic: values.logic,
            channels,
            template: values.template,
            cooldown,
            enabled: values.enabled,
            severity: values.severity,
        }
    }

    const onSubmit = (values: RuleFormValues) => {
        const payload = buildPayload(values)

        if (mode === 'create') {
            createMutation.mutate(payload)
        } else if (ruleId) {
            const updatePayload: UpdateAlertRuleRequest = {
                ...payload,
            }
            updateMutation.mutate({ id: ruleId, payload: updatePayload })
        }
    }

    return (
        <Form {...form}>
            <form
                onSubmit={form.handleSubmit(onSubmit)}
                className="space-y-6"
            >
                {submitError && (
                    <Alert variant="destructive">
                        <AlertCircle className="h-4 w-4" />
                        <AlertDescription>
                            {String(submitError)}
                        </AlertDescription>
                    </Alert>
                )}

                {/* 基本信息 */}
                <div className="space-y-4">
                    <h3 className="text-lg font-medium">
                        {t('alert.ruleBasicInfo', {
                            defaultValue: 'Basic Information',
                        })}
                    </h3>

                    <FormField
                        control={form.control}
                        name="name"
                        render={({ field }) => (
                            <FormItem>
                                <FormLabel>
                                    {t('alert.ruleName')} *
                                </FormLabel>
                                <FormControl>
                                    <Input
                                        placeholder={t(
                                            'alert.ruleNamePlaceholder',
                                            {
                                                defaultValue:
                                                    'High risk attack rule',
                                            },
                                        )}
                                        {...field}
                                    />
                                </FormControl>
                                <FormMessage />
                            </FormItem>
                        )}
                    />

                    <FormField
                        control={form.control}
                        name="description"
                        render={({ field }) => (
                            <FormItem>
                                <FormLabel>
                                    {t('description', {
                                        defaultValue: 'Description',
                                    })}
                                </FormLabel>
                                <FormControl>
                                    <Textarea
                                        rows={2}
                                        placeholder={t(
                                            'alert.ruleDescriptionPlaceholder',
                                            {
                                                defaultValue:
                                                    'Describe when this alert should be triggered',
                                            },
                                        )}
                                        {...field}
                                    />
                                </FormControl>
                                <FormMessage />
                            </FormItem>
                        )}
                    />

                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <FormField
                            control={form.control}
                            name="severity"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>
                                        {t('alert.severityLabel', {
                                            defaultValue: 'Severity',
                                        })}
                                    </FormLabel>
                                    <Select
                                        onValueChange={field.onChange}
                                        defaultValue={field.value}
                                    >
                                        <FormControl>
                                            <SelectTrigger>
                                                <SelectValue />
                                            </SelectTrigger>
                                        </FormControl>
                                        <SelectContent>
                                            <SelectItem
                                                value={AlertSeverity.Low}
                                            >
                                                {t('alert.severity.low')}
                                            </SelectItem>
                                            <SelectItem
                                                value={AlertSeverity.Medium}
                                            >
                                                {t('alert.severity.medium')}
                                            </SelectItem>
                                            <SelectItem
                                                value={AlertSeverity.High}
                                            >
                                                {t('alert.severity.high')}
                                            </SelectItem>
                                            <SelectItem
                                                value={AlertSeverity.Critical}
                                            >
                                                {t('alert.severity.critical')}
                                            </SelectItem>
                                        </SelectContent>
                                    </Select>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        <FormField
                            control={form.control}
                            name="logic"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>
                                        {t('operator', {
                                            defaultValue: 'Logic Operator',
                                        })}
                                    </FormLabel>
                                    <Select
                                        onValueChange={field.onChange}
                                        defaultValue={field.value}
                                    >
                                        <FormControl>
                                            <SelectTrigger>
                                                <SelectValue />
                                            </SelectTrigger>
                                        </FormControl>
                                        <SelectContent>
                                            <SelectItem value="AND">
                                                AND
                                            </SelectItem>
                                            <SelectItem value="OR">
                                                OR
                                            </SelectItem>
                                        </SelectContent>
                                    </Select>
                                    <FormDescription>
                                        {t('alert.logicDescription', {
                                            defaultValue:
                                                'Combine conditions with AND or OR.',
                                        })}
                                    </FormDescription>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        <FormField
                            control={form.control}
                            name="cooldown"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>
                                        {t('alert.cooldown', {
                                            defaultValue:
                                                'Cooldown (seconds)',
                                        })}
                                    </FormLabel>
                                    <FormControl>
                                        <Input
                                            type="number"
                                            min={0}
                                            placeholder="60"
                                            {...field}
                                        />
                                    </FormControl>
                                    <FormDescription>
                                        {t('alert.cooldownDescription', {
                                            defaultValue:
                                                'Minimum time interval between two alerts triggered by the same rule.',
                                        })}
                                    </FormDescription>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />
                    </div>

                    <FormField
                        control={form.control}
                        name="enabled"
                        render={({ field }) => (
                            <FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
                                <div className="space-y-0.5">
                                    <FormLabel className="text-base">
                                        {t('alert.enabled')}
                                    </FormLabel>
                                    <FormDescription>
                                        {t('alert.enableRuleDescription', {
                                            defaultValue:
                                                'Only enabled rules can generate alerts.',
                                        })}
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

                {/* 条件配置 */}
                <div className="space-y-4">
                    <h3 className="text-lg font-medium">
                        {t('alert.conditions', {
                            defaultValue: 'Conditions',
                        })}
                    </h3>
                    <FormDescription>
                        {t('alert.conditionsDescription', {
                            defaultValue:
                                'When any condition group is met, an alert will be triggered.',
                        })}
                    </FormDescription>

                    <div className="space-y-3">
                        {fields.map((field, index) => (
                            <div
                                key={field.id}
                                className="grid grid-cols-1 md:grid-cols-4 gap-3 items-end border rounded-md p-3"
                            >
                                <FormField
                                    control={form.control}
                                    name={`conditions.${index}.metric`}
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel>
                                                {t('metric', {
                                                    defaultValue: 'Metric',
                                                })}
                                            </FormLabel>
                                            <FormControl>
                                                <Input
                                                    placeholder="attack_rate"
                                                    {...field}
                                                />
                                            </FormControl>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />

                                <FormField
                                    control={form.control}
                                    name={`conditions.${index}.operator`}
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel>
                                                {t('operator', {
                                                    defaultValue: 'Operator',
                                                })}
                                            </FormLabel>
                                            <Select
                                                onValueChange={field.onChange}
                                                defaultValue={field.value}
                                            >
                                                <FormControl>
                                                    <SelectTrigger>
                                                        <SelectValue />
                                                    </SelectTrigger>
                                                </FormControl>
                                                <SelectContent>
                                                    <SelectItem
                                                        value={
                                                            ConditionOperator.GreaterThan
                                                        }
                                                    >
                                                        &gt;
                                                    </SelectItem>
                                                    <SelectItem
                                                        value={
                                                            ConditionOperator.GreaterThanOrEqual
                                                        }
                                                    >
                                                        &gt;=
                                                    </SelectItem>
                                                    <SelectItem
                                                        value={
                                                            ConditionOperator.LessThan
                                                        }
                                                    >
                                                        &lt;
                                                    </SelectItem>
                                                    <SelectItem
                                                        value={
                                                            ConditionOperator.LessThanOrEqual
                                                        }
                                                    >
                                                        &lt;=
                                                    </SelectItem>
                                                    <SelectItem
                                                        value={
                                                            ConditionOperator.Equal
                                                        }
                                                    >
                                                        =
                                                    </SelectItem>
                                                    <SelectItem
                                                        value={
                                                            ConditionOperator.NotEqual
                                                        }
                                                    >
                                                        !=
                                                    </SelectItem>
                                                </SelectContent>
                                            </Select>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />

                                <FormField
                                    control={form.control}
                                    name={`conditions.${index}.threshold`}
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel>
                                                {t('threshold', {
                                                    defaultValue: 'Threshold',
                                                })}
                                            </FormLabel>
                                            <FormControl>
                                                <Input
                                                    type="number"
                                                    placeholder="100"
                                                    {...field}
                                                />
                                            </FormControl>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />

                                <div className="flex gap-2 items-end">
                                    <FormField
                                        control={form.control}
                                        name={`conditions.${index}.duration`}
                                        render={({ field }) => (
                                            <FormItem className="flex-1">
                                                <FormLabel>
                                                    {t('duration', {
                                                        defaultValue:
                                                            'Duration (min)',
                                                    })}
                                                </FormLabel>
                                                <FormControl>
                                                    <Input
                                                        type="number"
                                                        placeholder="60"
                                                        {...field}
                                                    />
                                                </FormControl>
                                                <FormMessage />
                                            </FormItem>
                                        )}
                                    />
                                    <Button
                                        type="button"
                                        variant="ghost"
                                        size="icon"
                                        className="self-center mt-1"
                                        onClick={() => remove(index)}
                                        disabled={fields.length === 1}
                                    >
                                        <Trash2 className="h-4 w-4" />
                                    </Button>
                                </div>
                            </div>
                        ))}

                        <Button
                            type="button"
                            variant="outline"
                            size="sm"
                            onClick={() =>
                                append({
                                    metric: '',
                                    operator: ConditionOperator.GreaterThanOrEqual,
                                    threshold: '',
                                    duration: '',
                                })
                            }
                            className="flex items-center gap-2"
                        >
                            <PlusCircle className="h-4 w-4" />
                            {t('alert.addCondition', {
                                defaultValue: 'Add condition',
                            })}
                        </Button>
                    </div>
                </div>

                {/* 通道与模板 */}
                <div className="space-y-4">
                    <h3 className="text-lg font-medium">
                        {t('alert.notification', {
                            defaultValue: 'Notification',
                        })}
                    </h3>

                    <FormField
                        control={form.control}
                                        name="channels"
                        render={({ field }) => (
                            <FormItem>
                                <FormLabel>
                                    {t('alert.channels', {
                                        defaultValue: 'Channels',
                                    })}
                                </FormLabel>
                                <FormControl>
                                    <Input
                                        placeholder={t(
                                            'alert.channelIdsPlaceholder',
                                            {
                                                defaultValue:
                                                    'Comma-separated channel IDs, e.g. 1,2,3',
                                            },
                                        )}
                                        {...field}
                                    />
                                </FormControl>
                                <FormDescription>
                                    {t('alert.channelIdsDescription', {
                                        defaultValue:
                                            'Specify which alert channels to use by their IDs.',
                                    })}
                                </FormDescription>
                                <FormMessage />
                            </FormItem>
                        )}
                    />

                    <FormField
                        control={form.control}
                        name="template"
                        render={({ field }) => (
                            <FormItem>
                                <FormLabel>
                                    {t('alert.template', {
                                        defaultValue: 'Message template',
                                    })}
                                </FormLabel>
                                <FormControl>
                                    <Textarea
                                        rows={3}
                                        placeholder={t(
                                            'alert.templatePlaceholder',
                                            {
                                                defaultValue:
                                                    'You can use variables like {{ruleName}}, {{severity}}, {{message}}...',
                                            },
                                        )}
                                        {...field}
                                    />
                                </FormControl>
                                <FormMessage />
                            </FormItem>
                        )}
                    />
                </div>

                <div className="flex justify-end gap-2 pt-4">
                    <Button type="submit" disabled={isSubmitting}>
                        {isSubmitting && (
                            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                        )}
                        {mode === 'create'
                            ? t('common.create')
                            : t('common.save')}
                    </Button>
                </div>
            </form>
        </Form>
    )
}

