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

// ContestHandler 竞赛处理器
type ContestHandler struct {
	contestService *service.ContestService
}

// NewContestHandler 创建竞赛处理器
func NewContestHandler(contestService *service.ContestService) *ContestHandler {
	return &ContestHandler{contestService: contestService}
}

// CreateContest 创建竞赛
func (h *ContestHandler) CreateContest(c *gin.Context) {
	creatorID, _ := middleware.GetUserID(c)
	var schoolID *uint
	if sid, ok := middleware.GetSchoolID(c); ok {
		schoolID = &sid
	}

	var req request.CreateContestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	resp, err := h.contestService.CreateContest(c.Request.Context(), creatorID, schoolID, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

// UpdateContest 更新竞赛
func (h *ContestHandler) UpdateContest(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var req request.UpdateContestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	userID, _ := middleware.GetUserID(c)
	schoolID, _ := middleware.GetSchoolID(c)
	role, _ := middleware.GetRole(c)

	resp, err := h.contestService.UpdateContest(c.Request.Context(), uint(id), userID, schoolID, role, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

// GetContest 获取竞赛
func (h *ContestHandler) GetContest(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	userID, _ := middleware.GetUserID(c)
	schoolID, _ := middleware.GetSchoolID(c)
	role, _ := middleware.GetRole(c)

	resp, err := h.contestService.GetContest(c.Request.Context(), uint(id), userID, schoolID, role)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

// ListContests 获取竞赛列表
func (h *ContestHandler) ListContests(c *gin.Context) {
	schoolID, _ := middleware.GetSchoolID(c)

	var req request.ListContestsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	userID, _ := middleware.GetUserID(c)
	role, _ := middleware.GetRole(c)
	list, total, err := h.contestService.ListContests(c.Request.Context(), schoolID, userID, role, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SuccessPage(c, list, total, req.GetPage(), req.GetPageSize())
}

// CreateTeam 创建队伍
func (h *ContestHandler) CreateTeam(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	var schoolID *uint
	if sid, ok := middleware.GetSchoolID(c); ok {
		schoolID = &sid
	}

	var req request.CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	role, _ := middleware.GetRole(c)
	resp, err := h.contestService.CreateTeam(c.Request.Context(), userID, schoolID, role, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

// JoinTeam 加入队伍
func (h *ContestHandler) JoinTeam(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req request.JoinTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	if err := h.contestService.JoinTeam(c.Request.Context(), userID, &req); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// SubmitFlag 提交Flag
func (h *ContestHandler) SubmitFlag(c *gin.Context) {
	contestID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	userID, _ := middleware.GetUserID(c)

	var req request.SubmitFlagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	schoolID, _ := middleware.GetSchoolID(c)
	role, _ := middleware.GetRole(c)

	resp, err := h.contestService.SubmitFlag(c.Request.Context(), userID, schoolID, role, uint(contestID), &req, c.ClientIP())
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

// GetScoreboard 获取排行榜
func (h *ContestHandler) GetScoreboard(c *gin.Context) {
	contestID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	userID, _ := middleware.GetUserID(c)

	schoolID, _ := middleware.GetSchoolID(c)
	role, _ := middleware.GetRole(c)

	result, err := h.contestService.GetScoreboard(c.Request.Context(), uint(contestID), userID, schoolID, role)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, result)
}

// RegisterContest 报名竞赛
func (h *ContestHandler) RegisterContest(c *gin.Context) {
	contestID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	userID, _ := middleware.GetUserID(c)

	schoolID, _ := middleware.GetSchoolID(c)
	role, _ := middleware.GetRole(c)

	if err := h.contestService.RegisterContest(c.Request.Context(), userID, schoolID, role, uint(contestID)); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// GetMyTeam 获取我在竞赛中的队伍
func (h *ContestHandler) GetMyTeam(c *gin.Context) {
	contestID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	userID, _ := middleware.GetUserID(c)

	resp, err := h.contestService.GetMyTeam(c.Request.Context(), userID, uint(contestID))
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

// GetMyContestRecords 获取我的竞赛记录
func (h *ContestHandler) GetMyContestRecords(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req request.PaginationRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	list, total, err := h.contestService.GetMyContestRecords(c.Request.Context(), userID, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SuccessPage(c, list, total, req.GetPage(), req.GetPageSize())
}

// GetMyTeams 获取我的所有队伍
func (h *ContestHandler) GetMyTeams(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	list, err := h.contestService.GetMyTeams(c.Request.Context(), userID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, list)
}

// GetContestReviewChallenges 获取已结束解题赛的赛后题目回顾列表。
func (h *ContestHandler) GetContestReviewChallenges(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	schoolID, _ := middleware.GetSchoolID(c)
	role, _ := middleware.GetRole(c)

	list, err := h.contestService.GetContestReviewChallenges(c.Request.Context(), uint(id), schoolID, role)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, list)
}

// LeaveTeam 离开队伍
func (h *ContestHandler) LeaveTeam(c *gin.Context) {
	teamID, err := strconv.ParseUint(c.Param("team_id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	userID, _ := middleware.GetUserID(c)

	if err := h.contestService.LeaveTeam(c.Request.Context(), userID, uint(teamID)); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// InviteTeamMember 邀请队伍成员
func (h *ContestHandler) InviteTeamMember(c *gin.Context) {
	teamID, err := strconv.ParseUint(c.Param("team_id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	userID, _ := middleware.GetUserID(c)

	var req request.InviteTeamMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	if err := h.contestService.InviteTeamMember(c.Request.Context(), userID, uint(teamID), &req); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// RemoveTeamMember 移除队伍成员
func (h *ContestHandler) RemoveTeamMember(c *gin.Context) {
	teamID, err := strconv.ParseUint(c.Param("team_id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	memberID, err := strconv.ParseUint(c.Param("user_id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	userID, _ := middleware.GetUserID(c)

	if err := h.contestService.RemoveTeamMember(c.Request.Context(), userID, uint(teamID), uint(memberID)); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// GetContestChallenges 获取竞赛题目列表
func (h *ContestHandler) GetContestChallenges(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	userID, _ := middleware.GetUserID(c)

	schoolID, _ := middleware.GetSchoolID(c)
	role, _ := middleware.GetRole(c)

	list, err := h.contestService.GetContestChallenges(c.Request.Context(), uint(id), userID, schoolID, role)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, list)
}

// GetChallengeAttachmentAccessURL 获取题目附件访问地址（前端携带鉴权请求后再打开新窗口）。
func (h *ContestHandler) GetChallengeAttachmentAccessURL(c *gin.Context) {
	contestID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}
	challengeID, err := strconv.ParseUint(c.Param("challenge_id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}
	attachmentIndex, err := strconv.ParseInt(c.Param("attachment_index"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	userID, _ := middleware.GetUserID(c)
	schoolID, _ := middleware.GetSchoolID(c)
	role, _ := middleware.GetRole(c)

	url, err := h.contestService.GetChallengeAttachmentAccessURL(
		c.Request.Context(),
		uint(contestID),
		uint(challengeID),
		int(attachmentIndex),
		userID,
		schoolID,
		role,
	)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, gin.H{"url": url})
}

// PublishContest 发布竞赛
func (h *ContestHandler) PublishContest(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	userID, _ := middleware.GetUserID(c)
	schoolID, _ := middleware.GetSchoolID(c)
	role, _ := middleware.GetRole(c)

	if err := h.contestService.PublishContest(c.Request.Context(), uint(id), userID, schoolID, role); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// AddChallengeToContest 添加题目到竞赛
func (h *ContestHandler) AddChallengeToContest(c *gin.Context) {
	contestID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var req request.AddChallengeToContestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	userID, _ := middleware.GetUserID(c)
	schoolID, _ := middleware.GetSchoolID(c)
	role, _ := middleware.GetRole(c)

	resp, err := h.contestService.AddChallengeToContest(c.Request.Context(), uint(contestID), userID, schoolID, role, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

// RemoveChallengeFromContest 从竞赛移除题目
func (h *ContestHandler) RemoveChallengeFromContest(c *gin.Context) {
	contestID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	challengeID, err := strconv.ParseUint(c.Param("challenge_id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	userID, _ := middleware.GetUserID(c)
	schoolID, _ := middleware.GetSchoolID(c)
	role, _ := middleware.GetRole(c)

	if err := h.contestService.RemoveChallengeFromContest(c.Request.Context(), uint(contestID), uint(challengeID), userID, schoolID, role); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// ListContestChallengesAdmin 获取竞赛题目列表（管理视角）
func (h *ContestHandler) ListContestChallengesAdmin(c *gin.Context) {
	contestID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	userID, _ := middleware.GetUserID(c)
	schoolID, _ := middleware.GetSchoolID(c)
	role, _ := middleware.GetRole(c)

	list, err := h.contestService.ListContestChallengesAdmin(c.Request.Context(), uint(contestID), userID, schoolID, role)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, list)
}

// DeleteContest 删除竞赛
func (h *ContestHandler) DeleteContest(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	userID, _ := middleware.GetUserID(c)
	schoolID, _ := middleware.GetSchoolID(c)
	role, _ := middleware.GetRole(c)

	if err := h.contestService.DeleteContest(c.Request.Context(), uint(id), userID, schoolID, role); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// UploadAgentCode 上传智能体代码
func (h *ContestHandler) UploadAgentCode(c *gin.Context) {
	contestID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	userID, _ := middleware.GetUserID(c)

	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams.WithMessage("请上传文件"))
		return
	}

	version, err := h.contestService.UploadAgentCode(c.Request.Context(), uint(contestID), userID, file)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, gin.H{"version": version})
}

// StartChallengeEnv 启动题目环境
func (h *ContestHandler) StartChallengeEnv(c *gin.Context) {
	contestID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}
	challengeID, err := strconv.ParseUint(c.Param("challenge_id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}
	userID, _ := middleware.GetUserID(c)

	resp, err := h.contestService.StartChallengeEnv(c.Request.Context(), userID, uint(contestID), uint(challengeID))
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, resp)
}

// GetChallengeEnvStatus 获取题目环境状态
func (h *ContestHandler) GetChallengeEnvStatus(c *gin.Context) {
	contestID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}
	challengeID, err := strconv.ParseUint(c.Param("challenge_id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}
	userID, _ := middleware.GetUserID(c)

	resp, err := h.contestService.GetChallengeEnvStatus(c.Request.Context(), userID, uint(contestID), uint(challengeID))
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, resp)
}

// StopChallengeEnv 停止题目环境
func (h *ContestHandler) StopChallengeEnv(c *gin.Context) {
	contestID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}
	challengeID, err := strconv.ParseUint(c.Param("challenge_id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}
	userID, _ := middleware.GetUserID(c)

	if err := h.contestService.StopChallengeEnv(c.Request.Context(), userID, uint(contestID), uint(challengeID)); err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, nil)
}
