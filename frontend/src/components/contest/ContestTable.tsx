import { Button, Empty, Space, Table, Tag, Typography } from 'antd'
import { EditOutlined, EyeOutlined, RocketOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { ContestStatusConfig, StatusTag } from '@/components/common'
import type { Contest, PaginatedData } from '@/types'
import { ContestTypeMap } from '@/types'
import { formatDateTime } from '@/utils/format'

interface ContestTableProps {
  data: PaginatedData<Contest>
  loading: boolean
  emptyDescription?: string
  showEndTime?: boolean
  onView: (contest: Contest) => void
  onEdit: (contest: Contest) => void
  onPublish: (contestId: number) => void
  onPageChange: (page: number, pageSize: number) => void
}

export default function ContestTable({
  data,
  loading,
  emptyDescription,
  showEndTime = false,
  onView,
  onEdit,
  onPublish,
  onPageChange,
}: ContestTableProps) {
  const columns: ColumnsType<Contest> = [
    {
      title: '比赛名称',
      dataIndex: 'title',
      key: 'title',
      width: 280,
      render: (value: string, record: Contest) => (
        <div>
          <div className="font-medium text-slate-900">{value}</div>
          <Typography.Text type="secondary" className="text-xs">
            {record.type === 'agent_battle' ? '智能体对抗赛' : '解题赛'}
          </Typography.Text>
        </div>
      ),
    },
    {
      title: '比赛类型',
      dataIndex: 'type',
      key: 'type',
      width: 120,
      render: (value: Contest['type']) => <Tag color="blue">{ContestTypeMap[value]}</Tag>,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (value: string) => <StatusTag status={value} statusMap={ContestStatusConfig} />,
    },
    {
      title: '开始时间',
      dataIndex: 'start_time',
      key: 'start_time',
      width: 180,
      render: (value: string) => (
        <div className="text-sm text-slate-700">{formatDateTime(value, 'YYYY-MM-DD HH:mm')}</div>
      ),
    },
    ...(showEndTime ? [{
      title: '结束时间',
      dataIndex: 'end_time',
      key: 'end_time',
      width: 180,
      render: (value: string) => formatDateTime(value, 'YYYY-MM-DD HH:mm'),
    }] satisfies ColumnsType<Contest> : []),
    {
      title: '操作',
      key: 'action',
      width: 240,
      render: (_, record) => (
        <Space wrap size={[4, 4]}>
          <Button type="link" size="small" icon={<EyeOutlined />} onClick={() => onView(record)}>
            查看
          </Button>
          {record.status === 'draft' && (
            <>
              <Button type="link" size="small" icon={<EditOutlined />} onClick={() => onEdit(record)}>
                编辑
              </Button>
              <Button type="link" size="small" icon={<RocketOutlined />} onClick={() => onPublish(record.id)}>
                发布
              </Button>
            </>
          )}
        </Space>
      ),
    },
  ]

  return (
    <Table
      columns={columns}
      dataSource={data.list}
      rowKey="id"
      loading={loading}
      scroll={{ x: 960 }}
      locale={emptyDescription ? { emptyText: <Empty description={emptyDescription} /> } : undefined}
      pagination={{
        current: data.page,
        pageSize: data.page_size,
        total: data.total,
        showSizeChanger: true,
        showTotal: (total) => `共 ${total} 条`,
        onChange: onPageChange,
      }}
    />
  )
}
