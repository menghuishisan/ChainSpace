/**
 * 学校管理员 - 学校设置页面
 */
import { useState, useEffect } from 'react'
import { Card, Form, Input, Upload, Button, message, Switch, Spin, Divider } from 'antd'
import { UploadOutlined, SaveOutlined } from '@ant-design/icons'
import type { UploadProps } from 'antd'
import { PageHeader } from '@/components/common'
import { getSchoolInfo, updateSchoolInfo } from '@/api/school'
import { uploadImage } from '@/api/common'
import type { SchoolSettingsFormValues } from '@/types/presentation'

export default function SchoolSettings() {
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  // 获取学校信息
  useEffect(() => {
    const fetchData = async () => {
      setLoading(true)
      try {
        const data = await getSchoolInfo()
        form.setFieldsValue(data)
      } catch { /* 错误由拦截器处理 */ } finally { setLoading(false) }
    }
    fetchData()
  }, [form])

  // 保存设置
  const handleSave = async (values: SchoolSettingsFormValues) => {
    setSaving(true)
    try {
      await updateSchoolInfo(values)
      message.success('保存成功')
    } catch { /* 错误由拦截器处理 */ } finally { setSaving(false) }
  }

  // Logo上传
  const uploadProps: UploadProps = {
    showUploadList: false,
    accept: 'image/*',
    customRequest: async (options) => {
      const { file, onSuccess, onError } = options
      try {
        const result = await uploadImage(file as File)
        form.setFieldValue('logo_url', result.url)
        message.success('上传成功')
        onSuccess?.(result)
      } catch (error) {
        message.error('上传失败')
        onError?.(error as Error)
      }
    },
  }

  if (loading) return <div className="flex items-center justify-center h-64"><Spin size="large" /></div>

  return (
    <div>
      <PageHeader title="学校设置" subtitle="配置本校基本信息和功能开关" />

      <Card>
        <Form form={form} layout="vertical" onFinish={handleSave} initialValues={{ allow_cross_school_contest: false, allow_public_courses: false }}>
          {/* 基本信息 */}
          <h3 className="text-lg font-medium mb-4">基本信息</h3>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Form.Item label="学校名称" name="name" rules={[{ required: true, message: '请输入学校名称' }]}>
              <Input placeholder="请输入学校名称" />
            </Form.Item>
            <Form.Item label="学校代码" name="code" rules={[{ required: true, message: '请输入学校代码' }]}>
              <Input placeholder="如：PKU、THU" disabled />
            </Form.Item>
          </div>

          <Form.Item label="学校Logo" name="logo_url">
            <div className="flex items-center">
              {form.getFieldValue('logo_url') && (
                <img src={form.getFieldValue('logo_url')} alt="Logo" className="w-16 h-16 object-contain mr-4 border rounded" />
              )}
              <Upload {...uploadProps}>
                <Button icon={<UploadOutlined />}>上传Logo</Button>
              </Upload>
            </div>
          </Form.Item>

          <Form.Item label="学校简介" name="description">
            <Input.TextArea rows={3} placeholder="请输入学校简介" />
          </Form.Item>

          <Divider />

          {/* 联系方式 */}
          <h3 className="text-lg font-medium mb-4">联系方式</h3>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Form.Item label="联系邮箱" name="contact_email">
              <Input placeholder="admin@school.edu.cn" />
            </Form.Item>
            <Form.Item label="联系电话" name="contact_phone">
              <Input placeholder="010-12345678" />
            </Form.Item>
          </div>
          <Form.Item label="学校地址" name="address">
            <Input placeholder="请输入学校地址" />
          </Form.Item>

          <Divider />

          {/* 功能设置 */}
          <h3 className="text-lg font-medium mb-4">功能设置</h3>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Form.Item label="允许跨校比赛" name="allow_cross_school_contest" valuePropName="checked">
              <Switch checkedChildren="开启" unCheckedChildren="关闭" />
            </Form.Item>
            <Form.Item label="允许公开课程" name="allow_public_courses" valuePropName="checked">
              <Switch checkedChildren="开启" unCheckedChildren="关闭" />
            </Form.Item>
          </div>

          <Divider />

          {/* 配额设置 */}
          <h3 className="text-lg font-medium mb-4">配额设置</h3>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Form.Item label="最大学生数" name="max_students">
              <Input type="number" placeholder="不限制请留空" />
            </Form.Item>
            <Form.Item label="最大教师数" name="max_teachers">
              <Input type="number" placeholder="不限制请留空" />
            </Form.Item>
          </div>

          <div className="flex justify-end mt-4">
            <Button type="primary" htmlType="submit" icon={<SaveOutlined />} loading={saving}>保存设置</Button>
          </div>
        </Form>
      </Card>
    </div>
  )
}
