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

// ChallengeHandler 题目处理器。
type ChallengeHandler struct {
	challengeService *service.ChallengeService
}

// NewChallengeHandler 创建题目处理器。
func NewChallengeHandler(challengeService *service.ChallengeService) *ChallengeHandler {
	return &ChallengeHandler{challengeService: challengeService}
}

// ListChallenges 获取题目列表。
func (h *ChallengeHandler) ListChallenges(c *gin.Context) {
	var req request.ListChallengesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	userID, _ := middleware.GetUserID(c)
	schoolID, _ := middleware.GetSchoolID(c)
	role, _ := middleware.GetRole(c)

	list, total, err := h.challengeService.ListChallenges(c.Request.Context(), &req, userID, schoolID, role)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SuccessPage(c, buildChallengeManageResponses(list), total, req.GetPage(), req.GetPageSize())
}

// GetChallenge 获取题目详情。
func (h *ChallengeHandler) GetChallenge(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	userID, _ := middleware.GetUserID(c)
	schoolID, _ := middleware.GetSchoolID(c)
	role, _ := middleware.GetRole(c)

	challenge, err := h.challengeService.GetChallenge(c.Request.Context(), uint(id), userID, schoolID, role)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, buildChallengeManageResponse(challenge))
}

// CreateChallenge 创建题目。
func (h *ChallengeHandler) CreateChallenge(c *gin.Context) {
	var req request.CreateChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	userID, _ := middleware.GetUserID(c)
	role, _ := middleware.GetRole(c)
	var schoolID *uint
	if sid, ok := middleware.GetSchoolID(c); ok {
		schoolID = &sid
	}

	challenge, err := h.challengeService.CreateChallenge(c.Request.Context(), &req, userID, schoolID, role)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, buildChallengeManageResponse(challenge))
}

// UpdateChallenge 更新题目。
func (h *ChallengeHandler) UpdateChallenge(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var req request.UpdateChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	userID, _ := middleware.GetUserID(c)
	schoolID, _ := middleware.GetSchoolID(c)
	role, _ := middleware.GetRole(c)

	challenge, err := h.challengeService.UpdateChallenge(c.Request.Context(), uint(id), userID, schoolID, role, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, buildChallengeManageResponse(challenge))
}

// DeleteChallenge 删除题目。
func (h *ChallengeHandler) DeleteChallenge(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	userID, _ := middleware.GetUserID(c)
	schoolID, _ := middleware.GetSchoolID(c)
	role, _ := middleware.GetRole(c)

	if err := h.challengeService.DeleteChallenge(c.Request.Context(), uint(id), userID, schoolID, role); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// RequestPublish 申请公开题目。
func (h *ChallengeHandler) RequestPublish(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var req request.RequestChallengePublishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	userID, _ := middleware.GetUserID(c)
	schoolID, _ := middleware.GetSchoolID(c)
	role, _ := middleware.GetRole(c)

	if err := h.challengeService.RequestPublish(c.Request.Context(), uint(id), userID, schoolID, role, req.Reason); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}
