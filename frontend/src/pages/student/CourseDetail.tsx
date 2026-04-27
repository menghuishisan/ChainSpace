import { useCallback, useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { Card, Progress, Spin, Tabs, message } from 'antd'

import { PageHeader } from '@/components/common'
import { StudentCourseExperimentsTab, StudentCourseLearnTab } from '@/components/course'
import { usePersistedTab } from '@/hooks'
import { getChapters, getCourse, getCourseProgress } from '@/api/course'
import type { Chapter, Course, CourseProgress } from '@/types'

export default function StudentCourseDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const courseId = parseInt(id || '0', 10)

  const [loading, setLoading] = useState(true)
  const [course, setCourse] = useState<Course | null>(null)
  const [chapters, setChapters] = useState<Chapter[]>([])
  const [progress, setProgress] = useState<CourseProgress | null>(null)
  const [activeTab, setActiveTab] = usePersistedTab(
    `student_course_${courseId}`,
    'learn',
    ['learn', 'experiments', 'discussion'],
  )

  const fetchCourse = useCallback(async () => {
    if (!courseId) {
      return
    }

    setLoading(true)
    try {
      const [courseData, chaptersData, progressData] = await Promise.all([
        getCourse(courseId),
        getChapters(courseId),
        getCourseProgress(courseId),
      ])
      setCourse(courseData)
      setChapters(chaptersData)
      setProgress(progressData)
    } catch {
      message.error('获取课程信息失败')
      navigate('/student/courses')
    } finally {
      setLoading(false)
    }
  }, [courseId, navigate])

  useEffect(() => {
    void fetchCourse()
  }, [fetchCourse])

  if (loading) {
    return <div className="flex h-64 items-center justify-center"><Spin size="large" /></div>
  }

  if (!course) {
    return null
  }

  return (
    <div>
      <PageHeader title={course.title} subtitle={course.description} showBack backPath="/student/courses">
        {progress && (
          <div className="mt-4 max-w-md">
            <div className="mb-1 flex justify-between text-sm">
              <span className="text-text-secondary">学习进度</span>
              <span>{progress.progress_percent}%</span>
            </div>
            <Progress percent={progress.progress_percent} showInfo={false} />
            <div className="mt-1 flex justify-between text-xs text-text-secondary">
              <span>资料: {progress.completed_materials}/{progress.total_materials}</span>
              <span>实验: {progress.completed_experiments || 0}/{progress.total_experiments || 0}</span>
            </div>
          </div>
        )}
      </PageHeader>

      <Card>
        <Tabs
          activeKey={activeTab}
          onChange={setActiveTab}
          items={[
            { key: 'learn', label: '课程学习', children: <StudentCourseLearnTab courseId={courseId} chapters={chapters} /> },
            { key: 'experiments', label: '课程实验', children: <StudentCourseExperimentsTab courseId={courseId} /> },
          ]}
        />
      </Card>
    </div>
  )
}
