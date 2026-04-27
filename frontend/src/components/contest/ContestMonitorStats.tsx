import { Card, Col, Row } from 'antd'
import {
  CheckCircleOutlined,
  ClockCircleOutlined,
  PlayCircleOutlined,
  TeamOutlined,
  ThunderboltOutlined,
  TrophyOutlined,
} from '@ant-design/icons'

import type { Contest } from '@/types'
import type { ContestChallengeStat, ContestRoundInfo } from '@/types/presentation'
import { formatDuration } from '@/utils/format'
import {
  getFinishedRoundCount,
  getRunningRoundNumber,
} from '@/domains/contest/monitor'

interface ContestMonitorStatsProps {
  contest: Contest
  remainingTime: number
  participantCount: number
  challengeStats: ContestChallengeStat[]
  rounds: ContestRoundInfo[]
}

export default function ContestMonitorStats({
  contest,
  remainingTime,
  participantCount,
  challengeStats,
  rounds,
}: ContestMonitorStatsProps) {
  const isAgentBattle = contest.type === 'agent_battle'
  const stats = [
    {
      key: 'remaining',
      label: '剩余时间',
      value: formatDuration(remainingTime),
      icon: <ClockCircleOutlined />,
      tone: remainingTime < 3600 ? 'text-rose-600' : 'text-sky-600',
      description: remainingTime < 3600 ? '已进入最后一小时窗口' : '比赛仍处于正常推进阶段',
    },
    {
      key: 'participants',
      label: isAgentBattle ? '参赛队伍' : '参赛人数',
      value: participantCount,
      icon: <TeamOutlined />,
      tone: 'text-slate-900',
      description: isAgentBattle ? '参与当前对抗赛的队伍数量' : '已进入当前赛场的选手规模',
    },
    ...(isAgentBattle ? [
      {
        key: 'rounds',
        label: '轮次计划',
        value: `${rounds.length}`,
        icon: <ThunderboltOutlined />,
        tone: 'text-amber-600',
        description: `已完成 ${getFinishedRoundCount(rounds)} 个回合结算`,
      },
      {
        key: 'current-round',
        label: '当前轮次',
        value: getRunningRoundNumber(rounds),
        icon: <PlayCircleOutlined />,
        tone: 'text-emerald-600',
        description: '监控当前处于执行或结算中的比赛轮次',
      },
    ] : [
      {
        key: 'challenges',
        label: '赛题数量',
        value: challengeStats.length,
        icon: <TrophyOutlined />,
        tone: 'text-indigo-600',
        description: '当前解题赛已挂载的全部题目数量',
      },
      {
        key: 'solves',
        label: '累计解题',
        value: challengeStats.reduce((sum, item) => sum + item.solve_count, 0),
        icon: <CheckCircleOutlined />,
        tone: 'text-emerald-600',
        description: '方便识别热点题与当前解题密度',
      },
    ]),
  ]

  return (
    <Row gutter={[16, 16]} className="mb-6">
      {stats.map((item) => (
        <Col key={item.key} xs={24} sm={12} xl={6}>
          <Card className="h-full overflow-hidden border-0 shadow-sm" styles={{ body: { padding: 0 } }}>
            <div className="h-full bg-[linear-gradient(145deg,#f4f8ff_0%,#eef5ff_55%,#eefaf7_100%)] px-5 py-5 text-slate-900">
              <div className="flex items-start justify-between gap-4">
                <div>
                  <div className="text-xs uppercase tracking-[0.24em] text-slate-500">{item.label}</div>
                  <div className={`mt-3 text-3xl font-semibold ${item.tone}`}>{item.value}</div>
                </div>
                <div className="rounded-2xl border border-slate-200 bg-white p-3 text-lg text-slate-700">
                  {item.icon}
                </div>
              </div>
              <div className="mt-4 text-sm leading-6 text-slate-600">{item.description}</div>
            </div>
          </Card>
        </Col>
      ))}
    </Row>
  )
}
