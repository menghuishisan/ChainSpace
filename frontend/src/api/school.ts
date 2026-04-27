/**
 * 学校管理员 API
 * 使用后端 /users API 并通过 role 参数区分用户类型
 */
import { get, post, put, del, upload } from './request'
import type { User, Class, PaginatedData, School, CrossSchoolApplication } from '@/types'

// ====== 教师管理 ======

/**
 * 获取教师列表
 * 后端路由: GET /users?role=teacher
 */
export function getTeachers(params?: {
  page?: number
  page_size?: number
  keyword?: string
  status?: string
}): Promise<PaginatedData<User>> {
  return get<PaginatedData<User>>('/users', { ...params, role: 'teacher' })
}

/**
 * 添加教师
 * 后端路由: POST /users
 */
export function addTeacher(data: {
  real_name: string
  phone: string
  email?: string
  password: string
}): Promise<User> {
  return post<User>('/users', { ...data, role: 'teacher' })
}

/**
 * 更新教师
 * 后端路由: PUT /users/:id
 */
export function updateTeacher(id: number, data: {
  real_name?: string
  email?: string
  phone?: string
}): Promise<User> {
  return put<User>(`/users/${id}`, data)
}

/**
 * 更新教师状态
 * 后端路由: PUT /users/:id/status
 */
export function updateTeacherStatus(id: number, status: 'active' | 'disabled'): Promise<void> {
  return put(`/users/${id}/status`, { status })
}

/**
 * 删除教师
 * 后端路由: DELETE /users/:id
 */
export function deleteTeacher(id: number): Promise<void> {
  return del(`/users/${id}`)
}

// ====== 学生管理 ======

/**
 * 获取学生列表
 * 后端路由: GET /users?role=student
 */
export function getStudents(params?: {
  page?: number
  page_size?: number
  keyword?: string
  class_id?: number
  status?: string
}): Promise<PaginatedData<User>> {
  return get<PaginatedData<User>>('/users', { ...params, role: 'student' })
}

/**
 * 添加学生
 * 后端路由: POST /users
 */
export function addStudent(data: {
  real_name: string
  phone: string
  student_no: string
  class_id?: number
  email?: string
  password: string
}): Promise<User> {
  return post<User>('/users', { ...data, role: 'student' })
}

/**
 * 更新学生
 * 后端路由: PUT /users/:id
 */
export function updateStudent(id: number, data: {
  real_name?: string
  class_id?: number
  student_no?: string
  email?: string
  phone?: string
}): Promise<User> {
  return put<User>(`/users/${id}`, data)
}

/**
 * 更新学生状态
 * 后端路由: PUT /users/:id/status
 */
export function updateStudentStatus(id: number, status: 'active' | 'disabled'): Promise<void> {
  return put(`/users/${id}/status`, { status })
}

/**
 * 删除学生
 * 后端路由: DELETE /users/:id
 */
export function deleteStudent(id: number): Promise<void> {
  return del(`/users/${id}`)
}

/**
 * 批量导入学生
 * 后端路由: POST /users/batch-import
 */
export function importStudents(file: File): Promise<{
  success_count: number
  fail_count: number
  fail_details: Array<{ row: number; reason: string }>
}> {
  return upload('/users/batch-import', file, 'excel')
}

// ====== 班级管理 ======

/**
 * 获取班级列表
 * 后端路由: GET /classes
 */
export function getClasses(params?: {
  page?: number
  page_size?: number
}): Promise<PaginatedData<Class>> {
  return get<PaginatedData<Class>>('/classes', params)
}

/**
 * 获取班级详情
 * 后端路由: GET /classes/:id
 */
export function getClass(id: number): Promise<Class> {
  return get<Class>(`/classes/${id}`)
}

/**
 * 创建班级
 * 后端路由: POST /classes
 */
export function createClass(data: {
  name: string
  grade?: string
  major?: string
  description?: string
}): Promise<Class> {
  return post<Class>('/classes', data)
}

/**
 * 更新班级
 * 后端路由: PUT /classes/:id
 */
export function updateClass(id: number, data: {
  name?: string
  grade?: string
  major?: string
  description?: string
  status?: string
}): Promise<Class> {
  return put<Class>(`/classes/${id}`, data)
}

/**
 * 删除班级
 * 后端路由: DELETE /classes/:id
 */
export function deleteClass(id: number): Promise<void> {
  return del(`/classes/${id}`)
}

// ====== 跨校比赛 ======

/**
 * 获取跨校比赛申请列表
 * 后端路由: GET /system/cross-school
 */
export function getCrossSchoolApplications(): Promise<CrossSchoolApplication[]> {
  return get<CrossSchoolApplication[]>('/system/cross-school')
}

/**
 * 申请跨校比赛
 * 后端路由: POST /system/cross-school
 */
export function applyCrossSchoolContest(data: {
  contest_id: number
  target_school_ids: number[]
}): Promise<void> {
  return post('/system/cross-school', data)
}

/**
 * 处理跨校比赛邀请
 * 后端路由: POST /system/cross-school/:id/handle
 */
export function handleCrossSchoolInvitation(id: number, action: 'approve' | 'reject'): Promise<void> {
  return post(`/system/cross-school/${id}/handle`, { action })
}

// ====== 学校信息 ======

/**
 * 获取本校信息
 * 后端路由: GET /schools/current
 */
export function getSchoolInfo(): Promise<School> {
  return get<School>('/schools/current')
}

/**
 * 更新本校信息
 * 后端路由: PUT /schools/current
 */
export function updateSchoolInfo(data: {
  name?: string
  logo_url?: string
  contact_email?: string
  contact_phone?: string
}): Promise<void> {
  return put('/schools/current', data)
}

/**
 * 获取班级学生列表
 * 后端路由: GET /classes/:id/students
 */
export function getClassStudents(classId: number): Promise<PaginatedData<User>> {
  return get<PaginatedData<User>>(`/classes/${classId}/students`)
}

/**
 * 下载学生导入模板
 */
export function downloadStudentTemplate(): Promise<Blob> {
  return get('/users/import-template', {}, { responseType: 'blob' })
}
