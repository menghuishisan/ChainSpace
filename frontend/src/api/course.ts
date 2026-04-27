/**
 * 课程相关 API
 */
import { get, post, put, del } from './request'
import type { 
  Course, 
  Chapter, 
  Material, 
  CourseStudent, 
  CourseProgress,
  PaginatedData,
} from '@/types'

// ====== 课程管理（教师）======

/**
 * 获取我的课程列表
 * 后端路由: GET /courses
 */
export function getCourses(params?: {
  page?: number
  page_size?: number
  status?: string
}): Promise<PaginatedData<Course>> {
  return get<PaginatedData<Course>>('/courses', params)
}

/**
 * 创建课程
 * 后端路由: POST /courses
 */
export function createCourse(data: {
  title: string
  description?: string
  cover?: string
  is_public?: boolean
}): Promise<Course> {
  return post<Course>('/courses', data)
}

/**
 * 获取课程详情
 * 后端路由: GET /courses/:id
 */
export function getCourse(id: number): Promise<Course> {
  return get<Course>(`/courses/${id}`)
}

/**
 * 更新课程
 * 后端路由: PUT /courses/:id
 */
export function updateCourse(id: number, data: {
  title?: string
  description?: string
  cover?: string
  is_public?: boolean
}): Promise<Course> {
  return put<Course>(`/courses/${id}`, data)
}

/**
 * 更新课程状态
 * 后端路由: PUT /courses/:id/status
 */
export function updateCourseStatus(id: number, status: 'published' | 'archived'): Promise<void> {
  return put(`/courses/${id}/status`, { status })
}

/**
 * 删除课程
 * 后端路由: DELETE /courses/:id
 */
export function deleteCourse(id: number): Promise<void> {
  return del(`/courses/${id}`)
}

/**
 * 重置邀请码
 * 后端路由: POST /courses/:id/invite-code/reset
 */
export function resetInviteCode(id: number): Promise<{ invite_code: string }> {
  return post<{ invite_code: string }>(`/courses/${id}/invite-code/reset`)
}

// ====== 章节管理 ======

/**
 * 获取章节列表
 * 后端路由: GET /courses/:id/chapters
 */
export function getChapters(courseId: number): Promise<Chapter[]> {
  return get<Chapter[]>(`/courses/${courseId}/chapters`)
}

/**
 * 创建章节
 * 后端路由: POST /courses/:id/chapters
 */
export function createChapter(courseId: number, data: {
  title: string
  description?: string
}): Promise<Chapter> {
  return post<Chapter>(`/courses/${courseId}/chapters`, data)
}

/**
 * 更新章节
 * 后端路由: PUT /courses/:id/chapters/:chapter_id
 */
export function updateChapter(courseId: number, chapterId: number, data: {
  title?: string
  description?: string
}): Promise<Chapter> {
  return put<Chapter>(`/courses/${courseId}/chapters/${chapterId}`, data)
}

/**
 * 删除章节
 * 后端路由: DELETE /courses/:id/chapters/:chapter_id
 */
export function deleteChapter(courseId: number, chapterId: number): Promise<void> {
  return del(`/courses/${courseId}/chapters/${chapterId}`)
}

/**
 * 调整章节顺序
 * 后端路由: PUT /courses/:id/chapters/sort
 */
export function sortChapters(courseId: number, chapterIds: number[]): Promise<void> {
  return put(`/courses/${courseId}/chapters/sort`, { chapter_ids: chapterIds })
}

// ====== 学习资料管理 ======

/**
 * 创建资料
 * 后端路由: POST /courses/:id/chapters/:chapter_id/materials
 */
export function createMaterial(courseId: number, chapterId: number, data: {
  title: string
  type: string
  content?: string
  url?: string
  duration?: number
}): Promise<Material> {
  return post<Material>(`/courses/${courseId}/chapters/${chapterId}/materials`, data)
}

/**
 * 更新资料
 * 后端路由: PUT /courses/:id/chapters/:chapter_id/materials/:material_id
 */
export function updateMaterial(courseId: number, chapterId: number, materialId: number, data: {
  title?: string
  type?: string
  content?: string
  url?: string
  duration?: number
}): Promise<Material> {
  return put<Material>(`/courses/${courseId}/chapters/${chapterId}/materials/${materialId}`, data)
}

/**
 * 删除资料
 * 后端路由: DELETE /courses/:id/chapters/:chapter_id/materials/:material_id
 */
export function deleteMaterial(courseId: number, chapterId: number, materialId: number): Promise<void> {
  return del(`/courses/${courseId}/chapters/${chapterId}/materials/${materialId}`)
}

// ====== 课程学生管理 ======

/**
 * 获取课程学生列表
 * 后端路由: GET /courses/:id/students
 */
export function getCourseStudents(courseId: number, params?: {
  page?: number
  page_size?: number
  keyword?: string
}): Promise<PaginatedData<CourseStudent>> {
  return get<PaginatedData<CourseStudent>>(`/courses/${courseId}/students`, params)
}

/**
 * 添加学生到课程（通过手机号或学号）
 * 后端路由: POST /courses/:id/students
 */
export function addCourseStudents(courseId: number, data: {
  phones?: string[]
  student_nos?: string[]
}): Promise<void> {
  return post(`/courses/${courseId}/students`, data)
}

/**
 * 移除课程学生
 * 后端路由: DELETE /courses/:id/students
 */
export function removeCourseStudents(courseId: number, studentIds: number[]): Promise<void> {
  return del(`/courses/${courseId}/students`, { data: { student_ids: studentIds } })
}

/**
 * 批量导入学生到课程（通过Excel）
 * 后端路由: POST /courses/:id/students/import
 */
export function batchImportStudentsToCourse(courseId: number, file: File): Promise<{
  success: number
  failed: number
  errors?: string[]
}> {
  const formData = new FormData()
  formData.append('file', file)
  
  return post(`/courses/${courseId}/students/import`, formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  })
}

// ====== 学生课程操作 ======

/**
 * 通过邀请码加入课程
 * 后端路由: POST /courses/join
 */
export function joinCourse(inviteCode: string): Promise<Course> {
  return post<Course>('/courses/join', { code: inviteCode })
}

/**
 * 获取我的课程列表
 * 后端路由: GET /courses/my
 */
export function getMyCourses(params?: {
  page?: number
  page_size?: number
}): Promise<PaginatedData<Course>> {
  return get<PaginatedData<Course>>('/courses/my', params)
}

/**
 * 更新资料学习进度
 * 后端路由: PUT /courses/:id/materials/:material_id/progress
 */
export function updateMaterialProgress(courseId: number, materialId: number, data: {
  progress: number
  last_position?: number
}): Promise<void> {
  return put(`/courses/${courseId}/materials/${materialId}/progress`, data)
}

// ====== 学生端API ======

/**
 * 获取章节资料列表
 * 后端路由: GET /courses/:id/chapters/:chapter_id/materials
 */
export function getMaterials(courseId: number, chapterId: number): Promise<Material[]> {
  return get<Material[]>(`/courses/${courseId}/chapters/${chapterId}/materials`)
}

/**
 * 获取课程学习进度
 * 后端路由: GET /courses/:id/progress
 */
export function getCourseProgress(courseId: number): Promise<CourseProgress> {
  return get<CourseProgress>(`/courses/${courseId}/progress`)
}
