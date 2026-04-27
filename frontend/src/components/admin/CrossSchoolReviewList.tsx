import { useState } from 'react'
import { Button, Input, Modal, Space, Table, Tag, message } from 'antd'
import { CheckOutlined, CloseOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'

import { useCrossSchoolReviews } from '@/hooks'
import { formatDateTime } from '@/utils/format'
import type { CrossSchoolApplication } from '@/types'

const { TextArea } = Input

function getApplicantName(record: CrossSchoolApplication) {
  return record.applicant?.real_name || record.applicant?.student_no || record.applicant?.phone || '-'
}

export default function CrossSchoolReviewList() {
  const { loading, data, review } = useCrossSchoolReviews()
  const [reviewModalVisible, setReviewModalVisible] = useState(false)
  const [reviewingItem, setReviewingItem] = useState<CrossSchoolApplication | null>(null)
  const [reviewAction, setReviewAction] = useState<'approve' | 'reject'>('approve')
  const [reviewComment, setReviewComment] = useState('')

  const handleReview = (record: CrossSchoolApplication, action: 'approve' | 'reject') => {
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

  const columns: ColumnsType<CrossSchoolApplication> = [
    { title: '申请类型', dataIndex: 'type', key: 'type', width: 100, render: (value: string) => (value === 'contest' ? '比赛' : '题目') },
    { title: '申请学校', key: 'from_school', width: 150, render: (_: unknown, record: CrossSchoolApplication) => record.from_school?.name || '-' },
    { title: '申请人', key: 'applicant', width: 100, render: (_: unknown, record: CrossSchoolApplication) => getApplicantName(record) },
    { title: '目标学校', key: 'to_school', width: 150, render: (_: unknown, record: CrossSchoolApplication) => record.to_school?.name || '-' },
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
      render: (_: unknown, record: CrossSchoolApplication) =>
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
          <p>申请类型：{reviewingItem?.type === 'contest' ? '比赛' : '题目'}</p>
          <p>申请学校：{reviewingItem?.from_school?.name}</p>
          <p>目标学校：{reviewingItem?.to_school?.name}</p>
          <p>申请人：{reviewingItem ? getApplicantName(reviewingItem) : '-'}</p>
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
