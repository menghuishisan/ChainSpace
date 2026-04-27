import type {
  ChallengeCategory,
  ChallengeOrchestration,
  ChallengeRuntimeProfile,
  ChallengeServiceSpec,
} from '@/types'

export const CHALLENGE_RUNTIME_PROFILE_OPTIONS = [
  { label: '静态题', value: 'static' },
  { label: '单链实例题', value: 'single_chain_instance' },
  { label: 'Fork 复现题', value: 'fork_replay' },
  { label: '多服务拓扑题', value: 'multi_service_lab' },
]

export const CHALLENGE_SERVICE_OPTIONS = [
  { label: 'Anvil Fork 服务', value: 'anvil_fork' },
  { label: 'Chainlink 预言机服务', value: 'chainlink' },
  { label: 'Geth 节点服务', value: 'geth' },
  { label: 'Blockscout 区块浏览器', value: 'blockscout' },
  { label: 'The Graph 索引服务', value: 'thegraph' },
]

const challengeServiceTemplateMap: Record<string, ChallengeServiceSpec> = {
  anvil_fork: {
    key: 'anvil_fork',
    image: 'chainspace/eth-dev:latest',
    purpose: 'fork_replay',
    description: '为 Fork 复现场景提供本地 RPC 服务。',
  },
  chainlink: {
    key: 'chainlink',
    image: 'chainspace/chainlink:latest',
    purpose: 'oracle',
    description: '提供预言机数据与链下服务模拟。',
    ports: [
      { name: 'api', port: 6688, protocol: 'http', expose_as: 'api_debug' },
    ],
  },
  geth: {
    key: 'geth',
    image: 'chainspace/geth:latest',
    purpose: 'node_cluster',
    description: '提供链节点、共识与网络交互能力。',
    ports: [
      { name: 'rpc', port: 8545, protocol: 'http', expose_as: 'rpc' },
      { name: 'ws', port: 8546, protocol: 'ws', expose_as: 'rpc' },
    ],
  },
  blockscout: {
    key: 'blockscout',
    image: 'chainspace/blockscout:latest',
    purpose: 'explorer',
    description: '提供链上区块、交易、地址与合约状态浏览能力。',
    ports: [
      { name: 'web', port: 4000, protocol: 'http', expose_as: 'explorer' },
    ],
  },
  thegraph: {
    key: 'thegraph',
    image: 'chainspace/thegraph:latest',
    purpose: 'indexer',
    description: '提供链上数据索引与查询能力。',
    ports: [
      { name: 'graphql', port: 8000, protocol: 'http', expose_as: 'api_debug' },
    ],
  },
}

const runtimeProfileServiceMap: Record<ChallengeRuntimeProfile, string[]> = {
  static: [],
  single_chain_instance: [],
  fork_replay: ['anvil_fork'],
  multi_service_lab: ['geth'],
}

export function getDefaultRuntimeProfile(category: ChallengeCategory = 'contract_vuln'): ChallengeRuntimeProfile {
  switch (category) {
    case 'crypto':
    case 'reverse':
      return 'static'
    case 'defi':
      return 'fork_replay'
    case 'consensus':
    case 'cross_chain':
      return 'multi_service_lab'
    default:
      return 'single_chain_instance'
  }
}

function getDefaultWorkspaceImage(runtimeProfile: ChallengeRuntimeProfile): string {
  if (runtimeProfile === 'static') {
    return 'chainspace/crypto:latest'
  }
  return 'chainspace/eth-dev:latest'
}

function getDefaultWorkspaceLabel(runtimeProfile: ChallengeRuntimeProfile): string {
  switch (runtimeProfile) {
    case 'static':
      return '静态分析工作区'
    case 'fork_replay':
      return 'Fork 复现工作区'
    case 'multi_service_lab':
      return '多服务拓扑工作区'
    default:
      return '单链实例工作区'
  }
}

function getDefaultTemplate(runtimeProfile: ChallengeRuntimeProfile): string {
  switch (runtimeProfile) {
    case 'static':
      return 'static_analysis'
    case 'fork_replay':
      return 'fork_replay'
    case 'multi_service_lab':
      return 'workspace_with_services'
    default:
      return 'single_chain_instance'
  }
}

export function buildChallengeServiceSpecs(serviceKeys: string[]): ChallengeServiceSpec[] {
  return serviceKeys
    .map((key) => challengeServiceTemplateMap[key])
    .filter((service): service is ChallengeServiceSpec => Boolean(service))
    .map((service) => ({ ...service }))
}

export function getDefaultServiceKeys(runtimeProfile: ChallengeRuntimeProfile, category: ChallengeCategory): string[] {
  if (runtimeProfile === 'multi_service_lab' && category === 'cross_chain') {
    return ['geth', 'blockscout']
  }
  if (runtimeProfile === 'fork_replay' && category === 'defi') {
    return ['chainlink']
  }
  return [...runtimeProfileServiceMap[runtimeProfile]]
}

function getDefaultInteractionTools(runtimeProfile: ChallengeRuntimeProfile): string[] {
  switch (runtimeProfile) {
    case 'fork_replay':
      return ['ide', 'terminal', 'files', 'logs']
    case 'multi_service_lab':
      return ['ide', 'terminal', 'files', 'logs']
    case 'static':
      return []
    default:
      return ['ide', 'terminal', 'files']
  }
}

function getDefaultExposedEntries(runtimeProfile: ChallengeRuntimeProfile): string[] {
  switch (runtimeProfile) {
    case 'fork_replay':
      return ['workspace', 'rpc']
    case 'multi_service_lab':
      return ['workspace']
    case 'static':
      return ['workspace']
    default:
      return ['workspace', 'rpc']
  }
}

export function buildDefaultChallengeOrchestration(
  runtimeProfile: ChallengeRuntimeProfile,
  category: ChallengeCategory = 'contract_vuln',
): ChallengeOrchestration {
  const serviceKeys = getDefaultServiceKeys(runtimeProfile, category)
  const services = buildChallengeServiceSpecs(serviceKeys)
  const needsEnvironment = runtimeProfile !== 'static'

  return {
    mode: runtimeProfile,
    needs_environment: needsEnvironment,
    workspace: {
      image: getDefaultWorkspaceImage(runtimeProfile),
      display_name: getDefaultWorkspaceLabel(runtimeProfile),
      template: getDefaultTemplate(runtimeProfile),
      interaction_tools: getDefaultInteractionTools(runtimeProfile),
      resources: {
        cpu: '2',
        memory: '4Gi',
        storage: '10Gi',
      },
      init_scripts: [],
    },
    services,
    topology: {
      mode: services.length > 0 ? 'workspace_with_services' : 'workspace_only',
      exposed_entries: needsEnvironment ? getDefaultExposedEntries(runtimeProfile) : ['workspace'],
      shared_network: runtimeProfile === 'multi_service_lab',
    },
    fork: {
      enabled: runtimeProfile === 'fork_replay',
      chain: 'ethereum',
      chain_id: 1,
      rpc_url: 'https://eth-mainnet.g.alchemy.com/v2/${ALCHEMY_KEY}',
    },
    scenario: {
      attack_goal: '根据题目说明完成漏洞复现、状态分析或系统攻防验证。',
      init_steps: [],
      solve_steps: [],
      defense_goal: '',
    },
    lifecycle: {
      time_limit_minutes: 120,
      auto_destroy: true,
      reuse_running_env: true,
    },
    validation: {
      mode: needsEnvironment ? 'service' : 'static',
      description: needsEnvironment
        ? '通过环境中的服务、脚本或交互结果验证解题过程与最终结果。'
        : '通过附件、答案或静态分析结果验证解题过程。',
    },
  }
}

export function normalizeChallengeOrchestration(
  input: Partial<ChallengeOrchestration> | undefined,
  runtimeProfile: ChallengeRuntimeProfile,
  category: ChallengeCategory = 'contract_vuln',
): ChallengeOrchestration {
  const base = buildDefaultChallengeOrchestration(runtimeProfile, category)

  return {
    ...base,
    ...input,
    mode: runtimeProfile,
    needs_environment: runtimeProfile !== 'static',
    workspace: {
      ...base.workspace,
      ...(input?.workspace || {}),
      interaction_tools: input?.workspace?.interaction_tools?.length
        ? input.workspace.interaction_tools
        : base.workspace.interaction_tools,
      resources: {
        ...(base.workspace.resources || {}),
        ...(input?.workspace?.resources || {}),
      },
      init_scripts: input?.workspace?.init_scripts || base.workspace.init_scripts,
    },
    services: input?.services?.length ? input.services : base.services,
    topology: {
      ...base.topology,
      ...(input?.topology || {}),
      exposed_entries: input?.topology?.exposed_entries?.length
        ? input.topology.exposed_entries
        : base.topology.exposed_entries,
    },
    fork: {
      ...base.fork,
      ...(input?.fork || {}),
      enabled: runtimeProfile === 'fork_replay',
    },
    scenario: {
      ...base.scenario,
      ...(input?.scenario || {}),
      init_steps: input?.scenario?.init_steps || base.scenario.init_steps,
      solve_steps: input?.scenario?.solve_steps || base.scenario.solve_steps,
    },
    lifecycle: {
      ...base.lifecycle,
      ...(input?.lifecycle || {}),
    },
    validation: {
      ...base.validation,
      ...(input?.validation || {}),
    },
  }
}
