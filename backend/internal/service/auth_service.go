package service

import (
	"context"
	"fmt"

	"github.com/chainspace/backend/internal/dto/request"
	"github.com/chainspace/backend/internal/dto/response"
	"github.com/chainspace/backend/internal/model"
	"github.com/chainspace/backend/internal/pkg/jwt"
	"github.com/chainspace/backend/internal/pkg/password"
	"github.com/chainspace/backend/internal/repository"
	"github.com/chainspace/backend/pkg/errors"
	"gorm.io/gorm"
)

// AuthService 认证服务
type AuthService struct {
	userRepo   *repository.UserRepository
	schoolRepo *repository.SchoolRepository
	jwtManager *jwt.Manager
}

// NewAuthService 创建认证服务
func NewAuthService(userRepo *repository.UserRepository, schoolRepo *repository.SchoolRepository, jwtManager *jwt.Manager) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		schoolRepo: schoolRepo,
		jwtManager: jwtManager,
	}
}

// Login 登录（手机号登录）
func (s *AuthService) Login(ctx context.Context, req *request.LoginRequest, ip string) (*response.LoginResponse, error) {
	// 根据手机号查找用户
	user, err := s.userRepo.GetByPhone(ctx, req.Phone)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrInvalidCredentials
		}
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	// 验证密码
	if !password.Verify(req.Password, user.Password) {
		return nil, errors.ErrInvalidCredentials
	}

	// 检查用户状态
	if !user.IsActive() {
		return nil, errors.ErrAccountDisabled
	}

	// 检查学校状态（非平台管理员）
	if user.SchoolID != nil && !user.IsPlatformAdmin() {
		school, err := s.schoolRepo.GetByID(ctx, *user.SchoolID)
		if err != nil {
			return nil, errors.ErrDatabaseError.WithError(err)
		}
		if !school.IsActive() {
			return nil, errors.ErrAccountDisabled.WithMessage("所属学校已被禁用或过期")
		}
	}

	// 生成Token对
	tokenPair, err := s.jwtManager.GenerateTokenPair(user.ID, user.Role, user.SchoolID)
	if err != nil {
		return nil, errors.ErrInternal.WithError(err)
	}

	// 更新最后登录信息
	_ = s.userRepo.UpdateLastLogin(ctx, user.ID, ip)

	// 构建用户响应
	userResp := &response.UserResponse{}
	userResp.FromUser(user)

	return &response.LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		User:         *userResp,
	}, nil
}

// RefreshToken 刷新Token
func (s *AuthService) RefreshToken(ctx context.Context, req *request.RefreshTokenRequest) (*response.RefreshTokenResponse, error) {
	// 刷新Token对
	tokenPair, err := s.jwtManager.RefreshTokenPair(ctx, req.RefreshToken)
	if err != nil {
		if appErr, ok := errors.AsAppError(err); ok {
			return nil, appErr
		}
		return nil, errors.ErrRefreshTokenInvalid.WithError(err)
	}

	return &response.RefreshTokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	}, nil
}

// Logout 登出
func (s *AuthService) Logout(ctx context.Context, accessClaims *jwt.Claims, req *request.LogoutRequest) error {
	refreshClaims, err := s.jwtManager.ValidateRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		if appErr, ok := errors.AsAppError(err); ok {
			return appErr
		}
		return errors.ErrRefreshTokenInvalid.WithError(err)
	}

	if refreshClaims.UserID != accessClaims.UserID {
		return errors.ErrRefreshTokenInvalid.WithMessage("Refresh Token不匹配当前登录用户")
	}

	if err := s.jwtManager.BlacklistToken(ctx, refreshClaims); err != nil {
		return errors.ErrRedisError.WithError(err)
	}
	if err := s.jwtManager.BlacklistToken(ctx, accessClaims); err != nil {
		return errors.ErrRedisError.WithError(err)
	}
	return nil
}

// ChangePassword 修改密码
func (s *AuthService) ChangePassword(ctx context.Context, userID uint, req *request.ChangePasswordRequest) error {
	// 获取用户
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrUserNotFound
		}
		return errors.ErrDatabaseError.WithError(err)
	}

	// 验证旧密码
	if !password.Verify(req.OldPassword, user.Password) {
		return errors.ErrOldPasswordWrong
	}

	// 验证新密码强度
	if !password.IsStrong(req.NewPassword) {
		return errors.ErrPasswordTooWeak
	}

	// 加密新密码
	hashedPassword, err := password.Hash(req.NewPassword)
	if err != nil {
		return errors.ErrInternal.WithError(err)
	}

	// 更新密码
	if err := s.userRepo.UpdatePassword(ctx, userID, hashedPassword); err != nil {
		return errors.ErrDatabaseError.WithError(err)
	}

	// 注销该用户所有Token
	if err := s.jwtManager.RevokeAllUserTokens(ctx, userID); err != nil {
		return errors.ErrRedisError.WithError(err)
	}

	return nil
}

// ResetPassword 重置密码（管理员操作）
func (s *AuthService) ResetPassword(ctx context.Context, operatorID uint, req *request.ResetPasswordRequest) error {
	// 获取目标用户
	user, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrUserNotFound
		}
		return errors.ErrDatabaseError.WithError(err)
	}

	// 不能重置自己的密码
	if user.ID == operatorID {
		return errors.ErrCannotModifySelf.WithMessage("不能通过此方式重置自己的密码")
	}

	// 加密新密码
	hashedPassword, err := password.Hash(req.NewPassword)
	if err != nil {
		return errors.ErrInternal.WithError(err)
	}

	// 更新密码并标记需要修改密码
	if err := s.userRepo.UpdatePasswordPolicy(ctx, req.UserID, hashedPassword, true); err != nil {
		return errors.ErrDatabaseError.WithError(err)
	}

	// 注销该用户所有Token
	if err := s.jwtManager.RevokeAllUserTokens(ctx, req.UserID); err != nil {
		return errors.ErrRedisError.WithError(err)
	}

	return nil
}

// GetCurrentUser 获取当前用户信息
func (s *AuthService) GetCurrentUser(ctx context.Context, userID uint) (*response.UserResponse, error) {
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

// UpdateProfile 更新当前用户个人资料
func (s *AuthService) UpdateProfile(ctx context.Context, userID uint, req *request.UpdateProfileRequest) (*response.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrUserNotFound
		}
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	if req.RealName != "" {
		user.RealName = req.RealName
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	userResp := &response.UserResponse{}
	userResp.FromUser(user)
	return userResp, nil
}

// ValidatePassword 验证密码强度
func (s *AuthService) ValidatePassword(pwd string) *password.ValidationResult {
	result := password.Validate(pwd)
	return &result
}

// InitPlatformAdmin 初始化平台管理员
func (s *AuthService) InitPlatformAdmin(ctx context.Context, phone, realName, pwd, email string) error {
	// 检查是否已存在平台管理员
	_, total, err := s.userRepo.List(ctx, 0, model.RolePlatformAdmin, "", "", 1, 1)
	if err != nil {
		return fmt.Errorf("check existing admin: %w", err)
	}
	if total > 0 {
		return fmt.Errorf("platform admin already exists")
	}

	// 校验密码强度（管理员密码必须满足安全要求）
	if len(pwd) < 10 {
		return fmt.Errorf("admin password must be at least 10 characters")
	}
	validation := password.Validate(pwd)
	if !validation.Valid {
		return fmt.Errorf("password too weak: %v", validation.Errors)
	}

	// 加密密码
	hashedPassword, err := password.Hash(pwd)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	// 创建管理员
	admin := &model.User{
		Password:      hashedPassword,
		Phone:         phone,
		RealName:      realName,
		Email:         email,
		Role:          model.RolePlatformAdmin,
		Status:        model.StatusActive,
		MustChangePwd: true,
	}

	if err := s.userRepo.Create(ctx, admin); err != nil {
		return fmt.Errorf("create admin: %w", err)
	}

	return nil
}
