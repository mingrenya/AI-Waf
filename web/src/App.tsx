import { AppRouter } from "@/routes"
import { useEffect } from 'react'
import { useThemeStore } from '@/store'

function App() {
    // 初始化主题
    useEffect(() => {
        useThemeStore.getState().initTheme()
    }, [])

    return <AppRouter />
}

export default App
