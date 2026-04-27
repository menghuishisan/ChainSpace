import { useEffect, useMemo, useState } from 'react'
import { Button, Card, Descriptions, Space, Tag } from 'antd'
import { EyeOutlined, PlayCircleOutlined, RocketOutlined, TrophyOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'

import { ContestRankingTable } from '@/components/contest'
import { ContestStatusConfig, PageHeader, StatusTag } from '@/components/common'
import { getFinalRank, registerContest } from '@/api/contest'
import {
  getBattleConfigFromOrchestration,
  getRoundPhaseConfig,
  getRoundStatusConfig,
} from '@/domains/contest/battle'
import { buildContestAgentBattleDetailState, mapFinalRankToContestScores } from '@/domains/contest/detail'
import { getContestParticipationPresentation } from '@/domains/contest/presentation'
import { useContestStore } from '@/store'
import type { Contest, ContestScore } from '@/types'
import { ContestTypeMap } from '@/types'
import { formatDateTime } from '@/utils/format'

interface ContestAgentBattleDetailProps {
  contest: Contest
}

export default function ContestAgentBattleDetail({ contest }: ContestAgentBattleDetailProps) {
  const navigate = useNavigate()
  const contestId = contest.id

  const {
    scoreboard,
    currentRound,
    myTeam,
    hydrateBattleWorkspace,
    reset,
  } = useContestStore()

  const [isRegistered, setIsRegistered] = useState(Boolean(contest.is_registered))
  const [finalRank, setFinalRank] = useState<ContestScore[]>([])

  const effectiveContest = useMemo(() => ({
    ...contest,
    is_registered: isRegistered,
  }), [contest, isRegistered])

  const battleConfig = useMemo(
    () => getBattleConfigFromOrchestration(contest.battle_orchestration),
    [contest.battle_orchestration],
  )

  const detailState = useMemo(
    () => buildContestAgentBattleDetailState(contest),
    [contest],
  )

  const participation = useMemo(() => (
    getContestParticipationPresentation(effectiveContest, {
      hasTeam: Boolean(myTeam?.name),
    })
  ), [effectiveContest, myTeam?.name])

  const roundStatusConfig = getRoundStatusConfig(currentRound?.status)
  const roundPhaseConfig = getRoundPhaseConfig(currentRound?.phase)

  useEffect(() => {
    const init = async () => {
      try {
        await hydrateBattleWorkspace(contestId)
      } catch {
        // Keep detail page readable even if runtime state is temporarily unavailable.
      }

      if (contest.status !== 'ended') {
        setFinalRank([])
        return
      }

      try {
        const data = await getFinalRank(contestId)
        setFinalRank(mapFinalRankToContestScores(data))
      } catch {
        setFinalRank([])
      }
    }

    void init()

    return () => {
      reset()
    }
  }, [contest.status, contestId, hydrateBattleWorkspace, reset])

  const handleRegister = async () => {
    await registerContest(contestId)
    setIsRegistered(true)
    await hydrateBattleWorkspace(contestId)
  }

  return (
    <div className="space-y-4">
      <PageHeader
        title={contest.title}
        subtitle="共享链上的策略智能体对抗赛"
        showBack
        tags={(
          <>
            <Tag color="purple">{ContestTypeMap[contest.type]}</Tag>
            <StatusTag status={contest.status} statusMap={ContestStatusConfig} />
            {participation.badgeText && <Tag color={participation.badgeColor}>{participation.badgeText}</Tag>}
          </>
        )}
        extra={(
          <Space>
            {participation.canRegister && (
              <Button type="primary" icon={<RocketOutlined />} onClick={() => void handleRegister()}>
                立即报名
              </Button>
            )}
            {participation.canEnter && (
              <Button type="primary" icon={<PlayCircleOutlined />} onClick={() => navigate(`/contest/${contestId}/battle`)}>
                进入比赛
              </Button>
            )}
            {detailState.canSpectate && (
              <Button icon={<EyeOutlined />} onClick={() => navigate(`/contest/${contestId}/spectate`)}>
                进入观战
              </Button>
            )}
            {detailState.canReplay && (
              <Button onClick={() => navigate(`/contest/${contestId}/replay`)}>
                查看回放
              </Button>
            )}
          </Space>
        )}
      />

      <div className="grid gap-3 xl:grid-cols-[minmax(0,1fr)_290px]">
        <div className="space-y-3">
          <Card className="overflow-hidden border-0 shadow-sm" styles={{ body: { padding: 0 } }}>
            <div className="bg-[linear-gradient(135deg,#f4f8ff_0%,#eef5ff_55%,#eefaf7_100%)] px-4 py-3.5 text-slate-900">
            <div className="text-xs uppercase tracking-[0.28em] text-sky-600">对抗赛场</div>
            <div className="mt-2 flex flex-wrap gap-2">
              <Tag color="purple">{ContestTypeMap[contest.type]}</Tag>
              <StatusTag status={contest.status} statusMap={ContestStatusConfig} />
              {participation.badgeText ? <Tag color={participation.badgeColor}>{participation.badgeText}</Tag> : null}
            </div>
            <div className="mt-2.5 text-[1.65rem] font-semibold leading-tight">{contest.title}</div>
            <p className="mt-2 max-w-3xl whitespace-pre-wrap text-sm leading-6 text-slate-700">
              {contest.description || '本场比赛采用策略智能体对抗模式，围绕资源控制、攻防博弈与生存稳定进行多轮结算。'}
            </p>
            <div className="mt-3 grid gap-2 md:grid-cols-2 xl:grid-cols-4">
              <div className="rounded-2xl border border-slate-200 bg-white px-3.5 py-2">
                <div className="text-xs uppercase tracking-[0.18em] text-slate-500">当前轮次</div>
                <div className="mt-1 text-sm font-medium text-slate-900">{currentRound ? `第 ${currentRound.round_number} 轮` : '尚未开始'}</div>
              </div>
              <div className="rounded-2xl border border-slate-200 bg-white px-3.5 py-2">
                <div className="text-xs uppercase tracking-[0.18em] text-slate-500">阶段</div>
                <div className="mt-1 text-sm font-medium text-slate-900">{roundPhaseConfig?.text || '待推进'}</div>
              </div>
              <div className="rounded-2xl border border-slate-200 bg-white px-3.5 py-2">
                <div className="text-xs uppercase tracking-[0.18em] text-slate-500">参赛人数</div>
                <div className="mt-1 text-sm font-medium text-slate-900">{contest.participant_count || 0} 人</div>
              </div>
              <div className="rounded-2xl border border-slate-200 bg-white px-3.5 py-2">
                <div className="text-xs uppercase tracking-[0.18em] text-slate-500">当前身份</div>
                <div className="mt-1 text-sm font-medium text-slate-900">{myTeam?.name || participation.teamSummaryText}</div>
              </div>
            </div>
            </div>
          </Card>

          <Card
            title="比赛总览"
            className="border-0 shadow-sm"
            styles={{
              header: { background: 'linear-gradient(90deg, rgba(14,165,233,0.08), rgba(16,185,129,0.08))', minHeight: 44, padding: '0 16px' },
              body: { padding: 16 },
            }}
          >
            <Descriptions size="small" column={{ xs: 1, sm: 2 }}>
              <Descriptions.Item label="比赛类型">{ContestTypeMap[contest.type]}</Descriptions.Item>
              <Descriptions.Item label="开始时间">{formatDateTime(contest.start_time)}</Descriptions.Item>
              <Descriptions.Item label="结束时间">{formatDateTime(contest.end_time)}</Descriptions.Item>
              {contest.registration_end && (
                <Descriptions.Item label="报名截止">{formatDateTime(contest.registration_end)}</Descriptions.Item>
              )}
              <Descriptions.Item label="参赛人数">{contest.participant_count || 0} 人</Descriptions.Item>
              <Descriptions.Item label="队伍人数">
                {contest.team_min_size || 1} - {contest.team_max_size || 1} 人
              </Descriptions.Item>
              <Descriptions.Item label="当前轮次">
                {currentRound ? `第 ${currentRound.round_number} 轮` : '尚未开始'}
              </Descriptions.Item>
              <Descriptions.Item label="轮次状态">
                {roundStatusConfig ? <Tag color={roundStatusConfig.color}>{roundStatusConfig.text}</Tag> : '-'}
              </Descriptions.Item>
              <Descriptions.Item label="轮次阶段">
                {roundPhaseConfig ? <Tag color={roundPhaseConfig.color}>{roundPhaseConfig.text}</Tag> : '-'}
              </Descriptions.Item>
              <Descriptions.Item label="我的身份">
                {myTeam?.name || participation.teamSummaryText}
              </Descriptions.Item>
            </Descriptions>
          </Card>

          <Card
            title="规则模型"
            className="border-0 shadow-sm"
            styles={{
              header: { background: 'linear-gradient(90deg, rgba(14,165,233,0.08), rgba(34,197,94,0.08))', minHeight: 44, padding: '0 16px' },
              body: { padding: 16 },
            }}
          >
            <Descriptions bordered size="small" column={2}>
              <Descriptions.Item label="策略接口">{battleConfig.judge.strategy_interface || '-'}</Descriptions.Item>
              <Descriptions.Item label="资源模型">{battleConfig.judge.resource_model || '-'}</Descriptions.Item>
              <Descriptions.Item label="评分模型">{battleConfig.judge.scoring_model || '-'}</Descriptions.Item>
              <Descriptions.Item label="总轮数">{battleConfig.lifecycle.total_rounds || 0}</Descriptions.Item>
              <Descriptions.Item label="单轮时长">{battleConfig.lifecycle.round_duration_seconds || 0} 秒</Descriptions.Item>
              <Descriptions.Item label="升级窗口">{battleConfig.lifecycle.upgrade_window_seconds || 0} 秒</Descriptions.Item>
              <Descriptions.Item label="允许动作" span={2}>
                {(battleConfig.judge.allowed_actions || []).map((action) => (
                  <Tag key={action}>{action}</Tag>
                ))}
              </Descriptions.Item>
              <Descriptions.Item label="资源分权重">{detailState.scoreWeights.resource || 0}</Descriptions.Item>
              <Descriptions.Item label="攻击分权重">{detailState.scoreWeights.attack || 0}</Descriptions.Item>
              <Descriptions.Item label="防守分权重">{detailState.scoreWeights.defense || 0}</Descriptions.Item>
              <Descriptions.Item label="生存分权重">{detailState.scoreWeights.survival || 0}</Descriptions.Item>
            </Descriptions>
          </Card>
        </div>

        <div className="space-y-3 xl:sticky xl:top-5 xl:h-fit">
          <Card className="border-0 shadow-sm" styles={{ body: { padding: 14 } }}>
            <div className="rounded-2xl border border-slate-200 bg-slate-50 p-3.5">
              <div className="text-xs uppercase tracking-[0.18em] text-text-secondary">比赛状态</div>
              <div className="mt-1.5 text-xl font-semibold leading-tight text-slate-900">
                {contest.status === 'published' ? '报名准备阶段' : contest.status === 'ongoing' ? '比赛进行中' : contest.status === 'ended' ? '赛后复盘阶段' : '赛场搭建中'}
              </div>
              <p className="mt-1.5 text-sm leading-6 text-text-secondary">
                在这里直接完成报名、进入对局、观战和回放。
              </p>
              <Space wrap size={[8, 8]} className="mt-3">
                {participation.canRegister && (
                  <Button type="primary" icon={<RocketOutlined />} onClick={() => void handleRegister()}>
                    立即报名
                  </Button>
                )}
                {participation.canEnter && (
                  <Button type="primary" icon={<PlayCircleOutlined />} onClick={() => navigate(`/contest/${contestId}/battle`)}>
                    进入比赛
                  </Button>
                )}
                {detailState.canSpectate && (
                  <Button icon={<EyeOutlined />} onClick={() => navigate(`/contest/${contestId}/spectate`)}>
                    进入观战
                  </Button>
                )}
                {detailState.canReplay && (
                  <Button onClick={() => navigate(`/contest/${contestId}/replay`)}>
                    查看回放
                  </Button>
                )}
              </Space>
            </div>
          </Card>

          <Card title="赛场状态" className="border-0 shadow-sm" styles={{ header: { minHeight: 42, padding: '0 14px' }, body: { padding: 14 } }}>
            <div className="flex items-center gap-3 rounded-2xl border border-slate-200 bg-slate-50 px-3.5 py-3">
              <TrophyOutlined className="text-2xl text-warning" />
              <div className="text-sm">
                {contest.status === 'published' && <p className="text-success">比赛正在报名中</p>}
                {contest.status === 'ongoing' && <p className="text-primary">比赛进行中，可进入对局与观战页面</p>}
                {contest.status === 'ended' && <p className="text-text-secondary">比赛已结束，可查看最终排名与赛后回放</p>}
                {contest.status === 'draft' && <p className="text-text-secondary">比赛仍在准备阶段</p>}
              </div>
            </div>
          </Card>

          <Card title={contest.status === 'ended' ? '最终排名' : '实时排名'} className="border-0 shadow-sm" styles={{ header: { minHeight: 42, padding: '0 14px' }, body: { padding: 14 } }}>
            <ContestRankingTable
              data={contest.status === 'ended' ? finalRank : (scoreboard?.list || [])}
              emptyText={contest.status === 'ended' ? '暂无最终排名数据' : '暂无实时排名数据'}
            />
          </Card>
        </div>
      </div>
    </div>
  )
}
