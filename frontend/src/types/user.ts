/**
 * 用户相关类型定义
 */

// 用户角色枚举
export type UserRole = 'platform_admin' | 'school_admin' | 'teacher' | 'student'

// 用户状态枚举
export type UserStatus = 'active' | 'disabled'

// 用户信息接口
export interface User {
  id: number
  real_name: string
  role: UserRole
  email?: string
  phone?: string
  avatar?: string
  student_no?: string
  class_id?: number
  class_name?: string
  school_id?: number
  school_name?: string
  status: UserStatus
  must_change_pwd?: boolean
  last_login_at?: string
  created_at: string
}

// 登录响应
export interface LoginResponse {
  access_token: string
  access_token_expires_at: string
  refresh_token: string
  refresh_token_expires_at: string
  user: User
}

// 刷新Token响应
export interface RefreshTokenResponse {
  access_token: string
  access_token_expires_at: string
  refresh_token: string
  refresh_token_expires_at: string
}

// 角色显示名称映射
export const RoleNameMap: Record<UserRole, string> = {
  platform_admin: '平台管理员',
  school_admin: '学校管理员',
  teacher: '教师',
  student: '学生',
}

// 角色对应的默认路由
export const RoleDefaultRoute: Record<UserRole, string> = {
  platform_admin: '/admin/schools',
  school_admin: '/school/teachers',
  teacher: '/teacher/courses',
  student: '/student/courses',
}
