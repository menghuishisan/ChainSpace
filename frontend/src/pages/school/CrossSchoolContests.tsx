/**
 * 学校管理员 - 跨校比赛申请页面
 */
import { useState, useEffect, useCallback } from 'react'
import { Card, Table, Button, Tag, Space, Modal, Form, Input, DatePicker, InputNumber, message, Tabs } from 'antd'
import { PlusOutlined, CheckOutlined, CloseOutlined } from '@ant-design/icons'
import { PageHeader } from '@/components/common'
import { createContest } from '@/api/contest'
import { applyCrossSchoolContest, getCrossSchoolApplications, handleCrossSchoolInvitation } from '@/api/school'
import { formatDateTime } from '@/utils/format'
import { usePersistedTab } from '@/hooks'
import type { CrossSchoolApplication } from '@/types'

export default function CrossSchoolContests() {
  const [activeTab, setActiveTab] = usePersistedTab(
    'cross_school_contests',
    'send',
    ['send', 'received']
  )
  const [loading, setLoading] = useState(false)
  
  // 发起的申请
  const [sentApplications, setSentApplications] = useState<CrossSchoolApplication[]>([])
  // 收到的邀请
  const [receivedInvitations, setReceivedInvitations] = useState<CrossSchoolApplication[]>([])
  
  const [applyModalVisible, setApplyModalVisible] = useState(false)
  const [form] = Form.useForm()

  // 获取数据
  const fetchData = useCallback(async () => {
    setLoading(true)
    try {
      const result = await getCrossSchoolApplications()
      const applications = Array.isArray(result) ? result : []
      const sent = applications.filter((a) => a.from_school)
      const received = applications.filter((a) => a.to_school)
      setSentApplications(sent)
      setReceivedInvitations(received)
    } catch { /* */ } finally { setLoading(false) }
  }, [])

  useEffect(() => { fetchData() }, [fetchData])

  // 发起跨校比赛申请
  const handleApply = async () => {
    const values = await form.validateFields()
    try {
      // 先创建比赛
      const contest = await createContest({
        ...values,
        type: 'jeopardy',
        level: 'cross_school',
        start_time: values.time_range[0].toISOString(),
        end_time: values.time_range[1].toISOString(),
        registration_deadline: values.registration_deadline?.toISOString(),
      })
      // 申请跨校
      await applyCrossSchoolContest({ contest_id: contest.id, target_school_ids: values.target_school_ids })
      message.success('申请已提交')
      setApplyModalVisible(false)
      form.resetFields()
      fetchData()
    } catch { /* */ }
  }

  // 处理收到的邀请
  const handleInvitation = async (id: number, action: 'approve' | 'reject') => {
    try {
      await handleCrossSchoolInvitation(id, action)
      message.success(action === 'approve' ? '已同意' : '已拒绝')
      fetchData()
    } catch { /* */ }
  }

  // 发起的申请列表
  const sentColumns = [
    { title: '申请类型', dataIndex: 'type', key: 'type', render: (v: string) => v === 'contest' ? '竞赛' : '题目' },
    { title: '目标学校', key: 'to_school', render: (_: unknown, r: CrossSchoolApplication) => r.to_school?.name || '-' },
    { title: '状态', dataIndex: 'status', key: 'status', render: (v: string) => (
      <Tag color={v === 'approved' ? 'success' : v === 'rejected' ? 'error' : 'processing'}>
        {v === 'approved' ? '已通过' : v === 'rejected' ? '已拒绝' : '审核中'}
      </Tag>
    )},
    { title: '申请时间', dataIndex: 'created_at', key: 'created_at', render: (v: string) => formatDateTime(v) },
  ]

  // 收到的邀请列表
  const receivedColumns = [
    { title: '申请类型', dataIndex: 'type', key: 'type', render: (v: string) => v === 'contest' ? '竞赛' : '题目' },
    { title: '发起学校', key: 'from_school', render: (_: unknown, r: CrossSchoolApplication) => r.from_school?.name || '-' },
    { title: '状态', dataIndex: 'status', key: 'status', render: (v: string) => (
      <Tag color={v === 'approved' ? 'success' : v === 'rejected' ? 'error' : 'processing'}>
        {v === 'approved' ? '已同意' : v === 'rejected' ? '已拒绝' : '待处理'}
      </Tag>
    )},
    { title: '邀请时间', dataIndex: 'created_at', key: 'created_at', render: (v: string) => formatDateTime(v) },
    { title: '操作', key: 'action', render: (_: unknown, record: CrossSchoolApplication) => (
      record.status === 'pending' ? (
        <Space>
          <Button type="link" size="small" icon={<CheckOutlined />} onClick={() => handleInvitation(record.id, 'approve')}>同意</Button>
          <Button type="link" size="small" danger icon={<CloseOutlined />} onClick={() => handleInvitation(record.id, 'reject')}>拒绝</Button>
        </Space>
      ) : null
    )}
  ]

  return (
    <div>
      <PageHeader 
        title="跨校比赛申请" 
        subtitle="发起或处理跨校联合比赛" 
        extra={<Button type="primary" icon={<PlusOutlined />} onClick={() => setApplyModalVisible(true)}>发起跨校比赛</Button>} 
      />

      <Card>
        <Tabs activeKey={activeTab} onChange={setActiveTab} items={[
          {
            key: 'send',
            label: `我发起的申请 (${sentApplications.length})`,
            children: (
              <Table columns={sentColumns} dataSource={sentApplications} rowKey="id" loading={loading} pagination={false} />
            )
          },
          {
            key: 'received',
            label: `收到的邀请 (${receivedInvitations.filter(i => i.status === 'pending').length})`,
            children: (
              <Table columns={receivedColumns} dataSource={receivedInvitations} rowKey="id" loading={loading} pagination={false} />
            )
          },
        ]} />
      </Card>

      {/* 发起申请弹窗 */}
      <Modal title="发起跨校比赛" open={applyModalVisible} onCancel={() => setApplyModalVisible(false)} onOk={handleApply} width={600}>
        <Form form={form} layout="vertical">
          <Form.Item label="比赛名称" name="name" rules={[{ required: true, message: '请输入比赛名称' }]}>
            <Input placeholder="如：XX大学&YY大学联合CTF赛" />
          </Form.Item>
          <Form.Item label="比赛描述" name="description">
            <Input.TextArea rows={3} placeholder="请输入比赛描述" />
          </Form.Item>
          <Form.Item label="比赛时间" name="time_range" rules={[{ required: true, message: '请选择比赛时间' }]}>
            <DatePicker.RangePicker showTime className="w-full" />
          </Form.Item>
          <Form.Item label="报名截止时间" name="registration_deadline">
            <DatePicker showTime className="w-full" />
          </Form.Item>
          <div className="grid grid-cols-2 gap-4">
            <Form.Item label="最大参赛人数" name="max_participants">
              <InputNumber min={10} className="w-full" placeholder="不限制请留空" />
            </Form.Item>
            <Form.Item label="队伍人数" name="team_size">
              <InputNumber min={1} max={5} className="w-full" placeholder="个人赛留空" />
            </Form.Item>
          </div>
          <Form.Item label="邀请学校" name="target_school_ids" rules={[{ required: true, message: '请选择邀请的学校' }]} extra="请联系平台管理员获取其他学校ID">
            <Input placeholder="输入学校ID，多个用逗号分隔" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
