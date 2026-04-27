import { useCallback, useEffect, useMemo, useState } from 'react'
import { Button, Card, Form, message } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { createContest, getContests, publishContest, updateContest } from '@/api/contest'
import { ContestManageModal, ContestTable } from '@/components/contest'
import { PageHeader, SearchFilter } from '@/components/common'
import {
  buildContestFormInitialValues,
  buildContestListQueryParams,
  buildContestSubmitData,
  CONTEST_FILTER_CONFIG,
  DEFAULT_CONTEST_LIST_FILTERS,
  DEFAULT_CONTEST_PAGINATION,
  normalizeContestFilters,
} from '@/domains/contest/management'
import type { Contest, PaginatedData } from '@/types'
import type { ContestFormValues } from '@/types/presentation'

export default function TeacherContests() {
  const navigate = useNavigate()
  const [loading, setLoading] = useState(false)
  const [data, setData] = useState<PaginatedData<Contest>>({ list: [], total: 0, page: 1, page_size: 20 })
  const [filters, setFilters] = useState(DEFAULT_CONTEST_LIST_FILTERS)
  const [pagination, setPagination] = useState(DEFAULT_CONTEST_PAGINATION)
  const [modalVisible, setModalVisible] = useState(false)
  const [modalLoading, setModalLoading] = useState(false)
  const [editingContest, setEditingContest] = useState<Contest | null>(null)
  const [form] = Form.useForm<ContestFormValues>()

  const fetchData = useCallback(async () => {
    setLoading(true)
    try {
      const result = await getContests(buildContestListQueryParams(pagination, filters))
      setData(result)
    } finally {
      setLoading(false)
    }
  }, [filters, pagination])

  useEffect(() => {
    void fetchData()
  }, [fetchData])

  const overview = useMemo(() => ({
    total: data.list.length,
    draft: data.list.filter((item) => item.status === 'draft').length,
    published: data.list.filter((item) => item.status === 'published').length,
    battle: data.list.filter((item) => item.type === 'agent_battle').length,
  }), [data.list])

  const handleCreate = () => {
    setEditingContest(null)
    form.resetFields()
    setModalVisible(true)
  }

  const handleEdit = (record: Contest) => {
    setEditingContest(record)
    form.setFieldsValue(buildContestFormInitialValues(record))
    setModalVisible(true)
  }

  const handlePublish = async (id: number) => {
    try {
      await publishContest(id)
      message.success('比赛发布成功')
      await fetchData()
    } catch {
      // handled by request interceptor
    }
  }

  const handleSubmit = async (values: ContestFormValues) => {
    setModalLoading(true)
    try {
      const submitData = buildContestSubmitData(values, 'school')
      if (editingContest) {
        await updateContest(editingContest.id, submitData)
        message.success('比赛更新成功')
      } else {
        await createContest(submitData)
        message.success('比赛创建成功')
      }

      setModalVisible(false)
      await fetchData()
    } finally {
      setModalLoading(false)
    }
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="比赛管理"
        subtitle="管理解题赛与智能体对抗赛，统一处理创建、编辑和发布"
        extra={(
          <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
            创建比赛
          </Button>
        )}
      />

      <Card className="overflow-hidden border-0 shadow-sm" styles={{ body: { padding: 0 } }}>
        <div className="grid gap-px bg-slate-200 xl:grid-cols-[1.35fr_repeat(4,minmax(0,1fr))]">
          <div className="bg-[linear-gradient(135deg,#1b1239_0%,#1d2342_55%,#12355b_100%)] px-6 py-6 text-white">
            <div className="text-xs uppercase tracking-[0.28em] text-fuchsia-200">比赛管理台</div>
            <div className="mt-3 text-2xl font-semibold">统一管理解题赛与对抗赛</div>
            <p className="mt-3 text-sm leading-6 text-slate-200">
              在这里完成比赛创建、编辑、发布和状态查看，不再混入旧页面里的冗余说明。
            </p>
          </div>
          <div className="bg-white px-5 py-5"><div className="text-xs uppercase tracking-[0.2em] text-text-secondary">全部比赛</div><div className="mt-3 text-3xl font-semibold text-slate-900">{overview.total}</div></div>
          <div className="bg-white px-5 py-5"><div className="text-xs uppercase tracking-[0.2em] text-text-secondary">草稿</div><div className="mt-3 text-3xl font-semibold text-slate-900">{overview.draft}</div></div>
          <div className="bg-white px-5 py-5"><div className="text-xs uppercase tracking-[0.2em] text-text-secondary">已发布</div><div className="mt-3 text-3xl font-semibold text-slate-900">{overview.published}</div></div>
          <div className="bg-white px-5 py-5"><div className="text-xs uppercase tracking-[0.2em] text-text-secondary">对抗赛</div><div className="mt-3 text-3xl font-semibold text-slate-900">{overview.battle}</div></div>
        </div>
      </Card>

      <SearchFilter
        filters={CONTEST_FILTER_CONFIG}
        values={filters}
        onChange={(values) => setFilters(normalizeContestFilters(values))}
        onSearch={() => setPagination((current) => ({ ...current, page: 1 }))}
        onReset={() => {
          setFilters(DEFAULT_CONTEST_LIST_FILTERS)
          setPagination(DEFAULT_CONTEST_PAGINATION)
        }}
      />

      <Card className="border-0 shadow-sm">
        <ContestTable
          data={data}
          loading={loading}
          onView={(record) => navigate(`/contest/${record.id}`)}
          onEdit={handleEdit}
          onPublish={(contestId) => void handlePublish(contestId)}
          onPageChange={(page, pageSize) => setPagination({ page, page_size: pageSize })}
        />
      </Card>

      <ContestManageModal
        open={modalVisible}
        loading={modalLoading}
        editingContest={editingContest}
        form={form}
        onCancel={() => setModalVisible(false)}
        onSubmit={handleSubmit}
      />
    </div>
  )
}
