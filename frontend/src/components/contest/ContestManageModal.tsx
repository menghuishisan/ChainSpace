import { useEffect } from 'react'
import { Button, Form, Input, Modal, Select, Space, Switch } from 'antd'
import type { FormInstance } from 'antd'
import ContestBattleConfigFields from './ContestBattleConfigFields'
import { ensureBattleOrchestration, resolveBattleOrchestrationByContestType } from '@/domains/contest/editor'
import type { Contest, ContestType } from '@/types'
import type { ContestFormValues } from '@/types/presentation'

interface ContestManageModalProps {
  open: boolean
  loading: boolean
  editingContest: Contest | null
  form: FormInstance<ContestFormValues>
  onCancel: () => void
  onSubmit: (values: ContestFormValues) => Promise<void> | void
}

export default function ContestManageModal({
  open,
  loading,
  editingContest,
  form,
  onCancel,
  onSubmit,
}: ContestManageModalProps) {
  const contestType = Form.useWatch('type', form)

  useEffect(() => {
    if (contestType === 'agent_battle' && !form.getFieldValue('battle_orchestration')) {
      form.setFieldValue('battle_orchestration', ensureBattleOrchestration())
    }
  }, [contestType, form])

  return (
    <Modal
      title={editingContest ? '编辑比赛' : '创建比赛'}
      open={open}
      onCancel={onCancel}
      footer={null}
      width={760}
    >
      <Form form={form} layout="vertical" onFinish={onSubmit} className="mt-4">
        <Form.Item name="name" label="比赛名称" rules={[{ required: true, message: '请输入比赛名称' }]}>
          <Input placeholder="请输入比赛名称" />
        </Form.Item>

        <Form.Item name="description" label="比赛描述">
          <Input.TextArea placeholder="请输入比赛描述" rows={3} />
        </Form.Item>

        <div className="grid grid-cols-2 gap-4">
          <Form.Item name="cover" label="封面地址">
            <Input placeholder="可选，用于比赛列表封面展示" />
          </Form.Item>
          <Form.Item name="is_public" label="公开展示" valuePropName="checked">
            <Switch />
          </Form.Item>
        </div>

        <Form.Item name="rules" label="比赛规则">
          <Input.TextArea rows={3} placeholder="填写比赛规则、计分说明或注意事项" />
        </Form.Item>

        <Form.Item name="type" label="比赛类型" rules={[{ required: true, message: '请选择比赛类型' }]}>
          <Select
            placeholder="请选择比赛类型"
            options={[
              { label: '解题赛 (Jeopardy)', value: 'jeopardy' },
              { label: '智能体对抗赛', value: 'agent_battle' },
            ]}
            onChange={(value: ContestType) => {
              form.setFieldValue(
                'battle_orchestration',
                resolveBattleOrchestrationByContestType(value, form.getFieldValue('battle_orchestration')),
              )
            }}
          />
        </Form.Item>

        <ContestBattleConfigFields />

        <Form.Item className="mb-0 text-right">
          <Space>
            <Button onClick={onCancel}>取消</Button>
            <Button type="primary" htmlType="submit" loading={loading}>
              {editingContest ? '保存' : '创建'}
            </Button>
          </Space>
        </Form.Item>
      </Form>
    </Modal>
  )
}
