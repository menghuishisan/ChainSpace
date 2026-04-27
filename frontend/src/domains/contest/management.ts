import dayjs from 'dayjs'

import type {
  BattleOrchestration,
  Contest,
  CreateContestRequest,
} from '@/types'
import type {
  ContestFormValues,
  ContestListQueryParams,
  SearchFilterItem,
} from '@/types/presentation'
import { FILTER_OPTIONS } from '@/utils/constants'

export interface ContestListFilters extends Record<string, unknown> {
  keyword: string
  type: string
  status: string
}

export interface ContestPagination {
  page: number
  page_size: number
}

export const DEFAULT_CONTEST_LIST_FILTERS: ContestListFilters = {
  keyword: '',
  type: '',
  status: '',
}

export const DEFAULT_CONTEST_PAGINATION: ContestPagination = {
  page: 1,
  page_size: 20,
}

export const CONTEST_FILTER_CONFIG: SearchFilterItem[] = [
  { key: 'keyword', label: '关键字', type: 'input', placeholder: '搜索比赛名称' },
  { key: 'type', label: '类型', type: 'select', options: [...FILTER_OPTIONS.CONTEST_TYPE] },
  { key: 'status', label: '状态', type: 'select', options: [...FILTER_OPTIONS.CONTEST_STATUS] },
]

export const BATTLE_STRATEGY_OPTIONS = [
  { label: '标准策略接口 v1', value: 'strategy_agent_v1' },
]

export const BATTLE_RESOURCE_MODEL_OPTIONS = [
  { label: '共享资源控制', value: 'shared_resource_control' },
]

export const BATTLE_SCORING_MODEL_OPTIONS = [
  { label: '资源 / 攻击 / 防御 / 生存复合评分', value: 'resource_attack_defense_survival' },
]

function normalizeBattleInteractionTools(tools?: string[]): string[] {
  const allowed = new Set(['ide', 'terminal', 'files', 'logs', 'explorer', 'api_debug', 'visualization', 'network', 'rpc'])
  const values = (tools || [])
    .map((item) => item.trim())
    .filter((item) => Boolean(item) && allowed.has(item))
  return values.filter((item, index) => values.indexOf(item) === index)
}

export function buildDefaultBattleOrchestration(): BattleOrchestration {
  return {
    shared_chain: {
      image: 'chainspace/eth-dev:latest',
      chain_type: 'anvil',
      network_id: 31337,
      block_time: 2,
    },
    judge: {
      strategy_interface: 'strategy_agent_v1',
      resource_model: 'shared_resource_control',
      scoring_model: 'resource_attack_defense_survival',
      score_weights: {
        resource: 35,
        attack: 30,
        defense: 20,
        survival: 15,
      },
      allowed_actions: ['gather', 'attack', 'defend', 'fortify', 'recover', 'scout'],
    },
    team_workspace: {
      image: 'chainspace/eth-dev:latest',
      display_name: '策略开发工作区',
      interaction_tools: ['ide', 'terminal', 'files', 'logs', 'api_debug'],
      resources: {
        cpu: '2',
        memory: '4Gi',
      },
    },
    spectate: {
      enable_monitor: true,
      enable_replay: true,
    },
    lifecycle: {
      round_duration_seconds: 300,
      upgrade_window_seconds: 120,
      total_rounds: 5,
      auto_cleanup: true,
    },
  }
}

export function normalizeBattleOrchestration(input?: BattleOrchestration): BattleOrchestration {
  const defaults = buildDefaultBattleOrchestration()

  return {
    shared_chain: {
      ...defaults.shared_chain,
      ...(input?.shared_chain || {}),
    },
    judge: {
      ...defaults.judge,
      ...(input?.judge || {}),
      score_weights: {
        ...(defaults.judge.score_weights || {}),
        ...(input?.judge?.score_weights || {}),
      },
      allowed_actions: input?.judge?.allowed_actions?.length
        ? input.judge.allowed_actions
        : defaults.judge.allowed_actions,
      forbidden_calls: input?.judge?.forbidden_calls?.length
        ? input.judge.forbidden_calls
        : defaults.judge.forbidden_calls,
    },
    team_workspace: {
      ...defaults.team_workspace,
      ...(input?.team_workspace || {}),
      interaction_tools: input?.team_workspace?.interaction_tools?.length
        ? normalizeBattleInteractionTools(input.team_workspace.interaction_tools)
        : defaults.team_workspace.interaction_tools,
      resources: {
        ...(defaults.team_workspace.resources || {}),
        ...(input?.team_workspace?.resources || {}),
      },
    },
    spectate: {
      ...defaults.spectate,
      ...(input?.spectate || {}),
    },
    lifecycle: {
      ...defaults.lifecycle,
      ...(input?.lifecycle || {}),
    },
  }
}

export function buildContestListQueryParams(
  pagination: ContestPagination,
  filters: ContestListFilters,
  level?: string,
): ContestListQueryParams {
  const queryParams: ContestListQueryParams = { ...pagination }

  if (level) {
    queryParams.level = level
  }
  if (filters.keyword.trim()) {
    queryParams.keyword = filters.keyword.trim()
  }
  if (filters.type.trim()) {
    queryParams.type = filters.type
  }
  if (filters.status.trim()) {
    queryParams.status = filters.status
  }

  return queryParams
}

export function normalizeContestFilters(values: Record<string, unknown>): ContestListFilters {
  return {
    keyword: typeof values.keyword === 'string' ? values.keyword : '',
    type: typeof values.type === 'string' ? values.type : '',
    status: typeof values.status === 'string' ? values.status : '',
  }
}

export function getContestSearchFilterItems(
  options: {
    includeDraft?: boolean
  } = {},
): SearchFilterItem[] {
  const { includeDraft = true } = options

  return [
    { key: 'keyword', label: '搜索', type: 'input', placeholder: '搜索比赛名称' },
    { key: 'type', label: '类型', type: 'select', options: [...FILTER_OPTIONS.CONTEST_TYPE] },
    {
      key: 'status',
      label: '状态',
      type: 'select',
      options: includeDraft
        ? [...FILTER_OPTIONS.CONTEST_STATUS]
        : FILTER_OPTIONS.CONTEST_STATUS.filter((option) => option.value !== 'draft'),
    },
  ]
}

export function buildContestFormInitialValues(record: Contest): ContestFormValues {
  return {
    name: record.title,
    description: record.description,
    type: record.type,
    time_range: [dayjs(record.start_time), dayjs(record.end_time)],
    registration_start: record.registration_start ? dayjs(record.registration_start) : undefined,
    registration_end: record.registration_end ? dayjs(record.registration_end) : undefined,
    cover: record.cover,
    rules: record.rules,
    is_public: record.is_public,
    max_participants: record.max_participants,
    team_size_min: record.team_min_size,
    team_size_max: record.team_max_size,
    dynamic_score: record.dynamic_score,
    first_blood_bonus: record.first_blood_bonus,
    battle_orchestration: record.type === 'agent_battle'
      ? normalizeBattleOrchestration(record.battle_orchestration)
      : undefined,
  }
}

export function buildContestSubmitData(
  values: ContestFormValues,
  level: CreateContestRequest['level'],
): CreateContestRequest {
  return {
    title: values.name,
    description: values.description,
    type: values.type,
    level,
    cover: values.cover,
    rules: values.rules,
    start_time: values.time_range[0].toISOString(),
    end_time: values.time_range[1].toISOString(),
    registration_start: values.registration_start ? values.registration_start.toISOString() : undefined,
    registration_end: values.registration_end ? values.registration_end.toISOString() : undefined,
    is_public: values.is_public,
    max_participants: values.max_participants,
    team_min_size: values.team_size_min,
    team_max_size: values.team_size_max,
    dynamic_score: values.dynamic_score,
    first_blood_bonus: values.first_blood_bonus,
    battle_orchestration: values.type === 'agent_battle'
      ? normalizeBattleOrchestration(values.battle_orchestration)
      : undefined,
  }
}
