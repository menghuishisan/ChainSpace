import { Progress, Table, Tag, Typography } from 'antd'

import type { Scoreboard } from '@/types'
import type { ContestChallengeStat } from '@/types/presentation'
import { getChallengeSolveRate } from '@/domains/contest/monitor'

interface ContestChallengeStatsTableProps {
  challengeStats: ContestChallengeStat[]
  scoreboard: Scoreboard | null
  loading?: boolean
}

export default function ContestChallengeStatsTable({
  challengeStats,
  scoreboard,
  loading = false,
}: ContestChallengeStatsTableProps) {
  return (
    <Table
      columns={[
        {
          title: '题目',
          dataIndex: 'title',
          key: 'title',
          render: (value: string, record: ContestChallengeStat) => (
            <div>
              <div className="font-medium text-slate-900">{value}</div>
              <Typography.Text type="secondary" className="text-xs">
                Challenge #{record.id}
              </Typography.Text>
            </div>
          ),
        },
        { title: '分值', dataIndex: 'points', key: 'points', width: 80 },
        { title: '解题人数', dataIndex: 'solve_count', key: 'solve_count', width: 100 },
        {
          title: '解题率',
          key: 'rate',
          width: 150,
          render: (_: unknown, record: ContestChallengeStat) => (
            <Progress percent={getChallengeSolveRate(record, scoreboard)} size="small" />
          ),
        },
        {
          title: '一血',
          dataIndex: 'first_blood',
          key: 'first_blood',
          width: 140,
          render: (value?: string) => (value ? <Tag color="red">{value}</Tag> : '-'),
        },
      ]}
      dataSource={challengeStats}
      rowKey="id"
      pagination={false}
      size="middle"
      loading={loading}
      scroll={{ x: 720 }}
    />
  )
}
