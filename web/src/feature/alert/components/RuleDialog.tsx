import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog'
import { AnimatePresence, motion } from 'motion/react'
import {
    dialogEnterExitAnimation,
    dialogContentAnimation,
    dialogHeaderAnimation,
    dialogContentItemAnimation,
} from '@/components/ui/animation/dialog-animation'
import { useTranslation } from 'react-i18next'
import type { AlertRule } from '@/types/alert'
import { RuleForm, type RuleFormValues } from './RuleForm'

interface RuleDialogProps {
    open: boolean
    onOpenChange: (open: boolean) => void
    mode: 'create' | 'update'
    rule?: AlertRule | null
}

export function RuleDialog({
    open,
    onOpenChange,
    mode = 'create',
    rule = null,
}: RuleDialogProps) {
    const { t } = useTranslation()

    const title =
        mode === 'create'
            ? t('alert.dialog.createRuleTitle')
            : t('alert.dialog.updateRuleTitle')
    const description =
        mode === 'create'
            ? t('alert.dialog.createRuleDescription')
            : t('alert.dialog.updateRuleDescription')

    // 将后端 Rule 数据转换为表单默认值
    const defaultValues: RuleFormValues | undefined =
        mode === 'update' && rule
            ? {
                  name: rule.name,
                  description: rule.description ?? '',
                  enabled: rule.enabled,
                  severity: rule.severity,
                  logic: rule.logic as 'AND' | 'OR',
                  cooldown: rule.cooldown != null ? String(rule.cooldown) : '',
                  template: rule.template ?? '',
                  channels: (rule.channels || []).join(','),
                  conditions: (rule.conditions || []).map((c) => ({
                      metric: c.metric,
                      operator: c.operator as unknown as RuleFormValues['conditions'][number]['operator'],
                      threshold: c.threshold != null ? String(c.threshold) : '',
                      duration: c.duration != null ? String(c.duration) : '',
                  })),
              }
            : undefined

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <AnimatePresence mode="wait">
                {open && (
                    <motion.div {...dialogEnterExitAnimation}>
                        <DialogContent className="max-w-3xl max-h-[85vh] overflow-y-auto scrollbar-neon p-0">
                            <motion.div {...dialogContentAnimation}>
                                <motion.div {...dialogHeaderAnimation}>
                                    <DialogHeader className="p-6 pb-3">
                                        <DialogTitle className="text-xl">
                                            {title}
                                        </DialogTitle>
                                        <DialogDescription className="dark:text-shadow-glow-white">
                                            {description}
                                        </DialogDescription>
                                    </DialogHeader>
                                </motion.div>

                                <motion.div
                                    {...dialogContentItemAnimation}
                                    className="px-6 pb-6"
                                >
                                    <RuleForm
                                        mode={mode}
                                        ruleId={rule?.id}
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

