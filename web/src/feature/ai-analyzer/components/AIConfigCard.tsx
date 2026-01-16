import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Label } from "@/components/ui/label"
import { Switch } from "@/components/ui/switch"
import { Input } from "@/components/ui/input"
import { Info } from "lucide-react"
import type { AIAnalyzerConfig } from "@/types/ai-analyzer"

interface AIConfigCardProps {
    config: AIAnalyzerConfig
    onConfigChange: (config: AIAnalyzerConfig) => void
}

export function AIConfigCard({ config, onConfigChange }: AIConfigCardProps) {
    return (
        <Card>
            <CardHeader>
                <CardTitle>基础设置</CardTitle>
                <CardDescription>AI分析器的基本运行参数</CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
                <div className="flex items-center justify-between">
                    <div className="space-y-0.5">
                        <Label className="text-base">启用AI分析器</Label>
                        <div className="text-sm text-muted-foreground">
                            开启后将自动分析攻击模式并生成防护规则
                        </div>
                    </div>
                    <Switch
                        checked={config.enabled}
                        onCheckedChange={(checked) =>
                            onConfigChange({ ...config, enabled: checked })
                        }
                    />
                </div>

                <div className="space-y-2">
                    <Label htmlFor="analysis-interval">分析间隔（分钟）</Label>
                    <Input
                        id="analysis-interval"
                        type="number"
                        value={config.analysisInterval || 30}
                        onChange={(e) =>
                            onConfigChange({
                                ...config,
                                analysisInterval: parseInt(e.target.value) || 30,
                            })
                        }
                        min={5}
                        max={1440}
                        disabled={!config.enabled}
                    />
                    <div className="flex items-start gap-2 text-xs text-muted-foreground">
                        <Info className="h-3 w-3 mt-0.5 flex-shrink-0" />
                        <span>AI分析任务的执行间隔，范围: 5-1440 分钟</span>
                    </div>
                </div>
            </CardContent>
        </Card>
    )
}
