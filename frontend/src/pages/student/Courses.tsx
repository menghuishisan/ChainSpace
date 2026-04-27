/**
 * 学生 - 我的课程页面
 */
import { useState, useEffect, useCallback } from 'react'
import { Card, Row, Col, Button, Progress, Empty, Spin, Modal, Input, message, Tag } from 'antd'
import { PlusOutlined, BookOutlined, TeamOutlined, ClockCircleOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { PageHeader } from '@/components/common'
import type { Course, PaginatedData } from '@/types'
import { getMyCourses, joinCourse } from '@/api/course'

export default function StudentCourses() {
  const navigate = useNavigate()
  const [loading, setLoading] = useState(false)
  const [data, setData] = useState<PaginatedData<Course & { progress_percent?: number }>>({ list: [], total: 0, page: 1, page_size: 20 })
  const [joinModalVisible, setJoinModalVisible] = useState(false)
  const [inviteCode, setInviteCode] = useState('')
  const [joining, setJoining] = useState(false)

  // 获取课程列表
  const fetchData = useCallback(async () => {
    setLoading(true)
    try {
      const result = await getMyCourses({ page: 1, page_size: 100 })
      setData(result)
    } catch { /* 错误由拦截器处理 */ } finally { setLoading(false) }
  }, [])

  useEffect(() => { fetchData() }, [fetchData])

  // 加入课程
  const handleJoinCourse = async () => {
    if (!inviteCode.trim()) { message.error('请输入邀请码'); return }
    setJoining(true)
    try {
      await joinCourse(inviteCode.trim())
      message.success('加入成功')
      setJoinModalVisible(false)
      setInviteCode('')
      fetchData()
    } catch { /* 错误由拦截器处理 */ } finally { setJoining(false) }
  }

  return (
    <div>
      <PageHeader
        title="我的课程"
        subtitle="查看已加入的所有课程"
        extra={<Button type="primary" icon={<PlusOutlined />} onClick={() => setJoinModalVisible(true)}>加入课程</Button>}
      />

      <Spin spinning={loading}>
        {data.list.length === 0 ? (
          <Card><Empty description="暂未加入任何课程" image={Empty.PRESENTED_IMAGE_SIMPLE}><Button type="primary" onClick={() => setJoinModalVisible(true)}>通过邀请码加入</Button></Empty></Card>
        ) : (
          <Row gutter={[24, 24]}>
            {data.list.map((course) => (
              <Col xs={24} sm={12} lg={8} xl={6} key={course.id}>
                <Card
                  hoverable
                  onClick={() => navigate(`/student/courses/${course.id}`)}
                  cover={course.cover ? (<img alt={course.title} src={course.cover} className="h-40 object-cover" />) : (<div className="h-40 bg-gradient-to-r from-blue-400 to-blue-600 flex items-center justify-center"><BookOutlined className="text-5xl text-white" /></div>)}
                  styles={{ header: { background: 'linear-gradient(90deg, rgba(24,144,255,0.08), rgba(82,196,26,0.08))' } }}
                >
                  <Card.Meta
                    title={course.title}
                    description={
                      <div>
                        <p className="text-text-secondary text-sm line-clamp-2 h-10 mb-2">{course.description || '暂无描述'}</p>
                        <div className="flex items-center text-text-secondary text-xs mb-2"><TeamOutlined className="mr-1" /><span>教师: {course.teacher_name || '-'}</span></div>
                        <div className="flex items-center justify-between text-xs text-text-secondary">
                          <div className="flex items-center"><ClockCircleOutlined className="mr-1" />学习进度</div>
                          <Tag color={course.progress && course.progress >= 60 ? 'green' : 'blue'}>{course.progress || 0}%</Tag>
                        </div>
                        <Progress percent={course.progress || 0} size="small" />
                      </div>
                    }
                  />
                </Card>
              </Col>
            ))}
          </Row>
        )}
      </Spin>

      {/* 加入课程弹窗 */}
      <Modal title="加入课程" open={joinModalVisible} onCancel={() => setJoinModalVisible(false)} onOk={handleJoinCourse} confirmLoading={joining} okText="加入">
        <div className="py-4"><p className="mb-4 text-text-secondary">请输入教师提供的课程邀请码</p><Input placeholder="请输入邀请码" value={inviteCode} onChange={(e) => setInviteCode(e.target.value)} onPressEnter={handleJoinCourse} size="large" /></div>
      </Modal>
    </div>
  )
}
