import type { BattleOrchestration, ContestType } from '@/types'
import { buildDefaultBattleOrchestration, normalizeBattleOrchestration } from '@/domains/contest/management'

export function resolveBattleOrchestrationByContestType(
  type: ContestType,
  current?: BattleOrchestration,
): BattleOrchestration | undefined {
  if (type !== 'agent_battle') {
    return undefined
  }

  return current ? normalizeBattleOrchestration(current) : buildDefaultBattleOrchestration()
}

export function ensureBattleOrchestration(
  orchestration?: BattleOrchestration,
): BattleOrchestration {
  return orchestration ? normalizeBattleOrchestration(orchestration) : buildDefaultBattleOrchestration()
}
