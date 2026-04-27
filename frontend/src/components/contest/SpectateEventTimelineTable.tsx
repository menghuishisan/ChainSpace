import { Empty, Table, Tag } from 'antd'

import type { AgentBattleEvent } from '@/types'
import { EVENT_COLOR_MAP } from '@/domains/contest/spectate'
import { formatDateTime } from '@/utils/format'

interface SpectateEventTimelineTableProps {
  events: AgentBattleEvent[]
}

export default function SpectateEventTimelineTable({ events }: SpectateEventTimelineTableProps) {
  return (
    <Table
      dataSource={events}
      rowKey={(event) => `${event.block}-${event.event_type}-${event.description || ''}`}
      size="small"
      pagination={false}
      columns={[
        {
          title: '区块',
          dataIndex: 'block',
          key: 'block',
          width: 80,
        },
        {
          title: '事件',
          key: 'event',
          render: (_: unknown, event: AgentBattleEvent) => (
            <div className="py-1">
              <div className="mb-1">
                <Tag color={EVENT_COLOR_MAP[event.event_type] || 'blue'}>{event.event_type}</Tag>
              </div>
              {(event.actor_team || event.target_team) && (
                <div className="mb-1 text-xs">
                  {event.actor_team && <Tag color="purple">{event.actor_team}</Tag>}
                  {event.target_team && <Tag color="blue">{event.target_team}</Tag>}
                </div>
              )}
              <div className="text-xs text-text-secondary">
                {event.description || '暂无事件描述'}
              </div>
              {event.time && (
                <div className="mt-1 text-[11px] text-gray-400">
                  {formatDateTime(event.time)}
                </div>
              )}
            </div>
          ),
        },
      ]}
      locale={{
        emptyText: (
          <div className="py-6">
            <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无关键事件" />
          </div>
        ),
      }}
    />
  )
}
