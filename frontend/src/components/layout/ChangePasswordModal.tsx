import { Button, Form, Input, Modal } from 'antd'

type ChangePasswordValues = {
  old_password: string
  new_password: string
  confirm_password: string
}

interface ChangePasswordModalProps {
  loading: boolean
  open: boolean
  onCancel: () => void
  onSubmit: (values: { old_password: string; new_password: string }) => Promise<void>
}

export default function ChangePasswordModal({
  loading,
  open,
  onCancel,
  onSubmit,
}: ChangePasswordModalProps) {
  const [form] = Form.useForm<ChangePasswordValues>()

  const handleCancel = () => {
    form.resetFields()
    onCancel()
  }

  return (
    <Modal
      title="修改密码"
      open={open}
      onCancel={handleCancel}
      footer={null}
      destroyOnHidden
    >
      <Form
        form={form}
        layout="vertical"
        onFinish={async (values) => {
          await onSubmit({
            old_password: values.old_password,
            new_password: values.new_password,
          })
          form.resetFields()
        }}
      >
        <Form.Item name="old_password" label="当前密码" rules={[{ required: true, message: '请输入当前密码' }]}>
          <Input.Password />
        </Form.Item>
        <Form.Item
          name="new_password"
          label="新密码"
          rules={[
            { required: true, message: '请输入新密码' },
            { min: 6, message: '密码长度不能少于 6 位' },
          ]}
        >
          <Input.Password />
        </Form.Item>
        <Form.Item
          name="confirm_password"
          label="确认新密码"
          dependencies={['new_password']}
          rules={[
            { required: true, message: '请确认新密码' },
            ({ getFieldValue }) => ({
              validator(_, value) {
                if (!value || getFieldValue('new_password') === value) {
                  return Promise.resolve()
                }
                return Promise.reject(new Error('两次输入的密码不一致'))
              },
            }),
          ]}
        >
          <Input.Password />
        </Form.Item>
        <div className="text-right">
          <Button onClick={handleCancel} className="mr-2">取消</Button>
          <Button type="primary" htmlType="submit" loading={loading}>确认修改</Button>
        </div>
      </Form>
    </Modal>
  )
}
