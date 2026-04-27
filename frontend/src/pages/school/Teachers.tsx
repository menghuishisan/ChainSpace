/**
 * 学校管理员 - 教师管理页面
 */
import { useCallback, useEffect, useState } from 'react'
import { Button, Form, Input, message, Modal, Popconfirm, Space, Table } from 'antd'
import { CheckOutlined, EditOutlined, PlusOutlined, StopOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'

import { PageHeader, SearchFilter, StatusTag, UserStatusConfig } from '@/components/common'
import { addTeacher, getTeachers, updateTeacher, updateTeacherStatus } from '@/api/school'
import type { PaginatedData, User } from '@/types'
import { FILTER_OPTIONS } from '@/utils/constants'
import { formatDateTime } from '@/utils/format'

export default function SchoolTeachers() {
  const [loading, setLoading] = useState(false)
  const [data, setData] = useState<PaginatedData<User>>({ list: [], total: 0, page: 1, page_size: 20 })
  const [filters, setFilters] = useState<Record<string, unknown>>({})
  const [pagination, setPagination] = useState({ page: 1, page_size: 20 })
  const [modalVisible, setModalVisible] = useState(false)
  const [modalLoading, setModalLoading] = useState(false)
  const [editingTeacher, setEditingTeacher] = useState<User | null>(null)
  const [form] = Form.useForm()

  const fetchData = useCallback(async () => {
    setLoading(true)
    try {
      const result = await getTeachers({ ...pagination, ...filters })
      setData(result)
    } catch {
      // 错误由请求层统一处理
    } finally {
      setLoading(false)
    }
  }, [filters, pagination])

  useEffect(() => {
    void fetchData()
  }, [fetchData])

  const handleCreate = () => {
    setEditingTeacher(null)
    form.resetFields()
    setModalVisible(true)
  }

  const handleEdit = (record: User) => {
    setEditingTeacher(record)
    form.setFieldsValue({
      real_name: record.real_name,
      phone: record.phone,
      email: record.email,
    })
    setModalVisible(true)
  }

  const handleStatusChange = async (record: User) => {
    const nextStatus = record.status === 'active' ? 'disabled' : 'active'
    try {
      await updateTeacherStatus(record.id, nextStatus)
      message.success(`${nextStatus === 'active' ? '启用' : '禁用'}成功`)
      void fetchData()
    } catch {
      // 错误由请求层统一处理
    }
  }

  const handleSubmit = async (values: {
    real_name: string
    phone?: string
    email?: string
    password?: string
  }) => {
    setModalLoading(true)
    try {
      if (editingTeacher) {
        await updateTeacher(editingTeacher.id, {
          real_name: values.real_name,
          phone: values.phone,
          email: values.email,
        })
        message.success('更新成功')
      } else {
        await addTeacher({
          real_name: values.real_name,
          phone: values.phone || '',
          email: values.email,
          password: values.password || '',
        })
        message.success('添加成功')
      }
      setModalVisible(false)
      void fetchData()
    } catch {
      // 错误由请求层统一处理
    } finally {
      setModalLoading(false)
    }
  }

  const columns: ColumnsType<User> = [
    { title: '姓名', dataIndex: 'real_name', key: 'real_name', width: 120 },
    { title: '手机号', dataIndex: 'phone', key: 'phone', width: 150, render: (value?: string) => value || '-' },
    { title: '邮箱', dataIndex: 'email', key: 'email', width: 220, render: (value?: string) => value || '-' },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: User['status']) => <StatusTag status={status} statusMap={UserStatusConfig} />,
    },
    {
      title: '最后登录',
      dataIndex: 'last_login_at',
      key: 'last_login_at',
      width: 180,
      render: (value?: string) => (value ? formatDateTime(value) : '-'),
    },
    {
      title: '操作',
      key: 'action',
      width: 180,
      render: (_: unknown, record: User) => (
        <Space>
          <Button type="link" size="small" icon={<EditOutlined />} onClick={() => handleEdit(record)}>
            编辑
          </Button>
          <Popconfirm
            title={`确定要${record.status === 'active' ? '禁用' : '启用'}该教师吗？`}
            onConfirm={() => void handleStatusChange(record)}
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
    { key: 'keyword', label: '搜索', type: 'input' as const, placeholder: '姓名/手机号/邮箱', width: 240 },
    { key: 'status', label: '状态', type: 'select' as const, options: [...FILTER_OPTIONS.USER_STATUS] },
  ]

  return (
    <div>
      <PageHeader
        title="教师管理"
        subtitle="管理本校教师账号"
        extra={(
          <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
            添加教师
          </Button>
        )}
      />

      <SearchFilter
        filters={filterConfig}
        values={filters}
        onChange={setFilters}
        onSearch={() => setPagination((prev) => ({ ...prev, page: 1 }))}
        onReset={() => {
          setFilters({})
          setPagination({ page: 1, page_size: 20 })
        }}
      />

      <div className="card">
        <Table
          rowKey="id"
          columns={columns}
          dataSource={data.list}
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

      <Modal
        title={editingTeacher ? '编辑教师' : '添加教师'}
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        footer={null}
      >
        <Form form={form} layout="vertical" onFinish={handleSubmit} className="mt-4">
          <Form.Item name="real_name" label="姓名" rules={[{ required: true, message: '请输入姓名' }]}>
            <Input placeholder="请输入教师姓名" />
          </Form.Item>
          <Form.Item
            name="phone"
            label="手机号"
            rules={[
              { required: true, message: '请输入手机号' },
              { len: 11, message: '手机号必须为 11 位' },
            ]}
          >
            <Input placeholder="请输入 11 位手机号" maxLength={11} />
          </Form.Item>
          <Form.Item name="email" label="邮箱" rules={[{ type: 'email', message: '邮箱格式不正确' }]}>
            <Input placeholder="请输入邮箱" />
          </Form.Item>
          {!editingTeacher ? (
            <Form.Item
              name="password"
              label="初始密码"
              rules={[
                { required: true, message: '请输入初始密码' },
                { min: 8, message: '密码长度至少 8 位' },
              ]}
            >
              <Input.Password placeholder="请输入初始密码" />
            </Form.Item>
          ) : null}
          <Form.Item className="mb-0 text-right">
            <Space>
              <Button onClick={() => setModalVisible(false)}>取消</Button>
              <Button type="primary" htmlType="submit" loading={modalLoading}>
                {editingTeacher ? '保存' : '添加'}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
