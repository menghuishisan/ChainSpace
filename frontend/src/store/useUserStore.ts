/**
 * 用户状态管理
 */
import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { User, UserRole } from '@/types'
import { STORAGE_KEYS } from '@/utils/constants'
import { 
  setAccessToken, 
  setRefreshToken, 
  clearLoginInfo,
  getAccessToken,
  getRefreshToken,
} from '@/utils/storage'
import * as authApi from '@/api/auth'
import * as userApi from '@/api/user'

interface UserState {
  // 用户信息
  user: User | null
  // 是否已登录
  isLoggedIn: boolean
  // 加载状态
  loading: boolean
  
  // 操作方法
  login: (phone: string, password: string) => Promise<void>
  logout: () => Promise<void>
  fetchUser: () => Promise<void>
  updateUser: (data: Partial<User>) => Promise<void>
  setUser: (user: User | null) => void
  
  // 权限检查
  hasRole: (roles: UserRole | UserRole[]) => boolean
  isAdmin: () => boolean
  isSchoolAdmin: () => boolean
  isTeacher: () => boolean
  isStudent: () => boolean
}

export const useUserStore = create<UserState>()(
  persist(
    (set, get) => ({
      user: null,
      isLoggedIn: false,
      loading: false,

      // 登录
      login: async (phone: string, password: string) => {
        set({ loading: true })
        try {
          const response = await authApi.login(phone, password)
          
          // 保存 Token
          setAccessToken(response.access_token)
          setRefreshToken(response.refresh_token)
          
          // 设置用户信息
          set({
            user: response.user,
            isLoggedIn: true,
            loading: false,
          })
        } catch (error) {
          set({ loading: false })
          throw error
        }
      },

      // 登出
      logout: async () => {
        try {
          const refreshToken = getRefreshToken()
          if (refreshToken) {
            await authApi.logout(refreshToken)
          }
        } catch {
          // 忽略登出错误
        } finally {
          clearLoginInfo()
          set({ user: null, isLoggedIn: false })
        }
      },

      // 获取当前用户信息
      fetchUser: async () => {
        // 检查是否有 Token
        const token = getAccessToken()
        if (!token) {
          set({ user: null, isLoggedIn: false })
          return
        }

        set({ loading: true })
        try {
          const user = await userApi.getCurrentUser()
          set({ user, isLoggedIn: true, loading: false })
        } catch {
          // 获取失败，清除登录状态
          clearLoginInfo()
          set({ user: null, isLoggedIn: false, loading: false })
        }
      },

      // 更新用户信息
      updateUser: async (data: Partial<User>) => {
        const currentUser = get().user
        if (!currentUser) throw new Error('用户未登录')
        
        set({ loading: true })
        try {
          const user = await userApi.updateCurrentUser({
            real_name: data.real_name,
            email: data.email,
            phone: data.phone,
            avatar: data.avatar,
          })
          set({ user, loading: false })
        } catch (error) {
          set({ loading: false })
          throw error
        }
      },

      // 设置用户（用于初始化）
      setUser: (user: User | null) => {
        set({ user, isLoggedIn: !!user })
      },

      // 检查是否拥有指定角色
      hasRole: (roles: UserRole | UserRole[]) => {
        const { user } = get()
        if (!user) return false
        
        const roleArray = Array.isArray(roles) ? roles : [roles]
        return roleArray.includes(user.role)
      },

      // 是否是平台管理员
      isAdmin: () => get().hasRole('platform_admin'),

      // 是否是学校管理员
      isSchoolAdmin: () => get().hasRole('school_admin'),

      // 是否是教师
      isTeacher: () => get().hasRole('teacher'),

      // 是否是学生
      isStudent: () => get().hasRole('student'),
    }),
    {
      name: STORAGE_KEYS.USER_INFO,
      partialize: (state) => ({ 
        user: state.user, 
        isLoggedIn: state.isLoggedIn 
      }),
    }
  )
)
