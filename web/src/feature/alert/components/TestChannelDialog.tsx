import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
    DialogFooter
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import { AlertChannel } from '@/types/alert'
import { useMutation } from '@tanstack/react-query'
import { alertChannelApi } from '@/api/alert'
import { useToast } from '@/hooks/use-toast'
import { useTranslation } from 'react-i18next'
import { useState } from 'react'
import { Loader2 } from 'lucide-react'

interface TestChannelDialogProps {
    open: boolean
    onOpenChange: (open: boolean) => void
    channel: AlertChannel | null
}

export function TestChannelDialog({ open, onOpenChange, channel }: TestChannelDialogProps) {
    const { t } = useTranslation()
    const { toast } = useToast()
    const [message, setMessage] = useState('This is a test alert message from AI-WAF.')

    const testMutation = useMutation({
        mutationFn: (data: { id: string, message: string }) => 
            alertChannelApi.testChannel(data.id, { message: data.message }),
        onSuccess: () => {
            toast({ title: 'Success', description: t('alert.testChannelSuccess') })
            onOpenChange(false)
        },
        onError: (error: any) => {
            toast({ title: 'Error', description: error?.response?.data?.message || t('alert.testChannelFailed'), variant: 'destructive' })
        }
    })

    const handleTest = () => {
        if (channel && message.trim()) {
            testMutation.mutate({ id: channel.id, message })
        }
    }

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="max-w-md">
                <DialogHeader>
                    <DialogTitle>{t('alert.testChannel')}</DialogTitle>
                    <DialogDescription className='dark:text-shadow-glow-white'>
                        {t('alert.testChannelDescription')}
                    </DialogDescription>
                </DialogHeader>
                
                <div className="space-y-4 py-4">
                    <div className="space-y-2">
                        <Label htmlFor="test-message">{t('alert.testMessage')}</Label>
                        <Textarea
                            id="test-message"
                            value={message}
                            onChange={(e) => setMessage(e.target.value)}
                            placeholder={t('alert.testMessagePlaceholder')}
                            rows={4}
                            className="resize-none"
                        />
                    </div>
                </div>

                <DialogFooter>
                    <Button
                        variant="outline"
                        onClick={() => onOpenChange(false)}
                        disabled={testMutation.isPending}
                    >
                        {t('alert.deleteDialog.cancel')}
                    </Button>
                    <Button
                        onClick={handleTest}
                        disabled={testMutation.isPending || !message.trim()}
                    >
                        {testMutation.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                        {t('alert.sendTest')}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    )
}
