import type { ReactNode } from 'react'
import type { Chapter, Material } from './course'

export interface KeepAliveTabPanelProps {
  active: boolean
  children: ReactNode
}

export interface StudentCourseLearnTabProps {
  courseId: number
  chapters: Chapter[]
}

export interface StudentCourseExperimentsTabProps {
  courseId: number
}

export interface ChapterMaterialsMap {
  [chapterId: number]: Material[]
}
