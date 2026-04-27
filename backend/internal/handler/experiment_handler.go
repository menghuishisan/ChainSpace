package handler

import (
	"strconv"
	"strings"

	"github.com/chainspace/backend/internal/dto/request"
	"github.com/chainspace/backend/internal/middleware"
	"github.com/chainspace/backend/internal/pkg/response"
	"github.com/chainspace/backend/internal/service"
	"github.com/chainspace/backend/pkg/errors"
	"github.com/gin-gonic/gin"
)

// ExperimentHandler 实验处理器
type ExperimentHandler struct {
	experimentService *service.ExperimentService
}

// NewExperimentHandler 创建实验处理器
func NewExperimentHandler(experimentService *service.ExperimentService) *ExperimentHandler {
	return &ExperimentHandler{experimentService: experimentService}
}

// CreateExperiment 创建实验
func (h *ExperimentHandler) CreateExperiment(c *gin.Context) {
	schoolID, ok := middleware.GetSchoolID(c)
	if !ok {
		response.Error(c, errors.ErrNoPermission)
		return
	}
	creatorID, _ := middleware.GetUserID(c)

	var req request.CreateExperimentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	resp, err := h.experimentService.CreateExperiment(c.Request.Context(), schoolID, creatorID, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

// UpdateExperiment 更新实验
func (h *ExperimentHandler) UpdateExperiment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var req request.UpdateExperimentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	userID, _ := middleware.GetUserID(c)
	schoolID, _ := middleware.GetSchoolID(c)
	role, _ := middleware.GetRole(c)

	resp, err := h.experimentService.UpdateExperiment(c.Request.Context(), uint(id), userID, schoolID, role, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

// DeleteExperiment 删除实验
func (h *ExperimentHandler) DeleteExperiment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	userID, _ := middleware.GetUserID(c)
	schoolID, _ := middleware.GetSchoolID(c)
	role, _ := middleware.GetRole(c)

	if err := h.experimentService.DeleteExperiment(c.Request.Context(), uint(id), userID, schoolID, role); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// PublishExperiment 发布实验
func (h *ExperimentHandler) PublishExperiment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	userID, _ := middleware.GetUserID(c)
	schoolID, _ := middleware.GetSchoolID(c)
	role, _ := middleware.GetRole(c)

	if err := h.experimentService.PublishExperiment(c.Request.Context(), uint(id), userID, schoolID, role); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// GetExperiment 获取实验
func (h *ExperimentHandler) GetExperiment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var userID *uint
	if uid, ok := middleware.GetUserID(c); ok {
		userID = &uid
	}
	role, _ := middleware.GetRole(c)

	schoolID, _ := middleware.GetSchoolID(c)
	resp, err := h.experimentService.GetExperiment(c.Request.Context(), uint(id), userID, schoolID, role)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

// ListExperiments 获取实验列表
func (h *ExperimentHandler) ListExperiments(c *gin.Context) {
	schoolID, _ := middleware.GetSchoolID(c)
	role, _ := middleware.GetRole(c)

	var req request.ListExperimentsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	list, total, err := h.experimentService.ListExperiments(c.Request.Context(), schoolID, role, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SuccessPage(c, list, total, req.GetPage(), req.GetPageSize())
}

// StartEnv 启动实验环境
func (h *ExperimentHandler) StartEnv(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	schoolID, _ := middleware.GetSchoolID(c)
	role, _ := middleware.GetRole(c)

	var req request.StartEnvRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	resp, err := h.experimentService.StartEnv(c.Request.Context(), userID, schoolID, role, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

// StopEnv 停止实验环境
func (h *ExperimentHandler) StopEnv(c *gin.Context) {
	envID := c.Param("env_id")
	userID, _ := middleware.GetUserID(c)

	if err := h.experimentService.StopEnv(c.Request.Context(), envID, userID); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// GetEnvStatus 获取环境状态
func (h *ExperimentHandler) GetEnvStatus(c *gin.Context) {
	envID := c.Param("env_id")
	userID, _ := middleware.GetUserID(c)
	schoolID, _ := middleware.GetSchoolID(c)
	role, _ := middleware.GetRole(c)

	resp, err := h.experimentService.GetEnvStatus(c.Request.Context(), envID, userID, schoolID, role)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

// ExtendEnv 延长实验环境
func (h *ExperimentHandler) ExtendEnv(c *gin.Context) {
	envID := c.Param("env_id")
	userID, _ := middleware.GetUserID(c)

	var req request.ExtendEnvRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	if err := h.experimentService.ExtendEnv(c.Request.Context(), envID, userID, &req); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// SubmitExperiment 提交实验
func (h *ExperimentHandler) ListEnvs(c *gin.Context) {
	schoolID, _ := middleware.GetSchoolID(c)
	userID, _ := middleware.GetUserID(c)
	role, _ := middleware.GetRole(c)

	var req request.ListEnvsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	list, total, err := h.experimentService.ListEnvs(c.Request.Context(), userID, schoolID, role, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SuccessPage(c, list, total, req.GetPage(), req.GetPageSize())
}

func (h *ExperimentHandler) ListSessions(c *gin.Context) {
	schoolID, _ := middleware.GetSchoolID(c)
	userID, _ := middleware.GetUserID(c)
	role, _ := middleware.GetRole(c)

	var req request.ListExperimentSessionsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	list, total, err := h.experimentService.ListSessions(c.Request.Context(), userID, schoolID, role, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SuccessPage(c, list, total, req.GetPage(), req.GetPageSize())
}

func (h *ExperimentHandler) GetSession(c *gin.Context) {
	sessionKey := c.Param("session_key")
	schoolID, _ := middleware.GetSchoolID(c)
	userID, _ := middleware.GetUserID(c)
	role, _ := middleware.GetRole(c)

	resp, err := h.experimentService.GetSession(c.Request.Context(), sessionKey, userID, schoolID, role)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

func (h *ExperimentHandler) ListSessionMessages(c *gin.Context) {
	sessionKey := c.Param("session_key")
	schoolID, _ := middleware.GetSchoolID(c)
	userID, _ := middleware.GetUserID(c)
	role, _ := middleware.GetRole(c)

	var req request.ListExperimentSessionMessagesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	list, total, err := h.experimentService.ListSessionMessages(c.Request.Context(), sessionKey, userID, schoolID, role, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SuccessPage(c, list, total, req.GetPage(), req.GetPageSize())
}

func (h *ExperimentHandler) SendSessionMessage(c *gin.Context) {
	sessionKey := c.Param("session_key")
	schoolID, _ := middleware.GetSchoolID(c)
	userID, _ := middleware.GetUserID(c)
	role, _ := middleware.GetRole(c)

	var req request.SendExperimentSessionMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	resp, err := h.experimentService.SendSessionMessage(c.Request.Context(), sessionKey, userID, schoolID, role, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

func (h *ExperimentHandler) UpdateSessionMember(c *gin.Context) {
	sessionKey := c.Param("session_key")
	targetUserID, err := strconv.ParseUint(c.Param("user_id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}
	schoolID, _ := middleware.GetSchoolID(c)
	userID, _ := middleware.GetUserID(c)
	role, _ := middleware.GetRole(c)

	var req request.UpdateExperimentSessionMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	resp, err := h.experimentService.UpdateSessionMember(c.Request.Context(), sessionKey, uint(targetUserID), userID, schoolID, role, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

func (h *ExperimentHandler) JoinSession(c *gin.Context) {
	sessionKey := c.Param("session_key")
	schoolID, _ := middleware.GetSchoolID(c)
	userID, _ := middleware.GetUserID(c)
	role, _ := middleware.GetRole(c)

	resp, err := h.experimentService.JoinSession(c.Request.Context(), sessionKey, userID, schoolID, role)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

func (h *ExperimentHandler) LeaveSession(c *gin.Context) {
	sessionKey := c.Param("session_key")
	schoolID, _ := middleware.GetSchoolID(c)
	userID, _ := middleware.GetUserID(c)
	role, _ := middleware.GetRole(c)

	resp, err := h.experimentService.LeaveSession(c.Request.Context(), sessionKey, userID, schoolID, role)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

func (h *ExperimentHandler) GetSessionLogs(c *gin.Context) {
	sessionKey := c.Param("session_key")
	schoolID, _ := middleware.GetSchoolID(c)
	userID, _ := middleware.GetUserID(c)
	role, _ := middleware.GetRole(c)

	source := strings.TrimSpace(c.Query("source"))
	levels := strings.Split(c.Query("levels"), ",")
	logs, err := h.experimentService.ListSessionLogs(c.Request.Context(), sessionKey, userID, schoolID, role, source, levels)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, gin.H{"logs": logs})
}

func (h *ExperimentHandler) PauseEnv(c *gin.Context) {
	envID := c.Param("env_id")
	userID, _ := middleware.GetUserID(c)

	if err := h.experimentService.PauseEnv(c.Request.Context(), envID, userID); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

func (h *ExperimentHandler) ResumeEnv(c *gin.Context) {
	envID := c.Param("env_id")
	userID, _ := middleware.GetUserID(c)

	if err := h.experimentService.ResumeEnv(c.Request.Context(), envID, userID); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

func (h *ExperimentHandler) CreateSnapshot(c *gin.Context) {
	envID := c.Param("env_id")
	userID, _ := middleware.GetUserID(c)

	resp, err := h.experimentService.CreateSnapshot(c.Request.Context(), envID, userID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

func (h *ExperimentHandler) SubmitExperiment(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	schoolID, _ := middleware.GetSchoolID(c)

	var req request.SubmitExperimentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	role, _ := middleware.GetRole(c)
	resp, err := h.experimentService.SubmitExperiment(c.Request.Context(), userID, schoolID, role, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

// GradeSubmission 批改提交
func (h *ExperimentHandler) GradeSubmission(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	graderID, _ := middleware.GetUserID(c)

	var req request.GradeSubmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	schoolID, _ := middleware.GetSchoolID(c)
	role, _ := middleware.GetRole(c)

	resp, err := h.experimentService.GradeSubmission(c.Request.Context(), uint(id), graderID, schoolID, role, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

// ListSubmissions 获取提交列表
func (h *ExperimentHandler) ListSubmissions(c *gin.Context) {
	schoolID, _ := middleware.GetSchoolID(c)

	var req request.ListSubmissionsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	userID, _ := middleware.GetUserID(c)
	role, _ := middleware.GetRole(c)

	list, total, err := h.experimentService.ListSubmissions(c.Request.Context(), userID, schoolID, role, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SuccessPage(c, list, total, req.GetPage(), req.GetPageSize())
}
