import { useCallback, useEffect, useRef, useState } from 'react'
import { Button, message } from 'antd'
import { SyncOutlined } from '@ant-design/icons'

import { PageHeader } from '@/components/common'
import {
  VulnerabilityDetailModal,
  VulnerabilitySearchFilter,
  VulnerabilityTable,
} from '@/components/vulnerability'
import {
  convertVulnerability,
  enrichVulnerabilityCode,
  getVulnerabilities,
  skipVulnerability,
  syncVulnerabilities,
  updateVulnerability,
} from '@/api/admin'
import {
  buildVulnerabilityEditInitialValues,
  buildVulnerabilityListQueryParams,
  buildVulnerabilityUpdateData,
  DEFAULT_VULNERABILITY_EDIT_VALUES,
  DEFAULT_VULNERABILITY_FILTERS,
} from '@/domains/vulnerability/management'
import type { PaginatedData, VulnerabilityCandidate } from '@/types'
import type { VulnerabilityEditFormValues, VulnerabilityListFilters } from '@/types/presentation'

const refreshIntervalMs = 3000
const refreshAttempts = 12

export default function AdminVulnerabilities() {
  const pollingRef = useRef(false)
  const [loading, setLoading] = useState(false)
  const [syncing, setSyncing] = useState(false)
  const [data, setData] = useState<PaginatedData<VulnerabilityCandidate>>({
    list: [],
    total: 0,
    page: 1,
    page_size: 20,
  })
  const [filters, setFilters] = useState<VulnerabilityListFilters>(DEFAULT_VULNERABILITY_FILTERS)
  const [pagination, setPagination] = useState({ page: 1, page_size: 20 })
  const [activeVulnerability, setActiveVulnerability] = useState<VulnerabilityCandidate | null>(null)
  const [editForm, setEditForm] = useState<VulnerabilityEditFormValues>(DEFAULT_VULNERABILITY_EDIT_VALUES)

  const fetchData = useCallback(async () => {
    setLoading(true)
    try {
      const result = await getVulnerabilities(buildVulnerabilityListQueryParams(
        pagination.page,
        pagination.page_size,
        filters,
      ))
      setData(result)
    } finally {
      setLoading(false)
    }
  }, [filters, pagination.page, pagination.page_size])

  useEffect(() => {
    void fetchData()
  }, [fetchData])

  useEffect(() => () => {
    pollingRef.current = false
  }, [])

  const startRefreshPolling = useCallback(async () => {
    pollingRef.current = true
    for (let attempt = 0; attempt < refreshAttempts && pollingRef.current; attempt += 1) {
      await new Promise((resolve) => {
        window.setTimeout(resolve, refreshIntervalMs)
      })
      await fetchData()
    }
    pollingRef.current = false
  }, [fetchData])

  const openDetail = (record: VulnerabilityCandidate) => {
    setActiveVulnerability(record)
    setEditForm(buildVulnerabilityEditInitialValues(record))
  }

  const closeDetail = () => {
    setActiveVulnerability(null)
    setEditForm(DEFAULT_VULNERABILITY_EDIT_VALUES)
  }

  const handleSync = async () => {
    setSyncing(true)
    try {
      const result = await syncVulnerabilities()
      message.success(`同步任务已提交，任务 ID: ${result.task_id}`)
      void startRefreshPolling()
    } finally {
      setSyncing(false)
    }
  }

  const handleConvert = async (record: VulnerabilityCandidate) => {
    await convertVulnerability(record.id)
    message.success('漏洞案例已转化为比赛题目')
    await fetchData()
    closeDetail()
  }

  const handleSkip = async (record: VulnerabilityCandidate) => {
    await skipVulnerability(record.id)
    message.success('漏洞案例已标记为跳过')
    await fetchData()
    if (activeVulnerability?.id === record.id) {
      closeDetail()
    }
  }

  const handleEnrich = async (record: VulnerabilityCandidate) => {
    try {
      const result = await enrichVulnerabilityCode(record.id)
      message.success(`源码增强任务已提交，任务 ID: ${result.task_id}`)
      void startRefreshPolling()
      if (activeVulnerability?.id === record.id) {
        closeDetail()
      }
    } catch {
      // error handled by interceptor
    }
  }

  const handleSave = async () => {
    if (!activeVulnerability) {
      return
    }

    await updateVulnerability(activeVulnerability.id, buildVulnerabilityUpdateData(editForm))
    message.success('漏洞候选信息已保存')
    closeDetail()
    await fetchData()
  }

  return (
    <div>
      <PageHeader
        title="漏洞转化"
        subtitle="围绕多源聚合、自动增强、自动评分与自动转题管理真实漏洞案例"
        extra={(
          <Button type="primary" icon={<SyncOutlined />} onClick={handleSync} loading={syncing}>
            同步多源漏洞
          </Button>
        )}
      />

      <VulnerabilitySearchFilter
        values={filters}
        onChange={setFilters}
        onSearch={() => setPagination((current) => ({ ...current, page: 1 }))}
        onReset={() => {
          setFilters(DEFAULT_VULNERABILITY_FILTERS)
          setPagination({ page: 1, page_size: 20 })
        }}
      />

      <div className="card">
        <VulnerabilityTable
          data={data}
          loading={loading}
          onView={openDetail}
          onEdit={openDetail}
          onEnrich={(record) => void handleEnrich(record)}
          onConvert={(record) => void handleConvert(record)}
          onSkip={(record) => void handleSkip(record)}
          onPageChange={(page, pageSize) => setPagination({ page, page_size: pageSize })}
        />
      </div>

      <VulnerabilityDetailModal
        open={Boolean(activeVulnerability)}
        vulnerability={activeVulnerability}
        formValues={editForm}
        onChange={setEditForm}
        onSave={() => void handleSave()}
        onCancel={closeDetail}
      />
    </div>
  )
}
