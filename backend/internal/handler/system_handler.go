package handler

import (
	"strconv"

	"github.com/chainspace/backend/internal/dto/request"
	"github.com/chainspace/backend/internal/middleware"
	"github.com/chainspace/backend/internal/model"
	"github.com/chainspace/backend/internal/pkg/response"
	"github.com/chainspace/backend/internal/service"
	"github.com/chainspace/backend/pkg/errors"
	"github.com/gin-gonic/gin"
)

type SystemHandler struct {
	systemService        *service.SystemService
	vulnerabilityService *service.VulnerabilityAdminService
	schedulerService     *service.SchedulerService
	taskHandlerService   *service.TaskHandlerService
}

func NewSystemHandler(
	systemService *service.SystemService,
	vulnerabilityService *service.VulnerabilityAdminService,
	schedulerService *service.SchedulerService,
	taskHandlerService *service.TaskHandlerService,
) *SystemHandler {
	return &SystemHandler{
		systemService:        systemService,
		vulnerabilityService: vulnerabilityService,
		schedulerService:     schedulerService,
		taskHandlerService:   taskHandlerService,
	}
}

func (h *SystemHandler) GetStats(c *gin.Context) {
	var schoolID *uint
	role, _ := c.Get(middleware.ContextKeyRole)
	if role == model.RoleSchoolAdmin {
		if sid, ok := middleware.GetSchoolID(c); ok {
			schoolID = &sid
		}
	}

	stats, err := h.systemService.GetSystemStats(c.Request.Context(), schoolID)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, stats)
}

func (h *SystemHandler) HealthCheck(c *gin.Context) {
	result := h.systemService.HealthCheck(c.Request.Context())
	response.Success(c, result)
}

func (h *SystemHandler) GetConfig(c *gin.Context) {
	key := c.Param("key")
	config, err := h.systemService.GetConfig(c.Request.Context(), key)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, config)
}

func (h *SystemHandler) SetConfig(c *gin.Context) {
	var req struct {
		Key         string `json:"key" binding:"required"`
		Value       string `json:"value" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	if err := h.systemService.SetConfig(c.Request.Context(), req.Key, req.Value, req.Description); err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *SystemHandler) ListConfigs(c *gin.Context) {
	group := c.Query("group")
	configs, err := h.systemService.ListConfigs(c.Request.Context(), group)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, configs)
}

func (h *SystemHandler) ListOperationLogs(c *gin.Context) {
	var req request.ListOperationLogsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	if !middleware.IsPlatformAdmin(c) {
		if schoolID, ok := middleware.GetSchoolID(c); ok {
			req.SchoolID = &schoolID
		}
	}

	list, total, err := h.systemService.ListOperationLogs(c.Request.Context(), &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.SuccessPage(c, list, total, req.GetPage(), req.GetPageSize())
}

func (h *SystemHandler) CreateCrossSchoolApplication(c *gin.Context) {
	var req request.CreateCrossSchoolApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	userID, _ := middleware.GetUserID(c)
	req.ApplicantID = userID
	if schoolID, ok := middleware.GetSchoolID(c); ok {
		req.FromSchoolID = schoolID
	}

	app, err := h.systemService.CreateCrossSchoolApplication(c.Request.Context(), &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, app)
}

func (h *SystemHandler) HandleCrossSchoolApplication(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var req struct {
		Approved bool   `json:"approved"`
		Reason   string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	userID, _ := middleware.GetUserID(c)
	if err := h.systemService.HandleCrossSchoolApplication(c.Request.Context(), uint(id), userID, req.Approved, req.Reason); err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *SystemHandler) ListCrossSchoolApplications(c *gin.Context) {
	var req request.PaginationRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	schoolID, _ := middleware.GetSchoolID(c)
	status := c.Query("status")
	list, total, err := h.systemService.ListCrossSchoolApplications(c.Request.Context(), schoolID, status, req.GetPage(), req.GetPageSize())
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.SuccessPage(c, list, total, req.GetPage(), req.GetPageSize())
}

func (h *SystemHandler) GetSchedulerStatus(c *gin.Context) {
	status := h.schedulerService.GetTaskStatus()
	response.Success(c, status)
}

func (h *SystemHandler) RunSchedulerTask(c *gin.Context) {
	taskName := c.Param("task")
	if err := h.schedulerService.RunTaskNow(taskName); err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *SystemHandler) ListVulnerabilities(c *gin.Context) {
	var req request.ListVulnerabilitiesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	list, total, err := h.vulnerabilityService.ListVulnerabilities(c.Request.Context(), &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.SuccessPage(c, list, total, req.GetPage(), req.GetPageSize())
}

func (h *SystemHandler) UpdateVulnerability(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var req request.UpdateVulnerabilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	if err := h.vulnerabilityService.UpdateVulnerability(c.Request.Context(), uint(id), &req); err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *SystemHandler) ConvertVulnerability(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	userID, _ := middleware.GetUserID(c)
	challenge, err := h.vulnerabilityService.ConvertVulnerabilityToChallenge(c.Request.Context(), uint(id), userID)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, challenge)
}

func (h *SystemHandler) SkipVulnerability(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	if err := h.vulnerabilityService.SkipVulnerability(c.Request.Context(), uint(id)); err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *SystemHandler) SyncVulnerabilities(c *gin.Context) {
	sourceID, _ := strconv.ParseUint(c.Query("source_id"), 10, 32)

	taskID, err := h.taskHandlerService.SubmitVulnSyncTask(c.Request.Context(), uint(sourceID))
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SuccessWithMessage(c, "同步任务已提交", gin.H{
		"task_id": taskID,
		"type":    service.TaskTypeVulnSync,
	})
}

func (h *SystemHandler) EnrichVulnerabilityCode(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	taskID, err := h.taskHandlerService.SubmitVulnEnrichTask(c.Request.Context(), uint(id))
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SuccessWithMessage(c, "源码增强任务已提交", gin.H{
		"task_id": taskID,
		"type":    service.TaskTypeVulnEnrich,
	})
}

func (h *SystemHandler) ListChallengePublishRequests(c *gin.Context) {
	var req request.PaginationRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	schoolID, _ := middleware.GetSchoolID(c)
	status := c.Query("status")
	list, total, err := h.systemService.ListChallengePublishRequests(c.Request.Context(), schoolID, status, req.GetPage(), req.GetPageSize())
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.SuccessPage(c, list, total, req.GetPage(), req.GetPageSize())
}

func (h *SystemHandler) GetSystemMonitor(c *gin.Context) {
	monitor, err := h.systemService.GetSystemMonitor(c.Request.Context())
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, monitor)
}

func (h *SystemHandler) GetContainerStats(c *gin.Context) {
	stats, err := h.systemService.GetContainerStats(c.Request.Context())
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, stats)
}

func (h *SystemHandler) GetServiceHealth(c *gin.Context) {
	health, err := h.systemService.GetServiceHealth(c.Request.Context())
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, health)
}

func (h *SystemHandler) HandleChallengePublishRequest(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var req struct {
		Action  string `json:"action" binding:"required,oneof=approve reject"`
		Comment string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	userID, _ := middleware.GetUserID(c)
	approved := req.Action == "approve"
	if err := h.systemService.HandleChallengePublishRequest(c.Request.Context(), uint(id), userID, approved, req.Comment); err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, nil)
}
