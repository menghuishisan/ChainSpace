package request

// CreatePostRequest 创建帖子请求
type CreatePostRequest struct {
	CourseID     *uint    `json:"course_id" binding:"omitempty"`
	ExperimentID *uint    `json:"experiment_id" binding:"omitempty"`
	ContestID    *uint    `json:"contest_id" binding:"omitempty"`
	Title        string   `json:"title" binding:"required,max=200"`
	Content      string   `json:"content" binding:"required"`
	Tags         []string `json:"tags" binding:"omitempty"`
}

// UpdatePostRequest 更新帖子请求
type UpdatePostRequest struct {
	Title   string   `json:"title" binding:"omitempty,max=200"`
	Content string   `json:"content" binding:"omitempty"`
	Tags    []string `json:"tags" binding:"omitempty"`
}

// CreateReplyRequest 创建回复请求
type CreateReplyRequest struct {
	PostID   uint   `json:"post_id" binding:"required"`
	ParentID *uint  `json:"parent_id" binding:"omitempty"`
	Content  string `json:"content" binding:"required"`
}

// UpdateReplyRequest 更新回复请求
type UpdateReplyRequest struct {
	Content string `json:"content" binding:"required"`
}

// ListPostsRequest 帖子列表请求
type ListPostsRequest struct {
	PaginationRequest
	CourseID     uint   `form:"course_id"`
	ExperimentID uint   `form:"experiment_id"`
	ContestID    uint   `form:"contest_id"`
	AuthorID     uint   `form:"author_id"`
	Tag          string `form:"tag"`
	Status       string `form:"status"`
	Keyword      string `form:"keyword"`
	SortBy       string `form:"sort_by"`
}

// ListRepliesRequest 回复列表请求
type ListRepliesRequest struct {
	PaginationRequest
	PostID uint `form:"post_id" binding:"required"`
}

// PinPostRequest 置顶帖子请求
type PinPostRequest struct {
	IsPinned bool `json:"is_pinned"`
}

// LockPostRequest 锁定帖子请求
type LockPostRequest struct {
	IsLocked bool `json:"is_locked"`
}

// AcceptReplyRequest 采纳回复请求
type AcceptReplyRequest struct {
	ReplyID uint `json:"reply_id" binding:"required"`
}
