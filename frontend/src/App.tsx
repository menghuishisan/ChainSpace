/**
 * 链境 ChainSpace 应用根组件
 * 配置路由和全局状态
 */
import { BrowserRouter } from 'react-router-dom'
import { Suspense } from 'react'
import { Spin, App as AntdApp } from 'antd'

import AntdAppBridge from '@/components/common/AntdAppBridge'
import AppRouter from './router'

// 全局加载状态
const GlobalLoading = () => (
  <div className="flex items-center justify-center h-screen bg-gradient-main">
    <Spin size="large" tip="加载中..."><div /></Spin>
  </div>
)

function App() {
  return (
    <BrowserRouter>
      <AntdApp>
        <AntdAppBridge />
        <Suspense fallback={<GlobalLoading />}>
          <AppRouter />
        </Suspense>
      </AntdApp>
    </BrowserRouter>
  )
}

export default App
