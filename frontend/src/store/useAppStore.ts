/**
 * 应用全局状态管理
 */
import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { STORAGE_KEYS } from '@/utils/constants'

interface AppState {
  // 侧边栏折叠状态
  sidebarCollapsed: boolean
  // 主题模式 (预留)
  theme: 'light' | 'dark'
  // 全局加载状态
  globalLoading: boolean
  // 未读通知数
  unreadNotificationCount: number

  // 操作方法
  toggleSidebar: () => void
  setSidebarCollapsed: (collapsed: boolean) => void
  setTheme: (theme: 'light' | 'dark') => void
  setGlobalLoading: (loading: boolean) => void
  setUnreadNotificationCount: (count: number) => void
  incrementUnreadCount: () => void
  decrementUnreadCount: () => void
}

export const useAppStore = create<AppState>()(
  persist(
    (set, get) => ({
      sidebarCollapsed: false,
      theme: 'light',
      globalLoading: false,
      unreadNotificationCount: 0,

      // 切换侧边栏
      toggleSidebar: () => {
        set({ sidebarCollapsed: !get().sidebarCollapsed })
      },

      // 设置侧边栏状态
      setSidebarCollapsed: (collapsed: boolean) => {
        set({ sidebarCollapsed: collapsed })
      },

      // 设置主题
      setTheme: (theme: 'light' | 'dark') => {
        set({ theme })
      },

      // 设置全局加载状态
      setGlobalLoading: (loading: boolean) => {
        set({ globalLoading: loading })
      },

      // 设置未读通知数
      setUnreadNotificationCount: (count: number) => {
        set({ unreadNotificationCount: count })
      },

      // 增加未读数
      incrementUnreadCount: () => {
        set({ unreadNotificationCount: get().unreadNotificationCount + 1 })
      },

      // 减少未读数
      decrementUnreadCount: () => {
        const current = get().unreadNotificationCount
        set({ unreadNotificationCount: Math.max(0, current - 1) })
      },
    }),
    {
      name: STORAGE_KEYS.SIDEBAR_COLLAPSED,
      partialize: (state) => ({ 
        sidebarCollapsed: state.sidebarCollapsed,
        theme: state.theme,
      }),
    }
  )
)
