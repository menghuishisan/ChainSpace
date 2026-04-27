import {
  ApiOutlined,
  ClockCircleOutlined,
  InfoCircleOutlined,
  MessageOutlined,
  PauseCircleOutlined,
  PlayCircleOutlined,
  SaveOutlined,
  TeamOutlined,
} from '@ant-design/icons'
import { Button, Drawer, Input, Modal, Select, Space, Spin, Tag } from 'antd'
import { useEffect, useMemo, useState } from 'react'

import ApiDebugger from '../ApiDebugger'
import BlockExplorer from '../BlockExplorer'
import ExperimentVisualizationPanel from '../ExperimentVisualizationPanel'
import FileManager from '../FileManager'
import KeepAliveTabPanel from '../KeepAliveTabPanel'
import LogViewer from '../LogViewer'
import {
  getWorkbenchVisualizationModuleKey,
  isWorkbenchEnvRunning,
} from '@/domains/experiment/workbench'
import { EnvStatusMap } from '@/types'
import type { ExperimentWorkbenchProps, ExperimentWorkbenchTabKey } from '@/types/presentation'
import { formatDuration } from '@/utils/format'

function renderEnvPlaceholder(title: string, tips: string) {
  return (
    <div className="flex h-full min-h-[280px] flex-col items-center justify-center gap-3 bg-slate-50 text-slate-900">
      <div className="flex h-12 w-12 items-center justify-center rounded-full bg-sky-100">
        <ApiOutlined className="text-2xl text-primary" />
      </div>
      <div className="text-lg font-semibold">{title}</div>
      <div className="text-sm text-text-secondary">{tips}</div>
      <div className="text-xs text-text-secondary">你可以稍后重试，或切换到其他可用工具继续实验。</div>
    </div>
  )
}

function instanceKindText(kind: string): string {
  switch (kind) {
    case 'workspace':
      return '主操作区'
    case 'node':
      return '节点实例'
    case 'service':
      return '服务实例'
    default:
      return '运行实例'
  }
}

function instanceStatusText(status: string): string {
  const key = status as keyof typeof EnvStatusMap
  return EnvStatusMap[key]?.text || '状态未知'
}

export default function ExperimentWorkbench({
  experimentId,
  experiment,
  instances,
  currentMember,
  sessionMembers,
  sessionMessages,
  canManageSessionMembers,
  envStatus,
  remainingSeconds,
  availableTools,
  activeTab,
  mountedTabs,
  ideReady,
  submitModalVisible,
  submitting,
  submitReport,
  snapshotUrl,
  envErrorMessage,
  onSetActiveTab,
  onExtend,
  onPause,
  onResume,
  onCreateSnapshot,
  onRestoreSnapshot,
  onStop,
  onUpdateSessionMember,
  onExit,
  onOpenSubmit,
  onCloseSubmit,
  onSubmitReportChange,
  onSubmitExperiment,
  onSendMessage,
}: ExperimentWorkbenchProps) {
  const [selectedInstanceKey, setSelectedInstanceKey] = useState<string>()
  const [chatMessage, setChatMessage] = useState('')
  const [sidebarOpen, setSidebarOpen] = useState(false)
  const isTimeLow = remainingSeconds < 1800
  const statusConfig = EnvStatusMap[envStatus || 'terminated']
  const envRunning = isWorkbenchEnvRunning(envStatus)
  const useOverlaySidebar = activeTab === 'visualization'

  const roleConfig = useMemo(
    () => experiment?.blueprint?.collaboration?.roles?.find((role) => role.key === currentMember?.role_key),
    [currentMember?.role_key, experiment?.blueprint?.collaboration?.roles],
  )

  const allowedToolKeys = roleConfig?.tool_keys?.length ? new Set(roleConfig.tool_keys) : null

  const allowedNodeKeys = useMemo(() => {
    const keys = new Set<string>()

    if (currentMember?.assigned_node_key) {
      keys.add(currentMember.assigned_node_key)
    }

    for (const key of roleConfig?.node_keys || []) {
      if (key) {
        keys.add(key)
      }
    }

    return keys.size > 0 ? keys : null
  }, [currentMember?.assigned_node_key, roleConfig?.node_keys])

  const visibleInstances = useMemo(() => {
    if (!allowedNodeKeys) {
      return instances
    }

    return instances.filter((instance) => (
      instance.kind !== 'node' || allowedNodeKeys.has(instance.instance_key)
    ))
  }, [allowedNodeKeys, instances])

  const visibleInstanceKeys = useMemo(
    () => new Set(visibleInstances.map((instance) => instance.instance_key)),
    [visibleInstances],
  )

  useEffect(() => {
    if (visibleInstances.length === 0) {
      setSelectedInstanceKey(undefined)
      return
    }

    if (selectedInstanceKey && visibleInstances.some((instance) => instance.instance_key === selectedInstanceKey)) {
      return
    }

    setSelectedInstanceKey(visibleInstances[0].instance_key)
  }, [selectedInstanceKey, visibleInstances])

  useEffect(() => {
    setSidebarOpen(!useOverlaySidebar)
  }, [useOverlaySidebar])

  const toolGroups = useMemo(() => {
    const filteredTools = (allowedToolKeys
      ? availableTools.filter((tool) => allowedToolKeys.has(tool.key))
      : availableTools)
      .filter((tool) => !tool.target || visibleInstanceKeys.has(tool.target))

    return filteredTools.reduce<Map<ExperimentWorkbenchTabKey, ExperimentWorkbenchProps['availableTools']>>((map, tool) => {
      const list = map.get(tool.key) || []
      list.push(tool)
      map.set(tool.key, list)
      return map
    }, new Map())
  }, [allowedToolKeys, availableTools, visibleInstanceKeys])

  const resolvedTools = useMemo(() => new Map(
    Array.from(toolGroups.entries()).map(([key, tools]) => {
      const matched = selectedInstanceKey
        ? tools.find((tool) => tool.target === selectedInstanceKey)
        : undefined

      return [key, matched || tools[0]]
    }),
  ), [selectedInstanceKey, toolGroups])

  const toolTabs = useMemo(
    () => Array.from(resolvedTools.values()).filter((tool): tool is NonNullable<typeof tool> => Boolean(tool)),
    [resolvedTools],
  )

  const visualizationTool = resolvedTools.get('visualization')
  const visualizationModuleKey = visualizationTool?.moduleKey
    || experiment?.blueprint?.tools?.find((tool) => tool.key === 'visualization')?.kind
  const selectedInstance = visibleInstances.find((instance) => instance.instance_key === selectedInstanceKey)
  const totalCheckpoints = experiment?.blueprint?.grading?.checkpoints?.length || 0
  const totalAssets = experiment?.blueprint?.content?.assets?.length || 0
  const totalNodes = experiment?.blueprint?.nodes?.length || 0
  const totalServices = experiment?.blueprint?.services?.length || 0

  const renderIframePanel = (toolKey: ExperimentWorkbenchTabKey, title: string, tips: string) => {
    const tool = resolvedTools.get(toolKey)
    return tool?.accessUrl
      ? <iframe src={`${tool.accessUrl}/`} className="h-full w-full border-0" title={title} />
      : renderEnvPlaceholder(`${title} 暂不可用`, tips)
  }

  const renderToolPanel = (tabKey: ExperimentWorkbenchTabKey) => {
    switch (tabKey) {
      case 'ide':
        return resolvedTools.get('ide')?.accessUrl
          ? (
            ideReady
              ? renderIframePanel('ide', 'IDE', 'IDE 运行时尚未就绪。')
              : renderEnvPlaceholder('IDE 准备中', '正在等待编辑器完成初始化。')
          )
          : renderEnvPlaceholder('IDE 暂不可用', '当前环境没有提供 IDE 入口。')
      case 'terminal':
        return renderIframePanel('terminal', '终端', '当前实例没有可用的终端入口。')
      case 'files':
        return resolvedTools.get('files')?.accessUrl
          ? <FileManager experimentId={experimentId} accessUrl={resolvedTools.get('files')?.accessUrl} />
          : renderEnvPlaceholder('文件管理不可用', '当前实例没有可用的文件浏览入口。')
      case 'explorer':
        return resolvedTools.get('explorer')?.accessUrl
          ? <BlockExplorer accessUrl={resolvedTools.get('explorer')?.accessUrl} />
          : renderEnvPlaceholder('区块浏览器不可用', '当前环境没有提供区块浏览器入口。')
      case 'logs':
        return resolvedTools.get('logs')?.accessUrl
          ? <LogViewer accessUrl={resolvedTools.get('logs')?.accessUrl} />
          : renderEnvPlaceholder('日志不可用', '当前实例没有可用的日志入口。')
      case 'visualization':
        return visualizationTool?.accessUrl
          ? (
            <ExperimentVisualizationPanel
              experiment={experiment}
              accessUrl={visualizationTool.accessUrl}
              wsUrl={visualizationTool.wsUrl}
              moduleKey={visualizationModuleKey || getWorkbenchVisualizationModuleKey(visualizationTool)}
            />
          )
          : renderEnvPlaceholder('可视化不可用', '当前环境没有提供可视化入口。')
      case 'api_debug':
        return resolvedTools.get('api_debug')?.accessUrl
          ? (
            <ApiDebugger
              accessUrl={resolvedTools.get('api_debug')?.accessUrl}
              title="链上接口调试器"
              description="通过实验运行时代理直接调用当前链服务，可用于 REST / JSON-RPC 接口排查。"
            />
          )
          : renderEnvPlaceholder('API 调试不可用', '当前环境没有提供 API 调试入口。')
      case 'network':
        return renderIframePanel('network', '组网面板', '当前实例没有提供网络连接或节点关系视图。')
      case 'rpc':
        return resolvedTools.get('rpc')?.accessUrl
          ? (
            <ApiDebugger
              accessUrl={resolvedTools.get('rpc')?.accessUrl}
              title="链节点 RPC 调试器"
              description="直接向当前实验链节点发送 JSON-RPC 请求，用于查看区块、交易、账户与状态数据。"
            />
          )
          : renderEnvPlaceholder('RPC 不可用', '当前环境没有提供链节点 RPC 入口。')
      default:
        return renderEnvPlaceholder('工具不可用', '当前工具未提供可用入口。')
    }
  }

  const handleSendMessage = () => {
    const trimmed = chatMessage.trim()
    if (!trimmed) {
      return
    }

    onSendMessage(trimmed)
    setChatMessage('')
  }

  const memberRoleOptions = (experiment?.blueprint?.collaboration?.roles || []).map((role) => ({
    label: role.label || role.key,
    value: role.key,
  }))

  const memberNodeOptions = (experiment?.blueprint?.nodes || []).map((node) => ({
    label: node.name || node.key,
    value: node.key,
  }))

  const memberStatusText = (status?: string) => {
    switch (status) {
      case 'joined':
        return '已加入'
      case 'left':
        return '已离开'
      default:
        return status || '未知'
    }
  }

  const sidebarContent = (
    <div className="flex h-full min-h-0 flex-col gap-2.5 overflow-auto pr-1">
      <div className="rounded-2xl border border-slate-200 bg-white p-3">
        <div className="mb-2 flex items-center gap-2 text-sm font-semibold">
          <ApiOutlined />
          当前环境
        </div>
        {selectedInstance ? (
          <div className="space-y-2 text-xs text-slate-600">
            <div>当前区域：{selectedInstance.instance_key}</div>
            <div>区域类型：{instanceKindText(selectedInstance.kind)}</div>
            <div>运行状态：{instanceStatusText(selectedInstance.status)}</div>
            {experiment?.blueprint?.content?.assets?.length ? (
              <div>实验资料：{experiment.blueprint.content.assets.length} 项</div>
            ) : null}
          </div>
        ) : (
          <div className="text-xs text-slate-500">当前没有可用操作区域。</div>
        )}
      </div>

      <div className="rounded-2xl border border-slate-200 bg-white p-3">
        <div className="mb-2 text-sm font-semibold text-slate-900">环境概览</div>
        <div className="grid gap-2 sm:grid-cols-2 xl:grid-cols-1">
          <div className="rounded-xl border border-slate-200 bg-slate-50 px-3 py-2">
            <div className="text-xs uppercase tracking-[0.16em] text-slate-500">工具数量</div>
            <div className="mt-1.5 text-sm text-slate-900">{toolTabs.length} 项</div>
          </div>
          <div className="rounded-xl border border-slate-200 bg-slate-50 px-3 py-2">
            <div className="text-xs uppercase tracking-[0.16em] text-slate-500">实验时长</div>
            <div className="mt-1.5 text-sm text-slate-900">{experiment ? formatDuration(experiment.estimated_time * 60) : '-'}</div>
          </div>
          <div className="rounded-xl border border-slate-200 bg-slate-50 px-3 py-2">
            <div className="text-xs uppercase tracking-[0.16em] text-slate-500">节点 / 服务</div>
            <div className="mt-1.5 text-sm text-slate-900">{totalNodes} / {totalServices}</div>
          </div>
          <div className="rounded-xl border border-slate-200 bg-slate-50 px-3 py-2">
            <div className="text-xs uppercase tracking-[0.16em] text-slate-500">评分方式</div>
            <div className="mt-1.5 text-sm text-slate-900">{experiment?.auto_grade ? '自动评测' : '人工批改'}</div>
          </div>
        </div>
        {visibleInstances.length > 0 ? (
          <div className="mt-2 space-y-1.5">
            {visibleInstances.map((instance) => (
              <div key={instance.instance_key} className="rounded-xl border border-slate-200 px-3 py-2 text-xs">
                <div className="flex items-center justify-between gap-3">
                  <span className="font-medium text-slate-900">{instance.instance_key}</span>
                  <Tag color={instance.status === 'running' ? 'success' : 'default'}>{instanceStatusText(instance.status)}</Tag>
                </div>
              </div>
            ))}
          </div>
        ) : null}
      </div>

      <div className="rounded-2xl border border-slate-200 bg-white p-3">
        <div className="mb-2 text-sm font-semibold text-slate-900">实验信息</div>
        <div className="space-y-2 text-xs text-slate-600">
          <div className="rounded-xl border border-slate-200 bg-slate-50 px-3 py-2">
            <div className="text-[11px] uppercase tracking-[0.16em] text-slate-500">实验资料</div>
            <div className="mt-1.5 text-sm text-slate-900">{totalAssets} 项</div>
            <div className="mt-1 text-slate-500">
              {totalAssets > 0 ? '进入环境后会自动准备相关资料。' : '当前没有额外实验资料。'}
            </div>
          </div>
          <div className="rounded-xl border border-slate-200 bg-slate-50 px-3 py-2">
            <div className="text-[11px] uppercase tracking-[0.16em] text-slate-500">评测信息</div>
            <div className="mt-1.5 text-sm text-slate-900">
              {experiment?.auto_grade ? '自动评测' : '人工批改'} / {totalCheckpoints} 项检查
            </div>
            <div className="mt-1 text-slate-500">
              完成实验后可提交并查看结果。
            </div>
          </div>
        </div>
      </div>

      <div className="rounded-2xl border border-slate-200 bg-white p-3">
        <div className="mb-2 flex items-center gap-2 text-sm font-semibold">
          <TeamOutlined />
          协作成员
        </div>
        <div className="space-y-1.5">
          {sessionMembers.length > 0 ? sessionMembers.map((member) => (
            <div key={`${member.user_id}-${member.joined_at}`} className="rounded-xl border border-slate-200 px-3 py-2 text-xs">
              <div className="font-medium text-slate-900">
                {member.real_name || member.display_name || member.phone || `用户 ${member.user_id}`}
              </div>
              {canManageSessionMembers && onUpdateSessionMember ? (
                <div className="mt-2 space-y-2">
                  <Select
                    size="small"
                    className="w-full"
                    value={member.role_key || undefined}
                    placeholder="角色"
                    options={memberRoleOptions}
                    onChange={(value) => onUpdateSessionMember(member.user_id, { role_key: value })}
                  />
                  <Select
                    size="small"
                    className="w-full"
                    value={member.assigned_node_key || undefined}
                    placeholder="节点"
                    allowClear
                    options={memberNodeOptions}
                    onChange={(value) => onUpdateSessionMember(member.user_id, { assigned_node_key: value })}
                  />
                  <Select
                    size="small"
                    className="w-full"
                    value={member.join_status}
                    options={[
                      { label: '已加入', value: 'joined' },
                      { label: '已离开', value: 'left' },
                    ]}
                    onChange={(value) => onUpdateSessionMember(member.user_id, { join_status: value as 'joined' | 'left' })}
                  />
                </div>
              ) : (
                <>
                  <div className="mt-1 text-slate-600">角色：{member.role_key || '未分配'}</div>
                  <div className="text-slate-600">节点：{member.assigned_node_key || '未分配'}</div>
                  <div className="text-slate-600">状态：{memberStatusText(member.join_status)}</div>
                </>
              )}
            </div>
          )) : (
            <div className="text-xs text-slate-500">当前没有协作成员信息。</div>
          )}
        </div>
      </div>

      <div className="flex min-h-[180px] flex-1 flex-col rounded-2xl border border-slate-200 bg-white p-3 xl:min-h-0">
        <div className="mb-2 flex items-center gap-2 text-sm font-semibold">
          <MessageOutlined />
          协作消息
        </div>
        <div className="mb-2.5 min-h-0 flex-1 space-y-1.5 overflow-auto pr-1">
          {sessionMessages.length > 0 ? sessionMessages.map((message) => (
            <div key={message.id} className="rounded-xl border border-slate-200 px-3 py-2.5 text-xs">
              <div className="font-medium text-slate-900">
                {message.real_name || message.display_name || message.phone || `用户 ${message.user_id}`}
              </div>
              <div className="mt-1 whitespace-pre-wrap text-slate-700">{message.message}</div>
            </div>
          )) : (
            <div className="text-xs text-slate-500">当前会话还没有消息。</div>
          )}
        </div>
        <Space.Compact>
          <Input
            value={chatMessage}
            onChange={(event) => setChatMessage(event.target.value)}
            onPressEnter={handleSendMessage}
            placeholder="发送一条协作消息"
          />
          <Button type="primary" onClick={handleSendMessage}>发送</Button>
        </Space.Compact>
      </div>
    </div>
  )

  return (
    <>
      <div className="-m-6 flex h-[calc(100vh-64px)] min-h-[640px] flex-col overflow-hidden bg-slate-50 xl:rounded-[24px] xl:border xl:border-slate-200 xl:shadow-sm">
        <div className="border-b border-slate-200 bg-white px-4 py-2 text-slate-900">
          <div className="grid gap-2.5 xl:grid-cols-[minmax(0,1.45fr)_minmax(260px,360px)_auto] xl:items-center">
            <div className="min-w-0">
              <div className="flex flex-wrap items-center gap-2">
                <span className="rounded-full border border-slate-200 bg-slate-50 px-2 py-0.5 text-[11px] font-medium uppercase tracking-[0.18em] text-slate-500">
                  实验工作台
                </span>
                <span className="truncate text-base font-semibold">{experiment?.title}</span>
                <Tag color={statusConfig.color} bordered>
                  {statusConfig.text}
                </Tag>
                {visibleInstances.length > 0 ? (
                  <Tag color="cyan" bordered>
                    实例 {visibleInstances.length}
                  </Tag>
                ) : null}
                {sessionMembers.length > 0 ? (
                  <Tag color="purple" bordered icon={<TeamOutlined />}>
                    协作 {sessionMembers.length}
                  </Tag>
                ) : null}
              </div>
              <div className="mt-1.5 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-slate-600">
                <span>时长 {experiment ? formatDuration(experiment.estimated_time * 60) : '-'}</span>
                <span>节点 {totalNodes} / 服务 {totalServices}</span>
                <span>资料 {totalAssets} 项</span>
                <span>{experiment?.auto_grade ? '自动评测' : '人工批改'} / {totalCheckpoints} 项检查</span>
              </div>
            </div>

            <div className="flex flex-wrap items-center gap-2 rounded-2xl border border-slate-200 bg-slate-50 px-3 py-2">
              {remainingSeconds > 0 ? (
                <span className={`flex items-center rounded-full border border-slate-200 bg-white px-3 py-1 text-xs ${isTimeLow ? 'font-semibold text-warning' : 'text-slate-900'}`}>
                  <ClockCircleOutlined className="mr-2" />
                  {formatDuration(remainingSeconds)}
                </span>
              ) : null}
              {isTimeLow ? (
                <Button type="primary" size="small" onClick={onExtend}>
                  延长环境
                </Button>
              ) : null}
              {visibleInstances.length > 1 ? (
                <Select
                  size="small"
                  className="min-w-[188px] flex-1"
                  value={selectedInstanceKey}
                  onChange={setSelectedInstanceKey}
                  options={visibleInstances.map((instance) => ({
                    label: `${instance.instance_key} (${instanceKindText(instance.kind)})`,
                    value: instance.instance_key,
                  }))}
                />
              ) : selectedInstance ? (
                <span className="text-xs text-slate-600">
                  当前区域：<span className="font-medium text-slate-900">{selectedInstance.instance_key}</span>
                </span>
              ) : null}
              <Button
                size="small"
                icon={<InfoCircleOutlined />}
                onClick={() => setSidebarOpen(true)}
              >
                {useOverlaySidebar ? '实验摘要' : '打开摘要'}
              </Button>
            </div>

            <div className="flex flex-wrap items-center justify-start gap-2 xl:justify-end">
              {envStatus === 'running' ? (
                <Button size="small" icon={<PauseCircleOutlined />} onClick={onPause}>
                  暂停
                </Button>
              ) : null}

              {envStatus === 'paused' ? (
                <Button size="small" icon={<PlayCircleOutlined />} onClick={onResume}>
                  恢复
                </Button>
              ) : null}

              {(envStatus === 'running' || envStatus === 'paused') ? (
                <Button size="small" icon={<SaveOutlined />} onClick={onCreateSnapshot}>
                  快照
                </Button>
              ) : null}

              {snapshotUrl ? (
                <Button size="small" onClick={onRestoreSnapshot}>
                  恢复快照
                </Button>
              ) : null}

              {envRunning ? (
                <Button size="small" icon={<PauseCircleOutlined />} onClick={onStop}>
                  停止
                </Button>
              ) : null}

              <Button size="small" onClick={onExit}>
                退出
              </Button>
              <Button size="small" type="primary" onClick={onOpenSubmit}>
                提交实验
              </Button>
            </div>
          </div>
        </div>

        <div className="flex min-h-0 flex-1 flex-col xl:flex-row">
          <div className="border-b border-slate-200 bg-white xl:w-[76px] xl:border-r xl:border-b-0">
            <div className="flex gap-2 overflow-x-auto px-2.5 py-2.5 xl:flex-col xl:items-center xl:overflow-x-visible">
              {toolTabs.map((tool) => (
                <Button
                  key={tool.key}
                  type={activeTab === tool.key ? 'primary' : 'text'}
                  icon={tool.icon}
                  onClick={() => onSetActiveTab(tool.key)}
                  className="shrink-0"
                  title={tool.title}
                >
                  <span className="xl:hidden">{tool.label}</span>
                </Button>
              ))}
            </div>
          </div>

          <div className="flex min-h-[300px] min-w-0 flex-1 flex-col bg-white">
            {toolTabs.length > 0 && envRunning ? (
              <div className="border-b border-slate-200 bg-slate-50 px-3 py-1.5">
                <div className="flex flex-wrap items-center gap-2 text-xs text-slate-600">
                  <span className="uppercase tracking-[0.24em] text-slate-500">当前工具</span>
                  {toolTabs.map((tool) => (
                    <Tag key={tool.key} color={activeTab === tool.key ? 'processing' : 'default'}>
                      {tool.label}
                    </Tag>
                  ))}
                </div>
              </div>
            ) : null}

            <div className="min-h-0 flex-1">
              {!envRunning ? (
                <div className="flex h-full items-center justify-center text-slate-900">
                  <div className="text-center">
                    <p className="mb-4 text-lg">
                      环境
                      {envStatus === 'creating'
                        ? '创建中'
                        : envStatus === 'paused'
                          ? '已暂停'
                          : envStatus === 'failed'
                            ? '启动失败'
                            : '已停止'}
                    </p>
                    {envStatus === 'failed' ? (
                      <p className="mx-auto mb-4 max-w-xl text-sm text-slate-600">
                        {envErrorMessage || '运行时实例没有按预期就绪，请查看环境概览中的实例状态并切换到日志面板排查。'}
                      </p>
                    ) : null}
                    {envStatus === 'creating' ? <Spin /> : null}
                  </div>
                </div>
              ) : toolTabs.length === 0 ? (
                renderEnvPlaceholder('暂无可用工具', '当前环境还没有可用入口，请稍后重试。')
              ) : (
                <div className="h-full">
                  {mountedTabs.map((tabKey) => (
                    <KeepAliveTabPanel key={tabKey} active={activeTab === tabKey}>
                      {renderToolPanel(tabKey)}
                    </KeepAliveTabPanel>
                  ))}
                </div>
              )}
            </div>
          </div>

          {!useOverlaySidebar && sidebarOpen ? (
            <div className="w-full border-t border-slate-200 bg-white p-3 text-slate-900 xl:w-[320px] xl:border-l xl:border-t-0">
              {sidebarContent}
            </div>
          ) : null}
        </div>
      </div>

      {useOverlaySidebar ? (
        <Drawer
          title="实验信息与协作"
          placement="right"
          width={360}
          open={sidebarOpen}
          onClose={() => setSidebarOpen(false)}
          styles={{
            body: {
              padding: 12,
              background: '#f8fafc',
            },
            header: {
              background: '#f8fafc',
              color: '#0f172a',
              borderBottom: '1px solid rgba(148,163,184,0.28)',
            },
          }}
        >
          <div className="text-slate-900">
            {sidebarContent}
          </div>
        </Drawer>
      ) : null}

      <Modal
        title="提交实验"
        open={submitModalVisible}
        onCancel={onCloseSubmit}
        onOk={onSubmitExperiment}
        confirmLoading={submitting}
        okText="提交并结束"
        width={600}
      >
        <div className="py-4">
          <p className="mb-4 text-text-secondary">
            提交后实验环境会被停止，请确认已经保存必要的代码、日志和实验产物。
          </p>
          <Input.TextArea
            placeholder="实验报告为可选，用于记录你的思路、关键操作和结论。"
            value={submitReport}
            onChange={(event) => onSubmitReportChange(event.target.value)}
            rows={6}
          />
        </div>
      </Modal>
    </>
  )
}
