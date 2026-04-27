import { useCallback, useEffect, useMemo, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { Alert, Button, Card, Empty, Select, Slider, Spin, Statistic, Table, Tag } from 'antd'

import { PageHeader } from '@/components/common'
import { ReplaySnapshotTeamsTable } from '@/components/contest'
import { getAgentBattleRounds, getContest, getFinalRank, getReplayData } from '@/api/contest'
import { getRoundPhaseConfig } from '@/domains/contest/battle'
import {
  buildReplayRoundOptions,
  type FinalRankItem,
  getReplayCurrentSnapshot,
  type ReplaySnapshot,
} from '@/domains/contest/replay'
import { formatDateTime } from '@/utils/format'

export default function Replay() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const contestId = Number(id || '0')

  const [loading, setLoading] = useState(true)
  const [contestName, setContestName] = useState('')
  const [roundId, setRoundId] = useState<number | null>(null)
  const [rounds, setRounds] = useState<Array<{ id: number; round_number: number; phase?: string }>>([])
  const [roundOptions, setRoundOptions] = useState<Array<{ label: string; value: number }>>([])
  const [snapshots, setSnapshots] = useState<ReplaySnapshot[]>([])
  const [finalRank, setFinalRank] = useState<FinalRankItem[]>([])
  const [currentIndex, setCurrentIndex] = useState(0)
  const [selectedRoundPhase, setSelectedRoundPhase] = useState<string>('')
  const [blockRange, setBlockRange] = useState<{ start: number; end: number } | null>(null)

  const selectedRoundPhaseText = getRoundPhaseConfig(selectedRoundPhase)?.text || selectedRoundPhase

  const loadReplayData = useCallback(async (targetRoundId: number) => {
    const replay = await getReplayData(contestId, { round_id: targetRoundId })
    setSnapshots(replay.snapshots || [])
    setCurrentIndex(0)
    setBlockRange({
      start: replay.start_block || 0,
      end: replay.end_block || 0,
    })
  }, [contestId])

  useEffect(() => {
    const init = async () => {
      setLoading(true)
      try {
        const [contest, rounds, rankData] = await Promise.all([
          getContest(contestId),
          getAgentBattleRounds(contestId).catch(() => []),
          getFinalRank(contestId).catch(() => []),
        ])

        setContestName(contest.title || '')
        setFinalRank(rankData)
        setRounds(rounds)
        setRoundOptions(buildReplayRoundOptions(rounds))

        const latestRound = rounds[0]
        if (latestRound) {
          setRoundId(latestRound.id)
          setSelectedRoundPhase(latestRound.phase || '')
          await loadReplayData(latestRound.id)
        }
      } catch {
        navigate(-1)
      } finally {
        setLoading(false)
      }
    }

    void init()
  }, [contestId, loadReplayData, navigate])

  const currentSnapshot = useMemo(
    () => getReplayCurrentSnapshot(snapshots, currentIndex),
    [currentIndex, snapshots],
  )

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
        title="赛后回放"
        subtitle={contestName}
        showBack
        tags={<Tag color="purple">对抗赛复盘</Tag>}
      />

      <Card className="overflow-hidden border-0 shadow-sm" styles={{ body: { padding: 0 } }}>
        <div className="grid gap-px bg-slate-200 xl:grid-cols-[minmax(0,1.4fr)_320px]">
          <div className="bg-[linear-gradient(135deg,#f4f8ff_0%,#eef5ff_55%,#eefaf7_100%)] px-6 py-6 text-slate-900">
            <div className="text-xs uppercase tracking-[0.28em] text-sky-600">Battle Replay</div>
            <div className="mt-4 grid gap-3 md:grid-cols-2 xl:grid-cols-4">
              <div className="rounded-2xl border border-slate-200 bg-white px-4 py-4"><Statistic title="当前区块" value={currentSnapshot?.block || '-'} valueStyle={{ color: '#0f172a' }} /></div>
              <div className="rounded-2xl border border-slate-200 bg-white px-4 py-4"><Statistic title="当前帧" value={snapshots.length > 0 ? `${currentIndex + 1} / ${snapshots.length}` : '-'} valueStyle={{ color: '#0f172a' }} /></div>
              <div className="rounded-2xl border border-slate-200 bg-white px-4 py-4"><Statistic title="轮次阶段" value={selectedRoundPhaseText || '-'} valueStyle={{ color: '#0f172a' }} /></div>
              <div className="rounded-2xl border border-slate-200 bg-white px-4 py-4"><Statistic title="队伍数" value={currentSnapshot?.teams?.length || 0} valueStyle={{ color: '#0f172a' }} /></div>
            </div>
          </div>

          <div className="flex flex-col gap-4 bg-white px-6 py-6">
            <div className="text-xs uppercase tracking-[0.18em] text-text-secondary">Replay Control</div>
            <Select
              value={roundId || undefined}
              options={roundOptions}
              className="w-full"
              placeholder="选择轮次"
              onChange={async (value) => {
                setRoundId(value)
                const selectedRound = rounds.find((round) => round.id === value)
                setSelectedRoundPhase(selectedRound?.phase || '')
                await loadReplayData(value)
              }}
            />
            <Button onClick={() => setCurrentIndex(0)} disabled={snapshots.length === 0}>
              回到开头
            </Button>
            {blockRange ? (
              <div className="text-sm text-text-secondary">
                区块范围：{blockRange.start} - {blockRange.end}
              </div>
            ) : null}
          </div>
        </div>
      </Card>

      <div className="grid gap-6 xl:grid-cols-[minmax(0,1fr)_340px]">
        <div className="space-y-6">
          <Card title="时间轴控制" className="border-0 shadow-sm">
            <div className="mb-4 flex gap-4">
              <div className="text-sm text-text-secondary">
                通过时间轴回看某一轮中各个区块快照、队伍状态和关键事件。
              </div>
            </div>

            {snapshots.length > 0 ? (
              <>
                <div className="mb-3 text-sm text-text-secondary">
                  当前快照：第 {currentIndex + 1} / {snapshots.length} 帧，区块 {currentSnapshot?.block || '-'}
                </div>
                <Slider
                  min={0}
                  max={Math.max(snapshots.length - 1, 0)}
                  value={currentIndex}
                  onChange={setCurrentIndex}
                />
              </>
            ) : (
                <Alert type="info" showIcon message="当前轮次暂无回放快照" />
            )}
          </Card>

          <Card title="当前快照队伍状态" className="border-0 shadow-sm">
            <Table
              dataSource={currentSnapshot?.teams || []}
              rowKey="team_id"
              pagination={false}
              columns={[
                { title: '队伍 ID', dataIndex: 'team_id', key: 'team_id', width: 90 },
                { title: '得分', dataIndex: 'score', key: 'score', width: 120 },
                { title: '资源', dataIndex: 'resource', key: 'resource', width: 120 },
              ]}
              locale={{ emptyText: <Empty description="暂无队伍快照" /> }}
            />
          </Card>

          <Card title="当前快照事件" className="border-0 shadow-sm">
            <Table
              dataSource={currentSnapshot?.events || []}
              rowKey={(event, index) => `${event.event_type}-${index}`}
              pagination={false}
              columns={[
                { title: '事件类型', dataIndex: 'event_type', key: 'event_type', width: 120, render: (value: string) => <Tag color="blue">{value}</Tag> },
                { title: '发起方', dataIndex: 'actor_team', key: 'actor_team', width: 140, render: (value?: string) => value || '-' },
                { title: '目标', dataIndex: 'target_team', key: 'target_team', width: 140, render: (value?: string) => value || '-' },
                { title: '结果', dataIndex: 'action_result', key: 'action_result', width: 120, render: (value?: string) => value || '-' },
                { title: '描述', dataIndex: 'description', key: 'description' },
              ]}
              locale={{ emptyText: <Empty description="当前快照暂无事件" /> }}
            />
          </Card>
        </div>

        <div className="space-y-6 xl:sticky xl:top-6 xl:h-fit">
          <Card title="最终排名" className="border-0 shadow-sm">
            <ReplaySnapshotTeamsTable finalRank={finalRank} />
          </Card>

          <Card title="回放说明" className="border-0 shadow-sm">
            <Statistic title="当前区块" value={currentSnapshot?.block || '-'} />
            {selectedRoundPhaseText ? <div className="mt-3 text-sm text-text-secondary">轮次标签：{selectedRoundPhaseText}</div> : null}
            <Alert
              className="mt-4"
              type="info"
              showIcon
              message="复盘建议"
              description={`建议重点关注得分变化、资源变化和关键攻击事件，结合最终排名判断哪些动作真正改变了局势。更新时间：${formatDateTime(new Date().toISOString())}`}
            />
          </Card>
        </div>
      </div>
    </div>
  )
}
