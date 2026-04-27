import type {
  NodeState,
  SimulatorEvent,
  SimulatorState,
  VisualizationDisturbanceItem,
  VisualizationEntityCard,
  VisualizationEventItem,
  VisualizationMetricCard,
  VisualizationRuntimeSpec,
  VisualizationStateSection,
} from '@/types/visualizationDomain'
import type { VisualizationRecord } from '@/types/visualizationDomain'
import { getVisualizationEventSummary } from './visualizationEventFormatter'
import { getVisualizationEventLabel, isSystemVisualizationEvent } from './visualizationEventLabels'
import { getVisualizationModuleOption } from './visualizationRegistry'

export function asRecord(value: unknown): VisualizationRecord {
  return value && typeof value === 'object' && !Array.isArray(value) ? (value as VisualizationRecord) : {}
}

export function asArray<T>(value: unknown): T[] {
  return Array.isArray(value) ? (value as T[]) : []
}

export function asNumber(value: unknown, fallback = 0): number {
  if (typeof value === 'number' && Number.isFinite(value)) {
    return value
  }
  if (typeof value === 'string') {
    const parsed = Number(value.replace(/%/g, '').trim())
    if (Number.isFinite(parsed)) {
      return parsed
    }
  }
  return fallback
}

export function asString(value: unknown, fallback = ''): string {
  if (typeof value === 'string') {
    return value
  }
  return value === null || value === undefined ? fallback : String(value)
}

export function stringifyValue(value: unknown, fallback = '--'): string {
  if (value === null || value === undefined) {
    return fallback
  }
  if (typeof value === 'string') {
    return value
  }
  if (typeof value === 'number' || typeof value === 'boolean') {
    return String(value)
  }
  try {
    return JSON.stringify(value)
  } catch {
    return fallback
  }
}

export function shortHash(value: unknown, fallback = '--'): string {
  const raw = asString(value, fallback)
  return raw.length <= 18 ? raw : `${raw.slice(0, 8)}...${raw.slice(-6)}`
}

export function formatNumber(value: number, digits = 2): string {
  return Number.isFinite(value)
    ? value.toLocaleString('zh-CN', { minimumFractionDigits: 0, maximumFractionDigits: digits })
    : '0'
}

export function getGlobalData(state: SimulatorState): VisualizationRecord {
  return asRecord(state.global_data)
}

export function getData(state: SimulatorState): VisualizationRecord {
  return asRecord(state.data)
}

export function getNodes(state: SimulatorState): NodeState[] {
  if (!state.nodes) {
    return []
  }
  return Array.isArray(state.nodes) ? (state.nodes as unknown as NodeState[]) : Object.values(state.nodes)
}

export function getEvents(state: SimulatorState): SimulatorEvent[] {
  return asArray<SimulatorEvent>(getData(state).__events)
    .filter((event) => !isSystemVisualizationEvent(event.type))
    .slice(-32)
}

export function buildMetrics(
  state: SimulatorState,
  runtime: VisualizationRuntimeSpec,
  globalData: VisualizationRecord,
  events: SimulatorEvent[],
  nodes: NodeState[],
): VisualizationMetricCard[] {
  const moduleLabel = getVisualizationModuleOption(runtime.moduleKey)?.label || '当前场景'

  return [
    {
      key: 'tick',
      label: '当前步骤',
      value: String(state.tick),
      hint: '表示当前可视化已经推进到的教学步骤。',
    },
    {
      key: 'topic',
      label: '当前主题',
      value: moduleLabel,
      hint: '这里显示当前实验对应的学习主题。',
    },
    {
      key: 'objects',
      label: '活跃对象',
      value: nodes.length > 0 ? `${nodes.filter((node) => node.status !== 'offline').length}/${nodes.length}` : '0',
      hint: '当前在线或可用的节点、对象数量。',
    },
    {
      key: 'events',
      label: '最近事件',
      value: String(events.length),
      hint: '最近一段时间内可供观察的事件数量。',
    },
    {
      key: 'fields',
      label: '状态字段',
      value: String(Object.keys(globalData).filter((key) => !key.startsWith('__')).length),
      hint: '当前场景状态中可供观察的关键字段数。',
    },
  ]
}

export function buildEntities(nodes: NodeState[], fallbackPrefix: string): VisualizationEntityCard[] {
  return nodes.slice(0, 12).map((node, index) => {
    const nodeData = asRecord(node.data)
    return {
      id: node.id,
      title: asString(nodeData.name, node.id),
      subtitle: asString(nodeData.role || nodeData.type, `${fallbackPrefix} ${index + 1}`),
      status: node.status,
      details: Object.entries(nodeData).slice(0, 5).map(([key, value]) => ({
        label: key,
        value: typeof value === 'object' ? stringifyValue(value) : asString(value, '--'),
      })),
    }
  })
}

export function buildSections(state: SimulatorState, globalData: VisualizationRecord): VisualizationStateSection[] {
  const toItems = (record: VisualizationRecord) => Object.entries(record)
    .filter(([key]) => !key.startsWith('__'))
    .slice(0, 10)
    .map(([key, value]) => ({
      label: key,
      value: stringifyValue(value, '--'),
    }))

  const sections: VisualizationStateSection[] = []
  const globalItems = toItems(globalData)
  const dataItems = toItems(getData(state))

  if (globalItems.length > 0) {
    sections.push({ key: 'global', title: '全局状态', items: globalItems })
  }
  if (dataItems.length > 0) {
    sections.push({ key: 'data', title: '当前状态数据', items: dataItems })
  }

  const processFeedback = asRecord(globalData.process_feedback)
  const processItems = [
    {
      label: '当前阶段',
      value: asString(processFeedback.stage, '--'),
    },
    {
      label: '过程摘要',
      value: asString(processFeedback.summary, '--'),
    },
    {
      label: '下一步提示',
      value: asString(processFeedback.next_hint, '--'),
    },
    {
      label: '进度',
      value: processFeedback.progress === undefined ? '--' : `${asNumber(processFeedback.progress).toFixed(0)}%`,
    },
  ].filter((item) => item.value !== '--')

  if (processItems.length > 0) {
    sections.push({ key: 'process_feedback', title: '过程反馈', items: processItems })
  }

  const linkedEffects = asArray<VisualizationRecord>(globalData.linked_effects)
  const linkedEffectItems = linkedEffects.slice(0, 8).map((effect) => ({
    label: asString(effect.scope, '联动影响'),
    value: asString(effect.summary, '--'),
  })).filter((item) => item.value !== '--')

  if (linkedEffectItems.length > 0) {
    sections.push({ key: 'linked_effects', title: '联动影响', items: linkedEffectItems })
  }

  return sections
}

export function buildEventItems(events: SimulatorEvent[]): VisualizationEventItem[] {
  return events.slice(-8).reverse().map((event, index) => ({
    id: event.id || `event-${index}`,
    title: getVisualizationEventLabel(event.type),
    summary: getVisualizationEventSummary(event),
    tick: event.tick,
    source: event.source,
    target: event.target,
  }))
}

export function buildDisturbances(globalData: VisualizationRecord): VisualizationDisturbanceItem[] {
  const linkedEffects = asArray<VisualizationRecord>(globalData.linked_effects).map<VisualizationDisturbanceItem>((effect, index) => ({
    id: asString(effect.id, `effect-${index}`),
    type: asString(effect.scope, 'effect'),
    target: asString(effect.target, '--'),
    label: asString(effect.blocking) === 'true' || effect.blocking === true ? '阻断性联动' : '联动影响',
    summary: asString(effect.summary, '当前联动正在影响主过程。'),
  }))

  const faults = asArray<VisualizationRecord>(globalData.active_faults).map<VisualizationDisturbanceItem>((fault, index) => ({
    id: asString(fault.id, `fault-${index}`),
    type: asString(fault.type, 'fault'),
    target: asString(fault.target, '--'),
    label: '故障覆盖层',
    summary: `故障类型 ${asString(fault.type, '--')}，作用目标 ${asString(fault.target, '--')}。`,
  }))

  const attacks = asArray<VisualizationRecord>(globalData.active_attacks).map<VisualizationDisturbanceItem>((attack, index) => ({
    id: asString(attack.id, `attack-${index}`),
    type: asString(attack.type, 'attack'),
    target: asString(attack.target, '--'),
    label: '攻击覆盖层',
    summary: `攻击类型 ${asString(attack.type, '--')}，作用目标 ${asString(attack.target, '--')}。`,
  }))

  return [...linkedEffects, ...faults, ...attacks]
}
