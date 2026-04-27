/**
 * 教师 - 讨论管理页面
 */
import { useState, useEffect, useCallback } from 'react'
import { Card, Table, Button, Tag, Space, Modal, Select, message, Badge } from 'antd'
import { DeleteOutlined, EyeOutlined, PushpinOutlined, CheckOutlined, StopOutlined } from '@ant-design/icons'
import { PageHeader, SearchFilter } from '@/components/common'
import type { Course, PaginatedData, Post } from '@/types'
import { getCourses } from '@/api/course'
import { getPosts, deletePost, pinPost, lockPost } from '@/api/discussion'
import { formatDateTime, formatRelativeTime } from '@/utils/format'
import { FILTER_OPTIONS } from '@/utils/constants'

export default function TeacherDiscussionManage() {
  const [loading, setLoading] = useState(false)
  const [courses, setCourses] = useState<Course[]>([])
  const [selectedCourseId, setSelectedCourseId] = useState<number | null>(null)
  const [data, setData] = useState<PaginatedData<Post>>({ list: [], total: 0, page: 1, page_size: 20 })
  const [filters, setFilters] = useState<Record<string, unknown>>({ keyword: '', status: '' })
  const [detailVisible, setDetailVisible] = useState(false)
  const [selectedPost, setSelectedPost] = useState<Post | null>(null)

  // 获取课程列表
  useEffect(() => {
    const fetchCourses = async () => {
      try {
        const result = await getCourses({ page: 1, page_size: 100 })
        setCourses(result.list)
        if (result.list.length > 0) setSelectedCourseId(result.list[0].id)
      } catch { /* */ }
    }
    fetchCourses()
  }, [])

  // 获取帖子列表
  const fetchData = useCallback(async (page = 1) => {
    if (!selectedCourseId) return
    setLoading(true)
    try {
      const result = await getPosts({ course_id: selectedCourseId, page, page_size: data.page_size, ...filters })
      setData(result)
    } catch { /* */ } finally { setLoading(false) }
  }, [selectedCourseId, filters, data.page_size])

  useEffect(() => { fetchData() }, [fetchData])

  // 查看帖子
  const handleView = (post: Post) => {
    setSelectedPost(post)
    setDetailVisible(true)
  }

  // 置顶/取消置顶
  const handlePin = async (post: Post) => {
    try {
      await pinPost(post.id)
      message.success(post.is_pinned ? '已取消置顶' : '已置顶')
      fetchData(data.page)
    } catch { /* */ }
  }

  // 关闭/开放讨论
  const handleClose = async (post: Post) => {
    try {
      await lockPost(post.id)
      message.success(post.is_locked ? '已开放讨论' : '已关闭讨论')
      fetchData(data.page)
    } catch { /* */ }
  }

  // 删除帖子
  const handleDelete = (id: number) => {
    Modal.confirm({
      title: '确认删除',
      content: '删除后将无法恢复，确定要删除吗？',
      onOk: async () => {
        try { await deletePost(id); message.success('删除成功'); fetchData(data.page) } catch { /* */ }
      }
    })
  }

  const columns = [
    { 
      title: '标题', 
      dataIndex: 'title', 
      key: 'title',
      render: (v: string, record: Post) => (
        <Space>
          {record.is_pinned && <Tag color="red">置顶</Tag>}
          {record.is_locked && <Tag>已关闭</Tag>}
          <span className="cursor-pointer hover:text-primary" onClick={() => handleView(record)}>{v}</span>
        </Space>
      )
    },
    { title: '作者', dataIndex: 'author_name', key: 'author_name', width: 100 },
    { title: '回复', dataIndex: 'reply_count', key: 'reply_count', width: 80, render: (v: number) => <Badge count={v} showZero color="#1890FF" /> },
    { title: '浏览', dataIndex: 'view_count', key: 'view_count', width: 80 },
    { title: '发布时间', dataIndex: 'created_at', key: 'created_at', width: 150, render: formatRelativeTime },
    { title: '最后回复', dataIndex: 'last_reply_at', key: 'last_reply_at', width: 150, render: (v?: string) => v ? formatRelativeTime(v) : '-' },
    { title: '操作', key: 'action', width: 180, render: (_: unknown, record: Post) => (
      <Space size="small">
        <Button type="link" size="small" icon={<EyeOutlined />} onClick={() => handleView(record)}>查看</Button>
        <Button type="link" size="small" icon={<PushpinOutlined />} onClick={() => handlePin(record)}>{record.is_pinned ? '取消置顶' : '置顶'}</Button>
        <Button type="link" size="small" icon={record.is_locked ? <CheckOutlined /> : <StopOutlined />} onClick={() => handleClose(record)}>
          {record.is_locked ? '开放' : '关闭'}
        </Button>
        <Button type="link" size="small" danger icon={<DeleteOutlined />} onClick={() => handleDelete(record.id)}>删除</Button>
      </Space>
    )}
  ]

  return (
    <div>
      <PageHeader title="讨论管理" subtitle="管理课程讨论区的帖子" />

      <Card>
        {/* 课程选择 */}
        <div className="mb-4">
          <span className="mr-2">选择课程：</span>
          <Select
            value={selectedCourseId}
            onChange={setSelectedCourseId}
            options={courses.map(c => ({ label: c.title, value: c.id }))}
            className="w-64"
            placeholder="请选择课程"
          />
        </div>

        <SearchFilter
          filters={[
            { type: 'input', key: 'keyword', placeholder: '搜索标题或内容' },
            { type: 'select', key: 'status', placeholder: '状态', options: [...FILTER_OPTIONS.DISCUSSION_STATUS] },
          ]}
          values={filters}
          onChange={setFilters}
          onSearch={() => fetchData(1)}
        />

        <Table 
          columns={columns} 
          dataSource={data.list} 
          rowKey="id" 
          loading={loading}
          pagination={{ current: data.page, pageSize: data.page_size, total: data.total, onChange: fetchData }}
        />
      </Card>

      {/* 帖子详情弹窗 */}
      <Modal
        title={selectedPost?.title}
        open={detailVisible}
        onCancel={() => setDetailVisible(false)}
        footer={null}
        width={700}
      >
        {selectedPost && (
          <div>
            <div className="mb-4 text-text-secondary text-sm">
              <span>{selectedPost.author_name}</span>
              <span className="mx-2">·</span>
              <span>{formatDateTime(selectedPost.created_at)}</span>
              <span className="mx-2">·</span>
              <span>{selectedPost.view_count} 次浏览</span>
            </div>
            <div className="prose max-w-none whitespace-pre-wrap">{selectedPost.content}</div>
          </div>
        )}
      </Modal>
    </div>
  )
}
