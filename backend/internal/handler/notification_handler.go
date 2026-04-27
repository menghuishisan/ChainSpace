package handler

import (
	"strconv"

	"github.com/chainspace/backend/internal/dto/request"
	"github.com/chainspace/backend/internal/middleware"
	"github.com/chainspace/backend/internal/pkg/response"
	"github.com/chainspace/backend/internal/service"
	"github.com/chainspace/backend/pkg/errors"
	"github.com/gin-gonic/gin"
)

// NotificationHandler 通知处理器
type NotificationHandler struct {
	notifyService *service.NotificationService
}

// NewNotificationHandler 创建通知处理器
func NewNotificationHandler(notifyService *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notifyService: notifyService}
}

// ListNotifications 获取通知列表
func (h *NotificationHandler) ListNotifications(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req request.ListNotificationsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	list, total, err := h.notifyService.ListNotifications(c.Request.Context(), userID, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SuccessPage(c, list, total, req.GetPage(), req.GetPageSize())
}

// MarkAsRead 标记已读
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req request.MarkNotificationReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	if err := h.notifyService.MarkAsRead(c.Request.Context(), userID, req.NotificationIDs); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// MarkAllAsRead 标记全部已读
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	if err := h.notifyService.MarkAllAsRead(c.Request.Context(), userID); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// GetUnreadCount 获取未读数量
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	count, err := h.notifyService.GetUnreadCount(c.Request.Context(), userID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, gin.H{"count": count})
}

// SendNotification 发送通知（管理员）
func (h *NotificationHandler) SendNotification(c *gin.Context) {
	senderID, _ := middleware.GetUserID(c)

	var req request.SendNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	if err := h.notifyService.SendNotification(c.Request.Context(), &senderID, &req); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// BroadcastNotification 广播通知（管理员）
func (h *NotificationHandler) BroadcastNotification(c *gin.Context) {
	senderID, _ := middleware.GetUserID(c)

	var req request.BroadcastNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	if err := h.notifyService.BroadcastNotification(c.Request.Context(), &senderID, &req); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// DeleteNotification 删除通知
func (h *NotificationHandler) DeleteNotification(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	notificationID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	if err := h.notifyService.DeleteNotification(c.Request.Context(), userID, uint(notificationID)); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// BatchDeleteNotifications 批量删除通知
func (h *NotificationHandler) BatchDeleteNotifications(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req request.BatchDeleteNotificationsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	if err := h.notifyService.BatchDeleteNotifications(c.Request.Context(), userID, req.IDs); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}
