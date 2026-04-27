import { Empty, Table, Tag, Typography } from 'antd'

import { getRoundPhaseConfig, getRoundStatusConfig } from '@/domains/contest/battle'
import type { ContestRoundInfo } from '@/types/presentation'
import { formatDateTime } from '@/utils/format'

interface ContestRoundManagementTableProps {
  rounds: ContestRoundInfo[]
  loading?: boolean
}

export default function ContestRoundManagementTable({
  rounds,
  loading = false,
}: ContestRoundManagementTableProps) {
  return (
    <Table
      columns={[
        {
          title: '轮次',
          dataIndex: 'round_number',
          key: 'round_number',
          width: 140,
          render: (value: number, record: ContestRoundInfo) => (
            <div>
              <div className="font-medium text-slate-900">{`第 ${value} 轮`}</div>
              <Typography.Text type="secondary" className="text-xs">
                Round #{record.id}
              </Typography.Text>
            </div>
          ),
        },
        {
          title: '状态',
          dataIndex: 'status',
          key: 'status',
          width: 120,
          render: (value: string) => {
            const config = getRoundStatusConfig(value)
            return config ? <Tag color={config.color}>{config.text}</Tag> : value
          },
        },
        {
          title: '阶段',
          dataIndex: 'phase',
          key: 'phase',
          width: 120,
          render: (value?: string) => {
            const config = getRoundPhaseConfig(value)
            return config ? <Tag color={config.color}>{config.text}</Tag> : '-'
          },
        },
        {
          title: '开始时间',
          dataIndex: 'start_time',
          key: 'start_time',
          width: 180,
          render: (value?: string) => (value ? formatDateTime(value) : '-'),
        },
        {
          title: '结束时间',
          dataIndex: 'end_time',
          key: 'end_time',
          width: 180,
          render: (value?: string) => (value ? formatDateTime(value) : '-'),
        },
      ]}
      dataSource={rounds}
      rowKey="id"
      pagination={false}
      size="middle"
      loading={loading}
      scroll={{ x: 640 }}
      locale={{ emptyText: <Empty description="暂无轮次，可先在上方创建轮次计划" /> }}
    />
  )
}
