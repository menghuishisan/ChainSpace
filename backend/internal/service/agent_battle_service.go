package service

import (
	"context"
	"fmt"
	"math/big"
	"sort"
	"strings"
	"time"

	"github.com/chainspace/backend/internal/dto/request"
	"github.com/chainspace/backend/internal/dto/response"
	"github.com/chainspace/backend/internal/model"
	"github.com/chainspace/backend/internal/pkg/logger"
	"github.com/chainspace/backend/internal/pkg/safego"
	"github.com/chainspace/backend/internal/repository"
	"github.com/chainspace/backend/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AgentBattleService 负责对抗赛业务编排与对外入口。
type AgentBattleService struct {
	roundRepo        *repository.AgentBattleRoundRepository
	contractRepo     *repository.AgentContractRepository
	eventRepo        *repository.AgentBattleEventRepository
	scoreRepo        *repository.AgentBattleScoreRepository
	contestRepo      *repository.ContestRepository
	teamRepo         *repository.TeamRepository
	chainRuntime     *BattleChainRuntimeService
	scoreService     *AgentBattleScoreService
	contractService  *AgentContractService
	workspaceService *BattleWorkspaceService
}

func NewAgentBattleService(
	roundRepo *repository.AgentBattleRoundRepository,
	contractRepo *repository.AgentContractRepository,
	eventRepo *repository.AgentBattleEventRepository,
	scoreRepo *repository.AgentBattleScoreRepository,
	contestRepo *repository.ContestRepository,
	teamRepo *repository.TeamRepository,
	chainRuntime *BattleChainRuntimeService,
	scoreService *AgentBattleScoreService,
	contractService *AgentContractService,
	workspaceService *BattleWorkspaceService,
) *AgentBattleService {
	return &AgentBattleService{
		roundRepo:        roundRepo,
		contractRepo:     contractRepo,
		eventRepo:        eventRepo,
		scoreRepo:        scoreRepo,
		contestRepo:      contestRepo,
		teamRepo:         teamRepo,
		chainRuntime:     chainRuntime,
		scoreService:     scoreService,
		contractService:  contractService,
		workspaceService: workspaceService,
	}
}

func (s *AgentBattleService) CreateRound(ctx context.Context, contestID uint, req *request.CreateAgentBattleRoundRequest) (*response.AgentBattleRoundResponse, error) {
	contest, err := s.contestRepo.GetByID(ctx, contestID)
	if err != nil {
		return nil, errors.ErrContestNotFound
	}
	if contest.Type != "agent_battle" {
		return nil, errors.ErrBadRequest.WithMessage("竞赛类型不是智能体对抗")
	}

	currentRound, err := s.roundRepo.GetCurrentRound(ctx, contestID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, errors.ErrDatabaseError.WithError(err)
	}
	if currentRound == nil {
		latestRound, latestErr := s.roundRepo.GetLatestRound(ctx, contestID)
		if latestErr == nil {
			currentRound = latestRound
		}
	}

	roundNumber := 1
	if currentRound != nil {
		roundNumber = currentRound.RoundNumber + 1
	}
	if req.RoundNumber > 0 && req.RoundNumber != roundNumber {
		return nil, errors.ErrInvalidParams.WithMessage(fmt.Sprintf("轮次编号应为 %d", roundNumber))
	}

	round := &model.AgentBattleRound{
		ContestID:   contestID,
		RoundNumber: roundNumber,
		Status:      model.RoundStatusPending,
		Summary:     req.Description,
	}

	if req.StartTime != "" {
		t, err := time.Parse(time.RFC3339, req.StartTime)
		if err != nil {
			return nil, errors.ErrInvalidParams.WithMessage("开始时间格式错误，需要 RFC3339")
		}
		round.StartTime = t
	}
	if req.EndTime != "" {
		t, err := time.Parse(time.RFC3339, req.EndTime)
		if err != nil {
			return nil, errors.ErrInvalidParams.WithMessage("结束时间格式错误，需要 RFC3339")
		}
		round.EndTime = t
	}

	if err := s.roundRepo.Create(ctx, round); err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}
	contest, _ = s.contestRepo.GetByID(ctx, contestID)
	round.Contest = contest
	return s.buildRoundResponse(round), nil
}

func (s *AgentBattleService) GetCurrentRound(ctx context.Context, contestID uint) (*response.AgentBattleRoundResponse, error) {
	if err := s.SyncContestRoundLifecycle(ctx, contestID); err != nil {
		return nil, err
	}
	round, err := s.roundRepo.GetCurrentRound(ctx, contestID)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, errors.ErrDatabaseError.WithError(err)
		}
		latestRound, latestErr := s.roundRepo.GetLatestRound(ctx, contestID)
		if latestErr != nil {
			if latestErr == gorm.ErrRecordNotFound {
				return nil, nil
			}
			return nil, errors.ErrDatabaseError.WithError(latestErr)
		}
		contest, _ := s.contestRepo.GetByID(ctx, contestID)
		latestRound.Contest = contest
		return s.buildRoundResponse(latestRound), nil
	}
	contest, _ := s.contestRepo.GetByID(ctx, contestID)
	round.Contest = contest
	return s.buildRoundResponse(round), nil
}

func (s *AgentBattleService) ListRounds(ctx context.Context, contestID uint, req *request.ListAgentBattleRoundsRequest) ([]response.AgentBattleRoundResponse, int64, error) {
	if err := s.SyncContestRoundLifecycle(ctx, contestID); err != nil {
		return nil, 0, err
	}
	rounds, total, err := s.roundRepo.List(ctx, contestID, req.GetPage(), req.GetPageSize())
	if err != nil {
		return nil, 0, errors.ErrDatabaseError.WithError(err)
	}
	contest, _ := s.contestRepo.GetByID(ctx, contestID)

	list := make([]response.AgentBattleRoundResponse, len(rounds))
	for i, round := range rounds {
		round.Contest = contest
		list[i] = *s.buildRoundResponse(&round)
	}
	return list, total, nil
}

func (s *AgentBattleService) DeployContract(ctx context.Context, userID uint, req *request.DeployAgentContractRequest) (*response.AgentContractResponse, error) {
	if err := s.SyncContestRoundLifecycle(ctx, req.ContestID); err != nil {
		return nil, err
	}
	team, err := s.teamRepo.GetUserTeam(ctx, req.ContestID, userID)
	if err != nil {
		return nil, errors.ErrTeamNotFound
	}

	round, err := s.roundRepo.GetCurrentRound(ctx, req.ContestID)
	if err != nil {
		return nil, errors.ErrBattleNotStarted
	}

	existing, _ := s.contractRepo.GetByTeamAndRound(ctx, team.ID, round.ID)
	if existing != nil && existing.Status == model.ContractStatusDeployed {
		return nil, errors.ErrConflict.WithMessage("已部署合约")
	}

	if err := s.contractService.ValidateContractCode(req.SourceCode); err != nil {
		return nil, errors.ErrInvalidAgentCode.WithMessage(err.Error())
	}

	contract := &model.AgentContract{
		ContestID:   req.ContestID,
		TeamID:      team.ID,
		SubmitterID: userID,
		SourceCode:  req.SourceCode,
		Status:      model.ContractStatusPending,
	}
	if err := s.contractRepo.Create(ctx, contract); err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	safego.GoWithTimeout(fmt.Sprintf("deployContract:%d", contract.ID), 5*time.Minute, func(ctx context.Context) {
		s.contractService.DeployContractAsync(ctx, contract.ID, round.ChainRPCURL)
	})

	return s.buildContractResponse(contract, team.Name), nil
}

func (s *AgentBattleService) GetContractByTeam(ctx context.Context, contestID, teamID uint) (*response.AgentContractResponse, error) {
	if err := s.SyncContestRoundLifecycle(ctx, contestID); err != nil {
		return nil, err
	}
	round, err := s.roundRepo.GetCurrentRound(ctx, contestID)
	if err != nil {
		contract, latestErr := s.contractRepo.GetLatestByContestAndTeam(ctx, contestID, teamID)
		if latestErr != nil {
			return nil, errors.ErrNotFound
		}
		team, _ := s.teamRepo.GetByID(ctx, teamID)
		teamName := ""
		if team != nil {
			teamName = team.Name
		}
		return s.buildContractResponse(contract, teamName), nil
	}

	contract, err := s.contractRepo.GetByTeamAndRound(ctx, teamID, round.ID)
	if err != nil {
		contract, latestErr := s.contractRepo.GetLatestByContestAndTeam(ctx, contestID, teamID)
		if latestErr != nil {
			return nil, errors.ErrNotFound
		}
		team, _ := s.teamRepo.GetByID(ctx, teamID)
		teamName := ""
		if team != nil {
			teamName = team.Name
		}
		return s.buildContractResponse(contract, teamName), nil
	}

	team, _ := s.teamRepo.GetByID(ctx, teamID)
	teamName := ""
	if team != nil {
		teamName = team.Name
	}
	return s.buildContractResponse(contract, teamName), nil
}

func (s *AgentBattleService) GetScoreboard(ctx context.Context, contestID uint, roundID *uint) ([]response.AgentBattleScoreResponse, error) {
	if err := s.SyncContestRoundLifecycle(ctx, contestID); err != nil {
		return nil, err
	}
	var targetRoundID uint
	if roundID != nil {
		targetRoundID = *roundID
	} else {
		round, err := s.roundRepo.GetCurrentRound(ctx, contestID)
		if err != nil {
			round, err = s.roundRepo.GetLatestRound(ctx, contestID)
			if err != nil {
				return nil, errors.ErrBattleNotStarted
			}
		}
		targetRoundID = round.ID
	}

	scores, err := s.scoreRepo.GetScoreboard(ctx, targetRoundID)
	if err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	list := make([]response.AgentBattleScoreResponse, len(scores))
	for i, score := range scores {
		list[i] = response.AgentBattleScoreResponse{
			Rank:         i + 1,
			TeamID:       score.TeamID,
			Score:        score.Score,
			TokenBalance: score.TokenBalance,
			SuccessCount: score.SuccessCount,
			FailCount:    score.FailCount,
			ResourceHeld: battleResourceHeldFromScore(&score),
		}
		if score.Team != nil {
			list[i].TeamName = score.Team.Name
		}
	}
	return list, nil
}

func (s *AgentBattleService) ListEvents(ctx context.Context, roundID uint, req *request.ListAgentBattleEventsRequest) ([]model.AgentBattleEvent, int64, error) {
	return s.eventRepo.List(ctx, roundID, req.EventType, req.GetPage(), req.GetPageSize())
}

func (s *AgentBattleService) buildRoundResponse(round *model.AgentBattleRound) *response.AgentBattleRoundResponse {
	upgradeWindowEnd := s.getRoundUpgradeWindowEnd(round)

	return &response.AgentBattleRoundResponse{
		ID:               round.ID,
		ContestID:        round.ContestID,
		RoundNumber:      round.RoundNumber,
		Status:           round.Status,
		Phase:            s.getRoundPhase(round),
		StartTime:        round.StartTime,
		EndTime:          round.EndTime,
		UpgradeWindowEnd: upgradeWindowEnd,
		ChainRPCURL:      round.ChainRPCURL,
		BlockHeight:      round.BlockHeight,
		TxCount:          round.TxCount,
		CreatedAt:        round.CreatedAt,
	}
}

func (s *AgentBattleService) buildContractResponse(contract *model.AgentContract, teamName string) *response.AgentContractResponse {
	return &response.AgentContractResponse{
		ID:              contract.ID,
		ContestID:       contract.ContestID,
		TeamID:          contract.TeamID,
		TeamName:        teamName,
		ContractAddress: contract.ContractAddress,
		Status:          contract.Status,
		Version:         contract.Version,
		DeployedAt:      contract.DeployedAt,
		CreatedAt:       contract.CreatedAt,
	}
}

func (s *AgentBattleService) GetSpectateData(ctx context.Context, contestID uint) (*response.SpectateDataResponse, error) {
	if err := s.SyncContestRoundLifecycle(ctx, contestID); err != nil {
		return nil, err
	}
	contest, err := s.contestRepo.GetByID(ctx, contestID)
	if err != nil {
		return nil, errors.ErrContestNotFound
	}

	currentRound, _ := s.roundRepo.GetCurrentRound(ctx, contestID)
	if currentRound == nil {
		currentRound, _ = s.roundRepo.GetLatestRound(ctx, contestID)
	}

	teams := make([]response.SpectateTeamInfo, 0)
	if currentRound != nil {
		scores, _ := s.scoreRepo.GetScoreboard(ctx, currentRound.ID)
		for _, score := range scores {
			team, _ := s.teamRepo.GetByID(ctx, score.TeamID)
			if team != nil {
				teams = append(teams, response.SpectateTeamInfo{
					TeamID:       team.ID,
					TeamName:     team.Name,
					Score:        score.Score,
					ResourceHeld: battleResourceHeldFromScore(&score),
					IsAlive:      score.Score > 0,
				})
			}
		}
	}

	recentEvents := make([]response.SpectateEventInfo, 0)
	if currentRound != nil {
		eventList, _, _ := s.eventRepo.List(ctx, currentRound.ID, "", 1, 50)
		for _, event := range eventList {
			recentEvents = append(recentEvents, s.buildSpectateEventInfo(ctx, currentRound.RoundNumber, event))
		}
	}

	var currentRoundNum int
	var roundStatus string
	var roundPhase string
	var roundEndTime *time.Time
	if currentRound != nil {
		currentRound.Contest = contest
		currentRoundNum = currentRound.RoundNumber
		roundStatus = currentRound.Status
		roundPhase = s.getRoundPhase(currentRound)
		endTime := currentRound.EndTime
		roundEndTime = &endTime
	}

	return &response.SpectateDataResponse{
		ContestName:  contest.Title,
		CurrentRound: currentRoundNum,
		RoundStatus:  roundStatus,
		RoundPhase:   roundPhase,
		RoundEndTime: roundEndTime,
		Teams:        teams,
		RecentEvents: recentEvents,
	}, nil
}

func (s *AgentBattleService) GetReplayData(ctx context.Context, _ uint, roundID uint, fromBlock, toBlock *uint64) (*response.ReplayDataResponse, error) {
	round, err := s.roundRepo.GetByID(ctx, roundID)
	if err != nil {
		return nil, errors.ErrRoundNotFound
	}

	events, _, _ := s.eventRepo.List(ctx, roundID, "", 1, 10000)
	blockEvents := make(map[int64][]model.AgentBattleEvent)
	blocks := make([]int64, 0)
	for _, event := range events {
		if fromBlock != nil && uint64(event.BlockNumber) < *fromBlock {
			continue
		}
		if toBlock != nil && uint64(event.BlockNumber) > *toBlock {
			continue
		}
		if _, exists := blockEvents[event.BlockNumber]; !exists {
			blocks = append(blocks, event.BlockNumber)
		}
		blockEvents[event.BlockNumber] = append(blockEvents[event.BlockNumber], event)
	}
	sort.Slice(blocks, func(i, j int) bool { return blocks[i] < blocks[j] })

	scores, _ := s.scoreRepo.GetScoreboard(ctx, roundID)
	teamStateMap := make(map[uint]response.ReplayTeamState)
	for _, score := range scores {
		teamStateMap[score.TeamID] = response.ReplayTeamState{
			TeamID:   score.TeamID,
			Score:    0,
			Resource: 0,
		}
	}

	snapshots := make([]response.ReplaySnapshot, 0, len(blocks))
	for _, block := range blocks {
		evts := blockEvents[block]
		sort.SliceStable(evts, func(i, j int) bool {
			return evts[i].CreatedAt.Before(evts[j].CreatedAt)
		})

		for _, evt := range evts {
			if evt.TeamID == nil {
				continue
			}
			state := teamStateMap[*evt.TeamID]
			state.Score += int(readEventInt64(evt.EventData, "score_delta", "points", "point_delta"))
			state.Resource += int(readEventInt64(evt.EventData, "resource_delta", "resource_change", "resource"))
			teamStateMap[*evt.TeamID] = state
		}

		teamStates := make([]response.ReplayTeamState, 0, len(teamStateMap))
		for _, state := range teamStateMap {
			teamStates = append(teamStates, state)
		}
		sort.Slice(teamStates, func(i, j int) bool { return teamStates[i].TeamID < teamStates[j].TeamID })

		eventInfos := make([]response.ReplayEventInfo, 0, len(evts))
		for _, evt := range evts {
			eventInfos = append(eventInfos, s.buildReplayEventInfo(ctx, round.RoundNumber, evt))
		}

		snapshots = append(snapshots, response.ReplaySnapshot{
			Block:  uint64(block),
			Teams:  teamStates,
			Events: eventInfos,
		})
	}

	return &response.ReplayDataResponse{
		RoundID:    roundID,
		StartBlock: firstReplayBlock(blocks, uint64(round.BlockHeight)),
		EndBlock:   lastReplayBlock(blocks, uint64(round.BlockHeight)),
		Snapshots:  snapshots,
	}, nil
}

func (s *AgentBattleService) GetBattleStatus(ctx context.Context, contestID uint, currentUserID uint) (*response.BattleStatusResponse, error) {
	if err := s.SyncContestRoundLifecycle(ctx, contestID); err != nil {
		return nil, err
	}
	contest, err := s.contestRepo.GetByID(ctx, contestID)
	if err != nil {
		return nil, errors.ErrContestNotFound
	}

	rounds, _, err := s.roundRepo.List(ctx, contestID, 1, 1000)
	if err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	currentRound, err := s.roundRepo.GetCurrentRound(ctx, contestID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	displayRound := currentRound
	if displayRound == nil {
		latestRound, latestErr := s.roundRepo.GetLatestRound(ctx, contestID)
		if latestErr == nil {
			displayRound = latestRound
		}
	}

	status := "waiting"
	if contest.CurrentStatus() == model.ContestStatusEnded {
		status = "completed"
	}
	if currentRound != nil {
		switch currentRound.Status {
		case model.RoundStatusRunning:
			status = "running"
		case model.RoundStatusPending:
			status = "waiting"
		case model.RoundStatusFinished:
			status = "completed"
		default:
			status = currentRound.Status
		}
	}

	var currentBlock uint64
	if currentRound != nil && currentRound.ChainRPCURL != "" {
		currentBlock, _ = s.chainRuntime.GetCurrentBlock(ctx, currentRound.ChainRPCURL)
	}

	teams := make([]response.BattleTeamStatus, 0)
	contracts, _ := s.contractRepo.ListByContest(ctx, contestID)
	latestContracts := make(map[uint]model.AgentContract)
	var myRank int
	var myScore int64
	var myTeamID uint
	var agentStatus *response.BattleAgentStatus

	if currentUserID > 0 {
		if team, teamErr := s.teamRepo.GetUserTeam(ctx, contestID, currentUserID); teamErr == nil && team != nil {
			myTeamID = team.ID
		}
	}

	for _, contract := range contracts {
		if contract.TeamID == 0 {
			continue
		}
		existing, exists := latestContracts[contract.TeamID]
		if !exists || contract.RoundID > existing.RoundID || (contract.RoundID == existing.RoundID && contract.Version > existing.Version) {
			latestContracts[contract.TeamID] = contract
		}
	}

	if myTeamID > 0 {
		latestContract, latestErr := s.contractRepo.GetLatestByContestAndTeam(ctx, contestID, myTeamID)
		if latestErr == nil && latestContract != nil {
			uploadedAt := latestContract.CreatedAt
			if latestContract.DeployedAt != nil {
				uploadedAt = *latestContract.DeployedAt
			}
			agentStatus = &response.BattleAgentStatus{
				Version:    fmt.Sprintf("v%d", latestContract.Version),
				UploadedAt: &uploadedAt,
				IsValid:    latestContract.Status != model.ContractStatusFailed && latestContract.SourceCode != "",
			}
		}
	}

	for _, contract := range latestContracts {
		if contract.TeamID == 0 {
			continue
		}
		team, _ := s.teamRepo.GetByID(ctx, contract.TeamID)
		if team == nil {
			continue
		}

		var score *model.AgentBattleScore
		if displayRound != nil {
			score, _ = s.scoreRepo.GetByTeamAndRound(ctx, contract.TeamID, displayRound.ID)
		}

		var totalScore int
		if score != nil {
			totalScore = score.Score
			if myTeamID > 0 && contract.TeamID == myTeamID {
				myScore = int64(score.Score)
				myRank = score.Rank
			}
		}

		resourceHeld := 0
		if score != nil {
			resourceHeld = battleResourceHeldFromScore(score)
		}

		teams = append(teams, response.BattleTeamStatus{
			TeamID:          team.ID,
			TeamName:        team.Name,
			ContractAddress: contract.ContractAddress,
			IsAlive:         contract.Status == model.ContractStatusDeployed,
			ResourceHeld:    resourceHeld,
			TotalScore:      int64(totalScore),
		})
	}

	recentEvents := make([]response.BattleEventInfo, 0)
	if displayRound != nil {
		events, _, _ := s.eventRepo.List(ctx, displayRound.ID, "", 1, 20)
		for _, event := range events {
			recentEvents = append(recentEvents, s.buildBattleEventInfo(ctx, displayRound.RoundNumber, event))
		}
	}

	currentRoundNumber := 0
	roundPhase := ""
	if displayRound != nil {
		displayRound.Contest = contest
		currentRoundNumber = displayRound.RoundNumber
		roundPhase = s.getRoundPhase(displayRound)
	}
	totalRounds := len(rounds)
	if contest.BattleOrchestration.Lifecycle.TotalRounds > totalRounds {
		totalRounds = contest.BattleOrchestration.Lifecycle.TotalRounds
	}

	return &response.BattleStatusResponse{
		Status:       status,
		CurrentBlock: currentBlock,
		CurrentRound: currentRoundNumber,
		TotalRounds:  totalRounds,
		RoundPhase:   roundPhase,
		MyRank:       myRank,
		MyScore:      myScore,
		AgentStatus:  agentStatus,
		Teams:        teams,
		RecentEvents: recentEvents,
	}, nil
}

func (s *AgentBattleService) buildBattleEventInfo(ctx context.Context, roundNumber int, event model.AgentBattleEvent) response.BattleEventInfo {
	actorTeam := s.resolveEventTeamName(ctx, event.TeamID, event.EventData)
	targetTeam := readEventString(event.EventData, "target_team_name", "target_team", "target")
	actionResult := readEventString(event.EventData, "action_result", "result", "status")
	scoreDelta := readEventInt64(event.EventData, "score_delta", "points", "point_delta")
	resourceDelta := readEventInt64(event.EventData, "resource_delta", "resource_change", "resource")

	return response.BattleEventInfo{
		Block:         uint64(event.BlockNumber),
		RoundNumber:   roundNumber,
		EventType:     event.EventType,
		ActorTeam:     actorTeam,
		TargetTeam:    targetTeam,
		ActionResult:  actionResult,
		ScoreDelta:    scoreDelta,
		ResourceDelta: resourceDelta,
		Description:   s.buildBattleEventDescription(event.EventType, actorTeam, targetTeam, actionResult, scoreDelta, resourceDelta),
	}
}

func (s *AgentBattleService) buildSpectateEventInfo(ctx context.Context, roundNumber int, event model.AgentBattleEvent) response.SpectateEventInfo {
	base := s.buildBattleEventInfo(ctx, roundNumber, event)
	return response.SpectateEventInfo{
		Block:         base.Block,
		Time:          event.CreatedAt,
		RoundNumber:   base.RoundNumber,
		EventType:     base.EventType,
		ActorTeam:     base.ActorTeam,
		TargetTeam:    base.TargetTeam,
		ActionResult:  base.ActionResult,
		ScoreDelta:    base.ScoreDelta,
		ResourceDelta: base.ResourceDelta,
		Description:   base.Description,
	}
}

func (s *AgentBattleService) buildReplayEventInfo(ctx context.Context, roundNumber int, event model.AgentBattleEvent) response.ReplayEventInfo {
	base := s.buildBattleEventInfo(ctx, roundNumber, event)
	return response.ReplayEventInfo{
		EventType:     base.EventType,
		ActorTeam:     base.ActorTeam,
		TargetTeam:    base.TargetTeam,
		ActionResult:  base.ActionResult,
		ScoreDelta:    base.ScoreDelta,
		ResourceDelta: base.ResourceDelta,
		Description:   base.Description,
	}
}

func (s *AgentBattleService) resolveEventTeamName(ctx context.Context, teamID *uint, eventData model.JSONMap) string {
	if eventData != nil {
		if teamName := readEventString(eventData, "actor_team_name", "team_name", "actor_team"); teamName != "" {
			return teamName
		}
	}
	if teamID == nil {
		return ""
	}
	team, err := s.teamRepo.GetByID(ctx, *teamID)
	if err != nil || team == nil {
		return ""
	}
	return team.Name
}

func (s *AgentBattleService) buildBattleEventDescription(eventType, actorTeam, targetTeam, actionResult string, scoreDelta, resourceDelta int64) string {
	segments := make([]string, 0, 5)
	if actorTeam != "" {
		segments = append(segments, actorTeam)
	}
	segments = append(segments, eventType)
	if targetTeam != "" {
		segments = append(segments, "-> "+targetTeam)
	}
	if actionResult != "" {
		segments = append(segments, "结果: "+actionResult)
	}
	if scoreDelta != 0 {
		segments = append(segments, fmt.Sprintf("分数变化 %d", scoreDelta))
	}
	if resourceDelta != 0 {
		segments = append(segments, fmt.Sprintf("资源变化 %d", resourceDelta))
	}
	if len(segments) == 1 {
		return segments[0]
	}
	return strings.Join(segments, " | ")
}

func readEventString(eventData model.JSONMap, keys ...string) string {
	if eventData == nil {
		return ""
	}
	for _, key := range keys {
		value, exists := eventData[key]
		if !exists || value == nil {
			continue
		}
		if text, ok := value.(string); ok {
			return text
		}
	}
	return ""
}

func readEventInt64(eventData model.JSONMap, keys ...string) int64 {
	if eventData == nil {
		return 0
	}
	for _, key := range keys {
		value, exists := eventData[key]
		if !exists || value == nil {
			continue
		}
		switch typed := value.(type) {
		case int:
			return int64(typed)
		case int32:
			return int64(typed)
		case int64:
			return typed
		case float64:
			return int64(typed)
		case float32:
			return int64(typed)
		}
	}
	return 0
}

func (s *AgentBattleService) getRoundDurationSeconds(contest *model.Contest) int {
	if contest == nil || contest.BattleOrchestration.Lifecycle.RoundDurationSeconds <= 0 {
		return 300
	}
	return contest.BattleOrchestration.Lifecycle.RoundDurationSeconds
}

func (s *AgentBattleService) getRoundUpgradeWindowSeconds(contest *model.Contest) int {
	if contest == nil || contest.BattleOrchestration.Lifecycle.UpgradeWindowSeconds <= 0 {
		return 120
	}
	return contest.BattleOrchestration.Lifecycle.UpgradeWindowSeconds
}

func (s *AgentBattleService) getRoundUpgradeWindowEnd(round *model.AgentBattleRound) *time.Time {
	if round == nil || round.StartTime.IsZero() {
		return nil
	}
	end := round.StartTime.Add(time.Duration(s.getRoundUpgradeWindowSeconds(round.Contest)) * time.Second)
	return &end
}

func (s *AgentBattleService) getRoundPhase(round *model.AgentBattleRound) string {
	if round == nil {
		return model.RoundPhasePending
	}
	if round.Status == model.RoundStatusFinished {
		return model.RoundPhaseFinished
	}
	if round.Status != model.RoundStatusRunning {
		return model.RoundPhasePending
	}
	if round.StartTime.IsZero() {
		return model.RoundPhasePending
	}

	now := time.Now()
	upgradeWindowEnd := round.StartTime.Add(time.Duration(s.getRoundUpgradeWindowSeconds(round.Contest)) * time.Second)
	lockEnd := upgradeWindowEnd.Add(10 * time.Second)

	switch {
	case now.Before(upgradeWindowEnd):
		return model.RoundPhaseUpgradeWindow
	case now.Before(lockEnd):
		return model.RoundPhaseLocked
	case !round.EndTime.IsZero() && now.After(round.EndTime):
		return model.RoundPhaseSettling
	default:
		return model.RoundPhaseExecuting
	}
}

func battleResourceHeldFromScore(score *model.AgentBattleScore) int {
	if score == nil {
		return 0
	}
	if score.Details != nil {
		switch value := score.Details["resource_held"].(type) {
		case int:
			return value
		case int32:
			return int(value)
		case int64:
			return int(value)
		case float64:
			return int(value)
		}
	}
	if score.TokenBalance == "" {
		return 0
	}
	balance, ok := new(big.Int).SetString(score.TokenBalance, 10)
	if !ok {
		return 0
	}
	held := normalizeTokenBalance(balance)
	if held < 0 {
		return 0
	}
	if held > int64(^uint(0)>>1) {
		return int(^uint(0) >> 1)
	}
	return int(held)
}

func firstReplayBlock(blocks []int64, fallback uint64) uint64 {
	if len(blocks) == 0 {
		return fallback
	}
	return uint64(blocks[0])
}

func lastReplayBlock(blocks []int64, fallback uint64) uint64 {
	if len(blocks) == 0 {
		return fallback
	}
	return uint64(blocks[len(blocks)-1])
}

func (s *AgentBattleService) GetBattleConfig(ctx context.Context, contestID uint) (*response.BattleConfigResponse, error) {
	if err := s.SyncContestRoundLifecycle(ctx, contestID); err != nil {
		return nil, err
	}
	contest, err := s.contestRepo.GetByID(ctx, contestID)
	if err != nil {
		return nil, errors.ErrContestNotFound
	}

	roundList, _, _ := s.roundRepo.List(ctx, contestID, 1, 100)
	roundInfos := make([]response.BattleRoundInfo, 0, len(roundList))
	for _, r := range roundList {
		r.Contest = contest
		startTime := r.StartTime
		endTime := r.EndTime
		roundInfos = append(roundInfos, response.BattleRoundInfo{
			RoundNumber: r.RoundNumber,
			StartTime:   &startTime,
			EndTime:     &endTime,
			Status:      s.getRoundPhase(&r),
		})
	}

	var currentRoundNum int
	var chainRPC string
	currentRound, _ := s.roundRepo.GetCurrentRound(ctx, contestID)
	if currentRound != nil {
		currentRoundNum = currentRound.RoundNumber
		chainRPC = currentRound.ChainRPCURL
	}

	return &response.BattleConfigResponse{
		ContestID:    contestID,
		ContestName:  contest.Title,
		ChainRPC:     chainRPC,
		Rounds:       roundInfos,
		CurrentRound: currentRoundNum,
	}, nil
}

func (s *AgentBattleService) GetTeamWorkspaceForUser(ctx context.Context, userID, contestID uint) (*response.TeamWorkspaceResponse, error) {
	team, err := s.teamRepo.GetUserTeam(ctx, contestID, userID)
	if err != nil {
		return nil, errors.ErrTeamNotFound.WithMessage("请先报名参赛")
	}
	return s.workspaceService.GetTeamWorkspace(ctx, contestID, team.ID)
}

func (s *AgentBattleService) UpgradeContract(ctx context.Context, userID uint, req *request.UpgradeAgentContractRequest) (*response.AgentContractResponse, error) {
	if err := s.SyncContestRoundLifecycle(ctx, req.ContestID); err != nil {
		return nil, err
	}
	team, err := s.teamRepo.GetUserTeam(ctx, req.ContestID, userID)
	if err != nil {
		return nil, errors.ErrTeamNotFound
	}

	currentRound, err := s.roundRepo.GetCurrentRound(ctx, req.ContestID)
	if err != nil {
		return nil, errors.ErrRoundNotFound
	}
	contest, _ := s.contestRepo.GetByID(ctx, req.ContestID)
	currentRound.Contest = contest
	if s.getRoundPhase(currentRound) != model.RoundPhaseUpgradeWindow {
		return nil, errors.ErrInvalidParams.WithMessage("current round is outside the upgrade window")
	}

	contract, err := s.contractRepo.GetByTeamAndRound(ctx, team.ID, currentRound.ID)
	if err != nil {
		return nil, errors.ErrContractNotFound
	}

	bytecode, err := s.contractService.CompileContract(req.NewImplementation)
	if err != nil {
		return nil, errors.ErrContractCompileFailed.WithError(err)
	}

	safego.GoWithTimeout(fmt.Sprintf("deployUpgrade:%d", contract.ID), 5*time.Minute, func(ctx context.Context) {
		s.contractService.DeployUpgradeAsync(ctx, contract, currentRound, bytecode)
	})

	contract.Status = "upgrading"
	s.contractRepo.Update(ctx, contract)

	return &response.AgentContractResponse{
		ID:              contract.ID,
		TeamID:          contract.TeamID,
		ContractAddress: contract.ContractAddress,
		Status:          contract.Status,
	}, nil
}

func (s *AgentBattleService) GetFinalRank(ctx context.Context, contestID uint) ([]response.FinalRankResponse, error) {
	rounds, _, _ := s.roundRepo.List(ctx, contestID, 1, 100)

	teamScores := make(map[uint]int64)
	teamNames := make(map[uint]string)
	for _, round := range rounds {
		scores, _ := s.scoreRepo.GetScoreboard(ctx, round.ID)
		for _, score := range scores {
			teamScores[score.TeamID] += int64(score.Score)
			if _, ok := teamNames[score.TeamID]; !ok {
				team, _ := s.teamRepo.GetByID(ctx, score.TeamID)
				if team != nil {
					teamNames[score.TeamID] = team.Name
				}
			}
		}
	}

	type teamRank struct {
		teamID uint
		score  int64
	}

	rankList := make([]teamRank, 0, len(teamScores))
	for teamID, score := range teamScores {
		rankList = append(rankList, teamRank{teamID: teamID, score: score})
	}

	for i := 0; i < len(rankList); i++ {
		for j := i + 1; j < len(rankList); j++ {
			if rankList[j].score > rankList[i].score {
				rankList[i], rankList[j] = rankList[j], rankList[i]
			}
		}
	}

	result := make([]response.FinalRankResponse, 0, len(rankList))
	for rank, tr := range rankList {
		result = append(result, response.FinalRankResponse{
			Rank:       rank + 1,
			TeamID:     tr.teamID,
			TeamName:   teamNames[tr.teamID],
			TotalScore: tr.score,
		})
	}
	return result, nil
}

func (s *AgentBattleService) CreateOrGetTeamWorkspace(ctx context.Context, contestID, teamID uint) (*response.TeamWorkspaceResponse, error) {
	return s.workspaceService.CreateOrGetTeamWorkspace(ctx, contestID, teamID)
}

func (s *AgentBattleService) StopTeamWorkspace(ctx context.Context, _ uint, teamID uint) error {
	return s.workspaceService.StopTeamWorkspace(ctx, teamID)
}

func (s *AgentBattleService) SyncContestRoundLifecycle(ctx context.Context, contestID uint) error {
	contest, err := s.contestRepo.GetByID(ctx, contestID)
	if err != nil {
		return errors.ErrContestNotFound
	}
	if contest.Type != "agent_battle" {
		return nil
	}

	rounds, _, err := s.roundRepo.List(ctx, contestID, 1, 1000)
	if err != nil {
		return errors.ErrDatabaseError.WithError(err)
	}
	plannedTotalRounds := contest.BattleOrchestration.Lifecycle.TotalRounds
	if plannedTotalRounds <= 0 {
		plannedTotalRounds = 1
	}
	if len(rounds) < plannedTotalRounds {
		roundDurationSeconds := s.getRoundDurationSeconds(contest)
		existingByRoundNumber := make(map[int]struct{}, len(rounds))
		for index := range rounds {
			existingByRoundNumber[rounds[index].RoundNumber] = struct{}{}
		}
		for roundNumber := 1; roundNumber <= plannedTotalRounds; roundNumber++ {
			if _, exists := existingByRoundNumber[roundNumber]; exists {
				continue
			}
			startTime := contest.StartTime.Add(time.Duration(roundNumber-1) * time.Duration(roundDurationSeconds) * time.Second)
			endTime := startTime.Add(time.Duration(roundDurationSeconds) * time.Second)
			round := &model.AgentBattleRound{
				ContestID:   contestID,
				RoundNumber: roundNumber,
				Status:      model.RoundStatusPending,
				StartTime:   startTime,
				EndTime:     endTime,
			}
			if createErr := s.roundRepo.Create(ctx, round); createErr != nil {
				return errors.ErrDatabaseError.WithError(createErr)
			}
			round.Contest = contest
			rounds = append(rounds, *round)
		}
	}
	if len(rounds) == 0 {
		return nil
	}

	sort.Slice(rounds, func(i, j int) bool {
		return rounds[i].RoundNumber < rounds[j].RoundNumber
	})

	now := time.Now()
	var runningRound *model.AgentBattleRound
	var duePendingRound *model.AgentBattleRound
	for index := range rounds {
		rounds[index].Contest = contest
		switch rounds[index].Status {
		case model.RoundStatusRunning:
			if runningRound == nil {
				runningRound = &rounds[index]
			}
		case model.RoundStatusPending:
			if rounds[index].StartTime.IsZero() || now.Before(rounds[index].StartTime) {
				continue
			}
			if duePendingRound == nil {
				duePendingRound = &rounds[index]
			}
		}
	}

	if runningRound != nil && !runningRound.EndTime.IsZero() && !now.Before(runningRound.EndTime) {
		runningRound.Status = model.RoundStatusFinished
		if runningRound.ChainRPCURL != "" {
			if currentBlock, blockErr := s.chainRuntime.GetCurrentBlock(ctx, runningRound.ChainRPCURL); blockErr == nil {
				runningRound.BlockHeight = int64(currentBlock)
			}
		}
		if s.scoreService != nil {
			if err := s.scoreService.CalculateFinalScores(ctx, runningRound); err != nil {
				logger.Error("Failed to calculate final scores during lifecycle sync", zap.Uint("contestID", contestID), zap.Uint("roundID", runningRound.ID), zap.Error(err))
			}
		}
		if runningRound.ChainRPCURL != "" {
			if err := s.chainRuntime.StopSharedChain(ctx, runningRound); err != nil {
				logger.Error("Failed to stop battle chain during lifecycle sync", zap.Uint("contestID", contestID), zap.Uint("roundID", runningRound.ID), zap.Error(err))
			}
		}
		if err := s.roundRepo.Update(ctx, runningRound); err != nil {
			return errors.ErrDatabaseError.WithError(err)
		}
		runningRound = nil
	}

	if runningRound == nil && duePendingRound != nil {
		duePendingRound.Status = model.RoundStatusRunning
		duePendingRound.StartTime = now
		duePendingRound.EndTime = now.Add(time.Duration(s.getRoundDurationSeconds(contest)) * time.Second)

		chainRPCURL, err := s.chainRuntime.StartSharedChain(ctx, duePendingRound)
		if err != nil {
			logger.Error("Failed to start battle chain during lifecycle sync", zap.Uint("contestID", contestID), zap.Uint("roundID", duePendingRound.ID), zap.Error(err))
			duePendingRound.Status = model.RoundStatusPending
			duePendingRound.StartTime = time.Time{}
			duePendingRound.EndTime = time.Time{}
			_ = s.roundRepo.Update(ctx, duePendingRound)
			return nil
		}
		duePendingRound.ChainRPCURL = chainRPCURL

		if err := s.roundRepo.Update(ctx, duePendingRound); err != nil {
			return errors.ErrDatabaseError.WithError(err)
		}
	}

	return nil
}
