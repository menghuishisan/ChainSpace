package request

import "github.com/chainspace/backend/internal/model"

// CreateExperimentRequest 创建实验请求
type CreateExperimentRequest struct {
	ChapterID     uint                      `json:"chapter_id" binding:"required"`
	Title         string                    `json:"title" binding:"required,max=200"`
	Description   string                    `json:"description" binding:"omitempty"`
	Type          string                    `json:"type" binding:"required,oneof=visualization code_dev command_op data_analysis tool_usage config_debug reverse troubleshoot collaboration"`
	Difficulty    int                       `json:"difficulty" binding:"omitempty,min=1,max=5"`
	EstimatedTime int                       `json:"estimated_time" binding:"omitempty,min=1"`
	MaxScore      int                       `json:"max_score" binding:"omitempty,min=0"`
	PassScore     int                       `json:"pass_score" binding:"omitempty,min=0"`
	AutoGrade     bool                      `json:"auto_grade"`
	Blueprint     model.ExperimentBlueprint `json:"blueprint" binding:"required"`
	SortOrder     int                       `json:"sort_order" binding:"omitempty,min=0"`
	StartTime     string                    `json:"start_time" binding:"omitempty"`
	EndTime       string                    `json:"end_time" binding:"omitempty"`
	AllowLate     bool                      `json:"allow_late"`
	LateDeduction int                       `json:"late_deduction" binding:"omitempty,min=0,max=100"`
}

// UpdateExperimentRequest 更新实验请求
type UpdateExperimentRequest struct {
	ChapterID     *uint                      `json:"chapter_id" binding:"omitempty,min=1"`
	Title         string                     `json:"title" binding:"omitempty,max=200"`
	Description   string                     `json:"description" binding:"omitempty"`
	Type          string                     `json:"type" binding:"omitempty,oneof=visualization code_dev command_op data_analysis tool_usage config_debug reverse troubleshoot collaboration"`
	Difficulty    *int                       `json:"difficulty" binding:"omitempty,min=1,max=5"`
	EstimatedTime *int                       `json:"estimated_time" binding:"omitempty,min=1"`
	MaxScore      *int                       `json:"max_score" binding:"omitempty,min=0"`
	PassScore     *int                       `json:"pass_score" binding:"omitempty,min=0"`
	AutoGrade     *bool                      `json:"auto_grade" binding:"omitempty"`
	Blueprint     *model.ExperimentBlueprint `json:"blueprint" binding:"omitempty"`
	SortOrder     *int                       `json:"sort_order" binding:"omitempty,min=0"`
	StartTime     string                     `json:"start_time" binding:"omitempty"`
	EndTime       string                     `json:"end_time" binding:"omitempty"`
	AllowLate     *bool                      `json:"allow_late" binding:"omitempty"`
	LateDeduction *int                       `json:"late_deduction" binding:"omitempty,min=0,max=100"`
	Status        string                     `json:"status" binding:"omitempty,oneof=draft published archived"`
}

// StartEnvRequest 启动实验环境请求
type StartEnvRequest struct {
	ExperimentID uint   `json:"experiment_id" binding:"required"`
	SnapshotURL  string `json:"snapshot_url" binding:"omitempty,max=500"`
}

// ExtendEnvRequest 延长实验环境时间请求
type ExtendEnvRequest struct {
	Duration int `json:"duration" binding:"required,min=1,max=240"`
}

type CreateSnapshotRequest struct{}

// SubmitExperimentRequest 提交实验请求
type SubmitExperimentRequest struct {
	ExperimentID uint   `json:"experiment_id" binding:"required"`
	EnvID        string `json:"env_id" binding:"omitempty,max=50"`
	Content      string `json:"content" binding:"omitempty"`
	FileURL      string `json:"file_url" binding:"omitempty,max=500"`
}

// GradeSubmissionRequest 批改提交请求
type GradeSubmissionRequest struct {
	Score    int    `json:"score" binding:"required,min=0"`
	Feedback string `json:"feedback" binding:"omitempty"`
}

// ListExperimentsRequest 实验列表请求
type ListExperimentsRequest struct {
	PaginationRequest
	CourseID  uint   `form:"course_id"`
	ChapterID uint   `form:"chapter_id"`
	Type      string `form:"type"`
	Status    string `form:"status"`
	Keyword   string `form:"keyword"`
}

// ListEnvsRequest 实验环境列表请求
type ListEnvsRequest struct {
	PaginationRequest
	ExperimentID uint   `form:"experiment_id"`
	UserID       uint   `form:"user_id"`
	Status       string `form:"status"`
}

type ListExperimentSessionsRequest struct {
	PaginationRequest
	ExperimentID uint   `form:"experiment_id"`
	Status       string `form:"status"`
}

type ListExperimentSessionMessagesRequest struct {
	PaginationRequest
}

type SendExperimentSessionMessageRequest struct {
	Message string `json:"message" binding:"required"`
}

type UpdateExperimentSessionMemberRequest struct {
	RoleKey         *string `json:"role_key" binding:"omitempty,max=50"`
	AssignedNodeKey *string `json:"assigned_node_key" binding:"omitempty,max=50"`
	JoinStatus      *string `json:"join_status" binding:"omitempty,oneof=joined left"`
}

// ListSubmissionsRequest 提交列表请求
type ListSubmissionsRequest struct {
	PaginationRequest
	ExperimentID uint   `form:"experiment_id"`
	StudentID    uint   `form:"student_id"`
	Status       string `form:"status"`
}

// CreateDockerImageRequest 创建镜像请求
type CreateDockerImageRequest struct {
	Name        string                 `json:"name" binding:"required,max=100"`
	Tag         string                 `json:"tag" binding:"required,max=50"`
	Registry    string                 `json:"registry" binding:"omitempty,max=200"`
	Description string                 `json:"description" binding:"omitempty"`
	Category    string                 `json:"category" binding:"omitempty,max=50"`
	Features    []interface{}          `json:"features" binding:"omitempty"`
	EnvVars     map[string]interface{} `json:"env_vars" binding:"omitempty"`
	Ports       []interface{}          `json:"ports" binding:"omitempty"`
	BaseImage   string                 `json:"base_image" binding:"omitempty,max=100"`
}

// UpdateDockerImageRequest 更新镜像请求
type UpdateDockerImageRequest struct {
	Tag         string                 `json:"tag" binding:"omitempty,max=50"`
	Description string                 `json:"description" binding:"omitempty"`
	Category    string                 `json:"category" binding:"omitempty,max=50"`
	Features    []interface{}          `json:"features" binding:"omitempty"`
	EnvVars     map[string]interface{} `json:"env_vars" binding:"omitempty"`
	Ports       []interface{}          `json:"ports" binding:"omitempty"`
	Status      string                 `json:"status" binding:"omitempty,oneof=active inactive"`
}

// ListDockerImagesRequest 镜像列表请求
type ListDockerImagesRequest struct {
	PaginationRequest
	Category  string `form:"category"`
	Status    string `form:"status"`
	IsBuiltIn *bool  `form:"is_built_in"`
	Keyword   string `form:"keyword"`
}
