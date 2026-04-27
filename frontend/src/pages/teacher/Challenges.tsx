import { useCallback, useEffect, useState } from 'react'
import { Button, Form, message } from 'antd'
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

export default function TeacherChallenges() {
  const { user } = useUserStore()
  const [loading, setLoading] = useState(false)
  const [data, setData] = useState<PaginatedData<Challenge>>({ list: [], total: 0, page: 1, page_size: 20 })
  const [filters, setFilters] = useState<ChallengeListFilters>(DEFAULT_CHALLENGE_FILTERS)
  const [pagination, setPagination] = useState({ page: 1, page_size: 20 })
  const [modalVisible, setModalVisible] = useState(false)
  const [modalLoading, setModalLoading] = useState(false)
  const [editingChallenge, setEditingChallenge] = useState<Challenge | null>(null)
  const [form] = Form.useForm<ChallengeManageFormValues>()

  const fetchData = useCallback(async () => {
    setLoading(true)
    try {
      const result = await getChallenges(buildChallengeListQueryParams(
        pagination.page,
        pagination.page_size,
        filters,
      ))
      setData(result)
    } finally {
      setLoading(false)
    }
  }, [filters, pagination])

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

  const handleDelete = async (record: Challenge) => {
    try {
      await deleteChallenge(record.id)
      message.success('删除成功')
      await fetchData()
    } catch {
      // handled by interceptor
    }
  }

  const handleRequestPublish = async (record: Challenge) => {
    try {
      await requestPublishChallenge(record.id)
      message.success('已提交公开申请')
    } catch {
      // handled by interceptor
    }
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
      await fetchData()
    } finally {
      setModalLoading(false)
    }
  }

  return (
    <div>
      <PageHeader
        title="题目管理"
        subtitle="按知识主题和环境类型管理解题赛题目"
        extra={<Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>创建题目</Button>}
      />

      <ChallengeSearchFilter
        values={filters}
        showDifficulty={false}
        onChange={setFilters}
        onSearch={() => setPagination((current) => ({ ...current, page: 1 }))}
        onReset={() => {
          setFilters(DEFAULT_CHALLENGE_FILTERS)
          setPagination({ page: 1, page_size: 20 })
        }}
      />

      <div className="card">
        <ChallengeTable
          data={data}
          loading={loading}
          onEdit={handleEdit}
          onDelete={(record) => void handleDelete(record)}
          onRequestPublish={(record) => void handleRequestPublish(record)}
          canManage={(record) => Boolean(user && record.school_id === user.school_id && record.creator_id === user.id)}
          onPageChange={(page, pageSize) => setPagination({ page, page_size: pageSize })}
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
