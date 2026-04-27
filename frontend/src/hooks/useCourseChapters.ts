import { useCallback } from 'react'

import { createChapter, deleteChapter, updateChapter } from '@/api/course'

export function useCourseChapters(courseId: number) {
  const saveChapter = useCallback(async (
    chapterId: number | null,
    data: { title: string; description?: string },
  ) => {
    if (chapterId) {
      return updateChapter(courseId, chapterId, data)
    }
    return createChapter(courseId, data)
  }, [courseId])

  const removeChapter = useCallback(async (chapterId: number) => {
    await deleteChapter(courseId, chapterId)
  }, [courseId])

  return {
    saveChapter,
    removeChapter,
  }
}
