/**
 * 教师 - 批改作业页面
 */
import { useState, useEffect, useCallback } from 'react'
import { Table, Button, Select, Space, Modal, Form, InputNumber, Input, message, Drawer, Empty, Alert, Card } from 'antd'
import { EyeOutlined, CheckOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { PageHeader, StatusTag, SubmissionStatusConfig } from '@/components/common'
import type { Course, Experiment, Submission, PaginatedData } from '@/types'
import { getCourses } from '@/api/course'
import { getExperiments } from '@/api/experiment'
import { getSubmissions, gradeSubmission } from '@/api/experimentSubmission'
import { formatDateTime } from '@/utils/format'
import { FILTER_OPTIONS } from '@/utils/constants'

export default function TeacherGrading() {
  const [loading, setLoading] = useState(false)
  const [courses, setCourses] = useState<Course[]>([])
  const [experiments, setExperiments] = useState<Experiment[]>([])
  const [selectedCourseId, setSelectedCourseId] = useState<number | undefined>()
  const [selectedExperimentId, setSelectedExperimentId] = useState<number | undefined>()
  const [data, setData] = useState<PaginatedData<Submission>>({ list: [], total: 0, page: 1, page_size: 20 })
  const [pagination, setPagination] = useState({ page: 1, page_size: 20 })
  const [statusFilter, setStatusFilter] = useState<string>('')

  // 批改弹窗
  const [gradeModalVisible, setGradeModalVisible] = useState(false)
  const [gradingSubmission, setGradingSubmission] = useState<Submission | null>(null)
  const [gradeForm] = Form.useForm()

  // 详情抽屉
  const [detailVisible, setDetailVisible] = useState(false)
  const [viewingSubmission, setViewingSubmission] = useState<Submission | null>(null)

  // 获取课程列表
  useEffect(() => {
    const fetchCourses = async () => {
      try {
        const result = await getCourses({ page: 1, page_size: 100 })
        setCourses(result.list)
      } catch {
        // 错误由拦截器处理
      }
    }
    fetchCourses()
  }, [])

  // 获取实验列表
  useEffect(() => {
    if (!selectedCourseId) {
      setExperiments([])
      return
    }
    const fetchExperiments = async () => {
      try {
        const result = await getExperiments({ page: 1, page_size: 100 })
        setExperiments(result.list.filter((item) => item.course_id === selectedCourseId))
      } catch {
        // 错误由拦截器处理
      }
    }
    fetchExperiments()
  }, [selectedCourseId])

  // 获取提交列表
  const fetchData = useCallback(async () => {
    if (!selectedExperimentId) return
    setLoading(true)
    try {
      const result = await getSubmissions({
        experiment_id: selectedExperimentId,
        ...pagination,
        status: statusFilter || undefined,
      })
      setData(result)
    } catch {
      // 错误由拦截器处理
    } finally {
      setLoading(false)
    }
  }, [selectedExperimentId, pagination, statusFilter])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  const overview = {
    total: data.total,
    pending: data.list.filter((item) => item.status === 'pending').length,
    grading: data.list.filter((item) => item.status === 'grading').length,
    graded: data.list.filter((item) => item.status === 'graded').length,
  }

  // 打开批改弹窗
  const handleGrade = (submission: Submission) => {
    setGradingSubmission(submission)
    gradeForm.setFieldsValue({
      score: submission.score || 0,
      feedback: submission.feedback || '',
    })
    setGradeModalVisible(true)
  }

  // 提交批改
  const handleSubmitGrade = async (values: { score: number; feedback?: string }) => {
    if (!gradingSubmission) return
    try {
      await gradeSubmission(gradingSubmission.id, values)
      message.success('批改成功')
      setGradeModalVisible(false)
      fetchData()
    } catch {
      // 错误由拦截器处理
    }
  }

  // 查看详情
  const handleViewDetail = (submission: Submission) => {
    setViewingSubmission(submission)
    setDetailVisible(true)
  }

  // 表格列定义
  const columns: ColumnsType<Submission> = [
    { title: '学生姓名', dataIndex: 'student_name', key: 'student_name', width: 120 },
    {
      title: '提交时间',
      dataIndex: 'submitted_at',
      key: 'submitted_at',
      width: 180,
      render: (text) => formatDateTime(text),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status) => <StatusTag status={status} statusMap={SubmissionStatusConfig} />,
    },
    {
      title: '得分',
      dataIndex: 'score',
      key: 'score',
      width: 100,
      render: (score, record) => (
        record.status === 'graded' ? (
          <span className={score >= 60 ? 'text-success' : 'text-error'}>{score}分</span>
        ) : '-'
      ),
    },
    {
      title: '批改时间',
      dataIndex: 'graded_at',
      key: 'graded_at',
      width: 180,
      render: (text) => text ? formatDateTime(text) : '-',
    },
    {
      title: '操作',
      key: 'action',
      width: 160,
      render: (_, record) => (
        <Space>
          <Button
            type="link"
            size="small"
            icon={<EyeOutlined />}
            onClick={() => handleViewDetail(record)}
          >
            查看
          </Button>
          <Button
            type="link"
            size="small"
            icon={<CheckOutlined />}
            onClick={() => handleGrade(record)}
          >
            {record.status === 'graded' ? '修改' : '批改'}
          </Button>
        </Space>
      ),
    },
  ]

  return (
    <div className="space-y-6">
      <PageHeader title="批改作业" subtitle="批改学生提交的实验作业" />

      <Card className="overflow-hidden border-0 shadow-sm" styles={{ body: { padding: 0 } }}>
        <div className="grid gap-px bg-slate-200 xl:grid-cols-[1.35fr_repeat(4,minmax(0,1fr))]">
          <div className="bg-[linear-gradient(135deg,#0b2239_0%,#0f2744_55%,#113d67_100%)] px-6 py-6 text-white">
            <div className="text-xs uppercase tracking-[0.28em] text-cyan-200">Assessment Console</div>
            <div className="mt-3 text-2xl font-semibold">实验提交、检查点与人工批改统一处理</div>
            <p className="mt-3 text-sm leading-6 text-slate-200">
              评分页用于承接自动评测结果、手动批改和提交详情复核，构成实验平台的教师闭环。
            </p>
          </div>
          <div className="bg-white px-5 py-5"><div className="text-xs uppercase tracking-[0.2em] text-text-secondary">总提交</div><div className="mt-3 text-3xl font-semibold text-slate-900">{overview.total}</div></div>
          <div className="bg-white px-5 py-5"><div className="text-xs uppercase tracking-[0.2em] text-text-secondary">待处理</div><div className="mt-3 text-3xl font-semibold text-slate-900">{overview.pending}</div></div>
          <div className="bg-white px-5 py-5"><div className="text-xs uppercase tracking-[0.2em] text-text-secondary">批改中</div><div className="mt-3 text-3xl font-semibold text-slate-900">{overview.grading}</div></div>
          <div className="bg-white px-5 py-5"><div className="text-xs uppercase tracking-[0.2em] text-text-secondary">已完成</div><div className="mt-3 text-3xl font-semibold text-slate-900">{overview.graded}</div></div>
        </div>
      </Card>

      <Card className="border-0 shadow-sm">
        <Space wrap>
          <span className="text-text-secondary">课程:</span>
          <Select
            placeholder="请选择课程"
            value={selectedCourseId}
            onChange={(v) => {
              setSelectedCourseId(v)
              setSelectedExperimentId(undefined)
            }}
            style={{ width: 200 }}
            options={courses.map((c) => ({ label: c.title, value: c.id }))}
            allowClear
            notFoundContent={<Empty description="暂无课程" image={Empty.PRESENTED_IMAGE_SIMPLE} />}
          />
          <span className="text-text-secondary">实验:</span>
          <Select
            placeholder="请选择实验"
            value={selectedExperimentId}
            onChange={setSelectedExperimentId}
            style={{ width: 200 }}
            options={experiments.map((e) => ({ label: e.title, value: e.id }))}
            disabled={!selectedCourseId}
            allowClear
            notFoundContent={<Empty description="暂无实验" image={Empty.PRESENTED_IMAGE_SIMPLE} />}
          />
          <span className="text-text-secondary">状态:</span>
          <Select
            placeholder="全部"
            value={statusFilter}
            onChange={setStatusFilter}
            style={{ width: 120 }}
            options={[...FILTER_OPTIONS.GRADING_STATUS]}
            allowClear
          />
        </Space>
      </Card>

      <Card className="border-0 shadow-sm">
        {!selectedExperimentId ? (
          <Alert type="info" message="请先选择课程和实验后查看提交列表" showIcon className="text-left" />
        ) : (
          <Table
            columns={columns}
            dataSource={data.list}
            rowKey="id"
            loading={loading}
            locale={{
              emptyText: <Empty description="暂无提交，等待学生提交后再来批改" />,
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
        )}
      </Card>

      {/* 批改弹窗 */}
      <Modal
        title="批改作业"
        open={gradeModalVisible}
        onCancel={() => setGradeModalVisible(false)}
        footer={null}
      >
        <Form form={gradeForm} layout="vertical" onFinish={handleSubmitGrade} className="mt-4">
          <Form.Item label="学生" className="mb-2">
            <span>{gradingSubmission?.student_name}</span>
          </Form.Item>
          <Form.Item
            name="score"
            label="得分"
            rules={[{ required: true, message: '请输入得分' }]}
          >
            <InputNumber min={0} max={100} className="w-full" />
          </Form.Item>
          <Form.Item name="feedback" label="评语">
            <Input.TextArea placeholder="请输入评语（可选）" rows={4} />
          </Form.Item>
          <Form.Item className="mb-0 text-right">
            <Space>
              <Button onClick={() => setGradeModalVisible(false)}>取消</Button>
              <Button type="primary" htmlType="submit">提交</Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* 详情抽屉 */}
      <Drawer
        title="提交详情"
        open={detailVisible}
        onClose={() => setDetailVisible(false)}
        width={600}
      >
        {viewingSubmission && (
          <div>
            <div className="mb-4">
              <h4 className="text-text-secondary">学生</h4>
              <p>{viewingSubmission.student_name}</p>
            </div>
            <div className="mb-4">
              <h4 className="text-text-secondary">提交时间</h4>
              <p>{formatDateTime(viewingSubmission.submitted_at)}</p>
            </div>
            <div className="mb-4">
              <h4 className="text-text-secondary">提交内容</h4>
              <pre className="bg-gray-50 p-4 rounded text-sm overflow-auto">
                {JSON.stringify(viewingSubmission.content, null, 2) || '无'}
              </pre>
            </div>
            {viewingSubmission.file_url && (
              <div className="mb-4">
                <h4 className="text-text-secondary">附件</h4>
                <ul>
                  <li>
                    <a href={viewingSubmission.file_url} target="_blank" rel="noreferrer">
                      {viewingSubmission.file_url}
                    </a>
                  </li>
                </ul>
              </div>
            )}
            {viewingSubmission.status === 'graded' && (
              <>
                <div className="mb-4">
                  <h4 className="text-text-secondary">得分</h4>
                  <p className="text-lg font-semibold">{viewingSubmission.score}分</p>
                </div>
                {viewingSubmission.feedback && (
                  <div className="mb-4">
                    <h4 className="text-text-secondary">评语</h4>
                    <p>{viewingSubmission.feedback}</p>
                  </div>
                )}
              </>
            )}
          </div>
        )}
      </Drawer>
    </div>
  )
}
