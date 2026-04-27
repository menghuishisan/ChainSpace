/**
 * simulations 服务相关的前端类型定义。
 */

/** 模拟器运行状态。 */
export type SimulatorStatus = 'idle' | 'running' | 'paused' | 'stopped' | 'error'

/** 模拟器状态快照。 */
export interface SimulatorState {
  tick: number
  status: SimulatorStatus
  nodes?: Record<string, NodeState>
  global_data?: Record<string, unknown>
  data?: Record<string, unknown>
  updated_at?: string
}

/** 单个节点状态。 */
export interface NodeState {
  id: string
  status: string
  data?: Record<string, unknown>
  is_byzantine?: boolean
}

/** 区块状态。 */
export interface BlockState {
  index: number
  hash: string
  previousHash: string
  timestamp: number
  transactions: number
  miner?: string
}

/** 交易状态。 */
export interface TransactionState {
  id: string
  from: string
  to: string
  amount: number
  status: 'pending' | 'confirmed' | 'failed'
}

/** 消息状态。 */
export interface MessageState {
  id: string
  type: string
  from: string
  to: string
  timestamp: number
  data?: Record<string, unknown>
}

/** 模拟器事件。 */
export interface SimulatorEvent {
  id: string
  type: string
  tick: number
  timestamp: string
  source?: string
  target?: string
  data: Record<string, unknown>
}

/** 参数类型。 */
export type ParamType = 'int' | 'float' | 'string' | 'bool' | 'select' | 'slider'

/** 参数选项。 */
export interface ParamOption {
  label: string
  value: unknown
}

/** 模拟器参数定义。 */
export interface SimulatorParam {
  key: string
  name: string
  description?: string
  type: ParamType
  default?: unknown
  min?: number
  max?: number
  options?: ParamOption[]
  value: unknown
}

/** 组件类型。 */
export type ComponentType = 'tool' | 'demo' | 'process' | 'attack' | 'defi'

/** 组件能力。 */
export type Capability = 'param_panel' | 'time_control' | 'state_monitor' | 'event_log' | 'snapshot'

/** 模拟器描述信息。 */
export interface SimulatorDescription {
  id: string
  name: string
  description: string
  category: string
  type: ComponentType
  capabilities: Capability[]
  params: SimulatorParam[]
  version?: string
}

/** 故障注入配置。 */
export interface FaultConfig {
  type: string
  target: string
  params?: Record<string, unknown>
  duration?: number
}

/** 攻击注入配置。 */
export interface AttackConfig {
  type: string
  target: string
  params?: Record<string, unknown>
  duration?: number
}

/** 快照信息。 */
export interface SnapshotInfo {
  id: string
  name: string
  tick: number
  created_at: string
  size: number
}

/** WebSocket 返回消息。 */
export interface WSMessage {
  type: 'state_update' | 'event' | 'error' | 'status' | 'param_set' | 'speed_set' | 'fault_injected' | 'attack_injected'
  data: unknown
}

/** WebSocket 控制命令。 */
export interface WSCommand {
  action: string
  params?: Record<string, unknown>
}
