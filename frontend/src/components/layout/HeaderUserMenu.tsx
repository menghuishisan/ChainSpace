import { Avatar, Dropdown, Modal, Space } from 'antd'
import { LockOutlined, LogoutOutlined, UserOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'

import { RoleNameMap } from '@/types'
import type { User } from '@/types'

interface HeaderUserMenuProps {
  onLogout: () => Promise<void>
  onOpenPasswordModal: () => void
  user?: User | null
}

export default function HeaderUserMenu({
  onLogout,
  onOpenPasswordModal,
  user,
}: HeaderUserMenuProps) {
  const navigate = useNavigate()

  const handleLogout = () => {
    Modal.confirm({
      title: '确认退出',
      content: '确定要退出登录吗？',
      okText: '确定',
      cancelText: '取消',
      onOk: onLogout,
    })
  }

  const items = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: '个人信息',
      onClick: () => navigate('/profile'),
    },
    {
      key: 'password',
      icon: <LockOutlined />,
      label: '修改密码',
      onClick: onOpenPasswordModal,
    },
    { type: 'divider' as const },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      onClick: handleLogout,
    },
  ]

  return (
    <Dropdown menu={{ items }} placement="bottomRight">
      <Space className="cursor-pointer">
        <Avatar src={user?.avatar} icon={!user?.avatar && <UserOutlined />} size="small" />
        <span className="text-sm hide-on-mobile">
          {user?.real_name}
          <span className="ml-1 text-text-secondary">
            ({user ? RoleNameMap[user.role] : ''})
          </span>
        </span>
      </Space>
    </Dropdown>
  )
}
