package handler

import (
	"fmt"
	"strconv"

	"github.com/chainspace/backend/internal/dto/request"
	"github.com/chainspace/backend/internal/middleware"
	"github.com/chainspace/backend/internal/pkg/response"
	"github.com/chainspace/backend/internal/service"
	"github.com/chainspace/backend/pkg/errors"
	"github.com/gin-gonic/gin"
)

// CourseHandler 课程处理器
type CourseHandler struct {
	courseService *service.CourseService
}

// NewCourseHandler 创建课程处理器
func NewCourseHandler(courseService *service.CourseService) *CourseHandler {
	return &CourseHandler{courseService: courseService}
}

// CreateCourse 创建课程
func (h *CourseHandler) CreateCourse(c *gin.Context) {
	schoolID, ok := middleware.GetSchoolID(c)
	if !ok {
		response.Error(c, errors.ErrNoPermission)
		return
	}
	teacherID, _ := middleware.GetUserID(c)

	var req request.CreateCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	resp, err := h.courseService.CreateCourse(c.Request.Context(), schoolID, teacherID, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

// UpdateCourse 更新课程
func (h *CourseHandler) UpdateCourse(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var req request.UpdateCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	userID, _ := middleware.GetUserID(c)
	schoolID, _ := middleware.GetSchoolID(c)
	role, _ := middleware.GetRole(c)

	resp, err := h.courseService.UpdateCourse(c.Request.Context(), uint(id), userID, schoolID, role, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

// DeleteCourse 删除课程
func (h *CourseHandler) DeleteCourse(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	userID, _ := middleware.GetUserID(c)
	schoolID, _ := middleware.GetSchoolID(c)
	role, _ := middleware.GetRole(c)

	if err := h.courseService.DeleteCourse(c.Request.Context(), uint(id), userID, schoolID, role); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// GetCourse 获取课程
func (h *CourseHandler) GetCourse(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var userID *uint
	if uid, ok := middleware.GetUserID(c); ok {
		userID = &uid
	}
	schoolID, _ := middleware.GetSchoolID(c)
	role, _ := middleware.GetRole(c)

	resp, err := h.courseService.GetCourse(c.Request.Context(), uint(id), userID, schoolID, role)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

// ListCourses 获取课程列表
func (h *CourseHandler) ListCourses(c *gin.Context) {
	schoolID, _ := middleware.GetSchoolID(c)

	var req request.ListCoursesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	list, total, err := h.courseService.ListCourses(c.Request.Context(), schoolID, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SuccessPage(c, list, total, req.GetPage(), req.GetPageSize())
}

// ListMyCourses 获取我的课程
func (h *CourseHandler) ListMyCourses(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req request.PaginationRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	list, total, err := h.courseService.ListMyCourses(c.Request.Context(), userID, req.GetPage(), req.GetPageSize())
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SuccessPage(c, list, total, req.GetPage(), req.GetPageSize())
}

// JoinCourse 加入课程
func (h *CourseHandler) JoinCourse(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req request.JoinCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	if err := h.courseService.JoinCourse(c.Request.Context(), userID, &req); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// AddStudents 添加学生到课程
func (h *CourseHandler) AddStudents(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var req request.AddStudentsToCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	// 校验：至少要有一个有效的查询条件
	if len(req.Phones) == 0 && len(req.StudentNos) == 0 {
		response.Error(c, errors.ErrInvalidParams.WithError(fmt.Errorf("请提供手机号或学号")))
		return
	}

	userID, _ := middleware.GetUserID(c)
	schoolID, _ := middleware.GetSchoolID(c)
	role, _ := middleware.GetRole(c)

	if err := h.courseService.AddStudentsToCourse(c.Request.Context(), uint(id), userID, schoolID, role, req.Phones, req.StudentNos); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// RemoveStudents 从课程移除学生
func (h *CourseHandler) RemoveStudents(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var req request.RemoveStudentsFromCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	userID, _ := middleware.GetUserID(c)
	schoolID, _ := middleware.GetSchoolID(c)
	role, _ := middleware.GetRole(c)

	if err := h.courseService.RemoveStudentsFromCourse(c.Request.Context(), uint(id), userID, schoolID, role, req.StudentIDs); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// ListCourseStudents 获取课程学生列表
func (h *CourseHandler) ListCourseStudents(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var req request.ListCourseStudentsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	userID, _ := middleware.GetUserID(c)
	schoolID, _ := middleware.GetSchoolID(c)
	role, _ := middleware.GetRole(c)

	list, total, err := h.courseService.ListCourseStudents(c.Request.Context(), uint(id), userID, schoolID, role, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SuccessPage(c, list, total, req.GetPage(), req.GetPageSize())
}

// ImportStudentsFromExcel 从Excel导入学生到课程
func (h *CourseHandler) ImportStudentsFromExcel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	userID, _ := middleware.GetUserID(c)
	schoolID, _ := middleware.GetSchoolID(c)
	role, _ := middleware.GetRole(c)

	result, err := h.courseService.ImportStudentsFromExcel(c.Request.Context(), uint(id), userID, schoolID, role, file)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, result)
}

// CreateChapter 创建章节
func (h *CourseHandler) CreateChapter(c *gin.Context) {
	courseID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var req request.CreateChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	userID, _ := middleware.GetUserID(c)
	schoolID, _ := middleware.GetSchoolID(c)
	role, _ := middleware.GetRole(c)

	resp, err := h.courseService.CreateChapter(c.Request.Context(), uint(courseID), userID, schoolID, role, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

// UpdateChapter 更新章节
func (h *CourseHandler) UpdateChapter(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("chapter_id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var req request.UpdateChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	userID, _ := middleware.GetUserID(c)
	schoolID, _ := middleware.GetSchoolID(c)
	role, _ := middleware.GetRole(c)

	resp, err := h.courseService.UpdateChapter(c.Request.Context(), uint(id), userID, schoolID, role, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

// DeleteChapter 删除章节
func (h *CourseHandler) DeleteChapter(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("chapter_id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	userID, _ := middleware.GetUserID(c)
	schoolID, _ := middleware.GetSchoolID(c)
	role, _ := middleware.GetRole(c)

	if err := h.courseService.DeleteChapter(c.Request.Context(), uint(id), userID, schoolID, role); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// CreateMaterial 创建资料
func (h *CourseHandler) CreateMaterial(c *gin.Context) {
	chapterID, err := strconv.ParseUint(c.Param("chapter_id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var req request.CreateMaterialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	userID, _ := middleware.GetUserID(c)
	schoolID, _ := middleware.GetSchoolID(c)
	role, _ := middleware.GetRole(c)

	resp, err := h.courseService.CreateMaterial(c.Request.Context(), uint(chapterID), userID, schoolID, role, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

// UpdateMaterial 更新资料
func (h *CourseHandler) UpdateMaterial(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("material_id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var req request.UpdateMaterialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	userID, _ := middleware.GetUserID(c)
	schoolID, _ := middleware.GetSchoolID(c)
	role, _ := middleware.GetRole(c)

	resp, err := h.courseService.UpdateMaterial(c.Request.Context(), uint(id), userID, schoolID, role, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

// DeleteMaterial 删除资料
func (h *CourseHandler) DeleteMaterial(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("material_id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	userID, _ := middleware.GetUserID(c)
	schoolID, _ := middleware.GetSchoolID(c)
	role, _ := middleware.GetRole(c)

	if err := h.courseService.DeleteMaterial(c.Request.Context(), uint(id), userID, schoolID, role); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// UpdateMaterialProgress 更新学习进度
func (h *CourseHandler) UpdateMaterialProgress(c *gin.Context) {
	courseID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	materialID, err := strconv.ParseUint(c.Param("material_id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	userID, _ := middleware.GetUserID(c)
	schoolID, _ := middleware.GetSchoolID(c)
	role, _ := middleware.GetRole(c)

	var req request.UpdateMaterialProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	resp, err := h.courseService.UpdateMaterialProgress(c.Request.Context(), uint(courseID), uint(materialID), userID, schoolID, role, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

// ResetInviteCode 重置课程邀请码
func (h *CourseHandler) ResetInviteCode(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	userID, _ := middleware.GetUserID(c)
	schoolID, _ := middleware.GetSchoolID(c)
	role, _ := middleware.GetRole(c)

	newCode, err := h.courseService.ResetInviteCode(c.Request.Context(), uint(id), userID, schoolID, role)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, gin.H{"invite_code": newCode})
}

// UpdateCourseStatus 更新课程状态（发布/归档）
func (h *CourseHandler) UpdateCourseStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=draft published archived"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	userID, _ := middleware.GetUserID(c)
	schoolID, _ := middleware.GetSchoolID(c)
	role, _ := middleware.GetRole(c)

	if err := h.courseService.UpdateCourseStatus(c.Request.Context(), uint(id), userID, schoolID, role, req.Status); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// ListChapters 获取课程章节列表
func (h *CourseHandler) ListChapters(c *gin.Context) {
	courseID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var userID *uint
	if uid, ok := middleware.GetUserID(c); ok {
		userID = &uid
	}
	schoolID, _ := middleware.GetSchoolID(c)
	role, _ := middleware.GetRole(c)

	list, err := h.courseService.ListChapters(c.Request.Context(), uint(courseID), userID, schoolID, role)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, list)
}

// ListMaterials 获取章节资料列表
func (h *CourseHandler) ListMaterials(c *gin.Context) {
	courseID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	chapterID, err := strconv.ParseUint(c.Param("chapter_id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var userID *uint
	if uid, ok := middleware.GetUserID(c); ok {
		userID = &uid
	}
	schoolID, _ := middleware.GetSchoolID(c)
	role, _ := middleware.GetRole(c)

	list, err := h.courseService.ListMaterials(c.Request.Context(), uint(courseID), uint(chapterID), userID, schoolID, role)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, list)
}

// ReorderChapters 调整章节顺序
func (h *CourseHandler) ReorderChapters(c *gin.Context) {
	courseID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var req struct {
		ChapterIDs []uint `json:"chapter_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	userID, _ := middleware.GetUserID(c)
	schoolID, _ := middleware.GetSchoolID(c)
	role, _ := middleware.GetRole(c)

	if err := h.courseService.ReorderChapters(c.Request.Context(), uint(courseID), userID, schoolID, role, req.ChapterIDs); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// GetCourseProgress 获取课程学习进度
func (h *CourseHandler) GetCourseProgress(c *gin.Context) {
	courseID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	userID, _ := middleware.GetUserID(c)
	schoolID, _ := middleware.GetSchoolID(c)
	role, _ := middleware.GetRole(c)

	progress, err := h.courseService.GetCourseProgress(c.Request.Context(), uint(courseID), userID, schoolID, role)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, progress)
}
