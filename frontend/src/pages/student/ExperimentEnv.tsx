import { Modal, Spin, message } from 'antd'
import { useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'

import { ExperimentWorkbench } from '@/components/experiment'
import { probeRuntimeEndpoint } from '@/api/experimentRuntime'
import { submitExperiment } from '@/api/experimentSubmission'
import { RUNTIME_WORKBENCH_TOOL_ORDER } from '@/domains/runtime/workbench'
import type { ExperimentWorkbenchToolKey } from '@/domains/experiment/workbench'
import {
  getWorkbenchActiveTab,
  getWorkbenchIdeToolUrl,
  shouldProbeIdeRuntime,
} from '@/domains/experiment/workbench'
import { usePersistedTab } from '@/hooks'
import { useExperimentStore, useUserStore } from '@/store'

export default function StudentExperimentEnv() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const experimentId = Number.parseInt(id || '0', 10)
  const user = useUserStore((state) => state.user)

  const {
    currentExperiment,
    currentEnv,
    currentSession,
    currentInstances,
    currentTools,
    sessionMembers,
    sessionMessages,
    envStatus,
    remainingSeconds,
    loading,
    loadWorkbench,
    extendEnv,
    pauseEnv,
    postMessage,
    resumeEnv,
    createSnapshot,
    startEnv,
    stopEnv,
    updateSessionMember,
    stopStatusPolling,
    updateRemainingTime,
  } = useExperimentStore()

  const [submitModalVisible, setSubmitModalVisible] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [submitReport, setSubmitReport] = useState('')
  const [ideReady, setIdeReady] = useState(false)
  const [mountedTabs, setMountedTabs] = useState<ExperimentWorkbenchToolKey[]>([])
  const [activeTab, setActiveTab] = usePersistedTab(
    `experiment_env_${experimentId}`,
    'ide',
    RUNTIME_WORKBENCH_TOOL_ORDER,
  )
  const timerRef = useRef<NodeJS.Timeout | null>(null)

  useEffect(() => {
    const init = async () => {
      try {
        await loadWorkbench(experimentId)
      } catch {
        message.error('获取实验工作台失败')
        navigate(-1)
      }
    }

    void init()

    return () => {
      stopStatusPolling()
    }
  }, [experimentId, loadWorkbench, navigate, stopStatusPolling])

  useEffect(() => {
    if (timerRef.current) {
      clearInterval(timerRef.current)
    }

    if (envStatus === 'running' && remainingSeconds > 0) {
      timerRef.current = setInterval(() => {
        updateRemainingTime()
      }, 1000)
    }

    return () => {
      if (timerRef.current) {
        clearInterval(timerRef.current)
      }
    }
  }, [envStatus, remainingSeconds, updateRemainingTime])

  const availableTools = useMemo(() => currentTools, [currentTools])
  const currentTab = useMemo(
    () => getWorkbenchActiveTab(availableTools, activeTab as ExperimentWorkbenchToolKey | undefined),
    [activeTab, availableTools],
  )

  const currentMember = useMemo(
    () => sessionMembers.find((member) => member.user_id === user?.id) || null,
    [sessionMembers, user?.id],
  )

  useEffect(() => {
    if (!currentTab) {
      return
    }

    setMountedTabs((previous) => (
      previous.includes(currentTab) ? previous : [...previous, currentTab]
    ))
  }, [currentTab])

  const ideToolUrl = useMemo(() => getWorkbenchIdeToolUrl(availableTools), [availableTools])

  useEffect(() => {
    setIdeReady(false)

    if (!shouldProbeIdeRuntime(envStatus, ideToolUrl)) {
      return
    }

    const controller = new AbortController()
    let disposed = false

    const probeIde = async () => {
      for (let attempt = 0; attempt < 12 && !disposed; attempt += 1) {
        try {
          const ready = await probeRuntimeEndpoint(`${ideToolUrl}/`, controller.signal)
          if (ready) {
            if (!disposed) {
              setIdeReady(true)
            }
            return
          }
        } catch {
          // IDE 启动存在短暂延迟，继续轮询即可。
        }

        await new Promise((resolve) => setTimeout(resolve, 1000))
      }
    }

    void probeIde()

    return () => {
      disposed = true
      controller.abort()
    }
  }, [envStatus, ideToolUrl])

  const handleExtend = async () => {
    if (!currentEnv?.env_id) {
      return
    }

    await extendEnv(currentEnv.env_id)
    message.success('实验环境已延长')
  }

  const handlePause = async () => {
    if (!currentEnv?.env_id) {
      return
    }

    await pauseEnv(currentEnv.env_id)
    message.success('实验环境已暂停')
  }

  const handleResume = async () => {
    if (!currentEnv?.env_id) {
      return
    }

    await resumeEnv(currentEnv.env_id)
    message.success('实验环境已恢复')
  }

  const handleStop = async () => {
    if (!currentEnv?.env_id) {
      return
    }

    await stopEnv(currentEnv.env_id)
    message.success('实验环境已停止')
  }

  const handleCreateSnapshot = async () => {
    if (!currentEnv?.env_id) {
      return
    }

    const env = await createSnapshot(currentEnv.env_id)
    if (env?.snapshot_url) {
      message.success('快照创建成功')
      return
    }

    message.success('已提交快照创建请求')
  }

  const handleRestoreSnapshot = async () => {
    if (!currentEnv?.snapshot_url) {
      message.warning('当前还没有可恢复的快照')
      return
    }

    await startEnv(experimentId, currentEnv.snapshot_url)
    message.success('已从最近一次快照恢复实验环境')
  }

  const handleSendMessage = async (content: string) => {
    if (!currentSession?.session_key || !content.trim()) {
      return
    }

    await postMessage(currentSession.session_key, content.trim())
  }

  const handleUpdateSessionMember = async (
    userId: number,
    payload: { role_key?: string; assigned_node_key?: string; join_status?: 'joined' | 'left' },
  ) => {
    if (!currentSession?.session_key) {
      return
    }

    await updateSessionMember(currentSession.session_key, userId, payload)
    message.success('协作成员信息已更新')
  }

  const canManageSessionMembers = useMemo(() => {
    if (!currentSession || !user?.id) {
      return false
    }

    if (currentMember?.role_key === 'owner' || currentMember?.role_key === 'leader') {
      return true
    }

    return currentSession.members?.[0]?.user_id === user.id
  }, [currentMember?.role_key, currentSession, user?.id])

  const handleExit = () => {
    const envStillRunning = envStatus === 'running' || envStatus === 'creating' || envStatus === 'paused'

    Modal.confirm({
      title: '确认退出',
      content: envStillRunning
        ? '退出后实验环境会继续保留，你可以稍后回到工作台继续操作。'
        : '当前实验环境已经结束，退出后将返回实验详情页。',
      onOk: async () => {
        navigate(`/student/experiments/${experimentId}`)
      },
    })
  }

  const handleSubmit = async () => {
    setSubmitting(true)

    try {
      await submitExperiment(experimentId, {
        env_id: currentEnv?.env_id,
        content: submitReport,
      })

      message.success('实验提交成功')
      setSubmitModalVisible(false)

      if (currentEnv?.env_id) {
        await stopEnv(currentEnv.env_id)
      }

      navigate(`/student/experiments/${experimentId}`)
    } finally {
      setSubmitting(false)
    }
  }

  if (loading) {
    return (
      <div className="-m-6 flex h-[calc(100vh-64px)] items-center justify-center bg-gray-950">
        <Spin size="large" />
      </div>
    )
  }

  return (
    <ExperimentWorkbench
      experimentId={experimentId}
      experiment={currentExperiment}
      session={currentSession}
      instances={currentInstances || []}
      currentMember={currentMember}
      sessionMembers={sessionMembers}
      sessionMessages={sessionMessages}
      canManageSessionMembers={canManageSessionMembers}
      envStatus={envStatus}
      remainingSeconds={remainingSeconds}
      availableTools={availableTools}
      activeTab={currentTab}
      mountedTabs={mountedTabs}
      ideReady={ideReady}
      submitModalVisible={submitModalVisible}
      submitting={submitting}
      submitReport={submitReport}
      snapshotUrl={currentEnv?.snapshot_url}
      envErrorMessage={currentEnv?.error_message}
      onSetActiveTab={setActiveTab}
      onExtend={() => void handleExtend()}
      onPause={() => void handlePause()}
      onResume={() => void handleResume()}
      onCreateSnapshot={() => void handleCreateSnapshot()}
      onRestoreSnapshot={() => void handleRestoreSnapshot()}
      onStop={() => void handleStop()}
      onUpdateSessionMember={(userId, payload) => void handleUpdateSessionMember(userId, payload)}
      onExit={handleExit}
      onOpenSubmit={() => setSubmitModalVisible(true)}
      onCloseSubmit={() => setSubmitModalVisible(false)}
      onSubmitReportChange={setSubmitReport}
      onSubmitExperiment={() => void handleSubmit()}
      onSendMessage={(content) => void handleSendMessage(content)}
    />
  )
}
