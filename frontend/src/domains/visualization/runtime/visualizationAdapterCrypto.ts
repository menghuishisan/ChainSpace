import type {
  SimulatorState,
  VisualizationEntityCard,
  VisualizationMetricCard,
  VisualizationRecord,
  VisualizationStateSection,
} from '@/types/visualizationDomain'
import {
  asArray,
  asNumber,
  asRecord,
  asString,
  buildEventItems,
  formatNumber,
  getEvents,
  getGlobalData,
} from './visualizationAdapterCommon'

export function buildCryptoVisualizationData(state: SimulatorState): VisualizationRecord {
  const globalData = getGlobalData(state)
  const events = getEvents(state)
  const recentHistory = asArray<VisualizationRecord>(globalData.recent_history)
  const avalanche = asRecord(globalData.avalanche)
  const latestHistory = recentHistory[recentHistory.length - 1]
  const latestVerify = events.slice().reverse().find((event) => event.type === 'integrity_verified')
  const verifyPayload = latestVerify ? asRecord(latestVerify.data) : {}

  const metrics: VisualizationMetricCard[] = [
    { key: 'history_count', label: '累计计算次数', value: String(asNumber(globalData.history_count)), hint: '表示本轮已经完成多少次摘要计算。' },
    { key: 'latest_algorithm', label: '最近一次算法', value: asString(latestHistory?.algorithm, '未执行'), hint: '用于帮助学生确认当前结果来自哪种哈希算法。' },
    { key: 'latest_bits', label: '输出位长度', value: latestHistory ? `${asNumber(latestHistory.bit_length)} 位` : '--', hint: '不同算法的输出位长度不同。' },
    { key: 'integrity_result', label: '完整性校验', value: latestVerify ? (verifyPayload.valid ? '通过' : '失败') : '未执行', hint: '重点观察输入是否被篡改后仍能通过校验。' },
  ]

  const entities: VisualizationEntityCard[] = []
  if (latestHistory) {
    entities.push({
      id: 'latest-hash',
      title: '最近一次摘要结果',
      subtitle: '先核对输入、算法和输出是否一致',
      status: asString(latestHistory.algorithm, '未执行'),
      details: [
        { label: '输入文本', value: asString(latestHistory.input, '--') },
        { label: '输入十六进制', value: asString(latestHistory.input_hex, '--') },
        { label: '摘要输出', value: asString(latestHistory.output, '--') },
        { label: '输出位长度', value: `${asNumber(latestHistory.bit_length)} 位` },
      ],
    })
  }
  if (Object.keys(avalanche).length > 0) {
    entities.push({
      id: 'avalanche',
      title: '雪崩效应对照',
      subtitle: '输入轻微变化后，输出应大范围改变',
      status: `${formatNumber(asNumber(avalanche.change_percentage), 2)}% 改变`,
      details: [
        { label: '原始输入', value: asString(avalanche.original_input, '--') },
        { label: '修改后输入', value: asString(avalanche.modified_input, '--') },
        { label: '原始摘要', value: asString(avalanche.original_hash, '--') },
        { label: '修改后摘要', value: asString(avalanche.modified_hash, '--') },
      ],
    })
  }

  const sections: VisualizationStateSection[] = [
    {
      key: 'supported_algorithms',
      title: '算法族与输出长度',
      items: Object.entries(asRecord(globalData.algorithms)).map(([key, value]) => ({
        label: key.toUpperCase(),
        value: `${asNumber(value)} 位`,
      })),
    },
  ]
  if (Object.keys(avalanche).length > 0) {
    sections.push({
      key: 'avalanche_result',
      title: '雪崩效应结果',
      items: [
        { label: '不同位数', value: `${asNumber(avalanche.different_bits)} 位` },
        { label: '总位数', value: `${asNumber(avalanche.total_bits)} 位` },
        { label: '变化比例', value: `${formatNumber(asNumber(avalanche.change_percentage), 2)}%` },
      ],
    })
  }
  if (latestVerify) {
    sections.push({
      key: 'integrity',
      title: '完整性校验结果',
      items: [
        { label: '校验算法', value: asString(verifyPayload.algorithm, '--') },
        { label: '输入文本', value: asString(verifyPayload.input, '--') },
        { label: '期望摘要', value: asString(verifyPayload.expected_hash, '--') },
        { label: '校验结果', value: verifyPayload.valid ? '通过' : '失败' },
      ],
    })
  }

  return {
    title: '密码学演示：哈希算法',
    summary: '重点观察同一输入在不同算法下的摘要差异，以及输入轻微变化后输出是否发生大范围变化。只有摘要一致时，完整性校验才应该通过。',
    metrics,
    entities,
    sections,
    events: buildEventItems(events),
    observationTips: [
      '先看最近一次摘要结果，确认输入、算法和输出是否彼此对应。',
      '做多算法对比时，重点关注输出位长度和摘要内容差异，而不是只看是否成功执行。',
      '雪崩效应和完整性校验都要看结果是否合理，不能只看按钮是否执行成功。',
    ],
  }
}
