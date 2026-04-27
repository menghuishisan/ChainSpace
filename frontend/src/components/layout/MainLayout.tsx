/**
 * 主布局组件
 */
import { Outlet } from 'react-router-dom'
import { Layout } from 'antd'
import Header from './Header'
import Sidebar from './Sidebar'

const { Content } = Layout

/**
 * 主布局
 * 包含顶部导航栏、侧边栏菜单和主内容区
 */
export default function MainLayout() {
  return (
    <Layout className="min-h-screen">
      {/* 顶部导航栏 */}
      <Header />
      
      <Layout>
        {/* 侧边栏 */}
        <Sidebar />
        
        {/* 主内容区 */}
        <Content
          className="bg-gradient-main"
          style={{
            padding: 24,
            minHeight: 'calc(100vh - 64px)',
            overflow: 'auto',
          }}
        >
          <div className="fade-in">
            <Outlet />
          </div>
        </Content>
      </Layout>
    </Layout>
  )
}
