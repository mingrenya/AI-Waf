// Alert 组件统一导出文件

// 通道管理组件
export { ChannelTable } from './ChannelTable'
export { ChannelDialog } from './ChannelDialog'
export { ChannelForm } from './ChannelForm'
export { DeleteChannelDialog } from './DeleteChannelDialog'
export { TestChannelDialog } from './TestChannelDialog'

// 规则管理组件 - 占位符实现
// TODO: 需要完整实现这些组件
export const RuleTable: React.FC<any> = () => null
export const RuleDialog: React.FC<any> = () => null
export const DeleteRuleDialog: React.FC<any> = () => null

// 历史查看组件 - 占位符实现
// TODO: 需要完整实现这些组件
export const HistoryTable: React.FC<any> = () => null
export const HistoryDetailDialog: React.FC<any> = () => null
export const AlertStatsCards: React.FC = () => null

