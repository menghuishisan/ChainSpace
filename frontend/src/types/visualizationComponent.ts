import type { ReactNode } from 'react'
import type {
  SimulatorEvent,
  SimulatorParam,
  SimulatorState,
} from './simulation'
import type { VisualizationActionDefinition } from './visualizationAction'
import type { VisualizationDisturbanceItem } from './visualizationCanvas'

/**
 * SimulationWrapper 的组件输入定义。
 * 这里统一描述可视化沙箱的状态渲染、事件渲染和动作面板接线方式。
 */
export interface SimulationWrapperProps {
  module: string
  initialParams?: Record<string, unknown>
  nodeCount?: number
  renderState?: (state: SimulatorState) => ReactNode
  renderEvents?: (events: SimulatorEvent[]) => ReactNode
  renderParams?: (
    params: SimulatorParam[],
    onSetParam: (key: string, value: unknown) => void,
  ) => ReactNode
  actionDefinitions?: VisualizationActionDefinition[]
  onStateUpdate?: (state: SimulatorState) => void
  onEvent?: (event: SimulatorEvent) => void
  onConnectionChange?: (connected: boolean) => void
  accessUrl?: string
  wsUrl?: string
  children?: ReactNode
}

/**
 * SimulationActionPanel 的组件输入定义。
 * 动作面板只负责收集参数和触发执行，不直接关心具体模拟器实现。
 */
export interface SimulationActionPanelProps {
  actions: VisualizationActionDefinition[]
  onExecute: (
    action: VisualizationActionDefinition,
    values: Record<string, unknown>,
  ) => Promise<void>
}

export interface VisualizationSummaryMetricProps {
  label: string
  value: string
  accentClass: string
}

export interface VisualizationSummaryCardProps {
  title: string
  value: string
  hint: string
  valueClassName: string
}

export interface VisualizationContractCardProps {
  title: string
  tagColor: string
  tagLabel: string
  value: number
  accentClass: string
  description: string
  tips: string[]
}

export interface VisualizationDisturbancePanelProps {
  title?: string
  items: VisualizationDisturbanceItem[]
}
