import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react-swc'
import path from 'path'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:2333',
        changeOrigin: true,
        secure: false,
      }
    }
  },
  // SPA fallback: 直接访问子路由时返回 index.html
  appType: 'spa',
})
