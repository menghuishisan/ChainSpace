import { useCallback, useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { Card, Button, Popconfirm, Space, Spin, Tabs, Tooltip, message } from 'antd'
import { CopyOutlined, ReloadOutlined } from '@ant-design/icons'

import {
  TeacherCourseContentTab,
  TeacherCourseExperimentsTab,
  TeacherCourseStudentsTab,
} from '@/components/course'
import { CourseStatusConfig, PageHeader, StatusTag } from '@/components/common'
import { usePersistedTab } from '@/hooks'
import { getChapters, getCourse, resetInviteCode } from '@/api/course'
import type { Chapter, Course } from '@/types'

export default function TeacherCourseDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const courseId = parseInt(id || '0', 10)

  const [loading, setLoading] = useState(true)
  const [course, setCourse] = useState<Course | null>(null)
  const [chapters, setChapters] = useState<Chapter[]>([])
  const [activeTab, setActiveTab] = usePersistedTab(
    `teacher_course_${courseId}`,
    'content',
    ['content', 'students', 'discussion'],
  )

  const fetchCourse = useCallback(async () => {
    if (!courseId) {
      return
    }

    setLoading(true)
    try {
      const data = await getCourse(courseId)
      setCourse(data)
    } catch {
      message.error('获取课程信息失败')
      navigate('/teacher/courses')
    } finally {
      setLoading(false)
    }
  }, [courseId, navigate])

  const fetchChapters = useCallback(async () => {
    if (!courseId) {
      return
    }

    try {
      const data = await getChapters(courseId)
      setChapters(data || [])
    } catch {
      // 交给请求层统一处理
    }
  }, [courseId])

  useEffect(() => {
    void fetchCourse()
    void fetchChapters()
  }, [fetchCourse, fetchChapters])

  const handleResetInviteCode = async () => {
    if (!courseId) {
      return
    }

    try {
      const result = await resetInviteCode(courseId)
      setCourse((previous) => (previous ? { ...previous, invite_code: result.invite_code } : null))
      message.success('邀请码已重置')
    } catch {
      // 交给请求层统一处理
    }
  }

  const handleCopyInviteCode = async () => {
    if (!course?.invite_code) {
      return
    }

    await navigator.clipboard.writeText(course.invite_code)
    message.success('邀请码已复制')
  }

  if (loading) {
    return <div className="flex h-64 items-center justify-center"><Spin size="large" /></div>
  }

  if (!course) {
    return null
  }

  return (
    <div>
      <PageHeader
        title={course.title}
        subtitle={course.description}
        showBack
        backPath="/teacher/courses"
        tags={<StatusTag status={course.status} statusMap={CourseStatusConfig} />}
        extra={(
          <Space>
            <Tooltip title="点击复制">
              <Button icon={<CopyOutlined />} onClick={() => void handleCopyInviteCode()}>
                邀请码: {course.invite_code}
              </Button>
            </Tooltip>
            <Popconfirm title="确定要重置邀请码吗？" onConfirm={() => void handleResetInviteCode()}>
              <Button icon={<ReloadOutlined />}>重置邀请码</Button>
            </Popconfirm>
          </Space>
        )}
      />

      <Card>
        <Tabs
          activeKey={activeTab}
          onChange={setActiveTab}
          items={[
            {
              key: 'content',
              label: '课程内容',
              children: <TeacherCourseContentTab courseId={courseId} chapters={chapters} onRefresh={fetchChapters} />,
            },
            {
              key: 'experiments',
              label: '实验列表',
              children: <TeacherCourseExperimentsTab courseId={courseId} />,
            },
            {
              key: 'students',
              label: '学生管理',
              children: <TeacherCourseStudentsTab courseId={courseId} />,
            },
          ]}
        />
      </Card>
    </div>
  )
}
