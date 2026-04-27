import { useCallback, useEffect, useRef, useState } from 'react'
import { Alert, Button, Card, Descriptions, Drawer, message, Select, Slider, Space, Spin, Switch, Tag } from 'antd'
import {
  ApiOutlined,
  DisconnectOutlined,
  FileSearchOutlined,
  PauseCircleOutlined,
  PlayCircleOutlined,
  ReloadOutlined,
  SaveOutlined,
  SettingOutlined,
  StepForwardOutlined,
} from '@ant-design/icons'
import { SimulationClient } from '@/domains/visualization/runtime/simulationClient'
import {
  getVisualizationCapabilityLabel,
  getVisualizationComponentTypeLabel,
  getVisualizationModuleLabel,
  getVisualizationStatusLabel,
  hasVisualizationCapability,
} from '@/domains/visualization/runtime/visualizationMeta'
import type {
  SimulationWrapperProps,
  SimulatorDescription,
  SimulatorEvent,
  SimulatorParam,
  SimulatorState,
  VisualizationActionDefinition,
} from '@/types/visualizationDomain'
import { isSystemVisualizationEvent } from '@/domains/visualization/runtime/visualizationEventLabels'
import {
  getVisualizationEventSummary,
  getVisualizationEventTitle,
} from '@/domains/visualization/runtime/visualizationEventFormatter'
import SimulationActionPanel from './SimulationActionPanel'

const STRUCTURAL_PARAM_KEYS = new Set([
  'node_count',
  'byzantine_count',
  'topology_type',
  'network_size',
  'bootstrap_count',
])

const DEFAULT_SIMULATION_SPEED = 1

function getParamEffectiveValue(param: SimulatorParam): unknown {
  return param.value ?? param.default
}

function getParamPayload(params: SimulatorParam[]): Record<string, unknown> {
  return params.reduce<Record<string, unknown>>((result, item) => {
    result[item.key] = getParamEffectiveValue(item)
    return result
  }, {})
}

function getVisibleEvents(events: SimulatorEvent[]): SimulatorEvent[] {
  return events.filter((event) => !isSystemVisualizationEvent(event.type))
}

function getStateNodeIds(state: SimulatorState | null): string[] {
  if (!state?.nodes) {
    return []
  }

  if (Array.isArray(state.nodes)) {
    return state.nodes
      .map((node) => node?.id)
      .filter((nodeId): nodeId is string => typeof nodeId === 'string' && nodeId.length > 0)
  }

  return Object.keys(state.nodes)
}

function normalizeActionDefinitions(
  definitions: VisualizationActionDefinition[],
  state: SimulatorState | null,
): VisualizationActionDefinition[] {
  const nodeIds = getStateNodeIds(state)
  if (nodeIds.length === 0) {
    return definitions
  }

  return definitions.map((definition) => ({
    ...definition,
    fields: definition.fields?.map((field) => {
      if (field.key !== 'target' || field.type !== 'select') {
        return field
      }

      return {
        ...field,
        defaultValue: nodeIds.includes(String(field.defaultValue)) ? field.defaultValue : nodeIds[0],
        options: nodeIds.map((nodeId) => ({
          label: nodeId,
          value: nodeId,
        })),
      }
    }),
  }))
}

function normalizeErrorMessage(value: unknown): string {
  if (value instanceof Error && value.message) {
    return value.message
  }

  if (typeof value === 'string' && value.trim()) {
    return value
  }

  return '连接模拟器失败。'
}

/**
 * SimulationWrapper 统一负责 simulations 服务的初始化、控制和状态同步。
 */
export default function SimulationWrapper({
  module,
  initialParams,
  nodeCount = 4,
  renderState,
  renderEvents,
  renderParams,
  actionDefinitions = [],
  onStateUpdate,
  onEvent,
  onConnectionChange,
  accessUrl,
  wsUrl,
  children,
}: SimulationWrapperProps) {
  const clientRef = useRef<SimulationClient | null>(null)
  const pollTimerRef = useRef<number | null>(null)
  const lastEventTickRef = useRef(0)

  const [connected, setConnected] = useState(false)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [meta, setMeta] = useState<SimulatorDescription | null>(null)
  const [state, setState] = useState<SimulatorState | null>(null)
  const [events, setEvents] = useState<SimulatorEvent[]>([])
  const [params, setParams] = useState<SimulatorParam[]>([])
  const [speed, setSpeed] = useState(DEFAULT_SIMULATION_SPEED)
  const [drawerMode, setDrawerMode] = useState<'actions' | 'events' | 'settings' | null>(null)
  const [retryToken, setRetryToken] = useState(0)

  const updateState = useCallback((nextState: SimulatorState) => {
    setState(nextState)
    onStateUpdate?.(nextState)
  }, [onStateUpdate])

  const mergeEvents = useCallback((nextEvents: SimulatorEvent[]) => {
    if (nextEvents.length === 0) {
      return
    }

    setEvents((previous) => {
      const merged = [...previous, ...nextEvents]
      const deduped = new Map<string, SimulatorEvent>()

      for (const event of merged) {
        deduped.set(event.id, event)
      }

      return Array.from(deduped.values()).slice(-100)
    })

    for (const event of nextEvents) {
      onEvent?.(event)
      lastEventTickRef.current = Math.max(lastEventTickRef.current, event.tick)
    }
  }, [onEvent])

  const stopPolling = useCallback(() => {
    if (pollTimerRef.current !== null) {
      window.clearInterval(pollTimerRef.current)
      pollTimerRef.current = null
    }
  }, [])

  const refreshSnapshot = useCallback(async () => {
    const client = clientRef.current
    if (!client) {
      return
    }

    const [nextState, nextEvents, nextParams] = await Promise.all([
      client.getState(),
      client.getEvents(lastEventTickRef.current, 50),
      client.getParams(),
    ])

    updateState(nextState)
    mergeEvents(nextEvents)
    setParams(nextParams)
  }, [mergeEvents, updateState])

  const syncSimulatorSnapshot = useCallback(async (sinceTick = 0) => {
    const client = clientRef.current
    if (!client) {
      return
    }

    const [metaData, paramsData, stateData, eventData] = await Promise.all([
      client.getMeta(),
      client.getParams(),
      client.getState(),
      client.getEvents(sinceTick, 50),
    ])

    setMeta(metaData)
    setParams(paramsData)
    updateState(stateData)
    mergeEvents(eventData)
  }, [mergeEvents, updateState])

  const startPolling = useCallback(() => {
    stopPolling()
    pollTimerRef.current = window.setInterval(() => {
      refreshSnapshot().catch(() => {
        // 轮询失败时保留当前页面状态，避免界面闪烁。
      })
    }, 1000)
  }, [refreshSnapshot, stopPolling])

  useEffect(() => {
    let disposed = false
    const client = new SimulationClient(accessUrl, wsUrl)
    clientRef.current = client

    client.on('connected', () => {
      if (disposed) {
        return
      }

      setConnected(true)
      onConnectionChange?.(true)
      stopPolling()
    })

    client.on('disconnected', () => {
      if (disposed) {
        return
      }

      setConnected(false)
      onConnectionChange?.(false)
      startPolling()
    })

    client.on('state', (payload) => {
      if (disposed) {
        return
      }

      updateState(payload as SimulatorState)
    })

    client.on('event', (payload) => {
      if (disposed) {
        return
      }

      mergeEvents([payload as SimulatorEvent])
    })

    client.on('error', (payload) => {
      if (disposed || !client.isConnected()) {
        return
      }

      message.error(normalizeErrorMessage(payload))
    })

    const init = async () => {
      setLoading(true)
      setError(null)
      setEvents([])
      lastEventTickRef.current = 0

      try {
        try {
          await client.connect()
        } catch {
          // WebSocket 不可用时继续使用 HTTP 轮询模式。
        }

        await client.init(module, initialParams, nodeCount)
        await client.setSpeed(DEFAULT_SIMULATION_SPEED)
        await syncSimulatorSnapshot(0)
        setConnected(client.isConnected())
        setLoading(false)
        startPolling()
      } catch (initError) {
        if (disposed) {
          return
        }

        setError(normalizeErrorMessage(initError))
        setLoading(false)
      }
    }

    void init()

    return () => {
      disposed = true
      stopPolling()
      client.disconnect()
    }
  }, [
    accessUrl,
    initialParams,
    mergeEvents,
    module,
    nodeCount,
    onConnectionChange,
    retryToken,
    startPolling,
    stopPolling,
    syncSimulatorSnapshot,
    updateState,
    wsUrl,
  ])

  const runAction = useCallback(async (
    action: () => Promise<void>,
    successMessage?: string,
  ) => {
    try {
      await action()
      await refreshSnapshot()
      if (successMessage) {
        message.success(successMessage)
      }
    } catch (actionError) {
      message.error(normalizeErrorMessage(actionError))
    }
  }, [refreshSnapshot])

  const handleStart = useCallback(async () => {
    await runAction(async () => {
      await clientRef.current?.start()
    }, '已开始自动推进')
  }, [runAction])

  const handlePause = useCallback(async () => {
    await runAction(async () => {
      await clientRef.current?.pause()
    }, '已暂停当前过程')
  }, [runAction])

  const handleResume = useCallback(async () => {
    await runAction(async () => {
      await clientRef.current?.resume()
    }, '已继续当前过程')
  }, [runAction])

  const handleStep = useCallback(async () => {
    try {
      const nextState = await clientRef.current?.step()
      if (nextState) {
        updateState(nextState)
      }
      await refreshSnapshot()
    } catch (actionError) {
      message.error(normalizeErrorMessage(actionError))
    }
  }, [refreshSnapshot, updateState])

  const handleReset = useCallback(async () => {
    const client = clientRef.current
    if (!client) {
      return
    }

    const wasRunning = state?.status === 'running'
    const initParams = getParamPayload(params)
    const nextNodeCount = Number(initParams.node_count ?? nodeCount)

    try {
      setEvents([])
      lastEventTickRef.current = 0

      await client.reset()
      await client.init(module, initParams, nextNodeCount)
      await client.setSpeed(speed)
      await syncSimulatorSnapshot(0)

      if (wasRunning) {
        await client.start()
        await refreshSnapshot()
      }

      message.success('模拟已重置')
    } catch (actionError) {
      message.error(normalizeErrorMessage(actionError))
    }
  }, [module, nodeCount, params, refreshSnapshot, speed, state?.status, syncSimulatorSnapshot])

  const handleSpeedChange = useCallback(async (value: number) => {
    setSpeed(value)
    try {
      await clientRef.current?.setSpeed(value)
    } catch (actionError) {
      message.error(normalizeErrorMessage(actionError))
    }
  }, [])

  const handleSetParam = useCallback(async (key: string, value: unknown) => {
    const client = clientRef.current
    if (!client) {
      return
    }

    try {
      const nextParams = params.map((item) => (
        item.key === key ? { ...item, value } : item
      ))
      setParams(nextParams)

      if (STRUCTURAL_PARAM_KEYS.has(key)) {
        const initParams = getParamPayload(nextParams)
        setEvents([])
        lastEventTickRef.current = 0
        await client.init(module, initParams, Number(initParams.node_count ?? nodeCount))
        await client.setSpeed(speed)
        await syncSimulatorSnapshot(0)
        message.success('已按新的结构参数重新载入当前场景')
        return
      }

      await client.setParam(key, value)
      await refreshSnapshot()
    } catch (actionError) {
      message.error(normalizeErrorMessage(actionError))
    }
  }, [module, nodeCount, params, refreshSnapshot, speed, syncSimulatorSnapshot])

  const handleExecuteVisualizationAction = useCallback(async (
    definition: VisualizationActionDefinition,
    values: Record<string, unknown>,
  ) => {
    const client = clientRef.current
    if (!client) {
      return
    }

    const payload = {
      ...(definition.preset || {}),
      ...values,
    }

    try {
      if (definition.kind === 'module_action') {
        await client.executeAction(definition.action || '', payload)
      } else if (definition.kind === 'inject_fault') {
        await client.injectFault(payload as never)
      } else if (definition.kind === 'inject_attack') {
        await client.injectAttack(payload as never)
      }

      await refreshSnapshot()
      message.success(definition.successMessage)
    } catch (actionError) {
      message.error(normalizeErrorMessage(actionError))
    }
  }, [refreshSnapshot])

  const handleSaveSnapshot = useCallback(async () => {
    const snapshotName = `snapshot_${Date.now()}`
    await runAction(async () => {
      await clientRef.current?.saveSnapshot(snapshotName)
    }, '快照已保存')
  }, [runAction])

  const handleReconnect = useCallback(() => {
    setRetryToken((value) => value + 1)
  }, [])

  if (loading) {
    return (
      <div className="flex h-full items-center justify-center rounded-[24px] bg-[linear-gradient(180deg,#eff6ff_0%,#f8fafc_100%)]">
        <Spin size="large" tip="正在连接模拟器...">
          <div />
        </Spin>
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex h-full items-center justify-center rounded-[24px] bg-[linear-gradient(180deg,#eff6ff_0%,#f8fafc_100%)] p-4">
        <Alert
          type="error"
          message="连接模拟器失败"
          description={error}
          action={<Button onClick={handleReconnect}>重试</Button>}
        />
      </div>
    )
  }

  const isRunning = state?.status === 'running'
  const isPaused = state?.status === 'paused'
  const visibleEvents = getVisibleEvents(events)
  const normalizedActionDefinitions = normalizeActionDefinitions(actionDefinitions, state)
  const capabilitySummary = meta?.capabilities?.slice(0, 3) || []
  const canAdjustParams = hasVisualizationCapability(meta, 'param_panel') && params.length > 0
  const canUseTimeControl = hasVisualizationCapability(meta, 'time_control')
  const canShowStatusMonitor = hasVisualizationCapability(meta, 'state_monitor')
  const canShowEventLog = hasVisualizationCapability(meta, 'event_log')
  const canUseSnapshots = hasVisualizationCapability(meta, 'snapshot')
  const shouldShowSettings = canAdjustParams || Boolean(meta)
  const hasPrimaryActions = normalizedActionDefinitions.length > 0
  const eventPreview = visibleEvents.slice(-8).reverse()
  const drawerTitle = drawerMode === 'actions'
    ? '实验动作'
    : drawerMode === 'events'
      ? '最近记录'
      : '参数与说明'

  return (
    <div className="flex h-full min-h-0 flex-col overflow-hidden rounded-[22px] border border-slate-200 bg-[#f7fafc] text-slate-900 shadow-[0_18px_40px_rgba(15,23,42,0.08)]">
      <div className="border-b border-slate-200 bg-[linear-gradient(135deg,#eef5ff_0%,#e7f0ff_45%,#d9e8ff_100%)] px-3 py-2">
        <div className="flex flex-col gap-1.5 xl:flex-row xl:items-center xl:justify-between">
          <div className="flex min-w-0 flex-wrap items-center gap-2">
            <Tag color={connected ? 'green' : 'orange'} className="flex items-center gap-2">
              {connected ? <ApiOutlined /> : <DisconnectOutlined />}
              {connected ? '实时连接' : '轮询模式'}
            </Tag>
            {meta ? (
              <span className="truncate text-sm font-semibold text-slate-900 sm:text-base">{meta.name}</span>
            ) : null}
            {meta ? (
              <Tag color="blue">
                {getVisualizationComponentTypeLabel(meta.type)}
              </Tag>
            ) : null}
            {state && canShowStatusMonitor ? (
              <Tag color="gold">
                步骤 {state.tick}
              </Tag>
            ) : null}
            <Tag color={state?.status === 'running' ? 'success' : state?.status === 'paused' ? 'warning' : 'default'}>
              {getVisualizationStatusLabel(state?.status)}
            </Tag>
            {canShowEventLog ? <Tag color="cyan">记录 {visibleEvents.length}</Tag> : null}
            {canAdjustParams ? <Tag color="purple">参数 {params.length}</Tag> : null}
            {capabilitySummary.map((capability) => (
              <Tag key={capability}>
                {getVisualizationCapabilityLabel(capability)}
              </Tag>
            ))}
          </div>

          <Space wrap size={[8, 8]}>
            {canUseTimeControl ? (
              <>
                <span className="text-xs text-slate-600">推进速度</span>
                <Slider
                  min={0.1}
                  max={10}
                  step={0.1}
                  value={speed}
                  onChange={handleSpeedChange}
                  className="w-[120px] sm:w-[170px]"
                  tooltip={{ formatter: (value) => `${value}x` }}
                />
                {!isRunning ? (
                  <>
                    <Button icon={<PlayCircleOutlined />} type="primary" onClick={isPaused ? handleResume : handleStart}>
                      {isPaused ? '继续推进' : '开始推进'}
                    </Button>
                    <Button icon={<StepForwardOutlined />} onClick={handleStep}>
                      单步推进
                    </Button>
                  </>
                ) : (
                  <Button icon={<PauseCircleOutlined />} onClick={handlePause}>
                    暂停推进
                  </Button>
                )}
              </>
            ) : null}
            {hasPrimaryActions ? (
              <Button icon={<PlayCircleOutlined />} onClick={() => setDrawerMode('actions')}>
                实验动作
              </Button>
            ) : null}
            {canShowEventLog ? (
              <Button icon={<FileSearchOutlined />} onClick={() => setDrawerMode('events')}>
                最近记录
              </Button>
            ) : null}
            <Button icon={<ReloadOutlined />} onClick={handleReset}>
              重置
            </Button>
            {canUseSnapshots ? (
              <Button icon={<SaveOutlined />} onClick={handleSaveSnapshot}>
                保存快照
              </Button>
            ) : null}
            {shouldShowSettings ? (
              <Button
                icon={<SettingOutlined />}
                type={drawerMode === 'settings' ? 'primary' : 'default'}
                onClick={() => setDrawerMode('settings')}
              >
                参数与说明
              </Button>
            ) : null}
          </Space>
        </div>
      </div>

      <div className="flex min-h-0 flex-1 flex-col overflow-hidden bg-[linear-gradient(180deg,#edf4ff_0%,#f5f8fc_100%)] p-2">
        <section className="flex min-h-0 flex-1 flex-col overflow-hidden rounded-[20px] border border-slate-200 bg-[linear-gradient(180deg,#ffffff_0%,#f8fafc_100%)] shadow-[0_8px_24px_rgba(15,23,42,0.06)]">
          <div className="flex items-center justify-between gap-3 border-b border-slate-200 px-4 py-3">
            <div>
              <div className="text-sm font-semibold text-slate-900">主舞台</div>
              <div className="text-xs text-slate-600">
                当前场景的过程变化、结构关系和结果反馈会在这里持续更新。
              </div>
            </div>
            {meta?.description ? (
              <div className="hidden max-w-[420px] text-right text-xs leading-5 text-slate-600 lg:block">
                {meta.description}
              </div>
            ) : null}
          </div>

          <div className="min-h-0 flex-1 overflow-auto bg-[radial-gradient(circle_at_top,#ffffff_0%,#f8fafc_60%,#eef4ff_100%)] p-2">
            {children}
            {renderState && state && renderState({
              ...state,
              data: {
                ...(state.data || {}),
                __events: visibleEvents,
                __params: params,
              },
            })}
          </div>
        </section>
      </div>

      <Drawer
        title={drawerTitle}
        placement="right"
        width={360}
        open={drawerMode !== null}
        onClose={() => setDrawerMode(null)}
        styles={{
          body: {
            padding: 12,
            background: '#f2f6fb',
          },
          header: {
            background: '#eef4fb',
            color: '#0f172a',
            borderBottom: '1px solid rgba(148,163,184,0.24)',
          },
        }}
      >
        <div className="flex h-full flex-col overflow-auto">
          {drawerMode === 'actions' ? (
            hasPrimaryActions ? (
              <>
                <div className="mb-3 rounded-2xl border border-slate-200 bg-white p-3 text-xs leading-6 text-slate-600">
                  通过下拉选择当前要执行的实验动作，执行后再回到主舞台观察变化。
                </div>
                <SimulationActionPanel
                  actions={normalizedActionDefinitions}
                  onExecute={handleExecuteVisualizationAction}
                />
              </>
            ) : (
              <div className="rounded-2xl border border-slate-200 bg-white p-4 text-sm text-slate-500">
                当前场景没有额外动作可执行。
              </div>
            )
          ) : null}

          {drawerMode === 'events' ? (
            canShowEventLog ? (
              <section className="rounded-2xl border border-slate-200 bg-white p-3">
                <div className="mb-2 flex items-center justify-between gap-3">
                  <div>
                    <div className="text-sm font-semibold text-slate-900">最近记录</div>
                    <div className="text-xs text-slate-600">快速查看刚刚发生了什么。</div>
                  </div>
                  <Tag color="cyan" className="m-0">
                    共 {visibleEvents.length} 条
                  </Tag>
                </div>

                {renderEvents ? (
                  renderEvents(visibleEvents)
                ) : (
                  <div className="max-h-[480px] overflow-auto pr-1 text-xs">
                    {eventPreview.map((event) => (
                      <div key={event.id} className="border-b border-slate-200 py-2 text-slate-800 last:border-b-0">
                        <div className="flex items-center justify-between gap-3">
                          <span className="font-medium text-slate-900">{getVisualizationEventTitle(event)}</span>
                          <span className="text-slate-500">步骤 {event.tick}</span>
                        </div>
                        <div className="mt-1 leading-5 text-slate-600">
                          {getVisualizationEventSummary(event)}
                        </div>
                      </div>
                    ))}
                    {visibleEvents.length === 0 ? (
                      <div className="py-4 text-center text-slate-400">当前还没有可阅读的过程记录</div>
                    ) : null}
                  </div>
                )}
              </section>
            ) : (
              <div className="rounded-2xl border border-slate-200 bg-white p-4 text-sm text-slate-500">
                当前场景没有事件记录能力。
              </div>
            )
          ) : null}

          {drawerMode === 'settings' ? (
            <>
              {canAdjustParams ? (
                <Card title="参数设置" size="small" className="mb-3 border-slate-200 bg-white">
                  {renderParams ? (
                    renderParams(params, handleSetParam)
                  ) : (
                    <div className="space-y-2.5">
                      {params.map((param) => {
                        const effectiveValue = getParamEffectiveValue(param)
                        return (
                          <div key={param.key}>
                            <label className="mb-1 block text-sm font-medium text-slate-700">{param.name}</label>
                            {(param.type === 'int' || param.type === 'float' || param.type === 'slider') && (
                              <Slider
                                min={param.min || 0}
                                max={param.max || 100}
                                step={param.type === 'float' ? 0.1 : 1}
                                value={Number(effectiveValue ?? 0)}
                                onChange={(value) => handleSetParam(param.key, value)}
                              />
                            )}
                            {param.type === 'select' && (
                              <Select
                                value={effectiveValue}
                                onChange={(value) => handleSetParam(param.key, value)}
                                options={param.options}
                                className="w-full"
                              />
                            )}
                            {param.type === 'bool' && (
                              <div className="flex items-center justify-between rounded border border-slate-200 bg-slate-50 px-3 py-2">
                                <span className="text-xs text-slate-700">
                                  {Boolean(effectiveValue) ? '已开启' : '已关闭'}
                                </span>
                                <Switch
                                  checked={Boolean(effectiveValue)}
                                  onChange={(checked) => handleSetParam(param.key, checked)}
                                />
                              </div>
                            )}
                          </div>
                        )
                      })}
                    </div>
                  )}
                </Card>
              ) : null}

              {meta ? (
                <Card title="场景说明" size="small" className="mb-2 border-slate-200 bg-white">
                  <div className="mb-3 text-sm leading-6 text-slate-700">
                    {meta.description}
                  </div>
                  <Descriptions column={1} size="small">
                    <Descriptions.Item label="当前主题">{getVisualizationModuleLabel(module, meta.name)}</Descriptions.Item>
                    <Descriptions.Item label="类型">{getVisualizationComponentTypeLabel(meta.type)}</Descriptions.Item>
                    <Descriptions.Item label="支持能力">
                      {meta.capabilities.map((capability) => getVisualizationCapabilityLabel(capability)).join('、')}
                    </Descriptions.Item>
                  </Descriptions>
                </Card>
              ) : null}
            </>
          ) : null}
        </div>
      </Drawer>
    </div>
  )
}
