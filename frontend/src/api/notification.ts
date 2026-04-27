/**
 * 通知模块 API
 * 后端路由: /notifications/*
 */
import { get, post, del } from './request'
import type { PaginatedData, Notification } from '@/types'

/**
 * 获取通知列表
 * 后端路由: GET /notifications
 */
export function getNotifications(params?: {
  page?: number
  page_size?: number
  type?: string
  is_read?: boolean
}): Promise<PaginatedData<Notification>> {
  return get<PaginatedData<Notification>>('/notifications', params)
}

/**
 * 获取未读通知数量
 * 后端路由: GET /notifications/unread-count
 */
export function getUnreadCount(): Promise<{ count: number }> {
  return get<{ count: number }>('/notifications/unread-count')
}

/**
 * 标记通知为已读
 * 后端路由: POST /notifications/read
 */
export function markAsRead(notificationIds: number[]): Promise<void> {
  return post('/notifications/read', { notification_ids: notificationIds })
}

/**
 * 标记全部通知为已读
 * 后端路由: POST /notifications/read-all
 */
export function markAllAsRead(): Promise<void> {
  return post('/notifications/read-all')
}

/**
 * 删除通知
 * 后端路由: DELETE /notifications/:id
 */
export function deleteNotification(id: number): Promise<void> {
  return del(`/notifications/${id}`)
}

/**
 * 批量删除通知
 * 后端路由: POST /notifications/batch-delete
 */
export function batchDeleteNotifications(notificationIds: number[]): Promise<void> {
  return post('/notifications/batch-delete', { ids: notificationIds })
}

/**
 * 发送通知（管理员）
 * 后端路由: POST /notifications/send
 */
export function sendNotification(data: {
  user_ids: number[]
  title: string
  content: string
  type: string
}): Promise<void> {
  return post('/notifications/send', data)
}

/**
 * 广播通知（管理员）
 * 后端路由: POST /notifications/broadcast
 */
export function broadcastNotification(data: {
  title: string
  content: string
  type: string
  target_roles?: string[]
}): Promise<void> {
  return post('/notifications/broadcast', data)
}
