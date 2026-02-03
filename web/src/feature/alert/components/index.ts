// Alert 组件统一导出文件
import React from 'react'
import type { AlertHistory, AlertRule } from '@/types/alert'

// 通道管理组件
export { ChannelTable } from './ChannelTable'
export { ChannelDialog } from './ChannelDialog'
export { ChannelForm } from './ChannelForm'
export { DeleteChannelDialog } from './DeleteChannelDialog'
export { TestChannelDialog } from './TestChannelDialog'

// 规则管理组件类型定义
interface RuleTableProps {
    onEdit?: (rule: AlertRule) => void
    onDelete?: (id: string) => void
}

interface RuleDialogProps {
    open: boolean
    onOpenChange: React.Dispatch<React.SetStateAction<boolean>>
    mode: 'create' | 'update'
    rule?: AlertRule | null
}

interface DeleteRuleDialogProps {
    open: boolean
    onOpenChange: React.Dispatch<React.SetStateAction<boolean>>
    ruleId: string | null
}

// 历史查看组件类型定义
interface HistoryTableProps {
    onViewDetail?: (history: AlertHistory) => void
}

interface HistoryDetailDialogProps {
    open: boolean
    onOpenChange: React.Dispatch<React.SetStateAction<boolean>>
    history: AlertHistory | null
}

// 规则管理组件 - 占位符实现
// TODO: 需要完整实现这些组件
export const RuleTable: React.FC<RuleTableProps> = () => null
export const RuleDialog: React.FC<RuleDialogProps> = () => null
export const DeleteRuleDialog: React.FC<DeleteRuleDialogProps> = () => null

// 历史查看组件 - 占位符实现
// TODO: 需要完整实现这些组件
export const HistoryTable: React.FC<HistoryTableProps> = () => null
export const HistoryDetailDialog: React.FC<HistoryDetailDialogProps> = () => null
export const AlertStatsCards: React.FC = () => null

