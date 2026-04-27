package repository

import (
	"context"
	"crypto/rand"
	"math/big"

	"github.com/chainspace/backend/internal/model"
	"gorm.io/gorm"
)

// CourseRepository 课程仓库
type CourseRepository struct {
	*BaseRepository
}

// NewCourseRepository 创建课程仓库
func NewCourseRepository(db *gorm.DB) *CourseRepository {
	return &CourseRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create 创建课程
func (r *CourseRepository) Create(ctx context.Context, course *model.Course) error {
	return r.DB(ctx).Create(course).Error
}

// Update 更新课程
func (r *CourseRepository) Update(ctx context.Context, course *model.Course) error {
	return r.DB(ctx).Save(course).Error
}

// Delete 删除课程（软删除）
func (r *CourseRepository) Delete(ctx context.Context, id uint) error {
	return r.DB(ctx).Delete(&model.Course{}, id).Error
}

// GetByID 根据ID获取课程
func (r *CourseRepository) GetByID(ctx context.Context, id uint) (*model.Course, error) {
	var course model.Course
	err := r.DB(ctx).Preload("School").Preload("Teacher").First(&course, id).Error
	if err != nil {
		return nil, err
	}
	return &course, nil
}

// GetByCode 根据课程码获取课程
func (r *CourseRepository) GetByCode(ctx context.Context, code string) (*model.Course, error) {
	var course model.Course
	err := r.DB(ctx).Where("code = ?", code).First(&course).Error
	if err != nil {
		return nil, err
	}
	return &course, nil
}

// GetByInviteCode 根据邀请码获取课程
func (r *CourseRepository) GetByInviteCode(ctx context.Context, inviteCode string) (*model.Course, error) {
	var course model.Course
	err := r.DB(ctx).Where("invite_code = ?", inviteCode).First(&course).Error
	if err != nil {
		return nil, err
	}
	return &course, nil
}

// GetWithChapters 获取课程及其章节
func (r *CourseRepository) GetWithChapters(ctx context.Context, id uint) (*model.Course, error) {
	var course model.Course
	err := r.DB(ctx).
		Preload("Teacher").
		Preload("Chapters", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Preload("Chapters.Materials", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		First(&course, id).Error
	if err != nil {
		return nil, err
	}
	return &course, nil
}

// List 获取课程列表
func (r *CourseRepository) List(ctx context.Context, schoolID, teacherID uint, category, status, keyword string, isPublic *bool, page, pageSize int) ([]model.Course, int64, error) {
	var courses []model.Course
	var total int64

	query := r.DB(ctx).Model(&model.Course{})

	if schoolID > 0 {
		query = query.Where("school_id = ?", schoolID)
	}
	if teacherID > 0 {
		query = query.Where("teacher_id = ?", teacherID)
	}
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if isPublic != nil {
		query = query.Where("is_public = ?", *isPublic)
	}
	if keyword != "" {
		query = query.Where("title ILIKE ? OR description ILIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Scopes(Paginate(page, pageSize)).
		Preload("Teacher").
		Order("created_at DESC").
		Find(&courses).Error
	if err != nil {
		return nil, 0, err
	}

	return courses, total, nil
}

// ListByTeacher 获取教师的课程列表
func (r *CourseRepository) ListByTeacher(ctx context.Context, teacherID uint, page, pageSize int) ([]model.Course, int64, error) {
	return r.List(ctx, 0, teacherID, "", "", "", nil, page, pageSize)
}

// ListByStudent 获取学生加入的课程列表
func (r *CourseRepository) ListByStudent(ctx context.Context, studentID uint, page, pageSize int) ([]model.Course, int64, error) {
	var courses []model.Course
	var total int64

	subQuery := r.DB(ctx).Model(&model.CourseStudent{}).
		Select("course_id").
		Where("student_id = ? AND status = ?", studentID, model.StatusActive)

	query := r.DB(ctx).Model(&model.Course{}).Where("id IN (?)", subQuery)

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Scopes(Paginate(page, pageSize)).
		Preload("Teacher").
		Order("created_at DESC").
		Find(&courses).Error
	if err != nil {
		return nil, 0, err
	}

	return courses, total, nil
}

// CountStudents 统计课程学生数
func (r *CourseRepository) CountStudents(ctx context.Context, courseID uint) (int64, error) {
	var count int64
	err := r.DB(ctx).Model(&model.CourseStudent{}).
		Where("course_id = ? AND status = ?", courseID, model.StatusActive).
		Count(&count).Error
	return count, err
}

// CountChapters 统计课程章节数
func (r *CourseRepository) CountChapters(ctx context.Context, courseID uint) (int64, error) {
	var count int64
	err := r.DB(ctx).Model(&model.Chapter{}).Where("course_id = ?", courseID).Count(&count).Error
	return count, err
}

// GenerateCode 生成课程码
func (r *CourseRepository) GenerateCode(ctx context.Context) (string, error) {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	const codeLength = 6

	for {
		code := generateRandomCode(charset, codeLength)
		exists, err := r.codeExists(ctx, code)
		if err != nil {
			return "", err
		}
		if !exists {
			return code, nil
		}
	}
}

func (r *CourseRepository) codeExists(ctx context.Context, code string) (bool, error) {
	var count int64
	err := r.DB(ctx).Model(&model.Course{}).Where("code = ?", code).Count(&count).Error
	return count > 0, err
}

// inviteCodeExists 检查邀请码是否存在
func (r *CourseRepository) inviteCodeExists(ctx context.Context, code string) (bool, error) {
	var count int64
	err := r.DB(ctx).Model(&model.Course{}).Where("invite_code = ?", code).Count(&count).Error
	return count > 0, err
}

// GenerateInviteCode 生成邀请码
func (r *CourseRepository) GenerateInviteCode(ctx context.Context) (string, error) {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	const codeLength = 8

	for {
		code := generateRandomCode(charset, codeLength)
		exists, err := r.inviteCodeExists(ctx, code)
		if err != nil {
			return "", err
		}
		if !exists {
			return code, nil
		}
	}
}

// 生成随机码
func generateRandomCode(charset string, length int) string {
	code := make([]byte, length)
	charsetLen := big.NewInt(int64(len(charset)))
	for i := range code {
		idx, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			idx = big.NewInt(int64(i % len(charset)))
		}
		code[i] = charset[idx.Int64()]
	}
	return string(code)
}

// CountExperiments 统计课程实验数
func (r *CourseRepository) CountExperiments(ctx context.Context, courseID uint) (int64, error) {
	var count int64
	err := r.DB(ctx).Model(&model.Experiment{}).
		Joins("JOIN chapters ON chapters.id = experiments.chapter_id").
		Where("chapters.course_id = ? AND experiments.status = ?", courseID, model.ExperimentStatusPublished).
		Count(&count).Error
	return count, err
}

// CountCompletedExperiments 统计已完成的实验数
func (r *CourseRepository) CountCompletedExperiments(ctx context.Context, courseID, studentID uint) (int64, error) {
	var count int64
	err := r.DB(ctx).Model(&model.Submission{}).
		Joins("JOIN experiments ON experiments.id = submissions.experiment_id").
		Joins("JOIN chapters ON chapters.id = experiments.chapter_id").
		Where(
			"chapters.course_id = ? AND submissions.student_id = ? AND experiments.status = ? AND submissions.score IS NOT NULL AND submissions.score >= experiments.pass_score",
			courseID,
			studentID,
			model.ExperimentStatusPublished,
		).
		Distinct("submissions.experiment_id").
		Count(&count).Error
	return count, err
}
