import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
    DialogFooter,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import type { AlertHistory } from '@/types/alert'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { alertHistoryApi } from '@/api/alert'
import { toast } from '@/store'
import { useTranslation } from 'react-i18next'
import { useState } from 'react'
import { Loader2 } from 'lucide-react'
import { queryKeys } from '@/lib/query-keys'

interface HistoryDetailDialogProps {
    open: boolean
    onOpenChange: (open: boolean) => void
    history: AlertHistory | null
}

export function HistoryDetailDialog({
    open,
    onOpenChange,
    history,
}: HistoryDetailDialogProps) {
    const { t } = useTranslation()
    const queryClient = useQueryClient()
    const [note, setNote] = useState('')

    const acknowledgeMutation = useMutation({
        mutationFn: (id: string) =>
            alertHistoryApi.acknowledgeAlert(id, { comment: note || undefined }),
        onSuccess: () => {
            queryClient.invalidateQueries({
                queryKey: queryKeys.alert.history.lists(),
            })
            queryClient.invalidateQueries({
                queryKey: queryKeys.alert.history.stats(),
            })
            toast({
                title: t('common.success'),
                description: t('alert.acknowledgeSuccess', {
                    defaultValue: 'Alert acknowledged successfully',
                }),
                variant: 'success',
            })
            onOpenChange(false)
            setNote('')
        },
        onError: () => {
            toast({
                title: t('common.error'),
                description: t('alert.acknowledgeFailed', {
                    defaultValue: 'Failed to acknowledge alert',
                }),
                variant: 'destructive',
            })
        },
    })

    const canAcknowledge =
        !!history && history.status !== 'acknowledged' && !acknowledgeMutation.isPending

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="surface-modal max-w-2xl max-h-[85vh] overflow-y-auto scrollbar-custom" style={{ color: 'var(--text-primary)' }}>
                <DialogHeader>
                    <DialogTitle>
                        {t('alert.historyDetailTitle', {
                            defaultValue: 'Alert details',
                        })}
                    </DialogTitle>
                    <DialogDescription className="text-muted-foreground">
                        {t('alert.historyDetailDescription', {
                            defaultValue:
                                'View alert details and optionally acknowledge it.',
                        })}
                    </DialogDescription>
                </DialogHeader>

                {history && (
                    <div className="space-y-4 text-sm">
                        <div>
                            <div className="font-semibold">
                                {t('alert.ruleName')}
                            </div>
                            <div className="text-muted-foreground">
                                {history.ruleName}
                            </div>
                        </div>

                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                            <div>
                                <div className="font-semibold">
                                    {t('alert.severityLabel', {
                                        defaultValue: 'Severity',
                                    })}
                                </div>
                                <div className="text-muted-foreground">
                                    {t(`alert.severity.${history.severity}`)}
                                </div>
                            </div>
                            <div>
                                <div className="font-semibold">
                                    {t('alert.status')}
                                </div>
                                <div className="text-muted-foreground capitalize">
                                    {history.status}
                                </div>
                            </div>
                            <div>
                                <div className="font-semibold">
                                    {t('alert.triggeredAt', {
                                        defaultValue: 'Triggered at',
                                    })}
                                </div>
                                <div className="text-muted-foreground">
                                    {history.triggeredAt
                                        ? new Date(
                                              history.triggeredAt,
                                          ).toLocaleString()
                                        : '-'}
                                </div>
                            </div>
                            {history.sentAt && (
                                <div>
                                    <div className="font-semibold">
                                        {t('sentAt', { defaultValue: 'Sent at' })}
                                    </div>
                                    <div className="text-muted-foreground">
                                        {new Date(history.sentAt).toLocaleString()}
                                    </div>
                                </div>
                            )}
                        </div>

                        <div>
                            <div className="font-semibold">
                                {t('message', { defaultValue: 'Message' })}
                            </div>
                            <div className="text-muted-foreground whitespace-pre-wrap break-words">
                                {history.message}
                            </div>
                        </div>

                        {history.errorMessage && (
                            <div>
                                <div className="font-semibold">
                                    {t('error.message', {
                                        defaultValue: 'Error message',
                                    })}
                                </div>
                                <div className="text-red-500 whitespace-pre-wrap break-words">
                                    {history.errorMessage}
                                </div>
                            </div>
                        )}

                        <div>
                            <div className="font-semibold">
                                {t('alert.channelIds', {
                                    defaultValue: 'Channels',
                                })}
                            </div>
                            <div className="text-muted-foreground">
                                {history.channels?.join(', ') || '-'}
                            </div>
                        </div>

                        {history.acknowledgedBy && (
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <div>
                                    <div className="font-semibold">
                                        {t('alert.acknowledgedBy', {
                                            defaultValue: 'Acknowledged by',
                                        })}
                                    </div>
                                    <div className="text-muted-foreground">
                                        {history.acknowledgedBy}
                                    </div>
                                </div>
                                <div>
                                    <div className="font-semibold">
                                        {t('alert.acknowledgedAt', {
                                            defaultValue: 'Acknowledged at',
                                        })}
                                    </div>
                                    <div className="text-muted-foreground">
                                        {history.acknowledgedAt
                                            ? new Date(
                                                  history.acknowledgedAt,
                                              ).toLocaleString()
                                            : '-'}
                                    </div>
                                </div>
                            </div>
                        )}

                        <div>
                            <div className="font-semibold mb-1">
                                {t('alert.acknowledgeNote', {
                                    defaultValue: 'Acknowledge note',
                                })}
                            </div>
                            <Textarea
                                rows={3}
                                value={note}
                                onChange={(e) => setNote(e.target.value)}
                                placeholder={t(
                                    'alert.acknowledgeNotePlaceholder',
                                    {
                                        defaultValue:
                                            'Optional note when acknowledging this alert...',
                                    },
                                )}
                            />
                        </div>
                    </div>
                )}

                <DialogFooter className="mt-4">
                    <Button
                        variant="outline"
                        onClick={() => onOpenChange(false)}
                        disabled={acknowledgeMutation.isPending}
                    >
                        {t('alert.deleteDialog.cancel')}
                    </Button>
                    <Button
                        onClick={() => history && acknowledgeMutation.mutate(history.id)}
                        disabled={!canAcknowledge}
                    >
                        {acknowledgeMutation.isPending && (
                            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                        )}
                        {t('alert.acknowledge', {
                            defaultValue: 'Acknowledge',
                        })}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    )
}

