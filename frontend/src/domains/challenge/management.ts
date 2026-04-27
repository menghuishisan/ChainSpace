import type {
  Challenge,
  ChallengeCategory,
  CreateChallengeRequest,
  UpdateChallengeRequest,
} from '@/types'
import type { ChallengeListFilters, ChallengeManageFormValues } from '@/types/presentation'
import {
  buildChallengeServiceSpecs,
  buildDefaultChallengeOrchestration,
  getDefaultRuntimeProfile,
  normalizeChallengeOrchestration,
} from '@/domains/challenge/orchestration'

export const DEFAULT_CHALLENGE_FILTERS: ChallengeListFilters = {
  keyword: '',
  category: '',
  difficulty: '',
}

export function buildChallengeListQueryParams(
  page: number,
  pageSize: number,
  filters: ChallengeListFilters,
  extra?: {
    is_public?: boolean
  },
) {
  return {
    page,
    page_size: pageSize,
    keyword: filters.keyword.trim() || undefined,
    category: filters.category.trim() || undefined,
    difficulty: filters.difficulty && filters.difficulty.trim() ? Number(filters.difficulty) : undefined,
    ...extra,
  }
}

export function buildChallengeCreateInitialValues(): ChallengeManageFormValues {
  const category: ChallengeCategory = 'contract_vuln'
  const runtimeProfile = getDefaultRuntimeProfile(category)
  const orchestration = buildDefaultChallengeOrchestration(runtimeProfile, category)

  return {
    title: '',
    description: '',
    category,
    runtime_profile: runtimeProfile,
    difficulty: 3,
    base_points: 100,
    min_points: 50,
    decay_factor: 0.1,
    flag_type: 'dynamic',
    setup_code: '',
    deploy_script: '',
    check_script: '',
    hints: [],
    attachments: [],
    tags: [],
    is_public: false,
    challenge_orchestration: orchestration,
    service_keys: orchestration.services.map((service) => service.key),
  }
}

export function buildChallengeEditInitialValues(record: Challenge): ChallengeManageFormValues {
  const runtimeProfile = record.runtime_profile || getDefaultRuntimeProfile(record.category)
  const orchestration = normalizeChallengeOrchestration(record.challenge_orchestration, runtimeProfile, record.category)

  return {
    title: record.title,
    description: record.description,
    category: record.category,
    runtime_profile: runtimeProfile,
    difficulty: typeof record.difficulty === 'number' ? record.difficulty : 3,
    base_points: record.base_points,
    min_points: record.min_points,
    decay_factor: record.decay_factor,
    flag_type: record.flag_type,
    flag_template: record.flag_template,
    contract_code: record.contract_code,
    setup_code: record.setup_code,
    deploy_script: record.deploy_script,
    check_script: record.check_script,
    hints: record.hints,
    attachments: record.attachments,
    tags: record.tags,
    is_public: record.is_public,
    status: record.status,
    challenge_orchestration: orchestration,
    service_keys: orchestration.services.map((service) => service.key),
  }
}

export function buildChallengePresetValues(category: ChallengeCategory, runtimeProfile = getDefaultRuntimeProfile(category)) {
  const orchestration = normalizeChallengeOrchestration(undefined, runtimeProfile, category)

  return {
    runtime_profile: runtimeProfile,
    challenge_orchestration: orchestration,
    service_keys: orchestration.services.map((service) => service.key),
  }
}

export function buildChallengeSubmitData(values: ChallengeManageFormValues): CreateChallengeRequest {
  const serviceKeys = values.service_keys || []
  const orchestration = normalizeChallengeOrchestration(
    values.challenge_orchestration,
    values.runtime_profile,
    values.category,
  )
  const services = buildChallengeServiceSpecs(serviceKeys)

  return {
    title: values.title,
    description: values.description,
    category: values.category,
    runtime_profile: values.runtime_profile,
    difficulty: values.difficulty,
    base_points: values.base_points,
    min_points: values.min_points,
    decay_factor: values.decay_factor,
    flag_type: values.flag_type,
    flag_template: values.flag_template,
    contract_code: values.contract_code,
    setup_code: values.setup_code,
    deploy_script: values.deploy_script,
    check_script: values.check_script,
    hints: values.hints?.filter((hint) => hint.content.trim()),
    attachments: values.attachments?.filter((item) => item.trim()),
    tags: values.tags?.filter((item) => item.trim()),
    is_public: values.is_public ?? false,
    challenge_orchestration: {
      ...orchestration,
      services,
      topology: {
        ...orchestration.topology,
        mode: services.length > 0 ? 'workspace_with_services' : 'workspace_only',
        exposed_entries: orchestration.needs_environment ? ['workspace', 'rpc'] : ['workspace'],
      },
    },
  }
}

export function buildChallengeUpdateData(values: ChallengeManageFormValues): UpdateChallengeRequest {
  return {
    ...buildChallengeSubmitData(values),
    status: values.status,
  }
}
