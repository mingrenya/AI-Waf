"use client"

import { useState, useEffect, useCallback } from "react"
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Shield, CheckCircle2, Loader2, Zap, Lock, ShieldCheck } from "lucide-react"
import { ruleEnhancedApi } from "@/api/rule-enhanced"
import type { ProtectionProfile } from "@/types/rule-enhanced"
import { PROTECTION_LEVELS } from "@/types/rule-enhanced"
import { useToast } from "@/hooks/use-toast"
import { motion, AnimatePresence } from "motion/react"

interface ProtectionProfileDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSuccess?: () => void
}

export function ProtectionProfileDialog({ open, onOpenChange, onSuccess }: ProtectionProfileDialogProps) {
  const [profiles, setProfiles] = useState<ProtectionProfile[]>([])
  const [loading, setLoading] = useState(false)
  const [applying, setApplying] = useState(false)
  const [selectedProfile, setSelectedProfile] = useState<ProtectionProfile | null>(null)
  const { toast } = useToast()

  const fetchProfiles = useCallback(async () => {
    setLoading(true)
    try {
      const response = await ruleEnhancedApi.listProfiles()
      setProfiles(response.items || [])
    } catch (error) {
      const err = error as Error
      toast({
        variant: "destructive",
        title: "加载失败",
        description: err.message || "无法加载保护配置文件"
      })
    } finally {
      setLoading(false)
    }
  }, [toast])

  useEffect(() => {
    if (open) {
      fetchProfiles()
    }
  }, [open, fetchProfiles])

  const handleApply = async () => {
    if (!selectedProfile) return

    setApplying(true)
    try {
      const response = await ruleEnhancedApi.applyProfile({
        profile_id: selectedProfile.id
      })

      toast({
        title: "应用成功",
        description: `已创建 ${response.created_count} 条规则`,
        duration: 5000
      })

      onOpenChange(false)
      onSuccess?.()
    } catch (error) {
      const err = error as Error
      toast({
        variant: "destructive",
        title: "应用失败",
        description: err.message || "无法应用保护配置"
      })
    } finally {
      setApplying(false)
    }
  }

  const getLevelIcon = (level: string) => {
    switch (level) {
      case "basic": return <Zap className="h-5 w-5" />
      case "standard": return <Shield className="h-5 w-5" />
      case "strict": return <Lock className="h-5 w-5" />
      default: return <Shield className="h-5 w-5" />
    }
  }

  const getLevelColor = (level: string) => {
    switch (level) {
      case "basic": return "from-blue-500/20 to-cyan-500/20 border-blue-500/30"
      case "standard": return "from-emerald-500/20 to-green-500/20 border-emerald-500/30"
      case "strict": return "from-orange-500/20 to-red-500/20 border-orange-500/30"
      default: return "from-gray-500/20 to-slate-500/20 border-gray-500/30"
    }
  }

  const getLevelBadgeColor = (level: string) => {
    switch (level) {
      case "basic": return "bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400"
      case "standard": return "bg-emerald-100 text-emerald-800 dark:bg-emerald-900/30 dark:text-emerald-400"
      case "strict": return "bg-orange-100 text-orange-800 dark:bg-orange-900/30 dark:text-orange-400"
      default: return "bg-gray-100 text-gray-800 dark:bg-gray-900/30 dark:text-gray-400"
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-5xl max-h-[85vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-3 text-2xl">
            <div className="p-2 rounded-lg bg-gradient-to-br from-indigo-500/20 to-purple-500/20 border border-indigo-500/30">
              <ShieldCheck className="h-6 w-6 text-indigo-600 dark:text-indigo-400" />
            </div>
            一键保护配置
          </DialogTitle>
          <DialogDescription className="text-base">
            选择适合您应用的保护级别，快速部署完整的 OWASP Top 10 防护规则
          </DialogDescription>
        </DialogHeader>

        {loading ? (
          <div className="flex items-center justify-center py-16">
            <Loader2 className="h-8 w-8 animate-spin text-primary" />
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6 py-6">
            <AnimatePresence mode="wait">
              {profiles.map((profile, index) => (
                <motion.div
                  key={profile.id}
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: index * 0.1 }}
                >
                  <Card
                    className={`cursor-pointer transition-all duration-300 hover:shadow-xl relative overflow-hidden ${
                      selectedProfile?.id === profile.id
                        ? `ring-2 ring-primary shadow-lg bg-gradient-to-br ${getLevelColor(profile.level)}`
                        : "hover:border-primary/50"
                    }`}
                    onClick={() => setSelectedProfile(profile)}
                  >
                    {selectedProfile?.id === profile.id && (
                      <motion.div
                        initial={{ scale: 0 }}
                        animate={{ scale: 1 }}
                        className="absolute top-3 right-3"
                      >
                        <CheckCircle2 className="h-6 w-6 text-primary" />
                      </motion.div>
                    )}

                    <CardHeader className="space-y-3">
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2">
                          {getLevelIcon(profile.level)}
                          <CardTitle className="text-lg">{profile.name}</CardTitle>
                        </div>
                        {profile.is_default && (
                          <Badge variant="secondary" className="text-xs">
                            推荐
                          </Badge>
                        )}
                      </div>
                      <Badge className={`w-fit ${getLevelBadgeColor(profile.level)}`}>
                        {PROTECTION_LEVELS[profile.level as keyof typeof PROTECTION_LEVELS]}
                      </Badge>
                    </CardHeader>

                    <CardContent className="space-y-4">
                      <CardDescription className="text-sm leading-relaxed min-h-[60px]">
                        {profile.description}
                      </CardDescription>

                      <div className="space-y-2">
                        <p className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                          包含分类
                        </p>
                        <div className="flex flex-wrap gap-1.5">
                          {profile.categories.slice(0, 3).map((category) => (
                            <Badge
                              key={category}
                              variant="outline"
                              className="text-xs font-normal"
                            >
                              {category.split('_').map(w => w.charAt(0).toUpperCase() + w.slice(1)).join(' ')}
                            </Badge>
                          ))}
                          {profile.categories.length > 3 && (
                            <Badge variant="outline" className="text-xs">
                              +{profile.categories.length - 3}
                            </Badge>
                          )}
                        </div>
                      </div>

                      <div className="pt-3 border-t">
                        <div className="flex items-center justify-between text-sm">
                          <span className="text-muted-foreground">规则模板</span>
                          <span className="font-semibold">{profile.template_ids.length} 个</span>
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                </motion.div>
              ))}
            </AnimatePresence>
          </div>
        )}

        {selectedProfile && (
          <motion.div
            initial={{ opacity: 0, height: 0 }}
            animate={{ opacity: 1, height: "auto" }}
            className="rounded-lg bg-gradient-to-r from-primary/5 to-primary/10 border border-primary/20 p-4"
          >
            <div className="flex items-start gap-3">
              <div className="p-2 rounded-full bg-primary/10">
                <Shield className="h-4 w-4 text-primary" />
              </div>
              <div className="flex-1">
                <h4 className="font-medium mb-1">即将应用</h4>
                <p className="text-sm text-muted-foreground">
                  将根据 <span className="font-semibold">{selectedProfile.name}</span> 配置文件创建 {selectedProfile.template_ids.length} 条规则。
                  已存在的规则将被跳过。
                </p>
              </div>
            </div>
          </motion.div>
        )}

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={applying}>
            取消
          </Button>
          <Button
            onClick={handleApply}
            disabled={!selectedProfile || applying}
            className="min-w-[120px]"
          >
            {applying && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            {applying ? "应用中..." : "应用配置"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
