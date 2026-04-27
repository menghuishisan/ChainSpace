import type { Contest, ContestScore } from '@/types'
import { getBattleConfigFromOrchestration } from '@/domains/contest/battle'

export interface ContestAgentBattleDetailState {
  canSpectate: boolean
  canReplay: boolean
  scoreWeights: Record<string, number>
}

export function buildContestAgentBattleDetailState(contest: Contest): ContestAgentBattleDetailState {
  const battleConfig = getBattleConfigFromOrchestration(contest.battle_orchestration)

  return {
    canSpectate: contest.status === 'ongoing' || contest.status === 'ended',
    canReplay: contest.status === 'ended' && Boolean(battleConfig.spectate.enable_replay),
    scoreWeights: battleConfig.judge.score_weights || {},
  }
}

export function mapFinalRankToContestScores(
  data: Array<{
    rank: number
    team_id: number
    team_name: string
    total_score: number
  }>,
): ContestScore[] {
  return (data || []).map((item) => ({
    rank: item.rank,
    team_id: item.team_id,
    team_name: item.team_name,
    total_score: item.total_score,
    solve_count: 0,
  }))
}
