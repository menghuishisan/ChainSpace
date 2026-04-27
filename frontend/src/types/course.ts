/**
 * 课程相关类型定义
 */

// 课程状态
export type CourseStatus = 'draft' | 'published' | 'archived'

// 课程可见性
export type CourseVisibility = 'school' | 'public'

// 课程信息
export interface Course {
  id: number
  school_id: number
  teacher_id: number
  teacher_name?: string
  title: string
  description?: string
  cover?: string
  code?: string
  invite_code?: string
  category?: string
  tags?: string[]
  status: CourseStatus
  is_public: boolean
  start_date?: string
  end_date?: string
  max_students?: number
  student_count?: number
  chapter_count?: number
  progress?: number
  created_at: string
  updated_at?: string
  class_name?: string
}

// 章节信息
export interface Chapter {
  id: number
  course_id: number
  title: string
  description?: string
  sort_order: number
  material_count?: number
  experiment_count?: number
  created_at: string
}

// 学习资料类型
export type MaterialType = 'video' | 'document' | 'richtext' | 'ppt' | 'link'

// 学习资料
export interface Material {
  id: number
  chapter_id: number
  title: string
  type: MaterialType
  content?: string
  url?: string
  file_size?: number
  duration?: number
  sort_order: number
  status?: string
  progress?: number
  completed?: boolean
  created_at: string
}

// 学习进度
export interface MaterialProgress {
  id: number
  student_id: number
  material_id: number
  progress: number // 0-100
  last_position?: number // 视频播放位置
  completed_at?: string
}

// 课程学生关联
export interface CourseStudent {
  id: number
  student_id: number
  real_name?: string
  student_no?: string
  phone?: string
  class_name?: string
  progress: number
  last_access?: string
  joined_at: string
  status?: string
}

// 课程学习进度
export interface CourseProgress {
  course_id: number
  total_materials: number
  completed_materials: number
  total_experiments: number
  completed_experiments: number
  progress_percent: number
}

// 课程状态显示名称
export const CourseStatusMap: Record<CourseStatus, string> = {
  draft: '草稿',
  published: '已发布',
  archived: '已归档',
}

// 资料类型显示名称
export const MaterialTypeMap: Record<MaterialType, string> = {
  video: '视频',
  document: '文档',
  richtext: '富文本',
  ppt: 'PPT',
  link: '外链',
}
