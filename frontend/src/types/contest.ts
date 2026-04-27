/**
 * 比赛与挑战题目相关类型定义。
 * 这里统一承载解题赛、对抗赛、观战和回放相关的领域契约。
 */

import type { BattleOrchestration, ChallengeOrchestration } from './contestOrchestration'

export type ContestType = 'jeopardy' | 'agent_battle'
export type ContestLevel = 'practice' | 'school' | 'cross_school' | 'platform'
export type ContestStatus = 'draft' | 'published' | 'ongoing' | 'ended'

export type ChallengeDifficulty = 1 | 2 | 3 | 4 | 5

export type ChallengeCategory =
  | 'contract_vuln'
  | 'defi'
  | 'consensus'
  | 'crypto'
  | 'cross_chain'
  | 'nft'
  | 'reverse'
  | 'key_management'
  | 'misc'

export type ChallengeRuntimeProfile =
  | 'static'
  | 'single_chain_instance'
  | 'fork_replay'
  | 'multi_service_lab'

export type FlagType = 'static' | 'dynamic'
export type ChallengeSourceType = 'preset' | 'auto_converted' | 'user_created'
export type ChallengeStatus = 'draft' | 'active' | 'archived'

export interface Contest {
  id: number
  school_id?: number
  creator_id: number
  creator_name?: string
  title: string
  name?: string
  description?: string
  type: ContestType
  level?: ContestLevel
  cover?: string
  rules?: string
  start_time: string
  end_time: string
  registration_start?: string
  registration_end?: string
  dynamic_score?: boolean
  first_blood_bonus?: number
  max_participants?: number
  team_min_size?: number
  team_max_size?: number
  battle_orchestration?: BattleOrchestration
  status: ContestStatus
  is_public?: boolean
  created_at: string
  participant_count?: number
  challenge_count?: number
  is_registered?: boolean
}

export interface Challenge {
  id: number
  creator_id: number
  creator_name?: string
  school_id?: number
  title: string
  description: string
  category: ChallengeCategory
  runtime_profile: ChallengeRuntimeProfile
  difficulty: ChallengeDifficulty
  base_points: number
  points?: number
  min_points?: number
  decay_factor?: number
  flag_type: FlagType
  flag_template?: string
  contract_code?: string
  setup_code?: string
  deploy_script?: string
  check_script?: string
  validation_config?: Record<string, unknown>
  challenge_orchestration: ChallengeOrchestration
  hints?: ChallengeHint[]
  attachments?: string[]
  tags?: string[]
  source_type?: ChallengeSourceType
  source_ref?: string
  status: ChallengeStatus
  is_public: boolean
  solve_count?: number
  attempt_count?: number
  created_at: string
  is_solved?: boolean
  awarded_points?: number
  first_blood?: string
  first_blood_time?: string
}

export interface ChallengeHint {
  content: string
  cost: number
}

export interface CreateChallengeRequest {
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
  challenge_orchestration: ChallengeOrchestration
  hints?: ChallengeHint[]
  attachments?: string[]
  tags?: string[]
  is_public: boolean
}

export interface UpdateChallengeRequest extends Partial<CreateChallengeRequest> {
  status?: ChallengeStatus
}

export interface ContestChallenge {
  id: number
  contest_id: number
  challenge_id: number
  points: number
  current_points: number
  sort_order: number
  is_visible: boolean
  challenge: Challenge
  is_solved?: boolean
  solve_count?: number
}

export interface Team {
  id: number
  contest_id: number
  name: string
  captain_id: number
  leader_name?: string
  avatar?: string
  description?: string
  status: 'forming' | 'ready' | 'competing'
  invite_code?: string
  member_count?: number
  members?: TeamMember[]
  created_at: string
}

export interface TeamMember {
  id: number
  team_id: number
  user_id: number
  display_name?: string
  real_name?: string
  phone?: string
  student_no?: string
  avatar?: string
  role: 'captain' | 'member'
  is_captain?: boolean
  joined_at: string
}

export interface ContestScore {
  rank: number
  user_id?: number
  display_name?: string
  team_id?: number
  team_name?: string
  total_score: number
  solve_count: number
  last_solve_at?: string
  first_blood_count?: number
}

export interface Scoreboard {
  list: ContestScore[]
  my_rank?: number
  my_score?: number
}

export interface AgentBattleScore {
  rank: number
  team_id: number
  team_name: string
  score: number
  token_balance: string
  success_count: number
  fail_count: number
  resource_held: number
}

export interface ChallengeEnv {
  id: number
  env_id: string
  contest_id: number
  challenge_id: number
  status: string
  access_url?: string
  tools?: Array<{
    key: string
    label: string
    kind?: string
    target?: string
    port: number
    route: string
    instance_route?: string
  }>
  service_entries?: Array<{
    key: string
    label: string
    description?: string
    purpose?: string
    access_url?: string
    protocol?: string
    port?: number
    expose_as?: string
  }>
  started_at?: string
  expires_at?: string
  error_message?: string
  remaining: number
}

export interface CategoryGroup {
  category: string
  label: string
  challenges: Challenge[]
}

export interface TeamContractInfo {
  id: number
  contract_address: string
  status: string
  version: number
  deployed_at?: string
}

export type BattleRoundPhase =
  | 'pending'
  | 'upgrade_window'
  | 'locked'
  | 'executing'
  | 'settling'
  | 'finished'

export interface CurrentRoundInfo {
  id: number
  round_number: number
  status: string
  phase?: BattleRoundPhase | string
  start_time?: string
  end_time?: string
  upgrade_window_end?: string
}

export interface FlagSubmitResult {
  correct: boolean
  points: number
  message: string
}

export interface AgentBattleRound {
  round_number: number
  start_time: string
  end_time: string
  upgrade_window_end?: string
  status: string
  phase?: BattleRoundPhase | string
}

export interface AgentBattleTeamStatus {
  team_id: number
  team_name: string
  contract_address?: string
  is_alive: boolean
  resource_held: number
  total_score: number
  score_change?: string
}

export interface AgentBattleEvent {
  block: number
  round_number?: number
  time?: string
  event_type: string
  actor_team?: string
  target_team?: string
  action_result?: string
  score_delta?: number
  resource_delta?: number
  points?: number
  description?: string
}

export type BattleStatus = 'waiting' | 'running' | 'paused' | 'completed'

export interface AgentBattleStatus {
  status?: BattleStatus
  current_block?: number
  current_round?: number
  total_rounds?: number
  round_phase?: BattleRoundPhase | string
  my_rank?: number
  my_score?: number
  teams?: AgentBattleTeamStatus[]
  recent_events?: AgentBattleEvent[]
  agent_status?: {
    version: string
    uploaded_at: string
    is_valid: boolean
  }
}

export interface SpectateData {
  contest_name: string
  current_round: number
  current_block?: number
  round_status: string
  round_phase?: BattleRoundPhase | string
  round_end_time?: string
  teams?: AgentBattleTeamStatus[]
  recent_events?: AgentBattleEvent[]
  spectator_count?: number
  available_battles?: Array<{ id: string; name: string }>
  current_battle?: {
    round: number
    red_team: string
    blue_team: string
    current_turn: number
  }
  visualization_data?: Record<string, unknown>
}

export const ContestTypeMap: Record<ContestType, string> = {
  jeopardy: '解题赛',
  agent_battle: '智能体博弈战',
}

export const ContestLevelMap: Record<ContestLevel, string> = {
  practice: '练习赛',
  school: '校内赛',
  cross_school: '跨校赛',
  platform: '平台赛',
}

export const ContestStatusMap: Record<ContestStatus, { text: string; color: string }> = {
  draft: { text: '草稿', color: 'default' },
  published: { text: '已发布', color: 'processing' },
  ongoing: { text: '进行中', color: 'success' },
  ended: { text: '已结束', color: 'default' },
}

export const ChallengeRuntimeProfileMap: Record<ChallengeRuntimeProfile, string> = {
  static: '静态题',
  single_chain_instance: '链上实例题',
  fork_replay: 'Fork 复现题',
  multi_service_lab: '多服务拓扑题',
}

export const DifficultyMap: Record<string | number, { text: string; color: string; stars: number }> = {
  1: { text: '入门', color: 'blue', stars: 1 },
  2: { text: '简单', color: 'green', stars: 2 },
  3: { text: '中等', color: 'orange', stars: 3 },
  4: { text: '困难', color: 'red', stars: 4 },
  5: { text: '专家', color: 'purple', stars: 5 },
  easy: { text: '简单', color: 'green', stars: 1 },
  medium: { text: '中等', color: 'orange', stars: 2 },
  hard: { text: '困难', color: 'red', stars: 3 },
  expert: { text: '专家', color: 'purple', stars: 4 },
}

export const CategoryMap: Record<string, string> = {
  contract_vuln: '合约漏洞',
  defi: 'DeFi',
  consensus: '共识机制',
  crypto: '密码学',
  cross_chain: '跨链安全',
  nft: 'NFT/代币安全',
  reverse: '逆向分析',
  key_management: '密钥安全',
  misc: '其他',
}

export const BattleStatusMap: Record<BattleStatus, { text: string; color: string }> = {
  waiting: { text: '等待中', color: 'default' },
  running: { text: '进行中', color: 'success' },
  paused: { text: '已暂停', color: 'warning' },
  completed: { text: '已完成', color: 'blue' },
}

export const BattleRoundPhaseMap: Record<BattleRoundPhase, { text: string; color: string }> = {
  pending: { text: '待开始', color: 'default' },
  upgrade_window: { text: '升级窗口', color: 'processing' },
  locked: { text: '锁定阶段', color: 'warning' },
  executing: { text: '执行阶段', color: 'success' },
  settling: { text: '结算阶段', color: 'purple' },
  finished: { text: '已结束', color: 'default' },
}

export interface CreateContestRequest {
  title: string
  description?: string
  type: ContestType
  level?: string
  cover?: string
  rules?: string
  start_time: string
  end_time: string
  registration_start?: string
  registration_end?: string
  dynamic_score?: boolean
  first_blood_bonus?: number
  max_participants?: number
  team_min_size?: number
  team_max_size?: number
  is_public?: boolean
  battle_orchestration?: BattleOrchestration
}

export interface ContestRecord {
  contest_id: number
  contest_name: string
  contest_type: string
  team_name?: string
  rank?: number
  total_score: number
  status: ContestStatus
}
