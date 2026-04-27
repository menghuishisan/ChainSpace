import type {
  SimulatorEvent,
  SimulatorParam,
} from './simulation'

/**
 * simulations 服务 HTTP 返回包的统一结构。
 * 前端通过该类型统一校验业务 code，避免只看 HTTP 状态码导致误判。
 */
export interface SimulationResponse<T> {
  code: number
  message: string
  data: T
}

/**
 * 模块动作执行后的返回结果。
 * 用于承载教学动作、攻击注入或故障注入的执行反馈。
 */
export interface SimulatorActionResponse {
  success: boolean
  message?: string
  data?: Record<string, unknown>
}

/**
 * 事件查询接口返回的数据结构。
 */
export interface SimulatorEventsPayload {
  events?: SimulatorEvent[]
}

/**
 * 导出模拟器状态时的返回结果。
 */
export interface SimulatorStateExportPayload {
  state?: Record<string, unknown>
}

/**
 * 模拟器请求参数对象。
 * 统一抽离后，API 实现文件只负责通信，不再内嵌接口定义。
 */
export interface SimulationRequestOptions {
  method?: string
  body?: Record<string, unknown>
}

/**
 * 模拟器参数接口兼容数组和键值映射两种返回形态。
 */
export type SimulatorParamsPayload = Record<string, SimulatorParam> | SimulatorParam[]
