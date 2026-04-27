import { TrophyFilled } from '@ant-design/icons'
import { Card, Table, Tag } from 'antd'

import EmptyState from '@/components/common/EmptyState'
import type { AgentBattleSidebarProps } from '@/types/presentation'

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

export default function AgentBattleSidebar({
  scoreWeights,
  scoreboard,
  recentEvents,
}: AgentBattleSidebarProps) {
  return (
    <div className="space-y-6">
      <Card title="实时排行榜" styles={{ body: { padding: 0 } }} className="border-0 shadow-sm">
        <Table
          size="small"
          dataSource={scoreboard}
          rowKey="rank"
          pagination={false}
          columns={[
            {
              title: '排名',
              dataIndex: 'rank',
              key: 'rank',
              width: 60,
              render: renderRank,
            },
            { title: '队伍', dataIndex: 'team_name', key: 'team_name' },
            { title: '得分', dataIndex: 'total_score', key: 'total_score', width: 80 },
          ]}
          locale={{
            emptyText: <EmptyState description="暂无排行榜数据，等待比赛开始" />,
          }}
        />
      </Card>

      <Card title="评分结构" className="border-0 shadow-sm">
        <div className="space-y-3 text-sm">
          <div className="flex items-center justify-between rounded-xl bg-emerald-50 px-3 py-2">
            <span>资源控制</span>
            <Tag color="green">{scoreWeights.resource || 0}</Tag>
          </div>
          <div className="flex items-center justify-between rounded-xl bg-emerald-50 px-3 py-2">
            <span>攻击收益</span>
            <Tag color="green">{scoreWeights.attack || 0}</Tag>
          </div>
          <div className="flex items-center justify-between rounded-xl bg-sky-50 px-3 py-2">
            <span>防御保全</span>
            <Tag color="blue">{scoreWeights.defense || 0}</Tag>
          </div>
          <div className="flex items-center justify-between rounded-xl bg-fuchsia-50 px-3 py-2">
            <span>生存稳定</span>
            <Tag color="purple">{scoreWeights.survival || 0}</Tag>
          </div>
        </div>
      </Card>

      <Card title="最近事件" className="border-0 shadow-sm">
        {recentEvents.length > 0 ? (
          <div className="max-h-[520px] space-y-2 overflow-auto pr-1">
            {recentEvents.map((event) => (
              <div key={`${event.block}-${event.event_type}-${event.description || ''}`} className="rounded-xl bg-gray-50 p-3 text-sm">
                <div className="flex items-center justify-between gap-3">
                  <Tag color="blue">{event.event_type}</Tag>
                  <span className="text-text-secondary">区块 #{event.block}</span>
                </div>
                {(event.actor_team || event.target_team) && (
                  <div className="mt-2 flex flex-wrap gap-2">
                    {event.actor_team && <Tag color="purple">{event.actor_team}</Tag>}
                    {event.target_team && <Tag color="blue">{event.target_team}</Tag>}
                  </div>
                )}
                <div className="mt-2 text-text-secondary">
                  {event.description || '暂无事件描述'}
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className="py-4">
            <EmptyState image="simple" description="暂无最近事件" />
          </div>
        )}
      </Card>
    </div>
  )
}
