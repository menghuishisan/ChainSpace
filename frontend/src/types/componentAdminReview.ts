import type { ChallengePublishApplication, CrossSchoolApplication } from '@/types'

export interface CrossSchoolReviewState {
  list: CrossSchoolApplication[]
  total: number
  page: number
  page_size: number
}

export interface ChallengePublishReviewState {
  list: ChallengePublishApplication[]
  total: number
  page: number
  page_size: number
}
