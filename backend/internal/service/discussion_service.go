package service

import (
	"context"

	"github.com/chainspace/backend/internal/dto/request"
	"github.com/chainspace/backend/internal/dto/response"
	"github.com/chainspace/backend/internal/model"
	"github.com/chainspace/backend/internal/repository"
	"github.com/chainspace/backend/pkg/errors"
	"gorm.io/gorm"
)

// DiscussionService 讨论服务
type DiscussionService struct {
	postRepo      *repository.PostRepository
	replyRepo     *repository.ReplyRepository
	postLikeRepo  *repository.PostLikeRepository
	replyLikeRepo *repository.ReplyLikeRepository
}

// NewDiscussionService 创建讨论服务
func NewDiscussionService(
	postRepo *repository.PostRepository,
	replyRepo *repository.ReplyRepository,
	postLikeRepo *repository.PostLikeRepository,
	replyLikeRepo *repository.ReplyLikeRepository,
) *DiscussionService {
	return &DiscussionService{
		postRepo:      postRepo,
		replyRepo:     replyRepo,
		postLikeRepo:  postLikeRepo,
		replyLikeRepo: replyLikeRepo,
	}
}

// CreatePost 创建帖子
func (s *DiscussionService) CreatePost(ctx context.Context, authorID, schoolID uint, req *request.CreatePostRequest) (*response.PostResponse, error) {
	post := &model.Post{
		SchoolID:     schoolID,
		AuthorID:     authorID,
		CourseID:     req.CourseID,
		ExperimentID: req.ExperimentID,
		ContestID:    req.ContestID,
		Title:        req.Title,
		Content:      req.Content,
		Tags:         req.Tags,
		Status:       model.StatusActive,
	}

	if err := s.postRepo.Create(ctx, post); err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	post, _ = s.postRepo.GetByID(ctx, post.ID)
	return s.buildPostResponse(post), nil
}

// UpdatePost 更新帖子
func (s *DiscussionService) UpdatePost(ctx context.Context, postID, userID uint, req *request.UpdatePostRequest) (*response.PostResponse, error) {
	post, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrPostNotFound
		}
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	if post.AuthorID != userID {
		return nil, errors.ErrNoPermission
	}

	if req.Title != "" {
		post.Title = req.Title
	}
	if req.Content != "" {
		post.Content = req.Content
	}
	if req.Tags != nil {
		post.Tags = req.Tags
	}

	if err := s.postRepo.Update(ctx, post); err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	return s.buildPostResponse(post), nil
}

// DeletePost 删除帖子
func (s *DiscussionService) DeletePost(ctx context.Context, postID, userID uint, isAdmin bool) error {
	post, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return errors.ErrPostNotFound
	}

	if post.AuthorID != userID && !isAdmin {
		return errors.ErrNoPermission
	}

	return s.postRepo.Delete(ctx, postID)
}

// GetPost 获取帖子详情
func (s *DiscussionService) GetPost(ctx context.Context, postID uint, userID *uint) (*response.PostDetailResponse, error) {
	post, err := s.postRepo.GetWithReplies(ctx, postID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrPostNotFound
		}
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	// 增加浏览数
	s.postRepo.IncrementViewCount(ctx, postID)

	resp := &response.PostDetailResponse{
		PostResponse: *s.buildPostResponse(post),
	}

	// 检查是否已点赞
	if userID != nil {
		liked, _ := s.postLikeRepo.Exists(ctx, postID, *userID)
		resp.IsLiked = liked
	}

	// 构建回复列表
	resp.Replies = make([]response.ReplyResponse, len(post.Replies))
	for i, reply := range post.Replies {
		resp.Replies[i] = *s.buildReplyResponse(&reply)
		if userID != nil {
			liked, _ := s.replyLikeRepo.Exists(ctx, reply.ID, *userID)
			resp.Replies[i].IsLiked = liked
		}
	}

	return resp, nil
}

// ListPosts 获取帖子列表
func (s *DiscussionService) ListPosts(ctx context.Context, schoolID uint, req *request.ListPostsRequest) ([]response.PostResponse, int64, error) {
	posts, total, err := s.postRepo.List(ctx, schoolID, req.CourseID, req.ExperimentID, req.ContestID, req.AuthorID, req.Tag, req.Status, req.Keyword, req.SortBy, req.GetPage(), req.GetPageSize())
	if err != nil {
		return nil, 0, errors.ErrDatabaseError.WithError(err)
	}

	list := make([]response.PostResponse, len(posts))
	for i, post := range posts {
		list[i] = *s.buildPostResponse(&post)
	}

	return list, total, nil
}

// ListReplies 获取帖子回复列表
func (s *DiscussionService) ListReplies(ctx context.Context, postID uint, page, pageSize int) ([]response.ReplyResponse, int64, error) {
	replies, total, err := s.replyRepo.ListByPost(ctx, postID, page, pageSize)
	if err != nil {
		return nil, 0, errors.ErrDatabaseError.WithError(err)
	}

	list := make([]response.ReplyResponse, len(replies))
	for i, r := range replies {
		list[i] = *s.buildReplyResponse(&r)
	}

	return list, total, nil
}

// CreateReply 创建回复
func (s *DiscussionService) CreateReply(ctx context.Context, authorID uint, req *request.CreateReplyRequest) (*response.ReplyResponse, error) {
	// 检查帖子是否存在且未锁定
	post, err := s.postRepo.GetByID(ctx, req.PostID)
	if err != nil {
		return nil, errors.ErrPostNotFound
	}
	if post.IsLocked {
		return nil, errors.ErrPostLocked
	}

	reply := &model.Reply{
		PostID:   req.PostID,
		AuthorID: authorID,
		ParentID: req.ParentID,
		Content:  req.Content,
		Status:   model.StatusActive,
	}

	if err := s.replyRepo.Create(ctx, reply); err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	// 更新帖子回复数
	s.postRepo.IncrementReplyCount(ctx, req.PostID)

	reply, _ = s.replyRepo.GetByID(ctx, reply.ID)
	return s.buildReplyResponse(reply), nil
}

// DeleteReply 删除回复
func (s *DiscussionService) DeleteReply(ctx context.Context, replyID, userID uint, isAdmin bool) error {
	reply, err := s.replyRepo.GetByID(ctx, replyID)
	if err != nil {
		return errors.ErrReplyNotFound
	}

	if reply.AuthorID != userID && !isAdmin {
		return errors.ErrNoPermission
	}

	return s.replyRepo.Delete(ctx, replyID)
}

// LikePost 点赞帖子
func (s *DiscussionService) LikePost(ctx context.Context, postID, userID uint) error {
	exists, _ := s.postLikeRepo.Exists(ctx, postID, userID)
	if exists {
		return errors.ErrAlreadyLiked
	}

	like := &model.PostLike{
		PostID: postID,
		UserID: userID,
	}

	if err := s.postLikeRepo.Create(ctx, like); err != nil {
		return errors.ErrDatabaseError.WithError(err)
	}

	return s.postLikeRepo.UpdatePostLikeCount(ctx, postID, 1)
}

// UnlikePost 取消点赞帖子
func (s *DiscussionService) UnlikePost(ctx context.Context, postID, userID uint) error {
	exists, _ := s.postLikeRepo.Exists(ctx, postID, userID)
	if !exists {
		return nil
	}

	if err := s.postLikeRepo.Delete(ctx, postID, userID); err != nil {
		return errors.ErrDatabaseError.WithError(err)
	}

	return s.postLikeRepo.UpdatePostLikeCount(ctx, postID, -1)
}

// LikeReply 点赞回复
func (s *DiscussionService) LikeReply(ctx context.Context, replyID, userID uint) error {
	exists, _ := s.replyLikeRepo.Exists(ctx, replyID, userID)
	if exists {
		return errors.ErrAlreadyLiked
	}

	like := &model.ReplyLike{
		ReplyID: replyID,
		UserID:  userID,
	}

	if err := s.replyLikeRepo.Create(ctx, like); err != nil {
		return errors.ErrDatabaseError.WithError(err)
	}

	return s.replyLikeRepo.UpdateReplyLikeCount(ctx, replyID, 1)
}

// UnlikeReply 取消点赞回复
func (s *DiscussionService) UnlikeReply(ctx context.Context, replyID, userID uint) error {
	exists, _ := s.replyLikeRepo.Exists(ctx, replyID, userID)
	if !exists {
		return nil
	}

	if err := s.replyLikeRepo.Delete(ctx, replyID, userID); err != nil {
		return errors.ErrDatabaseError.WithError(err)
	}

	return s.replyLikeRepo.UpdateReplyLikeCount(ctx, replyID, -1)
}

// PinPost 置顶/取消置顶帖子
func (s *DiscussionService) PinPost(ctx context.Context, postID uint, isPinned bool) error {
	return s.postRepo.UpdatePinned(ctx, postID, isPinned)
}

// LockPost 锁定/解锁帖子
func (s *DiscussionService) LockPost(ctx context.Context, postID uint, isLocked bool) error {
	return s.postRepo.UpdateLocked(ctx, postID, isLocked)
}

// AcceptReply 采纳回复
func (s *DiscussionService) AcceptReply(ctx context.Context, postID, replyID, userID uint) error {
	post, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return errors.ErrPostNotFound
	}

	if post.AuthorID != userID {
		return errors.ErrNoPermission
	}

	// 清除其他采纳
	s.replyRepo.ClearAccepted(ctx, postID)

	return s.replyRepo.UpdateAccepted(ctx, replyID, true)
}

func (s *DiscussionService) buildPostResponse(p *model.Post) *response.PostResponse {
	resp := &response.PostResponse{
		ID:         p.ID,
		Title:      p.Title,
		Content:    p.Content,
		Tags:       p.Tags,
		ViewCount:  p.ViewCount,
		ReplyCount: p.ReplyCount,
		LikeCount:  p.LikeCount,
		IsPinned:   p.IsPinned,
		IsLocked:   p.IsLocked,
		Status:     p.Status,
		CreatedAt:  p.CreatedAt,
	}
	if p.Author != nil {
		resp.AuthorID = p.AuthorID
		resp.AuthorName = p.Author.RealName
		if resp.AuthorName == "" {
			resp.AuthorName = p.Author.DisplayName()
		}
		resp.AuthorAvatar = p.Author.Avatar
	}
	return resp
}

func (s *DiscussionService) buildReplyResponse(r *model.Reply) *response.ReplyResponse {
	resp := &response.ReplyResponse{
		ID:         r.ID,
		PostID:     r.PostID,
		ParentID:   r.ParentID,
		Content:    r.Content,
		LikeCount:  r.LikeCount,
		IsAccepted: r.IsAccepted,
		CreatedAt:  r.CreatedAt,
	}
	if r.Author != nil {
		resp.AuthorID = r.AuthorID
		resp.AuthorName = r.Author.RealName
		if resp.AuthorName == "" {
			resp.AuthorName = r.Author.DisplayName()
		}
		resp.AuthorAvatar = r.Author.Avatar
	}
	return resp
}
