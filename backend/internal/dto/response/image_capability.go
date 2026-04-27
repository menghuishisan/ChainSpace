package response

// ImageCapabilityResponse 描述镜像可用于编排时的能力摘要。
type ImageCapabilityResponse struct {
	ID               uint              `json:"id"`
	Name             string            `json:"name"`
	Tag              string            `json:"tag"`
	FullName         string            `json:"full_name"`
	Category         string            `json:"category"`
	Status           string            `json:"status"`
	ToolKeys         []string          `json:"tool_keys"`
	DefaultPorts     []int             `json:"default_ports"`
	DefaultResources map[string]string `json:"default_resources"`
	Features         []string          `json:"features"`
	Compatibility    []string          `json:"compatibility"`
}
