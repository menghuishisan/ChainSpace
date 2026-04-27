package response

// SystemMonitor 系统监控数据
type SystemMonitor struct {
	CPUUsage    float64   `json:"cpu_usage"`
	MemoryUsage float64   `json:"memory_usage"`
	MemoryTotal string    `json:"memory_total"`
	MemoryUsed  string    `json:"memory_used"`
	DiskUsage   float64   `json:"disk_usage"`
	DiskTotal   string    `json:"disk_total"`
	DiskUsed    string    `json:"disk_used"`
	Uptime      int64     `json:"uptime"`
	LoadAverage []float64 `json:"load_average"`
}

// ContainerInfo 容器信息
type ContainerInfo struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Status      string  `json:"status"`
	CPUPercent  float64 `json:"cpu_percent"`
	MemoryUsage string  `json:"memory_usage"`
	CreatedAt   string  `json:"created_at"`
}

// ContainerStats 容器统计
type ContainerStats struct {
	Total      int             `json:"total"`
	Running    int             `json:"running"`
	Paused     int             `json:"paused"`
	Stopped    int             `json:"stopped"`
	Containers []ContainerInfo `json:"containers"`
}

// ServiceHealth 服务健康状态
type ServiceHealth struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	Latency   int64  `json:"latency,omitempty"`
	LastCheck string `json:"last_check"`
	Message   string `json:"message,omitempty"`
}
