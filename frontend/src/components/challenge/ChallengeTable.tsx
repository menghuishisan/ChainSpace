import { Button, Space, Table, Tag } from 'antd'
import { DeleteOutlined, EditOutlined, EyeOutlined, GlobalOutlined, ShareAltOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import type { ChallengeTableProps } from '@/types/presentation'
import { CategoryMap, ChallengeRuntimeProfileMap, DifficultyMap } from '@/types'

export default function ChallengeTable({
  data,
  loading,
  onEdit,
  onView,
  onDelete,
  onRequestPublish,
  canManage,
  onPageChange,
}: ChallengeTableProps) {
  const columns: ColumnsType<typeof data.list[number]> = [
    { title: '题目名称', dataIndex: 'title', key: 'title', width: 220 },
    {
      title: '分类',
      dataIndex: 'category',
      key: 'category',
      width: 110,
      render: (value: string) => CategoryMap[value] || value,
    },
    {
      title: '环境类型',
      dataIndex: 'runtime_profile',
      key: 'runtime_profile',
      width: 140,
      render: (value: keyof typeof ChallengeRuntimeProfileMap) => <Tag color="blue">{ChallengeRuntimeProfileMap[value] || value}</Tag>,
    },
    {
      title: '难度',
      dataIndex: 'difficulty',
      key: 'difficulty',
      width: 100,
      render: (value) => {
        const config = DifficultyMap[value]
        return config ? <Tag color={config.color}>{config.text}</Tag> : value
      },
    },
    { title: '基础分', dataIndex: 'base_points', key: 'base_points', width: 90 },
    {
      title: '环境',
      key: 'environment',
      width: 100,
      render: (_, record) => (
        record.challenge_orchestration.needs_environment
          ? <Tag color="blue">环境题</Tag>
          : <Tag>静态题</Tag>
      ),
    },
    {
      title: '来源',
      dataIndex: 'source_type',
      key: 'source_type',
      width: 120,
      render: (value) => {
        const sourceMap: Record<string, string> = {
          preset: '预设',
          auto_converted: '自动转化',
          user_created: '用户创建',
        }
        return sourceMap[value || ''] || value || '-'
      },
    },
    {
      title: '可见性',
      dataIndex: 'is_public',
      key: 'is_public',
      width: 100,
      render: (value: boolean) => value ? <Tag color="green">公开</Tag> : <Tag>私有</Tag>,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (value) => {
        const statusMap: Record<string, { text: string; color: string }> = {
          draft: { text: '草稿', color: 'default' },
          active: { text: '启用', color: 'green' },
          archived: { text: '归档', color: 'default' },
        }
        const config = statusMap[value || '']
        return <Tag color={config?.color}>{config?.text || value || '-'}</Tag>
      },
    },
    {
      title: '操作',
      key: 'action',
      width: 260,
      render: (_, record) => {
        const manageable = canManage ? canManage(record) : true

        return (
          <Space size="small">
            {onView && (
              <Button type="link" size="small" icon={<EyeOutlined />} onClick={() => onView(record)}>
                查看
              </Button>
            )}
            {onEdit && manageable && (
              <Button type="link" size="small" icon={<EditOutlined />} onClick={() => onEdit(record)}>
                编辑
              </Button>
            )}
            {onRequestPublish && manageable && !record.is_public && (
              <Button
                type="link"
                size="small"
                icon={onDelete ? <ShareAltOutlined /> : <GlobalOutlined />}
                onClick={() => onRequestPublish(record)}
              >
                申请公开
              </Button>
            )}
            {onDelete && manageable && (
              <Button type="link" size="small" danger icon={<DeleteOutlined />} onClick={() => onDelete(record)}>
                删除
              </Button>
            )}
          </Space>
        )
      },
    },
  ]

  return (
    <Table
      columns={columns}
      dataSource={data.list}
      rowKey="id"
      loading={loading}
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
