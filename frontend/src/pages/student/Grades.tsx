import { useState, useEffect, useCallback } from 'react'
import { Table, Card, Row, Col, Statistic, Tag, Select, Space, Progress } from 'antd'
import { TrophyOutlined, CheckCircleOutlined, ClockCircleOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { PageHeader, StatusTag, SubmissionStatusConfig } from '@/components/common'
import type { Course, Experiment } from '@/types'
import { getMyCourses } from '@/api/course'
import { getStudentExperiments } from '@/api/experiment'

interface GradeRow {
  experiment_id: number
  title: string
  max_score: number
  status: string
  score?: number
}

export default function StudentGrades() {
  const [loading, setLoading] = useState(false)
  const [courses, setCourses] = useState<Course[]>([])
  const [selectedCourseId, setSelectedCourseId] = useState<number | undefined>()
  const [grades, setGrades] = useState<GradeRow[]>([])

  useEffect(() => {
    getMyCourses({ page: 1, page_size: 100 })
      .then((res: { list: Course[] }) => {
        setCourses(res.list)
        if (res.list.length > 0) {
          setSelectedCourseId(res.list[0].id)
        }
      })
      .catch(() => {})
  }, [])

  const fetchGrades = useCallback(async () => {
    if (!selectedCourseId) {
      return
    }

    setLoading(true)
    try {
      const result = await getStudentExperiments({ page: 1, page_size: 100 })
      const gradeData = result.list
        .filter((exp: Experiment) => exp.course_id === selectedCourseId)
        .map((exp: Experiment) => ({
          experiment_id: exp.id,
          title: exp.title,
          max_score: exp.max_score || 100,
          status: exp.my_status || 'pending',
          score: exp.my_score,
        }))
      setGrades(gradeData)
    } finally {
      setLoading(false)
    }
  }, [selectedCourseId])

  useEffect(() => {
    void fetchGrades()
  }, [fetchGrades])

  const gradedRows = grades.filter((item) => item.score !== undefined)
  const stats = {
    total: grades.length,
    completed: grades.filter((item) => item.status === 'graded').length,
    pending: grades.filter((item) => item.status === 'grading' || item.status === 'pending').length,
    avgScore: gradedRows.reduce((sum, item) => sum + (item.score || 0), 0) / (gradedRows.length || 1),
  }

  const columns: ColumnsType<GradeRow> = [
    { title: '实验名称', dataIndex: 'title', key: 'title' },
    { title: '满分', dataIndex: 'max_score', key: 'max_score', width: 80, render: (score) => `${score}分` },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status) => status === 'pending' ? <Tag>未提交</Tag> : <StatusTag status={status} statusMap={SubmissionStatusConfig} />,
    },
    {
      title: '得分',
      dataIndex: 'score',
      key: 'score',
      width: 100,
      render: (score, row) => score !== undefined ? (
        <span className={score >= row.max_score * 0.6 ? 'text-success' : 'text-error'}>
          {score}分
        </span>
      ) : '-',
    },
    {
      title: '得分率',
      key: 'rate',
      width: 120,
      render: (_, row) => row.score !== undefined
        ? <Progress percent={Math.round((row.score / row.max_score) * 100)} size="small" />
        : '-',
    },
  ]

  return (
    <div>
      <PageHeader title="我的成绩" subtitle="查看各课程的实验成绩" />

      <div className="card mb-4">
        <Space>
          <span className="text-text-secondary">选择课程:</span>
          <Select
            placeholder="请选择课程"
            value={selectedCourseId}
            onChange={setSelectedCourseId}
            style={{ width: 300 }}
            options={courses.map((course) => ({ label: course.title, value: course.id }))}
          />
        </Space>
      </div>

      <Row gutter={[24, 24]} className="mb-6">
        <Col xs={24} sm={12} lg={6}>
          <Card><Statistic title="实验总数" value={stats.total} prefix={<TrophyOutlined style={{ color: '#1890FF' }} />} /></Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card><Statistic title="已完成" value={stats.completed} prefix={<CheckCircleOutlined style={{ color: '#52C41A' }} />} /></Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card><Statistic title="待处理" value={stats.pending} prefix={<ClockCircleOutlined style={{ color: '#FA8C16' }} />} /></Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="平均分"
              value={stats.avgScore.toFixed(1)}
              suffix="分"
              valueStyle={{ color: stats.avgScore >= 60 ? '#52C41A' : '#FF4D4F' }}
            />
          </Card>
        </Col>
      </Row>

      <Card title="实验成绩">
        <Table columns={columns} dataSource={grades} rowKey="experiment_id" loading={loading} pagination={false} />
      </Card>
    </div>
  )
}
