import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Badge } from '@/components/ui/badge'
import { adaptiveThrottlingApi } from '@/api/adaptive-throttling'
import type { BaselineValue } from '@/types/adaptive-throttling'

export function BaselineChart() {
    const { t } = useTranslation()
    const [baselines, setBaselines] = useState<BaselineValue[]>([])
    const [loading, setLoading] = useState(true)
    const [activeType, setActiveType] = useState<'visit' | 'attack' | 'error'>('visit')

    useEffect(() => {
        fetchBaselines()
    }, [])

    const fetchBaselines = async () => {
        try {
            const response = await adaptiveThrottlingApi.getBaselines({})
            setBaselines(response.results)
        } catch (error) {
            console.error('Failed to fetch baselines:', error)
        } finally {
            setLoading(false)
        }
    }

    const getBaselineByType = (type: string) => {
        return baselines.find(b => b.type === type)
    }

    const renderBaselineCard = (type: 'visit' | 'attack' | 'error', title: string, description: string) => {
        const baseline = getBaselineByType(type)

        if (!baseline) {
            return (
                <Card>
                    <CardHeader>
                        <CardTitle>{title}</CardTitle>
                        <CardDescription>{description}</CardDescription>
                    </CardHeader>
                    <CardContent>
                        <p className="text-muted-foreground">
                            {t('adaptiveThrottling.baseline.noData', '暂无数据')}
                        </p>
                    </CardContent>
                </Card>
            )
        }

        return (
            <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <div>
                        <CardTitle>{title}</CardTitle>
                        <CardDescription>{description}</CardDescription>
                    </div>
                    <Badge variant={baseline.confidenceLevel >= 0.95 ? 'default' : 'secondary'}>
                        {t('adaptiveThrottling.baseline.confidence', '置信度')}: {(baseline.confidenceLevel * 100).toFixed(1)}%
                    </Badge>
                </CardHeader>
                <CardContent>
                    <div className="space-y-4">
                        <div className="flex items-baseline gap-2">
                            <span className="text-4xl font-bold">{baseline.value.toFixed(2)}</span>
                            <span className="text-sm text-muted-foreground">
                                {type === 'visit' 
                                    ? t('adaptiveThrottling.baseline.requestsPerSecond', '请求/秒')
                                    : t('adaptiveThrottling.baseline.perMinute', '次/分钟')
                                }
                            </span>
                        </div>

                        <div className="grid grid-cols-2 gap-4 pt-4 border-t">
                            <div>
                                <p className="text-sm text-muted-foreground">
                                    {t('adaptiveThrottling.baseline.sampleSize', '样本数量')}
                                </p>
                                <p className="text-lg font-semibold">{baseline.sampleSize.toLocaleString()}</p>
                            </div>
                            <div>
                                <p className="text-sm text-muted-foreground">
                                    {t('adaptiveThrottling.baseline.calculatedAt', '计算时间')}
                                </p>
                                <p className="text-lg font-semibold">
                                    {new Date(baseline.calculatedAt).toLocaleString()}
                                </p>
                            </div>
                        </div>

                        <div className="pt-2">
                            <p className="text-xs text-muted-foreground">
                                {t('adaptiveThrottling.baseline.lastUpdate', '最后更新')}: {new Date(baseline.updatedAt).toLocaleString()}
                            </p>
                        </div>
                    </div>
                </CardContent>
            </Card>
        )
    }

    if (loading) {
        return (
            <div className="grid gap-6">
                {[1, 2, 3].map(i => (
                    <Card key={i} className="animate-pulse">
                        <CardHeader className="space-y-2">
                            <div className="h-6 bg-gray-200 rounded w-1/4" />
                            <div className="h-4 bg-gray-100 rounded w-1/2" />
                        </CardHeader>
                        <CardContent>
                            <div className="h-20 bg-gray-200 rounded" />
                        </CardContent>
                    </Card>
                ))}
            </div>
        )
    }

    return (
        <div className="space-y-6">
            <Tabs value={activeType} onValueChange={(v) => setActiveType(v as any)}>
                <TabsList className="grid w-full grid-cols-3">
                    <TabsTrigger value="visit">
                        {t('adaptiveThrottling.baseline.visit', '访问基线')}
                    </TabsTrigger>
                    <TabsTrigger value="attack">
                        {t('adaptiveThrottling.baseline.attack', '攻击基线')}
                    </TabsTrigger>
                    <TabsTrigger value="error">
                        {t('adaptiveThrottling.baseline.error', '错误基线')}
                    </TabsTrigger>
                </TabsList>

                <TabsContent value="visit" className="mt-6">
                    {renderBaselineCard('visit', 
                        t('adaptiveThrottling.baseline.visitTitle', '访问流量基线'),
                        t('adaptiveThrottling.baseline.visitDesc', '基于历史访问流量计算的基线值')
                    )}
                </TabsContent>

                <TabsContent value="attack" className="mt-6">
                    {renderBaselineCard('attack',
                        t('adaptiveThrottling.baseline.attackTitle', '攻击流量基线'),
                        t('adaptiveThrottling.baseline.attackDesc', '基于历史攻击流量计算的基线值')
                    )}
                </TabsContent>

                <TabsContent value="error" className="mt-6">
                    {renderBaselineCard('error',
                        t('adaptiveThrottling.baseline.errorTitle', '错误流量基线'),
                        t('adaptiveThrottling.baseline.errorDesc', '基于历史错误流量计算的基线值')
                    )}
                </TabsContent>
            </Tabs>

            {/* 基线说明 */}
            <Card className="bg-blue-50 border-blue-200">
                <CardHeader>
                    <CardTitle className="text-blue-900">
                        {t('adaptiveThrottling.baseline.about', '关于基线')}
                    </CardTitle>
                </CardHeader>
                <CardContent className="text-sm text-blue-800 space-y-2">
                    <p>
                        {t('adaptiveThrottling.baseline.aboutDesc1', '基线值是通过分析历史流量数据计算得出的正常流量水平。')}
                    </p>
                    <p>
                        {t('adaptiveThrottling.baseline.aboutDesc2', '系统会定期更新基线值，并使用它来检测异常流量和自动调整限流阈值。')}
                    </p>
                    <p>
                        {t('adaptiveThrottling.baseline.aboutDesc3', '置信度越高，表示基线值越可靠。建议在置信度达到95%以上时启用自动调整。')}
                    </p>
                </CardContent>
            </Card>
        </div>
    )
}
