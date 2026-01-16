import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Badge } from '@/components/ui/badge'
import { Brain, TrendingUp, Settings, History } from 'lucide-react'
import { AdaptiveConfigForm } from '@/feature/adaptive-throttling/components/AdaptiveConfigForm'
import { BaselineChart } from '@/feature/adaptive-throttling/components/BaselineChart'
import { AdjustmentHistory } from '@/feature/adaptive-throttling/components/AdjustmentHistory'
import { RealTimeMonitor } from '@/feature/adaptive-throttling/components/RealTimeMonitor'

export default function AdaptiveThrottlingPage() {
    const { t } = useTranslation()
    const [activeTab, setActiveTab] = useState('config')

    return (
        <div className="flex flex-col h-full p-6 space-y-6">
            {/* 页面标题 */}
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight flex items-center gap-2">
                        <Brain className="h-8 w-8 text-primary" />
                        {t('adaptiveThrottling.title', '自适应限流')}
                    </h1>
                    <p className="text-muted-foreground mt-2">
                        {t('adaptiveThrottling.description', '基于流量模式的智能限流策略，自动学习和调整阈值')}
                    </p>
                </div>
                <Badge variant="outline" className="text-lg px-4 py-2">
                    <span className="w-2 h-2 rounded-full bg-green-500 mr-2 animate-pulse" />
                    {t('adaptiveThrottling.status.active', '运行中')}
                </Badge>
            </div>

            {/* 主内容区域 */}
            <Tabs value={activeTab} onValueChange={setActiveTab} className="flex-1 flex flex-col">
                <TabsList className="grid w-full grid-cols-4">
                    <TabsTrigger value="config" className="flex items-center gap-2">
                        <Settings className="h-4 w-4" />
                        {t('adaptiveThrottling.tabs.config', '配置管理')}
                    </TabsTrigger>
                    <TabsTrigger value="monitor" className="flex items-center gap-2">
                        <TrendingUp className="h-4 w-4" />
                        {t('adaptiveThrottling.tabs.monitor', '实时监控')}
                    </TabsTrigger>
                    <TabsTrigger value="baseline" className="flex items-center gap-2">
                        <Brain className="h-4 w-4" />
                        {t('adaptiveThrottling.tabs.baseline', '基线分析')}
                    </TabsTrigger>
                    <TabsTrigger value="history" className="flex items-center gap-2">
                        <History className="h-4 w-4" />
                        {t('adaptiveThrottling.tabs.history', '调整历史')}
                    </TabsTrigger>
                </TabsList>

                {/* 配置管理 */}
                <TabsContent value="config" className="flex-1 overflow-auto mt-6">
                    <Card>
                        <CardHeader>
                            <CardTitle>{t('adaptiveThrottling.config.title', '自适应限流配置')}</CardTitle>
                            <CardDescription>
                                {t('adaptiveThrottling.config.description', '配置学习模式、基线计算方法和自动调整策略')}
                            </CardDescription>
                        </CardHeader>
                        <CardContent>
                            <AdaptiveConfigForm />
                        </CardContent>
                    </Card>
                </TabsContent>

                {/* 实时监控 */}
                <TabsContent value="monitor" className="flex-1 overflow-auto mt-6">
                    <RealTimeMonitor />
                </TabsContent>

                {/* 基线分析 */}
                <TabsContent value="baseline" className="flex-1 overflow-auto mt-6">
                    <BaselineChart />
                </TabsContent>

                {/* 调整历史 */}
                <TabsContent value="history" className="flex-1 overflow-auto mt-6">
                    <AdjustmentHistory />
                </TabsContent>
            </Tabs>
        </div>
    )
}
