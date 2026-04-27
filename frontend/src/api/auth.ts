/**
 * 认证相关 API
 */
import { post, put } from './request'
import type { LoginResponse, RefreshTokenResponse } from '@/types'

/**
 * 用户登录
 * 后端路由: POST /auth/login
 */
export function login(phone: string, password: string): Promise<LoginResponse> {
  return post<LoginResponse>('/auth/login', { phone, password })
}

/**
 * 用户登出
 * 后端路由: POST /auth/logout
 */
export function logout(refresh_token: string): Promise<void> {
  return post('/auth/logout', { refresh_token })
}

/**
 * 修改密码
 * 后端路由: PUT /auth/password
 */
export function changePassword(old_password: string, new_password: string): Promise<void> {
  return put('/auth/password', { old_password, new_password })
}

/**
 * 刷新Token
 * 后端路由: POST /auth/refresh
 */
export function refreshToken(refresh_token: string): Promise<RefreshTokenResponse> {
  return post<RefreshTokenResponse>('/auth/refresh', { refresh_token })
}

/**
 * 重置用户密码（管理员）
 * 后端路由: POST /auth/reset-password
 */
export function resetPassword(userId: number, newPassword: string): Promise<void> {
  return post('/auth/reset-password', { user_id: userId, new_password: newPassword })
}
