package request

// SendNotificationRequest 发送通知请求
type SendNotificationRequest struct {
	UserIDs     []uint                 `json:"user_ids" binding:"required,min=1"`
	Type        string                 `json:"type" binding:"required,oneof=system course experiment contest discuss"`
	Title       string                 `json:"title" binding:"required,max=200"`
	Content     string                 `json:"content" binding:"omitempty"`
	Link        string                 `json:"link" binding:"omitempty,max=500"`
	RelatedID   *uint                  `json:"related_id" binding:"omitempty"`
	RelatedType string                 `json:"related_type" binding:"omitempty,max=50"`
	Extra       map[string]interface{} `json:"extra" binding:"omitempty"`
}

// BroadcastNotificationRequest 广播通知请求
type BroadcastNotificationRequest struct {
	SchoolID *uint                  `json:"school_id" binding:"omitempty"`
	Role     string                 `json:"role" binding:"omitempty,oneof=school_admin teacher student"`
	Type     string                 `json:"type" binding:"required,oneof=system course experiment contest discuss"`
	Title    string                 `json:"title" binding:"required,max=200"`
	Content  string                 `json:"content" binding:"omitempty"`
	Link     string                 `json:"link" binding:"omitempty,max=500"`
	Extra    map[string]interface{} `json:"extra" binding:"omitempty"`
}

// MarkNotificationReadRequest 标记通知已读请求
type MarkNotificationReadRequest struct {
	NotificationIDs []uint `json:"notification_ids" binding:"required,min=1"`
}

// BatchDeleteNotificationsRequest 批量删除通知请求
type BatchDeleteNotificationsRequest struct {
	IDs []uint `json:"ids" binding:"required,min=1"`
}

// ListNotificationsRequest 通知列表请求
type ListNotificationsRequest struct {
	PaginationRequest
	Type   string `form:"type"`
	IsRead *bool  `form:"is_read"`
}
