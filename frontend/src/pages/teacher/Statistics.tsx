/**
 * 教师 - 教学统计页面
 * 聚合课程相关数据展示统计信息
 */
import { useState, useEffect, useCallback } from 'react'
import { Card, Row, Col, Statistic, Select, Table, Spin } from 'antd'
import { TeamOutlined, ExperimentOutlined, BookOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { PageHeader } from '@/components/common'
import type { Course, Experiment } from '@/types'
import type { TeacherCourseStatRow, TeacherStatisticsSummary } from '@/types/presentation'
import { ExperimentTypeMap } from '@/types'
import { getCourses, getCourseStudents } from '@/api/course'
import { getExperiments } from '@/api/experiment'

export default function TeacherStatisticsPage() {
  const [loading, setLoading] = useState(false)
  const [courses, setCourses] = useState<Course[]>([])
  const [selectedCourseId, setSelectedCourseId] = useState<number | undefined>()
  const [stats, setStats] = useState<TeacherStatisticsSummary>({
    courseCount: 0,
    studentCount: 0,
    experimentCount: 0,
  })
  const [courseStats, setCourseStats] = useState<TeacherCourseStatRow[]>([])
  const [experiments, setExperiments] = useState<Experiment[]>([])

  // 获取课程列表并聚合统计
  useEffect(() => {
    const fetchData = async () => {
      setLoading(true)
      try {
        const coursesRes = await getCourses({ page: 1, page_size: 100 })
        setCourses(coursesRes.list)
        
        // 聚合每个课程的学生和实验数量
        let totalStudents = 0
        let totalExperiments = 0
        const statsRows: TeacherCourseStatRow[] = []
        
        for (const course of coursesRes.list) {
          const [studentsRes, experimentsRes] = await Promise.all([
            getCourseStudents(course.id, { page: 1, page_size: 1 }),
            getExperiments({ page: 1, page_size: 100 }),
          ])
          const experimentTotal = experimentsRes.list.filter((item) => item.course_id === course.id).length
          totalStudents += studentsRes.total
          totalExperiments += experimentTotal
          statsRows.push({
            courseId: course.id,
            courseName: course.title,
            studentCount: studentsRes.total,
            experimentCount: experimentTotal,
          })
        }
        
        setCourseStats(statsRows)
        setStats({
          courseCount: coursesRes.total,
          studentCount: totalStudents,
          experimentCount: totalExperiments,
        })
        
        if (coursesRes.list.length > 0) {
          setSelectedCourseId(coursesRes.list[0].id)
        }
      } catch {
        // 错误由拦截器处理
      } finally {
        setLoading(false)
      }
    }
    fetchData()
  }, [])

  // 获取选中课程的实验列表
  const fetchExperiments = useCallback(async () => {
    if (!selectedCourseId) return
    try {
      const res = await getExperiments({ page: 1, page_size: 100 })
      setExperiments(res.list.filter((item) => item.course_id === selectedCourseId))
    } catch {
      // 错误由拦截器处理
    }
  }, [selectedCourseId])

  useEffect(() => {
    fetchExperiments()
  }, [fetchExperiments])

  // 课程统计表格列
  const courseColumns: ColumnsType<TeacherCourseStatRow> = [
    { title: '课程名称', dataIndex: 'courseName', key: 'courseName' },
    { title: '学生人数', dataIndex: 'studentCount', key: 'studentCount', width: 120 },
    { title: '实验数量', dataIndex: 'experimentCount', key: 'experimentCount', width: 120 },
  ]

  // 实验列表表格列
  const experimentColumns: ColumnsType<Experiment> = [
    { title: '实验名称', dataIndex: 'title', key: 'title' },
    { title: '实验类型', dataIndex: 'type', key: 'type', width: 120, render: (type) => ExperimentTypeMap[type as keyof typeof ExperimentTypeMap] || type },
    { title: '满分', dataIndex: 'max_score', key: 'max_score', width: 80 },
    { title: '状态', dataIndex: 'status', key: 'status', width: 100, render: (status) => status === 'draft' ? '草稿' : status === 'published' ? '已发布' : status },
  ]

  return (
    <div>
      <PageHeader title="教学统计" subtitle="查看课程教学数据统计" />

      {/* 课程选择 */}
      <div className="card mb-4">
        <span className="text-text-secondary mr-2">选择课程:</span>
        <Select
          placeholder="全部课程"
          value={selectedCourseId}
          onChange={setSelectedCourseId}
          style={{ width: 300 }}
          allowClear
          options={[
            { label: '全部课程', value: undefined },
            ...courses.map((c) => ({ label: c.title, value: c.id }))
          ]}
        />
      </div>

      <Spin spinning={loading}>
        {/* 统计概览 */}
        <Row gutter={[24, 24]} className="mb-6">
          <Col xs={24} sm={12} lg={8}>
            <Card>
              <Statistic
                title="课程数量"
                value={stats.courseCount}
                prefix={<BookOutlined style={{ color: '#722ED1' }} />}
                valueStyle={{ color: '#722ED1' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={8}>
            <Card>
              <Statistic
                title="学生总数"
                value={stats.studentCount}
                prefix={<TeamOutlined style={{ color: '#1890FF' }} />}
                valueStyle={{ color: '#1890FF' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={8}>
            <Card>
              <Statistic
                title="实验数量"
                value={stats.experimentCount}
                prefix={<ExperimentOutlined style={{ color: '#FA8C16' }} />}
                valueStyle={{ color: '#FA8C16' }}
              />
            </Card>
          </Col>
        </Row>

        {/* 课程统计明细 */}
        <Card title="课程统计明细" className="mb-6">
          <Table
            columns={courseColumns}
            dataSource={courseStats}
            rowKey="courseId"
            pagination={false}
            size="middle"
          />
        </Card>

        {/* 选中课程的实验列表 */}
        {selectedCourseId && (
          <Card title={`实验列表 - ${courses.find(c => c.id === selectedCourseId)?.title || ''}`}>
            <Table
              columns={experimentColumns}
              dataSource={experiments}
              rowKey="id"
              pagination={false}
              size="middle"
            />
          </Card>
        )}
      </Spin>
    </div>
  )
}
