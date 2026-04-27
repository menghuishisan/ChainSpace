import { Badge, Button, Descriptions, Drawer, List, Modal, Space, Tag, message } from 'antd'
import { BellOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'

import { useNotificationCenter } from '@/hooks'
import type { Notification } from '@/types'
import { formatRelativeTime } from '@/utils/format'

export default function HeaderNotificationCenter() {
  const navigate = useNavigate()
  const {
    unreadNotificationCount,
    drawerOpen,
    notifications,
    loading,
    detailOpen,
    detail,
    selectedIds,
    deleteLoading,
    openDrawer,
    closeDrawer,
    openDetail,
    closeDetail,
    markNotificationRead,
    markEverythingRead,
    removeNotification,
    removeSelectedNotifications,
    toggleSelected,
  } = useNotificationCenter()

  const handleViewDetail = async (notification: Notification) => {
    if (notification.link) {
      await markNotificationRead(notification)
      navigate(notification.link)
      return
    }

    await openDetail(notification)
  }

  const handleDelete = (notification: Notification) => {
    Modal.confirm({
      title: '确认删除',
      content: '确定要删除这条通知吗？',
      okText: '确定',
      cancelText: '取消',
      okButtonProps: { danger: true },
      onOk: async () => {
        await removeNotification(notification)
        message.success('通知已删除')
      },
    })
  }

  const handleBatchDelete = async () => {
    if (selectedIds.length === 0) {
      message.warning('请先选择要删除的通知')
      return
    }

    Modal.confirm({
      title: '确认批量删除',
      content: `确定要删除选中的 ${selectedIds.length} 条通知吗？`,
      okText: '确定',
      cancelText: '取消',
      okButtonProps: { danger: true },
      onOk: async () => {
        const removedCount = await removeSelectedNotifications()
        message.success(`已删除 ${removedCount} 条通知`)
      },
    })
  }

  return (
    <>
      <Badge count={unreadNotificationCount} size="small" offset={[-2, 2]}>
        <Button
          type="text"
          icon={<BellOutlined className="text-lg" />}
          shape="circle"
          className="mr-4 hover:bg-[#f5f5f5]"
          onClick={() => void openDrawer()}
        />
      </Badge>

      <Drawer
        title={(
          <div className="flex items-center justify-between">
            <span>通知中心</span>
            <Space size="small">
              {selectedIds.length > 0 && (
                <Button type="primary" danger size="small" loading={deleteLoading} onClick={() => void handleBatchDelete()}>
                  删除 ({selectedIds.length})
                </Button>
              )}
              {notifications.some((item) => !item.is_read) && (
                <Button type="link" size="small" onClick={() => void markEverythingRead().then(() => message.success('已全部标记为已读'))}>
                  全部已读
                </Button>
              )}
            </Space>
          </div>
        )}
        open={drawerOpen}
        onClose={closeDrawer}
        width={420}
      >
        <List
          loading={loading}
          dataSource={notifications}
          locale={{ emptyText: '暂无通知' }}
          renderItem={(item) => (
            <List.Item
              className={`cursor-pointer hover:bg-gray-50 ${selectedIds.includes(item.id) ? 'bg-blue-50' : ''}`}
              onClick={(event) => {
                const target = event.target as HTMLElement
                if (target.closest('.notification-checkbox')) {
                  return
                }
                void handleViewDetail(item)
              }}
              extra={(
                <div className="flex items-center gap-2">
                  <input
                    type="checkbox"
                    className="notification-checkbox h-4 w-4 cursor-pointer"
                    checked={selectedIds.includes(item.id)}
                    onChange={() => toggleSelected(item.id)}
                    onClick={(event) => event.stopPropagation()}
                  />
                  <Button
                    type="text"
                    size="small"
                    danger
                    onClick={(event) => {
                      event.stopPropagation()
                      handleDelete(item)
                    }}
                  >
                    删除
                  </Button>
                </div>
              )}
            >
              <List.Item.Meta
                avatar={<div className={`mt-2 h-2 w-2 rounded-full ${!item.is_read ? 'bg-blue-500' : 'bg-gray-300'}`} />}
                title={(
                  <div className="flex items-center justify-between">
                    <span className={!item.is_read ? 'font-semibold' : ''}>{item.title}</span>
                    {!item.is_read && <Tag color="blue" className="ml-2">未读</Tag>}
                  </div>
                )}
                description={(
                  <div>
                    <div className="mb-1 line-clamp-2 text-sm text-text-secondary">{item.content}</div>
                    <div className="text-xs text-text-secondary">{formatRelativeTime(item.created_at)}</div>
                  </div>
                )}
              />
            </List.Item>
          )}
        />
      </Drawer>

      <Modal
        title={detail?.title || '通知详情'}
        open={detailOpen}
        onCancel={closeDetail}
        footer={[
          <Button key="close" onClick={closeDetail}>关闭</Button>,
          detail?.link ? (
            <Button
              key="goto"
              type="primary"
              onClick={() => {
                closeDetail()
                navigate(detail.link!)
              }}
            >
              查看详情
            </Button>
          ) : null,
        ]}
      >
        {detail && (
          <div className="space-y-4">
            <Descriptions column={1} size="small" bordered>
              <Descriptions.Item label="类型">
                <Tag color="blue">{detail.type}</Tag>
              </Descriptions.Item>
              <Descriptions.Item label="发送者">
                {detail.sender_name || '系统'}
              </Descriptions.Item>
              <Descriptions.Item label="发送时间">
                {detail.created_at}
              </Descriptions.Item>
              <Descriptions.Item label="状态">
                <Tag color={detail.is_read ? 'default' : 'blue'}>
                  {detail.is_read ? '已读' : '未读'}
                </Tag>
              </Descriptions.Item>
            </Descriptions>
            <div className="rounded-md bg-gray-50 p-4 text-sm leading-6 text-text-primary">
              {detail.content}
            </div>
          </div>
        )}
      </Modal>
    </>
  )
}
