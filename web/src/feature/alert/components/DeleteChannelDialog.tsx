import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { alertChannelApi } from '@/api/alert'
import { useToast } from '@/hooks/use-toast'
import { useTranslation } from 'react-i18next'

interface DeleteChannelDialogProps {
    open: boolean
    onOpenChange: (open: boolean) => void
    channelId: string | null
}

export function DeleteChannelDialog({ open, onOpenChange, channelId }: DeleteChannelDialogProps) {
    const { t } = useTranslation()
    const { toast } = useToast()
    const queryClient = useQueryClient()

    const deleteMutation = useMutation({
        mutationFn: (id: string) => alertChannelApi.deleteChannel(id),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['alertChannels'] })
            toast({ title: 'Success', description: t('alert.deleteChannelSuccess') })
            onOpenChange(false)
        },
        onError: () => {
            toast({ title: 'Error', description: t('alert.deleteChannelFailed'), variant: 'destructive' })
        }
    })

    const handleDelete = () => {
        if (channelId) {
            deleteMutation.mutate(channelId)
        }
    }

    return (
        <AlertDialog open={open} onOpenChange={onOpenChange}>
            <AlertDialogContent>
                <AlertDialogHeader>
                    <AlertDialogTitle>{t('alert.deleteChannelTitle')}</AlertDialogTitle>
                    <AlertDialogDescription className='dark:text-shadow-glow-white'>
                        {t('alert.deleteChannelDescription')}
                    </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                    <AlertDialogCancel>{t('alert.deleteDialog.cancel')}</AlertDialogCancel>
                    <AlertDialogAction
                        onClick={handleDelete}
                        disabled={deleteMutation.isPending}
                        className="bg-red-600 hover:bg-red-700"
                    >
                        {deleteMutation.isPending ? t('alert.deleteDialog.deleting') : t('alert.deleteDialog.delete')}
                    </AlertDialogAction>
                </AlertDialogFooter>
            </AlertDialogContent>
        </AlertDialog>
    )
}
