package repository

import (
	"context"
	"time"

	"github.com/chainspace/backend/internal/model"
	"gorm.io/gorm"
)

// CourseStudentRepository 课程学生关联仓库
type CourseStudentRepository struct {
	*BaseRepository
}

// NewCourseStudentRepository 创建课程学生关联仓库
func NewCourseStudentRepository(db *gorm.DB) *CourseStudentRepository {
	return &CourseStudentRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create 创建关联
func (r *CourseStudentRepository) Create(ctx context.Context, cs *model.CourseStudent) error {
	cs.JoinedAt = time.Now()
	return r.DB(ctx).Create(cs).Error
}

// Delete 删除关联
func (r *CourseStudentRepository) Delete(ctx context.Context, courseID, studentID uint) error {
	return r.DB(ctx).Where("course_id = ? AND student_id = ?", courseID, studentID).
		Delete(&model.CourseStudent{}).Error
}

// BatchDelete 批量删除关联
func (r *CourseStudentRepository) BatchDelete(ctx context.Context, courseID uint, studentIDs []uint) error {
	return r.DB(ctx).Where("course_id = ? AND student_id IN ?", courseID, studentIDs).
		Delete(&model.CourseStudent{}).Error
}

// Exists 检查是否存在关联
func (r *CourseStudentRepository) Exists(ctx context.Context, courseID, studentID uint) (bool, error) {
	var count int64
	err := r.DB(ctx).Model(&model.CourseStudent{}).
		Where("course_id = ? AND student_id = ?", courseID, studentID).
		Count(&count).Error
	return count > 0, err
}

// GetByID 根据课程和学生ID获取关联
func (r *CourseStudentRepository) GetByID(ctx context.Context, courseID, studentID uint) (*model.CourseStudent, error) {
	var cs model.CourseStudent
	err := r.DB(ctx).Preload("Student").
		Where("course_id = ? AND student_id = ?", courseID, studentID).
		First(&cs).Error
	if err != nil {
		return nil, err
	}
	return &cs, nil
}

// List 获取课程学生列表
func (r *CourseStudentRepository) List(ctx context.Context, courseID uint, status, keyword string, page, pageSize int) ([]model.CourseStudent, int64, error) {
	var list []model.CourseStudent
	var total int64

	query := r.DB(ctx).Model(&model.CourseStudent{}).
		Where("course_id = ?", courseID).
		Joins("JOIN users ON users.id = course_students.student_id")

	if status != "" {
		query = query.Where("course_students.status = ?", status)
	}
	if keyword != "" {
		query = query.Where("users.real_name ILIKE ? OR users.phone ILIKE ? OR users.student_no ILIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Scopes(Paginate(page, pageSize)).
		Preload("Student").
		Preload("Student.Class").
		Order("course_students.joined_at DESC").
		Find(&list).Error
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// UpdateProgress 更新学习进度
func (r *CourseStudentRepository) UpdateProgress(ctx context.Context, courseID, studentID uint, progress float64) error {
	return r.DB(ctx).Model(&model.CourseStudent{}).
		Where("course_id = ? AND student_id = ?", courseID, studentID).
		Updates(map[string]interface{}{
			"progress":    progress,
			"last_access": time.Now(),
		}).Error
}

// UpdateStatus 更新状态
func (r *CourseStudentRepository) UpdateStatus(ctx context.Context, courseID, studentID uint, status string) error {
	return r.DB(ctx).Model(&model.CourseStudent{}).
		Where("course_id = ? AND student_id = ?", courseID, studentID).
		Update("status", status).Error
}
