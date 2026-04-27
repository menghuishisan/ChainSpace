import { ClockCircleOutlined, FireOutlined, TrophyOutlined } from '@ant-design/icons'
import { Card, Space, Spin, Tag, Typography, message } from 'antd'
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'

import { probeRuntimeEndpoint } from '@/api/experimentRuntime'
import {
  getChallengeAttachmentAccessUrl,
  getChallengeEnvStatus,
  getContest,
  getPlayingChallenges,
  getScoreboard,
  startChallengeEnv,
  stopChallengeEnv,
  submitFlag,
} from '@/api/contest'
import { PageHeader } from '@/components/common'
import EmptyState from '@/components/common/EmptyState'
import {
  CodeViewerModal,
  ContestRuntimeWorkbench,
  JeopardyChallengeGrid,
  JeopardySidebar,
} from '@/components/contest'
import {
  buildChallengeCategoryGroups,
  extractContractName,
  getSolvedChallengeCount,
  normalizeChallengeEnvState,
  renderChallengeMarkdown,
} from '@/domains/contest/jeopardy'
import { RUNTIME_WORKBENCH_TOOL_ORDER, normalizeRuntimeWorkbenchToolKind } from '@/domains/runtime/workbench'
import type { Challenge, ChallengeEnv } from '@/types'
import type { ContestWorkbenchTabKey } from '@/types/presentation'
import { CategoryMap, ChallengeRuntimeProfileMap, DifficultyMap } from '@/types'
import { formatDuration } from '@/utils/format'

const { Paragraph } = Typography

function downloadTextFile(filename: string, content: string) {
  const blob = new Blob([content], { type: 'text/plain;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
}

function getInitialWorkbenchTab(env: ChallengeEnv | null): ContestWorkbenchTabKey | undefined {
  if (!env) {
    return undefined
  }

  const tools = env.tools || []

  for (const kind of RUNTIME_WORKBENCH_TOOL_ORDER) {
    const matched = tools.find((tool) => normalizeRuntimeWorkbenchToolKind(tool.kind || tool.key) === kind)
    if (matched) {
      return matched.key
    }
  }

  return tools[0]?.key
}

function getIdeToolUrl(env: ChallengeEnv | null): string | undefined {
  if (!env) {
    return undefined
  }

  const ideRoute = (env.tools || []).find((tool) => normalizeRuntimeWorkbenchToolKind(tool.kind || tool.key) === 'ide')
  return ideRoute?.route
}

export default function Jeopardy() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const contestId = Number(id || '0')

  const [loading, setLoading] = useState(true)
  const [contest, setContest] = useState<Awaited<ReturnType<typeof getContest>> | null>(null)
  const [challenges, setChallenges] = useState<Challenge[]>([])
  const [myScore, setMyScore] = useState(0)
  const [selectedChallenge, setSelectedChallenge] = useState<Challenge | null>(null)
  const [flagInput, setFlagInput] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [remainingTime, setRemainingTime] = useState(0)

  const [challengeEnv, setChallengeEnv] = useState<ChallengeEnv | null>(null)
  const [envLoading, setEnvLoading] = useState(false)
  const [envRemaining, setEnvRemaining] = useState(0)
  const [fetchingEnv, setFetchingEnv] = useState(false)

  const [codeViewerOpen, setCodeViewerOpen] = useState(false)
  const [codeViewerTitle, setCodeViewerTitle] = useState('')
  const [codeViewerCode, setCodeViewerCode] = useState('')

  const [activeWorkbenchTab, setActiveWorkbenchTab] = useState<ContestWorkbenchTabKey | undefined>()
  const [mountedWorkbenchTabs, setMountedWorkbenchTabs] = useState<ContestWorkbenchTabKey[]>([])
  const [ideReady, setIdeReady] = useState(false)
  const selectedChallengeIdRef = useRef<number | null>(null)
  const envRequestTokenRef = useRef(0)

  useEffect(() => {
    selectedChallengeIdRef.current = selectedChallenge?.id ?? null
  }, [selectedChallenge?.id])

  useEffect(() => {
    const init = async () => {
      setLoading(true)
      try {
        const [contestData, challengeList, scoreboardData] = await Promise.all([
          getContest(contestId),
          getPlayingChallenges(contestId),
          getScoreboard(contestId),
        ])

        const normalizedChallenges = (challengeList || []).map((item) => (
          item.is_solved && item.awarded_points
            ? { ...item, points: item.awarded_points }
            : item
        ))

        setContest(contestData)
        setChallenges(normalizedChallenges)
        setMyScore(scoreboardData.my_score || 0)
        setSelectedChallenge((current) => current || normalizedChallenges[0] || null)

        const endAt = new Date(contestData.end_time).getTime()
        setRemainingTime(Math.max(0, Math.floor((endAt - Date.now()) / 1000)))
      } catch {
        navigate(-1)
      } finally {
        setLoading(false)
      }
    }

    void init()
  }, [contestId, navigate])

  useEffect(() => {
    if (remainingTime <= 0) {
      return
    }

    const timer = window.setInterval(() => {
      setRemainingTime((current) => Math.max(0, current - 1))
    }, 1000)

    return () => window.clearInterval(timer)
  }, [remainingTime])

  const loadEnvStatus = useCallback(async (challengeId: number) => {
    const requestToken = envRequestTokenRef.current + 1
    envRequestTokenRef.current = requestToken
    setFetchingEnv(true)
    try {
      const env = normalizeChallengeEnvState(await getChallengeEnvStatus(contestId, challengeId))
      if (selectedChallengeIdRef.current !== challengeId || envRequestTokenRef.current != requestToken) {
        return
      }
      if (env) {
        setChallengeEnv(env)
        setEnvRemaining(env.remaining)
        return
      }

      setChallengeEnv(null)
      setEnvRemaining(0)
    } catch {
      setChallengeEnv(null)
      setEnvRemaining(0)
    } finally {
      if (selectedChallengeIdRef.current === challengeId && envRequestTokenRef.current === requestToken) {
        setFetchingEnv(false)
      }
    }
  }, [contestId])

  useEffect(() => {
    setActiveWorkbenchTab(undefined)
    setMountedWorkbenchTabs([])
    setIdeReady(false)

    if (!selectedChallenge) {
      setChallengeEnv(null)
      setEnvRemaining(0)
      return
    }

    if (!selectedChallenge.challenge_orchestration?.needs_environment) {
      setChallengeEnv(null)
      setEnvRemaining(0)
      return
    }

    void loadEnvStatus(selectedChallenge.id)
  }, [loadEnvStatus, selectedChallenge])

  useEffect(() => {
    if (!challengeEnv || challengeEnv.status !== 'running') {
      return
    }

    setEnvRemaining(challengeEnv.remaining)
  }, [challengeEnv])

  useEffect(() => {
    if (!challengeEnv || challengeEnv.status !== 'running' || envRemaining <= 0) {
      return
    }

    const timer = window.setInterval(() => {
      setEnvRemaining((current) => {
        if (current <= 1) {
          setChallengeEnv((env) => (env ? { ...env, status: 'expired', remaining: 0 } : null))
          return 0
        }
        return current - 1
      })
    }, 1000)

    return () => window.clearInterval(timer)
  }, [challengeEnv, envRemaining])

  useEffect(() => {
    if (!selectedChallenge || !challengeEnv || challengeEnv.challenge_id !== selectedChallenge.id || challengeEnv.status !== 'creating') {
      return
    }

    const timer = window.setInterval(() => {
      void loadEnvStatus(selectedChallenge.id)
    }, 3000)

    return () => window.clearInterval(timer)
  }, [challengeEnv, loadEnvStatus, selectedChallenge])

  useEffect(() => {
    const nextTab = getInitialWorkbenchTab(challengeEnv)
    setActiveWorkbenchTab(nextTab)
    setMountedWorkbenchTabs(nextTab ? [nextTab] : [])
  }, [challengeEnv?.env_id, challengeEnv?.status])

  useEffect(() => {
    if (!activeWorkbenchTab) {
      return
    }

    setMountedWorkbenchTabs((current) => (
      current.includes(activeWorkbenchTab) ? current : [...current, activeWorkbenchTab]
    ))
  }, [activeWorkbenchTab])

  const ideToolUrl = useMemo(() => getIdeToolUrl(challengeEnv), [challengeEnv])

  useEffect(() => {
    setIdeReady(false)

    if (!challengeEnv || challengeEnv.status !== 'running' || !ideToolUrl) {
      return
    }

    const controller = new AbortController()
    let disposed = false

    const probeIde = async () => {
      for (let attempt = 0; attempt < 12 && !disposed; attempt += 1) {
        const ready = await probeRuntimeEndpoint(`${ideToolUrl}/`, controller.signal)
        if (ready) {
          if (!disposed) {
            setIdeReady(true)
          }
          return
        }
        await new Promise((resolve) => setTimeout(resolve, 1000))
      }
    }

    void probeIde()

    return () => {
      disposed = true
      controller.abort()
    }
  }, [challengeEnv, ideToolUrl])

  const handleOpenCodeViewer = useCallback((code: string, label: string) => {
    setCodeViewerCode(code)
    setCodeViewerTitle(label)
    setCodeViewerOpen(true)
  }, [])

  const resetChallengeEnvState = useCallback(() => {
    setChallengeEnv(null)
    setEnvRemaining(0)
    setActiveWorkbenchTab(undefined)
    setMountedWorkbenchTabs([])
    setIdeReady(false)
  }, [])

  const handleStartEnv = async () => {
    if (!selectedChallenge) {
      return
    }

    setEnvLoading(true)
    try {
      const env = normalizeChallengeEnvState(await startChallengeEnv(contestId, selectedChallenge.id))
      if (selectedChallengeIdRef.current !== selectedChallenge.id) {
        return
      }
      setChallengeEnv(env)
      setEnvRemaining(env?.remaining ?? 0)
      message.success(env?.status === 'running' ? '题目环境启动成功' : '题目环境已开始创建')
    } catch (error: unknown) {
      message.error(error instanceof Error ? error.message : '题目环境启动失败')
    } finally {
      setEnvLoading(false)
    }
  }

  const handleStopEnv = async () => {
    if (!selectedChallenge || !challengeEnv) {
      return
    }

    setEnvLoading(true)
    try {
      await stopChallengeEnv(contestId, selectedChallenge.id)
      resetChallengeEnvState()
      message.success('题目环境已停止')
    } catch {
      message.error('停止环境失败')
    } finally {
      setEnvLoading(false)
    }
  }

  const handleSubmitFlag = async () => {
    if (!selectedChallenge || !flagInput.trim()) {
      message.error('请输入 Flag')
      return
    }

    setSubmitting(true)
    try {
      const solvedChallenge = selectedChallenge
      const result = await submitFlag(contestId, selectedChallenge.id, flagInput.trim())
      if (!result.correct) {
        message.error(result.message || 'Flag 错误')
        return
      }

      message.success(`提交正确，获得 ${result.points} 分`)
      setMyScore((current) => current + result.points)
      setChallenges((current) => current.map((item) => (
        item.id === selectedChallenge.id
          ? { ...item, is_solved: true, awarded_points: result.points, points: result.points }
          : item
      )))
      setSelectedChallenge((current) => (
        current
          ? { ...current, is_solved: true, awarded_points: result.points, points: result.points }
          : null
      ))
      setFlagInput('')
      if (solvedChallenge.challenge_orchestration?.needs_environment && challengeEnv) {
        try {
          await stopChallengeEnv(contestId, solvedChallenge.id)
          resetChallengeEnvState()
        } catch {
          message.warning('题目已解出，但自动停止环境失败，请手动停止')
        }
      }
      message.destroy()
      message.success(
        solvedChallenge.challenge_orchestration?.needs_environment
          ? `提交正确，获得 ${result.points} 分，已自动关闭题目环境`
          : `提交正确，获得 ${result.points} 分`,
      )
    } catch (error: unknown) {
      message.error(error instanceof Error ? error.message : '提交失败')
    } finally {
      setSubmitting(false)
    }
  }

  const categoryGroups = useMemo(
    () => buildChallengeCategoryGroups(challenges, CategoryMap),
    [challenges],
  )

  const solvedCount = useMemo(() => getSolvedChallengeCount(challenges), [challenges])
  const totalPoints = myScore
  const renderedDescription = selectedChallenge
    ? renderChallengeMarkdown(selectedChallenge.description || '暂无题目说明')
    : ''

  if (loading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <Spin size="large" />
      </div>
    )
  }

  if (!contest) {
    return null
  }

  return (
    <div className="space-y-6 pb-6">
      <PageHeader
        title={contest.title}
        subtitle="CTF 解题赛工作台"
        showBack
        tags={<Tag color="green">进行中</Tag>}
        extra={(
          <Space size="large">
            <span className="flex items-center">
              <ClockCircleOutlined className="mr-1" />
              <span className="font-mono">{formatDuration(remainingTime)}</span>
            </span>
            <span>当前得分 <span className="font-bold text-primary">{totalPoints}</span></span>
            <span>已解题数 <span className="font-bold text-primary">{solvedCount}/{challenges.length}</span></span>
          </Space>
        )}
      />

      <Card className="border-0 shadow-sm" styles={{ body: { padding: 0 } }}>
        <div className="grid gap-px overflow-hidden rounded-xl bg-slate-200 md:grid-cols-4">
          <div className="bg-slate-900 px-5 py-4 text-white">
            <div className="text-xs uppercase tracking-[0.2em] text-slate-300">当前阶段</div>
            <div className="mt-2 text-lg font-semibold">比赛进行中</div>
            <div className="mt-1 text-xs text-slate-300">已报名队伍可进入正式赛场。</div>
          </div>
          <div className="bg-white px-5 py-4">
            <div className="flex items-center gap-2 text-xs uppercase tracking-[0.18em] text-text-secondary">
              <TrophyOutlined className="text-primary" />
              <span>进度</span>
            </div>
            <div className="mt-2 text-lg font-semibold text-slate-900">{solvedCount}/{challenges.length}</div>
            <div className="mt-1 text-xs text-text-secondary">累计得分 {totalPoints}</div>
          </div>
          <div className="bg-white px-5 py-4">
            <div className="flex items-center gap-2 text-xs uppercase tracking-[0.18em] text-text-secondary">
              <ClockCircleOutlined className="text-sky-500" />
              <span>剩余时间</span>
            </div>
            <div className="mt-2 font-mono text-lg font-semibold text-slate-900">{formatDuration(remainingTime)}</div>
            <div className="mt-1 text-xs text-text-secondary">比赛剩余时间</div>
          </div>
          <div className="bg-white px-5 py-4">
            <div className="flex items-center gap-2 text-xs uppercase tracking-[0.18em] text-text-secondary">
              <FireOutlined className="text-orange-500" />
              <span>当前题目</span>
            </div>
            <div className="mt-2 line-clamp-1 text-lg font-semibold text-slate-900">
              {selectedChallenge?.title || '请选择题目'}
            </div>
            <div className="mt-1 text-xs text-text-secondary">
              {selectedChallenge
                ? `${ChallengeRuntimeProfileMap[selectedChallenge.runtime_profile]} · ${selectedChallenge.points ?? selectedChallenge.base_points} 分`
                : '从下方题库选择一道题进入'}
            </div>
          </div>
        </div>
      </Card>

      <div className="grid gap-6 xl:grid-cols-[minmax(0,1fr)_380px]">
        <div className="space-y-5">
          <Card
            className="border-0 shadow-sm"
            title="题库导航"
            extra={<span className="text-xs text-text-secondary">快速切题，不打断主工作区</span>}
            styles={{ body: { maxHeight: 260, overflow: 'auto' } }}
          >
            {categoryGroups.length === 0 ? (
              <EmptyState description="当前比赛暂无题目" />
            ) : (
              <JeopardyChallengeGrid
                compact
                categoryGroups={categoryGroups}
                selectedChallengeId={selectedChallenge?.id}
                runtimeLabels={ChallengeRuntimeProfileMap}
                difficultyMap={DifficultyMap}
                onSelectChallenge={setSelectedChallenge}
              />
            )}
          </Card>

          <Card className="border-0 shadow-sm">
            <div className="flex flex-wrap items-start justify-between gap-4">
              <div className="min-w-0">
                <div className="text-xs uppercase tracking-[0.18em] text-text-secondary">当前题目</div>
                <div className="mt-2 text-2xl font-semibold text-slate-900">{selectedChallenge?.title || '请选择题目'}</div>
                <div className="mt-3 flex flex-wrap gap-2">
                  {selectedChallenge ? <Tag color="blue">{ChallengeRuntimeProfileMap[selectedChallenge.runtime_profile]}</Tag> : null}
                  {selectedChallenge ? <Tag color={DifficultyMap[selectedChallenge.difficulty]?.color}>{DifficultyMap[selectedChallenge.difficulty]?.text}</Tag> : null}
                  {selectedChallenge?.is_solved ? <Tag color="success">已完成</Tag> : null}
                </div>
              </div>
              {selectedChallenge ? (
                <div className="rounded-2xl bg-slate-900 px-4 py-3 text-right text-white">
                  <div className="text-xs uppercase tracking-[0.18em] text-slate-300">分值</div>
                  <div className="mt-1 text-2xl font-semibold">{selectedChallenge.points ?? selectedChallenge.base_points}</div>
                </div>
              ) : null}
            </div>
            <Paragraph className="mt-4 mb-0 text-sm text-text-secondary">
              {selectedChallenge?.challenge_orchestration?.needs_environment
                ? '工作区和链环境会在下方主区域展开，右侧只保留运行说明、环境控制和提交。'
                : '当前题目为静态分析题，可直接阅读题面、查看代码并提交答案。'}
            </Paragraph>
          </Card>

          {selectedChallenge?.challenge_orchestration?.needs_environment ? (
            <ContestRuntimeWorkbench
              challenge={selectedChallenge}
              challengeEnv={challengeEnv}
              envRemaining={envRemaining}
              ideReady={ideReady}
              activeTab={activeWorkbenchTab}
              mountedTabs={mountedWorkbenchTabs}
              onSetActiveTab={setActiveWorkbenchTab}
            />
          ) : null}

          <Card
            className="border-0 shadow-sm"
            title="题目说明"
            extra={<span className="text-xs text-text-secondary">只展示解题需要的信息，不暴露运行时内部实现细节</span>}
            styles={{ body: { maxHeight: 520, overflow: 'auto' } }}
          >
            {selectedChallenge ? (
              <div className="prose prose-sm max-w-none" dangerouslySetInnerHTML={{ __html: renderedDescription }} />
            ) : (
              <EmptyState description="请选择一道题目查看题面" />
            )}
          </Card>
        </div>

        <div className="xl:sticky xl:top-6 xl:h-fit">
          <JeopardySidebar
            selectedChallenge={selectedChallenge}
            categoryMap={CategoryMap}
            runtimeLabels={ChallengeRuntimeProfileMap}
            difficultyMap={DifficultyMap}
            renderedDescription={renderedDescription}
            showDescriptionCard={false}
            firstBloodBonus={contest.first_blood_bonus}
            challengeEnv={challengeEnv}
            envLoading={envLoading}
            envRemaining={envRemaining}
            fetchingEnv={fetchingEnv}
            flagInput={flagInput}
            submitting={submitting}
            onRefreshEnv={() => {
              if (selectedChallenge) {
                void loadEnvStatus(selectedChallenge.id)
              }
            }}
            onStartEnv={() => void handleStartEnv()}
            onStopEnv={() => void handleStopEnv()}
            onOpenCodeViewer={handleOpenCodeViewer}
            onCopyCode={(code) => {
              navigator.clipboard.writeText(code)
                .then(() => message.success('已复制代码'))
                .catch(() => message.error('复制代码失败'))
            }}
            onDownloadCode={(filename, code) => downloadTextFile(filename, code)}
            onOpenAttachment={(attachmentIndex) => {
              if (!selectedChallenge) {
                return
              }
              void getChallengeAttachmentAccessUrl(contestId, selectedChallenge.id, attachmentIndex)
                .then((url) => {
                  const link = document.createElement('a')
                  link.href = url
                  link.rel = 'noopener noreferrer'
                  link.download = ''
                  document.body.appendChild(link)
                  link.click()
                  document.body.removeChild(link)
                })
                .catch((error: unknown) => {
                  message.error(error instanceof Error ? error.message : '附件打开失败')
                })
            }}
            onFlagInputChange={setFlagInput}
            onSubmitFlag={() => void handleSubmitFlag()}
          />
        </div>
      </div>

      <CodeViewerModal
        open={codeViewerOpen}
        code={codeViewerCode}
        title={codeViewerTitle || `${extractContractName(codeViewerCode)}.sol`}
        onClose={() => setCodeViewerOpen(false)}
      />
    </div>
  )
}
