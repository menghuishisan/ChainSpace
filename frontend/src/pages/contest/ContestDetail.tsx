/**
 * 比赛详情分流页。
 * 该页面只负责读取比赛基础信息，并根据比赛类型切换到对应的详情视图。
 */
import { useEffect, useState } from 'react'
import { Spin, message } from 'antd'
import { useNavigate, useParams } from 'react-router-dom'

import { getContest } from '@/api/contest'
import type { Contest } from '@/types'

import ContestAgentBattleDetail from './ContestAgentBattleDetail'
import ContestJeopardyDetail from './ContestJeopardyDetail'

export default function ContestDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const contestId = Number(id || '0')

  const [loading, setLoading] = useState(true)
  const [contest, setContest] = useState<Contest | null>(null)

  useEffect(() => {
    const loadContest = async () => {
      setLoading(true)
      try {
        const detail = await getContest(contestId)
        setContest(detail)
      } catch {
        message.error('获取比赛详情失败')
        navigate(-1)
      } finally {
        setLoading(false)
      }
    }

    loadContest()
  }, [contestId, navigate])

  if (loading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <Spin size="large" />
      </div>
    )
  }

  if (!contest) {
    return null
  }

  if (contest.type === 'agent_battle') {
    return <ContestAgentBattleDetail contest={contest} />
  }

  return <ContestJeopardyDetail contest={contest} />
}
