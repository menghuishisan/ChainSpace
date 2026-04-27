import { Table } from 'antd'
import type { ColumnsType } from 'antd/es/table'

import type { ContestScore } from '@/types'

interface ContestRankingTableProps {
  data: ContestScore[]
  size?: 'small' | 'middle' | 'large'
  emptyText: string
}

export default function ContestRankingTable({
  data,
  size = 'small',
  emptyText,
}: ContestRankingTableProps) {
  const columns: ColumnsType<ContestScore> = [
    { title: '排名', dataIndex: 'rank', key: 'rank', width: 80 },
    { title: '队伍', dataIndex: 'team_name', key: 'team_name' },
    { title: '总分', dataIndex: 'total_score', key: 'total_score', width: 120 },
  ]

  return (
    <Table
      columns={columns}
      dataSource={data}
      rowKey={(record) => `${record.rank}-${record.team_id || record.user_id || record.team_name || 'contest'}`}
      pagination={false}
      size={size}
      locale={{ emptyText }}
    />
  )
}
