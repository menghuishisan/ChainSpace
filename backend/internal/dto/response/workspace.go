package response

// WorkspaceFileItem 描述实验工作区中的单个文件或目录。
type WorkspaceFileItem struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Size       int64  `json:"size"`
	ModifiedAt string `json:"modified_at"`
	Path       string `json:"path"`
}

// WorkspaceLogEntry 描述实验环境中聚合后的一条日志记录。
type WorkspaceLogEntry struct {
	ID        string `json:"id"`
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Source    string `json:"source"`
	Message   string `json:"message"`
}
