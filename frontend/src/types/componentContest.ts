import type { ReactNode } from 'react'
import type {
  AgentBattleEvent,
  AgentBattleStatus,
  CategoryGroup,
  Challenge,
  ChallengeEnv,
  Contest,
  ContestScore,
  CurrentRoundInfo,
  TeamContractInfo,
} from './contest'
import type { BattleOrchestration } from './contestOrchestration'

export type ContestWorkbenchTabKey = string

export interface ContestTableProps {
  data: import('./common').PaginatedData<Contest>
  loading: boolean
  emptyDescription?: string
  showEndTime?: boolean
  onView: (contest: Contest) => void
  onEdit: (contest: Contest) => void
  onPublish: (contestId: number) => void
  onPageChange: (page: number, pageSize: number) => void
}

export interface JeopardyChallengeGridProps {
  categoryGroups: CategoryGroup[]
  selectedChallengeId?: number
  runtimeLabels: Record<string, string>
  difficultyMap: Record<string | number, { text: string; color: string; stars: number }>
  compact?: boolean
  onSelectChallenge: (challenge: Challenge) => void
}

export interface JeopardySidebarProps {
  selectedChallenge: Challenge | null
  categoryMap: Record<string, string>
  runtimeLabels: Record<string, string>
  difficultyMap: Record<string | number, { text: string; color: string; stars: number }>
  renderedDescription: string
  showDescriptionCard?: boolean
  firstBloodBonus?: number
  challengeEnv: ChallengeEnv | null
  envLoading: boolean
  envRemaining: number
  fetchingEnv: boolean
  flagInput: string
  submitting: boolean
  onRefreshEnv: () => void
  onStartEnv: () => void
  onStopEnv: () => void
  onOpenCodeViewer: (code: string, title: string) => void
  onCopyCode: (code: string) => void
  onDownloadCode: (filename: string, code: string) => void
  onOpenAttachment: (attachmentIndex: number) => void
  onFlagInputChange: (value: string) => void
  onSubmitFlag: () => void
}

export interface AgentBattleSummaryProps {
  contest: Contest
  battleStatus: AgentBattleStatus | null
  battleConfig: BattleOrchestration
  currentRound: CurrentRoundInfo | null
  contractInfo: TeamContractInfo | null
  sourceCode: string
  deploying: boolean
  uploading: boolean
  remainingTime: number
  onSourceCodeChange: (value: string) => void
  onDeploy: () => void
  onUploadFile: NonNullable<import('antd').UploadProps['customRequest']>
  onSpectate: () => void
}

export interface AgentBattleSidebarProps {
  scoreWeights: Record<string, number>
  scoreboard: ContestScore[]
  recentEvents: AgentBattleEvent[]
  emptyState?: ReactNode
}
