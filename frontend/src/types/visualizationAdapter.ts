/**
 * 可视化适配层使用的中间类型。
 * 这一层只服务于前端状态转换，不直接暴露给业务表单或 API。
 */
export type VisualizationRecord = Record<string, unknown>

/**
 * 区块字段在适配前的原始结构。
 * 区块结构模拟器会返回 name/value 形式的字段列表，
 * 适配层会把它转换成画布可直接消费的数据。
 */
export interface RawBlockchainField {
  name?: string
  value?: unknown
}
