import { type RouteObject } from "react-router"
import { Navigate } from "react-router"
import { Suspense, lazy, ReactElement } from "react"
import { RoutePath, ROUTES } from "./constants"
import { useTranslation } from 'react-i18next'
import { TFunction } from 'i18next'
import { ProtectedRoute } from "@/feature/auth/components/ProtectedRoute"

// 直接导入布局组件
import { RootLayout } from "@/components/layout/root-layout"
import { MonitorLayOut } from "@/pages/monitor/layout"
import { RulesLayOut } from "@/pages/rule/layout"
import { SettingLayOut } from "@/pages/setting/layout"
import { LogsLayout } from "@/pages/logs/layout"
import { AlertLayOut } from "@/pages/alert/layout"
import { AIAnalyzerLayOut } from "@/pages/ai-analyzer/layout"

// 直接导入子组件
import GlobalSettingPage from "@/pages/setting/pages/global-setting/page"
import CertificatesPage from "@/pages/setting/pages/certificate/page"
import EventsPage from "@/pages/logs/pages/event/page"
import LogsPage from "@/pages/logs/pages/log/page"
import SiteManagerPage from "@/pages/setting/pages/site/page"
import IPGroupPage from "@/pages/rule/pages/ip-group/page"
import MicroRulePage from "@/pages/rule/pages/micro-rule/page"
import StatsPage from "@/pages/monitor/pages/stats/page"
import ViewerPage from "@/pages/monitor/pages/security-dashboard/page"
import FlowControlPage from "@/pages/rule/pages/cc/page"
import AlertChannelPage from "@/pages/alert/pages/channel/page"
import AlertRulePage from "@/pages/alert/pages/rule/page"
import AlertHistoryPage from "@/pages/alert/pages/history/page"
import SecurityMetricsPage from "@/pages/security-metrics/page"
import AdaptiveThrottlingPage from "@/pages/rule/pages/adaptive-throttling/page"
import PatternsPage from "@/pages/ai-analyzer/pages/patterns/page"
import RulesPage from "@/pages/ai-analyzer/pages/rules/page"
import ConfigPage from "@/pages/ai-analyzer/pages/config/page"
import AIAssistantPage from "@/pages/ai-analyzer/pages/assistant/page"
import { LoadingFallback } from "@/components/common/loading-fallback"

// 懒加载认证页面
const LoginPage = lazy(() => import("@/pages/auth/login"))
const ResetPasswordPage = lazy(() => import("@/pages/auth/reset-password"))

// 懒加载组件包装器
const lazyLoad = (Component: React.ComponentType) => (
    <Suspense fallback={<LoadingFallback />}>
        <Component />
    </Suspense>
)

// 面包屑项类型定义
interface BreadcrumbItem {
    title: string
    path: string
    component: ReactElement
}

interface BreadcrumbConfig {
    defaultPath: string
    items: BreadcrumbItem[]
}

// 创建面包屑配置
export function createBreadcrumbConfig(t: TFunction): Record<RoutePath, BreadcrumbConfig> {
    return {
        [ROUTES.LOGS]: {
            defaultPath: "event",
            items: [
                { title: t('breadcrumb.logs.attack'), path: "event", component: <EventsPage /> },
                { title: t('breadcrumb.logs.protect'), path: "log", component: <LogsPage /> },
            ]
        },
        [ROUTES.MONITOR]: {
            defaultPath: "overview",
            items: [
                { title: t('breadcrumb.monitor.overview'), path: "overview", component: <StatsPage /> },
                { title: t('breadcrumb.monitor.dashboard'), path: "dashboard", component: <Navigate to="/security-dashboard" replace /> },
                { title: t('breadcrumb.monitor.securityMetrics'), path: "security-metrics", component: <SecurityMetricsPage /> },
            ]
        },
        [ROUTES.RULES]: {
            defaultPath: "user",
            items: [
                // { title: t('breadcrumb.rules.system'), path: "system", component: <SysRules /> },
                { title: t('breadcrumb.rules.user'), path: "user", component: <MicroRulePage /> },
                { title: t('breadcrumb.rules.ipGroup'), path: "ip-group", component: <IPGroupPage /> },
                { title: t('breadcrumb.rules.flowControl'), path: "flow-control", component: <FlowControlPage /> },
                { title: t('breadcrumb.rules.adaptiveThrottling'), path: "adaptive-throttling", component: <AdaptiveThrottlingPage /> }
            ]
        },
        [ROUTES.SETTINGS]: {
            defaultPath: "global",
            items: [
                { title: t('breadcrumb.settings.settings'), path: "global", component: <GlobalSettingPage /> },
                { title: t('breadcrumb.settings.siteManager'), path: "site", component: <SiteManagerPage /> },
                { title: t('breadcrumb.settings.certManager'), path: "cert", component: <CertificatesPage /> }
            ]
        },
        [ROUTES.ALERTS]: {
            defaultPath: "channel",
            items: [
                { title: t('breadcrumb.alerts.channel'), path: "channel", component: <AlertChannelPage /> },
                { title: t('breadcrumb.alerts.rule'), path: "rule", component: <AlertRulePage /> },
                { title: t('breadcrumb.alerts.history'), path: "history", component: <AlertHistoryPage /> }
            ]
        },
        [ROUTES.AI_ANALYZER]: {
            defaultPath: "patterns",
            items: [
                { title: t('breadcrumb.aiAnalyzer.patterns'), path: "patterns", component: <PatternsPage /> },
                { title: t('breadcrumb.aiAnalyzer.rules'), path: "rules", component: <RulesPage /> },
                { title: t('breadcrumb.aiAnalyzer.assistant'), path: "assistant", component: <AIAssistantPage /> },
                { title: t('breadcrumb.aiAnalyzer.config'), path: "config", component: <ConfigPage /> }
            ]
        }
    }
}

// 获取当前语言的面包屑配置
export function useBreadcrumbMap() {
    const { t } = useTranslation()
    return createBreadcrumbConfig(t)
}

// 生成子路由配置
function createChildRoutes(config: BreadcrumbConfig): RouteObject[] {
    return [
        {
            path: "",
            element: <Navigate to={config.defaultPath} replace />
        },
        ...config.items.map(item => ({
            path: item.path,
            element: item.component
        }))
    ]
}

// 路由配置
export function useRoutes(): RouteObject[] {
    const breadcrumbMap = useBreadcrumbMap()

    // 认证路由
    const authRoutes: RouteObject[] = [
        { path: "/login", element: lazyLoad(LoginPage) },
        { path: "/reset-password", element: lazyLoad(ResetPasswordPage) }
    ]

    // security-dashboard
    const securityDashboardRoutes: RouteObject[] = [
        { path: "/security-dashboard", element: lazyLoad(ViewerPage) }
    ]

    // 应用路由
    const appRoutes: RouteObject = {
        element: <ProtectedRoute />,
        children: [{
            element: <RootLayout />,
            children: [
                {
                    path: "/",
                    element: <Navigate to={`${ROUTES.MONITOR}/overview`} replace />
                },
                {
                    path: ROUTES.LOGS,
                    element: <LogsLayout />,
                    children: createChildRoutes(breadcrumbMap[ROUTES.LOGS])
                },
                {
                    path: ROUTES.MONITOR,
                    element: <MonitorLayOut />,
                    children: createChildRoutes(breadcrumbMap[ROUTES.MONITOR])
                },
                {
                    path: ROUTES.RULES,
                    element: <RulesLayOut />,
                    children: createChildRoutes(breadcrumbMap[ROUTES.RULES])
                },
                {
                    path: ROUTES.SETTINGS,
                    element: <SettingLayOut />,
                    children: createChildRoutes(breadcrumbMap[ROUTES.SETTINGS])
                },
                {
                    path: ROUTES.ALERTS,
                    element: <AlertLayOut />,
                    children: createChildRoutes(breadcrumbMap[ROUTES.ALERTS])
                },
                {
                    path: ROUTES.AI_ANALYZER,
                    element: <AIAnalyzerLayOut />,
                    children: createChildRoutes(breadcrumbMap[ROUTES.AI_ANALYZER])
                }
            ]
        }]
    }

    return [...authRoutes, appRoutes, ...securityDashboardRoutes]
}

// 默认面包屑配置，用于类型推断
export const breadcrumbMap = createBreadcrumbConfig(((key: string) => key) as unknown as TFunction) as ReturnType<typeof createBreadcrumbConfig>