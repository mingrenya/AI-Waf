import { StrictMode, Suspense } from 'react'
import { createRoot } from 'react-dom/client'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ReactQueryDevtools } from '@tanstack/react-query-devtools'
import './index.css'
import './i18n'
import App from './App.tsx'
import { ErrorBoundary } from './handler/error-boundary.tsx'
import { ENV } from './utils/env.ts'
import { Toaster } from './components/ui/toaster.tsx'
import { ConstantCategory } from './constant/index.ts'
import { getConstant } from './constant/index.ts'
import { I18nextProvider } from 'react-i18next'
import i18n from './i18n'
import { LoadingFallback } from './components/common/loading-fallback.tsx'
import { ThemeProvider } from './provider/theme-provider'

const queryClient = new QueryClient({
    defaultOptions: {
        queries: {
            staleTime: getConstant(ConstantCategory.FEATURE, 'QUERY_STALE_TIME', 5 * 60 * 1000),
            retry: getConstant(ConstantCategory.FEATURE, 'DEFAULT_QUERY_RETRY', 1),
            refetchOnWindowFocus: false
        }
    }
})

createRoot(document.getElementById('root')!).render(
    <StrictMode>
        <ErrorBoundary>
            <ThemeProvider defaultTheme="system" storageKey="waf-ui-theme">
                <QueryClientProvider client={queryClient}>
                    <I18nextProvider i18n={i18n}>
                        <Suspense fallback={<LoadingFallback />}>
                            <App />
                            <Toaster />
                        </Suspense>
                    </I18nextProvider>
                    {ENV.isDevelopment && <ReactQueryDevtools />}
                </QueryClientProvider>
            </ThemeProvider>
        </ErrorBoundary>
    </StrictMode>,
)

// React 挂载后采用更平滑的加载策略：最短展示时长 + 超时兜底
const APP_LOADING_MIN_VISIBLE_MS = 700
const APP_LOADING_FADE_OUT_MS = 500
const APP_LOADING_MAX_LIFETIME_MS = 5000

const safelyRemoveAppLoading = () => {
    const loadingElement = document.querySelector('.app-loading')
    if (!loadingElement) return

    loadingElement.classList.add('app-loading-fade-out')
    setTimeout(() => loadingElement.remove(), APP_LOADING_FADE_OUT_MS)
}

const appLoadingStart = Number((window as Window & { __APP_LOADING_START__?: number }).__APP_LOADING_START__ || Date.now())
const elapsed = Date.now() - appLoadingStart
const remainingVisibleTime = Math.max(0, APP_LOADING_MIN_VISIBLE_MS - elapsed)

setTimeout(() => {
    requestAnimationFrame(() => {
        safelyRemoveAppLoading()
    })
}, remainingVisibleTime)

setTimeout(() => {
    safelyRemoveAppLoading()
}, APP_LOADING_MAX_LIFETIME_MS)
