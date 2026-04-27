import type { Challenge, Contest, ContestScore, Scoreboard } from '@/types'
import type { ContestChallengeStat, ContestRoundInfo } from '@/types/presentation'

export function getContestMonitorRemainingSeconds(contest: Contest | null): number {
  if (!contest) {
    return 0
  }

  return Math.max(0, Math.floor((new Date(contest.end_time).getTime() - Date.now()) / 1000))
}

export function buildChallengeStats(challenges: Challenge[]): ContestChallengeStat[] {
  return challenges.map((challenge) => ({
    id: challenge.id,
    title: challenge.title,
    points: challenge.points || challenge.base_points || 100,
    solve_count: challenge.solve_count || 0,
    first_blood: challenge.first_blood,
    first_blood_time: challenge.first_blood_time,
  }))
}

export function getNextRoundNumber(rounds: ContestRoundInfo[]): number {
  return rounds.length > 0 ? Math.max(...rounds.map((round) => round.round_number)) + 1 : 1
}

export function getRunningRoundNumber(rounds: ContestRoundInfo[]): number | string {
  return rounds.find((round) => round.status === 'running')?.round_number || '-'
}

export function getFinishedRoundCount(rounds: ContestRoundInfo[]): number {
  return rounds.filter((round) => round.status === 'finished').length
}

export function getChallengeSolveRate(record: ContestChallengeStat, scoreboard: Scoreboard | null): number {
  return scoreboard?.list.length ? Math.round((record.solve_count / scoreboard.list.length) * 100) : 0
}

export function getContestScoreboardPreview(scoreboard: Scoreboard | null): ContestScore[] {
  return scoreboard?.list.slice(0, 20) || []
}

export const ROUND_STATUS_MAP: Record<string, { text: string; color: string }> = {
  pending: { text: '等待开始', color: 'default' },
  running: { text: '对抗进行中', color: 'success' },
  finished: { text: '已结束', color: 'blue' },
  failed: { text: '异常', color: 'error' },
}
