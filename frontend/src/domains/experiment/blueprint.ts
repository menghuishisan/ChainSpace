import type {
  DockerImage,
  ExperimentBlueprint,
  ExperimentEditorFormState,
  ExperimentServiceBlueprint,
  ExperimentToolBlueprint,
  ExperimentToolKey,
  ExperimentType,
  ServiceTemplateOption,
  TopologyTemplateOption,
} from '@/types'
import { VISUALIZATION_MODULE_OPTIONS } from '@/domains/visualization/runtime/visualizationRegistry'

export const TOPOLOGY_TEMPLATES: TopologyTemplateOption[] = [
  { key: 'workspace_only', label: '单工作区', description: '仅提供一个学生工作区实例，适合代码开发、命令操作与单体实验。' },
  { key: 'workspace_with_services', label: '工作区 + 服务', description: '工作区搭配链节点、浏览器、调试端口等辅助服务。' },
  { key: 'multi_role_lab', label: '多节点实验室', description: '适合多节点、多角色、协作式实验运行拓扑。' },
]

export const SERVICE_TEMPLATES: ServiceTemplateOption[] = [
  {
    key: 'simulation',
    name: '可视化模拟器',
    image: 'chainspace/simulation:latest',
    role: 'visualization',
    purpose: '为可视化实验提供 simulations 服务。',
    ports: [8080],
    student_facing: true,
  },
  {
    key: 'geth',
    name: '以太坊节点',
    image: 'chainspace/geth:latest',
    role: 'rpc',
    purpose: '提供链节点与 RPC 能力。',
    ports: [8545],
    student_facing: true,
  },
  {
    key: 'ipfs',
    name: 'IPFS 节点',
    image: 'chainspace/ipfs:latest',
    role: 'storage',
    purpose: '提供分布式存储实验能力。',
    ports: [5001],
  },
]

const defaultWorkspaceImageByType: Record<ExperimentType, string> = {
  visualization: 'chainspace/eth-dev:latest',
  code_dev: 'chainspace/eth-dev:latest',
  command_op: 'chainspace/eth-dev:latest',
  data_analysis: 'chainspace/eth-dev:latest',
  tool_usage: 'chainspace/security:latest',
  config_debug: 'chainspace/eth-dev:latest',
  reverse: 'chainspace/security:latest',
  troubleshoot: 'chainspace/security:latest',
  collaboration: 'chainspace/eth-dev:latest',
}

export const WORKSPACE_TOOL_OPTIONS: Array<{ label: string; value: ExperimentToolKey }> = [
  { label: 'IDE', value: 'ide' },
  { label: '终端', value: 'terminal' },
  { label: 'RPC', value: 'rpc' },
  { label: '文件', value: 'files' },
  { label: '区块浏览器', value: 'explorer' },
  { label: '日志', value: 'logs' },
  { label: '可视化', value: 'visualization' },
  { label: 'API 调试', value: 'api_debug' },
  { label: '组网面板', value: 'network' },
]

export const VISUALIZATION_TOOL_OPTIONS = VISUALIZATION_MODULE_OPTIONS.map((option) => ({
  label: option.label,
  value: option.key,
}))

export const DEFAULT_VISUALIZATION_MODULE_KEY = VISUALIZATION_TOOL_OPTIONS[0]?.value || 'blockchain/block_structure'

function defaultWorkspaceTools(type: ExperimentType): ExperimentToolKey[] {
  return type === 'visualization'
    ? ['terminal', 'logs', 'visualization']
    : ['ide', 'terminal', 'files', 'logs']
}

function findToolByKey(
  tools: ExperimentToolBlueprint[] | undefined,
  key: string,
  target?: string,
): ExperimentToolBlueprint | undefined {
  return tools?.find((tool) => tool.key === key && (target ? tool.target === target : true))
}

function deriveServiceToolBlueprints(
  services: ExperimentServiceBlueprint[],
  existingTools: ExperimentToolBlueprint[] = [],
): ExperimentToolBlueprint[] {
  const result: ExperimentToolBlueprint[] = []

  for (const service of services) {
    switch (service.key) {
      case 'simulation':
        result.push({
          key: 'visualization',
          label: 'visualization',
          kind: findToolByKey(existingTools, 'visualization', service.key)?.kind || DEFAULT_VISUALIZATION_MODULE_KEY,
          target: service.key,
          student_facing: true,
        })
        break
      case 'geth':
        result.push({ key: 'rpc', label: 'rpc', target: service.key, student_facing: true })
        break
      case 'ipfs':
        result.push({ key: 'api_debug', label: 'api_debug', target: service.key, student_facing: true })
        break
      default:
        break
    }
  }

  return result
}

function hasService(services: ExperimentServiceBlueprint[], key: string): boolean {
  return services.some((service) => service.key === key)
}

function resolveDraftToolTarget(
  toolKey: string,
  target: string,
  services: ExperimentServiceBlueprint[],
): string {
  if (toolKey === 'visualization') {
    return 'simulation'
  }

  if (['rpc', 'explorer', 'api_debug'].includes(toolKey) && hasService(services, 'geth')) {
    return 'geth'
  }

  if (toolKey === 'api_debug' && hasService(services, 'ipfs')) {
    return 'ipfs'
  }

  return target
}

export function createDefaultBlueprint(type: ExperimentType): ExperimentBlueprint {
  const services: ExperimentServiceBlueprint[] = type === 'visualization'
    ? [{ ...SERVICE_TEMPLATES.find((item) => item.key === 'simulation')! }]
    : []

  return normalizeExperimentBlueprintDraft({
    mode: type === 'collaboration' ? 'collaboration' : undefined,
    workspace: {
      image: defaultWorkspaceImageByType[type],
      resources: { cpu: '500m', memory: '512Mi', storage: '1Gi' },
      interaction_tools: defaultWorkspaceTools(type),
    },
    topology: {
      template: services.length > 0 ? 'workspace_with_services' : 'workspace_only',
    },
    services,
    tools: [],
    nodes: [],
    content: { assets: [], init_scripts: [] },
    grading: { strategy: 'checkpoint', checkpoints: [] },
  }, type)
}

export function createExperimentEditorFormState(type: ExperimentType = 'code_dev'): ExperimentEditorFormState {
  return {
    course_id: undefined,
    chapter_id: undefined,
    title: '',
    description: '',
    estimated_time: 60,
    type,
    max_score: 100,
    blueprint: createDefaultBlueprint(type),
  }
}

export function normalizeExperimentBlueprintDraft(
  blueprint: ExperimentBlueprint,
  type: ExperimentType,
): ExperimentBlueprint {
  const workspaceTools = (blueprint.workspace.interaction_tools || [])
    .filter((tool): tool is ExperimentToolKey =>
      WORKSPACE_TOOL_OPTIONS.some((option) => option.value === tool as ExperimentToolKey) || tool === 'visualization',
    )

  const services = blueprint.services || []
  const existingTools = blueprint.tools || []
  const tools: ExperimentToolBlueprint[] = [
    ...workspaceTools.map((tool) => ({
      key: tool,
      label: tool,
      kind: findToolByKey(existingTools, tool, 'workspace')?.kind,
      target: resolveDraftToolTarget(tool, 'workspace', services),
      student_facing: true,
    })),
    ...deriveServiceToolBlueprints(services, existingTools),
    ...existingTools.map((tool) => ({
      ...tool,
      target: resolveDraftToolTarget(tool.key, tool.target || 'workspace', services),
    })),
  ].filter((tool, index, list) => (
    list.findIndex((candidate) => candidate.key === tool.key && candidate.target === tool.target) === index
  ))

  return {
    ...blueprint,
    mode: blueprint.mode || (type === 'collaboration' ? 'collaboration' : undefined),
    workspace: {
      ...blueprint.workspace,
      interaction_tools: workspaceTools.length > 0 ? workspaceTools : defaultWorkspaceTools(type),
      resources: {
        cpu: blueprint.workspace.resources.cpu || '500m',
        memory: blueprint.workspace.resources.memory || '512Mi',
        storage: blueprint.workspace.resources.storage || '1Gi',
      },
      init_scripts: blueprint.workspace.init_scripts || [],
    },
    topology: {
      template: blueprint.topology?.template || (services.length > 0 ? 'workspace_with_services' : 'workspace_only'),
      shared_network: blueprint.topology?.shared_network || type === 'collaboration',
      exposed_entries: blueprint.topology?.exposed_entries || ['workspace'],
    },
    services,
    nodes: blueprint.nodes || [],
    tools,
    content: {
      assets: blueprint.content?.assets || [],
      init_scripts: blueprint.content?.init_scripts || [],
    },
    grading: {
      strategy: blueprint.grading?.strategy || 'checkpoint',
      checkpoints: blueprint.grading?.checkpoints || [],
    },
    collaboration: blueprint.collaboration,
  }
}

export function getSelectedVisualizationModule(formData: ExperimentEditorFormState): string {
  return formData.blueprint.tools?.find((tool) => tool.key === 'visualization')?.kind || DEFAULT_VISUALIZATION_MODULE_KEY
}

export function getImageLabel(image: DockerImage): string {
  return image.full_name || `${image.name}:${image.tag}`
}

export function describeImage(image: DockerImage): string {
  const parts = [image.description, image.features?.slice(0, 3).join(' / ')].filter(Boolean)
  return parts.join(' | ') || '暂无镜像用途说明'
}
