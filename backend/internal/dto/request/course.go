package request

// CreateCourseRequest 创建课程请求
type CreateCourseRequest struct {
	Title       string   `json:"title" binding:"required,max=200"`
	Description string   `json:"description" binding:"omitempty"`
	Cover       string   `json:"cover" binding:"omitempty,max=500"`
	Category    string   `json:"category" binding:"omitempty,max=50"`
	Tags        []string `json:"tags" binding:"omitempty"`
	IsPublic    bool     `json:"is_public"`
	StartDate   string   `json:"start_date" binding:"omitempty"`
	EndDate     string   `json:"end_date" binding:"omitempty"`
	MaxStudents int      `json:"max_students" binding:"omitempty,min=0"`
}

// UpdateCourseRequest 更新课程请求
type UpdateCourseRequest struct {
	Title       string   `json:"title" binding:"omitempty,max=200"`
	Description string   `json:"description" binding:"omitempty"`
	Cover       string   `json:"cover" binding:"omitempty,max=500"`
	Category    string   `json:"category" binding:"omitempty,max=50"`
	Tags        []string `json:"tags" binding:"omitempty"`
	IsPublic    *bool    `json:"is_public" binding:"omitempty"`
	StartDate   string   `json:"start_date" binding:"omitempty"`
	EndDate     string   `json:"end_date" binding:"omitempty"`
	MaxStudents *int     `json:"max_students" binding:"omitempty,min=0"`
	Status      string   `json:"status" binding:"omitempty,oneof=draft published archived"`
}

// CreateChapterRequest 创建章节请求
type CreateChapterRequest struct {
	Title       string `json:"title" binding:"required,max=200"`
	Description string `json:"description" binding:"omitempty"`
	SortOrder   int    `json:"sort_order" binding:"omitempty,min=0"`
}

// UpdateChapterRequest 更新章节请求
type UpdateChapterRequest struct {
	Title       string `json:"title" binding:"omitempty,max=200"`
	Description string `json:"description" binding:"omitempty"`
	SortOrder   *int   `json:"sort_order" binding:"omitempty,min=0"`
	Status      string `json:"status" binding:"omitempty,oneof=draft published"`
}

// ReorderChaptersRequest 章节排序请求
type ReorderChaptersRequest struct {
	ChapterIDs []uint `json:"chapter_ids" binding:"required,min=1"`
}

// CreateMaterialRequest 创建资料请求
type CreateMaterialRequest struct {
	Title     string `json:"title" binding:"required,max=200"`
	Type      string `json:"type" binding:"required,oneof=video document richtext ppt link"`
	Content   string `json:"content" binding:"omitempty"`
	URL       string `json:"url" binding:"omitempty,max=500"`
	Duration  int    `json:"duration" binding:"omitempty,min=0"`
	SortOrder int    `json:"sort_order" binding:"omitempty,min=0"`
}

// UpdateMaterialRequest 更新资料请求
type UpdateMaterialRequest struct {
	Title     string `json:"title" binding:"omitempty,max=200"`
	Type      string `json:"type" binding:"omitempty,oneof=video document richtext ppt link"`
	Content   string `json:"content" binding:"omitempty"`
	URL       string `json:"url" binding:"omitempty,max=500"`
	Duration  *int   `json:"duration" binding:"omitempty,min=0"`
	SortOrder *int   `json:"sort_order" binding:"omitempty,min=0"`
	Status    string `json:"status" binding:"omitempty,oneof=active inactive"`
}

// JoinCourseRequest 加入课程请求
type JoinCourseRequest struct {
	Code string `json:"code" binding:"required,max=20"`
}

// AddStudentsToCourseRequest 添加学生到课程请求
type AddStudentsToCourseRequest struct {
	// 手机号列表（与学号列表二选一，都填则只使用手机号）
	Phones []string `json:"phones"`
	// 学号列表
	StudentNos []string `json:"student_nos"`
}

// RemoveStudentsFromCourseRequest 从课程移除学生请求
type RemoveStudentsFromCourseRequest struct {
	StudentIDs []uint `json:"student_ids" binding:"required,min=1"`
}

// UpdateMaterialProgressRequest 更新学习进度请求
type UpdateMaterialProgressRequest struct {
	Progress     float64 `json:"progress" binding:"required,min=0,max=100"`
	Duration     int     `json:"duration" binding:"omitempty,min=0"`
	LastPosition int     `json:"last_position" binding:"omitempty,min=0"`
}

// ListCoursesRequest 课程列表请求
type ListCoursesRequest struct {
	PaginationRequest
	TeacherID uint   `form:"teacher_id"`
	Category  string `form:"category"`
	Status    string `form:"status"`
	IsPublic  *bool  `form:"is_public"`
	Keyword   string `form:"keyword"`
}

// ListChaptersRequest 章节列表请求
type ListChaptersRequest struct {
	Status string `form:"status"`
}

// ListMaterialsRequest 资料列表请求
type ListMaterialsRequest struct {
	Type   string `form:"type"`
	Status string `form:"status"`
}

// ListCourseStudentsRequest 课程学生列表请求
type ListCourseStudentsRequest struct {
	PaginationRequest
	Status  string `form:"status"`
	Keyword string `form:"keyword"`
}
