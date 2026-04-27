/**
 * Axios 请求封装
 * 支持双Token无感刷新机制
 */
import axios, { AxiosInstance, AxiosRequestConfig, AxiosError, InternalAxiosRequestConfig } from 'axios'
import { message } from 'antd'
import { API_BASE_URL } from '@/utils/constants'
import { 
  getAccessToken, 
  getRefreshToken, 
  setAccessToken, 
  setRefreshToken, 
  clearLoginInfo 
} from '@/utils/storage'
import type { ApiResponse, RefreshTokenResponse } from '@/types'

declare module 'axios' {
  interface AxiosRequestConfig {
    _silent?: boolean
  }

  interface InternalAxiosRequestConfig {
    _retry?: boolean
    _silent?: boolean
  }
}

export interface RequestConfig extends AxiosRequestConfig {
  _silent?: boolean
}

// 创建 axios 实例
const request: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 是否正在刷新 Token
let isRefreshing = false

// 等待刷新的请求队列
let refreshSubscribers: Array<(token: string | null) => void> = []

const tokenRefreshBufferMs = 30 * 1000

function decodeJwtPayload(token: string): Record<string, unknown> | null {
  const parts = token.split('.')
  if (parts.length < 2) {
    return null
  }

  try {
    const base64 = parts[1].replace(/-/g, '+').replace(/_/g, '/')
    const padded = base64.padEnd(Math.ceil(base64.length / 4) * 4, '=')
    return JSON.parse(window.atob(padded)) as Record<string, unknown>
  } catch {
    return null
  }
}

function shouldRefreshAccessToken(token: string): boolean {
  const payload = decodeJwtPayload(token)
  const exp = payload?.exp
  if (typeof exp !== 'number') {
    return false
  }

  const expiresAt = exp * 1000
  return expiresAt - Date.now() <= tokenRefreshBufferMs
}

/**
 * 将请求添加到等待队列
 */
function subscribeTokenRefresh(callback: (token: string | null) => void) {
  refreshSubscribers.push(callback)
}

/**
 * 执行队列中的请求
 */
function onTokenRefreshFinished(newToken: string | null) {
  refreshSubscribers.forEach(callback => callback(newToken))
  refreshSubscribers = []
}

function waitForRefreshedToken(): Promise<string> {
  return new Promise((resolve, reject) => {
    subscribeTokenRefresh((newToken: string | null) => {
      if (newToken) {
        resolve(newToken)
        return
      }
      reject(new Error('refresh token failed'))
    })
  })
}

/**
 * 刷新 Token
 */
async function refreshToken(): Promise<string | null> {
  const refresh_token = getRefreshToken()
  if (!refresh_token) {
    return null
  }

  try {
    // 使用新的 axios 实例发送刷新请求，避免循环
    const response = await axios.post<ApiResponse<RefreshTokenResponse>>(
      `${API_BASE_URL}/auth/refresh`,
      { refresh_token },
      { headers: { 'Content-Type': 'application/json' } }
    )

    if (response.data.code === 0) {
      const { access_token, refresh_token: newRefreshToken } = response.data.data
      setAccessToken(access_token)
      setRefreshToken(newRefreshToken)
      return access_token
    }
    return null
  } catch {
    return null
  }
}

// 请求拦截器
request.interceptors.request.use(
  async (config: InternalAxiosRequestConfig) => {
    // 获取 Token 并添加到请求头
    let token = getAccessToken()
    const isRefreshRequest = typeof config.url === 'string' && config.url.includes('/auth/refresh')

    if (token && !isRefreshRequest && shouldRefreshAccessToken(token)) {
      if (isRefreshing) {
        token = await waitForRefreshedToken()
      } else {
        isRefreshing = true
        try {
          const newToken = await refreshToken()
          if (newToken) {
            onTokenRefreshFinished(newToken)
            token = newToken
          } else {
            onTokenRefreshFinished(null)
          }
        } finally {
          isRefreshing = false
        }
      }
    }

    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error: AxiosError) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  (response) => {
    const data: ApiResponse = response.data
    
    // 业务错误处理
    if (data.code !== 0) {
      // 显示错误信息
      message.error(data.message || '请求失败')
      return Promise.reject(new Error(data.message || '请求失败'))
    }
    
    return response
  },
  async (error: AxiosError<ApiResponse>) => {
    const originalRequest = error.config
    if (!originalRequest) {
      return Promise.reject(error)
    }
    
    // Token 过期处理 (401)
    if (error.response?.status === 401 && !originalRequest._retry) {
      // 标记请求已重试
      originalRequest._retry = true

      // 如果正在刷新，则将请求加入队列
      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          subscribeTokenRefresh((newToken: string | null) => {
            if (!newToken) {
              reject(error)
              return
            }
            if (originalRequest.headers) {
              originalRequest.headers.Authorization = `Bearer ${newToken}`
            }
            resolve(request(originalRequest))
          })
        })
      }

      // 开始刷新 Token
      isRefreshing = true

      try {
        const newToken = await refreshToken()
        
        if (newToken) {
          // 刷新成功，执行队列中的请求
          onTokenRefreshFinished(newToken)
          
          // 重试原请求
          if (originalRequest.headers) {
            originalRequest.headers.Authorization = `Bearer ${newToken}`
          }
          return request(originalRequest)
        } else {
          onTokenRefreshFinished(null)
          // 刷新失败，清除登录信息并跳转登录页
          clearLoginInfo()
          message.error('登录已过期，请重新登录')
          window.location.href = '/login'
          return Promise.reject(error)
        }
      } finally {
        isRefreshing = false
      }
    }

    // 权限不足 (403)
    if (error.response?.status === 403) {
      message.error('权限不足，无法访问')
      return Promise.reject(error)
    }

    // 服务器错误 (500+)
    if (error.response?.status && error.response.status >= 500) {
      message.error('服务器错误，请稍后重试')
      return Promise.reject(error)
    }

    // 网络错误
    if (!error.response) {
      message.error('网络连接失败，请检查网络')
      return Promise.reject(error)
    }

    // 其他错误
    const errorMessage = error.response?.data?.message || '请求失败'
    const silent = error.config?._silent
    if (!silent) {
      message.error(errorMessage)
    }
    return Promise.reject(error)
  }
)

// ====== 请求方法封装 ======

/**
 * GET 请求
 */
export async function get<T>(url: string, params?: Record<string, unknown>, config?: RequestConfig): Promise<T> {
  const response = await request.get<ApiResponse<T>>(url, { params, ...config })
  return response.data.data
}

/**
 * POST 请求
 */
export async function post<T>(url: string, data?: unknown, config?: RequestConfig): Promise<T> {
  const response = await request.post<ApiResponse<T>>(url, data, config)
  return response.data.data
}

/**
 * PUT 请求
 */
export async function put<T>(url: string, data?: unknown, config?: RequestConfig): Promise<T> {
  const response = await request.put<ApiResponse<T>>(url, data, config)
  return response.data.data
}

/**
 * DELETE 请求
 */
export async function del<T>(url: string, config?: RequestConfig): Promise<T> {
  const response = await request.delete<ApiResponse<T>>(url, config)
  return response.data.data
}

/**
 * 上传文件
 */
export async function upload<T>(url: string, file: File, type: string = 'file'): Promise<T> {
  const formData = new FormData()
  formData.append('file', file)
  formData.append('type', type)
  
  const response = await request.post<ApiResponse<T>>(url, formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  })
  return response.data.data
}

export default request
