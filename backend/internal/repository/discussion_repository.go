package repository

import (
	"context"

	"github.com/chainspace/backend/internal/model"
	"gorm.io/gorm"
)

// PostRepository 帖子仓库
type PostRepository struct {
	*BaseRepository
}

// NewPostRepository 创建帖子仓库
func NewPostRepository(db *gorm.DB) *PostRepository {
	return &PostRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create 创建帖子
func (r *PostRepository) Create(ctx context.Context, post *model.Post) error {
	return r.DB(ctx).Create(post).Error
}

// Update 更新帖子
func (r *PostRepository) Update(ctx context.Context, post *model.Post) error {
	return r.DB(ctx).Save(post).Error
}

// Delete 删除帖子（软删除）
func (r *PostRepository) Delete(ctx context.Context, id uint) error {
	return r.DB(ctx).Delete(&model.Post{}, id).Error
}

// GetByID 根据ID获取帖子
func (r *PostRepository) GetByID(ctx context.Context, id uint) (*model.Post, error) {
	var post model.Post
	err := r.DB(ctx).Preload("Author").Preload("Course").Preload("Experiment").First(&post, id).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

// GetWithReplies 获取帖子及其回复
func (r *PostRepository) GetWithReplies(ctx context.Context, id uint) (*model.Post, error) {
	var post model.Post
	err := r.DB(ctx).
		Preload("Author").
		Preload("Replies", func(db *gorm.DB) *gorm.DB {
			return db.Where("status = ?", model.StatusActive).Order("created_at ASC")
		}).
		Preload("Replies.Author").
		First(&post, id).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

// List 获取帖子列表
func (r *PostRepository) List(ctx context.Context, schoolID, courseID, experimentID, contestID, authorID uint, tag, status, keyword, sortBy string, page, pageSize int) ([]model.Post, int64, error) {
	var posts []model.Post
	var total int64

	query := r.DB(ctx).Model(&model.Post{})

	if schoolID > 0 {
		query = query.Where("school_id = ?", schoolID)
	}
	if courseID > 0 {
		query = query.Where("course_id = ?", courseID)
	}
	if experimentID > 0 {
		query = query.Where("experiment_id = ?", experimentID)
	}
	if contestID > 0 {
		query = query.Where("contest_id = ?", contestID)
	}
	if authorID > 0 {
		query = query.Where("author_id = ?", authorID)
	}
	if tag != "" {
		query = query.Where("? = ANY(tags)", tag)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if keyword != "" {
		query = query.Where("title ILIKE ? OR content ILIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 排序
	orderClause := "is_pinned DESC, "
	switch sortBy {
	case "latest":
		orderClause += "created_at DESC"
	case "hot":
		orderClause += "reply_count DESC, view_count DESC"
	case "reply":
		orderClause += "last_reply_at DESC"
	default:
		orderClause += "created_at DESC"
	}

	err = query.Scopes(Paginate(page, pageSize)).
		Preload("Author").
		Order(orderClause).
		Find(&posts).Error
	if err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

// IncrementViewCount 增加浏览数
func (r *PostRepository) IncrementViewCount(ctx context.Context, id uint) error {
	return r.DB(ctx).Model(&model.Post{}).Where("id = ?", id).
		Update("view_count", gorm.Expr("view_count + 1")).Error
}

// IncrementReplyCount 增加回复数
func (r *PostRepository) IncrementReplyCount(ctx context.Context, id uint) error {
	return r.DB(ctx).Model(&model.Post{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"reply_count":   gorm.Expr("reply_count + 1"),
			"last_reply_at": gorm.Expr("NOW()"),
		}).Error
}

// UpdatePinned 更新置顶状态
func (r *PostRepository) UpdatePinned(ctx context.Context, id uint, isPinned bool) error {
	return r.DB(ctx).Model(&model.Post{}).Where("id = ?", id).Update("is_pinned", isPinned).Error
}

// UpdateLocked 更新锁定状态
func (r *PostRepository) UpdateLocked(ctx context.Context, id uint, isLocked bool) error {
	return r.DB(ctx).Model(&model.Post{}).Where("id = ?", id).Update("is_locked", isLocked).Error
}

// ReplyRepository 回复仓库
type ReplyRepository struct {
	*BaseRepository
}

// NewReplyRepository 创建回复仓库
func NewReplyRepository(db *gorm.DB) *ReplyRepository {
	return &ReplyRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create 创建回复
func (r *ReplyRepository) Create(ctx context.Context, reply *model.Reply) error {
	return r.DB(ctx).Create(reply).Error
}

// Update 更新回复
func (r *ReplyRepository) Update(ctx context.Context, reply *model.Reply) error {
	return r.DB(ctx).Save(reply).Error
}

// Delete 删除回复（软删除）
func (r *ReplyRepository) Delete(ctx context.Context, id uint) error {
	return r.DB(ctx).Delete(&model.Reply{}, id).Error
}

// GetByID 根据ID获取回复
func (r *ReplyRepository) GetByID(ctx context.Context, id uint) (*model.Reply, error) {
	var reply model.Reply
	err := r.DB(ctx).Preload("Author").Preload("Post").First(&reply, id).Error
	if err != nil {
		return nil, err
	}
	return &reply, nil
}

// ListByPost 获取帖子的回复列表
func (r *ReplyRepository) ListByPost(ctx context.Context, postID uint, page, pageSize int) ([]model.Reply, int64, error) {
	var replies []model.Reply
	var total int64

	query := r.DB(ctx).Model(&model.Reply{}).
		Where("post_id = ? AND status = ?", postID, model.StatusActive)

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Scopes(Paginate(page, pageSize)).
		Preload("Author").
		Order("created_at ASC").
		Find(&replies).Error
	if err != nil {
		return nil, 0, err
	}

	return replies, total, nil
}

// UpdateAccepted 更新采纳状态
func (r *ReplyRepository) UpdateAccepted(ctx context.Context, id uint, isAccepted bool) error {
	return r.DB(ctx).Model(&model.Reply{}).Where("id = ?", id).Update("is_accepted", isAccepted).Error
}

// ClearAccepted 清除帖子其他回复的采纳状态
func (r *ReplyRepository) ClearAccepted(ctx context.Context, postID uint) error {
	return r.DB(ctx).Model(&model.Reply{}).Where("post_id = ?", postID).Update("is_accepted", false).Error
}

// PostLikeRepository 帖子点赞仓库
type PostLikeRepository struct {
	*BaseRepository
}

// NewPostLikeRepository 创建帖子点赞仓库
func NewPostLikeRepository(db *gorm.DB) *PostLikeRepository {
	return &PostLikeRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create 创建点赞
func (r *PostLikeRepository) Create(ctx context.Context, like *model.PostLike) error {
	return r.DB(ctx).Create(like).Error
}

// Delete 删除点赞
func (r *PostLikeRepository) Delete(ctx context.Context, postID, userID uint) error {
	return r.DB(ctx).Where("post_id = ? AND user_id = ?", postID, userID).Delete(&model.PostLike{}).Error
}

// Exists 检查是否已点赞
func (r *PostLikeRepository) Exists(ctx context.Context, postID, userID uint) (bool, error) {
	var count int64
	err := r.DB(ctx).Model(&model.PostLike{}).
		Where("post_id = ? AND user_id = ?", postID, userID).
		Count(&count).Error
	return count > 0, err
}

// CountByPost 统计帖子点赞数
func (r *PostLikeRepository) CountByPost(ctx context.Context, postID uint) (int64, error) {
	var count int64
	err := r.DB(ctx).Model(&model.PostLike{}).Where("post_id = ?", postID).Count(&count).Error
	return count, err
}

// UpdatePostLikeCount 更新帖子点赞数
func (r *PostLikeRepository) UpdatePostLikeCount(ctx context.Context, postID uint, delta int) error {
	return r.DB(ctx).Model(&model.Post{}).Where("id = ?", postID).
		Update("like_count", gorm.Expr("like_count + ?", delta)).Error
}

// ReplyLikeRepository 回复点赞仓库
type ReplyLikeRepository struct {
	*BaseRepository
}

// NewReplyLikeRepository 创建回复点赞仓库
func NewReplyLikeRepository(db *gorm.DB) *ReplyLikeRepository {
	return &ReplyLikeRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create 创建点赞
func (r *ReplyLikeRepository) Create(ctx context.Context, like *model.ReplyLike) error {
	return r.DB(ctx).Create(like).Error
}

// Delete 删除点赞
func (r *ReplyLikeRepository) Delete(ctx context.Context, replyID, userID uint) error {
	return r.DB(ctx).Where("reply_id = ? AND user_id = ?", replyID, userID).Delete(&model.ReplyLike{}).Error
}

// Exists 检查是否已点赞
func (r *ReplyLikeRepository) Exists(ctx context.Context, replyID, userID uint) (bool, error) {
	var count int64
	err := r.DB(ctx).Model(&model.ReplyLike{}).
		Where("reply_id = ? AND user_id = ?", replyID, userID).
		Count(&count).Error
	return count > 0, err
}

// UpdateReplyLikeCount 更新回复点赞数
func (r *ReplyLikeRepository) UpdateReplyLikeCount(ctx context.Context, replyID uint, delta int) error {
	return r.DB(ctx).Model(&model.Reply{}).Where("id = ?", replyID).
		Update("like_count", gorm.Expr("like_count + ?", delta)).Error
}
