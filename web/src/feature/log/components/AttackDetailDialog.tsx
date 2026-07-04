import { useState } from "react"
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { Copy, Check, AlertTriangle, Shield } from "lucide-react"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import type { AttackDetailData } from "@/types/log"
import { format } from "date-fns"
import { Badge } from "@/components/ui/badge"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { AnimatePresence, motion } from "motion/react"
import {
    dialogEnterExitAnimation,
    dialogContentAnimation,
    dialogContentItemAnimation,
} from "@/components/ui/animation/dialog-animation"
import { useTranslation } from "react-i18next"
import { CopyableText } from "@/components/common/copyable-text"
import { useAttackerProfile } from "@/feature/situation/hooks/useSituationData"
import { quickAction } from "@/api/situation"
import { toast } from "@/store"
import { useAuthStore } from '@/store/auth'
import { hasPermission } from '@/lib/permissions'

import { TabsAnimationProvider } from "@/components/ui/animation/components/tab-animation"

interface AttackDetailDialogProps {
    open: boolean
    onOpenChange: (open: boolean) => void
    data: AttackDetailData | null
}

export function AttackDetailDialog({ open, onOpenChange, data }: AttackDetailDialogProps) {
    const [copyState, setCopyState] = useState<{ [key: string]: boolean }>({})
    const [encoding, setEncoding] = useState("UTF-8")
    const [activeTab, setActiveTab] = useState("request")
    const { t, i18n } = useTranslation()
    const sourceIP = data?.srcIp
    const { data: profile } = useAttackerProfile(sourceIP || '')
    const currentUser = useAuthStore((state) => state.user)

    const [quickActionLoading, setQuickActionLoading] = useState<string | null>(null)

    // Define the gradients for reuse
    const purpleGradient = `linear-gradient(135deg,
    rgba(147, 112, 219, 0.95) 0%,
    rgba(138, 100, 208, 0.9) 50%,
    rgba(123, 79, 214, 0.95) 100%)`

    const darkPurpleGradient = `linear-gradient(135deg,
    rgba(113, 70, 199, 0.8) 0%,
    rgba(91, 52, 171, 0.75) 50%,
    rgba(72, 38, 153, 0.8) 100%)`

    const handleCopy = (text: string, key: string) => {
        navigator.clipboard.writeText(text).then(() => {
            setCopyState((prev) => ({ ...prev, [key]: true }))
            setTimeout(() => setCopyState((prev) => ({ ...prev, [key]: false })), 2000)
        })
    }

    const handleQuickBlock = async (action: 'block' | 'blacklist' | 'both', hours: number, label: string) => {
        if (!sourceIP) return
        setQuickActionLoading(label)
        try {
            const result = await quickAction({
                source_ip: sourceIP,
                action,
                duration_hours: hours,
                reason: `从攻击日志详情处置 - ${data?.message || '未知攻击'}`,
                correlation_id: data?.requestId,
            })
            toast({ title: '处置成功', description: result.data.data.note, variant: 'success' })
        } catch {
            toast({ title: '处置失败', description: '请稍后重试', variant: 'destructive' })
        } finally {
            setQuickActionLoading(null)
        }
    }

    if (!data) return null

    // 构建curl命令
    const curlCommand = `curl -X GET "${data.target}"`

    // 为了演示，假设规则ID > 1000 的是高危规则
    const isHighRisk = data.ruleId > 1000

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <AnimatePresence>
                {open && (
                    <motion.div {...dialogEnterExitAnimation}>
                        <DialogContent className="surface-modal sm:max-w-[90vw] lg:max-w-[75vw] xl:max-w-[65vw] max-h-[90vh] w-full p-0 gap-0 overflow-hidden " style={{ color: 'var(--text-primary)' }}>
                            <motion.div {...dialogContentAnimation}>
                                <DialogHeader
                                    className="px-6 py-4"
                                    style={{
                                        background: purpleGradient,
                                    }}
                                >
                                    <motion.div {...dialogContentItemAnimation}>
                                        <div className="flex items-center gap-2">
                                            <DialogTitle className="text-xl font-semibold flex items-center gap-2">
                                                {isHighRisk && <AlertTriangle className="h-5 w-5 text-yellow-300 " />}
                                                {t("attackDetail.title")}
                                            </DialogTitle>
                                            {isHighRisk && (
                                                <Badge variant="destructive" className="ml-2 bg-purple-300 text-purple-800 hover:bg-purple-300 dark:bg-purple-900/70 dark:text-purple-200 ">
                                                    {t("attackDetail.highRiskAttack")}
                                                </Badge>
                                            )}
                                        </div>
                                    </motion.div>
                                </DialogHeader>

                                <ScrollArea scrollbarVariant="none" className="px-4 py-2 h-[calc(90vh-6rem)] transition-colors duration-200">
                                    <div className="space-y-2 p-0 max-w-full max-h-full">
                                        {/* 攻击概述卡片 */}
                                        <motion.div {...dialogContentItemAnimation}>
                                            <Card className="p-6 bg-card border-none shadow-none rounded-sm bg-gradient-to-r from-purple-50 to-white dark:from-purple-950/20 dark:to-accent/50 ">
                                                <h3 className="text-lg font-semibold mb-4 flex items-center gap-2 text-card-foreground dark:text-slate-200">
                                                    <Shield className="h-5 w-5 text-purple-600 dark:text-purple-400 " />
                                                    {t("attackDetail.overview")}
                                                </h3>
                                                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                                                    <div className="space-y-4">
                                                        <div>
                                                            <span className="text-muted-foreground text-sm block mb-1">
                                                                {t("attackDetail.target")}
                                                            </span>
                                                            <div className="font-medium truncate text-card-foreground dark:text-slate-200">
                                                                <CopyableText text={data.target} className="font-medium text-card-foreground dark:text-slate-200" />
                                                            </div>
                                                        </div>
                                                        <div>
                                                            <span className="text-muted-foreground text-sm block mb-1">
                                                                {t("attackDetail.message")}
                                                            </span>
                                                            <div className="font-medium text-card-foreground dark:text-slate-200">{data.message}</div>
                                                        </div>
                                                        <div>
                                                            <span className="text-muted-foreground text-sm block mb-1">{t("requestId")}</span>
                                                            <div className="font-mono text-sm flex items-center gap-1 text-card-foreground dark:text-slate-200">
                                                                {data.requestId}
                                                                <Button
                                                                    variant="ghost"
                                                                    size="icon"
                                                                    className="h-6 w-6 text-muted-foreground hover:text-card-foreground dark:text-slate-400 dark:hover:text-slate-200 "
                                                                    onClick={() => handleCopy(data.requestId, "requestId")}
                                                                >
                                                                    {copyState["requestId"] ? (
                                                                        <Check className="h-3 w-3 " />
                                                                    ) : (
                                                                        <Copy className="h-3 w-3 " />
                                                                    )}
                                                                </Button>
                                                            </div>
                                                        </div>
                                                    </div>
                                                    <div className="space-y-4">
                                                        <div>
                                                            <span className="text-muted-foreground text-sm block mb-1">{t("ruleId")}</span>
                                                            <div className="font-medium flex items-center gap-2 text-card-foreground dark:text-slate-200">
                                                                {data.ruleId}
                                                                {/* <Button
                                                                    variant="outline"
                                                                    size="sm"
                                                                    className="h-7 text-xs border-border hover:bg-accent dark:border-slate-700 dark:hover:bg-slate-800"
                                                                >
                                                                    {t("attackDetail.viewRuleDetail")}
                                                                    <ArrowUpRight className="h-3 w-3 ml-1 " />
                                                                </Button> */}
                                                            </div>
                                                        </div>
                                                        <div>
                                                            <span className="text-muted-foreground text-sm block mb-1">
                                                                {t("attackDetail.attackTime")}
                                                            </span>
                                                            <div className="font-medium text-card-foreground dark:text-slate-200">
                                                                {format(new Date(data.createdAt), "yyyy-MM-dd HH:mm:ss")}
                                                            </div>
                                                        </div>
                                                    </div>
                                                </div>
                                            </Card>
                                        </motion.div>

                                        {/* 载荷信息 */}
                                        <motion.div {...dialogContentItemAnimation}>
                                            <Card className="p-6 bg-card border-none shadow-none dark:bg-accent/20 ">
                                                <h3 className="text-lg font-semibold mb-4 text-card-foreground dark:text-slate-200">
                                                    {t("attackDetail.detectedPayload")}
                                                </h3>
                                                <div className="bg-muted rounded-md p-4 border-none bg-zinc-100 dark:bg-zinc-800/70">
                                                    <div className="flex items-center justify-between">
                                                        <code className="text-sm break-all font-mono text-card-foreground dark:text-slate-200 whitespace-pre-wrap break-words block w-full overflow-hidden">
                                                            {data.payload}
                                                        </code>
                                                        <Button
                                                            variant="ghost"
                                                            size="icon"
                                                            onClick={() => handleCopy(data.payload, "payload")}
                                                            className="text-muted-foreground hover:text-card-foreground dark:text-slate-400 dark:hover:text-slate-200 "
                                                        >
                                                            {copyState["payload"] ? <Check className="h-4 w-4 " /> : <Copy className="h-4 w-4 " />}
                                                        </Button>
                                                    </div>
                                                </div>
                                            </Card>
                                        </motion.div>

                                        {/* 攻击者画像简要面板 */}
                                        {profile && (
                                            <motion.div {...dialogContentItemAnimation}>
                                                <Card className="p-4 bg-card border rounded-lg bg-muted/30 dark:bg-accent/20 ">
                                                    <div className="flex items-center justify-between mb-3">
                                                        <h4 className="font-semibold text-sm text-card-foreground dark:text-slate-200">攻击者画像</h4>
                                                        <Badge variant={
                                                            profile.risk_label === 'critical' ? 'destructive' :
                                                                profile.risk_label === 'high' ? 'destructive' :
                                                                    profile.risk_label === 'medium' ? 'secondary' : 'outline'
                                                        }>
                                                            风险: {profile.risk_label} ({profile.risk_score})
                                                        </Badge>
                                                    </div>
                                                    <div className="grid grid-cols-2 gap-y-1 text-xs text-muted-foreground">
                                                        <span>攻击阶段: {profile.attack_phase}</span>
                                                        <span>24h攻击: {profile.total_attacks}次</span>
                                                        <span>攻击类型: {profile.unique_attack_types}种</span>
                                                        <span>疑似工具: {profile.tools_identified || '未识别'}</span>
                                                    </div>
                                                </Card>
                                            </motion.div>
                                        )}

                                        {/* 来源和目标信息 */}
                                        <motion.div {...dialogContentItemAnimation}>
                                            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                                                {/* 攻击来源 */}
                                                <Card className="p-6 bg-card border-none shadow-none bg-gradient-to-r from-purple-100 to-white dark:from-purple-900/20 dark:to-accent/40 rounded-sm ">
                                                    <h3 className="text-lg font-semibold mb-4 text-card-foreground dark:text-slate-200">
                                                        {t("attackDetail.attackSource")}
                                                    </h3>
                                                    <div className="space-y-4">
                                                        <div>
                                                            <span className="text-muted-foreground text-sm block mb-1">{t("srcIp")}</span>
                                                            <div className="font-medium text-card-foreground dark:text-slate-200">
                                                                <span className="break-all font-mono">{data.srcIp}</span>
                                                            </div>
                                                        </div>
                                                        <div>
                                                            <span className="text-muted-foreground text-sm block mb-1">{t("srcPort")}</span>
                                                            <div className="font-medium font-mono text-card-foreground dark:text-slate-200">{data.srcPort}</div>
                                                        </div>
                                                        {data.srcIpInfo && (
                                                            <div>
                                                                <span className="text-muted-foreground text-sm block mb-1">{t("attackDetail.location")}</span>
                                                                <div className="font-medium text-card-foreground dark:text-slate-200">
                                                                    {data.srcIpInfo.country ? (
                                                                        <div>
                                                                            {i18n.language === 'zh'
                                                                                ? data.srcIpInfo.country.nameZh
                                                                                : data.srcIpInfo.country.nameEn}
                                                                            {data.srcIpInfo.subdivision &&
                                                                                <span> - {i18n.language === 'zh'
                                                                                    ? data.srcIpInfo.subdivision.nameZh
                                                                                    : data.srcIpInfo.subdivision.nameEn}
                                                                                </span>
                                                                            }
                                                                            {data.srcIpInfo.city &&
                                                                                <span> - {i18n.language === 'zh'
                                                                                    ? data.srcIpInfo.city.nameZh
                                                                                    : data.srcIpInfo.city.nameEn}
                                                                                </span>
                                                                            }
                                                                        </div>
                                                                    ) : (
                                                                        t("attackDetail.noLocationInfo")
                                                                    )}
                                                                </div>
                                                            </div>
                                                        )}
                                                    </div>
                                                </Card>

                                                {/* 目标信息 */}
                                                <Card className="p-6 bg-card border-none shadow-none bg-gradient-to-r from-purple-100 to-white dark:from-purple-900/20 dark:to-accent/40 rounded-sm ">
                                                    <h3 className="text-lg font-semibold mb-4 text-card-foreground dark:text-slate-200">
                                                        {t("attackDetail.targetInfo")}
                                                    </h3>
                                                    <div className="space-y-4">
                                                        <div>
                                                            <span className="text-muted-foreground text-sm block mb-1">{t("dstIp")}</span>
                                                            <div className="font-medium font-mono break-all text-card-foreground dark:text-slate-200">{data.dstIp}</div>
                                                        </div>
                                                        <div>
                                                            <span className="text-muted-foreground text-sm block mb-1">{t("dstPort")}</span>
                                                            <div className="font-medium font-mono text-card-foreground dark:text-slate-200">{data.dstPort}</div>
                                                        </div>
                                                    </div>
                                                </Card>
                                            </div>
                                        </motion.div>

                                        {/* 请求详情选项卡 */}
                                        <motion.div {...dialogContentItemAnimation}>
                                            <Card className="p-6 bg-card border-none shadow-none rounded-sm dark:bg-accent/20 ">
                                                <Tabs defaultValue="request" className="w-full" onValueChange={(value) => setActiveTab(value)}>
                                                    <div className="flex justify-between items-center mb-4">
                                                        <h3 className="text-lg font-semibold text-card-foreground dark:text-slate-200">
                                                            {t("attackDetail.technicalDetails")}
                                                        </h3>
                                                        <div className="flex items-center gap-2">
                                                            <Button
                                                                variant="outline"
                                                                size="sm"
                                                                onClick={() => handleCopy(curlCommand, "curl")}
                                                                className="flex items-center gap-1 h-8 border-border hover:bg-purple-100 dark:border-slate-700 dark:hover:bg-purple-900/20 dark:text-slate-200 "
                                                            >
                                                                {copyState["curl"] ? <Check className="h-3 w-3 " /> : <Copy className="h-3 w-3 " />}
                                                                {t("attackDetail.copyCurl")}
                                                            </Button>

                                                            <Select value={encoding} onValueChange={setEncoding}>
                                                                <SelectTrigger className="w-[110px] h-8 border-purple-200 dark:border-purple-800 dark:bg-accent/40 dark:text-slate-200">
                                                                    <SelectValue placeholder={t("attackDetail.encoding")} />
                                                                </SelectTrigger>
                                                                <SelectContent className="dark:bg-accent dark:border-slate-700">
                                                                    <SelectItem value="UTF-8" className="dark:text-slate-200">UTF-8</SelectItem>
                                                                    <SelectItem value="GBK" className="dark:text-slate-200">GBK</SelectItem>
                                                                    <SelectItem value="ISO-8859-1" className="dark:text-slate-200">ISO-8859-1</SelectItem>
                                                                </SelectContent>
                                                            </Select>
                                                        </div>
                                                    </div>

                                                    <TabsList className="mb-3 w-full bg-purple-100 dark:bg-purple-900/30 transition-all duration-300 ease-in-out">
                                                        <TabsTrigger
                                                            value="request"
                                                            className="flex-1 transition-all duration-300 ease-in-out transform hover:scale-[1.02] text-purple-800 dark:text-purple-300"
                                                            style={{
                                                                background: activeTab === "request" ? (document.documentElement.classList.contains('dark') ? darkPurpleGradient : purpleGradient) : "transparent",
                                                                color: activeTab === "request" ? "white" : "inherit",
                                                            }}
                                                        >
                                                            {t("attackDetail.request")}
                                                        </TabsTrigger>
                                                        <TabsTrigger
                                                            value="response"
                                                            className="flex-1 transition-all duration-300 ease-in-out transform hover:scale-[1.02] text-purple-800 dark:text-purple-300"
                                                            style={{
                                                                background: activeTab === "response" ? (document.documentElement.classList.contains('dark') ? darkPurpleGradient : purpleGradient) : "transparent",
                                                                color: activeTab === "response" ? "white" : "inherit",
                                                            }}
                                                        >
                                                            {t("attackDetail.response")}
                                                        </TabsTrigger>
                                                        <TabsTrigger
                                                            value="logs"
                                                            className="flex-1 transition-all duration-300 ease-in-out transform hover:scale-[1.02] text-purple-800 dark:text-purple-300"
                                                            style={{
                                                                background: activeTab === "logs" ? (document.documentElement.classList.contains('dark') ? darkPurpleGradient : purpleGradient) : "transparent",
                                                                color: activeTab === "logs" ? "white" : "inherit",
                                                            }}
                                                        >
                                                            {t("attackDetail.logs")}
                                                        </TabsTrigger>
                                                    </TabsList>

                                                    <TabsAnimationProvider currentView={activeTab} animationVariant="slide">
                                                        {activeTab === "request" ? (
                                                            <TabsContent value="request" forceMount>
                                                                <div className="border rounded-md overflow-hidden bg-muted/10 border-border dark:border-slate-700 dark:bg-accent/10">
                                                                    <div className="flex justify-end p-2 bg-muted border-b border-border dark:bg-slate-800 dark:border-slate-700">
                                                                        <Button
                                                                            variant="ghost"
                                                                            size="sm"
                                                                            className="h-7 text-muted-foreground hover:text-card-foreground dark:text-slate-400 dark:hover:text-slate-200 "
                                                                            onClick={() => handleCopy(data.request, "requestCopy")}
                                                                        >
                                                                            {copyState["requestCopy"] ? (
                                                                                <Check className="h-3 w-3 mr-1 " />
                                                                            ) : (
                                                                                <Copy className="h-3 w-3 mr-1 " />
                                                                            )}
                                                                            {t("attackDetail.copyAll")}
                                                                        </Button>
                                                                    </div>
                                                                    <ScrollArea scrollbarVariant="neon" className="transition-colors duration-200 max-h-[18.75rem] overflow-auto scrollbar-custom">
                                                                        <pre className="p-4 text-sm whitespace-pre-wrap font-mono text-card-foreground bg-background dark:bg-slate-900 dark:text-slate-200">
                                                                            <code className="text-sm break-all font-mono text-card-foreground dark:text-slate-200 whitespace-pre-wrap break-words block w-full overflow-hidden">
                                                                                {data.request}
                                                                            </code>
                                                                        </pre>
                                                                    </ScrollArea>
                                                                </div>
                                                            </TabsContent>
                                                        ) : activeTab === "response" ? (
                                                            <TabsContent value="response" forceMount>
                                                                <div className="border rounded-md overflow-hidden bg-muted/10 border-border dark:border-slate-700 dark:bg-accent/10">
                                                                    <div className="flex justify-end p-2 bg-muted border-b border-border dark:bg-slate-800 dark:border-slate-700">
                                                                        <Button
                                                                            variant="ghost"
                                                                            size="sm"
                                                                            className="h-7 text-muted-foreground hover:text-card-foreground dark:text-slate-400 dark:hover:text-slate-200 "
                                                                            onClick={() => handleCopy(data.response, "responseCopy")}
                                                                        >
                                                                            {copyState["responseCopy"] ? (
                                                                                <Check className="h-3 w-3 mr-1 " />
                                                                            ) : (
                                                                                <Copy className="h-3 w-3 mr-1 " />
                                                                            )}
                                                                            {t("attackDetail.copyAll")}
                                                                        </Button>
                                                                    </div>
                                                                    <ScrollArea scrollbarVariant="neon" className="transition-colors duration-200 max-h-[18.75rem] overflow-auto">
                                                                        <pre className="p-4 text-sm whitespace-pre-wrap font-mono text-card-foreground bg-background dark:bg-slate-900 dark:text-slate-200">
                                                                            <code className="text-sm break-all font-mono text-card-foreground dark:text-slate-200 whitespace-pre-wrap break-words block w-full overflow-hidden">
                                                                                {data.response ? data.response : t("attackDetail.noResponse")}
                                                                            </code>
                                                                        </pre>
                                                                    </ScrollArea>
                                                                </div>
                                                            </TabsContent>
                                                        ) : (
                                                            <TabsContent value="logs" forceMount>
                                                                <div className="border rounded-md overflow-hidden bg-muted/10 border-border dark:border-slate-700 dark:bg-accent/10">
                                                                    <div className="flex justify-end p-2 bg-muted border-b border-border dark:bg-slate-800 dark:border-slate-700">
                                                                        <Button
                                                                            variant="ghost"
                                                                            size="sm"
                                                                            className="h-7 text-muted-foreground hover:text-card-foreground dark:text-slate-400 dark:hover:text-slate-200 "
                                                                            onClick={() => handleCopy(data.logs, "logsCopy")}
                                                                        >
                                                                            {copyState["logsCopy"] ? (
                                                                                <Check className="h-3 w-3 mr-1 " />
                                                                            ) : (
                                                                                <Copy className="h-3 w-3 mr-1 " />
                                                                            )}
                                                                            {t("attackDetail.copyAll")}
                                                                        </Button>
                                                                    </div>
                                                                    <ScrollArea scrollbarVariant="neon" className="transition-colors duration-200 max-h-[18.75rem] overflow-auto">
                                                                        <pre className="p-4 text-sm whitespace-pre-wrap font-mono text-card-foreground bg-background dark:bg-slate-900 dark:text-slate-200">
                                                                            <code className="text-sm break-all font-mono text-card-foreground dark:text-slate-200 whitespace-pre-wrap break-words block w-full overflow-hidden">
                                                                                {data.logs}
                                                                            </code>
                                                                        </pre>
                                                                    </ScrollArea>
                                                                </div>
                                                            </TabsContent>
                                                        )}
                                                    </TabsAnimationProvider>
                                                </Tabs>
                                            </Card>
                                        </motion.div>

                                        {/* 快速处置按钮 */}
                                        {sourceIP && hasPermission(currentUser, 'config:update') && (
                                            <motion.div {...dialogContentItemAnimation}>
                                                <div className="pt-3 border-t flex flex-wrap gap-2">
                                                    <span className="text-xs text-muted-foreground flex items-center gap-1 w-full mb-1">
                                                        <Shield className="h-3.5 w-3.5" />
                                                        快速处置:
                                                    </span>
                                                    <Button
                                                        size="sm"
                                                        variant="destructive"
                                                        className="text-xs h-7"
                                                        disabled={quickActionLoading !== null}
                                                        onClick={() => handleQuickBlock('block', 1, 'Block 1h')}
                                                    >
                                                        {quickActionLoading === 'Block 1h' ? (
                                                            <span className="flex items-center gap-1">
                                                                <span className="animate-spin h-3 w-3 border-2 border-current border-t-transparent rounded-full" />
                                                                处理中...
                                                            </span>
                                                        ) : '封禁 1h'}
                                                    </Button>
                                                    <Button
                                                        size="sm"
                                                        variant="destructive"
                                                        className="text-xs h-7"
                                                        disabled={quickActionLoading !== null}
                                                        onClick={() => handleQuickBlock('block', 24, 'Block 24h')}
                                                    >
                                                        {quickActionLoading === 'Block 24h' ? (
                                                            <span className="flex items-center gap-1">
                                                                <span className="animate-spin h-3 w-3 border-2 border-current border-t-transparent rounded-full" />
                                                                处理中...
                                                            </span>
                                                        ) : '封禁 24h'}
                                                    </Button>
                                                    <Button
                                                        size="sm"
                                                        variant="destructive"
                                                        className="text-xs h-7"
                                                        disabled={quickActionLoading !== null}
                                                        onClick={() => handleQuickBlock('block', 168, 'Block 7d')}
                                                    >
                                                        {quickActionLoading === 'Block 7d' ? (
                                                            <span className="flex items-center gap-1">
                                                                <span className="animate-spin h-3 w-3 border-2 border-current border-t-transparent rounded-full" />
                                                                处理中...
                                                            </span>
                                                        ) : '封禁 7d'}
                                                    </Button>
                                                    <Button
                                                        size="sm"
                                                        variant="destructive"
                                                        className="text-xs h-7"
                                                        disabled={quickActionLoading !== null}
                                                        onClick={() => handleQuickBlock('block', 87600, 'Block Permanent')}
                                                    >
                                                        {quickActionLoading === 'Block Permanent' ? (
                                                            <span className="flex items-center gap-1">
                                                                <span className="animate-spin h-3 w-3 border-2 border-current border-t-transparent rounded-full" />
                                                                处理中...
                                                            </span>
                                                        ) : '永久封禁'}
                                                    </Button>
                                                    <Button
                                                        size="sm"
                                                        variant="outline"
                                                        className="text-xs h-7"
                                                        disabled={quickActionLoading !== null}
                                                        onClick={() => handleQuickBlock('blacklist', 0, 'Add Blacklist')}
                                                    >
                                                        {quickActionLoading === 'Add Blacklist' ? (
                                                            <span className="flex items-center gap-1">
                                                                <span className="animate-spin h-3 w-3 border-2 border-current border-t-transparent rounded-full" />
                                                                处理中...
                                                            </span>
                                                        ) : '加入黑名单'}
                                                    </Button>
                                                </div>
                                            </motion.div>
                                        )}
                                    </div>
                                </ScrollArea>
                            </motion.div>
                        </DialogContent>
                    </motion.div>
                )}
            </AnimatePresence>
        </Dialog>
    )
}
