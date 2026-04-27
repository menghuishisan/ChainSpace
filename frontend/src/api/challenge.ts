/**
 * 题目管理 API
 * 后端路由: /challenges/*
 */
import { get, post, put, del } from './request'
import type { Challenge, PaginatedData, CreateChallengeRequest, UpdateChallengeRequest } from '@/types'

/**
 * 获取题目列表
 * 后端路由: GET /challenges
 */
export function getChallenges(params?: {
  page?: number
  page_size?: number
  keyword?: string
  category?: string
  difficulty?: number
  source_type?: 'preset' | 'auto_converted' | 'user_created'
  status?: 'draft' | 'active' | 'archived'
  is_public?: boolean
}): Promise<PaginatedData<Challenge>> {
  return get<PaginatedData<Challenge>>('/challenges', params)
}

/**
 * 获取题目详情
 * 后端路由: GET /challenges/:id
 */
export function getChallenge(id: number): Promise<Challenge> {
  return get<Challenge>(`/challenges/${id}`)
}

/**
 * 创建题目
 * 后端路由: POST /challenges
 */
export function createChallenge(data: CreateChallengeRequest): Promise<Challenge> {
  return post<Challenge>('/challenges', data)
}

/**
 * 更新题目
 * 后端路由: PUT /challenges/:id
 */
export function updateChallenge(id: number, data: UpdateChallengeRequest): Promise<Challenge> {
  return put<Challenge>(`/challenges/${id}`, data)
}

/**
 * 删除题目
 * 后端路由: DELETE /challenges/:id
 */
export function deleteChallenge(id: number): Promise<void> {
  return del(`/challenges/${id}`)
}

/**
 * 申请公开题目
 * 后端路由: POST /challenges/:id/publish-request
 */
export function requestPublishChallenge(id: number, reason?: string): Promise<void> {
  return post(`/challenges/${id}/publish-request`, { reason })
}
