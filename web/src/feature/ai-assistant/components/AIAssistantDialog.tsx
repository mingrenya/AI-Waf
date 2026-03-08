import { useState, useRef, useEffect } from 'react'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Badge } from '@/components/ui/badge'
import { Send, Loader2, Bot, User, Sparkles } from 'lucide-react'
import { cn } from '@/lib/utils'
import { sendChatMessage } from '@/api/mcp'
import { toast } from '@/store'
import type { AIAssistantMessage } from '@/types/mcp'

interface AIAssistantDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function AIAssistantDialog({ open, onOpenChange }: AIAssistantDialogProps) {
  const [messages, setMessages] = useState<AIAssistantMessage[]>([
    {
      id: 'welcome',
      role: 'assistant',
      content: '你好！我是AI安全助手。我可以帮你分析WAF日志、生成防护规则、评估攻击模式等。请告诉我你的需求。',
      timestamp: new Date().toISOString(),
    },
  ])
  const [input, setInput] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const scrollAreaRef = useRef<HTMLDivElement>(null)

  // 自动滚动到底部
  useEffect(() => {
    if (scrollAreaRef.current) {
      scrollAreaRef.current.scrollTop = scrollAreaRef.current.scrollHeight
    }
  }, [messages])

  const handleSend = async () => {
    if (!input.trim() || isLoading) return

    const userMessage: AIAssistantMessage = {
      id: Date.now().toString(),
      role: 'user',
      content: input,
      timestamp: new Date().toISOString(),
    }

    setMessages((prev) => [...prev, userMessage])
    const currentInput = input
    setInput('')
    setIsLoading(true)

    try {
      // 构建历史消息
      const historyMessages = messages.map(msg => ({
        role: msg.role,
        content: msg.content
      }))

      // 调用后端API
      const response = await sendChatMessage({
        message: currentInput,
        messages: historyMessages,
        stream: false
      })
      
      // 从响应中提取聊天数据
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const responseData = response as any
      const chatData = responseData.data?.data || responseData.data || responseData
      const aiMessage = chatData.message || '无响应'
      const aiToolCalls = chatData.toolCalls || []

      const assistantMessage: AIAssistantMessage = {
        id: (Date.now() + 1).toString(),
        role: 'assistant',
        content: aiMessage,
        timestamp: new Date().toISOString(),
        toolCalls: aiToolCalls.length > 0 ? aiToolCalls.map((tool: string) => ({
          id: Date.now().toString(),
          toolName: tool,
          timestamp: new Date().toISOString(),
          duration: 0,
          success: true
        })) : undefined
      }

      setMessages((prev) => [...prev, assistantMessage])
    } catch (error) {
      if (import.meta.env.DEV) {
        console.error('Chat error:', error)
      }
      toast({
        title: '错误',
        description: error instanceof Error ? error.message : '发送消息失败，请检查DEEPSEEK_API_KEY配置',
        variant: 'destructive'
      })
      
      // 添加错误消息
      const errorMessage: AIAssistantMessage = {
        id: (Date.now() + 1).toString(),
        role: 'assistant',
        content: '抱歉，我现在无法处理您的请求。请确保已配置DEEPSEEK_API_KEY环境变量。',
        timestamp: new Date().toISOString(),
      }
      setMessages((prev) => [...prev, errorMessage])
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-4xl h-[600px] flex flex-col">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Bot className="h-5 w-5" />
            AI 安全助手
            <Sparkles className="h-4 w-4 text-yellow-500" />
          </DialogTitle>
          <DialogDescription>
            通过与AI助手交互，执行安全分析任务
          </DialogDescription>
        </DialogHeader>

        <div className="flex-1 flex flex-col gap-4 overflow-hidden">
          {/* 消息列表 */}
          <ScrollArea className="flex-1 pr-4" ref={scrollAreaRef}>
            <div className="space-y-4">
              {messages.map((message) => (
                <div
                  key={message.id}
                  className={cn(
                    'flex gap-3',
                    message.role === 'user' && 'flex-row-reverse'
                  )}
                >
                  <div
                    className={cn(
                      'flex h-8 w-8 shrink-0 select-none items-center justify-center rounded-md',
                      message.role === 'assistant'
                        ? 'bg-primary text-primary-foreground'
                        : 'bg-muted'
                    )}
                  >
                    {message.role === 'assistant' ? (
                      <Bot className="h-4 w-4" />
                    ) : (
                      <User className="h-4 w-4" />
                    )}
                  </div>
                  <div
                    className={cn(
                      'flex-1 space-y-2 rounded-lg px-4 py-3',
                      message.role === 'assistant'
                        ? 'bg-muted'
                        : 'bg-primary text-primary-foreground'
                    )}
                  >
                    <p className="text-sm whitespace-pre-wrap">{message.content}</p>
                    <span className="text-xs opacity-60">
                      {new Date(message.timestamp).toLocaleTimeString('zh-CN')}
                    </span>
                    {message.toolCalls && message.toolCalls.length > 0 && (
                      <div className="flex flex-wrap gap-2 mt-2">
                        {message.toolCalls.map((tool) => (
                          <Badge key={tool.id} variant="secondary" className="text-xs">
                            {tool.toolName}
                          </Badge>
                        ))}
                      </div>
                    )}
                  </div>
                </div>
              ))}
              {isLoading && (
                <div className="flex gap-3">
                  <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-md bg-primary text-primary-foreground">
                    <Bot className="h-4 w-4" />
                  </div>
                  <div className="flex items-center gap-2 px-4 py-3 bg-muted rounded-lg">
                    <Loader2 className="h-4 w-4 animate-spin" />
                    <span className="text-sm text-muted-foreground">正在思考...</span>
                  </div>
                </div>
              )}
            </div>
          </ScrollArea>



          {/* 输入区域 */}
          <div className="flex gap-2">
            <Textarea
              value={input}
              onChange={(e) => setInput(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter' && !e.shiftKey) {
                  e.preventDefault()
                  handleSend()
                }
              }}
              placeholder="输入你的问题或指令... (Shift + Enter 换行)"
              className="min-h-[60px] resize-none"
              disabled={isLoading}
            />
            <Button
              onClick={handleSend}
              disabled={!input.trim() || isLoading}
              size="icon"
              className="h-[60px] w-[60px]"
            >
              {isLoading ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                <Send className="h-4 w-4" />
              )}
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
