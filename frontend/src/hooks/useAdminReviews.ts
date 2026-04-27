import { useCallback, useEffect, useState } from 'react'

import {
  getChallengePublishReviews,
  getCrossSchoolReviews,
  reviewChallengePublish,
  reviewCrossSchool,
} from '@/api/admin'
import type {
  ChallengePublishApplication,
  CrossSchoolApplication,
  PaginatedData,
} from '@/types'

export function useChallengePublishReviews() {
  const [loading, setLoading] = useState(false)
  const [data, setData] = useState<PaginatedData<ChallengePublishApplication>>({
    list: [],
    total: 0,
    page: 1,
    page_size: 20,
  })

  const refresh = useCallback(async () => {
    setLoading(true)
    try {
      setData(await getChallengePublishReviews({ status: 'pending' }))
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void refresh()
  }, [refresh])

  const review = useCallback(async (id: number, action: 'approve' | 'reject', comment?: string) => {
    await reviewChallengePublish(id, action, comment)
    await refresh()
  }, [refresh])

  return { loading, data, refresh, review }
}

export function useCrossSchoolReviews() {
  const [loading, setLoading] = useState(false)
  const [data, setData] = useState<PaginatedData<CrossSchoolApplication>>({
    list: [],
    total: 0,
    page: 1,
    page_size: 20,
  })

  const refresh = useCallback(async () => {
    setLoading(true)
    try {
      setData(await getCrossSchoolReviews({ status: 'pending' }))
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void refresh()
  }, [refresh])

  const review = useCallback(async (id: number, action: 'approve' | 'reject', comment?: string) => {
    await reviewCrossSchool(id, action, comment)
    await refresh()
  }, [refresh])

  return { loading, data, refresh, review }
}
