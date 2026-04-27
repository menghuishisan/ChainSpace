import type { Dayjs } from 'dayjs'

import type { BattleOrchestration } from './contestOrchestration'
import type { Class, Contest, Team } from '@/types'

export interface LoginFormValues {
  phone: string
  password: string
  remember: boolean
}

export type SchoolClassItem = Pick<Class, 'id' | 'name'>

export interface SchoolSettingsFormValues {
  name: string
  code: string
  logo_url?: string
  description?: string
  contact_email?: string
  contact_phone?: string
  address?: string
  allow_cross_school_contest?: boolean
  allow_public_courses?: boolean
  max_students?: number
  max_teachers?: number
}

export interface TeamWithContest extends Team {
  contest?: Contest
}

export interface TeacherCourseStatRow {
  courseId: number
  courseName: string
  studentCount: number
  experimentCount: number
}

export interface TeacherStatisticsSummary {
  courseCount: number
  studentCount: number
  experimentCount: number
}

export interface ContestChallengeStat {
  id: number
  title: string
  points: number
  solve_count: number
  first_blood?: string
  first_blood_time?: string
}

export interface ContestRoundInfo {
  id: number
  round_number: number
  status: string
  phase?: string
  upgrade_window_end?: string
  start_time?: string
  end_time?: string
}

export interface ContestFormValues {
  name: string
  description?: string
  type: 'jeopardy' | 'agent_battle'
  time_range: [Dayjs, Dayjs]
  registration_start?: Dayjs
  registration_end?: Dayjs
  cover?: string
  rules?: string
  is_public?: boolean
  max_participants?: number
  team_size_min?: number
  team_size_max?: number
  dynamic_score?: boolean
  first_blood_bonus?: number
  battle_orchestration?: BattleOrchestration
}

export interface ContestListQueryParams {
  page?: number
  page_size?: number
  keyword?: string
  type?: string
  level?: string
  status?: string
}
