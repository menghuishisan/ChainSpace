import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  build: {
    chunkSizeWarningLimit: 700,
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (!id.includes('node_modules')) {
            return undefined
          }

          // Keep the runtime-critical libraries together. The previous,
          // aggressive vendor splitting caused circular runtime dependencies
          // in production bundles on the deployed build.
          if (
            id.includes('react') ||
            id.includes('react-dom') ||
            id.includes('scheduler') ||
            id.includes('antd') ||
            id.includes('@ant-design') ||
            id.includes('rc-') ||
            id.includes('dayjs')
          ) {
            return 'vendor-ui'
          }

          if (id.includes('react-router')) {
            return 'vendor-router'
          }

          if (id.includes('echarts')) {
            return 'vendor-chart'
          }

          if (id.includes('@monaco-editor')) {
            return 'vendor-editor'
          }

          if (id.includes('dompurify')) {
            return 'vendor-sanitize'
          }

          return 'vendor-misc'
        },
      },
    },
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:3000',
        changeOrigin: true,
        ws: true,
      },
      '/ws': {
        target: 'ws://localhost:3000',
        ws: true,
      },
    },
  },
})
