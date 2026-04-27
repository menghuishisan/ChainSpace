import type {
  AttackStats,
  AttackSummary,
  AttackTimelineItem,
  BalanceCard,
  CallFrame,
  SimulatorState,
  StorageSlot,
  VisualizationRuntimeSpec,
} from '@/types/visualizationDomain'
import type { VisualizationRecord } from '@/types/visualizationDomain'
import { getAttackMechanism } from './visualizationAttack'
import {
  asArray,
  asNumber,
  asString,
  buildDisturbances,
  buildEntities,
  buildEventItems,
  buildMetrics,
  buildSections,
  formatNumber,
  getEvents,
  getGlobalData,
  getNodes,
} from './visualizationAdapterCommon'

export function buildAttackVisualizationData(
  state: SimulatorState,
  runtime: VisualizationRuntimeSpec,
): VisualizationRecord {
  const globalData = getGlobalData(state)
  const events = getEvents(state)
  const disturbances = buildDisturbances(globalData)
  const steps = asArray<VisualizationRecord>(globalData.steps)
  const metrics = buildMetrics(state, runtime, globalData, events, getNodes(state))
  const actors = buildEntities(getNodes(state), '攻击参与方')
  const sections = buildSections(state, globalData)
  const eventItems = buildEventItems(events)
  const victimBalance = asNumber(globalData.victim_balance)
  const attackerBalance = asNumber(globalData.attacker_balance)
  const mechanism = getAttackMechanism(runtime)
  const hasSteps = steps.length > 0

  const attackSummaryMap: Record<typeof mechanism, string> = {
    execution: hasSteps
      ? '当前已经形成清晰的调用链。建议沿着每一层调用观察漏洞是在状态更新之前暴露，还是被修复逻辑拦截。'
      : '当前还没有形成完整的执行攻击路径。先执行攻击动作，再观察调用链和关键状态变化。',
    economic: hasSteps
      ? '当前已经形成清晰的资金操纵路径。重点比较受害方资产下降、攻击者收益增长以及价格或清算指标的扭曲程度。'
      : '当前还没有形成经济攻击路径。先执行攻击动作，再观察资金流向、价格冲击和收益变化。',
    consensus: hasSteps
      ? '当前已经形成链竞争或投票破坏路径。重点确认攻击链如何取得优势，以及诚实链在哪一步失去主导。'
      : '当前还没有形成链安全攻击路径。先执行攻击动作，再观察分叉、私有链推进或投票失衡。',
    bridge: hasSteps
      ? '当前已经形成跨链攻击路径。重点确认源链、桥验证层和目标链之间，哪一步被伪造、绕过或重复利用。'
      : '当前还没有形成跨链攻击路径。先执行攻击动作，再观察消息如何穿过桥验证层并错误落到目标链。',
  }

  const observationTipsMap: Record<typeof mechanism, string[]> = {
    execution: [
      '先看调用链顺序，再判断外部交互发生在状态更新之前还是之后。',
      '重点比较受害状态槽位、权限标记或余额是否在危险步骤后立刻变化。',
      '如果切到修复模式，应该能明显看到关键路径被阻断，而不是只换一段文字说明。',
    ],
    economic: [
      '先确认受害方和攻击方资产如何变化，再判断价格、滑点、清算或收益率是否被扭曲。',
      '时间线不应只停在一次操作成功，而要能看到利润如何逐步累积。',
      '切换修复或防御策略后，资金路径和收益结果都应当明显收敛。',
    ],
    consensus: [
      '先看诚实链和攻击链如何分叉，再判断攻击者如何逐步取得优势。',
      '如果是投票类攻击，要观察票数、权重或签名阈值在哪一步被破坏。',
      '最终结果必须能看出哪条链成为主导，或系统为什么没能达成安全共识。',
    ],
    bridge: [
      '先确认源链、桥验证层和目标链的流程顺序，再看哪一步被伪造或绕过。',
      '只有看到签名、证明或挑战期被错误满足，才能说明这次攻击真正成立。',
      '最终要能在目标链看到错误执行、错误铸造或重复提取，而不是只停在源链事件。',
    ],
  }

  return {
    attack: {
      title: '攻击路径演示',
      summary: attackSummaryMap[mechanism],
      depth: asNumber(globalData.max_depth, steps.length),
      completedSteps: steps.length,
    } satisfies AttackSummary,
    balances: runtime.moduleKey === 'attacks/reentrancy' || victimBalance > 0 || attackerBalance > 0
      ? [
        {
          key: 'victim',
          label: '受害方余额',
          value: victimBalance,
          color: '#38bdf8',
          description: '如果这里持续下降，通常说明攻击路径仍在造成损失。',
        },
        {
          key: 'attacker',
          label: '攻击方收益',
          value: attackerBalance,
          color: '#f97316',
          description: '如果这里持续上升，通常说明攻击已经在兑现利润。',
        },
      ] satisfies BalanceCard[]
      : [],
    timeline: steps.map<AttackTimelineItem>((step, index) => ({
      id: `step-${index}`,
      title: asString(step.action, `步骤 ${index + 1}`),
      description: asString(step.description, '执行一次攻击步骤。'),
      amount: asString(step.amount, '0'),
      depth: asNumber(step.call_depth, index + 1),
    })),
    callFrames: steps.map<CallFrame>((step, index) => ({
      id: `frame-${index}`,
      title: asString(step.action, `步骤 ${index + 1}`),
      description: asString(step.description, '执行一次攻击步骤。'),
      caller: asString(step.caller, '攻击者'),
      callee: asString(step.target ?? step.function, '目标对象'),
      amount: asString(step.amount, '0'),
      depth: asNumber(step.call_depth, index + 1),
    })),
    storage: [
      { key: 'victim', label: '受害方余额', value: formatNumber(victimBalance, 2) },
      { key: 'attacker', label: '攻击方收益', value: formatNumber(attackerBalance, 2) },
      { key: 'depth', label: '攻击深度', value: String(asNumber(globalData.max_depth, steps.length)) },
    ] satisfies StorageSlot[],
    stats: {
      steps: steps.length,
      attackDepth: asNumber(globalData.max_depth, steps.length),
      drainedAmount: Math.max(attackerBalance, 0),
      remainingBalance: Math.max(victimBalance, 0),
    } satisfies AttackStats,
    metrics,
    actors,
    sections,
    events: eventItems,
    disturbances,
    observationTips: observationTipsMap[mechanism],
  }
}
