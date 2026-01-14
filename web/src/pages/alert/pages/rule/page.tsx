import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { Plus, RefreshCw } from "lucide-react"
import { useQueryClient } from "@tanstack/react-query"
import { useTranslation } from "react-i18next"
import { AnimatedButton } from "@/components/ui/animation/components/animated-button"
import { AnimatedIcon } from "@/components/ui/animation/components/animated-icon"
import { AlertRule } from "@/types/alert"
import { 
    RuleTable, 
    RuleDialog, 
    DeleteRuleDialog 
} from "@/feature/alert/components"

export default function AlertRulePage() {
    const { t } = useTranslation()
    const [isAddDialogOpen, setIsAddDialogOpen] = useState(false)
    const [isEditDialogOpen, setIsEditDialogOpen] = useState(false)
    const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false)
    const [selectedRule, setSelectedRule] = useState<AlertRule | null>(null)
    const [selectedRuleId, setSelectedRuleId] = useState<string | null>(null)
    const [isRefreshAnimating, setIsRefreshAnimating] = useState(false)

    const queryClient = useQueryClient()

    // 处理添加规则
    const handleAddRule = () => {
        setIsAddDialogOpen(true)
    }

    // 处理编辑规则
    const handleEditRule = (rule: AlertRule) => {
        setSelectedRule(rule)
        setIsEditDialogOpen(true)
    }

    // 处理删除规则
    const handleDeleteRule = (id: string) => {
        setSelectedRuleId(id)
        setIsDeleteDialogOpen(true)
    }

    // 刷新规则列表
    const refreshRules = () => {
        setIsRefreshAnimating(true)
        queryClient.invalidateQueries({ queryKey: ['alertRules'] })

        setTimeout(() => {
            setIsRefreshAnimating(false)
        }, 1000)
    }

    return (
        <Card className="p-6 w-full min-h-full border-none shadow-none rounded-none">
            <div className="flex justify-between items-center mb-6 bg-zinc-50 dark:bg-muted/30 rounded-md p-4 transition-colors duration-200">
                <h2 className="text-xl font-semibold text-primary dark:text-white">{t('alert.ruleManagement')}</h2>
                <div className="flex gap-2">
                    <AnimatedButton>
                        <Button
                            variant="outline"
                            size="sm"
                            onClick={refreshRules}
                            className="flex items-center gap-2 dark:text-shadow-glow-white"
                        >
                            <AnimatedIcon animationVariant="continuous-spin" isAnimating={isRefreshAnimating} className="h-4 w-4">
                                <RefreshCw className="h-4 w-4" />
                            </AnimatedIcon>
                            {t('refresh')}
                        </Button>
                    </AnimatedButton>
                    <AnimatedButton>
                        <Button
                            size="sm"
                            onClick={handleAddRule}
                            className="flex items-center gap-1 dark:text-shadow-glow-white"
                        >
                            <Plus className="h-4 w-4" />
                            {t('alert.addRule')}
                        </Button>
                    </AnimatedButton>
                </div>
            </div>

            <RuleTable
                onEdit={handleEditRule}
                onDelete={handleDeleteRule}
            />

            {/* 添加规则对话框 */}
            <RuleDialog
                open={isAddDialogOpen}
                onOpenChange={setIsAddDialogOpen}
                mode="create"
            />

            {/* 编辑规则对话框 */}
            <RuleDialog
                open={isEditDialogOpen}
                onOpenChange={setIsEditDialogOpen}
                mode="update"
                rule={selectedRule}
            />

            {/* 删除规则确认对话框 */}
            <DeleteRuleDialog
                open={isDeleteDialogOpen}
                onOpenChange={setIsDeleteDialogOpen}
                ruleId={selectedRuleId}
            />
        </Card>
    )
}
