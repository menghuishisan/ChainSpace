import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button, Layout, message } from 'antd'
import { MenuFoldOutlined, MenuUnfoldOutlined } from '@ant-design/icons'
import { Link } from 'lucide-react'

import { useAccountSecurity, useNotificationCenter } from '@/hooks'
import { useAppStore, useUserStore } from '@/store'
import ChangePasswordModal from './ChangePasswordModal'
import HeaderNotificationCenter from './HeaderNotificationCenter'
import HeaderUserMenu from './HeaderUserMenu'

const { Header: AntHeader } = Layout

export default function Header() {
  const navigate = useNavigate()
  const { user, logout } = useUserStore()
  const { sidebarCollapsed, toggleSidebar } = useAppStore()
  const { refreshUnreadCount } = useNotificationCenter()
  const { updatePassword } = useAccountSecurity()

  const [passwordModalVisible, setPasswordModalVisible] = useState(false)
  const [passwordLoading, setPasswordLoading] = useState(false)

  useEffect(() => {
    void refreshUnreadCount()
  }, [refreshUnreadCount])

  const handleChangePassword = async (values: { old_password: string; new_password: string }) => {
    setPasswordLoading(true)
    try {
      await updatePassword(values.old_password, values.new_password)
      message.success('密码修改成功')
      setPasswordModalVisible(false)
    } finally {
      setPasswordLoading(false)
    }
  }

  return (
    <>
      <AntHeader
        className="topbar-accent flex items-center justify-between bg-white px-4"
        style={{
          height: 64,
          padding: '0 24px',
          boxShadow: '0 6px 20px rgba(0, 0, 0, 0.06)',
          position: 'sticky',
          top: 0,
          zIndex: 100,
        }}
      >
        <div className="flex items-center">
          <Button
            type="text"
            icon={sidebarCollapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
            onClick={toggleSidebar}
            className="mr-4"
          />
          <div className="flex cursor-pointer items-center" onClick={() => navigate('/')}>
            <Link className="mr-2 h-6 w-6 text-primary" />
            <span className="hide-on-mobile text-lg font-semibold text-text-primary">
              链境
            </span>
          </div>
        </div>

        <div className="flex items-center">
          <HeaderNotificationCenter />
          <HeaderUserMenu
            user={user}
            onOpenPasswordModal={() => setPasswordModalVisible(true)}
            onLogout={async () => {
              await logout()
              navigate('/login')
            }}
          />
        </div>
      </AntHeader>

      <ChangePasswordModal
        open={passwordModalVisible}
        loading={passwordLoading}
        onCancel={() => setPasswordModalVisible(false)}
        onSubmit={handleChangePassword}
      />
    </>
  )
}
