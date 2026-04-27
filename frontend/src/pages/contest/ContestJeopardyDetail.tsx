import { PlusOutlined, TrophyOutlined } from '@ant-design/icons'
import {
  Button,
  Card,
  Descriptions,
  InputNumber,
  Modal,
  Select,
  Tag,
  message,
} from 'antd'
import { useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'

import { getChallenges } from '@/api/challenge'
import {
  addChallengeToContest,
  createTeam,
  joinTeam,
  registerContest,
  removeChallengeFromContest,
} from '@/api/contest'
import { ContestStatusConfig, PageHeader, StatusTag } from '@/components/common'
import {
  ContestChallengeAdminTable,
  ContestChallengePreviewTable,
  ContestJeopardyRegistrationModal,
  ContestTeamCard,
} from '@/components/contest'
import {
  getContestChallengePreviewAccess,
  getContestStatusSummary,
  getRuntimeProfileLabel,
} from '@/domains/contest/jeopardy'
import { getContestParticipationPresentation } from '@/domains/contest/presentation'
import { useContestStore, useUserStore } from '@/store'
import type { Challenge, Contest } from '@/types'
import { ContestTypeMap } from '@/types'
import { formatDateTime } from '@/utils/format'

interface ContestJeopardyDetailProps {
  contest: Contest
}

export default function ContestJeopardyDetail({ contest }: ContestJeopardyDetailProps) {
  const navigate = useNavigate()
  const { isAdmin, isSchoolAdmin, isTeacher } = useUserStore()
  const {
    challenges: publicChallenges,
    contestChallenges,
    myTeam,
    loading,
    hydrateJeopardyDetail,
    reset,
  } = useContestStore()

  const isManager = isAdmin() || isSchoolAdmin() || isTeacher()
  const contestId = contest.id
  const isTeamContest = (contest.team_max_size || 1) > 1

  const [registerModalVisible, setRegisterModalVisible] = useState(false)
  const [teamModalVisible, setTeamModalVisible] = useState(false)
  const [challengeModalVisible, setChallengeModalVisible] = useState(false)
  const [actionLoading, setActionLoading] = useState(false)
  const [joinTeamCode, setJoinTeamCode] = useState('')
  const [availableChallenges, setAvailableChallenges] = useState<Challenge[]>([])
  const [selectedChallengeId, setSelectedChallengeId] = useState<number | undefined>()
  const [addPoints, setAddPoints] = useState(100)
  const [isRegistered, setIsRegistered] = useState(Boolean(contest.is_registered))

  const effectiveContest = useMemo(() => ({
    ...contest,
    is_registered: isRegistered,
  }), [contest, isRegistered])

  const participation = useMemo(() => (
    getContestParticipationPresentation(effectiveContest, {
      isManager,
      hasTeam: Boolean(myTeam),
    })
  ), [effectiveContest, isManager, myTeam])

  const challengePreviewAccess = useMemo(
    () => getContestChallengePreviewAccess(contest.status, isRegistered, isManager),
    [contest.status, isManager, isRegistered],
  )

  const statusSummary = useMemo(
    () => getContestStatusSummary(contest.status, isRegistered),
    [contest.status, isRegistered],
  )

  useEffect(() => {
    void hydrateJeopardyDetail(contestId, {
      contest,
      isManager,
      isRegistered,
      isTeamContest,
      useReviewChallenges: challengePreviewAccess.useReviewChallenges,
      showChallengePreview: challengePreviewAccess.showChallengePreview,
    })

    return () => {
      reset()
    }
  }, [
    challengePreviewAccess.showChallengePreview,
    challengePreviewAccess.useReviewChallenges,
    contest,
    contestId,
    hydrateJeopardyDetail,
    isManager,
    isRegistered,
    isTeamContest,
    reset,
  ])

  const handleOpenAddChallenge = async () => {
    try {
      const result = await getChallenges({ page: 1, page_size: 100 })
      const existingIds = new Set(contestChallenges.map((item) => item.challenge_id))
      setAvailableChallenges((result.list || []).filter((challenge) => !existingIds.has(challenge.id)))
    } finally {
      setChallengeModalVisible(true)
    }
  }

  const refreshDetail = async (registered = isRegistered) => {
    await hydrateJeopardyDetail(contestId, {
      contest,
      isManager,
      isRegistered: registered,
      isTeamContest,
      useReviewChallenges: challengePreviewAccess.useReviewChallenges,
      showChallengePreview: challengePreviewAccess.showChallengePreview,
    })
  }

  const handleAddChallenge = async () => {
    if (!selectedChallengeId) {
      message.warning('请选择题目')
      return
    }

    await addChallengeToContest(contestId, {
      challenge_id: selectedChallengeId,
      points: addPoints,
      is_visible: true,
    })
    message.success('题目已加入比赛')
    setChallengeModalVisible(false)
    setSelectedChallengeId(undefined)
    setAddPoints(100)
    await refreshDetail()
  }

  const handleRemoveChallenge = async (challengeId: number) => {
    await removeChallengeFromContest(contestId, challengeId)
    message.success('题目已从比赛中移除')
    await refreshDetail()
  }

  const handleRegister = async () => {
    setActionLoading(true)
    try {
      await registerContest(contestId)
      message.success('报名成功')
      setRegisterModalVisible(false)
      setIsRegistered(true)
      await refreshDetail(true)
    } finally {
      setActionLoading(false)
    }
  }

  const handleCreateTeam = async (values: { name: string }) => {
    setActionLoading(true)
    try {
      const createdTeam = await createTeam({ name: values.name, contest_id: contestId })
      setIsRegistered(true)
      message.success(`队伍创建成功，邀请码：${createdTeam.invite_code || '-'}`)
      setTeamModalVisible(false)
      await refreshDetail(true)
    } finally {
      setActionLoading(false)
    }
  }

  const handleJoinTeam = async () => {
    if (!joinTeamCode.trim()) {
      message.error('请输入队伍邀请码')
      return
    }

    setActionLoading(true)
    try {
      await joinTeam(joinTeamCode.trim())
      setIsRegistered(true)
      message.success('加入队伍成功')
      setTeamModalVisible(false)
      setJoinTeamCode('')
      await refreshDetail(true)
    } finally {
      setActionLoading(false)
    }
  }

  const handleEnterContest = () => {
    navigate(`/contest/${contestId}/jeopardy`)
  }

  return (
    <div className="space-y-4">
      <PageHeader
        title={contest.title}
        subtitle="查看比赛信息、完成报名组队，并在比赛开始后进入赛场"
        showBack
        tags={(
          <>
            <Tag color="blue">{ContestTypeMap[contest.type]}</Tag>
            <StatusTag status={contest.status} statusMap={ContestStatusConfig} />
            {participation.badgeText ? <Tag color={participation.badgeColor}>{participation.badgeText}</Tag> : null}
          </>
        )}
        extra={
          participation.canEnter ? (
            <Button type="primary" size="large" onClick={handleEnterContest}>进入比赛</Button>
          ) : participation.canRegister ? (
            <Button
              type="primary"
              size="large"
              onClick={() => (isTeamContest ? setTeamModalVisible(true) : setRegisterModalVisible(true))}
            >
              立即报名
            </Button>
          ) : null
        }
      />

      <div className="grid gap-3 xl:grid-cols-[minmax(0,1fr)_280px]">
        <div className="space-y-3">
          <Card className="overflow-hidden border-0 shadow-sm" styles={{ body: { padding: 0 } }}>
            <div className="bg-[linear-gradient(135deg,#081a2f_0%,#0f2744_58%,#13325b_100%)] px-4 py-3.5 text-white">
              <div className="text-xs uppercase tracking-[0.28em] text-cyan-200">比赛概览</div>
              <div className="mt-2 flex flex-wrap gap-2">
                <Tag color="blue">{ContestTypeMap[contest.type]}</Tag>
                <StatusTag status={contest.status} statusMap={ContestStatusConfig} />
                {participation.badgeText ? <Tag color={participation.badgeColor}>{participation.badgeText}</Tag> : null}
              </div>
              <div className="mt-2.5 text-[1.55rem] font-semibold leading-tight">{contest.title}</div>
              <p className="mt-2 max-w-3xl whitespace-pre-wrap text-sm leading-6 text-slate-200">
                {contest.description || '暂无描述'}
              </p>
              <div className="mt-3 grid gap-2 md:grid-cols-2 xl:grid-cols-4">
                <div className="rounded-2xl border border-white/10 bg-white/5 px-3.5 py-2">
                  <div className="text-xs uppercase tracking-[0.18em] text-cyan-200">开始时间</div>
                  <div className="mt-1 text-sm font-medium">{formatDateTime(contest.start_time)}</div>
                </div>
                <div className="rounded-2xl border border-white/10 bg-white/5 px-3.5 py-2">
                  <div className="text-xs uppercase tracking-[0.18em] text-cyan-200">结束时间</div>
                  <div className="mt-1 text-sm font-medium">{formatDateTime(contest.end_time)}</div>
                </div>
                <div className="rounded-2xl border border-white/10 bg-white/5 px-3.5 py-2">
                  <div className="text-xs uppercase tracking-[0.18em] text-cyan-200">参赛人数</div>
                  <div className="mt-1 text-sm font-medium">{contest.participant_count || 0} 人</div>
                </div>
                <div className="rounded-2xl border border-white/10 bg-white/5 px-3.5 py-2">
                  <div className="text-xs uppercase tracking-[0.18em] text-cyan-200">队伍人数</div>
                  <div className="mt-1 text-sm font-medium">{contest.team_min_size || 1} - {contest.team_max_size || 1} 人</div>
                </div>
              </div>
            </div>
          </Card>

          <Card
            title="比赛信息"
            className="border-0 shadow-sm"
            styles={{
              header: { background: 'linear-gradient(90deg, rgba(24,144,255,0.08), rgba(82,196,26,0.08))', minHeight: 44, padding: '0 16px' },
              body: { padding: 16 },
            }}
          >
            <Descriptions size="small" column={{ xs: 1, md: 2 }}>
              <Descriptions.Item label="比赛类型">{ContestTypeMap[contest.type]}</Descriptions.Item>
              <Descriptions.Item label="当前状态">{statusSummary}</Descriptions.Item>
              <Descriptions.Item label="开始时间">{formatDateTime(contest.start_time)}</Descriptions.Item>
              <Descriptions.Item label="结束时间">{formatDateTime(contest.end_time)}</Descriptions.Item>
              {contest.registration_end ? (
                <Descriptions.Item label="报名截止">{formatDateTime(contest.registration_end)}</Descriptions.Item>
              ) : null}
              <Descriptions.Item label="参赛人数">{contest.participant_count || 0} 人</Descriptions.Item>
              <Descriptions.Item label="队伍人数">
                {contest.team_min_size || 1} - {contest.team_max_size || 1} 人
              </Descriptions.Item>
            </Descriptions>
          </Card>

          {challengePreviewAccess.showChallengePreview ? (
            <Card
              title={contest.status === 'ended' ? '赛后题目回顾' : '题目预览'}
              className="border-0 shadow-sm"
              extra={<Tag color="purple">{publicChallenges.length} 题</Tag>}
              styles={{ header: { minHeight: 44, padding: '0 16px' }, body: { padding: 16 } }}
            >
              <ContestChallengePreviewTable
                challenges={publicChallenges}
                reviewMode={challengePreviewAccess.useReviewChallenges}
              />
            </Card>
          ) : null}
        </div>

        <div className="space-y-3 xl:sticky xl:top-5 xl:h-fit">
          <Card className="border-0 shadow-sm" styles={{ body: { padding: 14 } }}>
            <div className="rounded-2xl border border-slate-200 bg-slate-50 p-3.5">
              <div className="text-xs uppercase tracking-[0.18em] text-text-secondary">比赛状态</div>
              <div className="mt-1.5 text-xl font-semibold leading-tight text-slate-900">{statusSummary}</div>
              <p className="mt-1.5 text-sm leading-6 text-text-secondary">
                {contest.status === 'published'
                  ? '当前重点是完成报名和组队，开赛后会开放正式赛场。'
                  : contest.status === 'ongoing'
                    ? '现在可以直接进入赛场完成答题与环境操作。'
                    : contest.status === 'ended'
                      ? '比赛已结束，可以继续查看赛后题目回顾与结果。'
                      : '比赛仍在准备中，当前可先查看信息并等待开放。'}
              </p>
              <div className="mt-3 flex flex-wrap gap-2">
                {participation.canEnter ? (
                  <Button type="primary" onClick={handleEnterContest}>进入比赛</Button>
                ) : participation.canRegister ? (
                  <Button
                    type="primary"
                    onClick={() => (isTeamContest ? setTeamModalVisible(true) : setRegisterModalVisible(true))}
                  >
                    立即报名
                  </Button>
                ) : null}
              </div>
            </div>
          </Card>

          {myTeam ? <ContestTeamCard team={myTeam} /> : null}

          <Card className="border-0 shadow-sm" title="赛场提示" styles={{ header: { minHeight: 42, padding: '0 14px' }, body: { padding: 14 } }}>
            <div className="space-y-1.5 text-sm text-text-secondary">
              <div>比赛阶段：{statusSummary}</div>
              <div>组队模式：{isTeamContest ? '需要先创建或加入队伍' : '个人独立报名'}</div>
              <div>开始后可直接进入赛场进行解题</div>
            </div>
          </Card>

          <Card className="border-0 shadow-sm" title="赛场状态" styles={{ header: { minHeight: 42, padding: '0 14px' }, body: { padding: 14 } }}>
            <div className="flex items-center gap-3 rounded-2xl border border-slate-200 bg-slate-50 px-3.5 py-3">
              <TrophyOutlined className="text-2xl text-warning" />
              <div className="min-w-0 text-sm">
                {contest.status === 'draft' ? <p className="text-text-secondary">比赛仍处于草稿状态</p> : null}
                {contest.status === 'published' ? <p className="text-success">比赛正在报名中</p> : null}
                {contest.status === 'ongoing' ? <p className="text-primary">比赛进行中</p> : null}
                {contest.status === 'ended' ? <p className="text-text-secondary">比赛已结束，可查看赛后题目回顾</p> : null}
              </div>
            </div>
          </Card>
        </div>
      </div>

      {isManager ? (
        <Card
          title={`比赛题目管理 (${contestChallenges.length})`}
          className="border-0 shadow-sm"
          loading={loading}
          extra={(
            contest.status === 'draft' || contest.status === 'published'
              ? (
                <Button type="primary" icon={<PlusOutlined />} size="small" onClick={() => void handleOpenAddChallenge()}>
                  添加题目
                </Button>
              )
              : null
          )}
        >
          <ContestChallengeAdminTable
            challenges={contestChallenges}
            onRemove={(challengeId) => void handleRemoveChallenge(challengeId)}
          />
        </Card>
      ) : null}

      <Modal
        title="添加题目"
        open={challengeModalVisible}
        onCancel={() => setChallengeModalVisible(false)}
        onOk={() => void handleAddChallenge()}
        okText="添加"
      >
        <div className="py-4">
          <div className="mb-4">
            <label className="mb-2 block font-medium">选择题目</label>
            <Select
              className="w-full"
              placeholder="搜索并选择题目"
              showSearch
              optionFilterProp="label"
              value={selectedChallengeId}
              onChange={setSelectedChallengeId}
              options={availableChallenges.map((challenge) => ({
                value: challenge.id,
                label: `${challenge.title} (${getRuntimeProfileLabel(challenge.runtime_profile)})`,
              }))}
            />
          </div>
          <div>
            <label className="mb-2 block font-medium">分值</label>
            <InputNumber min={1} value={addPoints} onChange={(value) => setAddPoints(value || 100)} className="w-full" />
          </div>
        </div>
      </Modal>

      <ContestJeopardyRegistrationModal
        registerOpen={registerModalVisible}
        teamOpen={teamModalVisible}
        joinTeamCode={joinTeamCode}
        loading={actionLoading}
        onRegisterCancel={() => setRegisterModalVisible(false)}
        onRegisterConfirm={() => void handleRegister()}
        onTeamCancel={() => setTeamModalVisible(false)}
        onCreateTeam={(values) => void handleCreateTeam(values)}
        onJoinTeamCodeChange={setJoinTeamCode}
        onJoinTeam={() => void handleJoinTeam()}
      />
    </div>
  )
}
