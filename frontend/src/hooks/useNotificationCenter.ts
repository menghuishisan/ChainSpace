import { useCallback, useState } from 'react'

import {
  batchDeleteNotifications,
  deleteNotification,
  getNotifications,
  getUnreadCount,
  markAllAsRead,
  markAsRead,
} from '@/api/notification'
import { useAppStore } from '@/store'
import type { Notification, PaginatedData } from '@/types'

export function useNotificationCenter() {
  const unreadNotificationCount = useAppStore((state) => state.unreadNotificationCount)
  const setUnreadNotificationCount = useAppStore((state) => state.setUnreadNotificationCount)

  const [drawerOpen, setDrawerOpen] = useState(false)
  const [notifications, setNotifications] = useState<Notification[]>([])
  const [loading, setLoading] = useState(false)
  const [detailOpen, setDetailOpen] = useState(false)
  const [detail, setDetail] = useState<Notification | null>(null)
  const [selectedIds, setSelectedIds] = useState<number[]>([])
  const [deleteLoading, setDeleteLoading] = useState(false)

  const syncReadState = useCallback((ids: number[]) => {
    if (ids.length === 0) {
      return
    }

    const targetIds = new Set(ids)
    const unreadChanged = notifications.filter((item) => targetIds.has(item.id) && !item.is_read).length
    setNotifications((current) => current.map((item) => (
      targetIds.has(item.id) ? { ...item, is_read: true } : item
    )))
    setUnreadNotificationCount(Math.max(0, unreadNotificationCount - unreadChanged))
  }, [notifications, setUnreadNotificationCount, unreadNotificationCount])

  const refreshUnreadCount = useCallback(async () => {
    const { count } = await getUnreadCount()
    setUnreadNotificationCount(count)
  }, [setUnreadNotificationCount])

  const openDrawer = useCallback(async () => {
    setDrawerOpen(true)
    setLoading(true)
    try {
      const result: PaginatedData<Notification> = await getNotifications({ page: 1, page_size: 50 })
      setNotifications(result.list)
    } finally {
      setLoading(false)
    }
  }, [])

  const closeDrawer = useCallback(() => {
    setDrawerOpen(false)
    setSelectedIds([])
  }, [])

  const openDetail = useCallback(async (notification: Notification) => {
    setDetail(notification)
    setDetailOpen(true)
    if (!notification.is_read) {
      await markAsRead([notification.id])
      syncReadState([notification.id])
    }
  }, [syncReadState])

  const closeDetail = useCallback(() => {
    setDetailOpen(false)
  }, [])

  const markNotificationRead = useCallback(async (notification: Notification) => {
    if (notification.is_read) {
      return
    }

    await markAsRead([notification.id])
    syncReadState([notification.id])
  }, [syncReadState])

  const markEverythingRead = useCallback(async () => {
    await markAllAsRead()
    setNotifications((current) => current.map((item) => ({ ...item, is_read: true })))
    setUnreadNotificationCount(0)
  }, [setUnreadNotificationCount])

  const removeNotification = useCallback(async (notification: Notification) => {
    await deleteNotification(notification.id)
    setNotifications((current) => current.filter((item) => item.id !== notification.id))
    if (!notification.is_read) {
      setUnreadNotificationCount(Math.max(0, unreadNotificationCount - 1))
    }
    setSelectedIds((current) => current.filter((id) => id !== notification.id))
  }, [setUnreadNotificationCount, unreadNotificationCount])

  const removeSelectedNotifications = useCallback(async () => {
    if (selectedIds.length === 0) {
      return 0
    }

    setDeleteLoading(true)
    try {
      await batchDeleteNotifications(selectedIds)
      const deletedIds = new Set(selectedIds)
      const unreadChanged = notifications.filter((item) => deletedIds.has(item.id) && !item.is_read).length
      setNotifications((current) => current.filter((item) => !deletedIds.has(item.id)))
      setUnreadNotificationCount(Math.max(0, unreadNotificationCount - unreadChanged))
      const removedCount = deletedIds.size
      setSelectedIds([])
      return removedCount
    } finally {
      setDeleteLoading(false)
    }
  }, [notifications, selectedIds, setUnreadNotificationCount, unreadNotificationCount])

  const toggleSelected = useCallback((id: number) => {
    setSelectedIds((current) => (
      current.includes(id) ? current.filter((item) => item !== id) : [...current, id]
    ))
  }, [])

  return {
    unreadNotificationCount,
    drawerOpen,
    notifications,
    loading,
    detailOpen,
    detail,
    selectedIds,
    deleteLoading,
    refreshUnreadCount,
    openDrawer,
    closeDrawer,
    openDetail,
    closeDetail,
    markNotificationRead,
    markEverythingRead,
    removeNotification,
    removeSelectedNotifications,
    toggleSelected,
  }
}
