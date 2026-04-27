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

// DiscussionHandler 讨论处理器
type DiscussionHandler struct {
	discussService *service.DiscussionService
}

// NewDiscussionHandler 创建讨论处理器
func NewDiscussionHandler(discussService *service.DiscussionService) *DiscussionHandler {
	return &DiscussionHandler{discussService: discussService}
}

// CreatePost 创建帖子
func (h *DiscussionHandler) CreatePost(c *gin.Context) {
	authorID, _ := middleware.GetUserID(c)
	schoolID, _ := middleware.GetSchoolID(c)

	var req request.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	resp, err := h.discussService.CreatePost(c.Request.Context(), authorID, schoolID, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

// UpdatePost 更新帖子
func (h *DiscussionHandler) UpdatePost(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	userID, _ := middleware.GetUserID(c)

	var req request.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	resp, err := h.discussService.UpdatePost(c.Request.Context(), uint(id), userID, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

// DeletePost 删除帖子
func (h *DiscussionHandler) DeletePost(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	userID, _ := middleware.GetUserID(c)
	isAdmin := middleware.IsPlatformAdmin(c) || middleware.IsSchoolAdmin(c)

	if err := h.discussService.DeletePost(c.Request.Context(), uint(id), userID, isAdmin); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// GetPost 获取帖子
func (h *DiscussionHandler) GetPost(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var userID *uint
	if uid, ok := middleware.GetUserID(c); ok {
		userID = &uid
	}

	resp, err := h.discussService.GetPost(c.Request.Context(), uint(id), userID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

// ListPosts 获取帖子列表
func (h *DiscussionHandler) ListPosts(c *gin.Context) {
	schoolID, _ := middleware.GetSchoolID(c)

	var req request.ListPostsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	list, total, err := h.discussService.ListPosts(c.Request.Context(), schoolID, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SuccessPage(c, list, total, req.GetPage(), req.GetPageSize())
}

// ListReplies 获取帖子回复列表
func (h *DiscussionHandler) ListReplies(c *gin.Context) {
	postID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var req request.PaginationRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	list, total, err := h.discussService.ListReplies(c.Request.Context(), uint(postID), req.GetPage(), req.GetPageSize())
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SuccessPage(c, list, total, req.GetPage(), req.GetPageSize())
}

// CreateReply 创建回复
func (h *DiscussionHandler) CreateReply(c *gin.Context) {
	authorID, _ := middleware.GetUserID(c)

	var req request.CreateReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	resp, err := h.discussService.CreateReply(c.Request.Context(), authorID, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

// DeleteReply 删除回复
func (h *DiscussionHandler) DeleteReply(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("reply_id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	userID, _ := middleware.GetUserID(c)
	isAdmin := middleware.IsPlatformAdmin(c) || middleware.IsSchoolAdmin(c)

	if err := h.discussService.DeleteReply(c.Request.Context(), uint(id), userID, isAdmin); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// LikePost 点赞帖子
func (h *DiscussionHandler) LikePost(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	userID, _ := middleware.GetUserID(c)

	if err := h.discussService.LikePost(c.Request.Context(), uint(id), userID); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// UnlikePost 取消点赞帖子
func (h *DiscussionHandler) UnlikePost(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	userID, _ := middleware.GetUserID(c)

	if err := h.discussService.UnlikePost(c.Request.Context(), uint(id), userID); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// LikeReply 点赞回复
func (h *DiscussionHandler) LikeReply(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("reply_id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	userID, _ := middleware.GetUserID(c)

	if err := h.discussService.LikeReply(c.Request.Context(), uint(id), userID); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// UnlikeReply 取消点赞回复
func (h *DiscussionHandler) UnlikeReply(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("reply_id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	userID, _ := middleware.GetUserID(c)

	if err := h.discussService.UnlikeReply(c.Request.Context(), uint(id), userID); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// PinPost 置顶帖子
func (h *DiscussionHandler) PinPost(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var req request.PinPostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	if err := h.discussService.PinPost(c.Request.Context(), uint(id), req.IsPinned); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// LockPost 锁定帖子
func (h *DiscussionHandler) LockPost(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var req request.LockPostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	if err := h.discussService.LockPost(c.Request.Context(), uint(id), req.IsLocked); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// AcceptReply 采纳回复
func (h *DiscussionHandler) AcceptReply(c *gin.Context) {
	postID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var req request.AcceptReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	userID, _ := middleware.GetUserID(c)

	if err := h.discussService.AcceptReply(c.Request.Context(), uint(postID), req.ReplyID, userID); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}
