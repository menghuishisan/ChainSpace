import { Alert, Button, Card, Empty, Space, Tag } from 'antd'
import type { ReactNode } from 'react'
import { useMemo } from 'react'

import { ApiDebugger, BlockExplorer, FileManager, KeepAliveTabPanel, LogViewer } from '@/components/experiment'
import {
  RUNTIME_WORKBENCH_TOOL_META,
  normalizeRuntimeWorkbenchToolKind,
} from '@/domains/runtime/workbench'
import type { Challenge, ChallengeEnv } from '@/types'
import type { ContestWorkbenchTabKey } from '@/types/presentation'
import { formatDuration } from '@/utils/format'

interface ContestWorkbenchTool {
  key: ContestWorkbenchTabKey
  label: string
  kind: string
  route: string
  target?: string
  icon: ReactNode
}

interface ContestRuntimeWorkbenchProps {
  challenge: Challenge
  challengeEnv: ChallengeEnv | null
  envRemaining: number
  ideReady: boolean
  activeTab?: ContestWorkbenchTabKey
  mountedTabs: ContestWorkbenchTabKey[]
  onSetActiveTab: (key: ContestWorkbenchTabKey) => void
}

const challengeEnvIconFallback: ReactNode = RUNTIME_WORKBENCH_TOOL_META.rpc.icon

function renderPlaceholder(title: string, description: string) {
  return (
    <div className="flex h-full items-center justify-center bg-slate-950 text-slate-100">
      <Empty
        image={Empty.PRESENTED_IMAGE_SIMPLE}
        description={(
          <div className="space-y-1">
            <div className="font-medium text-slate-100">{title}</div>
            <div className="text-xs text-slate-400">{description}</div>
          </div>
        )}
      />
    </div>
  )
}

function renderIframe(accessUrl: string, title: string) {
  return <iframe src={`${accessUrl}/`} className="h-full w-full border-0" title={title} />
}

function toolTitle(tool: ContestWorkbenchTool) {
  const normalizedKind = normalizeRuntimeWorkbenchToolKind(tool.kind)
  const base = (normalizedKind ? RUNTIME_WORKBENCH_TOOL_META[normalizedKind]?.title : undefined) || tool.kind || tool.label
  if (!tool.target || tool.target === 'workspace') {
    return base
  }
  return `${base} · ${tool.target}`
}

function toolBadge(tool: ContestWorkbenchTool) {
  if (!tool.target || tool.target === 'workspace') {
    return '主操作区'
  }
  if (tool.target === 'fork') {
    return '主网分叉链'
  }
  return `服务节点 (${tool.target})`
}

function apiDebuggerTitle(tool: ContestWorkbenchTool) {
  if (tool.kind === 'rpc' && tool.target === 'fork') {
    return '主网分叉链调试台'
  }
  if (tool.kind === 'rpc') {
    return tool.target && tool.target !== 'workspace' ? `${tool.label} 调试台` : '链上接口调试台'
  }
  return tool.target && tool.target !== 'workspace' ? `${tool.label} 调试台` : '接口调试台'
}

function apiDebuggerDescription(tool: ContestWorkbenchTool) {
  if (tool.kind === 'rpc' && tool.target === 'fork') {
    return '向当前题目的主网分叉链发送请求，用于复现真实链上状态。'
  }
  if (tool.kind === 'rpc') {
    return tool.target && tool.target !== 'workspace'
      ? `向服务节点 ${tool.target} 发送链上接口请求。`
      : '向当前题目环境发送链上接口请求。'
  }
  return tool.target && tool.target !== 'workspace'
    ? `向服务节点 ${tool.target} 发送接口请求。`
    : '向当前题目环境发送接口请求。'
}

function renderToolPanel(tool: ContestWorkbenchTool, ideReady: boolean) {
  switch (tool.kind) {
    case 'ide':
      return ideReady
        ? renderIframe(tool.route, toolTitle(tool))
        : renderPlaceholder('IDE 准备中', '工作区容器已经启动，正在等待 Web IDE 完成初始化。')
    case 'terminal':
      return renderIframe(tool.route, toolTitle(tool))
    case 'files':
      return <FileManager accessUrl={tool.route} />
    case 'explorer':
      return <BlockExplorer accessUrl={tool.route} />
    case 'logs':
      return <LogViewer accessUrl={tool.route} />
    case 'api_debug':
    case 'rpc':
      return (
        <ApiDebugger
          accessUrl={tool.route}
          title={apiDebuggerTitle(tool)}
          description={apiDebuggerDescription(tool)}
        />
      )
    case 'visualization':
      return renderIframe(tool.route, toolTitle(tool))
    default:
      return renderIframe(tool.route, toolTitle(tool))
  }
}

export default function ContestRuntimeWorkbench({
  challenge,
  challengeEnv,
  envRemaining,
  ideReady,
  activeTab,
  mountedTabs,
  onSetActiveTab,
}: ContestRuntimeWorkbenchProps) {
  const availableTools = useMemo<ContestWorkbenchTool[]>(() => {
    if (!challengeEnv) {
      return []
    }

    return (challengeEnv.tools || [])
      .filter((tool) => Boolean(tool.route && tool.key))
      .map((tool) => {
        const kind = tool.kind || tool.key
        const normalizedKind = normalizeRuntimeWorkbenchToolKind(kind)
        const meta = normalizedKind
          ? RUNTIME_WORKBENCH_TOOL_META[normalizedKind]
          : { title: tool.label || kind, icon: challengeEnvIconFallback }
        const normalizedLabel = (tool.label || '').trim().toLowerCase()
        const label = !normalizedLabel || normalizedLabel === kind
          ? meta.title
          : tool.label
        return {
          key: tool.key,
          label,
          kind,
          route: tool.route,
          target: tool.target,
          icon: meta.icon,
        }
      })
  }, [challengeEnv])

  const resolvedTools = useMemo(
    () => new Map(availableTools.map((tool) => [tool.key, tool])),
    [availableTools],
  )

  if (!challenge.challenge_orchestration?.needs_environment) {
    return null
  }

  if (!challengeEnv) {
    return (
      <Card className="border-0 shadow-sm">
        <Alert
          type="info"
          showIcon
          message="比赛工作台尚未就绪"
          description="请先启动题目环境，准备完成后即可开始操作。"
        />
      </Card>
    )
  }

  if (challengeEnv.status === 'creating') {
    return (
      <Card className="border-0 shadow-sm">
        <Alert
          type="info"
          showIcon
          message="题目环境创建中"
          description="环境正在准备中，页面会自动刷新状态。"
        />
      </Card>
    )
  }

  if (challengeEnv.status === 'failed') {
    return (
      <Card className="border-0 shadow-sm">
        <Alert
          type="error"
          showIcon
          message="题目环境启动失败"
          description={challengeEnv.error_message || '当前没有返回更详细的失败原因，请稍后重试。'}
        />
      </Card>
    )
  }

  if (challengeEnv.status !== 'running') {
    return (
      <Card className="border-0 shadow-sm">
        <Alert
          type="warning"
          showIcon
          message="当前环境尚未可用"
          description="环境当前不处于运行状态，请重新启动或刷新状态。"
        />
      </Card>
    )
  }

  if (!activeTab || availableTools.length === 0) {
    return (
      <Card className="border-0 shadow-sm">
        <Alert
          type="warning"
          showIcon
          message="当前环境没有可用工具入口"
          description="当前环境暂时没有可用入口，请刷新或重新启动环境。"
        />
      </Card>
    )
  }

  return (
    <Card
      className="border-0 shadow-sm"
      styles={{ body: { padding: 0 } }}
      title={(
        <div className="flex items-center justify-between gap-3">
          <div className="flex items-center gap-2">
            <span className="font-semibold">比赛工作台</span>
            <Tag color="processing">{challenge.title}</Tag>
          </div>
          <div className="text-xs text-text-secondary">环境剩余 {formatDuration(envRemaining)}</div>
        </div>
      )}
    >
      <div className="overflow-hidden rounded-b-2xl">
        <div className="border-b border-slate-200 bg-[linear-gradient(180deg,#f8fbff_0%,#f4f7fb_100%)] px-4 py-3">
          <div className="mb-2 text-xs uppercase tracking-[0.18em] text-text-secondary">可用工具</div>
          <Space wrap size="small">
            {availableTools.map((tool) => (
              <Button
                key={tool.key}
                type={activeTab === tool.key ? 'primary' : 'text'}
                className="h-auto rounded-xl px-3 py-2"
                icon={tool.icon}
                onClick={() => onSetActiveTab(tool.key)}
                title={toolTitle(tool)}
              >
                <span>{tool.label}</span>
                <span className="ml-2 text-[11px] opacity-70">{toolBadge(tool)}</span>
              </Button>
            ))}
          </Space>
        </div>

        <div className="min-w-0 bg-slate-950">
          <div className="h-[55vh] min-h-[420px] max-h-[780px]">
            {mountedTabs.map((tabKey) => {
              const tool = resolvedTools.get(tabKey)
              if (!tool) {
                return null
              }

              return (
                <KeepAliveTabPanel key={tabKey} active={activeTab === tabKey}>
                  {renderToolPanel(tool, ideReady)}
                </KeepAliveTabPanel>
              )
            })}
          </div>
        </div>

        {(challengeEnv.service_entries || []).length > 0 ? (
          <div className="border-t border-slate-200 bg-white px-4 py-4">
            <div className="mb-3 text-sm font-medium text-slate-900">附加服务</div>
            <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
              {challengeEnv.service_entries?.map((service) => (
                <div key={service.key} className="rounded-xl border border-slate-200 bg-slate-50 p-3">
                  <div className="font-medium text-slate-900">{service.label || service.key}</div>
                  <div className="mt-1 text-xs text-text-secondary">
                    {service.description || service.purpose || '暂无说明'}
                  </div>
                </div>
              ))}
            </div>
          </div>
        ) : null}
      </div>
    </Card>
  )
}
