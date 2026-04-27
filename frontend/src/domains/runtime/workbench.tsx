import type { ReactNode } from 'react'
import {
  ApiOutlined,
  BlockOutlined,
  CodeOutlined,
  ExperimentOutlined,
  FileSearchOutlined,
  FileTextOutlined,
  FolderOutlined,
} from '@ant-design/icons'

export type RuntimeWorkbenchToolKind =
  | 'ide'
  | 'terminal'
  | 'files'
  | 'explorer'
  | 'logs'
  | 'visualization'
  | 'api_debug'
  | 'network'
  | 'rpc'

export const RUNTIME_WORKBENCH_TOOL_ORDER: RuntimeWorkbenchToolKind[] = [
  'ide',
  'terminal',
  'files',
  'rpc',
  'explorer',
  'logs',
  'api_debug',
  'network',
  'visualization',
]

export const RUNTIME_WORKBENCH_TOOL_META: Record<RuntimeWorkbenchToolKind, { title: string; icon: ReactNode }> = {
  ide: { title: '在线编辑器', icon: <CodeOutlined /> },
  terminal: { title: '命令终端', icon: <FileTextOutlined /> },
  files: { title: '文件管理', icon: <FolderOutlined /> },
  explorer: { title: '区块浏览器', icon: <BlockOutlined /> },
  logs: { title: '日志', icon: <FileSearchOutlined /> },
  visualization: { title: '可视化', icon: <ExperimentOutlined /> },
  api_debug: { title: '接口调试台', icon: <ApiOutlined /> },
  network: { title: '节点协作面板', icon: <BlockOutlined /> },
  rpc: { title: '链上接口', icon: <ApiOutlined /> },
}

export function isRuntimeWorkbenchToolKind(value: string): value is RuntimeWorkbenchToolKind {
  return value in RUNTIME_WORKBENCH_TOOL_META
}

export function normalizeRuntimeWorkbenchToolKind(value?: string): RuntimeWorkbenchToolKind | undefined {
  if (!value) {
    return undefined
  }

  const normalized = value.trim().toLowerCase()
  return isRuntimeWorkbenchToolKind(normalized) ? normalized : undefined
}
