package service

import (
	"context"
	"fmt"

	"github.com/chainspace/backend/internal/dto/request"
	"github.com/chainspace/backend/internal/dto/response"
	"github.com/chainspace/backend/internal/model"
	"github.com/chainspace/backend/internal/pkg/password"
	"github.com/chainspace/backend/internal/repository"
	"github.com/chainspace/backend/pkg/errors"
	"gorm.io/gorm"
)

// UserService 用户服务
type UserService struct {
	userRepo   *repository.UserRepository
	schoolRepo *repository.SchoolRepository
	classRepo  *repository.ClassRepository
}

// NewUserService 创建用户服务
func NewUserService(userRepo *repository.UserRepository, schoolRepo *repository.SchoolRepository, classRepo *repository.ClassRepository) *UserService {
	return &UserService{
		userRepo:   userRepo,
		schoolRepo: schoolRepo,
		classRepo:  classRepo,
	}
}

// CreateUser 创建用户
func (s *UserService) CreateUser(ctx context.Context, req *request.CreateUserRequest, operatorSchoolID *uint) (*response.UserResponse, error) {
	exists, err := s.userRepo.ExistsByPhone(ctx, req.Phone)
	if err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}
	if exists {
		return nil, errors.ErrUserAlreadyExists.WithMessage("手机号已被使用")
	}

	if req.Email != "" {
		exists, err = s.userRepo.ExistsByEmail(ctx, req.Email)
		if err != nil {
			return nil, errors.ErrDatabaseError.WithError(err)
		}
		if exists {
			return nil, errors.ErrEmailExists
		}
	}

	var schoolID *uint
	if req.SchoolID > 0 {
		schoolID = &req.SchoolID
	} else if operatorSchoolID != nil {
		schoolID = operatorSchoolID
	}

	if schoolID != nil {
		_, err = s.schoolRepo.GetByID(ctx, *schoolID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, errors.ErrSchoolNotFound
			}
			return nil, errors.ErrDatabaseError.WithError(err)
		}
	}

	if req.Role == model.RoleStudent {
		if req.StudentNo == "" {
			return nil, errors.ErrInvalidParams.WithMessage("学生账号必须填写学号")
		}
		if schoolID != nil {
			exists, err = s.userRepo.ExistsByStudentNo(ctx, *schoolID, req.StudentNo)
			if err != nil {
				return nil, errors.ErrDatabaseError.WithError(err)
			}
			if exists {
				return nil, errors.ErrStudentNoExists
			}
		}
	}

	if !password.IsStrong(req.Password) {
		return nil, errors.ErrPasswordTooWeak
	}

	hashedPassword, err := password.Hash(req.Password)
	if err != nil {
		return nil, errors.ErrInternal.WithError(err)
	}

	user := &model.User{
		Password:      hashedPassword,
		Email:         req.Email,
		Phone:         req.Phone,
		RealName:      req.RealName,
		Role:          req.Role,
		SchoolID:      schoolID,
		StudentNo:     req.StudentNo,
		Status:        model.StatusActive,
		MustChangePwd: true,
	}

	if req.ClassID > 0 {
		user.ClassID = &req.ClassID
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	user, err = s.userRepo.GetByID(ctx, user.ID)
	if err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	userResp := &response.UserResponse{}
	userResp.FromUser(user)
	return userResp, nil
}

// UpdateUser 更新用户
func (s *UserService) UpdateUser(ctx context.Context, userID uint, req *request.UpdateUserRequest) (*response.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrUserNotFound
		}
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	if req.Email != "" && req.Email != user.Email {
		exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
		if err != nil {
			return nil, errors.ErrDatabaseError.WithError(err)
		}
		if exists {
			return nil, errors.ErrEmailExists
		}
		user.Email = req.Email
	}

	if req.Phone != "" && req.Phone != user.Phone {
		exists, err := s.userRepo.ExistsByPhone(ctx, req.Phone)
		if err != nil {
			return nil, errors.ErrDatabaseError.WithError(err)
		}
		if exists {
			return nil, errors.ErrUserAlreadyExists.WithMessage("手机号已被使用")
		}
		user.Phone = req.Phone
	}

	if req.StudentNo != "" && req.StudentNo != user.StudentNo && user.SchoolID != nil {
		exists, err := s.userRepo.ExistsByStudentNo(ctx, *user.SchoolID, req.StudentNo)
		if err != nil {
			return nil, errors.ErrDatabaseError.WithError(err)
		}
		if exists {
			return nil, errors.ErrStudentNoExists
		}
		user.StudentNo = req.StudentNo
	}

	if req.RealName != "" {
		user.RealName = req.RealName
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}
	if req.ClassID != nil {
		user.ClassID = req.ClassID
	}
	if req.Status != "" {
		user.Status = req.Status
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	userResp := &response.UserResponse{}
	userResp.FromUser(user)
	return userResp, nil
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(ctx context.Context, userID, operatorID uint) error {
	if userID == operatorID {
		return errors.ErrCannotModifySelf
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrUserNotFound
		}
		return errors.ErrDatabaseError.WithError(err)
	}

	if user.IsPlatformAdmin() {
		return errors.ErrNoPermission.WithMessage("不能删除平台管理员")
	}

	if err := s.userRepo.Delete(ctx, userID); err != nil {
		return errors.ErrDatabaseError.WithError(err)
	}

	return nil
}

// GetUser 获取用户
func (s *UserService) GetUser(ctx context.Context, userID uint) (*response.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrUserNotFound
		}
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	userResp := &response.UserResponse{}
	userResp.FromUser(user)
	return userResp, nil
}

// ListUsers 获取用户列表
func (s *UserService) ListUsers(ctx context.Context, req *request.ListUsersRequest, operatorSchoolID *uint) ([]response.UserListResponse, int64, error) {
	schoolID := req.SchoolID
	if operatorSchoolID != nil {
		schoolID = *operatorSchoolID
	}

	users, total, err := s.userRepo.List(ctx, schoolID, req.Role, req.Status, req.Keyword, req.GetPage(), req.GetPageSize())
	if err != nil {
		return nil, 0, errors.ErrDatabaseError.WithError(err)
	}

	list := make([]response.UserListResponse, len(users))
	for i, user := range users {
		list[i] = response.UserListResponse{
			ID:          user.ID,
			Email:       user.Email,
			Phone:       user.Phone,
			RealName:    user.RealName,
			Avatar:      user.Avatar,
			Role:        user.Role,
			StudentNo:   user.StudentNo,
			Status:      user.Status,
			LastLoginAt: user.LastLoginAt,
			CreatedAt:   user.CreatedAt,
		}
		if user.Class != nil {
			list[i].ClassName = user.Class.Name
		}
	}

	return list, total, nil
}

// BatchImportStudents 批量导入学生
func (s *UserService) BatchImportStudents(ctx context.Context, schoolID uint, req *request.BatchImportStudentRequest) (*response.BatchImportResult, error) {
	result := &response.BatchImportResult{
		Total:  len(req.Students),
		Errors: make([]response.BatchImportError, 0),
	}

	phoneSet := make(map[string]int)
	studentNoSet := make(map[string]int)
	for i, item := range req.Students {
		row := i + 1

		if item.Phone == "" {
			result.Failed++
			result.Errors = append(result.Errors, response.BatchImportError{
				Row:     row,
				Field:   "phone",
				Message: "手机号不能为空",
			})
			continue
		}
		if item.RealName == "" {
			result.Failed++
			result.Errors = append(result.Errors, response.BatchImportError{
				Row:     row,
				Field:   "real_name",
				Message: "姓名不能为空",
			})
			continue
		}
		if item.StudentNo == "" {
			result.Failed++
			result.Errors = append(result.Errors, response.BatchImportError{
				Row:     row,
				Field:   "student_no",
				Message: "学号不能为空",
			})
			continue
		}

		if prevRow, exists := phoneSet[item.Phone]; exists {
			result.Failed++
			result.Errors = append(result.Errors, response.BatchImportError{
				Row:     row,
				Field:   "phone",
				Message: fmt.Sprintf("手机号与第%d行重复", prevRow),
			})
			continue
		}
		phoneSet[item.Phone] = row

		if prevRow, exists := studentNoSet[item.StudentNo]; exists {
			result.Failed++
			result.Errors = append(result.Errors, response.BatchImportError{
				Row:     row,
				Field:   "student_no",
				Message: fmt.Sprintf("学号与第%d行重复", prevRow),
			})
			continue
		}
		studentNoSet[item.StudentNo] = row
	}

	if result.Failed > 0 {
		return result, nil
	}

	users := make([]model.User, 0, len(req.Students))

	for i, item := range req.Students {
		row := i + 1

		exists, err := s.userRepo.ExistsByPhone(ctx, item.Phone)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, response.BatchImportError{
				Row:     row,
				Field:   "phone",
				Message: "检查手机号失败",
			})
			continue
		}
		if exists {
			result.Failed++
			result.Errors = append(result.Errors, response.BatchImportError{
				Row:     row,
				Field:   "phone",
				Message: "手机号已存在",
			})
			continue
		}

		exists, err = s.userRepo.ExistsByStudentNo(ctx, schoolID, item.StudentNo)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, response.BatchImportError{
				Row:     row,
				Field:   "student_no",
				Message: "检查学号失败",
			})
			continue
		}
		if exists {
			result.Failed++
			result.Errors = append(result.Errors, response.BatchImportError{
				Row:     row,
				Field:   "student_no",
				Message: "学号已存在",
			})
			continue
		}

		pwd := item.Password
		if pwd == "" {
			pwd = item.StudentNo
		}

		hashedPassword, err := password.Hash(pwd)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, response.BatchImportError{
				Row:     row,
				Message: "密码加密失败",
			})
			continue
		}

		user := model.User{
			SchoolID:      &schoolID,
			Password:      hashedPassword,
			RealName:      item.RealName,
			StudentNo:     item.StudentNo,
			Email:         item.Email,
			Phone:         item.Phone,
			Role:          model.RoleStudent,
			Status:        model.StatusActive,
			MustChangePwd: true,
		}

		if req.ClassID > 0 {
			user.ClassID = &req.ClassID
		}

		users = append(users, user)
		result.Success++
	}

	if len(users) > 0 {
		if err := s.userRepo.Transaction(ctx, func(tx *gorm.DB) error {
			return tx.CreateInBatches(users, 100).Error
		}); err != nil {
			result.Success = 0
			result.Failed = len(req.Students)
			return nil, errors.ErrDatabaseError.WithError(fmt.Errorf("batch create users: %w", err))
		}
	}

	return result, nil
}

// UpdateUserStatus 更新用户状态
func (s *UserService) UpdateUserStatus(ctx context.Context, userID, operatorID uint, status string) error {
	if userID == operatorID {
		return errors.ErrCannotModifySelf
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrUserNotFound
		}
		return errors.ErrDatabaseError.WithError(err)
	}

	if user.IsPlatformAdmin() {
		return errors.ErrNoPermission.WithMessage("不能修改平台管理员状态")
	}

	if err := s.userRepo.UpdateStatus(ctx, userID, status); err != nil {
		return errors.ErrDatabaseError.WithError(err)
	}

	return nil
}
