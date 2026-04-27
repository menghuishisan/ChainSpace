import { getRoundPhaseConfig } from './battle'

export interface ReplaySnapshot {
  block: number
  teams: Array<{
    team_id: number
    score: number
    resource: number
  }>
  events: Array<{
    event_type: string
    actor_team?: string
    target_team?: string
    action_result?: string
    score_delta?: number
    resource_delta?: number
    description: string
  }>
}

export interface FinalRankItem {
  rank: number
  team_id: number
  team_name: string
  total_score: number
}

export function getReplayCurrentSnapshot(
  snapshots: ReplaySnapshot[],
  currentIndex: number,
): ReplaySnapshot | undefined {
  return snapshots[Math.min(currentIndex, Math.max(snapshots.length - 1, 0))]
}

export function buildReplayRoundOptions(
  rounds: Array<{ id: number; round_number: number; phase?: string }>,
): Array<{ label: string; value: number }> {
  return rounds.map((round) => ({
    label: round.phase
      ? `第 ${round.round_number} 轮 (${getRoundPhaseConfig(round.phase)?.text || round.phase})`
      : `第 ${round.round_number} 轮`,
    value: round.id,
  }))
}
