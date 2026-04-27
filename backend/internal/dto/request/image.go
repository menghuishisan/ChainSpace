package request

// ListImagesRequest 镜像列表请求
type ListImagesRequest struct {
	PaginationRequest
	Category  string `form:"category"`
	Status    string `form:"status"`
	Keyword   string `form:"keyword"`
	IsBuiltIn *bool  `form:"is_built_in"`
}

// DefaultResources 默认资源配置
type DefaultResources struct {
	CPU     float64 `json:"cpu" binding:"required"`
	Memory  string  `json:"memory" binding:"required"`
	Storage string  `json:"storage" binding:"required"`
}

// CreateImageRequest 创建镜像请求
type CreateImageRequest struct {
	Name             string           `json:"name" binding:"required,max=100"`
	Tag              string           `json:"tag" binding:"required,max=50"`
	Category         string           `json:"category" binding:"required,max=50"`
	Description      string           `json:"description" binding:"omitempty"`
	DefaultResources DefaultResources `json:"default_resources" binding:"required"`
}

// UpdateImageRequest 更新镜像请求
type UpdateImageRequest struct {
	Name             string            `json:"name" binding:"omitempty,max=100"`
	Tag              string            `json:"tag" binding:"omitempty,max=50"`
	Category         string            `json:"category" binding:"omitempty,max=50"`
	Description      string            `json:"description" binding:"omitempty"`
	DefaultResources *DefaultResources `json:"default_resources" binding:"omitempty"`
	IsActive         *bool             `json:"is_active" binding:"omitempty"`
}
