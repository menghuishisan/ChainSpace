/**
 * 教师 - 创建课程页面
 */
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Card, Form, Input, Select, Upload, Button, message, Space } from 'antd'
import { UploadOutlined, PlusOutlined } from '@ant-design/icons'
import type { UploadProps } from 'antd'
import { PageHeader } from '@/components/common'
import { createCourse } from '@/api/course'
import { uploadImage } from '@/api/common'

export default function TeacherCourseCreate() {
  const navigate = useNavigate()
  const [loading, setLoading] = useState(false)
  const [coverUrl, setCoverUrl] = useState<string>('')
  const [uploading, setUploading] = useState(false)
  const [form] = Form.useForm()

  // 处理封面上传
  const handleUpload: UploadProps['customRequest'] = async (options) => {
    const { file, onSuccess, onError } = options
    setUploading(true)
    try {
      const result = await uploadImage(file as File)
      setCoverUrl(result.url)
      onSuccess?.(result)
      message.success('封面上传成功')
    } catch (error) {
      onError?.(error as Error)
      message.error('封面上传失败')
    } finally {
      setUploading(false)
    }
  }

  // 提交表单
  const handleSubmit = async (values: Record<string, unknown>) => {
    setLoading(true)
    try {
      const course = await createCourse({
        title: values.name as string,
        description: values.description as string,
        cover: coverUrl || undefined,
        is_public: values.visibility === 'public',
      })
      message.success('课程创建成功')
      navigate(`/teacher/courses/${course.id}`)
    } catch {
      // 错误由拦截器处理
    } finally {
      setLoading(false)
    }
  }

  return (
    <div>
      <PageHeader
        title="创建课程"
        subtitle="创建一个新的课程"
        showBack
        backPath="/teacher/courses"
      />

      <Card>
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          initialValues={{ visibility: 'school' }}
          className="max-w-2xl"
        >
          <Form.Item
            name="name"
            label="课程名称"
            rules={[{ required: true, message: '请输入课程名称' }]}
          >
            <Input placeholder="请输入课程名称" maxLength={100} showCount />
          </Form.Item>

          <Form.Item name="description" label="课程描述">
            <Input.TextArea
              placeholder="请输入课程描述，帮助学生了解课程内容"
              rows={4}
              maxLength={500}
              showCount
            />
          </Form.Item>

          <Form.Item label="课程封面">
            <Upload
              accept="image/*"
              showUploadList={false}
              customRequest={handleUpload}
              listType="picture-card"
            >
              {coverUrl ? (
                <img src={coverUrl} alt="封面" className="w-full h-full object-cover" />
              ) : (
                <div>
                  {uploading ? <UploadOutlined spin /> : <PlusOutlined />}
                  <div className="mt-2">上传封面</div>
                </div>
              )}
            </Upload>
            <div className="text-text-secondary text-sm mt-1">
              建议尺寸：800x450，支持 JPG、PNG 格式
            </div>
          </Form.Item>

          <Form.Item
            name="visibility"
            label="可见性"
            rules={[{ required: true }]}
          >
            <Select>
              <Select.Option value="school">仅本校可见</Select.Option>
              <Select.Option value="public">全平台可见</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item className="mb-0">
            <Space>
              <Button onClick={() => navigate('/teacher/courses')}>取消</Button>
              <Button type="primary" htmlType="submit" loading={loading}>
                创建课程
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </div>
  )
}
