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
import { alertRuleApi } from '@/api/alert'
import { toast } from '@/store'
import { useTranslation } from 'react-i18next'
import { queryKeys } from '@/lib/query-keys'

interface DeleteRuleDialogProps {
    open: boolean
    onOpenChange: (open: boolean) => void
    ruleId: string | null
}

export function DeleteRuleDialog({
    open,
    onOpenChange,
    ruleId,
}: DeleteRuleDialogProps) {
    const { t } = useTranslation()
    const queryClient = useQueryClient()

    const deleteMutation = useMutation({
        mutationFn: (id: string) => alertRuleApi.deleteRule(id),
        onSuccess: () => {
            queryClient.invalidateQueries({
                queryKey: queryKeys.alert.rules.lists(),
            })
            toast({
                title: 'Success',
                description: t('alert.deleteRuleSuccess'),
                variant: 'success',
            })
            onOpenChange(false)
        },
        onError: () => {
            toast({
                title: 'Error',
                description: t('alert.deleteRuleFailed'),
                variant: 'destructive',
            })
        },
    })

    const handleDelete = () => {
        if (ruleId) {
            deleteMutation.mutate(ruleId)
        }
    }

    return (
        <AlertDialog open={open} onOpenChange={onOpenChange}>
            <AlertDialogContent>
                <AlertDialogHeader>
                    <AlertDialogTitle>
                        {t('alert.deleteRuleTitle')}
                    </AlertDialogTitle>
                    <AlertDialogDescription className="dark:text-shadow-glow-white">
                        {t('alert.deleteRuleDescription')}
                    </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                    <AlertDialogCancel>
                        {t('alert.deleteDialog.cancel')}
                    </AlertDialogCancel>
                    <AlertDialogAction
                        onClick={handleDelete}
                        disabled={deleteMutation.isPending}
                        className="bg-red-600 hover:bg-red-700"
                    >
                        {deleteMutation.isPending
                            ? t('alert.deleteDialog.deleting')
                            : t('alert.deleteDialog.delete')}
                    </AlertDialogAction>
                </AlertDialogFooter>
            </AlertDialogContent>
        </AlertDialog>
    )
}

