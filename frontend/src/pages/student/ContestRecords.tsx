/**
 * 学生比赛记录页。
 * 展示学生参与过的解题赛与对抗赛记录，并统一跳转到比赛详情页。
 */
import { useCallback, useEffect, useMemo, useState } from 'react'
import { Button, Card, Col, Empty, Row, Statistic, Table, Tag } from 'antd'
import { CheckCircleOutlined, ClockCircleOutlined, EyeOutlined, TrophyOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'

import { PageHeader } from '@/components/common'
import { getMyContestRecords } from '@/api/contest'
import type { ContestRecord, PaginatedData } from '@/types'
import { ContestTypeMap } from '@/types'

export default function StudentContestRecords() {
  const navigate = useNavigate()
  const [loading, setLoading] = useState(false)
  const [data, setData] = useState<PaginatedData<ContestRecord>>({ list: [], total: 0, page: 1, page_size: 20 })

  const stats = useMemo(() => {
    const rankedRecords = data.list.filter((record) => record.rank)
    return {
      total: data.total,
      completed: data.list.filter((record) => record.status === 'ended').length,
      totalScore: data.list.reduce((sum, record) => sum + record.total_score, 0),
      avgRank: Math.round(rankedRecords.reduce((sum, record) => sum + (record.rank || 0), 0) / (rankedRecords.length || 1)),
    }
  }, [data.list, data.total])

  const fetchData = useCallback(async (page = 1) => {
    setLoading(true)
    try {
      const result = await getMyContestRecords({ page, page_size: data.page_size })
      setData(result)
    } finally {
      setLoading(false)
    }
  }, [data.page_size])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  const handleViewContest = (record: ContestRecord) => {
    navigate(`/contest/${record.contest_id}`)
  }

  const columns = [
    {
      title: '比赛名称',
      dataIndex: 'contest_name',
      key: 'contest_name',
    },
    {
      title: '类型',
      dataIndex: 'contest_type',
      key: 'contest_type',
      width: 120,
      render: (value: string) => ContestTypeMap[value as keyof typeof ContestTypeMap] || value,
    },
    {
      title: '队伍',
      dataIndex: 'team_name',
      key: 'team_name',
      width: 140,
      render: (value?: string) => value || '个人参赛',
    },
    {
      title: '排名',
      dataIndex: 'rank',
      key: 'rank',
      width: 90,
      render: (value?: number) => {
        if (!value) {
          return '-'
        }
        const isTop3 = value <= 3
        return (
          <span className={isTop3 ? 'font-bold' : ''}>
            {isTop3 && <TrophyOutlined className={`mr-1 ${value === 1 ? 'text-yellow-500' : value === 2 ? 'text-gray-400' : 'text-amber-600'}`} />}
            {value}
          </span>
        )
      },
    },
    {
      title: '得分',
      dataIndex: 'total_score',
      key: 'total_score',
      width: 100,
      render: (value: number) => <span className="font-medium text-primary">{value}</span>,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 110,
      render: (value: string) => (
        <Tag
          color={value === 'ended' ? 'default' : value === 'ongoing' ? 'success' : 'processing'}
          icon={value === 'ended' ? <CheckCircleOutlined /> : <ClockCircleOutlined />}
        >
          {value === 'ended' ? '已结束' : value === 'ongoing' ? '进行中' : '未开始'}
        </Tag>
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 90,
      render: (_: unknown, record: ContestRecord) => (
        <Button type="link" size="small" icon={<EyeOutlined />} onClick={() => handleViewContest(record)}>
          查看
        </Button>
      ),
    },
  ]

  return (
    <div>
      <PageHeader title="比赛记录" subtitle="查看你参与过的解题赛与对抗赛历史" />

      <Row gutter={16} className="mb-6">
        <Col xs={12} sm={6}>
          <Card>
            <Statistic title="参与比赛" value={stats.total} suffix="场" valueStyle={{ color: '#1890FF' }} />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card>
            <Statistic title="已完成" value={stats.completed} suffix="场" valueStyle={{ color: '#52C41A' }} />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card>
            <Statistic title="累计得分" value={stats.totalScore} valueStyle={{ color: '#722ED1' }} />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card>
            <Statistic title="平均排名" value={stats.avgRank || '-'} valueStyle={{ color: '#FA8C16' }} />
          </Card>
        </Col>
      </Row>

      <Card>
        {data.list.length === 0 && !loading ? (
          <Empty description="暂无比赛记录，去参加一场比赛吧" />
        ) : (
          <Table
            columns={columns}
            dataSource={data.list}
            rowKey="contest_id"
            loading={loading}
            pagination={{
              current: data.page,
              pageSize: data.page_size,
              total: data.total,
              showSizeChanger: true,
              showTotal: (total) => `共 ${total} 条`,
              onChange: fetchData,
            }}
          />
        )}
      </Card>
    </div>
  )
}
