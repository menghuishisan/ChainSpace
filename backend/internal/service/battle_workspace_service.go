package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chainspace/backend/internal/dto/response"
	"github.com/chainspace/backend/internal/model"
	"github.com/chainspace/backend/internal/pkg/k8s"
	"github.com/chainspace/backend/internal/repository"
	"github.com/chainspace/backend/pkg/errors"
)

// BattleWorkspaceService 负责队伍工作区生命周期。
type BattleWorkspaceService struct {
	teamRepo            *repository.TeamRepository
	contestRepo         *repository.ContestRepository
	roundRepo           *repository.AgentBattleRoundRepository
	k8sClient           *k8s.Client
	provisioner         *RuntimeProvisionService
	challengeEnvService *ChallengeEnvService
}

func NewBattleWorkspaceService(
	teamRepo *repository.TeamRepository,
	contestRepo *repository.ContestRepository,
	roundRepo *repository.AgentBattleRoundRepository,
	k8sClient *k8s.Client,
	challengeEnvService *ChallengeEnvService,
) *BattleWorkspaceService {
	return &BattleWorkspaceService{
		teamRepo:            teamRepo,
		contestRepo:         contestRepo,
		roundRepo:           roundRepo,
		k8sClient:           k8sClient,
		provisioner:         NewRuntimeProvisionService(k8sClient),
		challengeEnvService: challengeEnvService,
	}
}

func (s *BattleWorkspaceService) GetTeamWorkspace(ctx context.Context, contestID, teamID uint) (*response.TeamWorkspaceResponse, error) {
	team, err := s.teamRepo.GetByID(ctx, teamID)
	if err != nil {
		return nil, errors.ErrTeamNotFound
	}
	wsConfig, _ := s.getBattleWorkspaceConfig(ctx, contestID)
	tools := buildBattleWorkspaceTools(team.WorkspacePodName, wsConfig)

	return &response.TeamWorkspaceResponse{
		TeamID:      team.ID,
		TeamName:    team.Name,
		EnvID:       team.WorkspacePodName,
		PodName:     team.WorkspacePodName,
		Status:      team.WorkspaceStatus,
		AccessURL:   s.getWorkspaceAccessURL(team.WorkspacePodName),
		ChainRPCURL: s.getSharedChainRPC(contestID),
		Tools:       tools,
	}, nil
}

func (s *BattleWorkspaceService) CreateOrGetTeamWorkspace(ctx context.Context, contestID, teamID uint) (*response.TeamWorkspaceResponse, error) {
	if s.k8sClient == nil || s.challengeEnvService == nil {
		return nil, errors.ErrInternal.WithMessage("battle workspace runtime is not initialized")
	}

	team, err := s.teamRepo.GetByID(ctx, teamID)
	if err != nil {
		return nil, errors.ErrTeamNotFound
	}

	if team.WorkspaceStatus == model.TeamWorkspaceStatusRunning && team.WorkspacePodName != "" {
		wsConfig, _ := s.getBattleWorkspaceConfig(ctx, contestID)
		return &response.TeamWorkspaceResponse{
			TeamID:      team.ID,
			TeamName:    team.Name,
			EnvID:       team.WorkspacePodName,
			PodName:     team.WorkspacePodName,
			Status:      team.WorkspaceStatus,
			AccessURL:   s.getWorkspaceAccessURL(team.WorkspacePodName),
			ChainRPCURL: s.getSharedChainRPC(contestID),
			Tools:       buildBattleWorkspaceTools(team.WorkspacePodName, wsConfig),
		}, nil
	}

	if team.WorkspacePodName != "" {
		_ = s.provisioner.StopInstance(ctx, team.WorkspacePodName)
	}

	wsConfig, err := s.getBattleWorkspaceConfig(ctx, contestID)
	if err != nil {
		return nil, err
	}

	podName := fmt.Sprintf("ws-battle-%d-%d", contestID, teamID)
	podInput := ChallengePodInput{
		EnvID:       podName,
		UserID:      team.LeaderID,
		TeamID:      team.ID,
		SchoolID:    0,
		ContestID:   contestID,
		ChallengeID: 0,
	}
	podCfg := s.challengeEnvService.BuildTeamWorkspacePodConfig(*wsConfig, podInput, 30*time.Minute)
	podCfg.EnvVars["CHAIN_RPC_URL"] = s.getSharedChainRPC(contestID)

	if err := s.provisioner.StartInstance(ctx, podCfg, 2*time.Minute); err != nil {
		return nil, errors.ErrInternal.WithMessage(fmt.Sprintf("创建工作区失败: %v", err))
	}

	team.WorkspacePodName = podName
	team.WorkspaceStatus = model.TeamWorkspaceStatusRunning
	if err := s.teamRepo.Update(ctx, team); err != nil {
		_ = s.provisioner.StopInstance(ctx, podName)
		return nil, errors.ErrDatabaseError.WithMessage("更新队伍工作区状态失败")
	}

	return &response.TeamWorkspaceResponse{
		TeamID:      team.ID,
		TeamName:    team.Name,
		EnvID:       podName,
		PodName:     podName,
		Status:      model.TeamWorkspaceStatusRunning,
		AccessURL:   s.getWorkspaceAccessURL(podName),
		ChainRPCURL: s.getSharedChainRPC(contestID),
		Tools:       buildBattleWorkspaceTools(podName, wsConfig),
	}, nil
}

func (s *BattleWorkspaceService) StopTeamWorkspace(ctx context.Context, teamID uint) error {
	if s.k8sClient == nil {
		return errors.ErrInternal.WithMessage("battle workspace runtime is not initialized")
	}

	team, err := s.teamRepo.GetByID(ctx, teamID)
	if err != nil {
		return errors.ErrTeamNotFound
	}

	if team.WorkspacePodName != "" {
		_ = s.provisioner.StopInstance(ctx, team.WorkspacePodName)
		team.WorkspacePodName = ""
		team.WorkspaceStatus = model.TeamWorkspaceStatusStopped
		return s.teamRepo.Update(ctx, team)
	}

	return nil
}

func (s *BattleWorkspaceService) CleanupAllTeamWorkspaces(ctx context.Context, contestID uint) error {
	if s.k8sClient == nil {
		return errors.ErrInternal.WithMessage("battle workspace runtime is not initialized")
	}

	page := 1
	pageSize := 100
	for {
		teams, total, err := s.teamRepo.List(ctx, contestID, "", "", page, pageSize)
		if err != nil {
			return err
		}
		for i := range teams {
			if teams[i].WorkspacePodName != "" {
				_ = s.provisioner.StopInstance(ctx, teams[i].WorkspacePodName)
				teams[i].WorkspacePodName = ""
				teams[i].WorkspaceStatus = model.TeamWorkspaceStatusStopped
				s.teamRepo.Update(ctx, &teams[i])
			}
		}
		if int(total) <= page*pageSize {
			break
		}
		page++
	}

	return nil
}

func (s *BattleWorkspaceService) getBattleWorkspaceConfig(ctx context.Context, contestID uint) (*model.BattleTeamWorkspaceSpec, error) {
	contest, err := s.contestRepo.GetByID(ctx, contestID)
	if err != nil {
		return &model.BattleTeamWorkspaceSpec{
			Image: defaultBattleWorkspaceImage(),
			InteractionTools: []string{
				"ide", "terminal", "files", "logs",
			},
			Resources: map[string]string{
				"cpu":     "1000m",
				"memory":  "2Gi",
				"storage": "10Gi",
			},
		}, nil
	}

	wsConfig := contest.BattleOrchestration.TeamWorkspace
	ws := &model.BattleTeamWorkspaceSpec{
		Image:            wsConfig.Image,
		InteractionTools: append([]string{}, wsConfig.InteractionTools...),
		Resources:        wsConfig.Resources,
	}
	if ws.Image == "" {
		ws.Image = defaultBattleWorkspaceImage()
	}
	if len(ws.InteractionTools) == 0 {
		ws.InteractionTools = []string{"ide", "terminal", "files", "logs"}
	}
	if ws.Resources == nil {
		ws.Resources = map[string]string{
			"cpu":     "1000m",
			"memory":  "2Gi",
			"storage": "10Gi",
		}
	}

	return ws, nil
}

func (s *BattleWorkspaceService) getWorkspaceAccessURL(podName string) string {
	if podName == "" {
		return ""
	}
	return fmt.Sprintf("/api/v1/battle-workspace/%s/ide", podName)
}

func (s *BattleWorkspaceService) getSharedChainRPC(contestID uint) string {
	if s.contestRepo != nil {
		if contest, err := s.contestRepo.GetByID(context.Background(), contestID); err == nil && contest != nil {
			if rpcURL := strings.TrimSpace(contest.BattleOrchestration.SharedChain.RPCURL); rpcURL != "" {
				return rpcURL
			}
		}
	}
	ctx := context.Background()
	round, err := s.roundRepo.GetCurrentRound(ctx, contestID)
	if err != nil {
		return ""
	}
	return round.ChainRPCURL
}

func buildBattleWorkspaceTools(envID string, wsConfig *model.BattleTeamWorkspaceSpec) []response.RuntimeToolResponse {
	if envID == "" {
		return nil
	}

	toolKeys := []string{"ide", "terminal", "files", "logs"}
	if wsConfig != nil && len(wsConfig.InteractionTools) > 0 {
		toolKeys = wsConfig.InteractionTools
	}
	workspaceSupported := map[string]struct{}{
		"ide":      {},
		"terminal": {},
		"files":    {},
		"logs":     {},
	}

	tools := make([]response.RuntimeToolResponse, 0, len(toolKeys))
	seen := map[string]struct{}{}
	for _, toolKey := range toolKeys {
		if _, ok := workspaceSupported[toolKey]; !ok {
			continue
		}
		if _, exists := seen[toolKey]; exists {
			continue
		}
		seen[toolKey] = struct{}{}
		routeSegment, ok := battleWorkspaceRouteSegment(toolKey)
		if !ok {
			continue
		}
		tools = append(tools, response.RuntimeToolResponse{
			Key:           toolKey,
			Label:         challengeToolLabel(toolKey),
			Kind:          toolKey,
			Target:        "workspace",
			InstanceKey:   "workspace",
			StudentFacing: true,
			Port:          int32(challengeRuntimeDefaultPortForTool(toolKey)),
			Route:         fmt.Sprintf("/api/v1/battle-workspace/%s/%s", envID, routeSegment),
		})
	}
	return tools
}

func battleWorkspaceRouteSegment(toolKey string) (string, bool) {
	switch toolKey {
	case "ide", "terminal", "files", "logs":
		return toolKey, true
	default:
		return "", false
	}
}
