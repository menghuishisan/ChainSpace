/**
 * 用户相关 API
 */
import { get, put } from './request'
import type { User } from '@/types'

/**
 * 获取用户详情
 * 后端路由: GET /users/:id
 */
export function getUser(id: number): Promise<User> {
  return get<User>(`/users/${id}`)
}

/**
 * 获取当前用户信息
 * 后端路由: GET /auth/me
 */
export function getCurrentUser(): Promise<User> {
  return get<User>('/auth/me')
}

/**
 * 更新当前用户信息
 * 后端路由: PUT /auth/me
 */
export function updateCurrentUser(data: {
  real_name?: string
  email?: string
  phone?: string
  avatar?: string
}): Promise<User> {
  return put<User>('/auth/me', data)
}
