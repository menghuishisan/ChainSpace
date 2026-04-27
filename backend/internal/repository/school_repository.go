package repository

import (
	"context"

	"github.com/chainspace/backend/internal/model"
	"gorm.io/gorm"
)

// SchoolRepository 学校仓库
type SchoolRepository struct {
	*BaseRepository
}

// NewSchoolRepository 创建学校仓库
func NewSchoolRepository(db *gorm.DB) *SchoolRepository {
	return &SchoolRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create 创建学校
func (r *SchoolRepository) Create(ctx context.Context, school *model.School) error {
	return r.DB(ctx).Create(school).Error
}

// Update 更新学校
func (r *SchoolRepository) Update(ctx context.Context, school *model.School) error {
	return r.DB(ctx).Save(school).Error
}

// Delete 删除学校（软删除）
func (r *SchoolRepository) Delete(ctx context.Context, id uint) error {
	return r.DB(ctx).Delete(&model.School{}, id).Error
}

// GetByID 根据ID获取学校
func (r *SchoolRepository) GetByID(ctx context.Context, id uint) (*model.School, error) {
	var school model.School
	err := r.DB(ctx).First(&school, id).Error
	if err != nil {
		return nil, err
	}
	return &school, nil
}

// GetByCode 根据代码获取学校
func (r *SchoolRepository) GetByCode(ctx context.Context, code string) (*model.School, error) {
	var school model.School
	err := r.DB(ctx).Where("code = ?", code).First(&school).Error
	if err != nil {
		return nil, err
	}
	return &school, nil
}

// ExistsByCode 检查学校代码是否存在
func (r *SchoolRepository) ExistsByCode(ctx context.Context, code string) (bool, error) {
	var count int64
	err := r.DB(ctx).Model(&model.School{}).Where("code = ?", code).Count(&count).Error
	return count > 0, err
}

// ExistsByName 检查学校名称是否存在
func (r *SchoolRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	var count int64
	err := r.DB(ctx).Model(&model.School{}).Where("name = ?", name).Count(&count).Error
	return count > 0, err
}

// List 获取学校列表
func (r *SchoolRepository) List(ctx context.Context, status, keyword string, page, pageSize int) ([]model.School, int64, error) {
	var schools []model.School
	var total int64

	query := r.DB(ctx).Model(&model.School{})

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if keyword != "" {
		query = query.Where("name ILIKE ? OR code ILIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Scopes(Paginate(page, pageSize)).
		Order("created_at DESC").
		Find(&schools).Error
	if err != nil {
		return nil, 0, err
	}

	return schools, total, nil
}

// ListAll 获取所有学校
func (r *SchoolRepository) ListAll(ctx context.Context) ([]model.School, error) {
	var schools []model.School
	err := r.DB(ctx).Where("status = ?", model.StatusActive).Order("name ASC").Find(&schools).Error
	return schools, err
}

// UpdateStatus 更新学校状态
func (r *SchoolRepository) UpdateStatus(ctx context.Context, id uint, status string) error {
	return r.DB(ctx).Model(&model.School{}).Where("id = ?", id).Update("status", status).Error
}

// Count 统计学校数量
func (r *SchoolRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.DB(ctx).Model(&model.School{}).Count(&count).Error
	return count, err
}

// ClassRepository 班级仓库
type ClassRepository struct {
	*BaseRepository
}

// NewClassRepository 创建班级仓库
func NewClassRepository(db *gorm.DB) *ClassRepository {
	return &ClassRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create 创建班级
func (r *ClassRepository) Create(ctx context.Context, class *model.Class) error {
	return r.DB(ctx).Create(class).Error
}

// Update 更新班级
func (r *ClassRepository) Update(ctx context.Context, class *model.Class) error {
	return r.DB(ctx).Save(class).Error
}

// Delete 删除班级（软删除）
func (r *ClassRepository) Delete(ctx context.Context, id uint) error {
	return r.DB(ctx).Delete(&model.Class{}, id).Error
}

// GetByID 根据ID获取班级
func (r *ClassRepository) GetByID(ctx context.Context, id uint) (*model.Class, error) {
	var class model.Class
	err := r.DB(ctx).Preload("School").First(&class, id).Error
	if err != nil {
		return nil, err
	}
	return &class, nil
}

// ExistsByName 检查班级名称是否存在（同一学校内）
func (r *ClassRepository) ExistsByName(ctx context.Context, schoolID uint, name string) (bool, error) {
	var count int64
	err := r.DB(ctx).Model(&model.Class{}).
		Where("school_id = ? AND name = ?", schoolID, name).
		Count(&count).Error
	return count > 0, err
}

// GetByName 根据名称获取班级
func (r *ClassRepository) GetByName(ctx context.Context, schoolID uint, name string) (*model.Class, error) {
	var class model.Class
	err := r.DB(ctx).Where("school_id = ? AND name = ?", schoolID, name).First(&class).Error
	if err != nil {
		return nil, err
	}
	return &class, nil
}

// List 获取班级列表
func (r *ClassRepository) List(ctx context.Context, schoolID uint, status, keyword string, page, pageSize int) ([]model.Class, int64, error) {
	var classes []model.Class
	var total int64

	query := r.DB(ctx).Model(&model.Class{}).Where("school_id = ?", schoolID)

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if keyword != "" {
		query = query.Where("name ILIKE ? OR grade ILIKE ? OR major ILIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Scopes(Paginate(page, pageSize)).
		Order("grade DESC, name ASC").
		Find(&classes).Error
	if err != nil {
		return nil, 0, err
	}

	return classes, total, nil
}

// ListBySchool 获取学校所有班级
func (r *ClassRepository) ListBySchool(ctx context.Context, schoolID uint) ([]model.Class, error) {
	var classes []model.Class
	err := r.DB(ctx).Where("school_id = ? AND status = ?", schoolID, model.StatusActive).
		Order("grade DESC, name ASC").
		Find(&classes).Error
	return classes, err
}

// CountStudents 统计班级学生数
func (r *ClassRepository) CountStudents(ctx context.Context, classID uint) (int64, error) {
	var count int64
	err := r.DB(ctx).Model(&model.User{}).Where("class_id = ?", classID).Count(&count).Error
	return count, err
}
