/**
 * 个人信息页面
 */
import { useState } from 'react'
import { Avatar, Button, Card, Descriptions, Form, Input, message, Space, Upload } from 'antd'
import { UploadOutlined, UserOutlined } from '@ant-design/icons'
import type { UploadProps } from 'antd'

import { PageHeader } from '@/components/common'
import { uploadImage } from '@/api/common'
import { useUserStore } from '@/store'
import { RoleNameMap } from '@/types'
import { formatDateTime } from '@/utils/format'

export default function Profile() {
  const { user, updateUser, loading } = useUserStore()
  const [form] = Form.useForm()
  const [avatarUrl, setAvatarUrl] = useState(user?.avatar || '')
  const [uploading, setUploading] = useState(false)

  const handleUpload: UploadProps['customRequest'] = async (options) => {
    const { file, onSuccess, onError } = options
    setUploading(true)
    try {
      const result = await uploadImage(file as File)
      setAvatarUrl(result.url)
      onSuccess?.(result)
      message.success('头像上传成功')
    } catch (error) {
      onError?.(error as Error)
      message.error('上传失败')
    } finally {
      setUploading(false)
    }
  }

  const handleSubmit = async (values: { real_name: string; email?: string; phone?: string }) => {
    try {
      await updateUser({ ...values, avatar: avatarUrl })
      message.success('保存成功')
    } catch {
      // 错误由请求层统一处理
    }
  }

  if (!user) {
    return null
  }

  return (
    <div>
      <PageHeader title="个人信息" subtitle="查看和修改您的账户信息" />

      <Card className="max-w-2xl">
        <Descriptions title="账户信息" column={1} className="mb-8">
          <Descriptions.Item label="姓名">{user.real_name}</Descriptions.Item>
          <Descriptions.Item label="角色">{RoleNameMap[user.role]}</Descriptions.Item>
          <Descriptions.Item label="手机号">{user.phone || '-'}</Descriptions.Item>
          {user.student_no ? <Descriptions.Item label="学号">{user.student_no}</Descriptions.Item> : null}
          {user.school_name ? <Descriptions.Item label="所属学校">{user.school_name}</Descriptions.Item> : null}
          {user.class_name ? <Descriptions.Item label="班级">{user.class_name}</Descriptions.Item> : null}
          <Descriptions.Item label="注册时间">{formatDateTime(user.created_at)}</Descriptions.Item>
          {user.last_login_at ? <Descriptions.Item label="最后登录">{formatDateTime(user.last_login_at)}</Descriptions.Item> : null}
        </Descriptions>

        <h3 className="mb-4 text-lg font-medium">编辑资料</h3>
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          initialValues={{ real_name: user.real_name, email: user.email, phone: user.phone }}
        >
          <Form.Item label="头像">
            <Space>
              <Avatar size={64} src={avatarUrl} icon={<UserOutlined />} />
              <Upload accept="image/*" showUploadList={false} customRequest={handleUpload}>
                <Button icon={<UploadOutlined />} loading={uploading}>更换头像</Button>
              </Upload>
            </Space>
          </Form.Item>

          <Form.Item name="real_name" label="姓名" rules={[{ required: true, message: '请输入姓名' }]}>
            <Input placeholder="请输入姓名" />
          </Form.Item>

          <Form.Item name="email" label="邮箱">
            <Input placeholder="请输入邮箱" type="email" />
          </Form.Item>

          <Form.Item name="phone" label="手机号">
            <Input placeholder="请输入手机号" />
          </Form.Item>

          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading}>保存修改</Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  )
}
