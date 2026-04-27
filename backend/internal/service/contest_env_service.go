package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/chainspace/backend/internal/dto/response"
	"github.com/chainspace/backend/internal/model"
	"github.com/chainspace/backend/internal/pkg/k8s"
	"github.com/chainspace/backend/internal/pkg/logger"
	"github.com/chainspace/backend/internal/pkg/safego"
	"github.com/chainspace/backend/internal/repository"
	"github.com/chainspace/backend/pkg/errors"
	"go.uber.org/zap"
)

// ContestRuntimeService only executes contest runtime actions and environment orchestration.
type ContestRuntimeService struct {
	challengeEnvRepo     *repository.ChallengeEnvRepository
	challengeRepo        *repository.ChallengeRepository
	uploadService        *UploadService
	k8sClient            *k8s.Client
	provisionService     *RuntimeProvisionService
	challengeEnvService  *ChallengeEnvService
	workspaceSeedService *ChallengeWorkspaceSeedService
}

func NewContestRuntimeService(
	challengeEnvRepo *repository.ChallengeEnvRepository,
	challengeRepo *repository.ChallengeRepository,
	uploadService *UploadService,
	k8sClient *k8s.Client,
	challengeEnvService *ChallengeEnvService,
	workspaceSeedService *ChallengeWorkspaceSeedService,
) *ContestRuntimeService {
	return &ContestRuntimeService{
		challengeEnvRepo:     challengeEnvRepo,
		challengeRepo:        challengeRepo,
		uploadService:        uploadService,
		k8sClient:            k8sClient,
		provisionService:     NewRuntimeProvisionService(k8sClient),
		challengeEnvService:  challengeEnvService,
		workspaceSeedService: workspaceSeedService,
	}
}

func (s *ContestRuntimeService) UploadAgentCode(ctx context.Context, contestID uint, teamID uint, file *multipart.FileHeader) (string, error) {
	result, err := s.uploadService.UploadAgentCode(ctx, file, contestID, teamID)
	if err != nil {
		return "", errors.ErrFileUploadFailed.WithError(err)
	}

	return fmt.Sprintf("v%d-%s", time.Now().Unix(), result.Filename), nil
}

func (s *ContestRuntimeService) StartChallengeEnv(ctx context.Context, userID uint, contestID uint, challenge *model.Challenge) (*model.ChallengeEnv, error) {
	if s.k8sClient == nil || s.challengeEnvService == nil {
		return nil, errors.ErrInternal.WithMessage("contest runtime is not initialized")
	}
	if challenge == nil {
		return nil, errors.ErrInvalidParams.WithMessage("challenge is required")
	}

	challengeID := challenge.ID
	existing, _ := s.challengeEnvRepo.GetActiveByUserAndChallenge(ctx, userID, contestID, challengeID)
	if existing != nil {
		if existing.Status == model.ChallengeEnvStatusRunning && existing.ExpiresAt != nil && existing.ExpiresAt.After(time.Now()) {
			return existing, nil
		}
		if existing.Status == model.ChallengeEnvStatusPending || existing.Status == model.ChallengeEnvStatusCreating {
			return existing, nil
		}
		s.cleanupPersistedChallengeEnv(ctx, existing)
		existing.Status = model.ChallengeEnvStatusTerminated
		_ = s.challengeEnvRepo.Update(ctx, existing)
	}

	envID := fmt.Sprintf("jeo-%d-%d-%s", contestID, userID, createInviteCode())
	flag := generateDynamicFlag(challenge, userID)
	now := time.Now()
	ttlMinutes := challenge.ChallengeOrchestration.Lifecycle.TimeLimitMinutes
	if ttlMinutes <= 0 {
		ttlMinutes = 120
	}
	timeout := time.Duration(ttlMinutes) * time.Minute

	env := &model.ChallengeEnv{
		EnvID:       envID,
		ContestID:   contestID,
		ChallengeID: challengeID,
		UserID:      userID,
		Status:      model.ChallengeEnvStatusCreating,
		Flag:        flag,
		StartedAt:   &now,
	}

	if err := s.challengeEnvRepo.Create(ctx, env); err != nil {
		return nil, errors.ErrDatabaseError.WithMessage("challenge environment create failed")
	}

	safego.GoWithTimeout("startChallengeEnv:"+envID, 10*time.Minute, func(asyncCtx context.Context) {
		s.provisionChallengeEnv(asyncCtx, env, challenge, timeout)
	})

	return env, nil
}

func (s *ContestRuntimeService) GetChallengeEnvStatus(ctx context.Context, userID uint, contestID uint, challengeID uint) (*model.ChallengeEnv, error) {
	env, err := s.challengeEnvRepo.GetActiveByUserAndChallenge(ctx, userID, contestID, challengeID)
	if err != nil {
		return nil, nil
	}
	if env.ExpiresAt != nil && env.ExpiresAt.Before(time.Now()) {
		env.Status = model.ChallengeEnvStatusExpired
		_ = s.challengeEnvRepo.Update(ctx, env)
	}
	return env, nil
}

func (s *ContestRuntimeService) StopChallengeEnv(ctx context.Context, userID uint, contestID uint, challengeID uint) error {
	if s.k8sClient == nil {
		return errors.ErrInternal.WithMessage("contest runtime is not initialized")
	}

	env, err := s.challengeEnvRepo.GetActiveByUserAndChallenge(ctx, userID, contestID, challengeID)
	if err != nil {
		return errors.ErrInvalidParams.WithMessage("active challenge environment not found")
	}

	if env.PodName != "" {
		s.cleanupPersistedChallengeEnv(ctx, env)
	}

	env.Status = model.ChallengeEnvStatusTerminated
	return s.challengeEnvRepo.Update(ctx, env)
}

func (s *ContestRuntimeService) provisionChallengeEnv(ctx context.Context, env *model.ChallengeEnv, challenge *model.Challenge, timeout time.Duration) {
	if env == nil || challenge == nil {
		return
	}

	podInput := ChallengePodInput{
		EnvID:       env.EnvID,
		UserID:      env.UserID,
		SchoolID:    0,
		ContestID:   env.ContestID,
		ChallengeID: env.ChallengeID,
	}

	podConfig := s.challengeEnvService.BuildChallengePodConfig(challenge.ChallengeOrchestration, podInput, timeout)
	if err := s.provisionService.StartInstance(ctx, podConfig, 3*time.Minute); err != nil {
		s.markChallengeEnvFailed(ctx, env, fmt.Sprintf("create challenge pod failed: %v", err))
		return
	}

	servicePodConfigs := s.challengeEnvService.BuildChallengeServicePodConfigs(challenge.ChallengeOrchestration, podInput, timeout, s.k8sClient.Namespace())
	createdServicePods := make([]string, 0, len(servicePodConfigs))
	for _, servicePodConfig := range servicePodConfigs {
		if err := s.provisionService.StartInstance(ctx, servicePodConfig, 2*time.Minute); err != nil {
			s.cleanupChallengeServicePods(ctx, createdServicePods)
			_ = s.provisionService.StopInstance(ctx, env.EnvID)
			s.markChallengeEnvFailed(ctx, env, fmt.Sprintf("create challenge service pod failed: %v", err))
			return
		}
		createdServicePods = append(createdServicePods, servicePodConfig.EnvID)
	}

	forkAccessURL := ""
	forkPodName := ""
	if challenge.ChallengeOrchestration.Fork.Enabled {
		forkConfig := s.challengeEnvService.BuildAnvilForkPodConfig(challenge.ChallengeOrchestration.Fork, podInput, timeout)
		if err := s.provisionService.StartInstance(ctx, forkConfig, 2*time.Minute); err != nil {
			s.cleanupChallengeServicePods(ctx, createdServicePods)
			_ = s.provisionService.StopInstance(ctx, env.EnvID)
			s.markChallengeEnvFailed(ctx, env, fmt.Sprintf("create fork runtime failed: %v", err))
			return
		}
		forkPodName = fmt.Sprintf("fork-%s", env.EnvID)
		forkAccessURL = fmt.Sprintf("http://fork-%s.%s.svc.cluster.local:8545", env.EnvID, s.k8sClient.Namespace())
	}

	if s.workspaceSeedService != nil {
		if err := s.workspaceSeedService.SeedChallengeWorkspace(ctx, env.EnvID, challenge, forkAccessURL); err != nil {
			s.cleanupChallengeServicePods(ctx, createdServicePods)
			_ = s.provisionService.StopInstance(ctx, env.EnvID)
			if forkPodName != "" {
				_ = s.provisionService.StopInstance(ctx, forkPodName)
			}
			s.markChallengeEnvFailed(ctx, env, fmt.Sprintf("seed challenge workspace failed: %v", err))
			return
		}
	}

	latestEnv, err := s.challengeEnvRepo.GetByEnvID(ctx, env.EnvID)
	if err == nil && latestEnv != nil && latestEnv.Status == model.ChallengeEnvStatusTerminated {
		s.cleanupChallengeServicePods(ctx, createdServicePods)
		_ = s.provisionService.StopInstance(ctx, env.EnvID)
		if forkPodName != "" {
			_ = s.provisionService.StopInstance(ctx, forkPodName)
		}
		return
	}

	startedAt := time.Now()
	expiresAt := startedAt.Add(timeout)
	env.Status = model.ChallengeEnvStatusRunning
	env.PodName = env.EnvID
	env.ForkPodName = forkPodName
	env.AccessURL = buildChallengeProxyRoute(env.EnvID, "ide")
	env.StartedAt = &startedAt
	env.ExpiresAt = &expiresAt
	env.ErrorMessage = ""
	if err := s.challengeEnvRepo.Update(ctx, env); err != nil {
		logger.Error("update challenge env after ready failed",
			zap.String("env_id", env.EnvID),
			zap.Uint("contest_id", env.ContestID),
			zap.Uint("challenge_id", env.ChallengeID),
			zap.Error(err),
		)
	}
}

func (s *ContestRuntimeService) markChallengeEnvFailed(ctx context.Context, env *model.ChallengeEnv, errorMessage string) {
	if env == nil {
		return
	}

	env.Status = model.ChallengeEnvStatusFailed
	env.ErrorMessage = errorMessage
	env.ExpiresAt = nil
	if err := s.challengeEnvRepo.Update(ctx, env); err != nil {
		logger.Error("update challenge env failed state failed",
			zap.String("env_id", env.EnvID),
			zap.Uint("contest_id", env.ContestID),
			zap.Uint("challenge_id", env.ChallengeID),
			zap.String("error_message", errorMessage),
			zap.Error(err),
		)
	}
}

func (s *ContestRuntimeService) buildChallengeEnvResponse(env *model.ChallengeEnv, challenge *model.Challenge) *response.ChallengeEnvResponse {
	remaining := 0
	if env.ExpiresAt != nil {
		remaining = int(time.Until(*env.ExpiresAt).Seconds())
		if remaining < 0 {
			remaining = 0
		}
	}

	tools, serviceEntries := buildChallengeRuntimeView(env, challenge)

	return &response.ChallengeEnvResponse{
		ID:             env.ID,
		EnvID:          env.EnvID,
		ContestID:      env.ContestID,
		ChallengeID:    env.ChallengeID,
		Status:         env.Status,
		AccessURL:      env.AccessURL,
		Tools:          tools,
		ServiceEntries: serviceEntries,
		StartedAt:      env.StartedAt,
		ExpiresAt:      env.ExpiresAt,
		ErrorMessage:   env.ErrorMessage,
		Remaining:      remaining,
	}
}

func (s *ContestRuntimeService) cleanupChallengeServicePods(ctx context.Context, podNames []string) {
	for _, podName := range podNames {
		_ = s.provisionService.StopInstance(ctx, podName)
	}
}

func (s *ContestRuntimeService) cleanupPersistedChallengeEnv(ctx context.Context, env *model.ChallengeEnv) {
	if env == nil || s.k8sClient == nil {
		return
	}

	if env.PodName != "" {
		_ = s.provisionService.StopInstance(ctx, env.PodName)
	}
	if s.challengeRepo != nil {
		if challenge, err := s.challengeRepo.GetByID(ctx, env.ChallengeID); err == nil && challenge != nil {
			for _, bundle := range buildChallengeRuntimeBundles(challenge.ChallengeOrchestration) {
				for _, component := range bundle.Components {
					_ = s.provisionService.StopInstance(ctx, challengeBundleComponentEnvID(env.EnvID, component.RuntimeKey))
				}
			}
		}
	}
	if env.ForkPodName != "" {
		_ = s.provisionService.StopInstance(ctx, env.ForkPodName)
	}
}

func buildChallengeRuntimeView(env *model.ChallengeEnv, challenge *model.Challenge) ([]response.RuntimeToolResponse, []response.RuntimeServiceResponse) {
	kernel := compileChallengeRuntimeKernel(env, challenge)
	return mapKernelToolsToRuntimeTools(kernel.state), kernel.services
}

func buildChallengeProxyRoute(envID, segment string) string {
	return fmt.Sprintf("/api/v1/contest-envs/%s/proxy/%s", envID, segment)
}

func buildChallengeServiceProxyRoute(envID, serviceKey string) string {
	return fmt.Sprintf("/api/v1/contest-envs/%s/proxy/services/%s", envID, serviceKey)
}

func challengeRouteSegmentByTool(toolKey string) (string, bool) {
	switch toolKey {
	case "ide", "terminal", "rpc", "explorer", "files", "logs", "network":
		return toolKey, true
	case "visualization":
		return "visualization", true
	case "api_debug":
		return "api_debug", true
	default:
		return "", false
	}
}

func challengeRouteSegmentByExposure(exposed string) (string, bool) {
	switch exposed {
	case "workspace", "ide":
		return "ide", true
	case "terminal":
		return "terminal", true
	case "rpc":
		return "rpc", true
	case "api_debug":
		return "api_debug", true
	case "explorer":
		return "explorer", true
	case "network":
		return "network", true
	case "visualization":
		return "visualization", true
	default:
		return "", false
	}
}

func challengeToolLabel(toolKey string) string {
	switch toolKey {
	case "ide":
		return "在线编辑器"
	case "terminal":
		return "命令终端"
	case "files":
		return "文件"
	case "rpc":
		return "链上接口"
	case "api_debug":
		return "接口调试台"
	case "explorer":
		return "区块浏览器"
	case "logs":
		return "日志"
	case "network":
		return "节点协作面板"
	case "visualization":
		return "可视化实验"
	default:
		return toolKey
	}
}

func challengeToolKindFromEntry(entry, routeSegment string) string {
	switch entry {
	case "workspace", "ide":
		return "ide"
	case "visualization":
		return "visualization"
	default:
		return normalizeChallengeToolKind(routeSegment)
	}
}

func normalizeChallengeToolKind(value string) string {
	switch value {
	case "ide", "terminal", "files", "rpc", "explorer", "logs", "visualization", "api_debug":
		return value
	default:
		return ""
	}
}

func challengeServiceToolLabel(serviceSpec model.ChallengeServiceSpec, kind string) string {
	base := serviceSpec.Description
	if base == "" {
		base = serviceSpec.Key
	}

	switch kind {
	case "rpc":
		return base + " RPC"
	case "api_debug":
		return base + " API"
	case "explorer":
		return base + " 浏览器"
	case "visualization":
		return base + " 可视化"
	default:
		return base
	}
}
