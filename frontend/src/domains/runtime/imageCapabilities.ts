import type { DockerImageCapability } from '@/types'

export function getCapabilityImageRef(capability: DockerImageCapability): string {
  return capability.full_name || `${capability.name}:${capability.tag}`
}

export function normalizeCapabilityMode(value?: string): string {
  return (value || '').trim().toLowerCase()
}

export function capabilitySupportsAllTools(capability: DockerImageCapability, toolKeys: string[]): boolean {
  if (!toolKeys.length) {
    return true
  }
  const supported = new Set((capability.tool_keys || []).map((key) => key.trim()).filter(Boolean))
  return toolKeys.every((key) => supported.has(key))
}

export function capabilitySupportsAnyTool(capability: DockerImageCapability, toolKeys: string[]): boolean {
  if (!toolKeys.length) {
    return true
  }
  const supported = new Set((capability.tool_keys || []).map((key) => key.trim()).filter(Boolean))
  return toolKeys.some((key) => supported.has(key))
}

export function capabilitySupportsMode(capability: DockerImageCapability, modes: string[]): boolean {
  if (!modes.length) {
    return true
  }
  const compat = new Set((capability.compatibility || []).map((mode) => normalizeCapabilityMode(mode)).filter(Boolean))
  return modes.some((mode) => compat.has(normalizeCapabilityMode(mode)))
}

export function describeCapability(capability: DockerImageCapability): string {
  const compat = (capability.compatibility || []).join(' / ')
  const tools = (capability.tool_keys || []).join(', ')
  const parts = [compat ? `兼容: ${compat}` : '', tools ? `工具: ${tools}` : ''].filter(Boolean)
  return parts.join(' | ')
}

