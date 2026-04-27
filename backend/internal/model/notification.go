package model

import (
	"time"
)

// Notification 通知
type Notification struct {
	BaseModel
	UserID      uint       `gorm:"index;not null" json:"user_id"`
	SchoolID    *uint      `gorm:"index" json:"school_id"`
	Type        string     `gorm:"size:30;not null" json:"type"`
	Title       string     `gorm:"size:200;not null" json:"title"`
	Content     string     `gorm:"type:text" json:"content"`
	Link        string     `gorm:"size:500" json:"link"`
	RelatedID   *uint      `json:"related_id"`
	RelatedType string     `gorm:"size:50" json:"related_type"`
	IsRead      bool       `gorm:"default:false;index" json:"is_read"`
	ReadAt      *time.Time `json:"read_at"`
	SenderID    *uint      `json:"sender_id"`
	Extra       JSONMap    `gorm:"type:jsonb" json:"extra"`

	// 关联
	User   *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Sender *User `gorm:"foreignKey:SenderID" json:"sender,omitempty"`
}

// TableName 表名
func (Notification) TableName() string {
	return "notifications"
}

// MarkAsRead 标记为已读
func (n *Notification) MarkAsRead() {
	n.IsRead = true
	now := time.Now()
	n.ReadAt = &now
}

// NotificationTemplate 通知模板
type NotificationTemplate struct {
	BaseModel
	Code        string    `gorm:"size:50;uniqueIndex;not null" json:"code"`
	Type        string    `gorm:"size:30;not null" json:"type"`
	Title       string    `gorm:"size:200;not null" json:"title"`
	Content     string    `gorm:"type:text" json:"content"`
	Description string    `gorm:"type:text" json:"description"`
	Variables   JSONArray `gorm:"type:jsonb" json:"variables"`
	Status      string    `gorm:"size:20;default:active" json:"status"`
}

// TableName 表名
func (NotificationTemplate) TableName() string {
	return "notification_templates"
}

// 通知模板代码常量
const (
	NotifyTemplateNewCourse      = "new_course"
	NotifyTemplateNewExperiment  = "new_experiment"
	NotifyTemplateGraded         = "graded"
	NotifyTemplateContestStart   = "contest_start"
	NotifyTemplateContestEnd     = "contest_end"
	NotifyTemplateTeamInvite     = "team_invite"
	NotifyTemplateNewReply       = "new_reply"
	NotifyTemplateSystemAnnounce = "system_announce"
)
