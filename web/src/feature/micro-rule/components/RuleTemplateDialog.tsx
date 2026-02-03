"use client"

import { useState, useEffect } from "react"
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Shield, Loader2, AlertTriangle } from "lucide-react"
import { ruleEnhancedApi } from "@/api/rule-enhanced"
import type { RuleTemplate } from "@/types/rule-enhanced"
import { OWASP_CATEGORIES, SEVERITY_LABELS } from "@/types/rule-enhanced"
import { useToast } from "@/hooks/use-toast"

interface RuleTemplateDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSuccess?: () => void
}

export function RuleTemplateDialog({ open, onOpenChange, onSuccess }: RuleTemplateDialogProps) {
  const [templates, setTemplates] = useState<RuleTemplate[]>([])
  const [loading, setLoading] = useState(false)
  const [creating, setCreating] = useState(false)
  const [selectedTemplate, setSelectedTemplate] = useState<RuleTemplate | null>(null)
  const [customName, setCustomName] = useState("")
  const [categoryFilter, setCategoryFilter] = useState<string>("all")
  const [severityFilter, setSeverityFilter] = useState<string>("all")
  const { toast } = useToast()

  useEffect(() => {
    if (open) {
      fetchTemplates()
    }
    // 移除过滤器从依赖数组，在fetchTemplates中处理
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open])

  // 当过滤器变化时重新获取模板
  useEffect(() => {
    if (open) {
      fetchTemplates()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [categoryFilter, severityFilter])

  const fetchTemplates = async () => {
    setLoading(true)
    try {
      const params: Record<string, string> = {}
      if (categoryFilter !== "all") params.category = categoryFilter
      if (severityFilter !== "all") params.severity = severityFilter

      const response = await ruleEnhancedApi.listTemplates(params)
      
      // 安全检查：确保response存在且格式正确
      if (!response || typeof response !== 'object') {
        setTemplates([])
        toast({
          variant: "destructive",
          title: "加载失败",
          description: "后端服务返回异常，请检查服务是否正常运行"
        })
        return
      }
      
      const items = Array.isArray(response.items) ? response.items : []
      
      setTemplates(items)
      
      // 如果没有模板，提示用户
      if (items.length === 0) {
        toast({
          title: "提示",
          description: "暂无可用的OWASP规则模板，请联系管理员初始化模板数据"
        })
      }
    } catch (error) {
      const err = error as Error & { response?: { status?: number }; message?: string }
      if (import.meta.env.DEV) {
        console.error('加载OWASP模板失败:', error)
      }
      setTemplates([])
      
      let errorMessage = "无法加载OWASP规则模板，请检查后端服务和权限设置"
      if (err.response?.status === 401) {
        errorMessage = "未授权访问，请重新登录"
      } else if (err.response?.status === 500) {
        errorMessage = "后端服务错误，请联系管理员"
      } else if (err.message) {
        errorMessage = err.message
      }
      
      toast({
        variant: "destructive",
        title: "加载失败",
        description: errorMessage
      })
    } finally {
      setLoading(false)
    }
  }

  const handleCreate = async () => {
    if (!selectedTemplate) return

    setCreating(true)
    try {
      await ruleEnhancedApi.createRuleFromTemplate({
        template_id: selectedTemplate.id,
        custom_name: customName || undefined
      })

      toast({
        title: "创建成功",
        description: `规则"${customName || selectedTemplate.name}"已创建`
      })

      onOpenChange(false)
      onSuccess?.()
    } catch (error) {
      const err = error as Error
      toast({
        variant: "destructive",
        title: "创建失败",
        description: err.message || "无法创建规则"
      })
    } finally {
      setCreating(false)
    }
  }

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case "critical": return "bg-red-100 text-red-800 dark:bg-red-900/20 dark:text-red-400"
      case "high": return "bg-orange-100 text-orange-800 dark:bg-orange-900/20 dark:text-orange-400"
      case "medium": return "bg-yellow-100 text-yellow-800 dark:bg-yellow-900/20 dark:text-yellow-400"
      case "low": return "bg-blue-100 text-blue-800 dark:bg-blue-900/20 dark:text-blue-400"
      default: return "bg-gray-100 text-gray-800 dark:bg-gray-900/20 dark:text-gray-400"
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-4xl max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Shield className="h-5 w-5" />
            从 OWASP Top 10 模板创建规则
          </DialogTitle>
          <DialogDescription>
            选择预定义的 OWASP Top 10 安全规则模板快速创建防护规则
          </DialogDescription>
        </DialogHeader>

        {/* 过滤器 */}
        <div className="grid grid-cols-2 gap-4 py-4">
          <div>
            <Label>分类筛选</Label>
            <Select value={categoryFilter} onValueChange={setCategoryFilter}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">全部分类</SelectItem>
                {Object.entries(OWASP_CATEGORIES).map(([key, label]) => (
                  <SelectItem key={key} value={key}>{label}</SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div>
            <Label>严重等级</Label>
            <Select value={severityFilter} onValueChange={setSeverityFilter}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">全部等级</SelectItem>
                {Object.entries(SEVERITY_LABELS).map(([key, label]) => (
                  <SelectItem key={key} value={key}>{label}</SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </div>

        {/* 模板列表 */}
        <div className="space-y-3">
          {loading ? (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="h-6 w-6 animate-spin" />
            </div>
          ) : templates.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <AlertTriangle className="h-12 w-12 mx-auto mb-2 opacity-50" />
              <p>没有找到匹配的模板</p>
            </div>
          ) : (
            templates.map((template) => (
              <div
                key={template.id}
                className={`border rounded-lg p-4 cursor-pointer transition-colors ${
                  selectedTemplate?.id === template.id
                    ? "border-primary bg-primary/5"
                    : "hover:border-primary/50"
                }`}
                onClick={() => {
                  setSelectedTemplate(template)
                  setCustomName(template.name)
                }}
              >
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <h4 className="font-medium mb-1">{template.name}</h4>
                    <p className="text-sm text-muted-foreground mb-2">
                      {template.description}
                    </p>
                    <div className="flex flex-wrap gap-2">
                      <Badge className={getSeverityColor(template.severity)}>
                        {SEVERITY_LABELS[template.severity as keyof typeof SEVERITY_LABELS]}
                      </Badge>
                      {template.tags.map(tag => (
                        <Badge key={tag} variant="outline" className="text-xs">
                          {tag}
                        </Badge>
                      ))}
                    </div>
                  </div>
                  <Badge variant="secondary">
                    优先级: {template.priority}
                  </Badge>
                </div>
              </div>
            ))
          )}
        </div>

        {/* 自定义规则名称 */}
        {selectedTemplate && (
          <div className="space-y-2 pt-4 border-t">
            <Label htmlFor="custom-name">规则名称（可选自定义）</Label>
            <Input
              id="custom-name"
              value={customName}
              onChange={(e) => setCustomName(e.target.value)}
              placeholder="留空使用模板默认名称"
            />
          </div>
        )}

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            取消
          </Button>
          <Button
            onClick={handleCreate}
            disabled={!selectedTemplate || creating}
          >
            {creating && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            创建规则
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
