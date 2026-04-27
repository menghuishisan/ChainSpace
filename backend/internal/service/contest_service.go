package service

import (
	"context"
	"mime/multipart"
	"strings"

	"github.com/chainspace/backend/internal/dto/request"
	"github.com/chainspace/backend/internal/dto/response"
	"github.com/chainspace/backend/internal/model"
	"github.com/chainspace/backend/internal/repository"
	"github.com/chainspace/backend/pkg/errors"
)

// ContestService handles contest-related business logic.
type ContestService struct {
	contestRepo           *repository.ContestRepository
	challengeRepo         *repository.ChallengeRepository
	contestChallengeRepo  *repository.ContestChallengeRepository
	teamRepo              *repository.TeamRepository
	teamMemberRepo        *repository.TeamMemberRepository
	contestSubmissionRepo *repository.ContestSubmissionRepository
	contestScoreRepo      *repository.ContestScoreRepository
	userRepo              *repository.UserRepository
	adminService          *ContestAdminService
	runtimeFacade         *ContestRuntimeFacade
	scoringFacade         *ContestScoringFacade
	uploadService         *UploadService
}

func NewContestService(
	contestRepo *repository.ContestRepository,
	challengeRepo *repository.ChallengeRepository,
	contestChallengeRepo *repository.ContestChallengeRepository,
	teamRepo *repository.TeamRepository,
	teamMemberRepo *repository.TeamMemberRepository,
	contestSubmissionRepo *repository.ContestSubmissionRepository,
	contestScoreRepo *repository.ContestScoreRepository,
	userRepo *repository.UserRepository,
	imageRepo *repository.DockerImageRepository,
	runtimeService *ContestRuntimeService,
	scoreService *ContestScoreService,
	uploadService *UploadService,
) *ContestService {
	return &ContestService{
		contestRepo:           contestRepo,
		challengeRepo:         challengeRepo,
		contestChallengeRepo:  contestChallengeRepo,
		teamRepo:              teamRepo,
		teamMemberRepo:        teamMemberRepo,
		contestSubmissionRepo: contestSubmissionRepo,
		contestScoreRepo:      contestScoreRepo,
		userRepo:              userRepo,
		adminService: NewContestAdminService(
			contestRepo,
			contestChallengeRepo,
			teamRepo,
			teamMemberRepo,
			imageRepo,
		),
		runtimeFacade: NewContestRuntimeFacade(
			contestRepo,
			challengeRepo,
			contestChallengeRepo,
			teamRepo,
			runtimeService,
		),
		scoringFacade: NewContestScoringFacade(scoreService),
		uploadService: uploadService,
	}
}

func (s *ContestService) CreateContest(ctx context.Context, creatorID uint, schoolID *uint, req *request.CreateContestRequest) (*response.ContestResponse, error) {
	if s.adminService == nil {
		return nil, errors.ErrInternal.WithMessage("contest admin service not initialized")
	}
	return s.adminService.CreateContest(ctx, creatorID, schoolID, req)
}

func (s *ContestService) UpdateContest(ctx context.Context, contestID, userID, schoolID uint, role string, req *request.UpdateContestRequest) (*response.ContestResponse, error) {
	if s.adminService == nil {
		return nil, errors.ErrInternal.WithMessage("contest admin service not initialized")
	}
	return s.adminService.UpdateContest(ctx, contestID, userID, schoolID, role, req)
}

func (s *ContestService) transitionContestStatus(ctx context.Context, contest *model.Contest, targetStatus string) error {
	if s.adminService == nil {
		return errors.ErrInternal.WithMessage("contest admin service not initialized")
	}
	return s.adminService.transitionContestStatus(ctx, contest, targetStatus)
}

func (s *ContestService) GetContest(ctx context.Context, contestID, userID, schoolID uint, role string) (*response.ContestResponse, error) {
	if s.adminService == nil {
		return nil, errors.ErrInternal.WithMessage("contest admin service not initialized")
	}
	return s.adminService.GetContest(ctx, contestID, userID, schoolID, role)
}

func (s *ContestService) ListContests(ctx context.Context, schoolID uint, userID uint, role string, req *request.ListContestsRequest) ([]response.ContestResponse, int64, error) {
	if s.adminService == nil {
		return nil, 0, errors.ErrInternal.WithMessage("contest admin service not initialized")
	}
	return s.adminService.ListContests(ctx, schoolID, userID, role, req)
}

func normalizeBattleOrchestration(orch model.BattleOrchestration, contestType string) model.BattleOrchestration {
	if contestType != model.ContestTypeAgentBattle {
		return orch
	}
	if orch.SharedChain.Image == "" {
		orch.SharedChain.Image = defaultBattleSharedChainImage(orch.SharedChain.ChainType)
	}
	if orch.SharedChain.ChainType == "" {
		orch.SharedChain.ChainType = "anvil"
	}
	if orch.SharedChain.NetworkID == 0 {
		orch.SharedChain.NetworkID = 31337
	}
	if orch.SharedChain.BlockTime == 0 {
		orch.SharedChain.BlockTime = 1
	}
	if orch.Judge.StrategyInterface == "" {
		orch.Judge.StrategyInterface = "strategy_agent_v1"
	}
	if orch.Judge.ResourceModel == "" {
		orch.Judge.ResourceModel = "shared_resource_control"
	}
	if orch.Judge.ScoringModel == "" {
		orch.Judge.ScoringModel = "resource_attack_defense_survival"
	}
	if orch.Judge.ScoreWeights == nil {
		orch.Judge.ScoreWeights = map[string]int{"resource": 35, "attack": 30, "defense": 20, "survival": 15}
	}
	if len(orch.Judge.AllowedActions) == 0 {
		orch.Judge.AllowedActions = []string{"gather", "attack", "defend", "fortify", "recover", "scout"}
	}
	if orch.TeamWorkspace.Image == "" {
		orch.TeamWorkspace.Image = defaultBattleWorkspaceImage()
	}
	if orch.TeamWorkspace.DisplayName == "" {
		orch.TeamWorkspace.DisplayName = "Battle Strategy Workspace"
	}
	orch.TeamWorkspace.InteractionTools = normalizeBattleInteractionTools(orch.TeamWorkspace.InteractionTools)
	if len(orch.TeamWorkspace.InteractionTools) == 0 {
		orch.TeamWorkspace.InteractionTools = []string{"ide", "terminal", "files", "logs", "api_debug"}
	}
	if orch.TeamWorkspace.Resources == nil {
		orch.TeamWorkspace.Resources = map[string]string{"cpu": "1000m", "memory": "2Gi", "storage": "10Gi"}
	}
	if !orch.Spectate.EnableMonitor {
		orch.Spectate.EnableMonitor = true
	}
	if !orch.Spectate.EnableReplay {
		orch.Spectate.EnableReplay = true
	}
	if orch.Lifecycle.RoundDurationSeconds == 0 {
		orch.Lifecycle.RoundDurationSeconds = 300
	}
	if orch.Lifecycle.UpgradeWindowSeconds == 0 {
		orch.Lifecycle.UpgradeWindowSeconds = 120
	}
	if orch.Lifecycle.TotalRounds == 0 {
		orch.Lifecycle.TotalRounds = 5
	}
	if !orch.Lifecycle.AutoCleanup {
		orch.Lifecycle.AutoCleanup = true
	}
	return orch
}

func (s *ContestService) UploadAgentCode(ctx context.Context, contestID uint, userID uint, file *multipart.FileHeader) (string, error) {
	if s.runtimeFacade == nil {
		return "", errors.ErrInternal.WithMessage("contest runtime facade not initialized")
	}
	return s.runtimeFacade.UploadAgentCode(ctx, contestID, userID, file)
}

func (s *ContestService) StartChallengeEnv(ctx context.Context, userID uint, contestID uint, challengeID uint) (*response.ChallengeEnvResponse, error) {
	if s.runtimeFacade == nil {
		return nil, errors.ErrInternal.WithMessage("contest runtime facade not initialized")
	}
	return s.runtimeFacade.StartChallengeEnv(ctx, userID, contestID, challengeID)
}

func (s *ContestService) GetChallengeEnvStatus(ctx context.Context, userID uint, contestID uint, challengeID uint) (*response.ChallengeEnvResponse, error) {
	if s.runtimeFacade == nil {
		return nil, errors.ErrInternal.WithMessage("contest runtime facade not initialized")
	}
	return s.runtimeFacade.GetChallengeEnvStatus(ctx, userID, contestID, challengeID)
}

func (s *ContestService) StopChallengeEnv(ctx context.Context, userID uint, contestID uint, challengeID uint) error {
	if s.runtimeFacade == nil {
		return errors.ErrInternal.WithMessage("contest runtime facade not initialized")
	}
	return s.runtimeFacade.StopChallengeEnv(ctx, userID, contestID, challengeID)
}

func (s *ContestService) SubmitFlag(ctx context.Context, userID uint, schoolID uint, role string, contestID uint, req *request.SubmitFlagRequest, ip string) (*response.FlagSubmitResponse, error) {
	if s.scoringFacade == nil {
		return nil, errors.ErrInternal.WithMessage("contest scoring facade not initialized")
	}
	return s.scoringFacade.SubmitFlag(ctx, userID, schoolID, role, contestID, req, ip)
}

func (s *ContestService) GetScoreboard(ctx context.Context, contestID, currentUserID, schoolID uint, role string) (*response.ScoreboardResponse, error) {
	if s.scoringFacade == nil {
		return nil, errors.ErrInternal.WithMessage("contest scoring facade not initialized")
	}
	return s.scoringFacade.GetScoreboard(ctx, contestID, currentUserID, schoolID, role)
}

func normalizeBattleInteractionTools(values []string) []string {
	allowed := map[string]struct{}{
		"ide":           {},
		"terminal":      {},
		"files":         {},
		"logs":          {},
		"explorer":      {},
		"api_debug":     {},
		"visualization": {},
		"network":       {},
		"rpc":           {},
	}
	result := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, item := range values {
		key := strings.TrimSpace(item)
		if key == "" {
			continue
		}
		if _, ok := allowed[key]; !ok {
			continue
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, key)
	}
	return result
}
