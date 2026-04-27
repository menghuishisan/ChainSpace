import { useCallback, useEffect, useState } from 'react'
import { Button, Form, Input, Modal, Popconfirm, Space, Table, message } from 'antd'
import { CheckOutlined, EditOutlined, PlusOutlined, StopOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'

import { getSchools, createSchool, updateSchool, updateSchoolStatus } from '@/api/admin'
import { PageHeader, SearchFilter, StatusTag, UserStatusConfig } from '@/components/common'
import type { CreateSchoolRequest, PaginatedData, School } from '@/types'
import { FILTER_OPTIONS } from '@/utils/constants'
import { formatDateTime } from '@/utils/format'

export default function Schools() {
  const [loading, setLoading] = useState(false)
  const [data, setData] = useState<PaginatedData<School>>({ list: [], total: 0, page: 1, page_size: 20 })
  const [filters, setFilters] = useState<{ keyword?: string; status?: string }>({})
  const [pagination, setPagination] = useState({ page: 1, page_size: 20 })
  const [modalVisible, setModalVisible] = useState(false)
  const [modalLoading, setModalLoading] = useState(false)
  const [editingSchool, setEditingSchool] = useState<School | null>(null)
  const [form] = Form.useForm<CreateSchoolRequest>()

  const fetchData = useCallback(async () => {
    setLoading(true)
    try {
      const result = await getSchools({ ...pagination, ...filters })
      setData(result)
    } finally {
      setLoading(false)
    }
  }, [filters, pagination])

  useEffect(() => {
    void fetchData()
  }, [fetchData])

  const handleCreate = () => {
    setEditingSchool(null)
    form.resetFields()
    setModalVisible(true)
  }

  const handleEdit = (record: School) => {
    setEditingSchool(record)
    form.setFieldsValue({
      name: record.name,
      logo_url: record.logo,
      contact_email: record.email,
      contact_phone: record.phone,
    })
    setModalVisible(true)
  }

  const handleStatusChange = async (record: School) => {
    const nextStatus = record.status === 'active' ? 'disabled' : 'active'
    await updateSchoolStatus(record.id, nextStatus)
    message.success(nextStatus === 'active' ? '学校已启用' : '学校已禁用')
    await fetchData()
  }

  const handleSubmit = async (values: CreateSchoolRequest) => {
    setModalLoading(true)
    try {
      if (editingSchool) {
        await updateSchool(editingSchool.id, values)
        message.success('学校信息已更新')
      } else {
        await createSchool(values)
        message.success('学校创建成功')
      }
      setModalVisible(false)
      await fetchData()
    } finally {
      setModalLoading(false)
    }
  }

  const columns: ColumnsType<School> = [
    { title: '学校名称', dataIndex: 'name', key: 'name', width: 220 },
    { title: '联系邮箱', dataIndex: 'email', key: 'email', width: 220, render: (value: string) => value || '-' },
    { title: '联系电话', dataIndex: 'phone', key: 'phone', width: 160, render: (value: string) => value || '-' },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status) => <StatusTag status={status} statusMap={UserStatusConfig} />,
    },
    { title: '教师数', dataIndex: 'teacher_count', key: 'teacher_count', width: 100, render: (value) => value || 0 },
    { title: '学生数', dataIndex: 'student_count', key: 'student_count', width: 100, render: (value) => value || 0 },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (value) => formatDateTime(value),
    },
    {
      title: '操作',
      key: 'action',
      width: 180,
      render: (_, record) => (
        <Space>
          <Button type="link" size="small" icon={<EditOutlined />} onClick={() => handleEdit(record)}>
            编辑
          </Button>
          <Popconfirm
            title={record.status === 'active' ? '确定禁用该学校吗？' : '确定启用该学校吗？'}
            onConfirm={() => void handleStatusChange(record)}
            okText="确定"
            cancelText="取消"
          >
            <Button
              type="link"
              size="small"
              danger={record.status === 'active'}
              icon={record.status === 'active' ? <StopOutlined /> : <CheckOutlined />}
            >
              {record.status === 'active' ? '禁用' : '启用'}
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  const filterConfig = [
    { key: 'keyword', label: '搜索', type: 'input' as const, placeholder: '学校名称', width: 220 },
    { key: 'status', label: '状态', type: 'select' as const, options: [...FILTER_OPTIONS.USER_STATUS] },
  ]

  return (
    <div>
      <PageHeader
        title="学校管理"
        subtitle="管理平台中的学校及其管理员"
        extra={(
          <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
            添加学校
          </Button>
        )}
      />

      <SearchFilter
        filters={filterConfig}
        values={filters}
        onChange={(values) => setFilters(values as typeof filters)}
        onSearch={() => setPagination((current) => ({ ...current, page: 1 }))}
        onReset={() => {
          setFilters({})
          setPagination({ page: 1, page_size: 20 })
        }}
      />

      <div className="card">
        <Table
          columns={columns}
          dataSource={data.list}
          rowKey="id"
          loading={loading}
          scroll={{ x: 1200 }}
          pagination={{
            current: data.page,
            pageSize: data.page_size,
            total: data.total,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 条`,
            onChange: (page, pageSize) => setPagination({ page, page_size: pageSize }),
          }}
        />
      </div>

      <Modal
        title={editingSchool ? '编辑学校' : '添加学校'}
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        footer={null}
        width={640}
      >
        <Form form={form} layout="vertical" onFinish={(values) => void handleSubmit(values)} className="mt-4">
          <Form.Item name="name" label="学校名称" rules={[{ required: true, message: '请输入学校名称' }]}>
            <Input placeholder="请输入学校名称" />
          </Form.Item>

          <Form.Item name="logo_url" label="学校 Logo">
            <Input placeholder="请输入 Logo 地址" />
          </Form.Item>

          <Form.Item name="contact_email" label="联系邮箱">
            <Input placeholder="请输入联系邮箱" />
          </Form.Item>

          <Form.Item name="contact_phone" label="联系电话">
            <Input placeholder="请输入联系电话" />
          </Form.Item>

          {!editingSchool && (
            <>
              <div className="mb-4 mt-6 border-t pt-4">
                <h4 className="font-medium text-text-primary">学校管理员信息</h4>
              </div>

              <Form.Item
                name="admin_phone"
                label="管理员手机号"
                rules={[
                  { required: true, message: '请输入管理员手机号' },
                  { pattern: /^1[3-9]\d{9}$/, message: '请输入正确的手机号' },
                ]}
              >
                <Input placeholder="请输入管理员手机号" />
              </Form.Item>

              <Form.Item
                name="admin_name"
                label="管理员姓名"
                rules={[{ required: true, message: '请输入管理员姓名' }]}
              >
                <Input placeholder="请输入管理员姓名" />
              </Form.Item>

              <Form.Item
                name="admin_password"
                label="初始密码"
                rules={[
                  { required: true, message: '请输入初始密码' },
                  { min: 8, message: '密码长度至少 8 位' },
                ]}
              >
                <Input.Password placeholder="请输入初始密码" />
              </Form.Item>
            </>
          )}

          <Form.Item className="mb-0 text-right">
            <Space>
              <Button onClick={() => setModalVisible(false)}>取消</Button>
              <Button type="primary" htmlType="submit" loading={modalLoading}>
                {editingSchool ? '保存' : '创建'}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
