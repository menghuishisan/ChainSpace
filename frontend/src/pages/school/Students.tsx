/**
 * 学校管理员 - 学生管理页面
 */
import { useCallback, useEffect, useMemo, useState } from 'react'
import { Button, Form, Input, message, Modal, Popconfirm, Select, Space, Table, Upload } from 'antd'
import { CheckOutlined, DownloadOutlined, EditOutlined, PlusOutlined, StopOutlined, UploadOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'

import { PageHeader, SearchFilter, StatusTag, UserStatusConfig } from '@/components/common'
import {
  addStudent,
  downloadStudentTemplate,
  getClasses,
  getStudents,
  importStudents,
  updateStudent,
  updateStudentStatus,
} from '@/api/school'
import type { Class, PaginatedData, User } from '@/types'
import { FILTER_OPTIONS } from '@/utils/constants'
import { formatDateTime } from '@/utils/format'

export default function SchoolStudents() {
  const [loading, setLoading] = useState(false)
  const [data, setData] = useState<PaginatedData<User>>({ list: [], total: 0, page: 1, page_size: 20 })
  const [filters, setFilters] = useState<Record<string, unknown>>({})
  const [pagination, setPagination] = useState({ page: 1, page_size: 20 })
  const [classes, setClasses] = useState<Class[]>([])
  const [modalVisible, setModalVisible] = useState(false)
  const [modalLoading, setModalLoading] = useState(false)
  const [editingStudent, setEditingStudent] = useState<User | null>(null)
  const [importModalVisible, setImportModalVisible] = useState(false)
  const [form] = Form.useForm()

  const classOptions = useMemo(
    () => classes.map((item) => ({ label: item.name, value: item.id })),
    [classes],
  )

  const fetchData = useCallback(async () => {
    setLoading(true)
    try {
      const result = await getStudents({ ...pagination, ...filters })
      setData(result)
    } catch {
      // 错误由请求层统一处理
    } finally {
      setLoading(false)
    }
  }, [filters, pagination])

  const fetchClasses = useCallback(async () => {
    try {
      const result = await getClasses()
      setClasses(result.list || [])
    } catch {
      // 错误由请求层统一处理
    }
  }, [])

  useEffect(() => {
    void fetchData()
  }, [fetchData])

  useEffect(() => {
    void fetchClasses()
  }, [fetchClasses])

  const handleCreate = () => {
    setEditingStudent(null)
    form.resetFields()
    setModalVisible(true)
  }

  const handleEdit = (record: User) => {
    setEditingStudent(record)
    form.setFieldsValue({
      real_name: record.real_name,
      phone: record.phone,
      email: record.email,
      student_no: record.student_no,
      class_id: record.class_id,
    })
    setModalVisible(true)
  }

  const handleStatusChange = async (record: User) => {
    const nextStatus = record.status === 'active' ? 'disabled' : 'active'
    try {
      await updateStudentStatus(record.id, nextStatus)
      message.success(`${nextStatus === 'active' ? '启用' : '禁用'}成功`)
      void fetchData()
    } catch {
      // 错误由请求层统一处理
    }
  }

  const handleSubmit = async (values: {
    real_name: string
    phone: string
    email?: string
    password?: string
    student_no: string
    class_id?: number
  }) => {
    setModalLoading(true)
    try {
      if (editingStudent) {
        await updateStudent(editingStudent.id, {
          real_name: values.real_name,
          phone: values.phone,
          email: values.email,
          student_no: values.student_no,
          class_id: values.class_id,
        })
        message.success('更新成功')
      } else {
        await addStudent({
          real_name: values.real_name,
          phone: values.phone,
          email: values.email,
          password: values.password || '',
          student_no: values.student_no,
          class_id: values.class_id,
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

  const handleDownloadTemplate = async () => {
    try {
      const blob = await downloadStudentTemplate()
      const url = window.URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.download = '学生导入模板.csv'
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      window.URL.revokeObjectURL(url)
    } catch {
      message.error('模板下载失败')
    }
  }

  const handleImport = async (file: File) => {
    try {
      const result = await importStudents(file)
      if (result.success_count > 0) {
        message.success(`成功导入 ${result.success_count} 名学生`)
      }
      if (result.fail_count > 0) {
        message.warning(`${result.fail_count} 条记录导入失败`)
      }
      setImportModalVisible(false)
      void fetchData()
    } catch {
      // 错误由请求层统一处理
    }
  }

  const columns: ColumnsType<User> = [
    { title: '学号', dataIndex: 'student_no', key: 'student_no', width: 140, render: (value?: string) => value || '-' },
    { title: '姓名', dataIndex: 'real_name', key: 'real_name', width: 120 },
    { title: '手机号', dataIndex: 'phone', key: 'phone', width: 150, render: (value?: string) => value || '-' },
    { title: '班级', dataIndex: 'class_name', key: 'class_name', width: 160, render: (value?: string) => value || '-' },
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
            title={`确定要${record.status === 'active' ? '禁用' : '启用'}该学生吗？`}
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
    { key: 'keyword', label: '搜索', type: 'input' as const, placeholder: '姓名/学号/手机号', width: 240 },
    { key: 'class_id', label: '班级', type: 'select' as const, options: [{ label: '全部', value: '' }, ...classOptions] },
    { key: 'status', label: '状态', type: 'select' as const, options: [...FILTER_OPTIONS.USER_STATUS] },
  ]

  return (
    <div>
      <PageHeader
        title="学生管理"
        subtitle="管理本校学生账号"
        extra={(
          <Space>
            <Button icon={<UploadOutlined />} onClick={() => setImportModalVisible(true)}>
              批量导入
            </Button>
            <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
              添加学生
            </Button>
          </Space>
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
        title={editingStudent ? '编辑学生' : '添加学生'}
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        footer={null}
      >
        <Form form={form} layout="vertical" onFinish={handleSubmit} className="mt-4">
          <Form.Item name="student_no" label="学号" rules={[{ required: true, message: '请输入学号' }]}>
            <Input placeholder="请输入学号" />
          </Form.Item>
          <Form.Item name="real_name" label="姓名" rules={[{ required: true, message: '请输入姓名' }]}>
            <Input placeholder="请输入学生姓名" />
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
          <Form.Item name="class_id" label="班级">
            <Select placeholder="请选择班级" allowClear options={classOptions} />
          </Form.Item>
          <Form.Item name="email" label="邮箱" rules={[{ type: 'email', message: '邮箱格式不正确' }]}>
            <Input placeholder="请输入邮箱" />
          </Form.Item>
          {!editingStudent ? (
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
                {editingStudent ? '保存' : '添加'}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title="批量导入学生"
        open={importModalVisible}
        onCancel={() => setImportModalVisible(false)}
        footer={null}
      >
        <div className="py-4">
          <p className="mb-4 text-text-secondary">
            请下载模板，按“手机号、学号、姓名、邮箱、密码”的顺序填写后再上传。
          </p>
          <Space direction="vertical" className="w-full">
            <Button icon={<DownloadOutlined />} onClick={() => void handleDownloadTemplate()}>
              下载导入模板
            </Button>
            <Upload
              accept=".xlsx,.xls,.csv"
              showUploadList={false}
              beforeUpload={(file) => {
                void handleImport(file)
                return false
              }}
            >
              <Button type="primary" icon={<UploadOutlined />}>
                上传文件
              </Button>
            </Upload>
          </Space>
        </div>
      </Modal>
    </div>
  )
}
