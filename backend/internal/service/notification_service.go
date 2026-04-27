package service

import (
	"context"
	"strconv"

	"github.com/chainspace/backend/internal/dto/request"
	"github.com/chainspace/backend/internal/dto/response"
	"github.com/chainspace/backend/internal/model"
	"github.com/chainspace/backend/internal/pkg/cache"
	"github.com/chainspace/backend/internal/pkg/logger"
	"github.com/chainspace/backend/internal/repository"
	"github.com/chainspace/backend/pkg/errors"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// NotificationService 通知服务
type NotificationService struct {
	notifyRepo *repository.NotificationRepository
	userRepo   *repository.UserRepository
	redis      *redis.Client
}

// NewNotificationService 创建通知服务
func NewNotificationService(notifyRepo *repository.NotificationRepository, userRepo *repository.UserRepository, redisClient *redis.Client) *NotificationService {
	return &NotificationService{
		notifyRepo: notifyRepo,
		userRepo:   userRepo,
		redis:      redisClient,
	}
}

// SendNotification 发送通知
func (s *NotificationService) SendNotification(ctx context.Context, senderID *uint, req *request.SendNotificationRequest) error {
	notifications := make([]model.Notification, len(req.UserIDs))
	for i, userID := range req.UserIDs {
		notifications[i] = model.Notification{
			UserID:      userID,
			Type:        req.Type,
			Title:       req.Title,
			Content:     req.Content,
			Link:        req.Link,
			RelatedID:   req.RelatedID,
			RelatedType: req.RelatedType,
			SenderID:    senderID,
			Extra:       req.Extra,
		}
	}

	if err := s.notifyRepo.BatchCreate(ctx, notifications); err != nil {
		return err
	}

	s.invalidateUnreadCountCache(ctx, req.UserIDs...)
	return nil
}

// BroadcastNotification 广播通知
func (s *NotificationService) BroadcastNotification(ctx context.Context, senderID *uint, req *request.BroadcastNotificationRequest) error {
	// 获取目标用户列表
	users, _, err := s.userRepo.List(ctx, 0, req.Role, model.StatusActive, "", 1, 10000)
	if err != nil {
		return errors.ErrDatabaseError.WithError(err)
	}

	if len(users) == 0 {
		return nil
	}

	notifications := make([]model.Notification, len(users))
	for i, user := range users {
		notifications[i] = model.Notification{
			UserID:   user.ID,
			SchoolID: user.SchoolID,
			Type:     req.Type,
			Title:    req.Title,
			Content:  req.Content,
			Link:     req.Link,
			SenderID: senderID,
			Extra:    req.Extra,
		}
	}

	if err := s.notifyRepo.BatchCreate(ctx, notifications); err != nil {
		return err
	}

	userIDs := make([]uint, len(users))
	for i, user := range users {
		userIDs[i] = user.ID
	}
	s.invalidateUnreadCountCache(ctx, userIDs...)
	return nil
}

// ListNotifications 获取通知列表
func (s *NotificationService) ListNotifications(ctx context.Context, userID uint, req *request.ListNotificationsRequest) ([]response.NotificationResponse, int64, error) {
	notifications, total, err := s.notifyRepo.List(ctx, userID, req.Type, req.IsRead, req.GetPage(), req.GetPageSize())
	if err != nil {
		return nil, 0, errors.ErrDatabaseError.WithError(err)
	}

	list := make([]response.NotificationResponse, len(notifications))
	for i, n := range notifications {
		list[i] = response.NotificationResponse{
			ID:          n.ID,
			Type:        n.Type,
			Title:       n.Title,
			Content:     n.Content,
			Link:        n.Link,
			RelatedID:   n.RelatedID,
			RelatedType: n.RelatedType,
			IsRead:      n.IsRead,
			ReadAt:      n.ReadAt,
			CreatedAt:   n.CreatedAt,
		}
		if n.Sender != nil {
			list[i].SenderName = n.Sender.RealName
			if list[i].SenderName == "" {
				list[i].SenderName = n.Sender.DisplayName()
			}
		}
	}

	return list, total, nil
}

// MarkAsRead 标记为已读
func (s *NotificationService) MarkAsRead(ctx context.Context, userID uint, notificationIDs []uint) error {
	if err := s.notifyRepo.BatchMarkAsRead(ctx, notificationIDs, userID); err != nil {
		return err
	}
	s.invalidateUnreadCountCache(ctx, userID)
	return nil
}

// MarkAllAsRead 标记全部为已读
func (s *NotificationService) MarkAllAsRead(ctx context.Context, userID uint) error {
	if err := s.notifyRepo.MarkAllAsRead(ctx, userID); err != nil {
		return err
	}
	s.invalidateUnreadCountCache(ctx, userID)
	return nil
}

// GetUnreadCount 获取未读数量
func (s *NotificationService) GetUnreadCount(ctx context.Context, userID uint) (int64, error) {
	if count, ok := s.getUnreadCountFromCache(ctx, userID); ok {
		return count, nil
	}

	count, err := s.notifyRepo.CountUnread(ctx, userID)
	if err != nil {
		return 0, err
	}

	s.cacheUnreadCount(ctx, userID, count)
	return count, nil
}

// DeleteNotification 删除通知
func (s *NotificationService) DeleteNotification(ctx context.Context, userID, notificationID uint) error {
	if err := s.notifyRepo.Delete(ctx, notificationID, userID); err != nil {
		return err
	}
	s.invalidateUnreadCountCache(ctx, userID)
	return nil
}

// BatchDeleteNotifications 批量删除通知
func (s *NotificationService) BatchDeleteNotifications(ctx context.Context, userID uint, ids []uint) error {
	if err := s.notifyRepo.BatchDelete(ctx, ids, userID); err != nil {
		return err
	}
	s.invalidateUnreadCountCache(ctx, userID)
	return nil
}

// CleanupOldNotifications 清理旧通知
func (s *NotificationService) CleanupOldNotifications(ctx context.Context, days int) error {
	return s.notifyRepo.DeleteOld(ctx, days)
}

func (s *NotificationService) getUnreadCountFromCache(ctx context.Context, userID uint) (int64, bool) {
	if s.redis == nil {
		return 0, false
	}

	value, err := s.redis.Get(ctx, cache.NotifyCountKey(userID)).Result()
	if err != nil {
		if err != redis.Nil {
			logger.Warn("Failed to get unread notification count cache", zap.Uint("user_id", userID), zap.Error(err))
		}
		return 0, false
	}

	count, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		logger.Warn("Failed to parse unread notification count cache", zap.Uint("user_id", userID), zap.Error(err))
		return 0, false
	}

	return count, true
}

func (s *NotificationService) cacheUnreadCount(ctx context.Context, userID uint, count int64) {
	if s.redis == nil {
		return
	}

	if err := s.redis.Set(ctx, cache.NotifyCountKey(userID), count, cache.TTLShort).Err(); err != nil {
		logger.Warn("Failed to set unread notification count cache", zap.Uint("user_id", userID), zap.Error(err))
	}
}

func (s *NotificationService) invalidateUnreadCountCache(ctx context.Context, userIDs ...uint) {
	if s.redis == nil || len(userIDs) == 0 {
		return
	}

	keys := make([]string, 0, len(userIDs))
	seen := make(map[uint]struct{}, len(userIDs))
	for _, userID := range userIDs {
		if userID == 0 {
			continue
		}
		if _, exists := seen[userID]; exists {
			continue
		}
		seen[userID] = struct{}{}
		keys = append(keys, cache.NotifyCountKey(userID))
	}
	if len(keys) == 0 {
		return
	}

	if err := s.redis.Del(ctx, keys...).Err(); err != nil {
		logger.Warn("Failed to invalidate unread notification count cache", zap.Int("user_count", len(keys)), zap.Error(err))
	}
}
