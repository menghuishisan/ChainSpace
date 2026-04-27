/**
 * 教师 - 实验列表页面
 */
import { useState, useEffect, useCallback } from 'react'
import { Table, Button, Tag, Space, Select, Empty, message, Card } from 'antd'
import { PlusOutlined, EyeOutlined, EditOutlined, UploadOutlined, DeploymentUnitOutlined, CheckCircleOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { useNavigate } from 'react-router-dom'
import { PageHeader, StatusTag, ExperimentStatusConfig } from '@/components/common'
import type { Experiment, PaginatedData, Course } from '@/types'
import { getCourses } from '@/api/course'
import { getExperiments, publishExperiment } from '@/api/experiment'
import { ExperimentTypeMap } from '@/types'
import { formatDurationCN } from '@/utils/format'

export default function TeacherExperiments() {
  const navigate = useNavigate()
  const [loading, setLoading] = useState(false)
  const [courses, setCourses] = useState<Course[]>([])
  const [selectedCourseId, setSelectedCourseId] = useState<number | undefined>()
  const [data, setData] = useState<PaginatedData<Experiment>>({ list: [], total: 0, page: 1, page_size: 20 })
  const [pagination, setPagination] = useState({ page: 1, page_size: 20 })
  const [publishingId, setPublishingId] = useState<number | null>(null)

  // 获取课程列表
  useEffect(() => {
    const fetchCourses = async () => {
      try {
        const result = await getCourses({ page: 1, page_size: 100 })
        setCourses(result.list)
        if (result.list.length > 0) {
          setSelectedCourseId(result.list[0].id)
        }
      } catch {
        // 错误由拦截器处理
      }
    }
    fetchCourses()
  }, [])

  // 获取实验列表
  const fetchData = useCallback(async () => {
    if (!selectedCourseId) return
    setLoading(true)
    try {
      const result = await getExperiments(pagination)
      const filtered = result.list.filter((item) => item.course_id === selectedCourseId)
      setData({ ...result, list: filtered, total: filtered.length })
    } catch {
      // 错误由拦截器处理
    } finally {
      setLoading(false)
    }
  }, [selectedCourseId, pagination])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  const overview = {
    total: data.list.length,
    draft: data.list.filter((item) => item.status === 'draft').length,
    published: data.list.filter((item) => item.status === 'published').length,
    collaborative: data.list.filter((item) => item.mode === 'collaboration' || item.mode === 'multi_node').length,
  }

  const handlePublish = async (experimentId: number) => {
    setPublishingId(experimentId)
    try {
      await publishExperiment(experimentId)
      message.success('实验已发布')
      await fetchData()
    } catch {
      // 错误由拦截器处理
    } finally {
      setPublishingId(null)
    }
  }

  // 表格列定义
  const columns: ColumnsType<Experiment> = [
    { title: '实验名称', dataIndex: 'title', key: 'title', width: 200 },
    {
      title: '实验类型',
      dataIndex: 'type',
      key: 'type',
      width: 140,
      render: (type) => (
        <Tag>{ExperimentTypeMap[type as keyof typeof ExperimentTypeMap] || type}</Tag>
      ),
    },
    {
      title: '所属章节',
      dataIndex: 'chapter_title',
      key: 'chapter_title',
      width: 140,
      render: (text) => text || '-',
    },
    {
      title: '预计时长',
      dataIndex: 'estimated_time',
      key: 'estimated_time',
      width: 100,
      render: (duration) => formatDurationCN(duration),
    },
    {
      title: '满分',
      dataIndex: 'max_score',
      key: 'max_score',
      width: 80,
      render: (score) => `${score}分`,
    },
    {
      title: '评分方式',
      dataIndex: 'auto_grade',
      key: 'auto_grade',
      width: 100,
      render: (v) => (v ? '自动评测' : '人工批改'),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status) => <StatusTag status={status} statusMap={ExperimentStatusConfig} />,
    },
    {
      title: '操作',
      key: 'action',
      width: 220,
      render: (_, record) => (
        <Space>
          <Button type="link" size="small" icon={<EyeOutlined />} onClick={() => navigate(`/teacher/experiments/${record.id}/edit`)}>
            查看
          </Button>
          <Button type="link" size="small" icon={<EditOutlined />} onClick={() => navigate(`/teacher/experiments/${record.id}/edit`)}>
            编辑
          </Button>
          {record.status === 'draft' && (
            <Button
              type="link"
              size="small"
              icon={<UploadOutlined />}
              loading={publishingId === record.id}
              onClick={() => handlePublish(record.id)}
            >
              发布
            </Button>
          )}
        </Space>
      ),
    },
  ]

  return (
    <div className="space-y-6">
      <PageHeader
        title="实验管理"
        subtitle="管理课程实验、发布状态和实验入口"
        extra={
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => navigate(`/teacher/experiments/create${selectedCourseId ? `?course_id=${selectedCourseId}` : ''}`)}
          >
            创建实验
          </Button>
        }
      />

      <Card className="overflow-hidden border-0 shadow-sm" styles={{ body: { padding: 0 } }}>
        <div className="grid gap-px bg-slate-200 xl:grid-cols-[1.35fr_repeat(4,minmax(0,1fr))]">
          <div className="bg-[linear-gradient(135deg,#0b2239_0%,#0f2744_55%,#113d67_100%)] px-6 py-6 text-white">
            <div className="text-xs uppercase tracking-[0.28em] text-cyan-200">实验管理台</div>
            <div className="mt-3 text-2xl font-semibold">统一管理实验内容、发布和评测</div>
            <p className="mt-3 text-sm leading-6 text-slate-200">
              这里重点处理实验内容、发布时间和评测方式，方便按课程组织实验任务。
            </p>
          </div>
          <div className="bg-white px-5 py-5"><div className="text-xs uppercase tracking-[0.2em] text-text-secondary">全部实验</div><div className="mt-3 text-3xl font-semibold text-slate-900">{overview.total}</div></div>
          <div className="bg-white px-5 py-5"><div className="text-xs uppercase tracking-[0.2em] text-text-secondary">草稿</div><div className="mt-3 text-3xl font-semibold text-slate-900">{overview.draft}</div></div>
          <div className="bg-white px-5 py-5"><div className="flex items-center gap-2 text-xs uppercase tracking-[0.2em] text-text-secondary"><CheckCircleOutlined className="text-emerald-500" />已发布</div><div className="mt-3 text-3xl font-semibold text-slate-900">{overview.published}</div></div>
          <div className="bg-white px-5 py-5"><div className="flex items-center gap-2 text-xs uppercase tracking-[0.2em] text-text-secondary"><DeploymentUnitOutlined className="text-violet-500" />复杂模式</div><div className="mt-3 text-3xl font-semibold text-slate-900">{overview.collaborative}</div></div>
        </div>
      </Card>

      <Card className="border-0 shadow-sm">
        <Space wrap>
          <span className="text-text-secondary">选择课程：</span>
          <Select
            placeholder="请选择课程"
            value={selectedCourseId}
            onChange={setSelectedCourseId}
            style={{ width: 300 }}
            options={courses.map((c) => ({ label: c.title, value: c.id }))}
            notFoundContent={<Empty description="暂无课程" />}
          />
        </Space>
      </Card>

      <Card className="border-0 shadow-sm">
        <Table
          columns={columns}
          dataSource={data.list}
          rowKey="id"
          loading={loading}
          locale={{
            emptyText: selectedCourseId
              ? <Empty description="暂无实验，尝试调整筛选或创建新实验" />
              : <Empty description="请先选择课程" />,
          }}
          pagination={{
            current: data.page,
            pageSize: data.page_size,
            total: data.total,
            showSizeChanger: true,
            showTotal: (total) => `共 ${total} 条`,
            onChange: (page, pageSize) => setPagination({ page, page_size: pageSize }),
          }}
        />
      </Card>
    </div>
  )
}
