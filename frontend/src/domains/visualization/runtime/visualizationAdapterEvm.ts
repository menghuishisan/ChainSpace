import type {
  EVMExecutionStats,
  EVMFrame,
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
  getEvents,
  getGlobalData,
  shortHash,
} from './visualizationAdapterCommon'

export function buildEvmVisualizationData(state: SimulatorState): VisualizationRecord {
  const globalData = getGlobalData(state)
  const events = getEvents(state)
  const framesRaw = asArray<VisualizationRecord>(globalData.frames)
  const traceTree = asArray<string>(globalData.trace_tree)

  const frames = framesRaw.map<EVMFrame>((frame, index) => ({
    id: `frame-${index}`,
    title: `${asString(frame.call_type, asString(frame.title, `执行帧 ${index + 1}`))} @ 深度 ${asNumber(frame.depth, index + 1)}`,
    opcode: asString(frame.opcode),
    description: `${asString(frame.from, '--')} -> ${asString(frame.to, '--')}${asString(frame.value) && asString(frame.value) !== '0' ? `，附带 ${asString(frame.value)} wei` : ''}`,
    depth: asNumber(frame.depth, index + 1),
    gas: asString(frame.gas),
  }))

  const metrics: VisualizationMetricCard[] = [
    { key: 'scenario', label: '当前场景', value: asString(globalData.scenario, '未生成'), hint: '说明当前调用链模拟的是哪类执行路径。' },
    { key: 'frame_count', label: '执行帧数量', value: String(frames.length), hint: '调用越复杂，执行帧通常越多。' },
    { key: 'max_depth', label: '最大调用深度', value: String(asNumber(globalData.max_depth)), hint: '用于观察调用是否出现深层嵌套。' },
    { key: 'total_gas', label: '累计 Gas', value: String(asNumber(globalData.total_gas)), hint: '用于帮助学生比较不同场景的执行成本。' },
  ]

  const sections: VisualizationStateSection[] = []
  if (traceTree.length > 0) {
    sections.push({
      key: 'trace_tree',
      title: '调用树摘要',
      items: traceTree.slice(0, 8).map((line, index) => ({
        label: `调用 ${index + 1}`,
        value: line,
      })),
    })
  }
  if (framesRaw.length > 0) {
    const root = framesRaw[0]
    sections.push({
      key: 'root_call',
      title: '入口调用检查',
      items: [
        { label: '调用类型', value: asString(root.call_type, '--') },
        { label: '调用方向', value: `${asString(root.from, '--')} -> ${asString(root.to, '--')}` },
        { label: 'Gas 使用', value: asString(root.gas, '--') },
        { label: '日志条数', value: String(asNumber(root.log_count)) },
      ],
    })
  }

  const entities = framesRaw.slice(0, 4).map<VisualizationEntityCard>((frame, index) => ({
    id: `evm-entity-${index}`,
    title: `${asString(frame.call_type, 'CALL')} 层 ${asNumber(frame.depth)}`,
    subtitle: `${asString(frame.from, '--')} -> ${asString(frame.to, '--')}`,
    status: asString(frame.error, '') ? 'error' : 'ok',
    details: [
      { label: '输入', value: shortHash(frame.input, '--') },
      { label: '输出', value: shortHash(frame.output, '--') },
      { label: 'Gas', value: asString(frame.gas, '--') },
      { label: '状态变化', value: String(Object.keys(asRecord(frame.storage_changes)).length) },
    ],
  }))

  return {
    title: 'EVM 执行演示',
    summary: '学生应当能沿着调用链确认谁调用了谁、每层消耗了多少 Gas、哪些步骤真的修改了状态，而不是只看到一串调用名称。',
    metrics,
    entities,
    sections,
    events: buildEventItems(events),
    observationTips: [
      '先看入口调用和最大深度，确认这条链路是不是你预期的执行路径。',
      '再沿执行帧向下看，判断每一层是谁调用了谁，以及 Gas 是如何被消耗的。',
      '如果某层没有输出或出现错误，应结合调用树判断是在哪一跳出问题。',
    ],
    frames,
    stats: {
      frameCount: frames.length,
      opcodeCount: frames.filter((frame) => Boolean(frame.opcode)).length,
      storageChanges: framesRaw.reduce((sum, frame) => sum + Object.keys(asRecord(frame.storage_changes)).length, 0),
    } satisfies EVMExecutionStats,
  }
}
