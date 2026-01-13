import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle
} from '@/components/ui/dialog'
import { ChannelForm } from './ChannelForm'
import { AlertChannel } from '@/types/alert'
import { AnimatePresence, motion } from "motion/react"
import {
    dialogEnterExitAnimation,
    dialogContentAnimation,
    dialogHeaderAnimation,
    dialogContentItemAnimation
} from '@/components/ui/animation/dialog-animation'
import { useTranslation } from 'react-i18next'

interface ChannelDialogProps {
    open: boolean
    onOpenChange: (open: boolean) => void
    mode: 'create' | 'update'
    channel?: AlertChannel | null
}

export function ChannelDialog({
    open,
    onOpenChange,
    mode = 'create',
    channel = null
}: ChannelDialogProps) {
    const { t } = useTranslation()

    const title = mode === 'create' ? t('alert.dialog.createChannelTitle') : t('alert.dialog.updateChannelTitle')
    const description = mode === 'create'
        ? t('alert.dialog.createChannelDescription')
        : t('alert.dialog.updateChannelDescription')

    const defaultValues = mode === 'update' && channel
        ? {
            name: channel.name,
            type: channel.type,
            config: channel.config,
            enabled: channel.enabled,
        }
        : undefined

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <AnimatePresence mode="wait">
                {open && (
                    <motion.div {...dialogEnterExitAnimation}>
                        <DialogContent className="max-w-2xl max-h-[85vh] overflow-y-auto scrollbar-neon p-0">
                            <motion.div {...dialogContentAnimation}>
                                <motion.div {...dialogHeaderAnimation}>
                                    <DialogHeader className="p-6 pb-3">
                                        <DialogTitle className="text-xl">{title}</DialogTitle>
                                        <DialogDescription className='dark:text-shadow-glow-white'>{description}</DialogDescription>
                                    </DialogHeader>
                                </motion.div>

                                <motion.div
                                    {...dialogContentItemAnimation}
                                    className="px-6 pb-6"
                                >
                                    <ChannelForm
                                        mode={mode}
                                        channelId={channel?.id}
                                        defaultValues={defaultValues}
                                        onSuccess={() => onOpenChange(false)}
                                    />
                                </motion.div>
                            </motion.div>
                        </DialogContent>
                    </motion.div>
                )}
            </AnimatePresence>
        </Dialog>
    )
}
