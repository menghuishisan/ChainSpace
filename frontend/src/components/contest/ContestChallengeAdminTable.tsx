import { Button, Table, Tag } from 'antd'
import { DeleteOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'

import type { ContestChallenge } from '@/types'
import { CategoryMap } from '@/types'
import { getRuntimeProfileLabel } from '@/domains/contest/jeopardy'

interface ContestChallengeAdminTableProps {
  challenges: ContestChallenge[]
  onRemove: (challengeId: number) => void
}

export default function ContestChallengeAdminTable({
  challenges,
  onRemove,
}: ContestChallengeAdminTableProps) {
  const columns: ColumnsType<ContestChallenge> = [
    {
      title: '题目名称',
      dataIndex: ['challenge', 'title'],
      key: 'title',
    },
    {
      title: '知识主题',
      dataIndex: ['challenge', 'category'],
      key: 'category',
      width: 140,
      render: (value: string) => CategoryMap[value] || value,
    },
    {
      title: '环境类型',
      dataIndex: ['challenge', 'runtime_profile'],
      key: 'runtime_profile',
      width: 150,
      render: (value: string) => getRuntimeProfileLabel(value),
    },
    {
      title: '分值',
      dataIndex: 'points',
      key: 'points',
      width: 90,
    },
    {
      title: '可见',
      dataIndex: 'is_visible',
      key: 'is_visible',
      width: 90,
      render: (value: boolean) => (value ? <Tag color="green">是</Tag> : <Tag color="red">否</Tag>),
    },
    {
      title: '操作',
      key: 'action',
      width: 100,
      render: (_, record) => (
        <Button
          type="link"
          danger
          size="small"
          icon={<DeleteOutlined />}
          onClick={() => onRemove(record.challenge_id)}
        >
          移除
        </Button>
      ),
    },
  ]

  return (
    <Table
      columns={columns}
      dataSource={challenges}
      rowKey="id"
      size="small"
      pagination={false}
      locale={{ emptyText: '当前比赛暂无题目，请点击“添加题目”补充内容。' }}
    />
  )
}
