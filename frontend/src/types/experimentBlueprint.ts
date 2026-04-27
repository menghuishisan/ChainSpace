export type ExperimentMode = 'single' | 'multi_node' | 'collaboration'

export type ExperimentToolKey =
  | 'ide'
  | 'terminal'
  | 'files'
  | 'logs'
  | 'explorer'
  | 'visualization'
  | 'api_debug'
  | 'network'
  | 'rpc'

export interface ExperimentResourceBlueprint {
  cpu: string
  memory: string
  storage: string
}

export interface ExperimentWorkspaceBlueprint {
  image: string
  display_name?: string
  resources: ExperimentResourceBlueprint
  interaction_tools?: string[]
  init_scripts?: string[]
}

export interface ExperimentTopologyBlueprint {
  template?: string
  shared_network?: boolean
  exposed_entries?: string[]
}

export interface ExperimentToolBlueprint {
  key: string
  label?: string
  kind?: string
  target?: string
  student_facing?: boolean
}

export interface ExperimentRoleBindingBlueprint {
  key: string
  label?: string
  node_keys?: string[]
  tool_keys?: string[]
}

export interface ExperimentCollaborationBlueprint {
  max_members?: number
  roles?: ExperimentRoleBindingBlueprint[]
}

export interface ExperimentNodeBlueprint {
  key: string
  name?: string
  image: string
  role?: string
  ports?: number[]
  resources?: ExperimentResourceBlueprint
  student_facing?: boolean
  interaction_tools?: string[]
  init_scripts?: string[]
}

export interface ExperimentServiceBlueprint {
  key: string
  name?: string
  image: string
  role?: string
  purpose?: string
  ports?: number[]
  student_facing?: boolean
  env_vars?: Record<string, string>
}

export interface ExperimentContentBlueprintAsset {
  key: string
  name?: string
  source_type?: string
  bucket?: string
  object_path?: string
  target?: string
  mount_path?: string
  required?: boolean
}

export interface ExperimentContentBlueprint {
  assets?: ExperimentContentBlueprintAsset[]
  init_scripts?: string[]
}

export interface ExperimentCheckpointBlueprint {
  key: string
  type: string
  target?: string
  path?: string
  command?: string
  expected?: string
  script?: string
  score?: number
}

export interface ExperimentGradingBlueprint {
  strategy?: string
  checkpoints?: ExperimentCheckpointBlueprint[]
}

export interface ExperimentBlueprint {
  mode?: ExperimentMode
  workspace: ExperimentWorkspaceBlueprint
  topology?: ExperimentTopologyBlueprint
  tools?: ExperimentToolBlueprint[]
  collaboration?: ExperimentCollaborationBlueprint
  nodes?: ExperimentNodeBlueprint[]
  services?: ExperimentServiceBlueprint[]
  content?: ExperimentContentBlueprint
  grading?: ExperimentGradingBlueprint
}

export interface TopologyTemplateOption {
  key: string
  label: string
  description: string
}

export interface ServiceTemplateOption {
  key: string
  name: string
  image: string
  role: string
  purpose: string
  ports: number[]
  student_facing?: boolean
}

export interface VisualizationModuleOption {
  key: string
  label: string
  description: string
  simulator_id?: string
}

export interface ExperimentEditorFormState {
  course_id?: number
  chapter_id?: number
  title: string
  description: string
  estimated_time: number
  type: import('./experiment').ExperimentType
  max_score: number
  blueprint: ExperimentBlueprint
}
