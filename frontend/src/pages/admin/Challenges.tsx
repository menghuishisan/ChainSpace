import { useCallback, useEffect, useState } from 'react'
import { getChallenges } from '@/api/challenge'
import { ChallengeDetailModal, ChallengeSearchFilter, ChallengeTable } from '@/components/challenge'
import { PageHeader } from '@/components/common'
import {
  buildChallengeListQueryParams,
  DEFAULT_CHALLENGE_FILTERS,
} from '@/domains/challenge/management'
import type { Challenge, PaginatedData } from '@/types'
import type { ChallengeListFilters } from '@/types/presentation'

export default function AdminChallenges() {
  const [loading, setLoading] = useState(false)
  const [data, setData] = useState<PaginatedData<Challenge>>({ list: [], total: 0, page: 1, page_size: 20 })
  const [filters, setFilters] = useState<ChallengeListFilters>(DEFAULT_CHALLENGE_FILTERS)
  const [pagination, setPagination] = useState({ page: 1, page_size: 20 })
  const [selectedChallenge, setSelectedChallenge] = useState<Challenge | null>(null)
  const [detailVisible, setDetailVisible] = useState(false)

  const fetchData = useCallback(async () => {
    setLoading(true)
    try {
      const result = await getChallenges(buildChallengeListQueryParams(
        pagination.page,
        pagination.page_size,
        filters,
        { is_public: true },
      ))
      setData(result)
    } finally {
      setLoading(false)
    }
  }, [filters, pagination])

  useEffect(() => {
    void fetchData()
  }, [fetchData])

  return (
    <div>
      <PageHeader
        title="公共题库"
        subtitle="平台侧按统一挑战模型审查公开题目质量"
      />

      <ChallengeSearchFilter
        values={filters}
        onChange={setFilters}
        onSearch={() => setPagination((current) => ({ ...current, page: 1 }))}
        onReset={() => {
          setFilters(DEFAULT_CHALLENGE_FILTERS)
          setPagination({ page: 1, page_size: 20 })
        }}
      />

      <div className="card">
        <ChallengeTable
          data={data}
          loading={loading}
          onView={(challenge) => {
            setSelectedChallenge(challenge)
            setDetailVisible(true)
          }}
          onPageChange={(page, pageSize) => setPagination({ page, page_size: pageSize })}
        />
      </div>

      <ChallengeDetailModal
        open={detailVisible}
        challenge={selectedChallenge}
        onCancel={() => setDetailVisible(false)}
      />
    </div>
  )
}
