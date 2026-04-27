import type { ReactNode } from 'react'
import type { Experiment } from './experiment'
import type {
  ExperimentCheckpointBlueprint,
  ExperimentCollaborationBlueprint,
  ExperimentContentBlueprintAsset,
  ExperimentEditorFormState,
  ExperimentNodeBlueprint,
} from './experimentBlueprint'
import type { DockerImage } from './common'
import type {
  EnvStatus,
  ExperimentRuntimeInstance,
  ExperimentSession,
  ExperimentSessionMember,
  ExperimentSessionMessage,
} from './experimentSession'

export interface FileItem {
  name: string
  type: 'file' | 'directory'
  size: number
  modified_at: string
  path: string
}

export interface FileManagerProps {
  experimentId?: number
  accessUrl?: string
  onFileOpen?: (file: FileItem) => void
}

export interface BlockExplorerProps {
  accessUrl?: string
}

export interface ApiDebuggerProps {
  accessUrl?: string
  title?: string
  description?: string
}

export interface ExperimentVisualizationPanelProps {
  experiment: Experiment | null
  accessUrl?: string
  wsUrl?: string
  moduleKey?: string
}

export interface LogEntry {
  id: string
  timestamp: string
  level: 'info' | 'warn' | 'error' | 'debug'
  source: string
  message: string
}

export interface LogViewerProps {
  accessUrl?: string
  sources?: string[]
}

export type TerminalConnectionStatus = 'connecting' | 'connected' | 'disconnected' | 'error'

export interface TerminalProps {
  experimentId: number
  accessUrl?: string
  wsUrl?: string
}

export interface ExperimentBlueprintEditorProps {
  formData: ExperimentEditorFormState
  images: DockerImage[]
  selectedWorkspaceImage?: DockerImage
  selectedServiceKeys: string[]
  selectedVisualizationModule?: string
  visualizationModuleOptions?: Array<{ label: string; value: string }>
  infoMessage: ReactNode
  infoDescription: ReactNode
  showWorkspaceResources?: boolean
  onWorkspaceImageChange: (image: string) => void
  onWorkspaceResourceChange?: (key: 'cpu' | 'memory' | 'storage', value: string) => void
  onInteractionToolsChange: (tools: string[]) => void
  onServicesChange: (keys: string[]) => void
  onTopologyTemplateChange: (template: string) => void
  onVisualizationModuleChange?: (moduleKey: string) => void
  onWorkspaceInitScriptsChange?: (scripts: string[]) => void
  onContentInitScriptsChange?: (scripts: string[]) => void
  onContentAssetsChange?: (assets: ExperimentContentBlueprintAsset[]) => void
  onNodesChange?: (nodes: ExperimentNodeBlueprint[]) => void
  onGradingStrategyChange?: (strategy: string) => void
  onCheckpointsChange?: (checkpoints: ExperimentCheckpointBlueprint[]) => void
  onCollaborationChange?: (collaboration: ExperimentCollaborationBlueprint) => void
  onAssetUpload?: (file: File) => Promise<ExperimentContentBlueprintAsset | undefined>
}

export type ExperimentWorkbenchTabKey =
  | 'ide'
  | 'terminal'
  | 'files'
  | 'explorer'
  | 'logs'
  | 'visualization'
  | 'api_debug'
  | 'network'
  | 'rpc'

export interface ExperimentWorkbenchToolDisplay {
  key: ExperimentWorkbenchTabKey
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

export interface ExperimentWorkbenchProps {
  experimentId: number
  experiment: Experiment | null
  session: ExperimentSession | null
  instances: ExperimentRuntimeInstance[]
  currentMember?: ExperimentSessionMember | null
  sessionMembers: ExperimentSessionMember[]
  sessionMessages: ExperimentSessionMessage[]
  canManageSessionMembers?: boolean
  envStatus: EnvStatus | null
  remainingSeconds: number
  availableTools: ExperimentWorkbenchToolDisplay[]
  activeTab?: ExperimentWorkbenchTabKey
  mountedTabs: ExperimentWorkbenchTabKey[]
  ideReady: boolean
  submitModalVisible: boolean
  submitting: boolean
  submitReport: string
  snapshotUrl?: string
  envErrorMessage?: string
  onSetActiveTab: (key: ExperimentWorkbenchTabKey) => void
  onExtend: () => void
  onPause: () => void
  onResume: () => void
  onCreateSnapshot: () => void
  onRestoreSnapshot?: () => void
  onStop: () => void
  onUpdateSessionMember?: (userId: number, payload: {
    role_key?: string
    assigned_node_key?: string
    join_status?: 'joined' | 'left'
  }) => void
  onExit: () => void
  onOpenSubmit: () => void
  onCloseSubmit: () => void
  onSubmitReportChange: (value: string) => void
  onSubmitExperiment: () => void
  onSendMessage: (message: string) => void
}
