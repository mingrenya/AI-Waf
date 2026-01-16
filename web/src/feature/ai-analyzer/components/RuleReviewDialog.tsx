import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Textarea } from "@/components/ui/textarea"
import { Label } from "@/components/ui/label"
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group"
import { useState } from "react"
import { useReviewRule } from "../hooks"
import type { GeneratedRule } from "@/types/ai-analyzer"

interface RuleReviewDialogProps {
    open: boolean
    onOpenChange: (open: boolean) => void
    rule: GeneratedRule | null
}

export function RuleReviewDialog({ open, onOpenChange, rule }: RuleReviewDialogProps) {
    const [action, setAction] = useState<"approve" | "reject">("approve")
    const [comment, setComment] = useState("")
    const reviewMutation = useReviewRule()

    const handleSubmit = () => {
        if (!rule) return
        reviewMutation.mutate({
            ruleId: rule.id,
            action,
            comment,
        })
        onOpenChange(false)
        setComment("")
        setAction("approve")
    }

    if (!rule) return null

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent>
                <DialogHeader>
                    <DialogTitle>审核规则</DialogTitle>
                </DialogHeader>

                <div className="space-y-4">
                    <div>
                        <Label>审核决定</Label>
                        <RadioGroup value={action} onValueChange={(v: "approve" | "reject") => setAction(v)} className="mt-2">
                            <div className="flex items-center space-x-2">
                                <RadioGroupItem value="approve" id="approve" />
                                <Label htmlFor="approve" className="cursor-pointer">批准部署</Label>
                            </div>
                            <div className="flex items-center space-x-2">
                                <RadioGroupItem value="reject" id="reject" />
                                <Label htmlFor="reject" className="cursor-pointer">拒绝规则</Label>
                            </div>
                        </RadioGroup>
                    </div>

                    <div>
                        <Label htmlFor="comment">审核意见（可选）</Label>
                        <Textarea
                            id="comment"
                            value={comment}
                            onChange={(e) => setComment(e.target.value)}
                            placeholder="请输入审核意见..."
                            className="mt-2"
                            rows={4}
                        />
                    </div>
                </div>

                <DialogFooter>
                    <Button variant="outline" onClick={() => onOpenChange(false)}>
                        取消
                    </Button>
                    <Button onClick={handleSubmit} disabled={reviewMutation.isPending}>
                        {reviewMutation.isPending ? "提交中..." : "提交"}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    )
}
