import type { ChallengeCategory, ChallengeRuntimeProfile, ChallengeStatus, FlagType } from './contest'
import type { ChallengeOrchestration } from './contestOrchestration'

// 题目列表筛选值。
export interface ChallengeListFilters extends Record<string, unknown> {
  keyword: string
  category: string
  difficulty?: string
}

// 题目管理表单值。
export interface ChallengeManageFormValues {
  title: string
  description: string
  category: ChallengeCategory
  runtime_profile: ChallengeRuntimeProfile
  difficulty: number
  base_points: number
  min_points?: number
  decay_factor?: number
  flag_type: FlagType
  flag_template?: string
  contract_code?: string
  setup_code?: string
  deploy_script?: string
  check_script?: string
  hints?: {
    content: string
    cost: number
  }[]
  attachments?: string[]
  tags?: string[]
  is_public?: boolean
  status?: ChallengeStatus
  service_keys?: string[]
  challenge_orchestration: ChallengeOrchestration
}
