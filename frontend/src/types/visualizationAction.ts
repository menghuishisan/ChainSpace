/**
 * 可视化沙箱动作类型定义。
 * 这一层只描述“前端要渲染什么动作”和“动作执行需要什么参数”，
 * 具体执行仍由 SimulationWrapper 和 simulations 服务负责。
 */

export type VisualizationActionKind = 'module_action' | 'inject_fault' | 'inject_attack'

export type VisualizationActionScope =
  | 'consensus'
  | 'attack'
  | 'blockchain'
  | 'network'
  | 'crypto'
  | 'crosschain'
  | 'evm'
  | 'defi'

export interface VisualizationActionOption {
  label: string
  value: string
}

export interface VisualizationActionField {
  key: string
  label: string
  type: 'select' | 'number' | 'text'
  defaultValue?: string | number
  min?: number
  max?: number
  options?: VisualizationActionOption[]
}

export interface VisualizationActionDefinition {
  key: string
  label: string
  description: string
  kind: VisualizationActionKind
  group?: string
  scope?: VisualizationActionScope
  linkedScopes?: VisualizationActionScope[]
  overlayLabel?: string
  action?: string
  preset?: Record<string, unknown>
  fields?: VisualizationActionField[]
  successMessage: string
}
