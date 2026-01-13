import { Outlet } from "react-router"
import { SubNav, type NavItem } from "@/components/layout/sub-nav"
import { Bell, Radio, History } from "lucide-react"
import { useTranslation } from "react-i18next"

export function AlertLayOut() {
    const { t } = useTranslation()

    const navItems: NavItem[] = [
        {
            title: t('alert.channels'),
            path: 'channel',
            icon: <Radio className="h-4 w-4" />
        },
        {
            title: t('alert.rules'),
            path: 'rule',
            icon: <Bell className="h-4 w-4" />
        },
        {
            title: t('alert.history'),
            path: 'history',
            icon: <History className="h-4 w-4" />
        }
    ]

    return (
        <div className="flex flex-col h-full">
            <SubNav items={navItems} />
            <div className="flex-1 overflow-auto">
                <Outlet />
            </div>
        </div>
    )
}
