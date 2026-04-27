/**
 * 上传模块 API
 * 后端路由: /upload/*
 */
import { upload as uploadRequest } from './request'

/**
 * 上传头像
 * 后端路由: POST /upload/avatar
 */
export function uploadAvatar(file: File): Promise<{ url: string }> {
  return uploadRequest<{ url: string }>('/upload/avatar', file, 'avatar')
}

/**
 * 上传课程封面
 * 后端路由: POST /upload/courses/:course_id/cover
 */
export function uploadCourseCover(courseId: number, file: File): Promise<{ url: string }> {
  return uploadRequest<{ url: string }>(`/upload/courses/${courseId}/cover`, file, 'cover')
}

/**
 * 上传课程资料
 * 后端路由: POST /upload/courses/:course_id/chapters/:chapter_id/materials
 */
export function uploadMaterial(
  courseId: number, 
  chapterId: number, 
  file: File
): Promise<{ url: string; filename: string; size: number }> {
  return uploadRequest<{ url: string; filename: string; size: number }>(
    `/upload/courses/${courseId}/chapters/${chapterId}/materials`, 
    file, 
    'material'
  )
}

/**
 * 上传实验提交文件
 * 后端路由: POST /upload/experiments/:experiment_id/submission
 */
export function uploadSubmission(
  experimentId: number, 
  file: File
): Promise<{ url: string; filename: string }> {
  return uploadRequest<{ url: string; filename: string }>(
    `/upload/experiments/${experimentId}/submission`, 
    file, 
    'submission'
  )
}

/**
 * 上传题目附件
 * 后端路由: POST /upload/challenges/:challenge_id/attachment
 */
export function uploadChallengeAttachment(
  challengeId: number, 
  file: File
): Promise<{ url: string; filename: string }> {
  return uploadRequest<{ url: string; filename: string }>(
    `/upload/challenges/${challengeId}/attachment`, 
    file, 
    'attachment'
  )
}

/**
 * 上传实验初始资源
 * 后端路由: POST /upload/experiments/assets
 */
export function uploadExperimentAsset(
  file: File,
): Promise<{ bucket: string; url: string; path: string; filename: string; size: number }> {
  return uploadRequest<{ bucket: string; url: string; path: string; filename: string; size: number }>(
    '/upload/experiments/assets',
    file,
    'experiment-asset'
  )
}
