import { useCallback, useEffect, useMemo, useState } from 'react'
import { Badge, Button, Card, Empty, Spin, Tag } from 'antd'
import {
  ClockCircleOutlined,
  FireOutlined,
  TeamOutlined,
  TrophyOutlined,
} from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'

import { getContests, registerContest } from '@/api/contest'
import { PageHeader, SearchFilter } from '@/components/common'
import {
  buildContestListQueryParams,
  DEFAULT_CONTEST_LIST_FILTERS,
  getContestSearchFilterItems,
  type ContestListFilters,
} from '@/domains/contest/management'
import { getContestParticipationPresentation } from '@/domains/contest/presentation'
import type { Contest, PaginatedData } from '@/types'
import { ContestStatusMap, ContestTypeMap } from '@/types'
import { formatDateTime, formatRelativeTime } from '@/utils/format'

function renderStatusBadge(contest: Contest) {
  const statusMeta = ContestStatusMap[contest.status]

  if (contest.status === 'published') {
    return <Badge status="processing" text={`${formatRelativeTime(contest.start_time)}开始`} />
  }
  if (contest.status === 'ongoing') {
    return <Badge status="processing" text={statusMeta.text} />
  }

  return <Badge status="default" text={statusMeta.text} />
}

function isTeamContest(contest: Contest) {
  return (contest.team_max_size || 1) > 1
}

export default function StudentContests() {
  const navigate = useNavigate()
  const [loading, setLoading] = useState(false)
  const [registeringId, setRegisteringId] = useState<number | null>(null)
  const [data, setData] = useState<PaginatedData<Contest>>({
    list: [],
    total: 0,
    page: 1,
    page_size: 20,
  })
  const [filters, setFilters] = useState<ContestListFilters>(DEFAULT_CONTEST_LIST_FILTERS)

  const fetchData = useCallback(async () => {
    setLoading(true)
    try {
      const result = await getContests(buildContestListQueryParams(
        { page: 1, page_size: 100 },
        filters,
      ))
      setData(result)
    } finally {
      setLoading(false)
    }
  }, [filters])

  useEffect(() => {
    void fetchData()
  }, [fetchData])

  const overview = useMemo(() => {
    const ongoing = data.list.filter((contest) => contest.status === 'ongoing').length
    const upcoming = data.list.filter((contest) => contest.status === 'published').length
    const teamBased = data.list.filter((contest) => isTeamContest(contest)).length

    return {
      ongoing,
      upcoming,
      teamBased,
      total: data.list.length,
    }
  }, [data.list])

  const handleRegister = async (contestId: number) => {
    setRegisteringId(contestId)
    try {
      await registerContest(contestId)
      await fetchData()
    } finally {
      setRegisteringId(null)
    }
  }

  return (
    <div className="space-y-6">
      <PageHeader title="比赛中心" subtitle="参与真实比赛训练，逐步沉淀区块链安全实战能力" />

      <Card className="overflow-hidden border-0 shadow-sm" styles={{ body: { padding: 0 } }}>
        <div className="grid gap-px overflow-hidden rounded-2xl bg-slate-200 xl:grid-cols-[1.35fr_repeat(4,minmax(0,1fr))]">
          <div className="bg-[linear-gradient(135deg,#081a2f_0%,#0f2744_55%,#13325b_100%)] px-6 py-6 text-white">
            <div className="text-xs uppercase tracking-[0.28em] text-cyan-200">比赛中心</div>
            <div className="mt-3 text-2xl font-semibold">真实区块链安全比赛入口</div>
            <p className="mt-3 max-w-2xl text-sm leading-6 text-slate-200">
              优先展示报名、进入赛场和赛后回顾这些关键动作，让你能更快找到当前应该做什么。
            </p>
          </div>

          <div className="bg-white px-5 py-5">
            <div className="flex items-center gap-2 text-xs uppercase tracking-[0.2em] text-text-secondary">
              <FireOutlined className="text-red-500" />
              进行中
            </div>
            <div className="mt-3 text-3xl font-semibold text-slate-900">{overview.ongoing}</div>
            <div className="mt-1 text-xs text-text-secondary">可直接进入正式赛场</div>
          </div>

          <div className="bg-white px-5 py-5">
            <div className="flex items-center gap-2 text-xs uppercase tracking-[0.2em] text-text-secondary">
              <ClockCircleOutlined className="text-sky-500" />
              报名中
            </div>
            <div className="mt-3 text-3xl font-semibold text-slate-900">{overview.upcoming}</div>
            <div className="mt-1 text-xs text-text-secondary">可提前组队或注册</div>
          </div>

          <div className="bg-white px-5 py-5">
            <div className="flex items-center gap-2 text-xs uppercase tracking-[0.2em] text-text-secondary">
              <TeamOutlined className="text-violet-500" />
              团队赛
            </div>
            <div className="mt-3 text-3xl font-semibold text-slate-900">{overview.teamBased}</div>
            <div className="mt-1 text-xs text-text-secondary">需要队伍协同参赛</div>
          </div>

          <div className="bg-white px-5 py-5">
            <div className="flex items-center gap-2 text-xs uppercase tracking-[0.2em] text-text-secondary">
              <TrophyOutlined className="text-amber-500" />
              全部比赛
            </div>
            <div className="mt-3 text-3xl font-semibold text-slate-900">{overview.total}</div>
            <div className="mt-1 text-xs text-text-secondary">覆盖解题赛与对抗赛</div>
          </div>
        </div>
      </Card>

      <SearchFilter
        filters={getContestSearchFilterItems({ includeDraft: false })}
        values={filters}
        onChange={(next) => setFilters({
          keyword: typeof next.keyword === 'string' ? next.keyword : '',
          type: typeof next.type === 'string' ? next.type : '',
          status: typeof next.status === 'string' ? next.status : '',
        })}
        onSearch={() => void fetchData()}
        onReset={() => setFilters(DEFAULT_CONTEST_LIST_FILTERS)}
      />

      <Spin spinning={loading}>
        {data.list.length === 0 ? (
          <Card>
            <Empty description="暂无比赛" />
          </Card>
        ) : (
          <div className="space-y-4">
            {data.list.map((contest) => {
              const presentation = getContestParticipationPresentation(contest)

              return (
                <Card
                  key={contest.id}
                  hoverable
                  className="overflow-hidden border-0 shadow-sm"
                  styles={{ body: { padding: 0 } }}
                  onClick={() => navigate(`/contest/${contest.id}`)}
                >
                  <div className="grid gap-px bg-slate-200 xl:grid-cols-[minmax(0,1.4fr)_240px]">
                    <div className="bg-white px-6 py-6">
                      <div className="flex flex-wrap items-start justify-between gap-4">
                        <div>
                          <div className="flex items-center gap-3">
                            <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-slate-900 text-white">
                              <TrophyOutlined className="text-lg" />
                            </div>
                            <div>
                              <h3 className="mb-1 text-xl font-semibold text-slate-900">{contest.title}</h3>
                              <div className="flex flex-wrap gap-2">
                                <Tag color="blue">{ContestTypeMap[contest.type]}</Tag>
                                {presentation.badgeText ? <Tag color={presentation.badgeColor}>{presentation.badgeText}</Tag> : null}
                                {isTeamContest(contest) ? <Tag color="purple">团队协作</Tag> : <Tag>个人参赛</Tag>}
                              </div>
                            </div>
                          </div>
                        </div>
                        <div className="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 text-right">
                          <div className="text-xs uppercase tracking-[0.18em] text-text-secondary">当前状态</div>
                          <div className="mt-2">{renderStatusBadge(contest)}</div>
                        </div>
                      </div>

                      <p className="mt-5 line-clamp-2 text-sm leading-6 text-text-secondary">
                        {contest.description || '暂无描述'}
                      </p>

                      <div className="mt-5 grid gap-3 md:grid-cols-3">
                        <div className="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3">
                          <div className="text-xs uppercase tracking-[0.16em] text-text-secondary">赛程</div>
                          <div className="mt-2 text-sm font-medium text-slate-900">
                            {formatDateTime(contest.start_time, 'MM-DD HH:mm')} - {formatDateTime(contest.end_time, 'MM-DD HH:mm')}
                          </div>
                          <div className="mt-1 text-xs text-text-secondary">
                            {contest.status === 'published' ? `${formatRelativeTime(contest.start_time)}开始` : '比赛时间窗口'}
                          </div>
                        </div>
                        <div className="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3">
                          <div className="text-xs uppercase tracking-[0.16em] text-text-secondary">参赛人数</div>
                          <div className="mt-2 text-sm font-medium text-slate-900">{contest.participant_count || 0} 人</div>
                          <div className="mt-1 text-xs text-text-secondary">当前已报名参赛人数</div>
                        </div>
                        <div className="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3">
                          <div className="text-xs uppercase tracking-[0.16em] text-text-secondary">参赛形式</div>
                          <div className="mt-2 text-sm font-medium text-slate-900">
                            {contest.team_min_size || 1} - {contest.team_max_size || 1} 人 / 队
                          </div>
                          <div className="mt-1 text-xs text-text-secondary">
                            {isTeamContest(contest) ? '适合协同攻防' : '个人独立完成'}
                          </div>
                        </div>
                      </div>
                    </div>

                    <div className="flex flex-col justify-between gap-4 bg-[linear-gradient(180deg,#f8fbff_0%,#eef5ff_100%)] px-6 py-6">
                      <div>
                        <div className="text-xs uppercase tracking-[0.18em] text-text-secondary">当前操作</div>
                        <div className="mt-3 text-lg font-semibold text-slate-900">
                          {presentation.canEnter ? '进入正式赛场' : presentation.canRegister ? '完成报名准备' : '查看比赛信息'}
                        </div>
                        <p className="mt-2 text-sm leading-6 text-text-secondary">
                          {presentation.canEnter
                            ? '直接进入比赛页面开始操作。'
                            : presentation.canRegister
                              ? isTeamContest(contest)
                                ? '先完成组队，再进入正式比赛。'
                                : '立即报名后即可在比赛开始时进入赛场。'
                              : contest.status === 'ended'
                                ? '比赛已结束，可查看结果或赛后内容。'
                                : '当前阶段仅可查看详情。'}
                        </p>
                      </div>

                      <div className="flex flex-wrap justify-end gap-3">
                        {presentation.canRegister ? (
                          <Button
                            type="primary"
                            loading={registeringId === contest.id}
                            onClick={(event) => {
                              event.stopPropagation()
                              if (isTeamContest(contest)) {
                                navigate(`/contest/${contest.id}`)
                                return
                              }
                              void handleRegister(contest.id)
                            }}
                          >
                            {isTeamContest(contest) ? '组队报名' : '立即报名'}
                          </Button>
                        ) : null}

                        {presentation.canEnter ? (
                          <Button
                            type="primary"
                            onClick={(event) => {
                              event.stopPropagation()
                              navigate(contest.type === 'agent_battle' ? `/contest/${contest.id}/battle` : `/contest/${contest.id}/jeopardy`)
                            }}
                          >
                            进入比赛
                          </Button>
                        ) : null}

                        {!presentation.canRegister && !presentation.canEnter && contest.status === 'ongoing' ? (
                          <Button
                            onClick={(event) => {
                              event.stopPropagation()
                              navigate(`/contest/${contest.id}`)
                            }}
                          >
                            查看详情
                          </Button>
                        ) : null}

                        {!presentation.canRegister && contest.status === 'ended' ? (
                          <Button
                            onClick={(event) => {
                              event.stopPropagation()
                              navigate(`/contest/${contest.id}`)
                            }}
                          >
                            查看结果
                          </Button>
                        ) : null}
                      </div>
                    </div>
                  </div>
                </Card>
              )
            })}
          </div>
        )}
      </Spin>
    </div>
  )
}
