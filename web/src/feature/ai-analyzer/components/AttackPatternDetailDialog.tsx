import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent } from "@/components/ui/card"
import type { AttackPattern } from "@/types/ai-analyzer"
import { format } from "date-fns"

interface AttackPatternDetailDialogProps {
    open: boolean
    onOpenChange: (open: boolean) => void
    pattern: AttackPattern | null
}

export function AttackPatternDetailDialog({
    open,
    onOpenChange,
    pattern,
}: AttackPatternDetailDialogProps) {
    if (!pattern) return null

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="max-w-3xl max-h-[80vh] overflow-y-auto">
                <DialogHeader>
                    <DialogTitle className="flex items-center justify-between">
                        <span>{pattern.attack_type}</span>
                        <Badge variant={pattern.severity === "critical" || pattern.severity === "high" ? "destructive" : "default"}>
                            {pattern.severity}
                        </Badge>
                    </DialogTitle>
                </DialogHeader>

                <div className="space-y-4">
                    {/* 基本信息 */}
                    <Card>
                        <CardContent className="pt-6">
                            <h3 className="font-semibold mb-3">基本信息</h3>
                            <div className="grid grid-cols-2 gap-4 text-sm">
                                <div>
                                    <span className="text-muted-foreground">样本数量：</span>
                                    <span className="font-medium">{pattern.sample_count}</span>
                                </div>
                                <div>
                                    <span className="text-muted-foreground">检测时间：</span>
                                    <span className="font-medium">
                                        {format(new Date(pattern.detected_at), "yyyy-MM-dd HH:mm:ss")}
                                    </span>
                                </div>
                            </div>
                            {pattern.description && (
                                <div className="mt-4">
                                    <span className="text-muted-foreground text-sm">描述：</span>
                                    <p className="mt-1">{pattern.description}</p>
                                </div>
                            )}
                        </CardContent>
                    </Card>

                    {/* 统计数据 */}
                    <Card>
                        <CardContent className="pt-6">
                            <h3 className="font-semibold mb-3">统计数据</h3>
                            <div className="grid grid-cols-3 gap-4 text-sm">
                                <div>
                                    <span className="text-muted-foreground">均值：</span>
                                    <span className="font-mono font-medium">
                                        {pattern.statistical_data?.mean?.toFixed(2)}
                                    </span>
                                </div>
                                <div>
                                    <span className="text-muted-foreground">标准差：</span>
                                    <span className="font-mono font-medium">
                                        {pattern.statistical_data?.std_dev?.toFixed(2)}
                                    </span>
                                </div>
                                <div>
                                    <span className="text-muted-foreground">Z-Score：</span>
                                    <span className="font-mono font-medium">
                                        {pattern.statistical_data?.z_score?.toFixed(2)}
                                    </span>
                                </div>
                            </div>
                        </CardContent>
                    </Card>

                    {/* 特征信息 */}
                    {pattern.features && (
                        <Card>
                            <CardContent className="pt-6">
                                <h3 className="font-semibold mb-3">特征信息</h3>
                                <div className="space-y-3">
                                    {pattern.features.ip_addresses?.length > 0 && (
                                        <div>
                                            <span className="text-sm text-muted-foreground">IP地址：</span>
                                            <div className="mt-1 flex flex-wrap gap-2">
                                                {pattern.features.ip_addresses.slice(0, 10).map((ip, i) => (
                                                    <Badge key={i} variant="outline">{ip}</Badge>
                                                ))}
                                                {pattern.features.ip_addresses.length > 10 && (
                                                    <Badge variant="secondary">+{pattern.features.ip_addresses.length - 10}</Badge>
                                                )}
                                            </div>
                                        </div>
                                    )}
                                    {pattern.features.urls?.length > 0 && (
                                        <div>
                                            <span className="text-sm text-muted-foreground">URL：</span>
                                            <div className="mt-1 space-y-1">
                                                {pattern.features.urls.slice(0, 5).map((url, i) => (
                                                    <div key={i} className="text-sm font-mono bg-muted px-2 py-1 rounded">
                                                        {url}
                                                    </div>
                                                ))}
                                            </div>
                                        </div>
                                    )}
                                </div>
                            </CardContent>
                        </Card>
                    )}
                </div>
            </DialogContent>
        </Dialog>
    )
}
