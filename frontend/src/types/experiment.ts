import type { ExperimentBlueprint } from './experimentBlueprint'
import type { EnvStatus } from './experimentSession'

export type ExperimentType =
  | 'visualization'
  | 'code_dev'
  | 'command_op'
  | 'data_analysis'
  | 'tool_usage'
  | 'config_debug'
  | 'reverse'
  | 'troubleshoot'
  | 'collaboration'

export type ExperimentStatus = 'draft' | 'published' | 'archived'
export type SubmissionStatus = 'pending' | 'grading' | 'graded'

export interface Experiment {
  id: number
  school_id: number
  course_id?: number
  chapter_id: number
  chapter_title?: string
  creator_id: number
  creator_name?: string
  title: string
  description: string
  type: ExperimentType
  mode: import('./experimentBlueprint').ExperimentMode
  difficulty: number
  estimated_time: number
  max_score: number
  pass_score: number
  auto_grade: boolean
  blueprint: ExperimentBlueprint
  sort_order: number
  status: ExperimentStatus
  start_time?: string
  end_time?: string
  allow_late: boolean
  late_deduction: number
  submission_count?: number
  my_score?: number
  my_status?: string
  created_at: string
}

export interface SubmissionCheckResult {
  id: number
  submission_id: number
  checkpoint_key: string
  checkpoint_type: string
  target?: string
  passed: boolean
  score: number
  details?: string
  sort_order: number
  created_at: string
  updated_at: string
}

export interface Submission {
  id: number
  experiment_id: number
  experiment_title?: string
  student_id: number
  student_name?: string
  env_id?: string
  content?: string
  file_url?: string
  snapshot_url?: string
  score?: number | null
  auto_score?: number | null
  manual_score?: number | null
  feedback?: string
  check_results?: SubmissionCheckResult[]
  status: SubmissionStatus
  submitted_at: string
  graded_at?: string
  grader_name?: string
  is_late: boolean
  attempt_number: number
}

export interface CreateExperimentRequest {
  chapter_id: number
  title: string
  description?: string
  type: ExperimentType
  difficulty?: number
  estimated_time?: number
  max_score?: number
  pass_score?: number
  auto_grade?: boolean
  blueprint: ExperimentBlueprint
  sort_order?: number
  start_time?: string
  end_time?: string
  allow_late?: boolean
  late_deduction?: number
}

export interface UpdateExperimentRequest extends Partial<CreateExperimentRequest> {
  status?: ExperimentStatus
}

export const ExperimentTypeMap: Record<ExperimentType, string> = {
  visualization: '可视化交互类',
  code_dev: '代码开发类',
  command_op: '命令操作类',
  data_analysis: '数据分析类',
  tool_usage: '工具使用类',
  config_debug: '配置调试类',
  reverse: '逆向分析类',
  troubleshoot: '故障排查类',
  collaboration: '多人协作类',
}

export const ExperimentTypeIcon: Record<ExperimentType, string> = {
  visualization: 'eye',
  code_dev: 'code',
  command_op: 'terminal',
  data_analysis: 'bar-chart',
  tool_usage: 'wrench',
  config_debug: 'settings',
  reverse: 'search',
  troubleshoot: 'bug',
  collaboration: 'users',
}

export const EnvStatusMap: Record<EnvStatus, { text: string; color: string }> = {
  pending: { text: '待启动', color: 'default' },
  creating: { text: '创建中', color: 'processing' },
  running: { text: '运行中', color: 'success' },
  paused: { text: '已暂停', color: 'warning' },
  terminated: { text: '已终止', color: 'default' },
  failed: { text: '启动失败', color: 'error' },
}
