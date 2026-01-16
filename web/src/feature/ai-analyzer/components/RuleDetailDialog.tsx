import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent } from "@/components/ui/card"
import type { GeneratedRule } from "@/types/ai-analyzer"
import { format } from "date-fns"

interface RuleDetailDialogProps {
    open: boolean
    onOpenChange: (open: boolean) => void
    rule: GeneratedRule | null
}

export function RuleDetailDialog({ open, onOpenChange, rule }: RuleDetailDialogProps) {
    if (!rule) return null

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="max-w-4xl max-h-[80vh] overflow-y-auto">
                <DialogHeader>
                    <DialogTitle className="flex items-center justify-between">
                        <span>规则详情</span>
                        <Badge>{rule.status}</Badge>
                    </DialogTitle>
                </DialogHeader>

                <div className="space-y-4">
                    <Card>
                        <CardContent className="pt-6">
                            <h3 className="font-semibold mb-3">基本信息</h3>
                            <div className="grid grid-cols-2 gap-4 text-sm">
                                <div>
                                    <span className="text-muted-foreground">规则类型：</span>
                                    <span className="font-medium">{rule.rule_type}</span>
                                </div>
                                <div>
                                    <span className="text-muted-foreground">置信度：</span>
                                    <span className="font-medium">{(rule.confidence * 100).toFixed(1)}%</span>
                                </div>
                                <div>
                                    <span className="text-muted-foreground">创建时间：</span>
                                    <span className="font-medium">
                                        {format(new Date(rule.created_at), "yyyy-MM-dd HH:mm:ss")}
                                    </span>
                                </div>
                                {rule.deployed_at && (
                                    <div>
                                        <span className="text-muted-foreground">部署时间：</span>
                                        <span className="font-medium">
                                            {format(new Date(rule.deployed_at), "yyyy-MM-dd HH:mm:ss")}
                                        </span>
                                    </div>
                                )}
                            </div>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardContent className="pt-6">
                            <h3 className="font-semibold mb-3">规则内容</h3>
                            <pre className="bg-muted p-4 rounded-md text-sm overflow-x-auto">
                                <code>{rule.rule_content}</code>
                            </pre>
                        </CardContent>
                    </Card>

                    {rule.review_comment && (
                        <Card>
                            <CardContent className="pt-6">
                                <h3 className="font-semibold mb-3">审核意见</h3>
                                <p className="text-sm">{rule.review_comment}</p>
                                {rule.reviewed_by && (
                                    <p className="text-sm text-muted-foreground mt-2">
                                        审核人：{rule.reviewed_by}
                                    </p>
                                )}
                            </CardContent>
                        </Card>
                    )}
                </div>
            </DialogContent>
        </Dialog>
    )
}
