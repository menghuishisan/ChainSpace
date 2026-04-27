import { useCallback, useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { Button, Spin, message } from 'antd'
import type { UploadProps } from 'antd'
import { DesktopOutlined } from '@ant-design/icons'

import { PageHeader } from '@/components/common'
import { AgentBattleSidebar, AgentBattleSummary } from '@/components/contest'
import {
  deployContract,
  createTeamWorkspace,
  getTeamWorkspace,
  getTeamContract,
  upgradeContract,
  uploadAgentCode,
} from '@/api/contest'
import {
  getBattleConfig,
  getBattleScoreTopList,
  getContestRemainingSeconds,
  STRATEGY_TEMPLATE,
} from '@/domains/contest/battle'
import { useContestStore } from '@/store'
import type { TeamContractInfo } from '@/types'

export default function AgentBattle() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const contestId = Number(id || '0')

  const {
    currentContest,
    battleStatus,
    currentRound,
    myTeam,
    scoreboard,
    loading,
    hydrateBattleWorkspace,
    fetchBattleStatus,
    fetchCurrentRound,
    startPolling,
    stopPolling,
    reset,
  } = useContestStore()

  const [uploading, setUploading] = useState(false)
  const [deploying, setDeploying] = useState(false)
  const [remainingTime, setRemainingTime] = useState(0)
  const [sourceCode, setSourceCode] = useState('')
  const [contractInfo, setContractInfo] = useState<TeamContractInfo | null>(null)
  const [workspaceLoading, setWorkspaceLoading] = useState(false)

  const fetchContractInfo = useCallback(async (teamId: number) => {
    try {
      const info = await getTeamContract(contestId, teamId)
      setContractInfo(info)
    } catch {
      setContractInfo(null)
    }
  }, [contestId])

  useEffect(() => {
    const init = async () => {
      try {
        await hydrateBattleWorkspace(contestId)
      } catch {
        navigate(-1)
      }
    }

    void init()
    setSourceCode(STRATEGY_TEMPLATE)

    return () => {
      stopPolling()
      reset()
    }
  }, [contestId, hydrateBattleWorkspace, navigate, reset, stopPolling])

  useEffect(() => {
    if (myTeam?.id) {
      void fetchContractInfo(myTeam.id)
    } else {
      setContractInfo(null)
    }
  }, [fetchContractInfo, myTeam?.id])

  useEffect(() => {
    if (!currentContest) {
      return
    }

    setRemainingTime(getContestRemainingSeconds(currentContest.end_time))
    startPolling(contestId, 'battle')
  }, [contestId, currentContest, startPolling])

  useEffect(() => {
    if (remainingTime <= 0) {
      return
    }

    const timer = window.setInterval(() => {
      setRemainingTime((current) => Math.max(0, current - 1))
    }, 1000)

    return () => window.clearInterval(timer)
  }, [remainingTime])

  const handleDeploy = async () => {
    if (!sourceCode.trim()) {
      message.error('请输入策略智能体源码')
      return
    }

    setDeploying(true)
    try {
      const result = contractInfo
        ? await upgradeContract({ contest_id: contestId, new_implementation: sourceCode })
        : await deployContract({ contest_id: contestId, source_code: sourceCode })

      setContractInfo({
        id: result.id,
        contract_address: result.contract_address,
        status: result.status,
        version: result.version ?? contractInfo?.version ?? 1,
        deployed_at: typeof result.deployed_at === 'string' ? result.deployed_at : undefined,
      })

      message.success(contractInfo ? '策略新版本已提交' : '策略智能体部署请求已提交')
      await Promise.all([
        fetchBattleStatus(contestId),
        fetchCurrentRound(contestId),
      ])
    } finally {
      setDeploying(false)
    }
  }

  const handleUpload: UploadProps['customRequest'] = async (options) => {
    const { file, onSuccess, onError } = options
    setUploading(true)

    try {
      if (!(file instanceof File)) {
        throw new Error('上传文件无效')
      }

      await uploadAgentCode(contestId, file)
      message.success('策略代码上传成功')
      onSuccess?.(null)
      await fetchBattleStatus(contestId)
    } catch (error) {
      onError?.(error instanceof Error ? error : new Error('上传失败'))
    } finally {
      setUploading(false)
    }
  }

  const handleOpenWorkspace = async () => {
    setWorkspaceLoading(true)
    try {
      let workspace = await getTeamWorkspace(contestId)
      if (!workspace.access_url || workspace.status !== 'running') {
        workspace = await createTeamWorkspace(contestId)
      }
      const ideToolRoute = (workspace.tools || []).find((tool) => (tool.kind || tool.key) === 'ide')?.route
      const openURL = ideToolRoute || workspace.access_url
      if (!openURL) {
        message.error('队伍工作区尚未就绪，请稍后再试')
        return
      }
      window.open(openURL, '_blank', 'noopener,noreferrer')
    } finally {
      setWorkspaceLoading(false)
    }
  }

  if (loading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <Spin size="large" />
      </div>
    )
  }

  if (!currentContest) {
    return null
  }

  const battleConfig = getBattleConfig(currentContest)
  const recentEvents = battleStatus?.recent_events || []

  return (
    <div className="space-y-6">
      <PageHeader
        title={currentContest.title}
        subtitle="智能体博弈战"
        showBack
        extra={(
          <Button icon={<DesktopOutlined />} loading={workspaceLoading} onClick={() => void handleOpenWorkspace()}>
            队伍工作区
          </Button>
        )}
      />

      <div className="grid gap-6 xl:grid-cols-[minmax(0,1fr)_360px]">
        <AgentBattleSummary
          contest={currentContest}
          battleStatus={battleStatus}
          battleConfig={battleConfig}
          currentRound={currentRound}
          contractInfo={contractInfo}
          sourceCode={sourceCode}
          deploying={deploying}
          uploading={uploading}
          remainingTime={remainingTime}
          onSourceCodeChange={setSourceCode}
          onDeploy={() => void handleDeploy()}
          onUploadFile={handleUpload}
          onSpectate={() => navigate(`/contest/${contestId}/spectate`)}
        />

        <div className="xl:sticky xl:top-6 xl:h-fit">
          <AgentBattleSidebar
            scoreWeights={battleConfig.judge.score_weights || {}}
            scoreboard={getBattleScoreTopList(scoreboard)}
            recentEvents={recentEvents}
          />
        </div>
      </div>
    </div>
  )
}
