import { useMemo } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { Layout, Menu } from 'antd'
import type { MenuProps } from 'antd'

import { useAppStore, useUserStore } from '@/store'
import { getSidebarMenuItems, getSidebarSelection } from './sidebarMenuConfig'

const { Sider } = Layout

export default function Sidebar() {
  const navigate = useNavigate()
  const location = useLocation()
  const { user } = useUserStore()
  const { sidebarCollapsed } = useAppStore()

  const menuItems = useMemo(() => (
    user ? getSidebarMenuItems(user.role) : []
  ), [user])

  const { selectedKeys, openKeys } = useMemo(() => (
    getSidebarSelection(location.pathname, menuItems)
  ), [location.pathname, menuItems])

  const handleMenuClick: MenuProps['onClick'] = ({ key }) => {
    if (key.startsWith('/')) {
      navigate(key)
    }
  }

  return (
    <Sider
      trigger={null}
      collapsible
      collapsed={sidebarCollapsed}
      width={200}
      collapsedWidth={64}
      className="bg-white"
      style={{
        overflow: 'auto',
        height: 'calc(100vh - 64px)',
        position: 'sticky',
        top: 64,
        left: 0,
      }}
    >
      <Menu
        mode="inline"
        selectedKeys={selectedKeys}
        defaultOpenKeys={openKeys}
        items={menuItems}
        onClick={handleMenuClick}
        className="menu-compact px-2"
        style={{ borderRight: 0 }}
      />
    </Sider>
  )
}
