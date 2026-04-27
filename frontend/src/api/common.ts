/**
 * 通用 API
 * 
 */
import { upload as uploadFile } from './request'
import type { UploadResponse } from '@/types'

// ====== 文件上传 ======

/**
 * 上传文件（通用）
 * 后端路由: POST /upload
 */
export function uploadCommon(file: File, type: string = 'file'): Promise<UploadResponse> {
  return uploadFile<UploadResponse>('/upload', file, type)
}

/**
 * 上传图片
 * 后端路由: POST /upload/image
 */
export function uploadImage(file: File): Promise<UploadResponse> {
  return uploadFile<UploadResponse>('/upload/image', file, 'image')
}
