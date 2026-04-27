import { useState } from 'react'
import { Button, Divider, Form, Input, List, Modal, Popconfirm, Select, Space, Tag, Upload, message } from 'antd'
import { FileTextOutlined, LinkOutlined, PlayCircleOutlined, PlusOutlined, UploadOutlined } from '@ant-design/icons'
import type { UploadProps } from 'antd'

import { useCourseMaterials } from '@/hooks'
import { MaterialTypeMap } from '@/types'
import type { Material } from '@/types'
import type { TeacherChapterMaterialsProps } from '@/types/presentation'

export default function TeacherChapterMaterials({
  courseId,
  chapterId,
  materials,
}: TeacherChapterMaterialsProps) {
  const [modalVisible, setModalVisible] = useState(false)
  const [editingMaterial, setEditingMaterial] = useState<Material | null>(null)
  const [uploading, setUploading] = useState(false)
  const [uploadedFileUrl, setUploadedFileUrl] = useState('')
  const [form] = Form.useForm()
  const { saveMaterial, removeMaterial, uploadCourseMaterial } = useCourseMaterials(courseId)

  const handleUpload: UploadProps['customRequest'] = async (options) => {
    const { file, onSuccess, onError } = options
    setUploading(true)
    try {
      const result = await uploadCourseMaterial(chapterId, file as File)
      setUploadedFileUrl(result.url)
      form.setFieldsValue({ url: result.url })
      message.success('文件上传成功')
      onSuccess?.(result)
    } catch (error) {
      onError?.(error as Error)
      message.error('文件上传失败')
    } finally {
      setUploading(false)
    }
  }

  const handleEdit = (material?: Material) => {
    setEditingMaterial(material || null)
    setUploadedFileUrl(material?.url || '')
    form.resetFields()
    if (material) {
      form.setFieldsValue({ ...material, url: material.url })
    }
    setModalVisible(true)
  }

  const handleTypeChange = (type: string) => {
    if (type === 'video' || type === 'document' || type === 'ppt') {
      setUploadedFileUrl('')
      form.setFieldsValue({ url: '' })
    }
  }

  const handleSave = async (values: Record<string, unknown>) => {
    try {
      await saveMaterial(chapterId, editingMaterial?.id || null, {
        title: String(values.title || ''),
        type: String(values.type || ''),
        content: values.content ? String(values.content) : undefined,
        url: uploadedFileUrl || (values.url ? String(values.url) : undefined),
        duration: values.duration ? Number(values.duration) : undefined,
      })
      message.success(editingMaterial ? '资料更新成功' : '资料添加成功')
      setModalVisible(false)
    } catch {
      // 交给请求层统一处理
    }
  }

  const handleDelete = async (materialId: number) => {
    try {
      await removeMaterial(chapterId, materialId)
      message.success('资料删除成功')
    } catch {
      // 交给请求层统一处理
    }
  }

  const getMaterialIcon = (type: string) => {
    switch (type) {
      case 'video':
        return <PlayCircleOutlined className="text-blue-500" />
      case 'link':
        return <LinkOutlined className="text-green-500" />
      default:
        return <FileTextOutlined className="text-orange-500" />
    }
  }

  return (
    <div>
      <div className="mb-3">
        <Button size="small" icon={<PlusOutlined />} onClick={() => handleEdit()}>
          添加资料
        </Button>
      </div>

      {materials.length === 0 ? (
        <div className="py-4 text-center text-text-secondary">暂无资料</div>
      ) : (
        <List
          size="small"
          dataSource={materials}
          renderItem={(item) => (
            <List.Item
              actions={[
                <Button key="edit" type="link" size="small" onClick={() => handleEdit(item)}>编辑</Button>,
                (
                  <Popconfirm key="delete" title="确定删除这条资料吗？" onConfirm={() => void handleDelete(item.id)}>
                    <Button type="link" size="small" danger>删除</Button>
                  </Popconfirm>
                ),
              ]}
            >
              <List.Item.Meta
                avatar={getMaterialIcon(item.type)}
                title={item.title}
                description={<Tag>{MaterialTypeMap[item.type as keyof typeof MaterialTypeMap] || item.type}</Tag>}
              />
            </List.Item>
          )}
        />
      )}

      <Modal
        title={editingMaterial ? '编辑资料' : '添加资料'}
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        footer={null}
        width={600}
      >
        <Form form={form} layout="vertical" onFinish={(values) => void handleSave(values)} className="mt-4">
          <Form.Item name="title" label="资料标题" rules={[{ required: true, message: '请输入标题' }]}>
            <Input placeholder="请输入资料标题" />
          </Form.Item>
          <Form.Item name="type" label="资料类型" rules={[{ required: true, message: '请选择类型' }]}>
            <Select placeholder="请选择资料类型" onChange={handleTypeChange}>
              <Select.Option value="video">视频</Select.Option>
              <Select.Option value="document">文档</Select.Option>
              <Select.Option value="richtext">富文本</Select.Option>
              <Select.Option value="ppt">PPT</Select.Option>
              <Select.Option value="link">外链</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item noStyle shouldUpdate={(prev, curr) => prev.type !== curr.type}>
            {({ getFieldValue }) => {
              const type = getFieldValue('type')
              if (type === 'video' || type === 'document' || type === 'ppt') {
                return (
                  <Form.Item label="上传文件">
                    <Upload
                      accept={type === 'video' ? 'video/*' : type === 'ppt' ? '.ppt,.pptx' : '.pdf,.doc,.docx,.txt'}
                      customRequest={handleUpload}
                      showUploadList={false}
                      maxCount={1}
                    >
                      <Button icon={<UploadOutlined />} loading={uploading}>
                        {uploadedFileUrl ? '重新上传' : '点击上传'}
                      </Button>
                    </Upload>
                    {uploadedFileUrl && (
                      <div className="mt-2 text-sm text-green-600">
                        已上传: {uploadedFileUrl.split('/').pop()}
                      </div>
                    )}
                    <div className="mt-1 text-xs text-text-secondary">
                      {type === 'video' && '支持 mp4、avi、mov 等视频格式'}
                      {type === 'document' && '支持 pdf、doc、docx、txt 等文档格式'}
                      {type === 'ppt' && '支持 ppt、pptx 格式'}
                    </div>
                  </Form.Item>
                )
              }
              return null
            }}
          </Form.Item>

          <Form.Item noStyle shouldUpdate={(prev, curr) => prev.type !== curr.type}>
            {({ getFieldValue }) => {
              if (getFieldValue('type') === 'link') {
                return (
                  <Form.Item name="url" label="文件链接" rules={[{ required: true, message: '请输入链接地址' }]}>
                    <Input placeholder="请输入文件或网页链接" />
                  </Form.Item>
                )
              }
              return null
            }}
          </Form.Item>

          <Form.Item noStyle shouldUpdate={(prev, curr) => prev.type !== curr.type}>
            {({ getFieldValue }) => {
              if (getFieldValue('type') === 'richtext') {
                return (
                  <Form.Item name="content" label="富文本内容" rules={[{ required: true, message: '请输入内容' }]}>
                    <Input.TextArea placeholder="请输入富文本内容或说明" rows={6} />
                  </Form.Item>
                )
              }
              return null
            }}
          </Form.Item>

          <Form.Item noStyle shouldUpdate={(prev, curr) => prev.type !== curr.type}>
            {({ getFieldValue }) => {
              if (getFieldValue('type') === 'video') {
                return (
                  <Form.Item name="duration" label="视频时长（分钟）">
                    <Input type="number" placeholder="请输入视频时长" />
                  </Form.Item>
                )
              }
              return null
            }}
          </Form.Item>

          <Divider />
          <Form.Item className="mb-0 text-right">
            <Space>
              <Button onClick={() => setModalVisible(false)}>取消</Button>
              <Button type="primary" htmlType="submit">保存</Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
