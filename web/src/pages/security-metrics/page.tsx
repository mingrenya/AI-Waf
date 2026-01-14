import { useState, useEffect } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { securityMetricsApi } from '@/api/security-metrics'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Skeleton } from '@/components/ui/skeleton'
import { Badge } from '@/components/ui/badge'
import { Progress } from '@/components/ui/progress'
import { Shield, Activity, AlertTriangle, BarChart3, TrendingUp, Globe, Clock } from 'lucide-react'
import type { SecurityMetricsResponse } from '@/types/security-metrics'

/**
 * 综合安全指标仪表板页面
 * 提供多维度的安全指标可视化展示
 */
const SecurityMetricsPage = () => {
    const { t } = useTranslation()
    const [timeRange, setTimeRange] = useState<'24h' | '7d' | '30d'>('24h')
    const [renderKey, setRenderKey] = useState(Date.now())

    // 获取综合安全指标 - 使用简化的缓存策略
    const { data, isLoading, error } = useQuery({
        queryKey: ['security-metrics', timeRange],
        queryFn: () => securityMetricsApi.getSecurityMetrics({ timeRange }),
        staleTime: 30000,
        gcTime: 60000, // 缓存时间
        refetchOnWindowFocus: false, // 禁用窗口聚焦时自动刷新
    })

    // 数据更新时生成新的渲染key，强制重新挂载
    useEffect(() => {
        if (data) {
            setRenderKey(Date.now())
        }
    }, [data])

    return (
        <div className="w-full min-h-full p-6 space-y-6">
            {/* 标题和时间范围选择器 */}
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold flex items-center gap-2">
                        <Shield className="h-8 w-8 text-primary" />
                        {t('securityMetrics.title', '综合安全指标')}
                    </h1>
                    <p className="text-muted-foreground mt-2">
                        {t('securityMetrics.description', '全面的安全态势分析和威胁情报展示')}
                    </p>
                </div>
                <Select value={timeRange} onValueChange={(value) => setTimeRange(value as '24h' | '7d' | '30d')}>
                    <SelectTrigger className="w-[180px]">
                        <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                        <SelectItem value="24h">{t('common.last24Hours', '最近24小时')}</SelectItem>
                        <SelectItem value="7d">{t('common.last7Days', '最近7天')}</SelectItem>
                        <SelectItem value="30d">{t('common.last30Days', '最近30天')}</SelectItem>
                    </SelectContent>
                </Select>
            </div>

            {error && (
                <Card className="p-6 border-destructive">
                    <div className="flex items-center gap-2 text-destructive">
                        <AlertTriangle className="h-5 w-5" />
                        <span>{t('common.errorLoadingData', '加载数据失败')}</span>
                    </div>
                </Card>
            )}

            {isLoading ? (
                <LoadingSkeleton />
            ) : data ? (
                <div key={renderKey} className="space-y-6">{/* 使用时间戳key强制重新挂载 */}
                    {/* 第一行: 核心指标卡片 */}
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                        <StatCard
                            icon={<Activity className="h-5 w-5" />}
                            title={t('securityMetrics.totalRequests', '总请求数')}
                            value={data.overview.totalRequests.toLocaleString()}
                            subValue={`QPS: ${data.overview.maxQPS}`}
                        />
                        <StatCard
                            icon={<Shield className="h-5 w-5" />}
                            title={t('securityMetrics.blockCount', '拦截数量')}
                            value={data.overview.blockCount.toLocaleString()}
                            subValue={`${((data.overview.blockCount / (data.overview.totalRequests || 1)) * 100).toFixed(2)}%`}
                            variant="warning"
                        />
                        <StatCard
                            icon={<AlertTriangle className="h-5 w-5" />}
                            title={t('securityMetrics.attackIPCount', '攻击IP数')}
                            value={data.overview.attackIPCount.toLocaleString()}
                            variant="danger"
                        />
                        <StatCard
                            icon={<BarChart3 className="h-5 w-5" />}
                            title={t('securityMetrics.error4xx', '4xx错误')}
                            value={data.overview.error4xx.toLocaleString()}
                            subValue={`${data.overview.error4xxRate.toFixed(2)}%`}
                        />
                    </div>

                    {/* 第二行: 规则引擎和威胁等级 */}
                    <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                        <RuleEngineCard key="rule-engine" data={data.ruleEngine} />
                        <ThreatLevelCard key="threat-level" data={data.threatLevel} />
                    </div>

                    {/* 第三行: Top触发规则 */}
                    <Card key="top-triggered-rules">
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2">
                                <TrendingUp className="h-5 w-5" />
                                {t('securityMetrics.topTriggeredRules', 'Top触发规则')}
                            </CardTitle>
                            <CardDescription>
                                {t('securityMetrics.topTriggeredRulesDesc', '触发次数最多的安全规则')}
                            </CardDescription>
                        </CardHeader>
                        <CardContent>
                            {data.topTriggeredRules && data.topTriggeredRules.length > 0 ? (
                                <div className="space-y-4">
                                    {data.topTriggeredRules.map((rule, index) => (
                                        <div key={`rule-${renderKey}-${index}`} className="flex items-center gap-4">
                                            <div className="flex-1">
                                                <div className="flex justify-between items-center mb-2">
                                                    <span className="font-medium">{rule.ruleName}</span>
                                                    <span className="text-sm text-muted-foreground">
                                                        {rule.count} ({rule.percentage.toFixed(1)}%)
                                                    </span>
                                                </div>
                                                <Progress value={rule.percentage} max={100} className="h-2" />
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            ) : (
                                <p className="text-sm text-muted-foreground text-center py-4">
                                    {t('common.noData', '暂无数据')}
                                </p>
                            )}
                        </CardContent>
                    </Card>

                    {/* 第四行: 严重等级分布和攻击类型 */}
                    <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                        <Card key="severity-distribution">
                            <CardHeader>
                                <CardTitle>{t('securityMetrics.severityDistribution', '严重等级分布')}</CardTitle>
                            </CardHeader>
                            <CardContent>
                                {data.severityDistribution && data.severityDistribution.length > 0 ? (
                                    <div className="space-y-3">
                                        {data.severityDistribution.map((severity, index) => (
                                            <div key={`severity-${renderKey}-${index}`} className="flex items-center gap-4">
                                                <Badge variant={getSeverityVariant(severity.level)}>
                                                    {severity.levelName}
                                                </Badge>
                                                <div className="flex-1">
                                                    <Progress value={severity.percentage} max={100} className="h-2" />
                                                </div>
                                                <span className="text-sm text-muted-foreground w-16 text-right">
                                                    {severity.count}
                                                </span>
                                            </div>
                                        ))}
                                    </div>
                                ) : (
                                    <p className="text-sm text-muted-foreground text-center py-4">
                                        {t('common.noData', '暂无数据')}
                                    </p>
                                )}
                            </CardContent>
                        </Card>

                        <Card key="attack-type-distribution">
                            <CardHeader>
                                <CardTitle>{t('securityMetrics.attackTypeDistribution', '攻击类型分布')}</CardTitle>
                            </CardHeader>
                            <CardContent>
                                {data.attackTypeDistribution && data.attackTypeDistribution.length > 0 ? (
                                    <div className="space-y-3">
                                        {data.attackTypeDistribution.map((type, index) => (
                                            <div key={`attack-${renderKey}-${index}`} className="flex items-center justify-between">
                                                <span className="text-sm">{type.category}</span>
                                                <div className="flex items-center gap-2">
                                                    <Progress value={type.percentage} max={100} className="h-2 w-24" />
                                                    <span className="text-sm text-muted-foreground w-16 text-right">
                                                        {type.count}
                                                    </span>
                                                </div>
                                            </div>
                                        ))}
                                    </div>
                                ) : (
                                    <p className="text-sm text-muted-foreground text-center py-4">
                                        {t('common.noData', '暂无数据')}
                                    </p>
                                )}
                            </CardContent>
                        </Card>
                    </div>

                    {/* 第五行: 攻击来源 */}
                    <Card key="top-attack-sources">
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2">
                                <Globe className="h-5 w-5" />
                                {t('securityMetrics.topAttackSources', 'Top攻击来源')}
                            </CardTitle>
                        </CardHeader>
                        <CardContent>
                            {data.topAttackSources && data.topAttackSources.length > 0 ? (
                                <div className="space-y-3">
                                    {data.topAttackSources.map((source, index) => (
                                        <div key={`source-${renderKey}-${index}`} className="flex items-center justify-between">
                                            <div className="flex items-center gap-2">
                                                <span className="font-mono text-sm">{source.countryCode}</span>
                                                <span>{source.country}</span>
                                                {source.city && <span className="text-muted-foreground">/ {source.city}</span>}
                                            </div>
                                            <div className="flex items-center gap-2">
                                                <Progress value={source.percentage} max={100} className="h-2 w-24" />
                                                <span className="text-sm text-muted-foreground w-16 text-right">
                                                    {source.count}
                                                </span>
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            ) : (
                                <p className="text-sm text-muted-foreground text-center py-4">
                                    {t('common.noData', '暂无数据')}
                                </p>
                            )}
                        </CardContent>
                    </Card>

                    {/* 第六行: 封禁IP和响应时间 */}
                    <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                        <BlockedIPCard key="blocked-ip" data={data.blockedIPMetrics} />
                        <ResponseTimeCard key="response-time" data={data.responseTime} />
                    </div>
                </div>
            ) : null}
        </div>
    )
}

// 获取严重等级的Badge变体
const getSeverityVariant = (level: number): 'default' | 'secondary' | 'destructive' | 'outline' => {
    if (level >= 4) return 'destructive'
    if (level >= 2) return 'default'
    return 'secondary'
}

// 规则引擎卡片
const RuleEngineCard = ({ data }: { data: SecurityMetricsResponse['ruleEngine'] }) => {
    const { t } = useTranslation()
    return (
        <Card>
            <CardHeader>
                <CardTitle>{t('securityMetrics.ruleEngine', '规则引擎')}</CardTitle>
                <CardDescription>{t('securityMetrics.ruleEngineDesc', '规则引擎性能和配置状态')}</CardDescription>
            </CardHeader>
            <CardContent>
                <div className="grid grid-cols-2 gap-4">
                    <div>
                        <p className="text-sm text-muted-foreground">{t('securityMetrics.totalRules', '总规则数')}</p>
                        <p className="text-2xl font-bold">{data.totalRules}</p>
                    </div>
                    <div>
                        <p className="text-sm text-muted-foreground">{t('securityMetrics.enabledRules', '已启用')}</p>
                        <p className="text-2xl font-bold text-green-600">{data.enabledRules}</p>
                    </div>
                    <div>
                        <p className="text-sm text-muted-foreground">{t('securityMetrics.blacklistRules', '黑名单')}</p>
                        <p className="text-2xl font-bold">{data.blacklistRules}</p>
                    </div>
                    <div>
                        <p className="text-sm text-muted-foreground">{t('securityMetrics.whitelistRules', '白名单')}</p>
                        <p className="text-2xl font-bold">{data.whitelistRules}</p>
                    </div>
                </div>
                <div className="mt-4 pt-4 border-t">
                    <div className="flex justify-between items-center">
                        <span className="text-sm">{t('securityMetrics.ruleEfficiency', '规则效率')}</span>
                        <span className="text-sm font-medium">{data.ruleEfficiency.toFixed(1)}%</span>
                    </div>
                    <Progress value={data.ruleEfficiency} max={100} className="h-2 mt-2" />
                </div>
            </CardContent>
        </Card>
    )
}

// 威胁等级卡片
const ThreatLevelCard = ({ data }: { data: SecurityMetricsResponse['threatLevel'] }) => {
    const { t } = useTranslation()
    const total = data.critical + data.high + data.medium + data.low

    return (
        <Card>
            <CardHeader>
                <CardTitle>{t('securityMetrics.threatLevel', '威胁等级')}</CardTitle>
                <CardDescription>{t('securityMetrics.threatLevelDesc', '当前威胁等级分布（最近1小时）')}</CardDescription>
            </CardHeader>
            <CardContent>
                <div className="grid grid-cols-2 gap-4">
                    <div className="p-4 border border-red-500/20 bg-red-500/5 rounded-lg">
                        <p className="text-sm text-muted-foreground">{t('securityMetrics.critical', '严重')}</p>
                        <p className="text-3xl font-bold text-red-600">{data.critical}</p>
                    </div>
                    <div className="p-4 border border-orange-500/20 bg-orange-500/5 rounded-lg">
                        <p className="text-sm text-muted-foreground">{t('securityMetrics.high', '高')}</p>
                        <p className="text-3xl font-bold text-orange-600">{data.high}</p>
                    </div>
                    <div className="p-4 border border-yellow-500/20 bg-yellow-500/5 rounded-lg">
                        <p className="text-sm text-muted-foreground">{t('securityMetrics.medium', '中')}</p>
                        <p className="text-3xl font-bold text-yellow-600">{data.medium}</p>
                    </div>
                    <div className="p-4 border border-blue-500/20 bg-blue-500/5 rounded-lg">
                        <p className="text-sm text-muted-foreground">{t('securityMetrics.low', '低')}</p>
                        <p className="text-3xl font-bold text-blue-600">{data.low}</p>
                    </div>
                </div>
                <div className="mt-4 pt-4 border-t">
                    <p className="text-sm text-center text-muted-foreground">
                        {t('securityMetrics.totalThreats', '总威胁数')}: <span className="font-bold">{total}</span>
                    </p>
                </div>
            </CardContent>
        </Card>
    )
}

// 封禁IP卡片
const BlockedIPCard = ({ data }: { data: SecurityMetricsResponse['blockedIPMetrics'] }) => {
    const { t } = useTranslation()
    return (
        <Card>
            <CardHeader>
                <CardTitle>{t('securityMetrics.blockedIPMetrics', '封禁IP统计')}</CardTitle>
            </CardHeader>
            <CardContent>
                <div className="space-y-3">
                    <div className="flex justify-between items-center">
                        <span className="text-sm">{t('securityMetrics.totalBlocked', '总封禁数')}</span>
                        <span className="font-bold">{data.totalBlocked}</span>
                    </div>
                    <div className="flex justify-between items-center">
                        <span className="text-sm">{t('securityMetrics.activeBlocked', '当前活跃')}</span>
                        <span className="font-bold text-green-600">{data.activeBlocked}</span>
                    </div>
                    <div className="flex justify-between items-center pt-3 border-t">
                        <span className="text-sm">{t('securityMetrics.highFrequencyVisit', '高频访问')}</span>
                        <span>{data.highFrequencyVisit}</span>
                    </div>
                    <div className="flex justify-between items-center">
                        <span className="text-sm">{t('securityMetrics.highFrequencyAttack', '高频攻击')}</span>
                        <span>{data.highFrequencyAttack}</span>
                    </div>
                    <div className="flex justify-between items-center">
                        <span className="text-sm">{t('securityMetrics.highFrequencyError', '高频错误')}</span>
                        <span>{data.highFrequencyError}</span>
                    </div>
                </div>
            </CardContent>
        </Card>
    )
}

// 响应时间卡片
const ResponseTimeCard = ({ data }: { data: SecurityMetricsResponse['responseTime'] }) => {
    const { t } = useTranslation()
    return (
        <Card>
            <CardHeader>
                <CardTitle className="flex items-center gap-2">
                    <Clock className="h-5 w-5" />
                    {t('securityMetrics.responseTime', '响应时间')}
                </CardTitle>
            </CardHeader>
            <CardContent>
                <div className="space-y-3">
                    <div className="flex justify-between items-center">
                        <span className="text-sm">{t('securityMetrics.avgResponseTime', '平均响应时间')}</span>
                        <span className="font-bold">{data.avgResponseTime.toFixed(2)} ms</span>
                    </div>
                    <div className="flex justify-between items-center">
                        <span className="text-sm">P50</span>
                        <span>{data.p50ResponseTime.toFixed(2)} ms</span>
                    </div>
                    <div className="flex justify-between items-center">
                        <span className="text-sm">P95</span>
                        <span>{data.p95ResponseTime.toFixed(2)} ms</span>
                    </div>
                    <div className="flex justify-between items-center">
                        <span className="text-sm">P99</span>
                        <span>{data.p99ResponseTime.toFixed(2)} ms</span>
                    </div>
                    <div className="flex justify-between items-center pt-3 border-t">
                        <span className="text-sm">{t('securityMetrics.maxResponseTime', '最大响应时间')}</span>
                        <span className="font-bold text-red-600">{data.maxResponseTime.toFixed(2)} ms</span>
                    </div>
                </div>
            </CardContent>
        </Card>
    )
}

// 统计卡片组件
interface StatCardProps {
    icon: React.ReactNode
    title: string
    value: string
    subValue?: string
    variant?: 'default' | 'warning' | 'danger'
}

const StatCard = ({ icon, title, value, subValue, variant = 'default' }: StatCardProps) => {
    const variantStyles = {
        default: 'border-blue-500/20 bg-blue-500/5',
        warning: 'border-yellow-500/20 bg-yellow-500/5',
        danger: 'border-red-500/20 bg-red-500/5',
    }

    return (
        <Card className={`p-6 ${variantStyles[variant]}`}>
            <div className="flex items-start justify-between">
                <div className="space-y-2">
                    <p className="text-sm text-muted-foreground">{title}</p>
                    <p className="text-3xl font-bold">{value}</p>
                    {subValue && <p className="text-sm text-muted-foreground">{subValue}</p>}
                </div>
                <div className={`p-3 rounded-lg ${variant === 'default' ? 'bg-blue-500/10' : variant === 'warning' ? 'bg-yellow-500/10' : 'bg-red-500/10'}`}>
                    {icon}
                </div>
            </div>
        </Card>
    )
}

// 加载骨架屏
const LoadingSkeleton = () => {
    return (
        <div className="space-y-6">
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                {[...Array(4)].map((_, i) => (
                    <Card key={`stat-skeleton-${i}`} className="p-6">
                        <Skeleton className="h-24 w-full" />
                    </Card>
                ))}
            </div>
            {[...Array(3)].map((_, i) => (
                <Card key={`card-skeleton-${i}`} className="p-6">
                    <Skeleton className="h-64 w-full" />
                </Card>
            ))}
        </div>
    )
}

export default SecurityMetricsPage
