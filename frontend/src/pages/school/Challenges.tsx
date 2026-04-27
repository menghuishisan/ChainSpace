import { useCallback, useEffect, useState } from 'react'
import { Button, Form, Modal, message } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { createChallenge, deleteChallenge, getChallenges, requestPublishChallenge, updateChallenge } from '@/api/challenge'
import { ChallengeManageModal, ChallengeSearchFilter, ChallengeTable } from '@/components/challenge'
import { PageHeader } from '@/components/common'
import {
  buildChallengeCreateInitialValues,
  buildChallengeEditInitialValues,
  buildChallengeListQueryParams,
  buildChallengeSubmitData,
  buildChallengeUpdateData,
  DEFAULT_CHALLENGE_FILTERS,
} from '@/domains/challenge/management'
import { useUserStore } from '@/store'
import type { Challenge, PaginatedData } from '@/types'
import type { ChallengeListFilters, ChallengeManageFormValues } from '@/types/presentation'

export default function SchoolChallenges() {
  const { user } = useUserStore()
  const [loading, setLoading] = useState(false)
  const [data, setData] = useState<PaginatedData<Challenge>>({ list: [], total: 0, page: 1, page_size: 20 })
  const [filters, setFilters] = useState<ChallengeListFilters>(DEFAULT_CHALLENGE_FILTERS)
  const [modalVisible, setModalVisible] = useState(false)
  const [modalLoading, setModalLoading] = useState(false)
  const [editingChallenge, setEditingChallenge] = useState<Challenge | null>(null)
  const [form] = Form.useForm<ChallengeManageFormValues>()

  const fetchData = useCallback(async (page = 1, pageSize = data.page_size) => {
    setLoading(true)
    try {
      const result = await getChallenges(buildChallengeListQueryParams(page, pageSize, filters, { is_public: false }))
      setData(result)
    } finally {
      setLoading(false)
    }
  }, [data.page_size, filters])

  useEffect(() => {
    void fetchData()
  }, [fetchData])

  const handleCreate = () => {
    setEditingChallenge(null)
    form.resetFields()
    form.setFieldsValue(buildChallengeCreateInitialValues())
    setModalVisible(true)
  }

  const handleEdit = (record: Challenge) => {
    setEditingChallenge(record)
    form.resetFields()
    form.setFieldsValue(buildChallengeEditInitialValues(record))
    setModalVisible(true)
  }

  const handleDelete = (record: Challenge) => {
    Modal.confirm({
      title: '确认删除',
      content: '删除后无法恢复，确定要删除吗？',
      onOk: async () => {
        try {
          await deleteChallenge(record.id)
          message.success('删除成功')
          await fetchData(data.page, data.page_size)
        } catch {
          // handled by interceptor
        }
      },
    })
  }

  const handleRequestPublish = (record: Challenge) => {
    Modal.confirm({
      title: '申请公开',
      content: `确定要将题目「${record.title}」提交到平台公共题库吗？`,
      onOk: async () => {
        try {
          await requestPublishChallenge(record.id)
          message.success('申请已提交')
          await fetchData(data.page, data.page_size)
        } catch {
          // handled by interceptor
        }
      },
    })
  }

  const handleSubmit = async (values: ChallengeManageFormValues) => {
    setModalLoading(true)
    try {
      if (editingChallenge) {
        await updateChallenge(editingChallenge.id, buildChallengeUpdateData(values))
        message.success('更新成功')
      } else {
        await createChallenge(buildChallengeSubmitData(values))
        message.success('创建成功')
      }

      setModalVisible(false)
      await fetchData(data.page, data.page_size)
    } finally {
      setModalLoading(false)
    }
  }

  return (
    <div>
      <PageHeader
        title="本校题库"
        subtitle="按知识主题和环境类型维护本校解题赛题目"
        extra={<Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>创建题目</Button>}
      />

      <div className="card">
        <ChallengeSearchFilter
          values={filters}
          onChange={setFilters}
          onSearch={() => void fetchData(1, data.page_size)}
          onReset={() => setFilters(DEFAULT_CHALLENGE_FILTERS)}
        />

        <ChallengeTable
          data={data}
          loading={loading}
          onView={handleEdit}
          onEdit={handleEdit}
          onDelete={handleDelete}
          onRequestPublish={handleRequestPublish}
          canManage={(record) => Boolean(user && record.school_id === user.school_id)}
          onPageChange={(page, pageSize) => void fetchData(page, pageSize)}
        />
      </div>

      <ChallengeManageModal
        open={modalVisible}
        loading={modalLoading}
        editingChallenge={editingChallenge}
        form={form}
        allowDirectPublic={false}
        onCancel={() => setModalVisible(false)}
        onSubmit={handleSubmit}
      />
    </div>
  )
}
