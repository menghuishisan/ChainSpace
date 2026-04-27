package model

type ExperimentWorkspace struct {
	BaseModel
	ExperimentID uint   `gorm:"uniqueIndex;not null" json:"experiment_id"`
	Image        string `gorm:"size:200;not null" json:"image"`
	DisplayName  string `gorm:"size:100" json:"display_name"`
	CPU          string `gorm:"size:20;not null" json:"cpu"`
	Memory       string `gorm:"size:20;not null" json:"memory"`
	Storage      string `gorm:"size:20;not null" json:"storage"`

	Tools []ExperimentWorkspaceTool `gorm:"foreignKey:WorkspaceID" json:"tools,omitempty"`
}

func (ExperimentWorkspace) TableName() string {
	return "experiment_workspaces"
}

type ExperimentWorkspaceTool struct {
	BaseModel
	WorkspaceID uint   `gorm:"index;not null" json:"workspace_id"`
	ToolKey     string `gorm:"size:50;not null" json:"tool_key"`
	SortOrder   int    `gorm:"default:0" json:"sort_order"`
}

func (ExperimentWorkspaceTool) TableName() string {
	return "experiment_workspace_tools"
}

type ExperimentTopology struct {
	BaseModel
	ExperimentID  uint   `gorm:"uniqueIndex;not null" json:"experiment_id"`
	Template      string `gorm:"size:50;not null" json:"template"`
	SharedNetwork bool   `gorm:"default:false" json:"shared_network"`

	ExposedEntries []ExperimentTopologyExposedEntry `gorm:"foreignKey:TopologyID" json:"exposed_entries,omitempty"`
}

func (ExperimentTopology) TableName() string {
	return "experiment_topologies"
}

type ExperimentTopologyExposedEntry struct {
	BaseModel
	TopologyID uint   `gorm:"index;not null" json:"topology_id"`
	EntryKey   string `gorm:"size:50;not null" json:"entry_key"`
	SortOrder  int    `gorm:"default:0" json:"sort_order"`
}

func (ExperimentTopologyExposedEntry) TableName() string {
	return "experiment_topology_exposed_entries"
}

type ExperimentTool struct {
	BaseModel
	ExperimentID   uint   `gorm:"index;not null" json:"experiment_id"`
	ToolKey        string `gorm:"size:50;not null" json:"tool_key"`
	Label          string `gorm:"size:100" json:"label"`
	Kind           string `gorm:"size:100" json:"kind"`
	Target         string `gorm:"size:50" json:"target"`
	StudentFacing  bool   `gorm:"default:false" json:"student_facing"`
	SortOrder      int    `gorm:"default:0" json:"sort_order"`
}

func (ExperimentTool) TableName() string {
	return "experiment_tools"
}

type ExperimentInitScript struct {
	BaseModel
	ExperimentID uint   `gorm:"index;not null" json:"experiment_id"`
	ScopeType    string `gorm:"size:20;not null" json:"scope_type"`
	ScopeKey     string `gorm:"size:50" json:"scope_key"`
	Script       string `gorm:"type:text;not null" json:"script"`
	SortOrder    int    `gorm:"default:0" json:"sort_order"`
}

func (ExperimentInitScript) TableName() string {
	return "experiment_init_scripts"
}

type ExperimentCollaboration struct {
	BaseModel
	ExperimentID uint   `gorm:"uniqueIndex;not null" json:"experiment_id"`
	MaxMembers   int    `gorm:"default:4" json:"max_members"`

	Roles []ExperimentRoleBinding `gorm:"foreignKey:CollaborationID" json:"roles,omitempty"`
}

func (ExperimentCollaboration) TableName() string {
	return "experiment_collaborations"
}

type ExperimentRoleBinding struct {
	BaseModel
	CollaborationID uint   `gorm:"index;not null" json:"collaboration_id"`
	RoleKey         string `gorm:"size:50;not null" json:"role_key"`
	Label           string `gorm:"size:100" json:"label"`
	SortOrder       int    `gorm:"default:0" json:"sort_order"`

	NodeAssignments []ExperimentRoleBindingNode `gorm:"foreignKey:RoleBindingID" json:"node_assignments,omitempty"`
	ToolAssignments []ExperimentRoleBindingTool `gorm:"foreignKey:RoleBindingID" json:"tool_assignments,omitempty"`
}

func (ExperimentRoleBinding) TableName() string {
	return "experiment_role_bindings"
}

type ExperimentRoleBindingNode struct {
	BaseModel
	RoleBindingID uint   `gorm:"index;not null" json:"role_binding_id"`
	NodeKey       string `gorm:"size:50;not null" json:"node_key"`
	SortOrder     int    `gorm:"default:0" json:"sort_order"`
}

func (ExperimentRoleBindingNode) TableName() string {
	return "experiment_role_binding_nodes"
}

type ExperimentRoleBindingTool struct {
	BaseModel
	RoleBindingID uint   `gorm:"index;not null" json:"role_binding_id"`
	ToolKey       string `gorm:"size:50;not null" json:"tool_key"`
	SortOrder     int    `gorm:"default:0" json:"sort_order"`
}

func (ExperimentRoleBindingTool) TableName() string {
	return "experiment_role_binding_tools"
}

type ExperimentNode struct {
	BaseModel
	ExperimentID   uint   `gorm:"index;not null" json:"experiment_id"`
	NodeKey        string `gorm:"size:50;not null" json:"node_key"`
	Name           string `gorm:"size:100" json:"name"`
	Image          string `gorm:"size:200;not null" json:"image"`
	Role           string `gorm:"size:50" json:"role"`
	CPU            string `gorm:"size:20;not null" json:"cpu"`
	Memory         string `gorm:"size:20;not null" json:"memory"`
	Storage        string `gorm:"size:20;not null" json:"storage"`
	StudentFacing  bool   `gorm:"default:false" json:"student_facing"`
	SortOrder      int    `gorm:"default:0" json:"sort_order"`

	Ports []ExperimentNodePort `gorm:"foreignKey:NodeID" json:"ports,omitempty"`
	Tools []ExperimentNodeTool `gorm:"foreignKey:NodeID" json:"tools,omitempty"`
}

func (ExperimentNode) TableName() string {
	return "experiment_nodes"
}

type ExperimentNodePort struct {
	BaseModel
	NodeID    uint  `gorm:"index;not null" json:"node_id"`
	Port      int32 `gorm:"not null" json:"port"`
	SortOrder int   `gorm:"default:0" json:"sort_order"`
}

func (ExperimentNodePort) TableName() string {
	return "experiment_node_ports"
}

type ExperimentNodeTool struct {
	BaseModel
	NodeID     uint   `gorm:"index;not null" json:"node_id"`
	ToolKey    string `gorm:"size:50;not null" json:"tool_key"`
	SortOrder  int    `gorm:"default:0" json:"sort_order"`
}

func (ExperimentNodeTool) TableName() string {
	return "experiment_node_tools"
}

type ExperimentService struct {
	BaseModel
	ExperimentID   uint   `gorm:"index;not null" json:"experiment_id"`
	ServiceKey     string `gorm:"size:50;not null" json:"service_key"`
	Name           string `gorm:"size:100" json:"name"`
	Image          string `gorm:"size:200;not null" json:"image"`
	Role           string `gorm:"size:50" json:"role"`
	Purpose        string `gorm:"size:100" json:"purpose"`
	StudentFacing  bool   `gorm:"default:false" json:"student_facing"`
	SortOrder      int    `gorm:"default:0" json:"sort_order"`

	Ports   []ExperimentServicePort   `gorm:"foreignKey:ServiceID" json:"ports,omitempty"`
	EnvVars []ExperimentServiceEnvVar `gorm:"foreignKey:ServiceID" json:"env_vars,omitempty"`
}

func (ExperimentService) TableName() string {
	return "experiment_services"
}

type ExperimentServicePort struct {
	BaseModel
	ServiceID  uint  `gorm:"index;not null" json:"service_id"`
	Port       int32 `gorm:"not null" json:"port"`
	SortOrder  int   `gorm:"default:0" json:"sort_order"`
}

func (ExperimentServicePort) TableName() string {
	return "experiment_service_ports"
}

type ExperimentServiceEnvVar struct {
	BaseModel
	ServiceID uint   `gorm:"index;not null" json:"service_id"`
	EnvKey    string `gorm:"size:100;not null" json:"env_key"`
	EnvValue  string `gorm:"type:text" json:"env_value"`
	SortOrder int    `gorm:"default:0" json:"sort_order"`
}

func (ExperimentServiceEnvVar) TableName() string {
	return "experiment_service_env_vars"
}

type ExperimentAsset struct {
	BaseModel
	ExperimentID uint   `gorm:"index;not null" json:"experiment_id"`
	AssetKey     string `gorm:"size:50;not null" json:"asset_key"`
	Name         string `gorm:"size:200" json:"name"`
	SourceType   string `gorm:"size:30;not null" json:"source_type"`
	Bucket       string `gorm:"size:100" json:"bucket"`
	ObjectPath   string `gorm:"size:500" json:"object_path"`
	MountPath    string `gorm:"size:500;not null" json:"mount_path"`
	Required     bool   `gorm:"default:false" json:"required"`
	SortOrder    int    `gorm:"default:0" json:"sort_order"`
}

func (ExperimentAsset) TableName() string {
	return "experiment_assets"
}

type ExperimentCheckpoint struct {
	BaseModel
	ExperimentID uint   `gorm:"index;not null" json:"experiment_id"`
	CheckpointKey string `gorm:"size:50;not null" json:"checkpoint_key"`
	Type         string `gorm:"size:30;not null" json:"type"`
	Target       string `gorm:"size:50" json:"target"`
	Path         string `gorm:"size:500" json:"path"`
	Command      string `gorm:"type:text" json:"command"`
	Expected     string `gorm:"type:text" json:"expected"`
	Script       string `gorm:"type:text" json:"script"`
	Score        int    `gorm:"default:0" json:"score"`
	SortOrder    int    `gorm:"default:0" json:"sort_order"`
}

func (ExperimentCheckpoint) TableName() string {
	return "experiment_checkpoints"
}

type ExperimentRuntimeInstance struct {
	BaseModel
	ExperimentEnvID uint   `gorm:"index;not null" json:"experiment_env_id"`
	InstanceKey     string `gorm:"size:50;not null" json:"instance_key"`
	Kind            string `gorm:"size:20;not null" json:"kind"`
	PodName         string `gorm:"size:100;not null" json:"pod_name"`
	Status          string `gorm:"size:20;default:pending" json:"status"`
	StudentFacing   bool   `gorm:"default:false" json:"student_facing"`

	Tools []ExperimentRuntimeTool `gorm:"foreignKey:RuntimeInstanceID" json:"tools,omitempty"`
}

func (ExperimentRuntimeInstance) TableName() string {
	return "experiment_runtime_instances"
}

type ExperimentRuntimeTool struct {
	BaseModel
	RuntimeInstanceID uint   `gorm:"index;not null" json:"runtime_instance_id"`
	ToolKey           string `gorm:"size:50;not null" json:"tool_key"`
	Port              int32  `gorm:"not null" json:"port"`
	SortOrder         int    `gorm:"default:0" json:"sort_order"`
}

func (ExperimentRuntimeTool) TableName() string {
	return "experiment_runtime_tools"
}

type SubmissionCheckResult struct {
	BaseModel
	SubmissionID   uint   `gorm:"index;not null" json:"submission_id"`
	CheckpointKey  string `gorm:"size:50;not null" json:"checkpoint_key"`
	CheckpointType string `gorm:"size:30;not null" json:"checkpoint_type"`
	Target         string `gorm:"size:50" json:"target"`
	Passed         bool   `gorm:"default:false" json:"passed"`
	Score          int    `gorm:"default:0" json:"score"`
	Details        string `gorm:"type:text" json:"details"`
	SortOrder      int    `gorm:"default:0" json:"sort_order"`
}

func (SubmissionCheckResult) TableName() string {
	return "submission_check_results"
}
