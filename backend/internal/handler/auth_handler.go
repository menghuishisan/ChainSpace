package handler

import (
	"github.com/chainspace/backend/internal/dto/request"
	"github.com/chainspace/backend/internal/middleware"
	"github.com/chainspace/backend/internal/pkg/response"
	"github.com/chainspace/backend/internal/service"
	"github.com/chainspace/backend/pkg/errors"
	"github.com/gin-gonic/gin"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Login 登录
// @Summary 用户登录
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body request.LoginRequest true "登录请求"
// @Success 200 {object} response.Response{data=response.LoginResponse}
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req request.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	resp, err := h.authService.Login(c.Request.Context(), &req, c.ClientIP())
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

// RefreshToken 刷新Token
// @Summary 刷新Token
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body request.RefreshTokenRequest true "刷新Token请求"
// @Success 200 {object} response.Response{data=response.RefreshTokenResponse}
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req request.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	resp, err := h.authService.RefreshToken(c.Request.Context(), &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

// Logout 登出
// @Summary 用户登出
// @Tags 认证
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		response.Error(c, errors.ErrLoginRequired)
		return
	}

	var req request.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	if err := h.authService.Logout(c.Request.Context(), claims, &req); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// GetCurrentUser 获取当前用户信息
// @Summary 获取当前用户信息
// @Tags 认证
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{data=response.UserResponse}
// @Router /api/v1/auth/me [get]
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Error(c, errors.ErrLoginRequired)
		return
	}

	resp, err := h.authService.GetCurrentUser(c.Request.Context(), userID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

// UpdateProfile 更新个人资料
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Error(c, errors.ErrLoginRequired)
		return
	}

	var req request.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	resp, err := h.authService.UpdateProfile(c.Request.Context(), userID, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Tags 认证
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body request.ChangePasswordRequest true "修改密码请求"
// @Success 200 {object} response.Response
// @Router /api/v1/auth/password [put]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Error(c, errors.ErrLoginRequired)
		return
	}

	var req request.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	if err := h.authService.ChangePassword(c.Request.Context(), userID, &req); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// ResetPassword 重置密码（管理员）
// @Summary 重置用户密码
// @Tags 认证
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body request.ResetPasswordRequest true "重置密码请求"
// @Success 200 {object} response.Response
// @Router /api/v1/auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Error(c, errors.ErrLoginRequired)
		return
	}

	var req request.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	if err := h.authService.ResetPassword(c.Request.Context(), userID, &req); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}
