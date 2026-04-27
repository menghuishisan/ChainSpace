import { TrophyFilled } from '@ant-design/icons'
import { Table, Tag } from 'antd'

import EmptyState from '@/components/common/EmptyState'
import type { RankedTeamView } from '@/domains/contest/spectate'

interface SpectateTeamStatusTableProps {
  teams: RankedTeamView[]
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

export default function SpectateTeamStatusTable({ teams }: SpectateTeamStatusTableProps) {
  return (
    <Table
      dataSource={teams}
      rowKey="team_id"
      pagination={false}
      columns={[
        {
          title: '排名',
          dataIndex: 'rank',
          key: 'rank',
          width: 70,
          render: renderRank,
        },
        {
          title: '队伍',
          dataIndex: 'team_name',
          key: 'team_name',
        },
        {
          title: '总分',
          dataIndex: 'total_score',
          key: 'total_score',
          width: 100,
          render: (value: number) => <span className="font-semibold text-primary">{value}</span>,
        },
        {
          title: '资源',
          dataIndex: 'resource_held',
          key: 'resource_held',
          width: 100,
        },
        {
          title: '状态',
          dataIndex: 'is_alive',
          key: 'is_alive',
          width: 100,
          render: (value: boolean) => <Tag color={value ? 'success' : 'default'}>{value ? '存活' : '失效'}</Tag>,
        },
      ]}
      locale={{
        emptyText: <EmptyState description="暂无队伍态势数据" />,
      }}
    />
  )
}
