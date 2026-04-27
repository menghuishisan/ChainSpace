import { useCallback, useState } from 'react'

import {
  createMaterial,
  deleteMaterial,
  getMaterials,
  updateMaterial,
} from '@/api/course'
import { uploadMaterial } from '@/api/upload'
import type { Material } from '@/types'

export interface CourseMaterialFormData {
  title: string
  type: string
  content?: string
  url?: string
  duration?: number
}

export function useCourseMaterials(courseId: number) {
  const [chapterMaterials, setChapterMaterials] = useState<Record<number, Material[]>>({})

  const fetchMaterials = useCallback(async (chapterId: number) => {
    const materials = await getMaterials(courseId, chapterId)
    setChapterMaterials((current) => ({ ...current, [chapterId]: materials }))
    return materials
  }, [courseId])

  const saveMaterial = useCallback(async (
    chapterId: number,
    materialId: number | null,
    data: CourseMaterialFormData,
  ) => {
    if (materialId) {
      await updateMaterial(courseId, chapterId, materialId, data)
    } else {
      await createMaterial(courseId, chapterId, data)
    }
    await fetchMaterials(chapterId)
  }, [courseId, fetchMaterials])

  const removeMaterial = useCallback(async (chapterId: number, materialId: number) => {
    await deleteMaterial(courseId, chapterId, materialId)
    await fetchMaterials(chapterId)
  }, [courseId, fetchMaterials])

  const uploadCourseMaterial = useCallback(async (chapterId: number, file: File) => {
    return uploadMaterial(courseId, chapterId, file)
  }, [courseId])

  return {
    chapterMaterials,
    fetchMaterials,
    saveMaterial,
    removeMaterial,
    uploadCourseMaterial,
  }
}
