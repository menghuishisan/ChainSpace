/**
 * 讨论区页面
 */
import { useState, useEffect, useCallback } from 'react'
import { useParams } from 'react-router-dom'
import { Card, List, Button, Modal, Form, Input, Avatar, message, Spin, Empty, Space, Pagination } from 'antd'
import { PlusOutlined, UserOutlined, MessageOutlined } from '@ant-design/icons'
import { PageHeader } from '@/components/common'
import type { PaginatedData, Post, Reply } from '@/types'
import { getPosts, createPost, getReplies, createReply } from '@/api/discussion'
import { formatRelativeTime } from '@/utils/format'

export default function Discussion() {
  const { courseId } = useParams<{ courseId: string }>()
  const cid = parseInt(courseId || '0')

  const [loading, setLoading] = useState(false)
  const [posts, setPosts] = useState<PaginatedData<Post>>({ list: [], total: 0, page: 1, page_size: 20 })
  const [selectedPost, setSelectedPost] = useState<Post | null>(null)
  const [replies, setReplies] = useState<Reply[]>([])
  const [repliesLoading, setRepliesLoading] = useState(false)
  const [postModalVisible, setPostModalVisible] = useState(false)
  const [replyContent, setReplyContent] = useState('')
  const [form] = Form.useForm()

  // 获取帖子列表
  const fetchPosts = useCallback(async (page = 1) => {
    setLoading(true)
    try {
      const result = await getPosts({ course_id: cid, page, page_size: 20 })
      setPosts(result)
    } catch { /* */ } finally { setLoading(false) }
  }, [cid])

  useEffect(() => { fetchPosts() }, [fetchPosts])

  // 获取回复列表
  const fetchReplies = async (postId: number) => {
    setRepliesLoading(true)
    try {
      const result = await getReplies(postId)
      setReplies(result.list)
    } catch { /* */ } finally { setRepliesLoading(false) }
  }

  // 选择帖子
  const handleSelectPost = (post: Post) => {
    setSelectedPost(post)
    fetchReplies(post.id)
  }

  // 发布帖子
  const handleCreatePost = async (values: { title: string; content: string }) => {
    try {
      await createPost({ course_id: cid, ...values })
      message.success('发布成功')
      setPostModalVisible(false)
      form.resetFields()
      fetchPosts()
    } catch { /* */ }
  }

  // 发布回复
  const handleCreateReply = async () => {
    if (!selectedPost || !replyContent.trim()) return
    try {
      await createReply(selectedPost.id, { content: replyContent })
      message.success('回复成功')
      setReplyContent('')
      fetchReplies(selectedPost.id)
    } catch { /* */ }
  }

  return (
    <div>
      <PageHeader
        title="课程讨论区"
        subtitle="与老师和同学交流学习问题"
        showBack
        extra={<Button type="primary" icon={<PlusOutlined />} onClick={() => setPostModalVisible(true)}>发帖</Button>}
      />

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* 帖子列表 */}
        <Card title="帖子列表" className="lg:col-span-1">
          <Spin spinning={loading}>
            {posts.list.length === 0 ? <Empty description="暂无帖子" /> : (
              <List
                dataSource={posts.list}
                renderItem={(item) => (
                  <List.Item
                    className={`cursor-pointer hover:bg-gray-50 ${selectedPost?.id === item.id ? 'bg-blue-50' : ''}`}
                    onClick={() => handleSelectPost(item)}
                  >
                    <List.Item.Meta
                      avatar={<Avatar src={item.author_avatar} icon={<UserOutlined />} />}
                      title={item.title}
                      description={<div className="text-xs text-text-secondary"><span>{item.author_name}</span><span className="ml-2">{formatRelativeTime(item.created_at)}</span><span className="ml-2"><MessageOutlined /> {item.reply_count || 0}</span></div>}
                    />
                  </List.Item>
                )}
              />
            )}
            {posts.total > 20 && <Pagination current={posts.page} total={posts.total} pageSize={20} onChange={(p) => fetchPosts(p)} className="mt-4" size="small" />}
          </Spin>
        </Card>

        {/* 帖子详情和回复 */}
        <Card title={selectedPost?.title || '选择帖子查看详情'} className="lg:col-span-2">
          {selectedPost ? (
            <div>
              {/* 帖子内容 */}
              <div className="mb-6 pb-6 border-b">
                <div className="flex items-center mb-4">
                  <Avatar src={selectedPost.author_avatar} icon={<UserOutlined />} />
                  <span className="ml-2 font-medium">{selectedPost.author_name}</span>
                  <span className="ml-2 text-text-secondary text-sm">{formatRelativeTime(selectedPost.created_at)}</span>
                </div>
                <div className="whitespace-pre-wrap">{selectedPost.content}</div>
              </div>

              {/* 回复列表 */}
              <h4 className="mb-4">回复 ({replies.length})</h4>
              <Spin spinning={repliesLoading}>
                {replies.length === 0 ? <div className="text-text-secondary text-center py-4">暂无回复</div> : (
                  <List
                    dataSource={replies}
                    renderItem={(item) => (
                      <List.Item>
                        <List.Item.Meta
                          avatar={<Avatar src={item.author_avatar} icon={<UserOutlined />} size="small" />}
                          title={<span className="text-sm">{item.author_name} <span className="text-text-secondary font-normal">{formatRelativeTime(item.created_at)}</span></span>}
                          description={item.content}
                        />
                      </List.Item>
                    )}
                  />
                )}
              </Spin>

              {/* 回复输入 */}
              <div className="mt-4 pt-4 border-t">
                <Input.TextArea placeholder="写下你的回复..." value={replyContent} onChange={(e) => setReplyContent(e.target.value)} rows={3} />
                <Button type="primary" className="mt-2" onClick={handleCreateReply} disabled={!replyContent.trim()}>发表回复</Button>
              </div>
            </div>
          ) : (
            <div className="text-center text-text-secondary py-12">请从左侧选择一个帖子</div>
          )}
        </Card>
      </div>

      {/* 发帖弹窗 */}
      <Modal title="发布帖子" open={postModalVisible} onCancel={() => setPostModalVisible(false)} footer={null}>
        <Form form={form} layout="vertical" onFinish={handleCreatePost} className="mt-4">
          <Form.Item name="title" label="标题" rules={[{ required: true, message: '请输入标题' }]}><Input placeholder="请输入帖子标题" /></Form.Item>
          <Form.Item name="content" label="内容" rules={[{ required: true, message: '请输入内容' }]}><Input.TextArea placeholder="请输入帖子内容" rows={6} /></Form.Item>
          <Form.Item className="mb-0 text-right"><Space><Button onClick={() => setPostModalVisible(false)}>取消</Button><Button type="primary" htmlType="submit">发布</Button></Space></Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
