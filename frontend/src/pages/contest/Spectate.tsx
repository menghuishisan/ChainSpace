import { useEffect, useMemo } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { Alert, Button, Card, Spin, Statistic, Tag } from 'antd'

import { PageHeader } from '@/components/common'
import EmptyState from '@/components/common/EmptyState'
import {
  SpectateEventTimelineTable,
  SpectateTeamStatusTable,
} from '@/components/contest'
import {
  getBattleScoreWeights,
  getRoundPhaseConfig,
  getRoundStatusConfig,
} from '@/domains/contest/battle'
import {
  buildRankedTeams,
  getSpectateEventTimeline,
  getSpectateHeadlineMetrics,
} from '@/domains/contest/spectate'
import { useContestStore } from '@/store'
import { formatDateTime } from '@/utils/format'

export default function Spectate() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const contestId = Number(id || '0')

  const {
    currentContest,
    currentRound,
    scoreboard,
    spectateData,
    loading,
    hydrateSpectateWorkspace,
    startPolling,
    stopPolling,
    reset,
  } = useContestStore()

  useEffect(() => {
    const init = async () => {
      try {
        await hydrateSpectateWorkspace(contestId)
        startPolling(contestId, 'spectate')
      } catch {
        navigate(-1)
      }
    }

    void init()

    return () => {
      stopPolling()
      reset()
    }
  }, [contestId, hydrateSpectateWorkspace, navigate, reset, startPolling, stopPolling])

  const rankedTeams = useMemo(
    () => buildRankedTeams(spectateData?.teams, scoreboard),
    [scoreboard, spectateData?.teams],
  )
  const eventTimeline = useMemo(
    () => getSpectateEventTimeline(spectateData),
    [spectateData],
  )
  const headline = useMemo(
    () => getSpectateHeadlineMetrics(rankedTeams),
    [rankedTeams],
  )
  const scoreWeights = getBattleScoreWeights(currentContest?.battle_orchestration)
  const roundStatusConfig = getRoundStatusConfig(spectateData?.round_status || currentRound?.status)
  const roundPhaseConfig = getRoundPhaseConfig(spectateData?.round_phase || currentRound?.phase)

  if (loading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <Spin size="large" />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="观战模式"
        subtitle={currentContest?.title || ''}
        showBack
        tags={<Tag color="purple">智能体博弈战</Tag>}
        extra={<Button onClick={() => navigate(`/contest/${contestId}/replay`)}>查看赛后回放</Button>}
      />

      <Card className="overflow-hidden border-0 shadow-sm" styles={{ body: { padding: 0 } }}>
        <div className="grid gap-px bg-slate-200 xl:grid-cols-[minmax(0,1.4fr)_320px]">
          <div className="bg-[linear-gradient(135deg,#f4f8ff_0%,#eef5ff_55%,#eefaf7_100%)] px-6 py-6 text-slate-900">
            <div className="text-xs uppercase tracking-[0.28em] text-sky-600">Live Spectate</div>
            <div className="mt-4 grid gap-3 md:grid-cols-2 xl:grid-cols-4">
              <div className="rounded-2xl border border-slate-200 bg-white px-4 py-4">
                <Statistic title="当前轮次" value={spectateData?.current_round || currentRound?.round_number || '-'} valueStyle={{ color: '#0f172a' }} />
              </div>
              <div className="rounded-2xl border border-slate-200 bg-white px-4 py-4">
                <Statistic title="轮次状态" value={roundStatusConfig?.text || spectateData?.round_status || currentRound?.status || '-'} valueStyle={{ color: '#0f172a' }} />
              </div>
              <div className="rounded-2xl border border-slate-200 bg-white px-4 py-4">
                <Statistic title="轮次阶段" value={roundPhaseConfig?.text || '-'} valueStyle={{ color: '#0f172a' }} />
              </div>
              <div className="rounded-2xl border border-slate-200 bg-white px-4 py-4">
                <Statistic title="总资源量" value={headline.totalResource} valueStyle={{ color: '#0f172a' }} />
              </div>
            </div>
            <Alert
              className="mt-5"
              type="info"
              showIcon
              message="观战重点"
              description="本页围绕局势变化、实时排名和关键事件展示比赛过程，而不是简单堆积日志。"
            />
          </div>

          <div className="flex flex-col gap-4 bg-white px-6 py-6">
            <Card size="small" className="border-0 bg-slate-50">
              <div className="text-xs uppercase tracking-[0.18em] text-text-secondary">Leading Team</div>
              <div className="mt-3 text-xl font-semibold text-slate-900">{headline.topTeam?.team_name || '暂无'}</div>
              <div className="mt-3 flex flex-wrap gap-2">
                <Tag color="blue">领先差距 {headline.scoreGap}</Tag>
                <Tag color="green">存活 {headline.aliveCount}</Tag>
              </div>
            </Card>
            <Card size="small" className="border-0 bg-slate-50">
              <div className="space-y-2 text-sm">
                <div className="flex justify-between"><span>资源控制</span><Tag color="green">{scoreWeights.resource || 0}</Tag></div>
                <div className="flex justify-between"><span>攻击收益</span><Tag color="red">{scoreWeights.attack || 0}</Tag></div>
                <div className="flex justify-between"><span>防守保全</span><Tag color="blue">{scoreWeights.defense || 0}</Tag></div>
                <div className="flex justify-between"><span>生存稳定</span><Tag color="purple">{scoreWeights.survival || 0}</Tag></div>
              </div>
            </Card>
            {spectateData?.round_end_time ? (
              <Alert
                type="warning"
                showIcon
                message="本轮结束时间"
                description={formatDateTime(spectateData.round_end_time)}
              />
            ) : null}
          </div>
        </div>
      </Card>

      <div className="grid gap-6 xl:grid-cols-[minmax(0,1fr)_360px]">
        <div className="space-y-6">
          <Card title="实时队伍态势" styles={{ body: { padding: 0 } }} className="border-0 shadow-sm">
            <SpectateTeamStatusTable teams={rankedTeams} />
          </Card>
        </div>

        <div className="space-y-6 xl:sticky xl:top-6 xl:h-fit">
          <Card title="关键事件流" styles={{ body: { padding: 0, maxHeight: 520, overflow: 'auto' } }} className="border-0 shadow-sm">
            <SpectateEventTimelineTable events={eventTimeline} />
          </Card>

          <Card title="比赛解读" className="border-0 shadow-sm">
            {headline.topTeam ? (
              <>
                <Alert
                  type="success"
                  showIcon
                  message={`当前领先队伍：${headline.topTeam.team_name}`}
                  description={`该队当前总分 ${headline.topTeam.total_score}，资源占有 ${headline.topTeam.resource_held}。`}
                />

                <div className="mt-4 space-y-3 text-sm">
                  <div className="rounded bg-gray-50 p-3">
                    <div className="mb-1 font-medium">领先差距</div>
                    <div className="text-text-secondary">
                      {headline.secondTeam
                        ? `相比第二名 ${headline.secondTeam.team_name} 领先 ${headline.scoreGap} 分。`
                        : '当前仅有一支队伍进入排行显示。'}
                    </div>
                  </div>

                  <div className="rounded bg-gray-50 p-3">
                    <div className="mb-1 font-medium">当前关注点</div>
                    <div className="text-text-secondary">
                      建议优先观察资源分是否持续拉开、攻击类事件是否连续发生，以及最近几条事件是否造成了排名变化。
                    </div>
                  </div>
                </div>
              </>
            ) : (
              <EmptyState description="比赛尚未进入可解读状态" />
            )}
          </Card>
        </div>
      </div>
    </div>
  )
}
