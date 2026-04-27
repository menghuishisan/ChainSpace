import { del, get, post, put } from './request'
import type {
  CreateExperimentRequest,
  Experiment,
  PaginatedData,
  UpdateExperimentRequest,
} from '@/types'

export function getExperiments(params?: {
  page?: number
  page_size?: number
  course_id?: number
  chapter_id?: number
  type?: string
  status?: string
  keyword?: string
}): Promise<PaginatedData<Experiment>> {
  return get<PaginatedData<Experiment>>('/experiments', params)
}

export function getStudentExperiments(params?: {
  page?: number
  page_size?: number
  course_id?: number
  chapter_id?: number
  type?: string
  status?: string
  keyword?: string
}): Promise<PaginatedData<Experiment>> {
  return getExperiments(params)
}

export function createExperiment(data: CreateExperimentRequest): Promise<Experiment> {
  return post<Experiment>('/experiments', data)
}

export function getExperiment(id: number): Promise<Experiment> {
  return get<Experiment>(`/experiments/${id}`)
}

export function updateExperiment(id: number, data: UpdateExperimentRequest): Promise<Experiment> {
  return put<Experiment>(`/experiments/${id}`, data)
}

export function publishExperiment(id: number): Promise<void> {
  return put<void>(`/experiments/${id}/publish`)
}

export function deleteExperiment(id: number): Promise<void> {
  return del<void>(`/experiments/${id}`)
}
