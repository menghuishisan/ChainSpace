/**
 * 平台管理员 API
 */
import { del, get, post, put } from './request'
import type {
  ChallengePublishApplication,
  ContainerStats,
  CrossSchoolApplication,
  DockerImage,
  DockerImageCapability,
  OperationLog,
  PaginatedData,
  School,
  ServiceHealth,
  SystemConfig,
  SystemMonitor,
  SystemStats,
  VulnerabilityCandidate,
  CreateSchoolRequest,
} from '@/types'

export type AsyncTaskSubmitResult = {
  task_id: string
  type: string
}

export function getSchools(params?: {
  page?: number
  page_size?: number
  keyword?: string
  status?: string
}): Promise<PaginatedData<School>> {
  return get<PaginatedData<School>>('/schools', params)
}

export function createSchool(data: CreateSchoolRequest): Promise<School> {
  return post<School>('/schools', data)
}

export function updateSchool(id: number, data: Partial<CreateSchoolRequest>): Promise<School> {
  return put<School>(`/schools/${id}`, data)
}

export function updateSchoolStatus(id: number, status: 'active' | 'disabled'): Promise<void> {
  return put(`/schools/${id}/status`, { status })
}

export function deleteSchool(id: number): Promise<void> {
  return del(`/schools/${id}`)
}

export function getImages(params?: {
  page?: number
  page_size?: number
  category?: string
}): Promise<PaginatedData<DockerImage>> {
  return get<PaginatedData<DockerImage>>('/images', params)
}

export function getAllImages(): Promise<DockerImage[]> {
  return get<DockerImage[]>('/images/all')
}

export function getImageCapabilities(): Promise<DockerImageCapability[]> {
  return get<DockerImageCapability[]>('/images/capabilities')
}

export function createImage(data: {
  name: string
  tag: string
  category: string
  description?: string
  default_resources: {
    cpu: number
    memory: string
    storage: string
  }
}): Promise<DockerImage> {
  return post<DockerImage>('/images', data)
}

export function updateImage(id: number, data: {
  name?: string
  tag?: string
  category?: string
  description?: string
  default_resources?: { cpu: number; memory: string; storage: string }
  is_active?: boolean
}): Promise<DockerImage> {
  return put<DockerImage>(`/images/${id}`, data)
}

export function deleteImage(id: number): Promise<void> {
  return del(`/images/${id}`)
}

export function getConfigs(): Promise<SystemConfig[]> {
  return get<SystemConfig[]>('/system/configs')
}

export function updateConfig(key: string, value: string): Promise<void> {
  return post('/system/configs', { key, value })
}

export function getSystemStats(): Promise<SystemStats> {
  return get<SystemStats>('/system/stats')
}

export function getSystemMonitor(): Promise<SystemMonitor> {
  return get('/system/monitor')
}

export function getContainerStats(): Promise<ContainerStats> {
  return get('/system/containers')
}

export function getServiceHealth(): Promise<ServiceHealth[]> {
  return get('/system/services')
}

export function getCrossSchoolReviews(params?: {
  page?: number
  page_size?: number
  status?: string
}): Promise<PaginatedData<CrossSchoolApplication>> {
  return get<PaginatedData<CrossSchoolApplication>>('/system/cross-school', params)
}

export function reviewCrossSchool(id: number, action: 'approve' | 'reject', comment?: string): Promise<void> {
  return post(`/system/cross-school/${id}/handle`, { action, comment })
}

export function getChallengePublishReviews(params?: {
  page?: number
  page_size?: number
  status?: string
}): Promise<PaginatedData<ChallengePublishApplication>> {
  return get<PaginatedData<ChallengePublishApplication>>('/system/challenge-reviews', params)
}

export function reviewChallengePublish(id: number, action: 'approve' | 'reject', comment?: string): Promise<void> {
  return post(`/system/challenge-reviews/${id}/handle`, { action, comment })
}

export function getVulnerabilities(params?: {
  page?: number
  page_size?: number
  keyword?: string
  status?: string
  category?: string
  severity?: string
  chain?: string
}): Promise<PaginatedData<VulnerabilityCandidate>> {
  return get<PaginatedData<VulnerabilityCandidate>>('/system/vulnerabilities', params)
}

export function convertVulnerability(id: number): Promise<void> {
  return post(`/system/vulnerabilities/${id}/convert`)
}

export function skipVulnerability(id: number): Promise<void> {
  return put(`/system/vulnerabilities/${id}/skip`)
}

export function updateVulnerability(id: number, data: {
  contract_address?: string
  attack_tx_hash?: string
  fork_block_number?: number
  related_contracts?: string[]
  related_tokens?: string[]
  attacker_addresses?: string[]
  victim_addresses?: string[]
  evidence_links?: string[]
  runtime_profile_suggestion?: 'static' | 'single_chain_instance' | 'fork_replay' | 'multi_service_lab'
}): Promise<void> {
  return put(`/system/vulnerabilities/${id}`, data)
}

export function enrichVulnerabilityCode(id: number): Promise<AsyncTaskSubmitResult> {
  return post<AsyncTaskSubmitResult>(`/system/vulnerabilities/${id}/enrich`)
}

export function syncVulnerabilities(): Promise<AsyncTaskSubmitResult> {
  return post<AsyncTaskSubmitResult>('/system/vulnerabilities/sync')
}

export function getOperationLogs(params?: {
  page?: number
  page_size?: number
}): Promise<PaginatedData<OperationLog>> {
  return get('/system/logs', params)
}
