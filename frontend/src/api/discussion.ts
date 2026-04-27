/**
 * 讨论区 API
 */
import { get, post, put, del } from './request'
import type { PaginatedData, Post, Reply } from '@/types'

/**
 * 获取帖子列表
 * 后端路由: GET /posts
 */
export function getPosts(params?: {
  page?: number
  page_size?: number
  course_id?: number
  keyword?: string
}): Promise<PaginatedData<Post>> {
  return get<PaginatedData<Post>>('/posts', params)
}

/**
 * 获取帖子详情
 * 后端路由: GET /posts/:id
 */
export function getPost(postId: number): Promise<Post & { replies: Reply[] }> {
  return get<Post & { replies: Reply[] }>(`/posts/${postId}`)
}

/**
 * 创建帖子
 * 后端路由: POST /posts
 */
export function createPost(data: {
  course_id: number
  title: string
  content: string
}): Promise<Post> {
  return post<Post>('/posts', data)
}

/**
 * 回复帖子
 * 后端路由: POST /posts/:id/replies
 */
export function replyPost(postId: number, content: string): Promise<Reply> {
  return post<Reply>(`/posts/${postId}/replies`, { content })
}

/**
 * 获取回复列表
 * 后端路由: GET /posts/:id/replies
 */
export function getReplies(postId: number, params?: {
  page?: number
  page_size?: number
}): Promise<PaginatedData<Reply>> {
  return get<PaginatedData<Reply>>(`/posts/${postId}/replies`, params)
}

/**
 * 创建回复
 * 后端路由: POST /posts/:id/replies
 */
export function createReply(postId: number, data: { content: string }): Promise<Reply> {
  return post<Reply>(`/posts/${postId}/replies`, data)
}

/**
 * 更新帖子
 * 后端路由: PUT /posts/:id
 */
export function updatePost(postId: number, data: {
  title?: string
  content?: string
}): Promise<Post> {
  return put<Post>(`/posts/${postId}`, data)
}

/**
 * 删除帖子
 * 后端路由: DELETE /posts/:id
 */
export function deletePost(postId: number): Promise<void> {
  return del(`/posts/${postId}`)
}

/**
 * 点赞帖子
 * 后端路由: POST /posts/:id/like
 */
export function likePost(postId: number): Promise<void> {
  return post(`/posts/${postId}/like`)
}

/**
 * 取消点赞帖子
 * 后端路由: DELETE /posts/:id/like
 */
export function unlikePost(postId: number): Promise<void> {
  return del(`/posts/${postId}/like`)
}

/**
 * 删除回复
 * 后端路由: DELETE /posts/:id/replies/:reply_id
 */
export function deleteReply(postId: number, replyId: number): Promise<void> {
  return del(`/posts/${postId}/replies/${replyId}`)
}

/**
 * 点赞回复
 * 后端路由: POST /posts/:id/replies/:reply_id/like
 */
export function likeReply(postId: number, replyId: number): Promise<void> {
  return post(`/posts/${postId}/replies/${replyId}/like`)
}

/**
 * 取消点赞回复
 * 后端路由: DELETE /posts/:id/replies/:reply_id/like
 */
export function unlikeReply(postId: number, replyId: number): Promise<void> {
  return del(`/posts/${postId}/replies/${replyId}/like`)
}

/**
 * 采纳回复
 * 后端路由: POST /posts/:id/accept
 */
export function acceptReply(postId: number, replyId: number): Promise<void> {
  return post(`/posts/${postId}/accept`, { reply_id: replyId })
}

/**
 * 置顶帖子
 * 后端路由: POST /posts/:id/pin
 */
export function pinPost(postId: number): Promise<void> {
  return post(`/posts/${postId}/pin`)
}

/**
 * 锁定帖子
 * 后端路由: POST /posts/:id/lock
 */
export function lockPost(postId: number): Promise<void> {
  return post(`/posts/${postId}/lock`)
}
