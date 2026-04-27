package model

type ExperimentBlueprint struct {
	Mode          string                       `json:"mode,omitempty"`
	Workspace     ExperimentWorkspaceBlueprint `json:"workspace"`
	Topology      ExperimentTopologyBlueprint  `json:"topology,omitempty"`
	Tools         []ExperimentToolBlueprint    `json:"tools,omitempty"`
	Collaboration ExperimentCollabBlueprint    `json:"collaboration,omitempty"`
	Nodes         []ExperimentNodeBlueprint    `json:"nodes,omitempty"`
	Services      []ExperimentServiceBlueprint `json:"services,omitempty"`
	Content       ExperimentContentBlueprint   `json:"content,omitempty"`
	Grading       ExperimentGradingBlueprint   `json:"grading,omitempty"`
}

type ExperimentWorkspaceBlueprint struct {
	Image            string                      `json:"image"`
	DisplayName      string                      `json:"display_name,omitempty"`
	Resources        ExperimentResourceBlueprint `json:"resources"`
	InteractionTools []string                    `json:"interaction_tools,omitempty"`
	InitScripts      []string                    `json:"init_scripts,omitempty"`
}

type ExperimentTopologyBlueprint struct {
	Template       string   `json:"template,omitempty"`
	SharedNetwork  bool     `json:"shared_network,omitempty"`
	ExposedEntries []string `json:"exposed_entries,omitempty"`
}

type ExperimentToolBlueprint struct {
	Key           string `json:"key"`
	Label         string `json:"label,omitempty"`
	Kind          string `json:"kind,omitempty"`
	Target        string `json:"target,omitempty"`
	StudentFacing bool   `json:"student_facing,omitempty"`
}

type ExperimentCollabBlueprint struct {
	MaxMembers int                              `json:"max_members,omitempty"`
	Roles      []ExperimentRoleBindingBlueprint `json:"roles,omitempty"`
}

type ExperimentRoleBindingBlueprint struct {
	Key      string   `json:"key"`
	Label    string   `json:"label,omitempty"`
	NodeKeys []string `json:"node_keys,omitempty"`
	ToolKeys []string `json:"tool_keys,omitempty"`
}

type ExperimentNodeBlueprint struct {
	Key              string                      `json:"key"`
	Name             string                      `json:"name,omitempty"`
	Image            string                      `json:"image"`
	Role             string                      `json:"role,omitempty"`
	Ports            []int32                     `json:"ports,omitempty"`
	Resources        ExperimentResourceBlueprint `json:"resources,omitempty"`
	StudentFacing    bool                        `json:"student_facing,omitempty"`
	InteractionTools []string                    `json:"interaction_tools,omitempty"`
	InitScripts      []string                    `json:"init_scripts,omitempty"`
}

type ExperimentResourceBlueprint struct {
	CPU     string `json:"cpu"`
	Memory  string `json:"memory"`
	Storage string `json:"storage"`
}

type ExperimentServiceBlueprint struct {
	Key           string            `json:"key"`
	Name          string            `json:"name,omitempty"`
	Image         string            `json:"image"`
	Role          string            `json:"role,omitempty"`
	Purpose       string            `json:"purpose,omitempty"`
	Ports         []int32           `json:"ports,omitempty"`
	StudentFacing bool              `json:"student_facing,omitempty"`
	EnvVars       map[string]string `json:"env_vars,omitempty"`
}

type ExperimentContentBlueprint struct {
	Assets      []ExperimentContentBlueprintAsset `json:"assets,omitempty"`
	InitScripts []string                          `json:"init_scripts,omitempty"`
}

type ExperimentGradingBlueprint struct {
	Strategy    string                          `json:"strategy,omitempty"`
	Checkpoints []ExperimentCheckpointBlueprint `json:"checkpoints,omitempty"`
}

type ExperimentContentBlueprintAsset struct {
	Key        string `json:"key"`
	Name       string `json:"name,omitempty"`
	SourceType string `json:"source_type,omitempty"`
	Bucket     string `json:"bucket,omitempty"`
	ObjectPath string `json:"object_path,omitempty"`
	Target     string `json:"target,omitempty"`
	MountPath  string `json:"mount_path,omitempty"`
	Required   bool   `json:"required,omitempty"`
}

type ExperimentCheckpointBlueprint struct {
	Key      string `json:"key"`
	Type     string `json:"type"`
	Target   string `json:"target,omitempty"`
	Path     string `json:"path,omitempty"`
	Command  string `json:"command,omitempty"`
	Expected string `json:"expected,omitempty"`
	Script   string `json:"script,omitempty"`
	Score    int    `json:"score,omitempty"`
}

type ExperimentRuntimeState struct {
	SessionMode        string                    `json:"session_mode"`
	PrimaryInstanceKey string                    `json:"primary_instance_key"`
	Instances          []ExperimentRuntimeTarget `json:"instances"`
	ToolTargets        map[string]RuntimeToolRef `json:"tool_targets,omitempty"`
}

type ExperimentRuntimeTarget struct {
	Key              string            `json:"key"`
	Kind             string            `json:"kind"`
	PodName          string            `json:"pod_name"`
	Status           string            `json:"status,omitempty"`
	Ports            []int32           `json:"ports,omitempty"`
	StudentFacing    bool              `json:"student_facing,omitempty"`
	InteractionTools []string          `json:"interaction_tools,omitempty"`
	EnvVars          map[string]string `json:"env_vars,omitempty"`
}

type RuntimeToolRef struct {
	InstanceKey string `json:"instance_key"`
	Port        int32  `json:"port"`
}
