import { Button } from "@/components/ui/button"
import { Loader2, Settings, Save, Play, CheckCircle2 } from "lucide-react"
import {
    useAnalyzerConfig,
    useUpdateAnalyzerConfig,
    useTriggerAnalysis,
    AIConfigCard,
    PatternDetectionCard,
    RuleGenerationCard,
} from "@/feature/ai-analyzer"
import { useState, useEffect } from "react"
import type { AIAnalyzerConfig } from "@/types/ai-analyzer"
import { toast } from "@/hooks/use-toast"

export default function ConfigPage() {
    const { data, isLoading } = useAnalyzerConfig()
    const updateMutation = useUpdateAnalyzerConfig()
    const triggerMutation = useTriggerAnalysis()

    const [config, setConfig] = useState<AIAnalyzerConfig>({
        enabled: false,
        patternDetection: {},
        ruleGeneration: {},
        analysisInterval: 30,
    })

    useEffect(() => {
        if (data?.data) {
            setConfig(data.data)
        }
    }, [data])

    const handleSave = () => {
        updateMutation.mutate(config, {
            onSuccess: () => {
                toast({
                    title: "保存成功",
                    description: "AI分析器配置已更新",
                })
            },
            onError: (error: any) => {
                toast({
                    title: "保存失败",
                    description: error?.message || "配置更新失败，请重试",
                    variant: "destructive",
                })
            },
        })
    }

    const handleTrigger = () => {
        triggerMutation.mutate(undefined, {
            onSuccess: () => {
                toast({
                    title: "分析已触发",
                    description: "AI分析任务已开始执行",
                })
            },
            onError: (error: any) => {
                toast({
                    title: "触发失败",
                    description: error?.message || "触发分析失败，请重试",
                    variant: "destructive",
                })
            },
        })
    }

    if (isLoading) {
        return (
            <div className="flex items-center justify-center h-96">
                <Loader2 className="h-8 w-8 animate-spin text-primary" />
            </div>
        )
    }

    return (
        <div className="space-y-6">
            {/* 页面头部 */}
            <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
                <div>
                    <h2 className="text-3xl font-bold tracking-tight flex items-center gap-2">
                        <Settings className="h-8 w-8" />
                        AI分析器配置
                    </h2>
                    <p className="text-muted-foreground mt-2">
                        配置AI安全分析引擎的各项参数和MCP集成设置
                    </p>
                </div>
                <div className="flex gap-2">
                    <Button
                        onClick={handleTrigger}
                        disabled={triggerMutation.isPending || !config.enabled}
                        variant="outline"
                    >
                        {triggerMutation.isPending ? (
                            <>
                                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                                分析中
                            </>
                        ) : (
                            <>
                                <Play className="mr-2 h-4 w-4" />
                                立即分析
                            </>
                        )}
                    </Button>
                    <Button
                        onClick={handleSave}
                        disabled={updateMutation.isPending}
                    >
                        {updateMutation.isPending ? (
                            <>
                                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                                保存中
                            </>
                        ) : (
                            <>
                                <Save className="mr-2 h-4 w-4" />
                                保存配置
                            </>
                        )}
                    </Button>
                </div>
            </div>

            {/* 配置卡片 */}
            <div className="grid gap-6">
                <AIConfigCard config={config} onConfigChange={setConfig} />
                <PatternDetectionCard config={config} onConfigChange={setConfig} />
                <RuleGenerationCard config={config} onConfigChange={setConfig} />
            </div>

            {/* 底部保存按钮 */}
            <div className="flex justify-end items-center gap-4 pt-4 border-t">
                {updateMutation.isSuccess && (
                    <div className="flex items-center gap-2 text-sm text-green-600">
                        <CheckCircle2 className="h-4 w-4" />
                        <span>配置已保存</span>
                    </div>
                )}
                <Button onClick={handleSave} disabled={updateMutation.isPending} size="lg">
                    {updateMutation.isPending ? (
                        <>
                            <Loader2 className="mr-2 h-5 w-5 animate-spin" />
                            保存中
                        </>
                    ) : (
                        <>
                            <Save className="mr-2 h-5 w-5" />
                            保存配置
                        </>
                    )}
                </Button>
            </div>
        </div>
    )
}
