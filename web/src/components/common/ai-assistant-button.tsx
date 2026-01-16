/**
 * AI助手按钮组件
 * 提供快速访问AI助手的入口
 */
import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { MessageSquare, Sparkles } from 'lucide-react'
import { cn } from '@/lib/utils'
import { AIAssistantDialog } from '@/feature/ai-assistant/components/AIAssistantDialog'

interface AIAssistantButtonProps {
  className?: string
  variant?: 'default' | 'ghost' | 'outline'
}

export function AIAssistantButton({ className, variant = 'ghost' }: AIAssistantButtonProps) {
  const [open, setOpen] = useState(false)

  return (
    <>
      <Button
        variant={variant}
        size="sm"
        className={cn("gap-2", className)}
        onClick={() => setOpen(true)}
      >
        <MessageSquare className="h-4 w-4" />
        <Sparkles className="h-3 w-3 text-yellow-500" />
        <span>AI 助手</span>
      </Button>

      <AIAssistantDialog open={open} onOpenChange={setOpen} />
    </>
  )
}
