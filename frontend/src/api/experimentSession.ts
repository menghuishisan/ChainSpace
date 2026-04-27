import { get, post, put } from './request'
import type {
  ExperimentEnv,
  ExperimentSession,
  ExperimentSessionMessage,
  PaginatedData,
  WorkspaceLogEntry,
} from '@/types'

export function listExperimentSessions(params?: {
  page?: number
  page_size?: number
  experiment_id?: number
  status?: string
}): Promise<PaginatedData<ExperimentSession>> {
  return get<PaginatedData<ExperimentSession>>('/experiment-sessions', params)
}

export function getExperimentSession(sessionKey: string): Promise<ExperimentSession> {
  return get<ExperimentSession>(`/experiment-sessions/${sessionKey}`)
}

export function listExperimentSessionMessages(
  sessionKey: string,
  params?: { page?: number; page_size?: number },
): Promise<PaginatedData<ExperimentSessionMessage>> {
  return get<PaginatedData<ExperimentSessionMessage>>(`/experiment-sessions/${sessionKey}/messages`, params)
}

export function sendExperimentSessionMessage(
  sessionKey: string,
  message: string,
): Promise<ExperimentSessionMessage> {
  return post<ExperimentSessionMessage>(`/experiment-sessions/${sessionKey}/messages`, { message })
}

export function joinExperimentSession(sessionKey: string): Promise<ExperimentSession> {
  return post<ExperimentSession>(`/experiment-sessions/${sessionKey}/join`)
}

export function leaveExperimentSession(sessionKey: string): Promise<ExperimentSession> {
  return post<ExperimentSession>(`/experiment-sessions/${sessionKey}/leave`)
}

export function updateExperimentSessionMember(
  sessionKey: string,
  userId: number,
  data: {
    role_key?: string
    assigned_node_key?: string
    join_status?: 'joined' | 'left'
  },
): Promise<ExperimentSession> {
  return put<ExperimentSession>(`/experiment-sessions/${sessionKey}/members/${userId}`, data)
}

export function getExperimentSessionLogs(
  sessionKey: string,
  params?: { source?: string; levels?: string[] },
): Promise<{ logs: WorkspaceLogEntry[] }> {
  return get<{ logs: WorkspaceLogEntry[] }>(`/experiment-sessions/${sessionKey}/logs`, {
    source: params?.source,
    levels: params?.levels?.join(','),
  })
}

export function startExperimentEnv(experimentId: number, snapshotUrl?: string): Promise<ExperimentEnv> {
  return post<ExperimentEnv>('/envs/start', {
    experiment_id: experimentId,
    snapshot_url: snapshotUrl,
  })
}

export function getExperimentEnv(envId: string): Promise<ExperimentEnv> {
  return get<ExperimentEnv>(`/envs/${envId}`)
}

export function listExperimentEnvs(params?: {
  page?: number
  page_size?: number
  experiment_id?: number
  status?: string
}): Promise<PaginatedData<ExperimentEnv>> {
  return get<PaginatedData<ExperimentEnv>>('/envs', params)
}

export function extendExperimentEnv(envId: string, duration: number): Promise<void> {
  return post<void>(`/envs/${envId}/extend`, { duration })
}

export function stopExperimentEnv(envId: string): Promise<void> {
  return post<void>(`/envs/${envId}/stop`)
}

export function pauseExperimentEnv(envId: string): Promise<void> {
  return post<void>(`/envs/${envId}/pause`)
}

export function resumeExperimentEnv(envId: string): Promise<void> {
  return post<void>(`/envs/${envId}/resume`)
}

export function createExperimentSnapshot(envId: string): Promise<ExperimentEnv> {
  return post<ExperimentEnv>(`/envs/${envId}/snapshots`)
}
