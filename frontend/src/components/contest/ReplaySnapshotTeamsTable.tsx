import { TrophyFilled } from '@ant-design/icons'
import { Empty, Table, Tag } from 'antd'

import type { FinalRankItem } from '@/domains/contest/replay'

interface ReplaySnapshotTeamsTableProps {
  finalRank: FinalRankItem[]
}

function renderRank(rank: number) {
  if (rank === 1) {
    return <TrophyFilled style={{ color: '#faad14' }} />
  }
  if (rank === 2) {
    return <TrophyFilled style={{ color: '#bfbfbf' }} />
  }
  if (rank === 3) {
    return <TrophyFilled style={{ color: '#d48806' }} />
  }

  return rank
}

export default function ReplaySnapshotTeamsTable({ finalRank }: ReplaySnapshotTeamsTableProps) {
  return (
    <Table
      dataSource={finalRank}
      rowKey="team_id"
      pagination={false}
      columns={[
        {
          title: '排名',
          dataIndex: 'rank',
          key: 'rank',
          width: 80,
          render: renderRank,
        },
        { title: '队伍', dataIndex: 'team_name', key: 'team_name' },
        {
          title: '总分',
          dataIndex: 'total_score',
          key: 'total_score',
          width: 100,
          render: (value: number) => <Tag color="blue">{value}</Tag>,
        },
      ]}
      locale={{ emptyText: <Empty description="暂无最终排名" /> }}
    />
  )
}
