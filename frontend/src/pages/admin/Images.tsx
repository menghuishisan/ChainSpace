/**
 * 平台管理员 - 镜像管理页面
 */
import { useState, useEffect, useCallback } from 'react'
import { Table, Button, Modal, Form, Input, Select, InputNumber, Switch, message, Space, Tag } from 'antd'
import { PlusOutlined, EditOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { PageHeader, SearchFilter } from '@/components/common'
import type { DockerImage, PaginatedData } from '@/types'
import { getImages, createImage, updateImage } from '@/api/admin'
import { IMAGE_CATEGORIES } from '@/utils/constants'
import { formatDateTime } from '@/utils/format'

export default function Images() {
  // 状态
  const [loading, setLoading] = useState(false)
  const [data, setData] = useState<PaginatedData<DockerImage>>({ list: [], total: 0, page: 1, page_size: 20 })
  const [filters, setFilters] = useState<{ category?: string }>({})
  const [pagination, setPagination] = useState({ page: 1, page_size: 20 })
  
  // 弹窗状态
  const [modalVisible, setModalVisible] = useState(false)
  const [modalLoading, setModalLoading] = useState(false)
  const [editingImage, setEditingImage] = useState<DockerImage | null>(null)
  const [form] = Form.useForm()

  // 获取数据
  const fetchData = useCallback(async () => {
    setLoading(true)
    try {
      const result = await getImages({ ...pagination, ...filters })
      setData(result)
    } catch {
      // 错误由拦截器处理
    } finally {
      setLoading(false)
    }
  }, [pagination, filters])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  // 处理筛选变化
  const handleFilterChange = (values: Record<string, unknown>) => {
    setFilters(values as typeof filters)
  }

  // 处理搜索
  const handleSearch = () => {
    setPagination((prev) => ({ ...prev, page: 1 }))
  }

  // 处理重置
  const handleReset = () => {
    setFilters({})
    setPagination({ page: 1, page_size: 20 })
  }

  // 打开创建弹窗
  const handleCreate = () => {
    setEditingImage(null)
    form.resetFields()
    form.setFieldsValue({
      default_resources: { cpu: 2, memory: '4GB', storage: '10GB' },
    })
    setModalVisible(true)
  }

  // 打开编辑弹窗
  const handleEdit = (record: DockerImage) => {
    setEditingImage(record)
    form.setFieldsValue({
      ...record,
      default_resources: record.default_resources,
    })
    setModalVisible(true)
  }

  // 处理状态切换
  const handleToggleActive = async (record: DockerImage) => {
    try {
      await updateImage(record.id, { is_active: !record.is_active })
      message.success(`${!record.is_active ? '启用' : '禁用'}成功`)
      fetchData()
    } catch {
      // 错误由拦截器处理
    }
  }

  // 处理表单提交
  const handleSubmit = async (values: Record<string, unknown>) => {
    setModalLoading(true)
    try {
      const submitData = {
        name: values.name as string,
        tag: values.tag as string,
        category: values.category as string,
        description: values.description as string,
        default_resources: values.default_resources as NonNullable<DockerImage['default_resources']>,
      }

      if (editingImage) {
        await updateImage(editingImage.id, submitData)
        message.success('更新成功')
      } else {
        await createImage(submitData)
        message.success('创建成功')
      }
      setModalVisible(false)
      fetchData()
    } catch {
      // 错误由拦截器处理
    } finally {
      setModalLoading(false)
    }
  }

  // 表格列定义
  const columns: ColumnsType<DockerImage> = [
    {
      title: '镜像名称',
      dataIndex: 'name',
      key: 'name',
      width: 200,
    },
    {
      title: '标签',
      dataIndex: 'tag',
      key: 'tag',
      width: 100,
      render: (tag) => <Tag>{tag}</Tag>,
    },
    {
      title: '分类',
      dataIndex: 'category',
      key: 'category',
      width: 120,
      render: (category) => {
        const cat = IMAGE_CATEGORIES.find((c) => c.key === category)
        return cat?.label || category
      },
    },
    {
      title: '默认资源',
      key: 'resources',
      width: 200,
      render: (_, record) => {
        const res = record.default_resources
        if (!res) return '-'
        return `${res.cpu || 0}核 / ${res.memory || '-'} / ${res.storage || '-'}`
      },
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
      render: (text) => text || '-',
    },
    {
      title: '状态',
      dataIndex: 'is_active',
      key: 'is_active',
      width: 100,
      render: (isActive, record) => (
        <Switch
          checked={isActive}
          onChange={() => handleToggleActive(record)}
          checkedChildren="启用"
          unCheckedChildren="禁用"
        />
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (text) => formatDateTime(text),
    },
    {
      title: '操作',
      key: 'action',
      width: 80,
      render: (_, record) => (
        <Button type="link" size="small" icon={<EditOutlined />} onClick={() => handleEdit(record)}>
          编辑
        </Button>
      ),
    },
  ]

  // 筛选配置
  const filterConfig = [
    { key: 'keyword', label: '关键词', type: 'input' as const, placeholder: '搜索镜像名称', width: 200 },
    {
      key: 'category',
      label: '分类',
      type: 'select' as const,
      options: [
        { label: '全部', value: '' },
        ...IMAGE_CATEGORIES.map((c) => ({ label: c.label, value: c.key })),
      ],
    },
  ]

  return (
    <div>
      <PageHeader
        title="镜像管理"
        subtitle="管理实验环境所使用的Docker镜像"
        extra={
          <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
            添加镜像
          </Button>
        }
      />

      <SearchFilter
        filters={filterConfig}
        values={filters}
        onChange={handleFilterChange}
        onSearch={handleSearch}
        onReset={handleReset}
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
            showTotal: (total) => `共 ${total} 条`,
            onChange: (page, pageSize) => setPagination({ page, page_size: pageSize }),
          }}
        />
      </div>

      {/* 创建/编辑弹窗 */}
      <Modal
        title={editingImage ? '编辑镜像' : '添加镜像'}
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        footer={null}
        width={600}
      >
        <Form form={form} layout="vertical" onFinish={handleSubmit} className="mt-4">
          <Form.Item
            name="name"
            label="镜像名称"
            rules={[{ required: true, message: '请输入镜像名称' }]}
          >
            <Input placeholder="例如：chainspace/eth-dev" />
          </Form.Item>

          <Form.Item
            name="tag"
            label="镜像标签"
            rules={[{ required: true, message: '请输入镜像标签' }]}
          >
            <Input placeholder="例如：latest 或 1.0.0" />
          </Form.Item>

          <Form.Item
            name="category"
            label="分类"
            rules={[{ required: true, message: '请选择分类' }]}
          >
            <Select
              placeholder="请选择分类"
              options={IMAGE_CATEGORIES.map((c) => ({ label: c.label, value: c.key }))}
            />
          </Form.Item>

          <Form.Item name="description" label="描述">
            <Input.TextArea placeholder="请输入描述" rows={3} />
          </Form.Item>

          <div className="mb-4">
            <h4 className="text-text-primary font-medium mb-2">默认资源配置</h4>
          </div>

          <div className="grid grid-cols-3 gap-4">
            <Form.Item
              name={['default_resources', 'cpu']}
              label="CPU核数"
              rules={[{ required: true, message: '请输入CPU核数' }]}
            >
              <InputNumber min={1} max={16} className="w-full" />
            </Form.Item>

            <Form.Item
              name={['default_resources', 'memory']}
              label="内存"
              rules={[{ required: true, message: '请输入内存' }]}
            >
              <Input placeholder="例如：4GB" />
            </Form.Item>

            <Form.Item
              name={['default_resources', 'storage']}
              label="存储"
              rules={[{ required: true, message: '请输入存储' }]}
            >
              <Input placeholder="例如：10GB" />
            </Form.Item>
          </div>

          <Form.Item className="mb-0 text-right">
            <Space>
              <Button onClick={() => setModalVisible(false)}>取消</Button>
              <Button type="primary" htmlType="submit" loading={modalLoading}>
                {editingImage ? '保存' : '创建'}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
