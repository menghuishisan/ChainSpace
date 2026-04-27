import type { AgentBattleEvent, AgentBattleTeamStatus, Scoreboard, SpectateData } from '@/types'

export interface RankedTeamView {
  rank: number
  team_id: number
  team_name: string
  total_score: number
  resource_held: number
  is_alive: boolean
}

export const EVENT_COLOR_MAP: Record<string, string> = {
  gather: 'green',
  attack: 'red',
  defend: 'blue',
  fortify: 'gold',
  recover: 'cyan',
  scout: 'purple',
}

export function buildRankedTeams(
  teamStatusList: AgentBattleTeamStatus[] | undefined,
  scoreboard: Scoreboard | null,
): RankedTeamView[] {
  if (teamStatusList?.length) {
    return [...teamStatusList]
      .map((team) => {
        const legacyTeam = team as AgentBattleTeamStatus & { score?: number }
        return {
          ...team,
          total_score: team.total_score ?? legacyTeam.score ?? 0,
          resource_held: team.resource_held ?? 0,
        }
      })
      .sort((left, right) => right.total_score - left.total_score)
      .map((team, index) => ({
        rank: index + 1,
        team_id: team.team_id,
        team_name: team.team_name,
        total_score: team.total_score,
        resource_held: team.resource_held,
        is_alive: team.is_alive,
      }))
  }

  return (scoreboard?.list || []).map((team) => ({
    rank: team.rank,
    team_id: team.team_id || 0,
    team_name: team.team_name || `队伍 ${team.rank}`,
    total_score: team.total_score,
    resource_held: 0,
    is_alive: true,
  }))
}

export function getSpectateEventTimeline(data: SpectateData | null): AgentBattleEvent[] {
  return data?.recent_events || []
}

export function getSpectateHeadlineMetrics(rankedTeams: RankedTeamView[]) {
  const topTeam = rankedTeams[0]
  const secondTeam = rankedTeams[1]

  return {
    topTeam,
    secondTeam,
    scoreGap: topTeam && secondTeam ? topTeam.total_score - secondTeam.total_score : 0,
    aliveCount: rankedTeams.filter((team) => team.is_alive).length,
    totalResource: rankedTeams.reduce((sum, team) => sum + team.resource_held, 0),
  }
}
