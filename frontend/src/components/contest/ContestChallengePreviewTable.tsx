import { Table, Tag } from 'antd'
import type { ColumnsType } from 'antd/es/table'

import type { Challenge } from '@/types'
import { CategoryMap, DifficultyMap } from '@/types'
import { getRuntimeProfileLabel } from '@/domains/contest/jeopardy'

interface ContestChallengePreviewTableProps {
  challenges: Challenge[]
  reviewMode?: boolean
}

export default function ContestChallengePreviewTable({
  challenges,
  reviewMode = false,
}: ContestChallengePreviewTableProps) {
  const columns: ColumnsType<Challenge> = [
    {
      title: '题目名称',
      dataIndex: 'title',
      key: 'title',
    },
    {
      title: '知识主题',
      dataIndex: 'category',
      key: 'category',
      width: 140,
      render: (value: string) => CategoryMap[value] || value,
    },
    {
      title: '环境类型',
      dataIndex: 'runtime_profile',
      key: 'runtime_profile',
      width: 150,
      render: (value: string) => getRuntimeProfileLabel(value),
    },
    {
      title: '难度',
      dataIndex: 'difficulty',
      key: 'difficulty',
      width: 110,
      render: (value: string | number) => {
        const difficulty = DifficultyMap[value]
        return <Tag color={difficulty?.color}>{difficulty?.text || String(value)}</Tag>
      },
    },
    {
      title: '分值',
      key: 'points',
      width: 90,
      render: (_, record) => record.points ?? record.base_points,
    },
    {
      title: reviewMode ? '历史解出人数' : '解出人数',
      dataIndex: 'solve_count',
      key: 'solve_count',
      width: 110,
      render: (value: number | undefined) => value || 0,
    },
  ]

  return (
    <Table
      columns={columns}
      dataSource={challenges}
      rowKey="id"
      pagination={false}
      locale={{ emptyText: '当前暂无可展示的题目' }}
    />
  )
}
