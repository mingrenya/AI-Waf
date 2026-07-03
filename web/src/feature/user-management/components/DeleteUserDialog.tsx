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

interface DeleteUserDialogProps {
    open: boolean
    onOpenChange: (open: boolean) => void
    username?: string
    submitting?: boolean
    onConfirm: () => Promise<void>
}

export function DeleteUserDialog({ open, onOpenChange, username, submitting = false, onConfirm }: DeleteUserDialogProps) {
    return (
        <AlertDialog open={open} onOpenChange={onOpenChange}>
            <AlertDialogContent className="bg-white/10 backdrop-blur-xl border border-white/20 rounded-2xl">
                <AlertDialogHeader>
                    <AlertDialogTitle>确认删除用户</AlertDialogTitle>
                    <AlertDialogDescription>
                        确定删除用户 {username ? `“${username}”` : ''} 吗？此操作不可撤销。
                    </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                    <AlertDialogCancel disabled={submitting}>取消</AlertDialogCancel>
                    <AlertDialogAction
                        disabled={submitting}
                        onClick={(e) => {
                            e.preventDefault()
                            void onConfirm()
                        }}
                    >
                        {submitting ? '删除中...' : '删除'}
                    </AlertDialogAction>
                </AlertDialogFooter>
            </AlertDialogContent>
        </AlertDialog>
    )
}
