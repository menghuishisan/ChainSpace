import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

function matchAntdGroup(id: string): string | undefined {
  // 按组件能力拆分 Ant Design 生态，避免整套 UI 依赖落到单个超大 chunk。
  if (id.includes('@ant-design/icons')) {
    return 'vendor-antd-icons'
  }

  if (
    id.includes('@ant-design/cssinjs') ||
    id.includes('@ant-design/cssinjs-utils') ||
    id.includes('@ant-design/colors') ||
    id.includes('@ant-design/fast-color') ||
    id.includes('rc-util')
  ) {
    return 'vendor-antd-style'
  }

  if (
    id.includes('rc-table') ||
    id.includes('rc-pagination') ||
    id.includes('rc-resize-observer') ||
    id.includes('rc-virtual-list')
  ) {
    return 'vendor-antd-table'
  }

  if (
    id.includes('rc-field-form') ||
    id.includes('rc-select') ||
    id.includes('rc-tree-select') ||
    id.includes('rc-cascader') ||
    id.includes('rc-upload') ||
    id.includes('rc-input') ||
    id.includes('rc-textarea') ||
    id.includes('rc-mentions')
  ) {
    return 'vendor-antd-form'
  }

  if (
    id.includes('rc-picker') ||
    id.includes('rc-time-picker') ||
    id.includes('dayjs')
  ) {
    return 'vendor-antd-picker'
  }

  if (
    id.includes('rc-dialog') ||
    id.includes('rc-drawer') ||
    id.includes('rc-notification') ||
    id.includes('rc-tooltip') ||
    id.includes('rc-trigger') ||
    id.includes('rc-motion') ||
    id.includes('rc-dropdown') ||
    id.includes('rc-menu')
  ) {
    return 'vendor-antd-feedback'
  }

  if (
    id.includes('antd/es/layout') ||
    id.includes('antd/es/grid') ||
    id.includes('antd/es/space') ||
    id.includes('antd/es/flex') ||
    id.includes('antd/es/splitter') ||
    id.includes('antd/es/menu') ||
    id.includes('antd/es/dropdown') ||
    id.includes('antd/es/breadcrumb') ||
    id.includes('antd/es/tabs') ||
    id.includes('antd/es/steps') ||
    id.includes('antd/es/anchor')
  ) {
    return 'vendor-antd-layout'
  }

  if (
    id.includes('antd/es/table') ||
    id.includes('antd/es/list') ||
    id.includes('antd/es/descriptions') ||
    id.includes('antd/es/statistic')
  ) {
    return 'vendor-antd-data'
  }

  if (
    id.includes('antd/es/form') ||
    id.includes('antd/es/input') ||
    id.includes('antd/es/input-number') ||
    id.includes('antd/es/select') ||
    id.includes('antd/es/tree-select') ||
    id.includes('antd/es/cascader') ||
    id.includes('antd/es/radio') ||
    id.includes('antd/es/checkbox') ||
    id.includes('antd/es/switch') ||
    id.includes('antd/es/slider') ||
    id.includes('antd/es/rate') ||
    id.includes('antd/es/upload') ||
    id.includes('antd/es/transfer') ||
    id.includes('antd/es/auto-complete') ||
    id.includes('antd/es/mentions')
  ) {
    return 'vendor-antd-entry'
  }

  if (
    id.includes('antd/es/button') ||
    id.includes('antd/es/card') ||
    id.includes('antd/es/tag') ||
    id.includes('antd/es/badge') ||
    id.includes('antd/es/avatar') ||
    id.includes('antd/es/typography') ||
    id.includes('antd/es/empty') ||
    id.includes('antd/es/divider') ||
    id.includes('antd/es/progress') ||
    id.includes('antd/es/result') ||
    id.includes('antd/es/skeleton') ||
    id.includes('antd/es/image')
  ) {
    return 'vendor-antd-display'
  }

  if (id.includes('antd') || id.includes('@ant-design') || id.includes('rc-')) {
    return 'vendor-antd-core'
  }

  return undefined
}

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

          if (id.includes('react-router')) {
            return 'vendor-router'
          }

          const antdGroup = matchAntdGroup(id)
          if (antdGroup) {
            return antdGroup
          }

          if (id.includes('react') || id.includes('scheduler')) {
            return 'vendor-react'
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
