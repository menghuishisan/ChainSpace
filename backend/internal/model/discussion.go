package model

import (
	"time"
)

// Post 帖子/讨论
type Post struct {
	BaseModel
	SchoolID     uint       `gorm:"index;not null" json:"school_id"`
	AuthorID     uint       `gorm:"index;not null" json:"author_id"`
	CourseID     *uint      `gorm:"index" json:"course_id"`
	ExperimentID *uint      `gorm:"index" json:"experiment_id"`
	ContestID    *uint      `gorm:"index" json:"contest_id"`
	Title        string     `gorm:"size:200;not null" json:"title"`
	Content      string     `gorm:"type:text;not null" json:"content"`
	Tags         StringList `gorm:"type:jsonb" json:"tags"`
	ViewCount    int        `gorm:"default:0" json:"view_count"`
	ReplyCount   int        `gorm:"default:0" json:"reply_count"`
	LikeCount    int        `gorm:"default:0" json:"like_count"`
	IsPinned     bool       `gorm:"default:false" json:"is_pinned"`
	IsLocked     bool       `gorm:"default:false" json:"is_locked"`
	Status       string     `gorm:"size:20;default:active" json:"status"`
	LastReplyAt  *time.Time `json:"last_reply_at"`
	LastReplyBy  *uint      `json:"last_reply_by"`

	// 关联
	School     *School     `gorm:"foreignKey:SchoolID" json:"school,omitempty"`
	Author     *User       `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
	Course     *Course     `gorm:"foreignKey:CourseID" json:"course,omitempty"`
	Experiment *Experiment `gorm:"foreignKey:ExperimentID" json:"experiment,omitempty"`
	Contest    *Contest    `gorm:"foreignKey:ContestID" json:"contest,omitempty"`
	Replies    []Reply     `gorm:"foreignKey:PostID" json:"replies,omitempty"`
}

// TableName 表名
func (Post) TableName() string {
	return "posts"
}

// Reply 回复
type Reply struct {
	BaseModel
	PostID     uint   `gorm:"index;not null" json:"post_id"`
	AuthorID   uint   `gorm:"index;not null" json:"author_id"`
	ParentID   *uint  `gorm:"index" json:"parent_id"`
	Content    string `gorm:"type:text;not null" json:"content"`
	LikeCount  int    `gorm:"default:0" json:"like_count"`
	IsAccepted bool   `gorm:"default:false" json:"is_accepted"`
	Status     string `gorm:"size:20;default:active" json:"status"`

	// 关联
	Post     *Post   `gorm:"foreignKey:PostID" json:"post,omitempty"`
	Author   *User   `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
	Parent   *Reply  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children []Reply `gorm:"foreignKey:ParentID" json:"children,omitempty"`
}

// TableName 表名
func (Reply) TableName() string {
	return "replies"
}

// PostLike 帖子点赞
type PostLike struct {
	BaseModel
	PostID uint `gorm:"uniqueIndex:idx_post_user_like;not null" json:"post_id"`
	UserID uint `gorm:"uniqueIndex:idx_post_user_like;not null" json:"user_id"`

	// 关联
	Post *Post `gorm:"foreignKey:PostID" json:"post,omitempty"`
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 表名
func (PostLike) TableName() string {
	return "post_likes"
}

// ReplyLike 回复点赞
type ReplyLike struct {
	BaseModel
	ReplyID uint `gorm:"uniqueIndex:idx_reply_user_like;not null" json:"reply_id"`
	UserID  uint `gorm:"uniqueIndex:idx_reply_user_like;not null" json:"user_id"`

	// 关联
	Reply *Reply `gorm:"foreignKey:ReplyID" json:"reply,omitempty"`
	User  *User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 表名
func (ReplyLike) TableName() string {
	return "reply_likes"
}
