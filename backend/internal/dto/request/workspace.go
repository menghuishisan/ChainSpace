package request

// WorkspaceMkdirRequest 描述实验工作区中新建目录的请求体。
type WorkspaceMkdirRequest struct {
	Path string `json:"path" binding:"required"`
}
