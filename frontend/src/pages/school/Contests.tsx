/**
 * 学校侧比赛管理页面。
 * 学校可以维护本校正式比赛，并通过统一弹窗配置解题赛和对抗赛。
 */
import { useCallback, useEffect, useState } from 'react'
import { Button, Form, message, Space, Table, Tag } from 'antd'
import { DeleteOutlined, EditOutlined, EyeOutlined, PlusOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { useNavigate } from 'react-router-dom'

import { createContest, deleteContest, getContests, publishContest, updateContest } from '@/api/contest'
import { ContestManageModal } from '@/components/contest'
import { ContestStatusConfig, PageHeader, SearchFilter, StatusTag } from '@/components/common'
import {
  buildContestFormInitialValues,
  buildContestListQueryParams,
  buildContestSubmitData,
  CONTEST_FILTER_CONFIG,
  DEFAULT_CONTEST_LIST_FILTERS,
  DEFAULT_CONTEST_PAGINATION,
  normalizeContestFilters,
} from '@/domains/contest/management'
import type { Contest, PaginatedData } from '@/types'
import type { ContestFormValues } from '@/types/presentation'
import { ContestTypeMap } from '@/types'
import { formatDateTime } from '@/utils/format'

export default function SchoolContests() {
  const navigate = useNavigate()
  const [loading, setLoading] = useState(false)
  const [data, setData] = useState<PaginatedData<Contest>>({ list: [], total: 0, page: 1, page_size: 20 })
  const [filters, setFilters] = useState(DEFAULT_CONTEST_LIST_FILTERS)
  const [pagination, setPagination] = useState(DEFAULT_CONTEST_PAGINATION)
  const [modalVisible, setModalVisible] = useState(false)
  const [modalLoading, setModalLoading] = useState(false)
  const [editingContest, setEditingContest] = useState<Contest | null>(null)
  const [form] = Form.useForm<ContestFormValues>()

  const fetchData = useCallback(async () => {
    setLoading(true)
    try {
      const result = await getContests(buildContestListQueryParams(pagination, filters, 'school'))
      setData(result)
    } finally {
      setLoading(false)
    }
  }, [filters, pagination])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  const handleCreate = () => {
    setEditingContest(null)
    form.resetFields()
    setModalVisible(true)
  }

  const handleEdit = (record: Contest) => {
    setEditingContest(record)
    form.setFieldsValue(buildContestFormInitialValues(record))
    setModalVisible(true)
  }

  const handlePublish = async (id: number) => {
    try {
      await publishContest(id)
      message.success('比赛发布成功')
      fetchData()
    } catch {
      // 统一由请求拦截器处理
    }
  }

  const handleDelete = async (id: number) => {
    try {
      await deleteContest(id)
      message.success('比赛删除成功')
      fetchData()
    } catch {
      // 统一由请求拦截器处理
    }
  }

  const handleSubmit = async (values: ContestFormValues) => {
    setModalLoading(true)
    try {
      const submitData = buildContestSubmitData(values, 'school')
      if (editingContest) {
        await updateContest(editingContest.id, submitData)
        message.success('比赛更新成功')
      } else {
        await createContest(submitData)
        message.success('比赛创建成功')
      }

      setModalVisible(false)
      fetchData()
    } finally {
      setModalLoading(false)
    }
  }

  const columns: ColumnsType<Contest> = [
    {
      title: '比赛名称',
      dataIndex: 'title',
      key: 'title',
      width: 220,
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
      title: '参赛人数',
      dataIndex: 'participant_count',
      key: 'participant_count',
      width: 100,
      render: (value?: number) => value || 0,
    },
    {
      title: '开始时间',
      dataIndex: 'start_time',
      key: 'start_time',
      width: 180,
      render: (value: string) => formatDateTime(value, 'YYYY-MM-DD HH:mm'),
    },
    {
      title: '结束时间',
      dataIndex: 'end_time',
      key: 'end_time',
      width: 180,
      render: (value: string) => formatDateTime(value, 'YYYY-MM-DD HH:mm'),
    },
    {
      title: '操作',
      key: 'action',
      width: 260,
      render: (_, record) => (
        <Space>
          <Button type="link" size="small" icon={<EyeOutlined />} onClick={() => navigate(`/contest/${record.id}`)}>
            查看
          </Button>
          {record.status === 'draft' && (
            <>
              <Button type="link" size="small" icon={<EditOutlined />} onClick={() => handleEdit(record)}>
                编辑
              </Button>
              <Button type="link" size="small" onClick={() => handlePublish(record.id)}>
                发布
              </Button>
              <Button type="link" size="small" danger icon={<DeleteOutlined />} onClick={() => handleDelete(record.id)}>
                删除
              </Button>
            </>
          )}
        </Space>
      ),
    },
  ]

  return (
    <div>
      <PageHeader
        title="校内比赛"
        subtitle="管理本校解题赛与智能体博弈战"
        extra={(
          <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
            创建比赛
          </Button>
        )}
      />

      <SearchFilter
        filters={CONTEST_FILTER_CONFIG}
        values={filters}
        onChange={(values) => setFilters(normalizeContestFilters(values))}
        onSearch={() => setPagination((current) => ({ ...current, page: 1 }))}
        onReset={() => {
          setFilters(DEFAULT_CONTEST_LIST_FILTERS)
          setPagination(DEFAULT_CONTEST_PAGINATION)
        }}
      />

      <div className="card">
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
            onChange: (page, pageSize) => setPagination({ page, page_size: pageSize }),
          }}
        />
      </div>

      <ContestManageModal
        open={modalVisible}
        loading={modalLoading}
        editingContest={editingContest}
        form={form}
        onCancel={() => setModalVisible(false)}
        onSubmit={handleSubmit}
      />
    </div>
  )
}
