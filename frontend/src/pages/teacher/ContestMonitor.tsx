import { PlusOutlined, ReloadOutlined } from '@ant-design/icons'
import {
  Alert,
  Button,
  Card,
  DatePicker,
  Empty,
  Form,
  InputNumber,
  Modal,
  Select,
  Space,
  Typography,
  message,
} from 'antd'
import { useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'

import { createRound, getContests } from '@/api/contest'
import { PageHeader } from '@/components/common'
import {
  ContestChallengeStatsTable,
  ContestMonitorStats,
  ContestRankingTable,
  ContestRoundManagementTable,
} from '@/components/contest'
import {
  buildChallengeStats,
  getContestMonitorRemainingSeconds,
  getContestScoreboardPreview,
  getNextRoundNumber,
} from '@/domains/contest/monitor'
import { useContestStore } from '@/store'
import type { Contest } from '@/types'

export default function TeacherContestMonitor() {
  const navigate = useNavigate()
  const {
    currentContest,
    challenges,
    scoreboard,
    rounds,
    loading,
    setContest,
    hydrateMonitorWorkspace,
    startPolling,
    stopPolling,
  } = useContestStore()

  const [contests, setContests] = useState<Contest[]>([])
  const [selectedContestId, setSelectedContestId] = useState<number | null>(null)
  const [remainingTime, setRemainingTime] = useState(0)
  const [roundModalVisible, setRoundModalVisible] = useState(false)
  const [roundActionLoading, setRoundActionLoading] = useState(false)
  const [roundForm] = Form.useForm()

  const selectedContest = currentContest
  const isAgentBattle = selectedContest?.type === 'agent_battle'
  const challengeStats = useMemo(() => buildChallengeStats(challenges), [challenges])

  useEffect(() => {
    const fetchContests = async () => {
      try {
        const result = await getContests({ page: 1, page_size: 100, status: 'ongoing' })
        setContests(result.list)
        if (result.list.length === 0) {
          return
        }

        const firstContest = result.list[0]
        setSelectedContestId(firstContest.id)
        setContest(firstContest)
        await hydrateMonitorWorkspace(firstContest.id, firstContest)
        startPolling(firstContest.id, 'monitor')
      } catch {
        setContests([])
      }
    }

    void fetchContests()

    return () => {
      stopPolling()
    }
  }, [hydrateMonitorWorkspace, setContest, startPolling, stopPolling])

  useEffect(() => {
    setRemainingTime(getContestMonitorRemainingSeconds(selectedContest))

    if (!selectedContest) {
      return
    }

    const timer = window.setInterval(() => {
      setRemainingTime(getContestMonitorRemainingSeconds(selectedContest))
    }, 1000)

    return () => window.clearInterval(timer)
  }, [selectedContest])

  const handleContestChange = async (contestId: number) => {
    const contest = contests.find((item) => item.id === contestId) || null
    setSelectedContestId(contestId)
    setContest(contest)
    stopPolling()
    await hydrateMonitorWorkspace(contestId, contest)
    startPolling(contestId, 'monitor')
  }

  const handleCreateRound = async () => {
    if (!selectedContestId) {
      return
    }

    const values = await roundForm.validateFields()
    const startTime = values.start_time.toDate()
    const endTime = values.end_time.toDate()

    if (endTime <= startTime) {
      message.error('结束时间必须晚于开始时间')
      return
    }

    try {
      setRoundActionLoading(true)
      await createRound(selectedContestId, {
        round_number: values.round_number,
        start_time: startTime.toISOString(),
        end_time: endTime.toISOString(),
      })
      message.success(`第 ${values.round_number} 轮已创建`)
      setRoundModalVisible(false)
      roundForm.resetFields()
      await hydrateMonitorWorkspace(selectedContestId, selectedContest)
    } finally {
      setRoundActionLoading(false)
    }
  }

  const openCreateRoundModal = () => {
    roundForm.setFieldsValue({ round_number: getNextRoundNumber(rounds) })
    setRoundModalVisible(true)
  }

  if (contests.length === 0) {
    return (
      <div>
        <PageHeader title="比赛监控" subtitle="实时查看比赛状态、得分变化与轮次进展" />
        <Card className="border-0 shadow-sm">
          <Empty description="当前没有进行中的比赛" />
        </Card>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="比赛监控"
        subtitle="查看实时赛况、轮次进度和榜单变化"
        extra={(
          <Space wrap>
            <Select
              value={selectedContestId || undefined}
              onChange={(value) => void handleContestChange(value)}
              options={contests.map((contest) => ({ label: contest.title, value: contest.id }))}
              className="w-64"
            />
            <Button
              icon={<ReloadOutlined />}
              onClick={() => selectedContestId && void hydrateMonitorWorkspace(selectedContestId, selectedContest)}
              loading={loading}
            >
              刷新
            </Button>
          </Space>
        )}
      />

      {selectedContest ? (
        <>
          <Card className="overflow-hidden border-0 shadow-sm" styles={{ body: { padding: 0 } }}>
            <div className="grid gap-px bg-slate-200 xl:grid-cols-[1.4fr_repeat(3,minmax(0,1fr))]">
              <div className="bg-[linear-gradient(135deg,#09111f_0%,#132238_58%,#1d4ed8_100%)] px-6 py-6 text-white">
                <div className="text-xs uppercase tracking-[0.28em] text-cyan-200">实时监控</div>
                <div className="mt-3 text-2xl font-semibold">{selectedContest.title}</div>
                <p className="mt-3 text-sm leading-6 text-slate-200">
                  {isAgentBattle
                    ? '当前页聚焦轮次计划、实时排行、回合状态与结算节奏，方便教师值守比赛。'
                    : '当前页聚焦赛题热度、解题趋势与排行榜变化，适合解题赛过程监控与教学观摩。'}
                </p>
              </div>
              <div className="bg-white px-5 py-5">
                <div className="text-xs uppercase tracking-[0.22em] text-text-secondary">比赛类型</div>
                <div className="mt-3 text-2xl font-semibold text-slate-900">
                  {isAgentBattle ? '对抗赛' : '解题赛'}
                </div>
                <div className="mt-2 text-sm text-text-secondary">{selectedContest.status}</div>
              </div>
              <div className="bg-white px-5 py-5">
                <div className="text-xs uppercase tracking-[0.22em] text-text-secondary">观测目标</div>
                <div className="mt-3 text-2xl font-semibold text-slate-900">
                  {isAgentBattle ? `${rounds.length} 轮计划` : `${challengeStats.length} 道题`}
                </div>
                <div className="mt-2 text-sm text-text-secondary">
                  {isAgentBattle ? '覆盖准备、执行和结算阶段' : '覆盖解题率、一血与榜单变化'}
                </div>
              </div>
              <div className="bg-white px-5 py-5">
                <div className="text-xs uppercase tracking-[0.22em] text-text-secondary">当前入口</div>
                <div className="mt-3 text-2xl font-semibold text-slate-900">监控页</div>
                <div className="mt-2 text-sm text-text-secondary">与参赛页分离，便于教师集中查看</div>
              </div>
            </div>
          </Card>

          <ContestMonitorStats
            contest={selectedContest}
            remainingTime={remainingTime}
            participantCount={scoreboard?.list.length || 0}
            challengeStats={challengeStats}
            rounds={rounds}
          />

          <div className="grid gap-6 xl:grid-cols-[minmax(0,1.55fr)_minmax(320px,0.95fr)]">
            <Card
              title="实时排行榜"
              className="border-0 shadow-sm"
              extra={scoreboard && scoreboard.list.length > 20 ? (
                <Button type="link" onClick={() => navigate(`/contest/${selectedContestId}`)}>
                  查看完整详情
                </Button>
              ) : null}
            >
              <Alert
                type="info"
                showIcon
                className="mb-4"
                message={isAgentBattle ? '榜单用于观察回合得分与阶段波动' : '榜单用于观察解题进度与拉分节奏'}
                description={isAgentBattle
                  ? '建议结合右侧轮次计划与事件变化一起查看，快速判断当前轮的推进情况。'
                  : '建议结合右侧题目统计识别热点题、一血题以及高分低解题率题。'}
              />
              <div className="overflow-hidden rounded-2xl border border-slate-200">
                <ContestRankingTable
                  data={getContestScoreboardPreview(scoreboard)}
                  emptyText="暂无实时排行数据"
                />
              </div>
            </Card>

            <div className="space-y-6">
              {isAgentBattle ? (
                <Card
                  title="轮次管理"
                  className="border-0 shadow-sm"
                  extra={(
                    <Button type="primary" size="small" icon={<PlusOutlined />} onClick={openCreateRoundModal}>
                      创建轮次
                    </Button>
                  )}
                >
                  <Typography.Paragraph type="secondary" className="mb-4">
                    轮次会按计划时间自动推进，这里只保留时间和编号配置，不再额外堆叠控制说明。
                  </Typography.Paragraph>
                  <div className="overflow-hidden rounded-2xl border border-slate-200">
                    <ContestRoundManagementTable rounds={rounds} loading={loading} />
                  </div>
                </Card>
              ) : (
                <Card title="题目统计" className="border-0 shadow-sm">
                  <Alert
                    type="info"
                    showIcon
                    className="mb-4"
                    message="热点题洞察"
                    description="解题率和一血信息会帮助教师快速判断赛题难度分布，以及是否需要在赛中补充提示。"
                  />
                  <div className="overflow-hidden rounded-2xl border border-slate-200">
                    <ContestChallengeStatsTable
                      challengeStats={challengeStats}
                      scoreboard={scoreboard}
                      loading={loading}
                    />
                  </div>
                </Card>
              )}
            </div>
          </div>
        </>
      ) : null}

      <Modal
        title="创建新轮次"
        open={roundModalVisible}
        onCancel={() => setRoundModalVisible(false)}
        onOk={() => void handleCreateRound()}
        confirmLoading={roundActionLoading}
      >
        <Form form={roundForm} layout="vertical" className="mt-4">
          <Form.Item
            name="round_number"
            label="轮次编号"
            rules={[{ required: true, message: '请输入轮次编号' }]}
          >
            <InputNumber min={1} className="w-full" />
          </Form.Item>
          <Form.Item
            name="start_time"
            label="计划开始时间"
            rules={[{ required: true, message: '请选择开始时间' }]}
          >
            <DatePicker showTime format="YYYY-MM-DD HH:mm" className="w-full" />
          </Form.Item>
          <Form.Item
            name="end_time"
            label="计划结束时间"
            rules={[{ required: true, message: '请选择结束时间' }]}
          >
            <DatePicker showTime format="YYYY-MM-DD HH:mm" className="w-full" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
