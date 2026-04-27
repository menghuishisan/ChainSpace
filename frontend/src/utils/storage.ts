/**
 * 本地存储工具函数
 */
import { STORAGE_KEYS } from './constants'

/**
 * 获取存储项
 */
export function getStorageItem<T>(key: string): T | null {
  try {
    const item = localStorage.getItem(key)
    if (item === null) return null
    return JSON.parse(item) as T
  } catch {
    return null
  }
}

/**
 * 设置存储项
 */
export function setStorageItem<T>(key: string, value: T): void {
  try {
    localStorage.setItem(key, JSON.stringify(value))
  } catch (error) {
    console.error('Storage setItem error:', error)
  }
}

/**
 * 移除存储项
 */
export function removeStorageItem(key: string): void {
  try {
    localStorage.removeItem(key)
  } catch (error) {
    console.error('Storage removeItem error:', error)
  }
}

/**
 * 清除所有存储
 */
export function clearStorage(): void {
  try {
    localStorage.clear()
  } catch (error) {
    console.error('Storage clear error:', error)
  }
}

// ====== Token 存储操作 ======

/**
 * 获取 Access Token
 */
export function getAccessToken(): string | null {
  return localStorage.getItem(STORAGE_KEYS.ACCESS_TOKEN)
}

/**
 * 设置 Access Token
 */
export function setAccessToken(token: string): void {
  localStorage.setItem(STORAGE_KEYS.ACCESS_TOKEN, token)
}

/**
 * 获取 Refresh Token
 */
export function getRefreshToken(): string | null {
  return localStorage.getItem(STORAGE_KEYS.REFRESH_TOKEN)
}

/**
 * 设置 Refresh Token
 */
export function setRefreshToken(token: string): void {
  localStorage.setItem(STORAGE_KEYS.REFRESH_TOKEN, token)
}

/**
 * 清除所有 Token
 */
export function clearTokens(): void {
  localStorage.removeItem(STORAGE_KEYS.ACCESS_TOKEN)
  localStorage.removeItem(STORAGE_KEYS.REFRESH_TOKEN)
}

/**
 * 设置登录信息（Token + 用户信息）
 */
export function setLoginInfo(
  accessToken: string,
  refreshToken: string,
  userInfo: unknown
): void {
  setAccessToken(accessToken)
  setRefreshToken(refreshToken)
  setStorageItem(STORAGE_KEYS.USER_INFO, userInfo)
}

/**
 * 清除登录信息
 */
export function clearLoginInfo(): void {
  clearTokens()
  removeStorageItem(STORAGE_KEYS.USER_INFO)
}
