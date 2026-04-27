import type { SimulatorEvent } from '@/types/visualizationDomain'
import { getVisualizationEventLabel } from './visualizationEventLabels'

function asRecord(value: unknown): Record<string, unknown> {
  if (!value || typeof value !== 'object' || Array.isArray(value)) {
    return {}
  }

  return value as Record<string, unknown>
}

function asString(value: unknown, fallback = ''): string {
  if (typeof value === 'string') {
    return value
  }

  if (value === null || value === undefined) {
    return fallback
  }

  return String(value)
}

function asNumber(value: unknown, fallback = 0): number {
  if (typeof value === 'number' && Number.isFinite(value)) {
    return value
  }

  if (typeof value === 'string') {
    const parsed = Number(value)
    if (Number.isFinite(parsed)) {
      return parsed
    }
  }

  return fallback
}

function shortValue(value: unknown, fallback = '--'): string {
  const raw = asString(value, fallback)
  if (raw.length <= 18) {
    return raw
  }

  return `${raw.slice(0, 8)}...${raw.slice(-6)}`
}

/**
 * 为事件日志生成统一标题，避免页面直接暴露底层事件枚举值。
 */
export function getVisualizationEventTitle(event: SimulatorEvent): string {
  return getVisualizationEventLabel(event.type)
}

/**
 * 为事件日志生成可阅读的中文摘要。
 * 这里统一吸收 simulations 暴露的 source、target 和 data 字段，保证各类模块都能有基本可读性。
 */
export function getVisualizationEventSummary(event: SimulatorEvent): string {
  const payload = asRecord(event.data)
  const source = asString(event.source, '系统')
  const target = asString(event.target, '')
  const sequence = asNumber(payload.sequence, -1)
  const digest = shortValue(payload.digest, '')

  switch (event.type) {
    case 'client_request':
      return `客户端向 ${target || '主节点'} 发起序号 ${sequence >= 0 ? sequence : '--'} 的请求`
    case 'pre_prepare':
      return `${source} 广播序号 ${sequence >= 0 ? sequence : '--'} 的预准备消息`
    case 'prepare':
      return `${source} 确认摘要 ${digest || '--'}，为序号 ${sequence >= 0 ? sequence : '--'} 投出准备票`
    case 'commit':
      return `${source} 为序号 ${sequence >= 0 ? sequence : '--'} 广播提交确认`
    case 'committed':
      return `${source} 完成序号 ${sequence >= 0 ? sequence : '--'} 的提交并向客户端返回结果`
    case 'view_change':
      return `系统切换到视图 ${asNumber(payload.new_view, 0)}，新主节点为 ${asString(payload.new_primary, '--')}`
    case 'fault_injected':
      return `${source} 被注入故障，行为为 ${asString(payload.behavior, asString(payload.fault_type, '未知'))}`
    case 'request_vote':
      return `${source} 发起新一轮选举，请求其他节点投票`
    case 'vote_response':
      return `${source} 向 ${target || '候选节点'} 返回投票结果`
    case 'leader_elected':
      return `${source || '系统'} 完成本轮选举，新的领导者已经产生`
    case 'append_entries':
      return `${source} 向其他节点复制日志并维持心跳`
    case 'command_submitted':
      return `客户端向 ${target || source} 提交新的状态机命令`
    case 'attack_started':
      return `${source} 开始执行攻击流程`
    case 'attack_completed':
      return `${source} 完成攻击流程，结果已更新`
    case 'swap_executed':
      return `完成一次兑换，价格影响 ${asNumber(payload.price_impact_percent, 0).toFixed(2)}%`
    case 'liquidity_added':
      return `向资金池添加流动性，池子状态已更新`
    case 'liquidity_removed':
      return `从资金池移除流动性，池子状态已更新`
    case 'field_changed':
      return `已修改字段 ${asString(payload.field, '--')}，区块哈希重新计算`
    case 'transaction_added':
      return `当前区块新增一笔交易 ${shortValue(payload.tx_hash, shortValue(payload.hash, '--'))}`
    case 'block_mined':
      return `找到满足难度条件的新区块 ${shortValue(payload.hash, '--')}`
    case 'difficulty_increased':
      return `网络难度已提高到 ${asString(payload.new_difficulty, asString(payload.difficulty, '--'))}`
    case 'difficulty_decreased':
      return `网络难度已降低到 ${asString(payload.new_difficulty, asString(payload.difficulty, '--'))}`
    case 'block_proposed':
      return `${source} 发起新区块提案，高度 ${asNumber(payload.height, 0)}`
    case 'block_finalized':
      return `高度 ${asNumber(payload.height, 0)} 的区块已经完成最终确认`
    case 'epoch_transition':
      return `系统进入 Epoch ${asNumber(payload.epoch, 0)}，验证者集合已更新`
    case 'validator_slashed':
      return `${source} 因违规被惩罚，惩罚量为 ${asString(payload.amount, '--')}`
    case 'vote_cast':
      return `${source} 已提交投票权重，当前权重 ${asNumber(payload.weight, 0)}`
    case 'delegates_elected':
      return `系统已选出 ${asNumber(payload.active_count, 0)} 个活跃委托者，当前领先者为 ${asString(payload.top_delegate, '--')}`
    case 'block_produced':
      return `${source} 在第 ${asNumber(payload.round, 0)} 轮产出高度 ${asNumber(payload.height, 0)} 的区块`
    case 'block_missed':
      return `${source} 错过本轮出块机会，累计漏块 ${asNumber(payload.missed_total, 0)} 次`
    case 'propose':
      return `${source} 发起新提案，等待验证者投票`
    case 'vote_received':
      return `${source} 的投票已被领导者接收，协议继续推进`
    case 'prepare_qc_formed':
      return `Prepare QC 已形成，系统可以进入 Precommit`
    case 'precommit_qc_formed':
      return `Precommit QC 已形成，系统继续进入 Commit`
    case 'commit_qc_formed':
      return `Commit QC 已形成，区块已满足最终提交条件`
    case 'block_committed':
      return `区块已经正式提交到主链，下一轮可以开始`
    case 'leader_rotated':
      return `系统已切换到新的领导者 ${asString(payload.new_leader, target || '--')}`
    case 'prevote':
      return `${source} 正在广播 Prevote，等待达到法定阈值`
    case 'precommit':
      return `${source} 正在广播 Precommit，准备完成提交`
    case 'new_round':
      return `当前轮没有完成最终提交，系统已进入新一轮`
    case 'vertex_created':
      return `${source} 创建了新的 DAG 顶点 ${shortValue(payload.vertex_id, '--')}`
    case 'vertex_confirmed':
      return `顶点 ${shortValue(payload.vertex_id, '--')} 已被确认并进入稳定状态`
    case 'node_selected':
      return `${source} 已被 VRF 选为当前轮候选者`
    case 'election_complete':
      return `本轮 VRF 抽签完成，共选出 ${asNumber(payload.selected_count, 0)} 个候选者`
    case 'block_created':
      return `新的候选区块 ${shortValue(payload.hash, '--')} 已加入链头竞争`
    case 'fork_created':
      return `网络出现新的分叉，当前分支数为 ${asNumber(payload.fork_count, 0)}`
    case 'chain_reorg':
      return `系统已切换规范链头，新主链高度为 ${asNumber(payload.height, 0)}`
    case 'hash_computed':
      return `使用 ${asString(payload.algorithm, '默认算法')} 生成摘要 ${shortValue(payload.output, '--')}`
    case 'hash_multiple':
      return `对同一输入完成多算法对比，共返回 ${Object.keys(asRecord(payload.results)).length} 份摘要`
    case 'avalanche_demo':
      return `输入轻微变化后，有 ${asNumber(payload.different_bits, 0)} 位发生改变，占比 ${asNumber(payload.change_percentage, 0).toFixed(2)}%`
    case 'integrity_verified':
      return `完整性校验${payload.valid ? '通过' : '失败'}，算法为 ${asString(payload.algorithm, '默认算法')}`
    case 'bridge_initiated':
      return `用户 ${asString(payload.user, '--')} 从 ${asString(payload.source_chain, '--')} 向 ${asString(payload.dest_chain, '--')} 发起 ${asString(payload.amount, '--')} ${asString(payload.token, '')} 跨链`
    case 'bridge_confirmed':
      return `源链确认数已达到 ${asNumber(payload.confirmations, 0)}，交易进入可签名阶段`
    case 'bridge_signed':
      return `验证者 ${asString(payload.validator, '--')} 已签名，当前 ${asNumber(payload.signatures, 0)}/${asNumber(payload.required, 0)}`
    case 'bridge_completed':
      return `目标链已完成执行，到账数量为 ${asString(payload.net_amount, '--')}，目标交易 ${shortValue(payload.dest_tx_hash, '--')}`
    case 'trace_complete':
      return `已生成 ${asString(payload.scenario, '默认场景')} 的完整调用跟踪，最大深度 ${asNumber(payload.max_depth, 0)}`
    default: {
      const detail = target ? `${source} -> ${target}` : source
      return detail === '系统' ? getVisualizationEventLabel(event.type) : detail
    }
  }
}
