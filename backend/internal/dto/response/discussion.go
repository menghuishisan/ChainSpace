package response

import (
	"time"
)

// PostResponse 帖子响应
type PostResponse struct {
	ID           uint      `json:"id"`
	AuthorID     uint      `json:"author_id"`
	AuthorName   string    `json:"author_name"`
	AuthorAvatar string    `json:"author_avatar"`
	CourseID     *uint     `json:"course_id,omitempty"`
	ExperimentID *uint     `json:"experiment_id,omitempty"`
	ContestID    *uint     `json:"contest_id,omitempty"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	Tags         []string  `json:"tags,omitempty"`
	ViewCount    int       `json:"view_count"`
	ReplyCount   int       `json:"reply_count"`
	LikeCount    int       `json:"like_count"`
	IsPinned     bool      `json:"is_pinned"`
	IsLocked     bool      `json:"is_locked"`
	Status       string    `json:"status"`
	IsLiked      bool      `json:"is_liked,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// PostDetailResponse 帖子详情响应
type PostDetailResponse struct {
	PostResponse
	Replies []ReplyResponse `json:"replies"`
}

// ReplyResponse 回复响应
type ReplyResponse struct {
	ID           uint      `json:"id"`
	PostID       uint      `json:"post_id"`
	AuthorID     uint      `json:"author_id"`
	AuthorName   string    `json:"author_name"`
	AuthorAvatar string    `json:"author_avatar"`
	ParentID     *uint     `json:"parent_id,omitempty"`
	Content      string    `json:"content"`
	LikeCount    int       `json:"like_count"`
	IsAccepted   bool      `json:"is_accepted"`
	IsLiked      bool      `json:"is_liked,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}
