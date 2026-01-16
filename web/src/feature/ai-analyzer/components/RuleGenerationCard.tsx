import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Label } from "@/components/ui/label"
import { Switch } from "@/components/ui/switch"
import { Input } from "@/components/ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Info, Shield } from "lucide-react"
import type { AIAnalyzerConfig } from "@/types/ai-analyzer"

interface RuleGenerationCardProps {
    config: AIAnalyzerConfig
    onConfigChange: (config: AIAnalyzerConfig) => void
}

export function RuleGenerationCard({ config, onConfigChange }: RuleGenerationCardProps) {
    const ruleGeneration = config.ruleGeneration || {}

    return (
        <Card>
            <CardHeader>
                <div className="flex items-center gap-2">
                    <Shield className="h-5 w-5 text-primary" />
                    <CardTitle>规则生成配置</CardTitle>
                </div>
                <CardDescription>自动生成ModSecurity和MicroRule防护规则</CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
                <div className="flex items-center justify-between">
                    <div className="space-y-0.5">
                        <Label className="text-base">启用规则生成</Label>
                        <div className="text-sm text-muted-foreground">
                            基于检测到的攻击模式自动生成防护规则
                        </div>
                    </div>
                    <Switch
                        checked={ruleGeneration.enabled}
                        onCheckedChange={(checked) =>
                            onConfigChange({
                                ...config,
                                ruleGeneration: {
                                    ...ruleGeneration,
                                    enabled: checked,
                                },
                            })
                        }
                        disabled={!config.enabled}
                    />
                </div>

                <div className="space-y-2">
                    <Label htmlFor="confidence-threshold">置信度阈值</Label>
                    <Input
                        id="confidence-threshold"
                        type="number"
                        step="0.01"
                        value={ruleGeneration.confidenceThreshold || 0.7}
                        onChange={(e) =>
                            onConfigChange({
                                ...config,
                                ruleGeneration: {
                                    ...ruleGeneration,
                                    confidenceThreshold: parseFloat(e.target.value) || 0.7,
                                },
                            })
                        }
                        min={0}
                        max={1}
                        disabled={!config.enabled || !ruleGeneration.enabled}
                    />
                    <div className="flex items-start gap-2 text-xs text-muted-foreground">
                        <Info className="h-3 w-3 mt-0.5 flex-shrink-0" />
                        <span>规则生成的最低置信度，范围 0-1（推荐 0.7）</span>
                    </div>
                </div>

                <div className="space-y-4 border-t pt-4">
                    <div className="flex items-center justify-between">
                        <div className="space-y-0.5">
                            <Label className="text-base">自动部署规则</Label>
                            <div className="text-sm text-muted-foreground">
                                高置信度规则自动部署到生产环境
                            </div>
                        </div>
                        <Switch
                            checked={ruleGeneration.autoDeploy}
                            onCheckedChange={(checked) =>
                                onConfigChange({
                                    ...config,
                                    ruleGeneration: {
                                        ...ruleGeneration,
                                        autoDeploy: checked,
                                    },
                                })
                            }
                            disabled={!config.enabled || !ruleGeneration.enabled}
                        />
                    </div>

                    <div className="flex items-center justify-between">
                        <div className="space-y-0.5">
                            <Label className="text-base">需要人工审核</Label>
                            <div className="text-sm text-muted-foreground">
                                生成的规则需要人工审核后才能部署
                            </div>
                        </div>
                        <Switch
                            checked={ruleGeneration.reviewRequired}
                            onCheckedChange={(checked) =>
                                onConfigChange({
                                    ...config,
                                    ruleGeneration: {
                                        ...ruleGeneration,
                                        reviewRequired: checked,
                                    },
                                })
                            }
                            disabled={!config.enabled || !ruleGeneration.enabled}
                        />
                    </div>
                </div>

                <div className="space-y-2">
                    <Label htmlFor="default-action">默认动作</Label>
                    <Select
                        value={ruleGeneration.defaultAction || "block"}
                        onValueChange={(value) =>
                            onConfigChange({
                                ...config,
                                ruleGeneration: {
                                    ...ruleGeneration,
                                    defaultAction: value,
                                },
                            })
                        }
                        disabled={!config.enabled || !ruleGeneration.enabled}
                    >
                        <SelectTrigger id="default-action">
                            <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value="block">拦截（Block）</SelectItem>
                            <SelectItem value="log">仅记录（Log）</SelectItem>
                        </SelectContent>
                    </Select>
                    <div className="flex items-start gap-2 text-xs text-muted-foreground">
                        <Info className="h-3 w-3 mt-0.5 flex-shrink-0" />
                        <span>生成规则的默认处理动作</span>
                    </div>
                </div>
            </CardContent>
        </Card>
    )
}
