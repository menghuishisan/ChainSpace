package response

type RuntimeToolResponse struct {
	Key           string `json:"key"`
	Label         string `json:"label"`
	Kind          string `json:"kind,omitempty"`
	ModuleKey     string `json:"module_key,omitempty"`
	Target        string `json:"target,omitempty"`
	InstanceKey   string `json:"instance_key,omitempty"`
	InstanceKind  string `json:"instance_kind,omitempty"`
	StudentFacing bool   `json:"student_facing"`
	Port          int32  `json:"port"`
	Route         string `json:"route"`
	WSRoute       string `json:"ws_route,omitempty"`
	InstanceRoute string `json:"instance_route,omitempty"`
}

type RuntimeServiceResponse struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
	Purpose     string `json:"purpose,omitempty"`
	AccessURL   string `json:"access_url,omitempty"`
	Protocol    string `json:"protocol,omitempty"`
	Port        int    `json:"port,omitempty"`
	ExposeAs    string `json:"expose_as,omitempty"`
}
