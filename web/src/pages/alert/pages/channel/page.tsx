import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { Plus, RefreshCw } from "lucide-react"
import { useQueryClient } from "@tanstack/react-query"
import { useTranslation } from "react-i18next"
import { AnimatedButton } from "@/components/ui/animation/components/animated-button"
import { AnimatedIcon } from "@/components/ui/animation/components/animated-icon"
import { AlertChannel } from "@/types/alert"
import { 
    ChannelTable, 
    ChannelDialog, 
    DeleteChannelDialog, 
    TestChannelDialog 
} from "@/feature/alert/components"

export default function AlertChannelPage() {
    const { t } = useTranslation()
    const [isAddDialogOpen, setIsAddDialogOpen] = useState(false)
    const [isEditDialogOpen, setIsEditDialogOpen] = useState(false)
    const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false)
    const [isTestDialogOpen, setIsTestDialogOpen] = useState(false)
    const [selectedChannel, setSelectedChannel] = useState<AlertChannel | null>(null)
    const [selectedChannelId, setSelectedChannelId] = useState<string | null>(null)
    const [isRefreshAnimating, setIsRefreshAnimating] = useState(false)

    const queryClient = useQueryClient()

    // 处理添加通道
    const handleAddChannel = () => {
        setIsAddDialogOpen(true)
    }

    // 处理编辑通道
    const handleEditChannel = (channel: AlertChannel) => {
        setSelectedChannel(channel)
        setIsEditDialogOpen(true)
    }

    // 处理删除通道
    const handleDeleteChannel = (id: string) => {
        setSelectedChannelId(id)
        setIsDeleteDialogOpen(true)
    }

    // 处理测试通道
    const handleTestChannel = (channel: AlertChannel) => {
        setSelectedChannel(channel)
        setIsTestDialogOpen(true)
    }

    // 刷新通道列表
    const refreshChannels = () => {
        setIsRefreshAnimating(true)
        queryClient.invalidateQueries({ queryKey: ['alertChannels'] })

        setTimeout(() => {
            setIsRefreshAnimating(false)
        }, 1000)
    }

    return (
        <Card className="p-6 w-full min-h-full border-none shadow-none rounded-none">
            <div className="flex justify-between items-center mb-6 bg-zinc-50 dark:bg-muted/30 rounded-md p-4 transition-colors duration-200">
                <h2 className="text-xl font-semibold text-primary dark:text-white">{t('alert.channelManagement')}</h2>
                <div className="flex gap-2">
                    <AnimatedButton>
                        <Button
                            variant="outline"
                            size="sm"
                            onClick={refreshChannels}
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
                            onClick={handleAddChannel}
                            className="flex items-center gap-1 dark:text-shadow-glow-white"
                        >
                            <Plus className="h-4 w-4" />
                            {t('alert.addChannel')}
                        </Button>
                    </AnimatedButton>
                </div>
            </div>

            <ChannelTable
                onEdit={handleEditChannel}
                onDelete={handleDeleteChannel}
                onTest={handleTestChannel}
            />

            {/* 添加通道对话框 */}
            <ChannelDialog
                open={isAddDialogOpen}
                onOpenChange={setIsAddDialogOpen}
                mode="create"
            />

            {/* 编辑通道对话框 */}
            <ChannelDialog
                open={isEditDialogOpen}
                onOpenChange={setIsEditDialogOpen}
                mode="update"
                channel={selectedChannel}
            />

            {/* 删除通道确认对话框 */}
            <DeleteChannelDialog
                open={isDeleteDialogOpen}
                onOpenChange={setIsDeleteDialogOpen}
                channelId={selectedChannelId}
            />

            {/* 测试通道对话框 */}
            <TestChannelDialog
                open={isTestDialogOpen}
                onOpenChange={setIsTestDialogOpen}
                channel={selectedChannel}
            />
        </Card>
    )
}
