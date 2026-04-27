/**
 * 比赛环境编排相关类型定义。
 * 这里专门承载解题赛与对抗赛的环境/赛制模型，避免和基础比赛类型混在一起。
 */

// 解题赛题目的主工作区配置。
export interface ChallengeWorkspaceSpec {
  image: string
  display_name?: string
  template?: string
  interaction_tools: string[]
  resources?: Record<string, string>
  init_scripts?: string[]
}

// 解题赛题目的附加服务端口。
export interface ChallengeServicePort {
  name: string
  port: number
  protocol: string
  expose_as?: string
}

// 解题赛题目的附加服务组件。
export interface ChallengeServiceSpec {
  key: string
  image: string
  purpose?: string
  description?: string
  ports?: ChallengeServicePort[]
  env?: Record<string, {} | undefined>
}

// 解题赛题目的拓扑说明。
export interface ChallengeTopologySpec {
  mode: string
  exposed_entries: string[]
  shared_network?: boolean
}

// 解题赛题目的 Fork 复现配置。
export interface ChallengeForkSpec {
  enabled: boolean
  chain?: string
  chain_id?: number
  label?: string
  rpc_url?: string
  block_number?: number
  target_tx_hash?: string
}

// 解题赛题目的场景说明。
export interface ChallengeScenarioSpec {
  contract_address?: string
  attack_goal?: string
  init_steps?: string[]
  solve_steps?: string[]
  defense_goal?: string
}

// 解题赛题目的生命周期配置。
export interface ChallengeLifecycleSpec {
  time_limit_minutes: number
  auto_destroy: boolean
  reuse_running_env: boolean
}

// 解题赛题目的验证配置。
export interface ChallengeValidationSpec {
  mode?: string
  description?: string
}

// 解题赛题目的正式环境编排模型。
export interface ChallengeOrchestration {
  mode: string
  needs_environment: boolean
  workspace: ChallengeWorkspaceSpec
  services: ChallengeServiceSpec[]
  topology: ChallengeTopologySpec
  fork: ChallengeForkSpec
  scenario: ChallengeScenarioSpec
  lifecycle: ChallengeLifecycleSpec
  validation: ChallengeValidationSpec
}

// 对抗赛共享链配置。
export interface BattleSharedChainSpec {
  image?: string
  chain_type?: string
  network_id?: number
  block_time?: number
  initial_balances?: Record<string, string>
  rpc_url?: string
}

// 对抗赛裁判与规则配置。
export interface BattleJudgeSpec {
  image?: string
  judge_contract?: string
  token_contract?: string
  strategy_interface?: string
  resource_model?: string
  scoring_model?: string
  score_weights?: Record<string, number>
  allowed_actions?: string[]
  forbidden_calls?: string[]
}

// 对抗赛队伍工作区模板。
export interface BattleTeamWorkspaceSpec {
  image?: string
  display_name?: string
  interaction_tools?: string[]
  resources?: Record<string, string>
}

// 对抗赛观战配置。
export interface BattleSpectateSpec {
  enable_monitor?: boolean
  enable_replay?: boolean
}

// 对抗赛生命周期配置。
export interface BattleLifecycleSpec {
  round_duration_seconds?: number
  upgrade_window_seconds?: number
  total_rounds?: number
  auto_cleanup?: boolean
}

// 对抗赛正式编排模型。
export interface BattleOrchestration {
  shared_chain: BattleSharedChainSpec
  judge: BattleJudgeSpec
  team_workspace: BattleTeamWorkspaceSpec
  spectate: BattleSpectateSpec
  lifecycle: BattleLifecycleSpec
}
