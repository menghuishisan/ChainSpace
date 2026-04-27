package handler

import (
	"net/http"
	"strconv"

	"github.com/chainspace/backend/internal/dto/request"
	"github.com/chainspace/backend/internal/middleware"
	"github.com/chainspace/backend/internal/pkg/response"
	"github.com/chainspace/backend/internal/service"
	"github.com/chainspace/backend/pkg/errors"
	"github.com/gin-gonic/gin"
)

// UserHandler 用户处理器
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler 创建用户处理器
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// CreateUser 创建用户
// @Summary 创建用户
// @Tags 用户管理
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body request.CreateUserRequest true "创建用户请求"
// @Success 200 {object} response.Response{data=response.UserResponse}
// @Router /api/v1/users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req request.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	var operatorSchoolID *uint
	if schoolID, ok := middleware.GetSchoolID(c); ok {
		operatorSchoolID = &schoolID
	}

	resp, err := h.userService.CreateUser(c.Request.Context(), &req, operatorSchoolID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

// UpdateUser 更新用户
// @Summary 更新用户
// @Tags 用户管理
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Param request body request.UpdateUserRequest true "更新用户请求"
// @Success 200 {object} response.Response{data=response.UserResponse}
// @Router /api/v1/users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var req request.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	resp, err := h.userService.UpdateUser(c.Request.Context(), uint(id), &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

// DeleteUser 删除用户
// @Summary 删除用户
// @Tags 用户管理
// @Security BearerAuth
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response
// @Router /api/v1/users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	operatorID, _ := middleware.GetUserID(c)

	if err := h.userService.DeleteUser(c.Request.Context(), uint(id), operatorID); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// GetUser 获取用户
// @Summary 获取用户详情
// @Tags 用户管理
// @Security BearerAuth
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response{data=response.UserResponse}
// @Router /api/v1/users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	resp, err := h.userService.GetUser(c.Request.Context(), uint(id))
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

// ListUsers 获取用户列表
// @Summary 获取用户列表
// @Tags 用户管理
// @Security BearerAuth
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Param school_id query int false "学校ID"
// @Param role query string false "角色"
// @Param status query string false "状态"
// @Param keyword query string false "关键词"
// @Success 200 {object} response.Response{data=response.PaginationData}
// @Router /api/v1/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	var req request.ListUsersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	var operatorSchoolID *uint
	if !middleware.IsPlatformAdmin(c) {
		if schoolID, ok := middleware.GetSchoolID(c); ok {
			operatorSchoolID = &schoolID
		}
	}

	list, total, err := h.userService.ListUsers(c.Request.Context(), &req, operatorSchoolID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SuccessPage(c, list, total, req.GetPage(), req.GetPageSize())
}

// BatchImportStudents 批量导入学生
// @Summary 批量导入学生
// @Tags 用户管理
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body request.BatchImportStudentRequest true "批量导入请求"
// @Success 200 {object} response.Response{data=response.BatchImportResult}
// @Router /api/v1/users/batch-import [post]
func (h *UserHandler) BatchImportStudents(c *gin.Context) {
	schoolID, ok := middleware.GetSchoolID(c)
	if !ok {
		response.Error(c, errors.ErrNoPermission)
		return
	}

	var req request.BatchImportStudentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	result, err := h.userService.BatchImportStudents(c.Request.Context(), schoolID, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, result)
}

// UpdateUserStatus 更新用户状态
// @Summary 更新用户状态
// @Tags 用户管理
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Param request body request.StatusUpdateRequest true "状态更新请求"
// @Success 200 {object} response.Response
// @Router /api/v1/users/{id}/status [put]
func (h *UserHandler) UpdateUserStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var req request.StatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	operatorID, _ := middleware.GetUserID(c)

	if err := h.userService.UpdateUserStatus(c.Request.Context(), uint(id), operatorID, req.Status); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// DownloadImportTemplate 下载学生导入模板
func (h *UserHandler) DownloadImportTemplate(c *gin.Context) {
	// BOM + CSV header for Excel UTF-8 compatibility
	bom := "\xEF\xBB\xBF"
	header := "手机号,学号,姓名,邮箱,密码\n"
	example := "13800138001,20230001,张三,zhangsan@example.com,ChainSpace2025\n"

	c.Header("Content-Disposition", "attachment; filename=student_import_template.csv")
	c.Data(http.StatusOK, "text/csv; charset=utf-8", []byte(bom+header+example))
}
