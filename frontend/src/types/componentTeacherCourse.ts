import type { Chapter, CourseStudent, Material } from '@/types'

export interface TeacherCourseContentTabProps {
  courseId: number
  chapters: Chapter[]
  onRefresh: () => void
}

export interface TeacherChapterMaterialsProps {
  courseId: number
  chapterId: number
  materials: Material[]
}

export interface TeacherCourseExperimentsTabProps {
  courseId: number
}

export interface TeacherCourseStudentsTabProps {
  courseId: number
}

export interface TeacherCourseStudentsState {
  list: CourseStudent[]
  total: number
  page: number
  page_size: number
}
