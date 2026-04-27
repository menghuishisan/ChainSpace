package repository

import (
	"context"
	"fmt"

	"github.com/chainspace/backend/internal/model"
	"gorm.io/gorm"
)

// NotificationRepository 通知仓库
type NotificationRepository struct {
	*BaseRepository
}

// NewNotificationRepository 创建通知仓库
func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create 创建通知
func (r *NotificationRepository) Create(ctx context.Context, notification *model.Notification) error {
	return r.DB(ctx).Create(notification).Error
}

// BatchCreate 批量创建通知
func (r *NotificationRepository) BatchCreate(ctx context.Context, notifications []model.Notification) error {
	return r.DB(ctx).CreateInBatches(notifications, 100).Error
}

// GetByID 根据 ID 获取通知
func (r *NotificationRepository) GetByID(ctx context.Context, id uint) (*model.Notification, error) {
	var notification model.Notification
	err := r.DB(ctx).First(&notification, id).Error
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

// List 获取通知列表
func (r *NotificationRepository) List(ctx context.Context, userID uint, notifyType string, isRead *bool, page, pageSize int) ([]model.Notification, int64, error) {
	var notifications []model.Notification
	var total int64

	query := r.DB(ctx).Model(&model.Notification{}).Where("user_id = ?", userID)

	if notifyType != "" {
		query = query.Where("type = ?", notifyType)
	}
	if isRead != nil {
		query = query.Where("is_read = ?", *isRead)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Scopes(Paginate(page, pageSize)).
		Preload("Sender").
		Order("created_at DESC").
		Find(&notifications).Error
	if err != nil {
		return nil, 0, err
	}

	return notifications, total, nil
}

// MarkAsRead 标记为已读
func (r *NotificationRepository) MarkAsRead(ctx context.Context, id, userID uint) error {
	return r.DB(ctx).Model(&model.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": gorm.Expr("NOW()"),
		}).Error
}

// MarkAllAsRead 标记全部为已读
func (r *NotificationRepository) MarkAllAsRead(ctx context.Context, userID uint) error {
	return r.DB(ctx).Model(&model.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": gorm.Expr("NOW()"),
		}).Error
}

// BatchMarkAsRead 批量标记为已读
func (r *NotificationRepository) BatchMarkAsRead(ctx context.Context, ids []uint, userID uint) error {
	return r.DB(ctx).Model(&model.Notification{}).
		Where("id IN ? AND user_id = ?", ids, userID).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": gorm.Expr("NOW()"),
		}).Error
}

// Delete 删除通知
func (r *NotificationRepository) Delete(ctx context.Context, id, userID uint) error {
	return r.DB(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&model.Notification{}).Error
}

// BatchDelete 批量删除通知
func (r *NotificationRepository) BatchDelete(ctx context.Context, ids []uint, userID uint) error {
	return r.DB(ctx).Where("id IN ? AND user_id = ?", ids, userID).Delete(&model.Notification{}).Error
}

// CountUnread 统计未读数量
func (r *NotificationRepository) CountUnread(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.DB(ctx).Model(&model.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error
	return count, err
}

// DeleteOld 删除已读旧通知
func (r *NotificationRepository) DeleteOld(ctx context.Context, days int) error {
	cutoffExpr := fmt.Sprintf("NOW() - INTERVAL '%d days'", days)

	return r.DB(ctx).
		Where("created_at < "+cutoffExpr+" AND is_read = ?", true).
		Delete(&model.Notification{}).Error
}
