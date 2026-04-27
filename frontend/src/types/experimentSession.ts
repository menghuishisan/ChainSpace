export type EnvStatus = 'pending' | 'creating' | 'running' | 'paused' | 'terminated' | 'failed'

export type RuntimeInstanceKind = 'workspace' | 'node' | 'service'

export interface ExperimentRuntimeTool {
  id: number
  runtime_instance_id: number
  tool_key: string
  port: number
  sort_order: number
  created_at: string
  updated_at: string
}

export interface ExperimentRuntimeInstance {
  id: number
  experiment_env_id: number
  instance_key: string
  kind: RuntimeInstanceKind | string
  pod_name: string
  status: EnvStatus
  student_facing: boolean
  tools?: ExperimentRuntimeTool[]
  created_at: string
  updated_at: string
}

export interface ExperimentEnvTool {
  key: string
  label: string
  kind?: string
  module_key?: string
  target?: string
  port: number
  route: string
  ws_route?: string
  instance_route?: string
}

export interface ExperimentSessionMember {
  user_id: number
  display_name?: string
  real_name?: string
  phone?: string
  role_key?: string
  assigned_node_key?: string
  join_status: string
  joined_at: string
}

export interface ExperimentSession {
  id: number
  session_key: string
  experiment_id: number
  mode: import('./experimentBlueprint').ExperimentMode | string
  status: string
  primary_env_id?: string
  max_members: number
  current_member_count: number
  started_at?: string
  expires_at?: string
  members?: ExperimentSessionMember[]
}

export interface ExperimentSessionMessage {
  id: number
  session_id: number
  user_id: number
  display_name?: string
  real_name?: string
  phone?: string
  message: string
  message_type: string
  created_at: string
}

export interface ExperimentEnv {
  id: number
  env_id: string
  experiment_id: number
  experiment_title?: string
  user_id: number
  display_name?: string
  status: EnvStatus
  session?: ExperimentSession
  session_mode?: import('./experimentBlueprint').ExperimentMode | string
  primary_instance_key?: string
  instances?: ExperimentRuntimeInstance[]
  tools?: ExperimentEnvTool[]
  started_at?: string
  expires_at?: string
  extend_count: number
  snapshot_url?: string
  error_message?: string
  created_at: string
}

export interface WorkspaceLogEntry {
  timestamp: string
  level: 'info' | 'warn' | 'error' | 'debug' | string
  source: string
  message: string
}
