import { useState } from 'react'
import { Button, Input, Modal, Space, Table, Tag, message } from 'antd'
import { CheckOutlined, CloseOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'

import { useChallengePublishReviews } from '@/hooks'
import { formatDateTime } from '@/utils/format'
import type { ChallengePublishApplication } from '@/types'

const { TextArea } = Input

export default function ChallengePublishReviewList() {
  const { loading, data, review } = useChallengePublishReviews()
  const [reviewModalVisible, setReviewModalVisible] = useState(false)
  const [reviewingItem, setReviewingItem] = useState<ChallengePublishApplication | null>(null)
  const [reviewAction, setReviewAction] = useState<'approve' | 'reject'>('approve')
  const [reviewComment, setReviewComment] = useState('')

  const handleReview = (record: ChallengePublishApplication, action: 'approve' | 'reject') => {
    setReviewingItem(record)
    setReviewAction(action)
    setReviewComment('')
    setReviewModalVisible(true)
  }

  const handleSubmitReview = async () => {
    if (!reviewingItem) {
      return
    }

    try {
      await review(reviewingItem.id, reviewAction, reviewComment)
      message.success(reviewAction === 'approve' ? '审核通过' : '已拒绝')
      setReviewModalVisible(false)
    } catch {
      // 交给请求层统一处理
    }
  }

  const columns: ColumnsType<ChallengePublishApplication> = [
    { title: '题目名称', dataIndex: 'challenge_title', key: 'challenge_title', width: 200 },
    { title: '申请人', dataIndex: 'applicant_name', key: 'applicant_name', width: 120 },
    { title: '申请说明', dataIndex: 'reason', key: 'reason', ellipsis: true },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => {
        const map = {
          pending: { text: '待审核', color: 'processing' },
          approved: { text: '已通过', color: 'success' },
          rejected: { text: '已拒绝', color: 'error' },
        }
        const config = map[status as keyof typeof map]
        return config ? <Tag color={config.color}>{config.text}</Tag> : status
      },
    },
    {
      title: '申请时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (text: string) => formatDateTime(text),
    },
    {
      title: '操作',
      key: 'action',
      width: 160,
      render: (_: unknown, record: ChallengePublishApplication) =>
        record.status === 'pending' && (
          <Space>
            <Button type="link" size="small" icon={<CheckOutlined />} onClick={() => handleReview(record, 'approve')}>
              通过
            </Button>
            <Button type="link" size="small" danger icon={<CloseOutlined />} onClick={() => handleReview(record, 'reject')}>
              拒绝
            </Button>
          </Space>
        ),
    },
  ]

  return (
    <>
      <Table columns={columns} dataSource={data.list} rowKey="id" loading={loading} pagination={false} />
      <Modal
        title={reviewAction === 'approve' ? '审核通过' : '拒绝申请'}
        open={reviewModalVisible}
        onCancel={() => setReviewModalVisible(false)}
        onOk={() => void handleSubmitReview()}
        okText="确定"
        cancelText="取消"
      >
        <div className="mb-4">
          <p>题目名称：{reviewingItem?.challenge_title}</p>
          <p>申请人：{reviewingItem?.applicant_name}</p>
          <p>申请说明：{reviewingItem?.reason || '无'}</p>
        </div>
        <TextArea
          placeholder="请输入审核意见（可选）"
          value={reviewComment}
          onChange={(event) => setReviewComment(event.target.value)}
          rows={4}
        />
      </Modal>
    </>
  )
}
