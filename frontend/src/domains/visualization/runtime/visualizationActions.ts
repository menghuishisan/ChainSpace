import type {
  VisualizationActionDefinition,
  VisualizationRuntimeSpec,
} from '@/types/visualizationDomain'
import { VISUALIZATION_CATALOG } from './visualizationRegistry'

function inferScope(runtime: VisualizationRuntimeSpec): VisualizationActionDefinition['scope'] {
  switch (runtime.renderer) {
    case 'consensus':
    case 'attack':
    case 'blockchain':
    case 'network':
    case 'crypto':
    case 'crosschain':
    case 'evm':
    case 'defi':
      return runtime.renderer
    default:
      return undefined
  }
}

function inferGroup(runtime: VisualizationRuntimeSpec, definition: VisualizationActionDefinition): string {
  if (definition.group) {
    return definition.group
  }

  if (definition.kind === 'module_action') {
    switch (runtime.renderer) {
      case 'consensus':
        return '协议推进'
      case 'attack':
        return '攻击演示'
      case 'defi':
        return '协议操作'
      case 'crosschain':
        return '跨链流程'
      case 'network':
        return '网络操作'
      default:
        return '实验动作'
    }
  }

  if (definition.kind === 'inject_fault') {
    return runtime.renderer === 'consensus' ? '故障与攻击联动' : '故障注入'
  }

  return '攻击联动'
}

function inferOverlayLabel(
  runtime: VisualizationRuntimeSpec,
  definition: VisualizationActionDefinition,
): string | undefined {
  if (definition.overlayLabel) {
    return definition.overlayLabel
  }

  if (definition.kind === 'inject_fault') {
    return runtime.renderer === 'consensus' ? '共识过程扰动' : '故障覆盖层'
  }

  if (definition.kind === 'inject_attack') {
    return '攻击覆盖层'
  }

  return undefined
}

export function getVisualizationActions(runtime: VisualizationRuntimeSpec): VisualizationActionDefinition[] {
  const catalogEntry = VISUALIZATION_CATALOG[runtime.moduleKey]
  if (!catalogEntry?.getActions) {
    return []
  }

  return catalogEntry.getActions().map((definition) => ({
    ...definition,
    scope: inferScope(runtime),
    group: inferGroup(runtime, definition),
    overlayLabel: inferOverlayLabel(runtime, definition),
  }))
}
