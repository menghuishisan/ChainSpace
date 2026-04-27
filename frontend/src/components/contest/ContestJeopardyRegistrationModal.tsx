import { Button, Form, Input, Modal } from 'antd'

interface ContestJeopardyRegistrationModalProps {
  registerOpen: boolean
  teamOpen: boolean
  joinTeamCode: string
  loading?: boolean
  onRegisterCancel: () => void
  onRegisterConfirm: () => void
  onTeamCancel: () => void
  onCreateTeam: (values: { name: string }) => void
  onJoinTeamCodeChange: (value: string) => void
  onJoinTeam: () => void
}

export default function ContestJeopardyRegistrationModal({
  registerOpen,
  teamOpen,
  joinTeamCode,
  loading = false,
  onRegisterCancel,
  onRegisterConfirm,
  onTeamCancel,
  onCreateTeam,
  onJoinTeamCodeChange,
  onJoinTeam,
}: ContestJeopardyRegistrationModalProps) {
  const [form] = Form.useForm<{ name: string }>()

  return (
    <>
      <Modal title="确认报名" open={registerOpen} onCancel={onRegisterCancel} onOk={onRegisterConfirm} okText="确认报名" confirmLoading={loading}>
        <p>确认要报名参加这场比赛吗？</p>
      </Modal>

      <Modal title="队伍报名" open={teamOpen} onCancel={onTeamCancel} footer={null}>
        <div className="py-4">
          <h4 className="mb-4">创建新队伍</h4>
          <Form
            form={form}
            onFinish={(values) => {
              onCreateTeam(values)
              form.resetFields()
            }}
          >
            <Form.Item name="name" rules={[{ required: true, message: '请输入队伍名称' }]}>
              <Input placeholder="请输入队伍名称" />
            </Form.Item>
            <Button type="primary" htmlType="submit" block loading={loading}>创建队伍</Button>
          </Form>

          <div className="my-6 text-center text-text-secondary">或</div>

          <h4 className="mb-4">加入已有队伍</h4>
          <Input
            placeholder="请输入队伍邀请码"
            value={joinTeamCode}
            onChange={(event) => onJoinTeamCodeChange(event.target.value)}
            className="mb-4"
          />
          <Button block onClick={onJoinTeam} loading={loading}>加入队伍</Button>
        </div>
      </Modal>
    </>
  )
}
