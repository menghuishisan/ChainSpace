package response

import (
	"time"
)

// NotificationResponse 通知响应
type NotificationResponse struct {
	ID          uint                   `json:"id"`
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	Content     string                 `json:"content"`
	Link        string                 `json:"link,omitempty"`
	RelatedID   *uint                  `json:"related_id,omitempty"`
	RelatedType string                 `json:"related_type,omitempty"`
	IsRead      bool                   `json:"is_read"`
	ReadAt      *time.Time             `json:"read_at,omitempty"`
	SenderID    *uint                  `json:"sender_id,omitempty"`
	SenderName  string                 `json:"sender_name,omitempty"`
	Extra       map[string]interface{} `json:"extra,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

// UnreadCountResponse 未读数量响应
type UnreadCountResponse struct {
	Count int64 `json:"count"`
}
