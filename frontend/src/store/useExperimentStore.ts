import { create } from 'zustand'
import { buildWorkbenchInstances, buildWorkbenchTools } from '@/domains/experiment/workbench'
import type {
  EnvStatus,
  Experiment,
  ExperimentEnv,
  ExperimentSession,
  ExperimentSessionMember,
  ExperimentSessionMessage,
  WorkspaceLogEntry,
} from '@/types'
import { getExperiment } from '@/api/experiment'
import {
  createExperimentSnapshot,
  extendExperimentEnv,
  getExperimentEnv,
  getExperimentSession,
  getExperimentSessionLogs,
  joinExperimentSession,
  listExperimentSessionMessages,
  leaveExperimentSession,
  pauseExperimentEnv,
  resumeExperimentEnv,
  sendExperimentSessionMessage,
  startExperimentEnv,
  stopExperimentEnv,
  updateExperimentSessionMember,
} from '@/api/experimentSession'

const envStartInflight = new Map<number, Promise<ExperimentEnv>>()

interface ExperimentState {
  currentExperiment: Experiment | null
  currentEnv: ExperimentEnv | null
  currentSession: ExperimentSession | null
  currentInstances: ExperimentEnv['instances']
  currentTools: ReturnType<typeof buildWorkbenchTools>
  sessionMembers: ExperimentSessionMember[]
  sessionMessages: ExperimentSessionMessage[]
  sessionLogs: WorkspaceLogEntry[]
  envStatus: EnvStatus | null
  remainingSeconds: number
  loading: boolean
  statusPollingId: number | null
  loadWorkbench: (experimentId: number) => Promise<void>
  setExperiment: (experiment: Experiment | null) => void
  setEnv: (env: ExperimentEnv | null) => void
  refreshSession: (sessionKey: string) => Promise<void>
  refreshMessages: (sessionKey: string) => Promise<void>
  refreshLogs: (sessionKey: string, params?: { source?: string; levels?: string[] }) => Promise<void>
  postMessage: (sessionKey: string, message: string) => Promise<void>
  joinSession: (sessionKey: string) => Promise<void>
  leaveSession: (sessionKey: string) => Promise<void>
  updateSessionMember: (
    sessionKey: string,
    userId: number,
    payload: { role_key?: string; assigned_node_key?: string; join_status?: 'joined' | 'left' },
  ) => Promise<void>
  startEnv: (experimentId: number, snapshotUrl?: string) => Promise<void>
  fetchEnvStatus: (envId: string) => Promise<void>
  extendEnv: (envId: string, duration?: number) => Promise<void>
  pauseEnv: (envId: string) => Promise<void>
  resumeEnv: (envId: string) => Promise<void>
  createSnapshot: (envId: string) => Promise<ExperimentEnv | null>
  stopEnv: (envId: string) => Promise<void>
  startStatusPolling: (envId: string) => void
  stopStatusPolling: () => void
  updateRemainingTime: () => void
  reset: () => void
}

function getRemainingSeconds(expiresAt?: string) {
  if (!expiresAt) {
    return 0
  }

  return Math.max(0, Math.floor((new Date(expiresAt).getTime() - Date.now()) / 1000))
}

export const useExperimentStore = create<ExperimentState>((set, get) => ({
  currentExperiment: null,
  currentEnv: null,
  currentSession: null,
  currentInstances: [],
  currentTools: [],
  sessionMembers: [],
  sessionMessages: [],
  sessionLogs: [],
  envStatus: null,
  remainingSeconds: 0,
  loading: false,
  statusPollingId: null,

  loadWorkbench: async (experimentId) => {
    set({ loading: true })
    try {
      const experiment = await getExperiment(experimentId)
      set({ currentExperiment: experiment })
      await get().startEnv(experimentId)
    } finally {
      set({ loading: false })
    }
  },

  setExperiment: (experiment) => {
    set({ currentExperiment: experiment })
  },

  setEnv: (env) => {
    const session = env?.session || null
    set({
      currentEnv: env,
      currentSession: session,
      currentInstances: buildWorkbenchInstances(env),
      currentTools: buildWorkbenchTools(env),
      sessionMembers: session?.members || [],
      envStatus: env?.status || null,
      remainingSeconds: getRemainingSeconds(env?.expires_at),
    })
  },

  refreshSession: async (sessionKey) => {
    const session = await getExperimentSession(sessionKey)
    set({
      currentSession: session,
      sessionMembers: session.members || [],
    })
  },

  refreshMessages: async (sessionKey) => {
    const response = await listExperimentSessionMessages(sessionKey, {
      page: 1,
      page_size: 100,
    })
    set({ sessionMessages: response.list || [] })
  },

  refreshLogs: async (sessionKey, params) => {
    const response = await getExperimentSessionLogs(sessionKey, params)
    set({ sessionLogs: response.logs || [] })
  },

  postMessage: async (sessionKey, content) => {
    const created = await sendExperimentSessionMessage(sessionKey, content)
    set((state) => ({
      sessionMessages: [...state.sessionMessages, created],
    }))
  },

  joinSession: async (sessionKey) => {
    const session = await joinExperimentSession(sessionKey)
    set((state) => ({
      currentSession: session,
      sessionMembers: session.members || [],
      currentEnv: state.currentEnv ? { ...state.currentEnv, session } : state.currentEnv,
    }))
  },

  leaveSession: async (sessionKey) => {
    const session = await leaveExperimentSession(sessionKey)
    set((state) => ({
      currentSession: session,
      sessionMembers: session.members || [],
      currentEnv: state.currentEnv ? { ...state.currentEnv, session } : state.currentEnv,
    }))
  },

  updateSessionMember: async (sessionKey, userId, payload) => {
    const session = await updateExperimentSessionMember(sessionKey, userId, payload)
    set((state) => ({
      currentSession: session,
      sessionMembers: session.members || [],
      currentEnv: state.currentEnv ? { ...state.currentEnv, session } : state.currentEnv,
    }))
  },

  startEnv: async (experimentId, snapshotUrl) => {
    try {
      let request = envStartInflight.get(experimentId)
      if (!request) {
        request = startExperimentEnv(experimentId, snapshotUrl)
        envStartInflight.set(experimentId, request)
      }

      const env = await request
      get().setEnv(env)

      if (env.session?.session_key) {
        await get().joinSession(env.session.session_key)
      }

      if (env.session?.session_key) {
        await Promise.all([
          get().refreshMessages(env.session.session_key),
          get().refreshLogs(env.session.session_key),
        ])
      } else {
        set({ sessionMessages: [], sessionLogs: [] })
      }

      if (env.env_id) {
        get().startStatusPolling(env.env_id)
      }
    } finally {
      envStartInflight.delete(experimentId)
    }
  },

  fetchEnvStatus: async (envId) => {
    try {
      const env = await getExperimentEnv(envId)
      get().setEnv(env)

      if (env.session?.session_key) {
        await Promise.all([
          get().refreshSession(env.session.session_key),
          get().refreshMessages(env.session.session_key),
          get().refreshLogs(env.session.session_key),
        ])
      }

      if (env.status === 'terminated' || env.status === 'failed') {
        get().stopStatusPolling()
      }
    } catch {
      set({
        currentEnv: null,
        currentSession: null,
        currentInstances: [],
        currentTools: [],
        sessionMembers: [],
        sessionMessages: [],
        sessionLogs: [],
        envStatus: null,
        remainingSeconds: 0,
      })
    }
  },

  extendEnv: async (envId, duration = 60) => {
    set({ loading: true })
    try {
      await extendExperimentEnv(envId, duration)
      await get().fetchEnvStatus(envId)
    } finally {
      set({ loading: false })
    }
  },

  pauseEnv: async (envId) => {
    set({ loading: true })
    try {
      await pauseExperimentEnv(envId)
      await get().fetchEnvStatus(envId)
    } finally {
      set({ loading: false })
    }
  },

  resumeEnv: async (envId) => {
    set({ loading: true })
    try {
      await resumeExperimentEnv(envId)
      await get().fetchEnvStatus(envId)
    } finally {
      set({ loading: false })
    }
  },

  createSnapshot: async (envId) => {
    set({ loading: true })
    try {
      const env = await createExperimentSnapshot(envId)
      get().setEnv(env)
      return env
    } finally {
      set({ loading: false })
    }
  },

  stopEnv: async (envId) => {
    set({ loading: true })
    try {
      await stopExperimentEnv(envId)
      set({ envStatus: 'terminated' })
      get().stopStatusPolling()
    } finally {
      set({ loading: false })
    }
  },

  startStatusPolling: (envId) => {
    get().stopStatusPolling()

    const pollingId = window.setInterval(() => {
      void get().fetchEnvStatus(envId)
    }, 3000)

    set({ statusPollingId: pollingId })
  },

  stopStatusPolling: () => {
    const { statusPollingId } = get()
    if (statusPollingId) {
      window.clearInterval(statusPollingId)
      set({ statusPollingId: null })
    }
  },

  updateRemainingTime: () => {
    const { remainingSeconds, envStatus } = get()
    if (envStatus === 'running' && remainingSeconds > 0) {
      set({ remainingSeconds: remainingSeconds - 1 })
    }
  },

  reset: () => {
    get().stopStatusPolling()
    set({
      currentExperiment: null,
      currentEnv: null,
      currentSession: null,
      currentInstances: [],
      currentTools: [],
      sessionMembers: [],
      sessionMessages: [],
      sessionLogs: [],
      envStatus: null,
      remainingSeconds: 0,
      loading: false,
    })
  },
}))
