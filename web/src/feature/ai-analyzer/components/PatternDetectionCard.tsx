import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Label } from "@/components/ui/label"
import { Switch } from "@/components/ui/switch"
import { Input } from "@/components/ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Info, TrendingUp } from "lucide-react"
import type { AIAnalyzerConfig } from "@/types/ai-analyzer"

interface PatternDetectionCardProps {
    config: AIAnalyzerConfig
    onConfigChange: (config: AIAnalyzerConfig) => void
}

export function PatternDetectionCard({ config, onConfigChange }: PatternDetectionCardProps) {
    const patternDetection = config.patternDetection || {}

    return (
        <Card>
            <CardHeader>
                <div className="flex items-center gap-2">
                    <TrendingUp className="h-5 w-5 text-primary" />
                    <CardTitle>模式检测配置</CardTitle>
                </div>
                <CardDescription>攻击模式识别和异常检测参数</CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
                <div className="flex items-center justify-between">
                    <div className="space-y-0.5">
                        <Label className="text-base">启用模式检测</Label>
                        <div className="text-sm text-muted-foreground">
                            使用机器学习算法检测攻击模式
                        </div>
                    </div>
                    <Switch
                        checked={patternDetection.enabled}
                        onCheckedChange={(checked) =>
                            onConfigChange({
                                ...config,
                                patternDetection: {
                                    ...patternDetection,
                                    enabled: checked,
                                },
                            })
                        }
                        disabled={!config.enabled}
                    />
                </div>

                <div className="grid gap-4 md:grid-cols-2">
                    <div className="space-y-2">
                        <Label htmlFor="min-samples">最小样本数量</Label>
                        <Input
                            id="min-samples"
                            type="number"
                            value={patternDetection.minSamples || 100}
                            onChange={(e) =>
                                onConfigChange({
                                    ...config,
                                    patternDetection: {
                                        ...patternDetection,
                                        minSamples: parseInt(e.target.value) || 100,
                                    },
                                })
                            }
                            min={10}
                            max={10000}
                            disabled={!config.enabled || !patternDetection.enabled}
                        />
                        <div className="flex items-start gap-2 text-xs text-muted-foreground">
                            <Info className="h-3 w-3 mt-0.5 flex-shrink-0" />
                            <span>触发分析所需的最小日志数量（10-10000）</span>
                        </div>
                    </div>

                    <div className="space-y-2">
                        <Label htmlFor="anomaly-threshold">异常阈值（Z-Score）</Label>
                        <Input
                            id="anomaly-threshold"
                            type="number"
                            step="0.1"
                            value={patternDetection.anomalyThreshold || 2.0}
                            onChange={(e) =>
                                onConfigChange({
                                    ...config,
                                    patternDetection: {
                                        ...patternDetection,
                                        anomalyThreshold: parseFloat(e.target.value) || 2.0,
                                    },
                                })
                            }
                            min={0.5}
                            max={10}
                            disabled={!config.enabled || !patternDetection.enabled}
                        />
                        <div className="flex items-start gap-2 text-xs text-muted-foreground">
                            <Info className="h-3 w-3 mt-0.5 flex-shrink-0" />
                            <span>统计异常检测的Z-Score阈值（0.5-10）</span>
                        </div>
                    </div>
                </div>

                <div className="grid gap-4 md:grid-cols-2">
                    <div className="space-y-2">
                        <Label htmlFor="clustering-method">聚类算法</Label>
                        <Select
                            value={patternDetection.clusteringMethod || "kmeans"}
                            onValueChange={(value) =>
                                onConfigChange({
                                    ...config,
                                    patternDetection: {
                                        ...patternDetection,
                                        clusteringMethod: value,
                                    },
                                })
                            }
                            disabled={!config.enabled || !patternDetection.enabled}
                        >
                            <SelectTrigger id="clustering-method">
                                <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="kmeans">K-Means</SelectItem>
                                <SelectItem value="dbscan">DBSCAN</SelectItem>
                            </SelectContent>
                        </Select>
                        <div className="flex items-start gap-2 text-xs text-muted-foreground">
                            <Info className="h-3 w-3 mt-0.5 flex-shrink-0" />
                            <span>用于攻击模式聚类的算法</span>
                        </div>
                    </div>

                    <div className="space-y-2">
                        <Label htmlFor="time-window">时间窗口（小时）</Label>
                        <Input
                            id="time-window"
                            type="number"
                            value={patternDetection.timeWindow || 24}
                            onChange={(e) =>
                                onConfigChange({
                                    ...config,
                                    patternDetection: {
                                        ...patternDetection,
                                        timeWindow: parseInt(e.target.value) || 24,
                                    },
                                })
                            }
                            min={1}
                            max={168}
                            disabled={!config.enabled || !patternDetection.enabled}
                        />
                        <div className="flex items-start gap-2 text-xs text-muted-foreground">
                            <Info className="h-3 w-3 mt-0.5 flex-shrink-0" />
                            <span>分析的时间窗口范围（1-168小时）</span>
                        </div>
                    </div>
                </div>
            </CardContent>
        </Card>
    )
}
