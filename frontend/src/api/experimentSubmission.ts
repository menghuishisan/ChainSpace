import { get, post } from './request'
import type { PaginatedData, Submission } from '@/types'

export function submitExperiment(experimentId: number, data: {
  env_id?: string
  content?: string
  file_url?: string
}): Promise<Submission> {
  return post<Submission>('/submissions', { ...data, experiment_id: experimentId })
}

export function getSubmissions(params?: {
  page?: number
  page_size?: number
  experiment_id?: number
  student_id?: number
  status?: string
}): Promise<PaginatedData<Submission>> {
  return get<PaginatedData<Submission>>('/submissions', params)
}

export function gradeSubmission(submissionId: number, data: {
  score: number
  feedback?: string
}): Promise<Submission> {
  return post<Submission>(`/submissions/${submissionId}/grade`, data)
}
