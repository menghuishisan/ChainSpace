import { useCallback, useEffect, useState } from 'react'

import {
  addCourseStudents,
  batchImportStudentsToCourse,
  getCourseStudents,
  removeCourseStudents,
} from '@/api/course'
import type { CourseStudent, PaginatedData } from '@/types'

export function useCourseStudents(courseId: number) {
  const [loading, setLoading] = useState(false)
  const [students, setStudents] = useState<PaginatedData<CourseStudent>>({
    list: [],
    total: 0,
    page: 1,
    page_size: 20,
  })

  const refreshStudents = useCallback(async () => {
    setLoading(true)
    try {
      const result = await getCourseStudents(courseId, { page: 1, page_size: 100 })
      setStudents(result)
    } finally {
      setLoading(false)
    }
  }, [courseId])

  useEffect(() => {
    void refreshStudents()
  }, [refreshStudents])

  const addStudentsByPhones = useCallback(async (phones: string[]) => {
    await addCourseStudents(courseId, { phones })
    await refreshStudents()
  }, [courseId, refreshStudents])

  const removeStudent = useCallback(async (studentId: number) => {
    await removeCourseStudents(courseId, [studentId])
    await refreshStudents()
  }, [courseId, refreshStudents])

  const importStudents = useCallback(async (file: File) => {
    const result = await batchImportStudentsToCourse(courseId, file)
    if (result.success > 0) {
      await refreshStudents()
    }
    return result
  }, [courseId, refreshStudents])

  return {
    loading,
    students,
    refreshStudents,
    addStudentsByPhones,
    removeStudent,
    importStudents,
  }
}
