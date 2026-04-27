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

// AgentBattleHandler 负责对抗赛相关 HTTP 请求处理。
type AgentBattleHandler struct {
	battleService *service.AgentBattleService
}

func NewAgentBattleHandler(battleService *service.AgentBattleService) *AgentBattleHandler {
	return &AgentBattleHandler{battleService: battleService}
}

func (h *AgentBattleHandler) CreateRound(c *gin.Context) {
	contestID, err := strconv.ParseUint(c.Param("contest_id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var req request.CreateAgentBattleRoundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	resp, err := h.battleService.CreateRound(c.Request.Context(), uint(contestID), &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

func (h *AgentBattleHandler) GetCurrentRound(c *gin.Context) {
	contestID, err := strconv.ParseUint(c.Param("contest_id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	resp, err := h.battleService.GetCurrentRound(c.Request.Context(), uint(contestID))
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

func (h *AgentBattleHandler) ListRounds(c *gin.Context) {
	contestID, err := strconv.ParseUint(c.Param("contest_id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var req request.ListAgentBattleRoundsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	list, total, err := h.battleService.ListRounds(c.Request.Context(), uint(contestID), &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SuccessPage(c, list, total, req.GetPage(), req.GetPageSize())
}

func (h *AgentBattleHandler) DeployContract(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req request.DeployAgentContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	resp, err := h.battleService.DeployContract(c.Request.Context(), userID, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

func (h *AgentBattleHandler) GetContract(c *gin.Context) {
	contestID, err := strconv.ParseUint(c.Param("contest_id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	teamID, err := strconv.ParseUint(c.Param("team_id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	resp, err := h.battleService.GetContractByTeam(c.Request.Context(), uint(contestID), uint(teamID))
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

func (h *AgentBattleHandler) GetScoreboard(c *gin.Context) {
	contestID, err := strconv.ParseUint(c.Param("contest_id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var roundID *uint
	if roundIDStr := c.Query("round_id"); roundIDStr != "" {
		if id, parseErr := strconv.ParseUint(roundIDStr, 10, 32); parseErr == nil {
			rid := uint(id)
			roundID = &rid
		}
	}

	list, err := h.battleService.GetScoreboard(c.Request.Context(), uint(contestID), roundID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, list)
}

func (h *AgentBattleHandler) ListEvents(c *gin.Context) {
	roundID, err := strconv.ParseUint(c.Param("round_id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var req request.ListAgentBattleEventsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	list, total, err := h.battleService.ListEvents(c.Request.Context(), uint(roundID), &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SuccessPage(c, list, total, req.GetPage(), req.GetPageSize())
}

func (h *AgentBattleHandler) GetSpectateData(c *gin.Context) {
	contestID, err := strconv.ParseUint(c.Param("contest_id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	resp, err := h.battleService.GetSpectateData(c.Request.Context(), uint(contestID))
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

func (h *AgentBattleHandler) GetReplayData(c *gin.Context) {
	contestID, err := strconv.ParseUint(c.Param("contest_id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	roundID, err := strconv.ParseUint(c.Query("round_id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var fromBlock, toBlock *uint64
	if fb := c.Query("from_block"); fb != "" {
		if v, parseErr := strconv.ParseUint(fb, 10, 64); parseErr == nil {
			fromBlock = &v
		}
	}
	if tb := c.Query("to_block"); tb != "" {
		if v, parseErr := strconv.ParseUint(tb, 10, 64); parseErr == nil {
			toBlock = &v
		}
	}

	resp, err := h.battleService.GetReplayData(c.Request.Context(), uint(contestID), uint(roundID), fromBlock, toBlock)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

func (h *AgentBattleHandler) GetBattleConfig(c *gin.Context) {
	contestID, err := strconv.ParseUint(c.Param("contest_id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	resp, err := h.battleService.GetBattleConfig(c.Request.Context(), uint(contestID))
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

func (h *AgentBattleHandler) UpgradeContract(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req request.UpgradeAgentContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	resp, err := h.battleService.UpgradeContract(c.Request.Context(), userID, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

func (h *AgentBattleHandler) GetFinalRank(c *gin.Context) {
	contestID, err := strconv.ParseUint(c.Param("contest_id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	resp, err := h.battleService.GetFinalRank(c.Request.Context(), uint(contestID))
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

func (h *AgentBattleHandler) GetStatus(c *gin.Context) {
	contestID, err := strconv.ParseUint(c.Param("contest_id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	userID, _ := middleware.GetUserID(c)
	resp, err := h.battleService.GetBattleStatus(c.Request.Context(), uint(contestID), userID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

func (h *AgentBattleHandler) GetContestEvents(c *gin.Context) {
	contestID, err := strconv.ParseUint(c.Param("contest_id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var req request.ListAgentBattleEventsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	currentRound, err := h.battleService.GetCurrentRound(c.Request.Context(), uint(contestID))
	if err != nil {
		response.HandleError(c, err)
		return
	}
	if currentRound == nil {
		response.Success(c, []interface{}{})
		return
	}

	list, _, err := h.battleService.ListEvents(c.Request.Context(), currentRound.ID, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, list)
}

func (h *AgentBattleHandler) GetMyTeamWorkspace(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	contestID, err := strconv.ParseUint(c.Param("contest_id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	resp, err := h.battleService.GetTeamWorkspaceForUser(c.Request.Context(), userID, uint(contestID))
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

func (h *AgentBattleHandler) CreateTeamWorkspace(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	contestID, err := strconv.ParseUint(c.Param("contest_id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	team, err := h.battleService.GetTeamWorkspaceForUser(c.Request.Context(), userID, uint(contestID))
	if err != nil {
		response.HandleError(c, err)
		return
	}

	resp, err := h.battleService.CreateOrGetTeamWorkspace(c.Request.Context(), uint(contestID), team.TeamID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, resp)
}

func (h *AgentBattleHandler) StopTeamWorkspace(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	contestID, err := strconv.ParseUint(c.Param("contest_id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	team, err := h.battleService.GetTeamWorkspaceForUser(c.Request.Context(), userID, uint(contestID))
	if err != nil {
		response.HandleError(c, err)
		return
	}

	if err := h.battleService.StopTeamWorkspace(c.Request.Context(), uint(contestID), team.TeamID); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}
