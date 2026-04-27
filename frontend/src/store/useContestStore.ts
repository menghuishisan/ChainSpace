import { create } from 'zustand'

import * as contestApi from '@/api/contest'
import type {
  AgentBattleStatus,
  Challenge,
  Contest,
  ContestChallenge,
  CurrentRoundInfo,
  Scoreboard,
  SpectateData,
  Team,
} from '@/types'

interface ContestState {
  currentContest: Contest | null
  challenges: Challenge[]
  contestChallenges: ContestChallenge[]
  myTeam: Team | null
  scoreboard: Scoreboard | null
  battleStatus: AgentBattleStatus | null
  spectateData: SpectateData | null
  currentRound: CurrentRoundInfo | null
  rounds: CurrentRoundInfo[]
  loading: boolean
  pollingId: number | null
  setContest: (contest: Contest | null) => void
  fetchContest: (contestId: number) => Promise<Contest>
  fetchChallenges: (contestId: number) => Promise<void>
  fetchMyTeam: (contestId: number) => Promise<void>
  fetchContestChallengesAdmin: (contestId: number) => Promise<ContestChallenge[]>
  fetchScoreboard: (contestId: number, options?: { agentBattle?: boolean }) => Promise<void>
  fetchBattleStatus: (contestId: number) => Promise<void>
  fetchSpectateData: (contestId: number) => Promise<void>
  fetchCurrentRound: (contestId: number) => Promise<void>
  fetchRounds: (contestId: number) => Promise<void>
  hydrateBattleWorkspace: (contestId: number) => Promise<void>
  hydrateSpectateWorkspace: (contestId: number) => Promise<void>
  hydrateMonitorWorkspace: (contestId: number, contest?: Contest | null) => Promise<void>
  hydrateJeopardyDetail: (contestId: number, options: {
    contest: Contest
    isManager: boolean
    isRegistered: boolean
    isTeamContest: boolean
    useReviewChallenges: boolean
    showChallengePreview: boolean
  }) => Promise<void>
  submitFlag: (contestId: number, challengeId: number, flag: string) => Promise<boolean>
  startPolling: (contestId: number, type: 'scoreboard' | 'battle' | 'spectate' | 'monitor') => void
  stopPolling: () => void
  reset: () => void
}

async function getContestScoreboard(contestId: number, agentBattle = false): Promise<Scoreboard> {
  return agentBattle
    ? contestApi.getAgentBattleScoreboard(contestId)
    : contestApi.getScoreboard(contestId)
}

export const useContestStore = create<ContestState>((set, get) => ({
  currentContest: null,
  challenges: [],
  contestChallenges: [],
  myTeam: null,
  scoreboard: null,
  battleStatus: null,
  spectateData: null,
  currentRound: null,
  rounds: [],
  loading: false,
  pollingId: null,

  setContest: (contest) => {
    set({ currentContest: contest })
  },

  fetchContest: async (contestId) => {
    const contest = await contestApi.getContest(contestId)
    set({ currentContest: contest })
    return contest
  },

  fetchChallenges: async (contestId) => {
    const challenges = await contestApi.getPlayingChallenges(contestId)
    set({ challenges })
  },

  fetchMyTeam: async (contestId) => {
    try {
      const team = await contestApi.getMyTeam(contestId)
      set({ myTeam: team })
    } catch {
      set({ myTeam: null })
    }
  },

  fetchContestChallengesAdmin: async (contestId) => {
    try {
      const contestChallenges = await contestApi.getContestChallengesAdmin(contestId)
      set({ contestChallenges: contestChallenges || [] })
      return contestChallenges || []
    } catch {
      set({ contestChallenges: [] })
      return []
    }
  },

  fetchScoreboard: async (contestId, options) => {
    try {
      const scoreboard = await getContestScoreboard(contestId, options?.agentBattle)
      set({ scoreboard })
    } catch {
      set({ scoreboard: null })
    }
  },

  fetchBattleStatus: async (contestId) => {
    try {
      const status = await contestApi.getAgentBattleStatus(contestId)
      set({ battleStatus: status })
    } catch {
      set({ battleStatus: null })
    }
  },

  fetchSpectateData: async (contestId) => {
    try {
      const data = await contestApi.getSpectateData(contestId)
      set({ spectateData: data })
    } catch {
      set({ spectateData: null })
    }
  },

  fetchCurrentRound: async (contestId) => {
    try {
      const round = await contestApi.getCurrentRound(contestId)
      set({ currentRound: round })
    } catch {
      set({ currentRound: null })
    }
  },

  fetchRounds: async (contestId) => {
    try {
      const rounds = await contestApi.getAgentBattleRounds(contestId)
      set({ rounds: Array.isArray(rounds) ? rounds : [] })
    } catch {
      set({ rounds: [] })
    }
  },

  hydrateBattleWorkspace: async (contestId) => {
    set({ loading: true })
    try {
      const contest = await get().fetchContest(contestId)
      await Promise.all([
        get().fetchMyTeam(contestId),
        get().fetchBattleStatus(contestId),
        get().fetchScoreboard(contestId, { agentBattle: true }),
        get().fetchCurrentRound(contestId),
      ])
      set({ currentContest: contest })
    } finally {
      set({ loading: false })
    }
  },

  hydrateSpectateWorkspace: async (contestId) => {
    set({ loading: true })
    try {
      const contest = await get().fetchContest(contestId)
      await Promise.all([
        get().fetchSpectateData(contestId),
        get().fetchScoreboard(contestId, { agentBattle: true }),
        get().fetchCurrentRound(contestId),
      ])
      set({ currentContest: contest })
    } finally {
      set({ loading: false })
    }
  },

  hydrateMonitorWorkspace: async (contestId, contest) => {
    set({ loading: true })
    try {
      const resolvedContest = contest || await get().fetchContest(contestId)
      if (resolvedContest.type === 'agent_battle') {
        await Promise.all([
          get().fetchScoreboard(contestId, { agentBattle: true }),
          get().fetchRounds(contestId),
        ])
        set({ challenges: [], contestChallenges: [] })
      } else {
        await Promise.all([
          get().fetchScoreboard(contestId),
          get().fetchChallenges(contestId),
        ])
        set({ rounds: [], contestChallenges: [] })
      }
    } finally {
      set({ loading: false })
    }
  },

  hydrateJeopardyDetail: async (contestId, options) => {
    set({ loading: true, currentContest: options.contest })
    try {
      const tasks: Array<Promise<void>> = []

      if (options.isTeamContest && options.isRegistered) {
        tasks.push(get().fetchMyTeam(contestId))
      } else {
        set({ myTeam: null })
      }

      if (options.isManager) {
        tasks.push(
          get().fetchContestChallengesAdmin(contestId).then((contestChallenges) => {
            if (options.showChallengePreview && !options.useReviewChallenges) {
              set({
                challenges: contestChallenges
                  .filter((item) => item.is_visible && item.challenge)
                  .map((item) => ({
                    ...item.challenge,
                    points: item.current_points || item.points,
                  })),
              })
            }
          }),
        )
      } else {
        set({ contestChallenges: [] })
      }

      if (options.showChallengePreview) {
        if (options.useReviewChallenges) {
          tasks.push(
            contestApi.getContestReviewChallenges(contestId).then((challenges) => set({ challenges: challenges || [] })),
          )
        } else if (!options.isManager) {
          tasks.push(get().fetchChallenges(contestId))
        }
      } else {
        set({ challenges: [] })
      }

      await Promise.all(tasks)
    } finally {
      set({ loading: false })
    }
  },

  submitFlag: async (contestId, challengeId, flag) => {
    set({ loading: true })
    try {
      const result = await contestApi.submitFlag(contestId, challengeId, flag)
      if (result.correct) {
        set((state) => ({
          challenges: state.challenges.map((challenge) => (
            challenge.id === challengeId
              ? { ...challenge, is_solved: true }
              : challenge
          )),
        }))
      }
      return result.correct
    } finally {
      set({ loading: false })
    }
  },

  startPolling: (contestId, type) => {
    get().stopPolling()

    const refresh = () => {
      switch (type) {
        case 'scoreboard':
          void get().fetchScoreboard(contestId)
          break
        case 'battle':
          void Promise.all([
            get().fetchBattleStatus(contestId),
            get().fetchScoreboard(contestId, { agentBattle: true }),
            get().fetchCurrentRound(contestId),
          ])
          break
        case 'spectate':
          void Promise.all([
            get().fetchSpectateData(contestId),
            get().fetchScoreboard(contestId, { agentBattle: true }),
            get().fetchCurrentRound(contestId),
          ])
          break
        case 'monitor': {
          const contest = get().currentContest
          void get().hydrateMonitorWorkspace(contestId, contest)
          break
        }
        default:
          break
      }
    }

    refresh()
    const pollingId = window.setInterval(refresh, 5000)
    set({ pollingId })
  },

  stopPolling: () => {
    const { pollingId } = get()
    if (pollingId) {
      window.clearInterval(pollingId)
      set({ pollingId: null })
    }
  },

  reset: () => {
    get().stopPolling()
    set({
      currentContest: null,
      challenges: [],
      contestChallenges: [],
      myTeam: null,
      scoreboard: null,
      battleStatus: null,
      spectateData: null,
      currentRound: null,
      rounds: [],
      loading: false,
    })
  },
}))
