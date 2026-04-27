package handler

import (
	"fmt"

	"github.com/chainspace/backend/internal/middleware"
	"github.com/chainspace/backend/internal/pkg/response"
	"github.com/chainspace/backend/internal/service"
	"github.com/chainspace/backend/pkg/errors"
	"github.com/gin-gonic/gin"
)

// UploadHandler 上传处理器
type UploadHandler struct {
	uploadService *service.UploadService
}

// NewUploadHandler 创建上传处理器
func NewUploadHandler(uploadService *service.UploadService) *UploadHandler {
	return &UploadHandler{uploadService: uploadService}
}

// UploadAvatar 上传头像
func (h *UploadHandler) UploadAvatar(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	result, err := h.uploadService.UploadAvatar(c.Request.Context(), file, userID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, result)
}

// UploadCourseCover 上传课程封面
func (h *UploadHandler) UploadCourseCover(c *gin.Context) {
	var courseID uint
	if id, err := parseUint(c.Param("course_id")); err == nil {
		courseID = id
	}

	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	result, err := h.uploadService.UploadCourseCover(c.Request.Context(), file, courseID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, result)
}

// UploadMaterial 上传课程资料
func (h *UploadHandler) UploadMaterial(c *gin.Context) {
	courseID, _ := parseUint(c.Param("course_id"))
	chapterID, _ := parseUint(c.Param("chapter_id"))

	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	result, err := h.uploadService.UploadMaterial(c.Request.Context(), file, courseID, chapterID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, result)
}

// UploadSubmission 上传实验提交
func (h *UploadHandler) UploadSubmission(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	expID, _ := parseUint(c.Param("experiment_id"))

	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	result, err := h.uploadService.UploadSubmission(c.Request.Context(), file, expID, userID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, result)
}

// UploadExperimentAsset 上传实验资源文件
func (h *UploadHandler) UploadExperimentAsset(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	result, err := h.uploadService.UploadExperimentAsset(c.Request.Context(), file)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, result)
}

// UploadChallengeAttachment 上传题目附件
func (h *UploadHandler) UploadChallengeAttachment(c *gin.Context) {
	challengeID, _ := parseUint(c.Param("challenge_id"))

	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	result, err := h.uploadService.UploadChallengeAttachment(c.Request.Context(), file, challengeID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, result)
}

func parseUint(s string) (uint, error) {
	var id uint
	_, err := fmt.Sscanf(s, "%d", &id)
	return id, err
}
