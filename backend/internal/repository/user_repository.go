package repository

import (
	"context"

	"github.com/chainspace/backend/internal/model"
	"gorm.io/gorm"
)

// UserRepository 用户仓储
type UserRepository struct {
	*BaseRepository
}

// NewUserRepository 创建用户仓储
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create 创建用户
func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	return r.DB(ctx).Create(user).Error
}

// Update 更新用户
func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	return r.DB(ctx).Save(user).Error
}

// Delete 删除用户（软删除）
func (r *UserRepository) Delete(ctx context.Context, id uint) error {
	return r.DB(ctx).Delete(&model.User{}, id).Error
}

// GetByID 根据 ID 获取用户
func (r *UserRepository) GetByID(ctx context.Context, id uint) (*model.User, error) {
	var user model.User
	err := r.DB(ctx).Preload("School").Preload("Class").First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail 根据邮箱获取用户
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.DB(ctx).Preload("School").Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByPhone 根据手机号获取用户
func (r *UserRepository) GetByPhone(ctx context.Context, phone string) (*model.User, error) {
	var user model.User
	err := r.DB(ctx).Preload("School").Where("phone = ?", phone).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ExistsByEmail 检查邮箱是否存在
func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.DB(ctx).Model(&model.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

// ExistsByPhone 检查手机号是否存在
func (r *UserRepository) ExistsByPhone(ctx context.Context, phone string) (bool, error) {
	var count int64
	err := r.DB(ctx).Model(&model.User{}).Where("phone = ?", phone).Count(&count).Error
	return count > 0, err
}

// ExistsByStudentNo 检查学号是否存在
func (r *UserRepository) ExistsByStudentNo(ctx context.Context, schoolID uint, studentNo string) (bool, error) {
	var count int64
	err := r.DB(ctx).Model(&model.User{}).
		Where("school_id = ? AND student_no = ?", schoolID, studentNo).
		Count(&count).Error
	return count > 0, err
}

// GetByStudentNo 根据学号获取用户
func (r *UserRepository) GetByStudentNo(ctx context.Context, schoolID uint, studentNo string) (*model.User, error) {
	var user model.User
	query := r.DB(ctx).Where("student_no = ?", studentNo)
	if schoolID > 0 {
		query = query.Where("school_id = ?", schoolID)
	}
	err := query.First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// List 获取用户列表
func (r *UserRepository) List(ctx context.Context, schoolID uint, role, status, keyword string, page, pageSize int) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	query := r.DB(ctx).Model(&model.User{})

	if schoolID > 0 {
		query = query.Where("school_id = ?", schoolID)
	}
	if role != "" {
		query = query.Where("role = ?", role)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if keyword != "" {
		query = query.Where("real_name ILIKE ? OR phone ILIKE ? OR email ILIKE ? OR student_no ILIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Scopes(Paginate(page, pageSize)).
		Preload("School").Preload("Class").
		Order("created_at DESC").
		Find(&users).Error
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// ListBySchool 获取学校用户列表
func (r *UserRepository) ListBySchool(ctx context.Context, schoolID uint, role string, page, pageSize int) ([]model.User, int64, error) {
	return r.List(ctx, schoolID, role, "", "", page, pageSize)
}

// ListByClass 获取班级学生列表
func (r *UserRepository) ListByClass(ctx context.Context, classID uint, page, pageSize int) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	query := r.DB(ctx).Model(&model.User{}).Where("class_id = ?", classID)

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Scopes(Paginate(page, pageSize)).
		Order("student_no ASC").
		Find(&users).Error
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// BatchCreate 批量创建用户
func (r *UserRepository) BatchCreate(ctx context.Context, users []model.User) error {
	return r.DB(ctx).CreateInBatches(users, 100).Error
}

// UpdatePassword 更新密码
func (r *UserRepository) UpdatePassword(ctx context.Context, userID uint, password string) error {
	return r.DB(ctx).Model(&model.User{}).Where("id = ?", userID).
		Updates(map[string]interface{}{
			"password":        password,
			"must_change_pwd": false,
		}).Error
}

func (r *UserRepository) UpdatePasswordPolicy(ctx context.Context, userID uint, password string, mustChange bool) error {
	return r.DB(ctx).Model(&model.User{}).Where("id = ?", userID).
		Updates(map[string]interface{}{
			"password":        password,
			"must_change_pwd": mustChange,
		}).Error
}

// UpdateLastLogin 更新最后登录信息
func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID uint, ip string) error {
	return r.DB(ctx).Model(&model.User{}).Where("id = ?", userID).
		Updates(map[string]interface{}{
			"last_login_at": gorm.Expr("NOW()"),
			"last_login_ip": ip,
		}).Error
}

// UpdateStatus 更新用户状态
func (r *UserRepository) UpdateStatus(ctx context.Context, userID uint, status string) error {
	return r.DB(ctx).Model(&model.User{}).Where("id = ?", userID).
		Update("status", status).Error
}

// CountBySchool 统计学校用户数
func (r *UserRepository) CountBySchool(ctx context.Context, schoolID uint) (int64, error) {
	var count int64
	err := r.DB(ctx).Model(&model.User{}).Where("school_id = ?", schoolID).Count(&count).Error
	return count, err
}

// CountByRole 按角色统计用户数
func (r *UserRepository) CountByRole(ctx context.Context, schoolID uint, role string) (int64, error) {
	var count int64
	query := r.DB(ctx).Model(&model.User{})
	if schoolID > 0 {
		query = query.Where("school_id = ?", schoolID)
	}
	if role != "" {
		query = query.Where("role = ?", role)
	}
	err := query.Count(&count).Error
	return count, err
}
