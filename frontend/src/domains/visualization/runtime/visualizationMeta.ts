import type {
  Capability,
  ComponentType,
  SimulatorDescription,
  SimulatorStatus,
} from '@/types/visualizationDomain'
import { getVisualizationModuleOption } from './visualizationRegistry'

const COMPONENT_TYPE_LABELS: Record<ComponentType, string> = {
  tool: '工具演示',
  demo: '结构演示',
  process: '过程模拟',
  attack: '攻击演示',
  defi: 'DeFi 机制',
}

const CAPABILITY_LABELS: Record<Capability, string> = {
  param_panel: '可调参数',
  time_control: '过程控制',
  state_monitor: '状态观察',
  event_log: '事件记录',
  snapshot: '场景快照',
}

const STATUS_LABELS: Record<SimulatorStatus, string> = {
  idle: '等待开始',
  running: '运行中',
  paused: '已暂停',
  stopped: '已停止',
  error: '异常',
}

export function getVisualizationComponentTypeLabel(type?: ComponentType): string {
  if (!type) {
    return '可视化实验'
  }
  return COMPONENT_TYPE_LABELS[type] || '可视化实验'
}

export function getVisualizationCapabilityLabel(capability: Capability): string {
  return CAPABILITY_LABELS[capability] || capability
}

export function getVisualizationStatusLabel(status?: SimulatorStatus): string {
  if (!status) {
    return '等待开始'
  }
  return STATUS_LABELS[status] || '等待开始'
}

export function getVisualizationModuleLabel(moduleKey?: string, fallback = '当前场景'): string {
  return getVisualizationModuleOption(moduleKey)?.label || fallback
}

export function hasVisualizationCapability(
  meta: Pick<SimulatorDescription, 'capabilities'> | null | undefined,
  capability: Capability,
): boolean {
  return Boolean(meta?.capabilities?.includes(capability))
}
