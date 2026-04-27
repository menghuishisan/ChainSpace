package handler

import (
	"strconv"

	"github.com/chainspace/backend/internal/dto/request"
	"github.com/chainspace/backend/internal/middleware"
	"github.com/chainspace/backend/internal/pkg/response"
	"github.com/chainspace/backend/internal/service"
	"github.com/chainspace/backend/pkg/errors"
	"github.com/gin-gonic/gin"
)

// SchoolHandler 学校处理器
type SchoolHandler struct {
	schoolService *service.SchoolService
}

// NewSchoolHandler 创建学校处理器
func NewSchoolHandler(schoolService *service.SchoolService) *SchoolHandler {
	return &SchoolHandler{schoolService: schoolService}
}

// CreateSchool 创建学校
func (h *SchoolHandler) CreateSchool(c *gin.Context) {
	var req request.CreateSchoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	resp, err := h.schoolService.CreateSchool(c.Request.Context(), &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

// UpdateSchool 更新学校
func (h *SchoolHandler) UpdateSchool(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var req request.UpdateSchoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	resp, err := h.schoolService.UpdateSchool(c.Request.Context(), uint(id), &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

// DeleteSchool 删除学校
func (h *SchoolHandler) DeleteSchool(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	if err := h.schoolService.DeleteSchool(c.Request.Context(), uint(id)); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// UpdateSchoolStatus 更新学校状态
func (h *SchoolHandler) UpdateSchoolStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=active disabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	if err := h.schoolService.UpdateSchoolStatus(c.Request.Context(), uint(id), req.Status); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// GetSchool 获取学校
func (h *SchoolHandler) GetSchool(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	resp, err := h.schoolService.GetSchool(c.Request.Context(), uint(id))
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

// ListSchools 获取学校列表
func (h *SchoolHandler) ListSchools(c *gin.Context) {
	var req request.ListSchoolsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	list, total, err := h.schoolService.ListSchools(c.Request.Context(), &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SuccessPage(c, list, total, req.GetPage(), req.GetPageSize())
}

// GetCurrentSchool 获取当前登录用户所属学校信息
func (h *SchoolHandler) GetCurrentSchool(c *gin.Context) {
	schoolID, ok := middleware.GetSchoolID(c)
	if !ok {
		response.Error(c, errors.ErrNoPermission)
		return
	}

	resp, err := h.schoolService.GetSchool(c.Request.Context(), schoolID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

// UpdateCurrentSchool 更新当前登录用户所属学校信息
func (h *SchoolHandler) UpdateCurrentSchool(c *gin.Context) {
	schoolID, ok := middleware.GetSchoolID(c)
	if !ok {
		response.Error(c, errors.ErrNoPermission)
		return
	}

	var req request.UpdateSchoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	resp, err := h.schoolService.UpdateSchool(c.Request.Context(), schoolID, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

// CreateClass 创建班级
func (h *SchoolHandler) CreateClass(c *gin.Context) {
	schoolID, ok := middleware.GetSchoolID(c)
	if !ok {
		response.Error(c, errors.ErrNoPermission)
		return
	}

	var req request.CreateClassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	resp, err := h.schoolService.CreateClass(c.Request.Context(), schoolID, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

// UpdateClass 更新班级
func (h *SchoolHandler) UpdateClass(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var req request.UpdateClassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	resp, err := h.schoolService.UpdateClass(c.Request.Context(), uint(id), &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

// DeleteClass 删除班级
func (h *SchoolHandler) DeleteClass(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	if err := h.schoolService.DeleteClass(c.Request.Context(), uint(id)); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// GetClass 获取班级
func (h *SchoolHandler) GetClass(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	resp, err := h.schoolService.GetClass(c.Request.Context(), uint(id))
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

// ListClasses 获取班级列表
func (h *SchoolHandler) ListClasses(c *gin.Context) {
	schoolID, ok := middleware.GetSchoolID(c)
	if !ok {
		response.Error(c, errors.ErrNoPermission)
		return
	}

	var req request.ListClassesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	list, total, err := h.schoolService.ListClasses(c.Request.Context(), schoolID, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SuccessPage(c, list, total, req.GetPage(), req.GetPageSize())
}

// ListClassStudents 获取班级学生列表
func (h *SchoolHandler) ListClassStudents(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var req request.PaginationRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	list, total, err := h.schoolService.ListClassStudents(c.Request.Context(), uint(id), req.GetPage(), req.GetPageSize())
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SuccessPage(c, list, total, req.GetPage(), req.GetPageSize())
}
