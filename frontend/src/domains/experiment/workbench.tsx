import type { ReactNode } from 'react'

import type { ExperimentEnv, ExperimentEnvTool, ExperimentRuntimeInstance } from '@/types'
import {
  type RuntimeWorkbenchToolKind,
  RUNTIME_WORKBENCH_TOOL_META,
  normalizeRuntimeWorkbenchToolKind,
} from '@/domains/runtime/workbench'
import { API_SERVER_ORIGIN } from '@/utils/constants'

export type ExperimentWorkbenchToolKey = RuntimeWorkbenchToolKind

export interface ExperimentWorkbenchToolView {
  key: ExperimentWorkbenchToolKey
  title: string
  icon: ReactNode
  label: string
  kind?: string
  moduleKey?: string
  route: string
  accessUrl: string
  wsUrl?: string
  instanceRoute?: string
  instanceAccessUrl?: string
  target?: string
}

function normalizeToolRoute(tool: ExperimentEnvTool): string {
  return tool.instance_route || tool.route
}

function resolveToolAccessUrl(route?: string): string {
  return route ? `${API_SERVER_ORIGIN}${route}` : ''
}

function resolveToolLabel(key: ExperimentWorkbenchToolKey, rawLabel?: string): string {
  const fallback = RUNTIME_WORKBENCH_TOOL_META[key].title
  const label = rawLabel?.trim()
  if (!label) {
    return fallback
  }

  const normalized = label.toLowerCase()
  if (normalized === key || normalized === normalized.replace(/_/g, '') || normalized.includes('_')) {
    return fallback
  }

  return label
}

export function buildWorkbenchTools(env: ExperimentEnv | null): ExperimentWorkbenchToolView[] {
  if (!env?.tools?.length) {
    return []
  }

  return env.tools.reduce<ExperimentWorkbenchToolView[]>((tools, tool) => {
      const key = normalizeRuntimeWorkbenchToolKind(tool.key) || normalizeRuntimeWorkbenchToolKind(tool.kind)
      if (!key) {
        return tools
      }

      tools.push({
        key,
        title: RUNTIME_WORKBENCH_TOOL_META[key].title,
        icon: RUNTIME_WORKBENCH_TOOL_META[key].icon,
        label: resolveToolLabel(key, tool.label),
        kind: tool.kind,
        moduleKey: tool.module_key,
        route: normalizeToolRoute(tool),
        accessUrl: resolveToolAccessUrl(normalizeToolRoute(tool)),
        wsUrl: resolveToolAccessUrl(tool.ws_route),
        instanceRoute: tool.instance_route,
        instanceAccessUrl: resolveToolAccessUrl(tool.instance_route),
        target: tool.target,
      })
      return tools
    }, [])
}

export function buildWorkbenchInstances(env: ExperimentEnv | null): ExperimentRuntimeInstance[] {
  return env?.instances?.filter((instance) => instance.student_facing) || []
}

export function resolveWorkbenchToolUrl(route?: string): string {
  return resolveToolAccessUrl(route)
}

export function getWorkbenchActiveTab(
  availableTools: ExperimentWorkbenchToolView[],
  activeTab?: ExperimentWorkbenchToolKey,
): ExperimentWorkbenchToolKey | undefined {
  const validTabKeys = availableTools.map((tool) => tool.key)
  if (activeTab && validTabKeys.includes(activeTab)) {
    return activeTab
  }

  return validTabKeys[0]
}

export function getWorkbenchIdeToolUrl(availableTools: ExperimentWorkbenchToolView[]): string {
  return availableTools.find((item) => item.key === 'ide')?.accessUrl || ''
}

export function isWorkbenchEnvRunning(status?: string | null): boolean {
  return status === 'running'
}

export function shouldProbeIdeRuntime(status: string | null | undefined, ideToolUrl: string): boolean {
  return isWorkbenchEnvRunning(status) && Boolean(ideToolUrl)
}

export function getWorkbenchVisualizationModuleKey(tool?: ExperimentWorkbenchToolView): string | undefined {
  return tool?.key === 'visualization' ? tool.moduleKey : undefined
}
