import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { TrendingUp, TrendingDown, Activity, AlertTriangle } from 'lucide-react'
import { adaptiveThrottlingApi } from '@/api/adaptive-throttling'
import type { AdaptiveThrottlingStats } from '@/types/adaptive-throttling'

export function RealTimeMonitor() {
    const { t } = useTranslation()
    const [stats, setStats] = useState<AdaptiveThrottlingStats | null>(null)
    const [loading, setLoading] = useState(true)

    useEffect(() => {
        fetchStats()
        const interval = setInterval(fetchStats, 5000) // 每5秒更新一次
        return () => clearInterval(interval)
    }, [])

    const fetchStats = async () => {
        try {
            const data = await adaptiveThrottlingApi.getStats()
            setStats(data)
        } catch (error) {
            console.error('Failed to fetch stats:', error)
        } finally {
            setLoading(false)
        }
    }

    if (loading || !stats) {
        return (
            <div className="grid grid-cols-3 gap-6">
                {[1, 2, 3, 4, 5, 6].map(i => (
                    <Card key={i} className="animate-pulse">
                        <CardHeader className="space-y-2">
                            <div className="h-4 bg-gray-200 rounded w-3/4" />
                            <div className="h-3 bg-gray-100 rounded w-1/2" />
                        </CardHeader>
                        <CardContent>
                            <div className="h-8 bg-gray-200 rounded w-1/3" />
                        </CardContent>
                    </Card>
                ))}
            </div>
        )
    }

    return (
        <div className="space-y-6">
            {/* 异常检测警告 */}
            {stats.anomalyDetected && (
                <Card className="border-orange-500 bg-orange-50">
                    <CardContent className="flex items-center gap-4 pt-6">
                        <AlertTriangle className="h-8 w-8 text-orange-500" />
                        <div>
                            <h3 className="font-semibold text-orange-900">
                                {t('adaptiveThrottling.monitor.anomalyDetected', '检测到异常流量')}
                            </h3>
                            <p className="text-sm text-orange-700">
                                {t('adaptiveThrottling.monitor.anomalyDesc', '系统正在自动调整限流阈值')}
                            </p>
                        </div>
                    </CardContent>
                </Card>
            )}

            {/* 当前基线值 */}
            <div>
                <h3 className="text-lg font-medium mb-4">{t('adaptiveThrottling.monitor.currentBaseline', '当前基线值')}</h3>
                <div className="grid grid-cols-3 gap-4">
                    <Card>
                        <CardHeader>
                            <CardTitle className="text-sm font-medium text-muted-foreground">
                                {t('adaptiveThrottling.monitor.visitBaseline', '访问基线')}
                            </CardTitle>
                        </CardHeader>
                        <CardContent>
                            <div className="text-2xl font-bold">{stats.currentBaseline.visit.toFixed(2)}</div>
                            <p className="text-xs text-muted-foreground mt-1">
                                {t('adaptiveThrottling.monitor.requestsPerSecond', '请求/秒')}
                            </p>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader>
                            <CardTitle className="text-sm font-medium text-muted-foreground">
                                {t('adaptiveThrottling.monitor.attackBaseline', '攻击基线')}
                            </CardTitle>
                        </CardHeader>
                        <CardContent>
                            <div className="text-2xl font-bold">{stats.currentBaseline.attack.toFixed(2)}</div>
                            <p className="text-xs text-muted-foreground mt-1">
                                {t('adaptiveThrottling.monitor.attacksPerMinute', '次/分钟')}
                            </p>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader>
                            <CardTitle className="text-sm font-medium text-muted-foreground">
                                {t('adaptiveThrottling.monitor.errorBaseline', '错误基线')}
                            </CardTitle>
                        </CardHeader>
                        <CardContent>
                            <div className="text-2xl font-bold">{stats.currentBaseline.error.toFixed(2)}</div>
                            <p className="text-xs text-muted-foreground mt-1">
                                {t('adaptiveThrottling.monitor.errorsPerMinute', '次/分钟')}
                            </p>
                        </CardContent>
                    </Card>
                </div>
            </div>

            {/* 当前阈值 */}
            <div>
                <h3 className="text-lg font-medium mb-4">{t('adaptiveThrottling.monitor.currentThreshold', '当前限流阈值')}</h3>
                <div className="grid grid-cols-3 gap-4">
                    <Card>
                        <CardHeader>
                            <CardTitle className="text-sm font-medium text-muted-foreground">
                                {t('adaptiveThrottling.monitor.visitThreshold', '访问阈值')}
                            </CardTitle>
                        </CardHeader>
                        <CardContent>
                            <div className="flex items-center justify-between">
                                <div className="text-2xl font-bold">{stats.currentThreshold.visit}</div>
                                <TrendingUp className="h-5 w-5 text-green-500" />
                            </div>
                            <p className="text-xs text-muted-foreground mt-1">
                                {t('adaptiveThrottling.monitor.perMinute', '次/分钟')}
                            </p>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader>
                            <CardTitle className="text-sm font-medium text-muted-foreground">
                                {t('adaptiveThrottling.monitor.attackThreshold', '攻击阈值')}
                            </CardTitle>
                        </CardHeader>
                        <CardContent>
                            <div className="flex items-center justify-between">
                                <div className="text-2xl font-bold">{stats.currentThreshold.attack}</div>
                                <TrendingDown className="h-5 w-5 text-orange-500" />
                            </div>
                            <p className="text-xs text-muted-foreground mt-1">
                                {t('adaptiveThrottling.monitor.perMinute', '次/分钟')}
                            </p>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader>
                            <CardTitle className="text-sm font-medium text-muted-foreground">
                                {t('adaptiveThrottling.monitor.errorThreshold', '错误阈值')}
                            </CardTitle>
                        </CardHeader>
                        <CardContent>
                            <div className="flex items-center justify-between">
                                <div className="text-2xl font-bold">{stats.currentThreshold.error}</div>
                                <Activity className="h-5 w-5 text-blue-500" />
                            </div>
                            <p className="text-xs text-muted-foreground mt-1">
                                {t('adaptiveThrottling.monitor.perMinute', '次/分钟')}
                            </p>
                        </CardContent>
                    </Card>
                </div>
            </div>

            {/* 统计信息 */}
            <div className="grid grid-cols-2 gap-4">
                <Card>
                    <CardHeader>
                        <CardTitle>{t('adaptiveThrottling.monitor.learningProgress', '学习进度')}</CardTitle>
                        <CardDescription>
                            {t('adaptiveThrottling.monitor.learningProgressDesc', '数据收集完成度')}
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <div className="space-y-2">
                            <div className="flex items-center justify-between">
                                <span className="text-2xl font-bold">{stats.learningProgress.toFixed(1)}%</span>
                                <Badge variant={stats.learningProgress === 100 ? 'default' : 'secondary'}>
                                    {stats.learningProgress === 100 
                                        ? t('adaptiveThrottling.monitor.completed', '已完成')
                                        : t('adaptiveThrottling.monitor.learning', '学习中')
                                    }
                                </Badge>
                            </div>
                            <div className="w-full bg-gray-200 rounded-full h-2">
                                <div 
                                    className="bg-primary h-2 rounded-full transition-all duration-300"
                                    style={{ width: `${stats.learningProgress}%` }}
                                />
                            </div>
                        </div>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader>
                        <CardTitle>{t('adaptiveThrottling.monitor.recentAdjustments', '近期调整次数')}</CardTitle>
                        <CardDescription>
                            {t('adaptiveThrottling.monitor.recentAdjustmentsDesc', '最近24小时内的自动调整')}
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <div className="space-y-2">
                            <div className="text-2xl font-bold">{stats.recentAdjustments}</div>
                            <p className="text-sm text-muted-foreground">
                                {t('adaptiveThrottling.monitor.lastUpdate', '最后更新')}: {new Date(stats.lastUpdateTime).toLocaleString()}
                            </p>
                        </div>
                    </CardContent>
                </Card>
            </div>
        </div>
    )
}
