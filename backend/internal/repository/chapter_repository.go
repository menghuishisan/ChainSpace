package repository

import (
	"context"

	"github.com/chainspace/backend/internal/model"
	"gorm.io/gorm"
)

// ChapterRepository 章节仓库
type ChapterRepository struct {
	*BaseRepository
}

// NewChapterRepository 创建章节仓库
func NewChapterRepository(db *gorm.DB) *ChapterRepository {
	return &ChapterRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create 创建章节
func (r *ChapterRepository) Create(ctx context.Context, chapter *model.Chapter) error {
	return r.DB(ctx).Create(chapter).Error
}

// Update 更新章节
func (r *ChapterRepository) Update(ctx context.Context, chapter *model.Chapter) error {
	return r.DB(ctx).Save(chapter).Error
}

// Delete 删除章节（软删除）
func (r *ChapterRepository) Delete(ctx context.Context, id uint) error {
	return r.DB(ctx).Delete(&model.Chapter{}, id).Error
}

// GetByID 根据ID获取章节
func (r *ChapterRepository) GetByID(ctx context.Context, id uint) (*model.Chapter, error) {
	var chapter model.Chapter
	err := r.DB(ctx).Preload("Course").First(&chapter, id).Error
	if err != nil {
		return nil, err
	}
	return &chapter, nil
}

// GetWithMaterials 获取章节及其资料
func (r *ChapterRepository) GetWithMaterials(ctx context.Context, id uint) (*model.Chapter, error) {
	var chapter model.Chapter
	err := r.DB(ctx).
		Preload("Materials", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		First(&chapter, id).Error
	if err != nil {
		return nil, err
	}
	return &chapter, nil
}

// ListByCourse 获取课程的章节列表
func (r *ChapterRepository) ListByCourse(ctx context.Context, courseID uint, status string) ([]model.Chapter, error) {
	var chapters []model.Chapter

	query := r.DB(ctx).Where("course_id = ?", courseID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Order("sort_order ASC").Find(&chapters).Error
	return chapters, err
}

// GetMaxSortOrder 获取最大排序号
func (r *ChapterRepository) GetMaxSortOrder(ctx context.Context, courseID uint) (int, error) {
	var maxOrder int
	err := r.DB(ctx).Model(&model.Chapter{}).
		Where("course_id = ?", courseID).
		Select("COALESCE(MAX(sort_order), 0)").
		Scan(&maxOrder).Error
	return maxOrder, err
}

// UpdateSortOrder 更新排序号
func (r *ChapterRepository) UpdateSortOrder(ctx context.Context, id uint, sortOrder int) error {
	return r.DB(ctx).Model(&model.Chapter{}).Where("id = ?", id).Update("sort_order", sortOrder).Error
}

// BatchUpdateSortOrder 批量更新排序号
func (r *ChapterRepository) BatchUpdateSortOrder(ctx context.Context, orders map[uint]int) error {
	return r.DB(ctx).Transaction(func(tx *gorm.DB) error {
		for id, order := range orders {
			if err := tx.Model(&model.Chapter{}).Where("id = ?", id).Update("sort_order", order).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// MaterialRepository 资料仓库
type MaterialRepository struct {
	*BaseRepository
}

// NewMaterialRepository 创建资料仓库
func NewMaterialRepository(db *gorm.DB) *MaterialRepository {
	return &MaterialRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create 创建资料
func (r *MaterialRepository) Create(ctx context.Context, material *model.Material) error {
	return r.DB(ctx).Create(material).Error
}

// Update 更新资料
func (r *MaterialRepository) Update(ctx context.Context, material *model.Material) error {
	return r.DB(ctx).Save(material).Error
}

// Delete 删除资料（软删除）
func (r *MaterialRepository) Delete(ctx context.Context, id uint) error {
	return r.DB(ctx).Delete(&model.Material{}, id).Error
}

// GetByID 根据ID获取资料
func (r *MaterialRepository) GetByID(ctx context.Context, id uint) (*model.Material, error) {
	var material model.Material
	err := r.DB(ctx).Preload("Chapter").First(&material, id).Error
	if err != nil {
		return nil, err
	}
	return &material, nil
}

// ListByChapter 获取章节的资料列表
func (r *MaterialRepository) ListByChapter(ctx context.Context, chapterID uint, materialType, status string) ([]model.Material, error) {
	var materials []model.Material

	query := r.DB(ctx).Where("chapter_id = ?", chapterID)
	if materialType != "" {
		query = query.Where("type = ?", materialType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Order("sort_order ASC").Find(&materials).Error
	return materials, err
}

// GetMaxSortOrder 获取最大排序号
func (r *MaterialRepository) GetMaxSortOrder(ctx context.Context, chapterID uint) (int, error) {
	var maxOrder int
	err := r.DB(ctx).Model(&model.Material{}).
		Where("chapter_id = ?", chapterID).
		Select("COALESCE(MAX(sort_order), 0)").
		Scan(&maxOrder).Error
	return maxOrder, err
}

// CountByChapter 统计章节资料数
func (r *MaterialRepository) CountByChapter(ctx context.Context, chapterID uint) (int64, error) {
	var count int64
	err := r.DB(ctx).Model(&model.Material{}).Where("chapter_id = ?", chapterID).Count(&count).Error
	return count, err
}

// CountByCourse 统计课程资料数
func (r *MaterialRepository) CountByCourse(ctx context.Context, courseID uint) (int64, error) {
	var count int64
	err := r.DB(ctx).Model(&model.Material{}).
		Joins("JOIN chapters ON chapters.id = materials.chapter_id").
		Where("chapters.course_id = ?", courseID).
		Count(&count).Error
	return count, err
}

// MaterialProgressRepository 学习进度仓库
type MaterialProgressRepository struct {
	*BaseRepository
}

// NewMaterialProgressRepository 创建学习进度仓库
func NewMaterialProgressRepository(db *gorm.DB) *MaterialProgressRepository {
	return &MaterialProgressRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Upsert 创建或更新进度
func (r *MaterialProgressRepository) Upsert(ctx context.Context, progress *model.MaterialProgress) error {
	return r.DB(ctx).Save(progress).Error
}

// GetByID 根据资料和学生ID获取进度
func (r *MaterialProgressRepository) GetByID(ctx context.Context, materialID, studentID uint) (*model.MaterialProgress, error) {
	var progress model.MaterialProgress
	err := r.DB(ctx).Where("material_id = ? AND student_id = ?", materialID, studentID).First(&progress).Error
	if err != nil {
		return nil, err
	}
	return &progress, nil
}

// ListByStudent 获取学生的学习进度列表
func (r *MaterialProgressRepository) ListByStudent(ctx context.Context, studentID uint, chapterID uint) ([]model.MaterialProgress, error) {
	var list []model.MaterialProgress

	query := r.DB(ctx).Where("student_id = ?", studentID)

	if chapterID > 0 {
		subQuery := r.DB(ctx).Model(&model.Material{}).Select("id").Where("chapter_id = ?", chapterID)
		query = query.Where("material_id IN (?)", subQuery)
	}

	err := query.Find(&list).Error
	return list, err
}

// GetCourseProgress 计算课程学习进度
func (r *MaterialProgressRepository) GetCourseProgress(ctx context.Context, courseID, studentID uint) (float64, error) {
	// 获取课程下所有资料ID
	var materialIDs []uint
	err := r.DB(ctx).Model(&model.Material{}).
		Joins("JOIN chapters ON chapters.id = materials.chapter_id").
		Where("chapters.course_id = ?", courseID).
		Pluck("materials.id", &materialIDs).Error
	if err != nil {
		return 0, err
	}

	if len(materialIDs) == 0 {
		return 0, nil
	}

	// 计算已完成的资料数
	var completedCount int64
	err = r.DB(ctx).Model(&model.MaterialProgress{}).
		Where("student_id = ? AND material_id IN ? AND completed = ?", studentID, materialIDs, true).
		Count(&completedCount).Error
	if err != nil {
		return 0, err
	}

	return float64(completedCount) / float64(len(materialIDs)) * 100, nil
}

// CountCompletedByCourse 统计课程已完成资料数
func (r *MaterialProgressRepository) CountCompletedByCourse(ctx context.Context, courseID, studentID uint) (int64, error) {
	var count int64
	err := r.DB(ctx).Model(&model.MaterialProgress{}).
		Joins("JOIN materials ON materials.id = material_progresses.material_id").
		Joins("JOIN chapters ON chapters.id = materials.chapter_id").
		Where("chapters.course_id = ? AND material_progresses.student_id = ? AND material_progresses.completed = ?", courseID, studentID, true).
		Count(&count).Error
	return count, err
}
