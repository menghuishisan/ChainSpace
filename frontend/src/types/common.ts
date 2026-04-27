/**
 * 通用类型定义
 */

// API统一响应格式
export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data: T
}

// 分页请求参数
export interface PaginationParams {
  page?: number
  page_size?: number
}

// 分页响应数据
export interface PaginatedData<T> {
  list: T[]
  total: number
  page: number
  page_size: number
}

// 学校信息
export interface School {
  id: number
  name: string
  code?: string
  logo?: string
  address?: string
  contact?: string
  phone?: string
  email?: string
  website?: string
  description?: string
  status: 'active' | 'disabled'
  expire_at?: string
  teacher_count?: number
  student_count?: number
  created_at: string
}

// 班级信息
export interface Class {
  id: number
  school_id: number
  name: string
  grade?: string
  major?: string
  description?: string
  status: 'active' | 'disabled'
  student_count?: number
  created_at: string
}

// 通知类型
export type NotificationType =
  | 'system'          // 系统通知
  | 'course'          // 课程通知
  | 'experiment'      // 实验通知
  | 'contest'         // 竞赛通知
  | 'grade'           // 成绩通知
  | 'submission'      // 提交通知
  | 'announce'        // 公告

// 通知信息（对齐后端 NotificationResponse）
export interface Notification {
  id: number
  user_id?: number
  type: NotificationType
  title: string
  content?: string
  link?: string
  related_id?: number
  related_type?: string
  is_read: boolean
  read_at?: string
  sender_id?: number
  sender_name?: string
  extra?: Record<string, unknown>
  created_at: string
}

// 帖子信息
export interface Post {
  id: number
  course_id: number
  author_id: number
  author_name?: string
  author_avatar?: string
  title: string
  content: string
  tags?: string[]
  is_pinned: boolean
  is_locked: boolean
  status: string
  reply_count: number
  view_count: number
  like_count: number
  is_liked?: boolean
  created_at: string
  last_reply_at?: string
}

// 回复信息
export interface Reply {
  id: number
  post_id: number
  author_id: number
  author_name?: string
  author_avatar?: string
  parent_id?: number
  content: string
  like_count: number
  is_accepted: boolean
  is_liked?: boolean
  created_at: string
}

// Docker镜像配置
export interface DockerImage {
  id: number
  name: string
  tag: string
  full_name?: string
  registry?: string
  category: string
  description?: string
  features?: string[]
  env_vars?: Record<string, unknown>
  ports?: number[]
  base_image?: string
  default_resources?: {
    cpu: number
    memory: string
    storage: string
  }
  status?: string
  is_active?: boolean
  is_built_in?: boolean
  created_at: string
}

export interface DockerImageCapability {
  id: number
  name: string
  tag: string
  full_name: string
  category: string
  status: string
  tool_keys: string[]
  default_ports: number[]
  default_resources: Record<string, string>
  features: string[]
  compatibility: string[]
}

// 镜像别名
export type Image = DockerImage

// 系统配置
export interface SystemConfig {
  key: string
  value: string
  description?: string
}

// 系统统计数据（对应后端 /system/stats）
export interface SystemStats {
  total_users: number
  total_schools: number
  total_courses: number
  total_experiments: number
  total_contests: number
  active_envs: number
  online_users: number
  server_uptime: string
  go_version: string
  num_goroutine: number
}

// 文件上传响应
export interface UploadResponse {
  url: string
  filename: string
  size: number
}

// 选项类型（用于下拉框等）
export interface SelectOption<T = string | number> {
  label: string
  value: T
  disabled?: boolean
}

// 表格列定义扩展
export interface TableColumnExtra {
  searchable?: boolean
  sortable?: boolean
  filterable?: boolean
  filterOptions?: SelectOption[]
}

// 跨校申请（对应后端 /system/cross-school）
export interface CrossSchoolApplication {
  id: number
  from_school_id: number
  to_school_id: number
  applicant_id: number
  type: string
  target_id: number
  target_type: string
  reason?: string
  status: 'pending' | 'approved' | 'rejected'
  reviewer_id?: number
  reviewed_at?: string
  reject_reason?: string
  created_at: string
  // 关联对象（后端 Preload 返回）
  from_school?: { id: number; name: string }
  to_school?: { id: number; name: string }
  applicant?: { id: number; real_name?: string; phone?: string; student_no?: string }
  reviewer?: { id: number; real_name?: string; phone?: string }
}

// 题目公开申请
export interface ChallengePublishApplication {
  id: number
  challenge_id: number
  challenge_title?: string
  applicant_id: number
  applicant_name?: string
  reason?: string
  status: 'pending' | 'approved' | 'rejected'
  created_at: string
}

// 漏洞数据（对齐后端 Vulnerability 模型 JSON tag）
export interface Vulnerability {
  id: number
  source_id: number
  external_id: string
  title: string
  description: string
  severity: string
  category: string
  technique: string
  chain: string
  amount: number
  attack_date: string
  reference: string
  contract_address?: string
  vuln_code?: string
  status: 'active' | 'converted' | 'skipped' | 'pending'
  converted_id?: number
  created_at: string
}

// 创建学校请求
export interface CreateSchoolRequest {
  name: string
  logo_url?: string
  contact_email?: string
  contact_phone?: string
  admin_phone: string
  admin_password: string
  admin_name: string
}

// 系统监控数据
export interface SystemMonitor {
  cpu_usage: number
  memory_usage: number
  memory_total: string
  memory_used: string
  disk_usage: number
  disk_total: string
  disk_used: string
  uptime: number
  load_average: number[]
}

// 容器信息
export interface ContainerInfo {
  id: string
  name: string
  status: string
  cpu_percent: number
  memory_usage: string
  created_at: string
}

// 容器统计
export interface ContainerStats {
  total: number
  running: number
  paused: number
  stopped: number
  containers: ContainerInfo[]
}

// 服务健康状态
export interface ServiceHealth {
  name: string
  status: 'healthy' | 'unhealthy' | 'unknown'
  latency?: number
  last_check: string
  message?: string
}

// 操作日志
export interface OperationLog {
  id: number
  user_id: number
  school_id?: number
  module: string
  action: string
  target_type?: string
  target_id?: number
  description?: string
  request_ip?: string
  user_agent?: string
  request_data?: Record<string, unknown>
  response_code?: number
  created_at: string
  // 关联对象
  user?: { id: number; real_name?: string; phone?: string; student_no?: string }
}
