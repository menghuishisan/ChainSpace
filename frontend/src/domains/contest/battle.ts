import { buildDefaultBattleOrchestration, normalizeBattleOrchestration } from '@/domains/contest/management'
import type {
  AgentBattleStatus,
  BattleOrchestration,
  BattleRoundPhase,
  Contest,
  Scoreboard,
} from '@/types'
import { BattleRoundPhaseMap, BattleStatusMap } from '@/types'

export const ROUND_STATUS_MAP: Record<string, { text: string; color: string }> = {
  pending: { text: '等待开始', color: 'default' },
  running: { text: '进行中', color: 'success' },
  finished: { text: '已结束', color: 'blue' },
  failed: { text: '异常', color: 'error' },
}

export const STRATEGY_TEMPLATE = `// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

contract StrategyAgentV1 {
    enum ActionType {
        Gather,
        Attack,
        Defend,
        Fortify,
        Recover,
        Scout
    }

    struct GameState {
        uint256 roundNumber;
        uint256 selfResource;
        uint256 selfScore;
        uint256 targetCount;
    }

    function decideAction(GameState calldata state) external pure returns (ActionType) {
        if (state.selfResource < 20) {
            return ActionType.Gather;
        }
        if (state.selfScore > 100) {
            return ActionType.Defend;
        }
        return ActionType.Attack;
    }

    function selectTarget(GameState calldata state) external pure returns (uint256) {
        if (state.targetCount == 0) {
            return 0;
        }
        return state.roundNumber % state.targetCount;
    }

    function allocateResources(GameState calldata state) external pure returns (uint256 attackBudget, uint256 defenseBudget) {
        attackBudget = state.selfResource * 60 / 100;
        defenseBudget = state.selfResource - attackBudget;
    }
}`

export function getBattleConfig(contest: Contest | null): BattleOrchestration {
  if (!contest?.battle_orchestration) {
    return buildDefaultBattleOrchestration()
  }

  return normalizeBattleOrchestration(contest.battle_orchestration)
}

export function getBattleConfigFromOrchestration(orchestration?: BattleOrchestration): BattleOrchestration {
  return orchestration ? normalizeBattleOrchestration(orchestration) : buildDefaultBattleOrchestration()
}

export function getBattleScoreWeights(orchestration?: BattleOrchestration) {
  return getBattleConfigFromOrchestration(orchestration).judge.score_weights || {}
}

export function getBattleStatusConfig(status?: AgentBattleStatus['status']) {
  return status ? BattleStatusMap[status] : null
}

export function getRoundStatusConfig(status?: string) {
  return status ? ROUND_STATUS_MAP[status] || null : null
}

export function getRoundPhaseConfig(phase?: BattleRoundPhase | string) {
  if (!phase) {
    return null
  }

  return BattleRoundPhaseMap[phase as BattleRoundPhase] || null
}

export function getBattleScoreTopList(scoreboard: Scoreboard | null) {
  return scoreboard?.list.slice(0, 10) || []
}

export function getContestRemainingSeconds(endTime: string): number {
  return Math.max(0, Math.floor((new Date(endTime).getTime() - Date.now()) / 1000))
}
