/**
 * 学生 - 我的实验页面
 */
import { useState, useEffect, useCallback } from 'react'
import { Table, Button, Tag, Space, Card } from 'antd'
import { PlayCircleOutlined, RadarChartOutlined, DeploymentUnitOutlined, CheckCircleOutlined, TrophyOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { useNavigate } from 'react-router-dom'
import { PageHeader, SearchFilter, StatusTag, SubmissionStatusConfig } from '@/components/common'
import type { Experiment, PaginatedData, Course, Chapter, ExperimentEnv, Submission } from '@/types'
import { getMyCourses, getChapters } from '@/api/course'
import { getStudentExperiments } from '@/api/experiment'
import { getSubmissions } from '@/api/experimentSubmission'
import { listExperimentEnvs } from '@/api/experimentSession'
import { ExperimentTypeMap } from '@/types'
import { formatDurationCN } from '@/utils/format'
import EmptyState from '@/components/common/EmptyState'

type StudentExperimentRow = Experiment & {
  submission_status?: string
  score?: number
  env_status?: string
}

type StudentExperimentOverview = {
  total: number
  active: number
  submitted: number
  graded: number
}

type ExperimentSubmissionMeta = {
  status?: string
  score?: number
}

const activeEnvStatuses = new Set(['pending', 'creating', 'running', 'paused'])

export default function StudentExperiments() {
  const navigate = useNavigate()
  const [loading, setLoading] = useState(false)
  const [courses, setCourses] = useState<Course[]>([])
  const [chapters, setChapters] = useState<Chapter[]>([])
  const [data, setData] = useState<PaginatedData<Experiment>>({ list: [], total: 0, page: 1, page_size: 20 })
  const [overview, setOverview] = useState<StudentExperimentOverview>({ total: 0, active: 0, submitted: 0, graded: 0 })
  const [submissionMetaByExperiment, setSubmissionMetaByExperiment] = useState<Record<number, ExperimentSubmissionMeta>>({})
  const [envStatusByExperiment, setEnvStatusByExperiment] = useState<Record<number, string>>({})
  const [filters, setFilters] = useState<Record<string, unknown>>({ keyword: '', course_id: '', chapter_id: '' })
  const [pagination, setPagination] = useState({ page: 1, page_size: 20 })

  const buildQueryParams = useCallback((pageInfo?: { page: number; page_size: number }) => {
    const queryParams: Record<string, unknown> = {
      ...(pageInfo || pagination),
    }
    if (filters.keyword && String(filters.keyword).trim()) {
      queryParams.keyword = String(filters.keyword).trim()
    }
    if (filters.course_id) {
      queryParams.course_id = filters.course_id
    }
    if (filters.chapter_id) {
      queryParams.chapter_id = filters.chapter_id
    }
    return queryParams
  }, [filters, pagination])

  // 获取课程列表
  useEffect(() => {
    getMyCourses({ page: 1, page_size: 100 })
      .then((res: { list: Course[] }) => setCourses(res.list || []))
      .catch(() => setCourses([]))
  }, [])

  // 当选择课程时，获取该课程的章节列表
  const selectedCourseId = filters.course_id as number | undefined
  useEffect(() => {
    if (selectedCourseId) {
      getChapters(selectedCourseId)
        .then((res) => {
          const chaptersData = Array.isArray(res) ? res : []
          setChapters(chaptersData as unknown as Chapter[])
        })
        .catch(() => setChapters([]))
    } else {
      setChapters([])
    }
    // 选择新课程时重置章节筛选
    setFilters(prev => ({ ...prev, chapter_id: '' }))
  }, [selectedCourseId])

  // 获取实验列表
  const fetchData = useCallback(async () => {
    setLoading(true)
    try {
      const queryParams = buildQueryParams()
      const result = await getStudentExperiments(queryParams as Parameters<typeof getStudentExperiments>[0])
      setData(result)
    } catch { /* */ } finally { setLoading(false) }
  }, [buildQueryParams])

  useEffect(() => { fetchData() }, [fetchData])

  useEffect(() => {
    const fetchAllPages = async <T,>(
      requestPage: (page: number, pageSize: number) => Promise<PaginatedData<T>>,
    ): Promise<T[]> => {
      const pageSize = 100
      const firstPage = await requestPage(1, pageSize)
      const all = [...(firstPage.list || [])]
      const total = firstPage.total || all.length
      const totalPages = Math.max(1, Math.ceil(total / pageSize))
      for (let page = 2; page <= totalPages; page += 1) {
        const pageResult = await requestPage(page, pageSize)
        all.push(...(pageResult.list || []))
      }
      return all
    }

    let active = true

    const fetchOverview = async () => {
      try {
        const baseQuery = buildQueryParams({ page: 1, page_size: 100 })
        const [allExperiments, allSubmissions, allEnvs] = await Promise.all([
          fetchAllPages((page, pageSize) => getStudentExperiments({
            ...baseQuery,
            page,
            page_size: pageSize,
          } as Parameters<typeof getStudentExperiments>[0])),
          fetchAllPages((page, pageSize) => getSubmissions({ page, page_size: pageSize })),
          fetchAllPages((page, pageSize) => listExperimentEnvs({ page, page_size: pageSize })),
        ])

        if (!active) {
          return
        }

        const experimentIds = new Set(allExperiments.map((item) => item.id))

        const latestSubmissionByExperiment: Record<number, ExperimentSubmissionMeta & { submitted_at?: string; attempt_number?: number }> = {}
        allSubmissions.forEach((submission: Submission) => {
          if (!experimentIds.has(submission.experiment_id)) {
            return
          }
          const current = latestSubmissionByExperiment[submission.experiment_id]
          const shouldReplace = !current
            || new Date(submission.submitted_at).getTime() > new Date(current.submitted_at || 0).getTime()
            || submission.attempt_number > (current.attempt_number || 0)
          if (shouldReplace) {
            latestSubmissionByExperiment[submission.experiment_id] = {
              status: submission.status,
              score: submission.score ?? undefined,
              submitted_at: submission.submitted_at,
              attempt_number: submission.attempt_number,
            }
          }
        })

        const latestEnvByExperiment: Record<number, { status: string; created_at: string }> = {}
        allEnvs.forEach((env: ExperimentEnv) => {
          if (!experimentIds.has(env.experiment_id)) {
            return
          }
          const current = latestEnvByExperiment[env.experiment_id]
          if (!current || new Date(env.created_at).getTime() > new Date(current.created_at).getTime()) {
            latestEnvByExperiment[env.experiment_id] = {
              status: env.status,
              created_at: env.created_at,
            }
          }
        })

        const nextSubmissionMeta: Record<number, ExperimentSubmissionMeta> = {}
        Object.entries(latestSubmissionByExperiment).forEach(([experimentId, meta]) => {
          nextSubmissionMeta[Number(experimentId)] = {
            status: meta.status,
            score: meta.score,
          }
        })

        const nextEnvStatus: Record<number, string> = {}
        Object.entries(latestEnvByExperiment).forEach(([experimentId, meta]) => {
          nextEnvStatus[Number(experimentId)] = meta.status
        })

        setSubmissionMetaByExperiment(nextSubmissionMeta)
        setEnvStatusByExperiment(nextEnvStatus)
        setOverview({
          total: allExperiments.length,
          active: Object.values(nextEnvStatus).filter((status) => activeEnvStatuses.has(status)).length,
          submitted: Object.keys(nextSubmissionMeta).length,
          graded: Object.values(nextSubmissionMeta).filter((meta) => meta.status === 'graded').length,
        })
      } catch {
        if (!active) {
          return
        }
        setSubmissionMetaByExperiment({})
        setEnvStatusByExperiment({})
        setOverview({ total: 0, active: 0, submitted: 0, graded: 0 })
      }
    }

    fetchOverview()

    return () => {
      active = false
    }
  }, [buildQueryParams])

  const tableData: StudentExperimentRow[] = data.list.map((item) => ({
    ...item,
    submission_status: submissionMetaByExperiment[item.id]?.status || item.my_status,
    score: submissionMetaByExperiment[item.id]?.score ?? item.my_score,
    env_status: envStatusByExperiment[item.id],
  }))

  const columns: ColumnsType<StudentExperimentRow> = [
    { title: '实验名称', dataIndex: 'title', key: 'title', width: 220 },
    { title: '所属章节', dataIndex: 'chapter_title', key: 'chapter_title', width: 160 },
    { title: '类型', dataIndex: 'type', key: 'type', width: 120, render: (t) => <Tag color="blue">{ExperimentTypeMap[t as keyof typeof ExperimentTypeMap] || t}</Tag> },
    { title: '预计时长', dataIndex: 'estimated_time', key: 'estimated_time', width: 100, render: (d) => formatDurationCN(d) },
    { title: '满分', dataIndex: 'max_score', key: 'max_score', width: 80, render: (s) => `${s}分` },
    { title: '提交状态', key: 'submission_status', width: 120, render: (_, r) => r.submission_status ? <StatusTag status={r.submission_status} statusMap={SubmissionStatusConfig} /> : <Tag>未提交</Tag> },
    { title: '得分', key: 'score', width: 80, render: (_, r) => r.score !== undefined ? <span className={r.score >= 60 ? 'text-success' : 'text-error'}>{r.score}分</span> : '-' },
    { title: '操作', key: 'action', width: 120, render: (_, r) => (<Space><Button type="link" size="small" icon={<PlayCircleOutlined />} onClick={() => navigate(`/student/experiments/${r.id}`)}>进入</Button></Space>) },
  ]

  // 筛选配置：课程筛选 + 章节筛选 + 关键词搜索
  const filterConfig = [
    { key: 'course_id', label: '课程', type: 'select' as const, placeholder: '全部课程', options: [{ label: '全部课程', value: '' }, ...courses.map(c => ({ label: c.title, value: c.id }))] },
    { key: 'chapter_id', label: '章节', type: 'select' as const, placeholder: '全部章节', options: [{ label: '全部章节', value: '' }, ...chapters.map(ch => ({ label: ch.title, value: ch.id }))] },
    { key: 'keyword', label: '关键词', type: 'input' as const, placeholder: '搜索实验名称或描述' },
  ]

  return (
    <div className="space-y-6">
      <PageHeader title="我的实验" subtitle="查看课程实验、继续未完成任务并进入实验环境" />
      <Card className="overflow-hidden border-0 shadow-sm" styles={{ body: { padding: 0 } }}>
        <div className="grid gap-px bg-slate-200 xl:grid-cols-[1.35fr_repeat(4,minmax(0,1fr))]">
          <div className="bg-[linear-gradient(135deg,#0b2239_0%,#0f2744_55%,#113d67_100%)] px-6 py-5 text-white">
            <div className="text-xs uppercase tracking-[0.28em] text-cyan-200">实验中心</div>
            <div className="mt-2 text-2xl font-semibold">开始你的实验任务</div>
            <p className="mt-2 text-sm leading-6 text-slate-200">
              查看自己可以参与的实验、继续进行中的任务，并进入实验环境。
            </p>
          </div>
          <div className="bg-white px-5 py-5">
            <div className="flex items-center gap-2 text-xs uppercase tracking-[0.2em] text-text-secondary"><RadarChartOutlined className="text-sky-500" />我的实验</div>
            <div className="mt-3 text-3xl font-semibold text-slate-900">{overview.total}</div>
          </div>
          <div className="bg-white px-5 py-5">
            <div className="flex items-center gap-2 text-xs uppercase tracking-[0.2em] text-text-secondary"><DeploymentUnitOutlined className="text-violet-500" />进行中的实验</div>
            <div className="mt-3 text-3xl font-semibold text-slate-900">{overview.active}</div>
          </div>
          <div className="bg-white px-5 py-5">
            <div className="flex items-center gap-2 text-xs uppercase tracking-[0.2em] text-text-secondary"><CheckCircleOutlined className="text-emerald-500" />已提交实验</div>
            <div className="mt-3 text-3xl font-semibold text-slate-900">{overview.submitted}</div>
          </div>
          <div className="bg-white px-5 py-5">
            <div className="flex items-center gap-2 text-xs uppercase tracking-[0.2em] text-text-secondary"><TrophyOutlined className="text-amber-500" />已获得成绩</div>
            <div className="mt-3 text-3xl font-semibold text-slate-900">{overview.graded}</div>
          </div>
        </div>
      </Card>
      <SearchFilter
        filters={filterConfig}
        values={filters}
        onChange={setFilters}
        onSearch={() => setPagination(p => ({ ...p, page: 1 }))}
        onReset={() => { setFilters({ keyword: '', course_id: '', chapter_id: '' }); setPagination({ page: 1, page_size: 20 }) }}
      />
      <Card
        className="border-0 shadow-sm"
        styles={{ body: { padding: 0 } }}
      >
        <Table
          columns={columns}
          dataSource={tableData}
          rowKey="id"
          loading={loading}
          pagination={{
            current: data.page,
            pageSize: data.page_size,
            total: data.total,
            showSizeChanger: true,
            showTotal: (t) => `共 ${t} 条`,
            onChange: (p, s) => setPagination({ page: p, page_size: s }),
          }}
          locale={{
            emptyText: <EmptyState description="暂无实验数据，尝试更换筛选条件" />,
          }}
        />
      </Card>
    </div>
  )
}
