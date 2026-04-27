/**
 * 教师 - 我的课程页面
 */
import { useState, useEffect, useCallback } from 'react'
import { Card, Row, Col, Button, Tag, Empty, Spin, Dropdown, Modal, message } from 'antd'
import { PlusOutlined, EyeOutlined, EditOutlined, MoreOutlined, TeamOutlined, BookOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { PageHeader, SearchFilter, StatusTag, CourseStatusConfig } from '@/components/common'
import type { Course, PaginatedData } from '@/types'
import { getCourses, updateCourseStatus, deleteCourse } from '@/api/course'
import { FILTER_OPTIONS } from '@/utils/constants'

export default function TeacherCourses() {
  const navigate = useNavigate()
  const [loading, setLoading] = useState(false)
  const [data, setData] = useState<PaginatedData<Course>>({ list: [], total: 0, page: 1, page_size: 12 })
  const [filters, setFilters] = useState<Record<string, unknown>>({ keyword: '', status: '' })

  // 获取课程列表
  const fetchData = useCallback(async () => {
    setLoading(true)
    try {
      // 构建查询参数，只传递有值的筛选条件
      const queryParams: Record<string, unknown> = { page: 1, page_size: 100 }
      if (filters.keyword && String(filters.keyword).trim()) {
        queryParams.keyword = String(filters.keyword).trim()
      }
      if (filters.status && String(filters.status).trim()) {
        queryParams.status = filters.status
      }
      const result = await getCourses(queryParams as Parameters<typeof getCourses>[0])
      setData(result)
    } catch {
      // 错误由拦截器处理
    } finally {
      setLoading(false)
    }
  }, [filters])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  // 发布课程
  const handlePublish = async (course: Course) => {
    try {
      await updateCourseStatus(course.id, 'published')
      message.success('发布成功')
      fetchData()
    } catch {
      // 错误由拦截器处理
    }
  }

  // 归档课程
  const handleArchive = async (course: Course) => {
    Modal.confirm({
      title: '确认归档',
      content: '归档后学生将无法访问该课程，确定要归档吗？',
      onOk: async () => {
        try {
          await updateCourseStatus(course.id, 'archived')
          message.success('归档成功')
          fetchData()
        } catch {
          // 错误由拦截器处理
        }
      },
    })
  }

  // 删除课程
  const handleDelete = async (course: Course) => {
    Modal.confirm({
      title: '确认删除',
      content: '删除后无法恢复，确定要删除该课程吗？',
      okType: 'danger',
      onOk: async () => {
        try {
          await deleteCourse(course.id)
          message.success('删除成功')
          fetchData()
        } catch {
          // 错误由拦截器处理
        }
      },
    })
  }

  // 获取操作菜单
  const getMenuItems = (course: Course) => {
    const items = [
      { key: 'view', label: '查看详情', icon: <EyeOutlined /> },
      { key: 'edit', label: '编辑课程', icon: <EditOutlined /> },
    ]
    
    if (course.status === 'draft') {
      items.push({ key: 'publish', label: '发布课程', icon: <BookOutlined /> })
      items.push({ key: 'delete', label: '删除课程', icon: <BookOutlined /> })
    } else if (course.status === 'published') {
      items.push({ key: 'archive', label: '归档课程', icon: <BookOutlined /> })
    }
    
    return items
  }

  // 处理菜单点击
  const handleMenuClick = (key: string, course: Course) => {
    switch (key) {
      case 'view':
        navigate(`/teacher/courses/${course.id}`)
        break
      case 'edit':
        navigate(`/teacher/courses/${course.id}?edit=true`)
        break
      case 'publish':
        handlePublish(course)
        break
      case 'archive':
        handleArchive(course)
        break
      case 'delete':
        handleDelete(course)
        break
    }
  }

  // 筛选配置
  const filterConfig = [
    { key: 'keyword', label: '关键词', type: 'input' as const, placeholder: '搜索课程名称' },
    { key: 'status', label: '状态', type: 'select' as const, options: [...FILTER_OPTIONS.COURSE_STATUS] },
  ]

  return (
    <div>
      <PageHeader
        title="我的课程"
        subtitle="管理您创建的所有课程"
        extra={
          <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/teacher/courses/create')}>
            创建课程
          </Button>
        }
      />

      <SearchFilter
        filters={filterConfig}
        values={filters}
        onChange={setFilters}
        onSearch={fetchData}
        onReset={() => setFilters({ keyword: '', status: '' })}
      />

      <Spin spinning={loading}>
        {data.list.length === 0 ? (
          <Card>
            <Empty
              description="暂无课程"
              image={Empty.PRESENTED_IMAGE_SIMPLE}
            >
              <Button type="primary" onClick={() => navigate('/teacher/courses/create')}>
                创建第一个课程
              </Button>
            </Empty>
          </Card>
        ) : (
          <Row gutter={[24, 24]}>
            {data.list.map((course) => (
              <Col xs={24} sm={12} lg={8} xl={6} key={course.id}>
                <Card
                  hoverable
                  styles={{ header: { background: 'linear-gradient(90deg, rgba(24,144,255,0.08), rgba(82,196,26,0.08))' } }}
                  cover={
                    course.cover ? (
                      <img
                        alt={course.title}
                        src={course.cover}
                        className="h-40 object-cover"
                      />
                    ) : (
                      <div className="h-40 bg-gradient-to-r from-blue-400 to-blue-600 flex items-center justify-center">
                        <BookOutlined className="text-5xl text-white" />
                      </div>
                    )
                  }
                  actions={[
                    <Button 
                      type="link" 
                      key="view" 
                      onClick={() => navigate(`/teacher/courses/${course.id}`)}
                    >
                      进入课程
                    </Button>,
                    <Dropdown
                      key="more"
                      menu={{
                        items: getMenuItems(course),
                        onClick: ({ key }) => handleMenuClick(key, course),
                      }}
                    >
                      <MoreOutlined />
                    </Dropdown>,
                  ]}
                >
                  <Card.Meta
                    title={
                      <div className="flex items-center justify-between">
                        <span className="truncate">{course.title}</span>
                        <StatusTag status={course.status} statusMap={CourseStatusConfig} />
                      </div>
                    }
                    description={
                      <div>
                        <p className="text-text-secondary text-sm line-clamp-2 h-10 mb-2">
                          {course.description || '暂无描述'}
                        </p>
                        <div className="flex items-center text-text-secondary text-xs">
                          <TeamOutlined className="mr-1" />
                          <span>{course.student_count || 0} 名学生</span>
                          <span className="mx-2">|</span>
                          <span>{course.chapter_count || 0} 个章节</span>
                        </div>
                        <div className="mt-2 text-text-secondary text-xs">
                          邀请码：<Tag>{course.invite_code}</Tag>
                        </div>
                      </div>
                    }
                  />
                </Card>
              </Col>
            ))}
          </Row>
        )}
      </Spin>
    </div>
  )
}
