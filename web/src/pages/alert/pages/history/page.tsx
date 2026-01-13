import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { RefreshCw } from "lucide-react"
import { useQueryClient } from "@tanstack/react-query"
import { useTranslation } from "react-i18next"
import { AnimatedButton } from "@/components/ui/animation/components/animated-button"
import { AnimatedIcon } from "@/components/ui/animation/components/animated-icon"
import { AlertHistory } from "@/types/alert"
import { 
    HistoryTable, 
    HistoryDetailDialog, 
    AlertStatsCards 
} from "@/feature/alert/components"

export default function AlertHistoryPage() {
    const { t } = useTranslation()
    const [isDetailDialogOpen, setIsDetailDialogOpen] = useState(false)
    const [selectedHistory, setSelectedHistory] = useState<AlertHistory | null>(null)
    const [isRefreshAnimating, setIsRefreshAnimating] = useState(false)

    const queryClient = useQueryClient()

    // 处理查看详情
    const handleViewDetail = (history: AlertHistory) => {
        setSelectedHistory(history)
        setIsDetailDialogOpen(true)
    }

    // 刷新历史列表
    const refreshHistory = () => {
        setIsRefreshAnimating(true)
        queryClient.invalidateQueries({ queryKey: ['alertHistory'] })
        queryClient.invalidateQueries({ queryKey: ['alertStats'] })

        setTimeout(() => {
            setIsRefreshAnimating(false)
        }, 1000)
    }

    return (
        <Card className="p-6 w-full h-full border-none shadow-none rounded-none">
            <div className="flex justify-between items-center mb-6 bg-zinc-50 dark:bg-muted/30 rounded-md p-4 transition-colors duration-200">
                <h2 className="text-xl font-semibold text-primary dark:text-white">{t('alert.historyManagement')}</h2>
                <div className="flex gap-2">
                    <AnimatedButton>
                        <Button
                            variant="outline"
                            size="sm"
                            onClick={refreshHistory}
                            className="flex items-center gap-2 dark:text-shadow-glow-white"
                        >
                            <AnimatedIcon animationVariant="continuous-spin" isAnimating={isRefreshAnimating} className="h-4 w-4">
                                <RefreshCw className="h-4 w-4" />
                            </AnimatedIcon>
                            {t('refresh')}
                        </Button>
                    </AnimatedButton>
                </div>
            </div>

            {/* 统计卡片 */}
            <AlertStatsCards />

            {/* 历史记录表格 */}
            <div className="mt-6">
                <HistoryTable onViewDetail={handleViewDetail} />
            </div>

            {/* 详情对话框 */}
            <HistoryDetailDialog
                open={isDetailDialogOpen}
                onOpenChange={setIsDetailDialogOpen}
                history={selectedHistory}
            />
        </Card>
    )
}
