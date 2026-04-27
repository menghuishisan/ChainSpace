import { useCallback, useEffect, useState } from 'react'

import { getExperiments } from '@/api/experiment'
import type { Experiment } from '@/types'

export function useCourseExperiments(courseId: number) {
  const [loading, setLoading] = useState(false)
  const [experiments, setExperiments] = useState<Experiment[]>([])

  const refreshExperiments = useCallback(async () => {
    setLoading(true)
    try {
      const result = await getExperiments({ page: 1, page_size: 100 })
      setExperiments(result.list.filter((item) => item.course_id === courseId))
    } finally {
      setLoading(false)
    }
  }, [courseId])

  useEffect(() => {
    void refreshExperiments()
  }, [refreshExperiments])

  return {
    loading,
    experiments,
    refreshExperiments,
  }
}
