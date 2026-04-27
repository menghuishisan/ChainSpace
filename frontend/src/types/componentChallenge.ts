import type { FormInstance } from 'antd'
import type { Challenge } from './contest'
import type { PaginatedData } from './common'
import type { ChallengeListFilters, ChallengeManageFormValues } from './pageChallenge'

export interface ChallengeTableProps {
  data: PaginatedData<Challenge>
  loading: boolean
  onEdit?: (challenge: Challenge) => void
  onView?: (challenge: Challenge) => void
  onDelete?: (challenge: Challenge) => void
  onRequestPublish?: (challenge: Challenge) => void
  canManage?: (challenge: Challenge) => boolean
  onPageChange: (page: number, pageSize: number) => void
}

export interface ChallengeManageModalProps {
  open: boolean
  loading: boolean
  editingChallenge: Challenge | null
  form: FormInstance<ChallengeManageFormValues>
  allowDirectPublic?: boolean
  onCancel: () => void
  onSubmit: (values: ChallengeManageFormValues) => Promise<void> | void
}

export interface ChallengeDetailModalProps {
  open: boolean
  challenge: Challenge | null
  onCancel: () => void
}

export interface ChallengeSearchFilterProps {
  values: ChallengeListFilters
  showDifficulty?: boolean
  onChange: (filters: ChallengeListFilters) => void
  onSearch: () => void
  onReset: () => void
}
