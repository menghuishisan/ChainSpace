package model

// RuntimeKernelState 是实验/解题赛/对抗赛共享的统一运行时视图。
// 上层业务语义分离，但都通过该结构暴露实例、工具、资产与策略。
type RuntimeKernelState struct {
	SessionKey  string                    `json:"session_key,omitempty"`
	SessionType string                    `json:"session_type,omitempty"`
	SessionMode string                    `json:"session_mode,omitempty"`
	Instances   []RuntimeKernelInstance   `json:"instances,omitempty"`
	Tools       []RuntimeKernelTool       `json:"tools,omitempty"`
	Assets      []RuntimeKernelAssetMount `json:"assets,omitempty"`
	Policy      RuntimeKernelPolicy       `json:"policy,omitempty"`
}

type RuntimeKernelInstance struct {
	Key           string            `json:"key"`
	Kind          string            `json:"kind"`
	PodName       string            `json:"pod_name,omitempty"`
	Status        string            `json:"status,omitempty"`
	StudentFacing bool              `json:"student_facing,omitempty"`
	Ports         []int32           `json:"ports,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`
}

type RuntimeKernelTool struct {
	Key           string `json:"key"`
	Label         string `json:"label,omitempty"`
	Kind          string `json:"kind,omitempty"`
	Target        string `json:"target,omitempty"`
	InstanceKey   string `json:"instance_key,omitempty"`
	Port          int32  `json:"port,omitempty"`
	Route         string `json:"route,omitempty"`
	WSRoute       string `json:"ws_route,omitempty"`
	StudentFacing bool   `json:"student_facing,omitempty"`
}

type RuntimeKernelAssetMount struct {
	Key        string `json:"key"`
	SourceType string `json:"source_type,omitempty"`
	Bucket     string `json:"bucket,omitempty"`
	ObjectPath string `json:"object_path,omitempty"`
	Target     string `json:"target,omitempty"`
	MountPath  string `json:"mount_path,omitempty"`
	Required   bool   `json:"required,omitempty"`
}

type RuntimeKernelPolicy struct {
	TenantKey            string   `json:"tenant_key,omitempty"`
	AllowedToolKeys      []string `json:"allowed_tool_keys,omitempty"`
	AllowedInstanceKinds []string `json:"allowed_instance_kinds,omitempty"`
}
