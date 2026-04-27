import type { AttackMechanism, VisualizationRuntimeSpec } from '@/types/visualizationDomain'

const EXECUTION_ATTACKS = new Set([
  'attacks/access_control',
  'attacks/delegatecall_attack',
  'attacks/dos',
  'attacks/integer_overflow',
  'attacks/reentrancy',
  'attacks/selfdestruct',
  'attacks/signature_replay',
  'attacks/timestamp_manipulation',
  'attacks/tx_origin',
  'attacks/weak_randomness',
])

const ECONOMIC_ATTACKS = new Set([
  'attacks/fake_deposit',
  'attacks/flashloan',
  'attacks/frontrun',
  'attacks/governance',
  'attacks/infinite_mint',
  'attacks/liquidation',
  'attacks/oracle_manipulation',
  'attacks/sandwich',
])

const CONSENSUS_ATTACKS = new Set([
  'attacks/attack_51',
  'attacks/bribery',
  'attacks/long_range',
  'attacks/nothing_at_stake',
  'attacks/selfish_mining',
])

const BRIDGE_ATTACKS = new Set([
  'attacks/bridge_attack',
])

/**
 * 根据攻击模块所属机制，决定使用哪一类攻击主舞台。
 * 攻击类不能继续共用一张统一画布，否则不同漏洞和攻击路径会被错误压平。
 */
export function getAttackMechanism(runtime: VisualizationRuntimeSpec): AttackMechanism {
  if (BRIDGE_ATTACKS.has(runtime.moduleKey)) {
    return 'bridge'
  }

  if (CONSENSUS_ATTACKS.has(runtime.moduleKey)) {
    return 'consensus'
  }

  if (ECONOMIC_ATTACKS.has(runtime.moduleKey)) {
    return 'economic'
  }

  if (EXECUTION_ATTACKS.has(runtime.moduleKey)) {
    return 'execution'
  }

  return 'execution'
}

/**
 * 为学生界面提供当前攻击机制的简短标签，避免把不相干的观察重点放到同一个页面。
 */
export function getAttackMechanismLabel(mechanism: AttackMechanism): string {
  switch (mechanism) {
    case 'execution':
      return '执行链路与状态篡改'
    case 'economic':
      return '经济操纵与资金路径'
    case 'consensus':
      return '链竞争与共识破坏'
    case 'bridge':
      return '跨链验证与消息伪造'
    default:
      return '攻击过程'
  }
}
