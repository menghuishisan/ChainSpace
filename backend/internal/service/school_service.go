package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chainspace/backend/internal/dto/request"
	"github.com/chainspace/backend/internal/dto/response"
	"github.com/chainspace/backend/internal/model"
	"github.com/chainspace/backend/internal/pkg/logger"
	"github.com/chainspace/backend/internal/pkg/password"
	"github.com/chainspace/backend/internal/repository"
	"github.com/chainspace/backend/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SchoolService 学校服务
type SchoolService struct {
	schoolRepo *repository.SchoolRepository
	userRepo   *repository.UserRepository
	classRepo  *repository.ClassRepository
}

// NewSchoolService 创建学校服务
func NewSchoolService(schoolRepo *repository.SchoolRepository, userRepo *repository.UserRepository, classRepo *repository.ClassRepository) *SchoolService {
	return &SchoolService{
		schoolRepo: schoolRepo,
		userRepo:   userRepo,
		classRepo:  classRepo,
	}
}

// CreateSchool 创建学校（含管理员）
func (s *SchoolService) CreateSchool(ctx context.Context, req *request.CreateSchoolRequest) (*response.SchoolResponse, error) {
	exists, err := s.schoolRepo.ExistsByName(ctx, req.Name)
	if err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}
	if exists {
		return nil, errors.ErrSchoolAlreadyExists
	}

	code := req.Code
	if code == "" {
		code = strings.ToUpper(strings.ReplaceAll(req.Name, " ", ""))
		if len(code) > 20 {
			code = code[:20]
		}
	}

	exists, err = s.schoolRepo.ExistsByCode(ctx, code)
	if err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}
	if exists {
		code = fmt.Sprintf("%s_%d", code, time.Now().Unix()%10000)
	}

	exists, err = s.userRepo.ExistsByPhone(ctx, req.AdminPhone)
	if err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}
	if exists {
		return nil, errors.ErrUserAlreadyExists.WithMessage("管理员手机号已存在")
	}

	hashedPassword, err := password.Hash(req.AdminPassword)
	if err != nil {
		return nil, errors.ErrInternal.WithError(err)
	}

	school := &model.School{
		Name:        req.Name,
		Code:        code,
		Logo:        req.LogoURL,
		Address:     req.Address,
		Phone:       req.ContactPhone,
		Email:       req.ContactEmail,
		Website:     req.Website,
		Description: req.Description,
		Status:      model.StatusActive,
	}

	if req.ExpireAt != "" {
		expireAt, err := time.Parse(time.RFC3339, req.ExpireAt)
		if err != nil {
			return nil, errors.ErrInvalidParams.WithMessage("过期时间格式错误，需要 RFC3339 格式")
		}
		school.ExpireAt = &expireAt
	}

	err = s.schoolRepo.Transaction(ctx, func(tx *gorm.DB) error {
		if err := tx.Create(school).Error; err != nil {
			return err
		}

		admin := &model.User{
			Password:      hashedPassword,
			RealName:      req.AdminName,
			Phone:         req.AdminPhone,
			Email:         req.ContactEmail,
			Role:          model.RoleSchoolAdmin,
			SchoolID:      &school.ID,
			Status:        model.StatusActive,
			MustChangePwd: true,
		}
		return tx.Create(admin).Error
	})
	if err != nil {
		logger.Error("Failed to create school", zap.Error(err))
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	resp := &response.SchoolResponse{}
	resp.FromSchool(school)
	return resp, nil
}

// UpdateSchool 更新学校
func (s *SchoolService) UpdateSchool(ctx context.Context, schoolID uint, req *request.UpdateSchoolRequest) (*response.SchoolResponse, error) {
	school, err := s.schoolRepo.GetByID(ctx, schoolID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrSchoolNotFound
		}
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	if req.Name != "" && req.Name != school.Name {
		exists, err := s.schoolRepo.ExistsByName(ctx, req.Name)
		if err != nil {
			return nil, errors.ErrDatabaseError.WithError(err)
		}
		if exists {
			return nil, errors.ErrSchoolAlreadyExists
		}
		school.Name = req.Name
	}

	if req.LogoURL != "" {
		school.Logo = req.LogoURL
	}
	if req.Address != "" {
		school.Address = req.Address
	}
	if req.ContactPhone != "" {
		school.Phone = req.ContactPhone
	}
	if req.ContactEmail != "" {
		school.Email = req.ContactEmail
	}
	if req.Website != "" {
		school.Website = req.Website
	}
	if req.Description != "" {
		school.Description = req.Description
	}
	if req.Status != "" {
		school.Status = req.Status
	}
	if req.ExpireAt != "" {
		expireAt, err := time.Parse(time.RFC3339, req.ExpireAt)
		if err != nil {
			return nil, errors.ErrInvalidParams.WithMessage("过期时间格式错误，需要 RFC3339 格式")
		}
		school.ExpireAt = &expireAt
	}

	if err := s.schoolRepo.Update(ctx, school); err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	resp := &response.SchoolResponse{}
	resp.FromSchool(school)
	return resp, nil
}

// DeleteSchool 删除学校
func (s *SchoolService) DeleteSchool(ctx context.Context, schoolID uint) error {
	school, err := s.schoolRepo.GetByID(ctx, schoolID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrSchoolNotFound
		}
		return errors.ErrDatabaseError.WithError(err)
	}

	userCount, err := s.userRepo.CountBySchool(ctx, school.ID)
	if err != nil {
		return errors.ErrDatabaseError.WithError(err)
	}
	if userCount > 0 {
		return errors.ErrConflict.WithMessage("学校下还有用户，无法删除")
	}

	if err := s.schoolRepo.Delete(ctx, schoolID); err != nil {
		return errors.ErrDatabaseError.WithError(err)
	}

	return nil
}

// GetSchool 获取学校
func (s *SchoolService) GetSchool(ctx context.Context, schoolID uint) (*response.SchoolResponse, error) {
	school, err := s.schoolRepo.GetByID(ctx, schoolID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrSchoolNotFound
		}
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	resp := &response.SchoolResponse{}
	resp.FromSchool(school)
	resp.TeacherCount, _ = s.userRepo.CountByRole(ctx, schoolID, model.RoleTeacher)
	resp.StudentCount, _ = s.userRepo.CountByRole(ctx, schoolID, model.RoleStudent)
	return resp, nil
}

// ListSchools 获取学校列表
func (s *SchoolService) ListSchools(ctx context.Context, req *request.ListSchoolsRequest) ([]response.SchoolResponse, int64, error) {
	schools, total, err := s.schoolRepo.List(ctx, req.Status, req.Keyword, req.GetPage(), req.GetPageSize())
	if err != nil {
		return nil, 0, errors.ErrDatabaseError.WithError(err)
	}

	list := make([]response.SchoolResponse, len(schools))
	for i, school := range schools {
		resp := &response.SchoolResponse{}
		resp.FromSchool(&school)
		resp.TeacherCount, _ = s.userRepo.CountByRole(ctx, school.ID, model.RoleTeacher)
		resp.StudentCount, _ = s.userRepo.CountByRole(ctx, school.ID, model.RoleStudent)
		list[i] = *resp
	}

	return list, total, nil
}

// UpdateSchoolStatus 更新学校状态
func (s *SchoolService) UpdateSchoolStatus(ctx context.Context, schoolID uint, status string) error {
	_, err := s.schoolRepo.GetByID(ctx, schoolID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrSchoolNotFound
		}
		return errors.ErrDatabaseError.WithError(err)
	}

	if err := s.schoolRepo.UpdateStatus(ctx, schoolID, status); err != nil {
		return errors.ErrDatabaseError.WithError(err)
	}

	return nil
}

// CreateClass 创建班级
func (s *SchoolService) CreateClass(ctx context.Context, schoolID uint, req *request.CreateClassRequest) (*response.ClassResponse, error) {
	_, err := s.schoolRepo.GetByID(ctx, schoolID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrSchoolNotFound
		}
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	exists, err := s.classRepo.ExistsByName(ctx, schoolID, req.Name)
	if err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}
	if exists {
		return nil, errors.ErrClassAlreadyExists
	}

	class := &model.Class{
		SchoolID:    schoolID,
		Name:        req.Name,
		Grade:       req.Grade,
		Major:       req.Major,
		Description: req.Description,
		Status:      model.StatusActive,
	}

	if err := s.classRepo.Create(ctx, class); err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	resp := &response.ClassResponse{}
	resp.FromClass(class)
	return resp, nil
}

// UpdateClass 更新班级
func (s *SchoolService) UpdateClass(ctx context.Context, classID uint, req *request.UpdateClassRequest) (*response.ClassResponse, error) {
	class, err := s.classRepo.GetByID(ctx, classID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrClassNotFound
		}
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	if req.Name != "" && req.Name != class.Name {
		exists, err := s.classRepo.ExistsByName(ctx, class.SchoolID, req.Name)
		if err != nil {
			return nil, errors.ErrDatabaseError.WithError(err)
		}
		if exists {
			return nil, errors.ErrClassAlreadyExists
		}
		class.Name = req.Name
	}

	if req.Grade != "" {
		class.Grade = req.Grade
	}
	if req.Major != "" {
		class.Major = req.Major
	}
	if req.Description != "" {
		class.Description = req.Description
	}
	if req.Status != "" {
		class.Status = req.Status
	}

	if err := s.classRepo.Update(ctx, class); err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	resp := &response.ClassResponse{}
	resp.FromClass(class)
	return resp, nil
}

// DeleteClass 删除班级
func (s *SchoolService) DeleteClass(ctx context.Context, classID uint) error {
	class, err := s.classRepo.GetByID(ctx, classID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrClassNotFound
		}
		return errors.ErrDatabaseError.WithError(err)
	}

	studentCount, err := s.classRepo.CountStudents(ctx, class.ID)
	if err != nil {
		return errors.ErrDatabaseError.WithError(err)
	}
	if studentCount > 0 {
		return errors.ErrConflict.WithMessage("班级下还有学生，无法删除")
	}

	if err := s.classRepo.Delete(ctx, classID); err != nil {
		return errors.ErrDatabaseError.WithError(err)
	}

	return nil
}

// GetClass 获取班级
func (s *SchoolService) GetClass(ctx context.Context, classID uint) (*response.ClassResponse, error) {
	class, err := s.classRepo.GetByID(ctx, classID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrClassNotFound
		}
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	resp := &response.ClassResponse{}
	resp.FromClass(class)
	studentCount, _ := s.classRepo.CountStudents(ctx, classID)
	resp.StudentCount = studentCount
	return resp, nil
}

// ListClasses 获取班级列表
func (s *SchoolService) ListClasses(ctx context.Context, schoolID uint, req *request.ListClassesRequest) ([]response.ClassResponse, int64, error) {
	classes, total, err := s.classRepo.List(ctx, schoolID, req.Status, req.Keyword, req.GetPage(), req.GetPageSize())
	if err != nil {
		return nil, 0, errors.ErrDatabaseError.WithError(err)
	}

	list := make([]response.ClassResponse, len(classes))
	for i, class := range classes {
		resp := &response.ClassResponse{}
		resp.FromClass(&class)
		studentCount, _ := s.classRepo.CountStudents(ctx, class.ID)
		resp.StudentCount = studentCount
		list[i] = *resp
	}

	return list, total, nil
}

// ListClassStudents 获取班级学生列表
func (s *SchoolService) ListClassStudents(ctx context.Context, classID uint, page, pageSize int) ([]response.UserResponse, int64, error) {
	users, total, err := s.userRepo.ListByClass(ctx, classID, page, pageSize)
	if err != nil {
		return nil, 0, errors.ErrDatabaseError.WithError(err)
	}

	list := make([]response.UserResponse, len(users))
	for i, u := range users {
		resp := &response.UserResponse{}
		resp.FromUser(&u)
		list[i] = *resp
	}

	return list, total, nil
}
