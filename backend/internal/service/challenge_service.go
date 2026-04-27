package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/chainspace/backend/internal/dto/request"
	"github.com/chainspace/backend/internal/model"
	"github.com/chainspace/backend/internal/repository"
	"github.com/chainspace/backend/pkg/errors"
)

// ChallengeService 负责题库核心业务、访问控制与题目基础 CRUD。
type ChallengeService struct {
	challengeRepo *repository.ChallengeRepository
	imageRepo     *repository.DockerImageRepository
}

func NewChallengeService(challengeRepo *repository.ChallengeRepository, imageRepo *repository.DockerImageRepository) *ChallengeService {
	return &ChallengeService{
		challengeRepo: challengeRepo,
		imageRepo:     imageRepo,
	}
}

func (s *ChallengeService) canAccessChallenge(challenge *model.Challenge, userID, schoolID uint, role string) bool {
	if role == model.RolePlatformAdmin {
		return true
	}
	if challenge.IsPublic {
		return true
	}
	if challenge.SchoolID != nil && schoolID > 0 && *challenge.SchoolID != schoolID {
		return false
	}
	if role == model.RoleSchoolAdmin {
		return true
	}
	return challenge.CreatorID == userID
}

func (s *ChallengeService) canManageChallenge(challenge *model.Challenge, userID, schoolID uint, role string) bool {
	if role == model.RolePlatformAdmin {
		return true
	}
	if challenge.SchoolID == nil || schoolID == 0 || *challenge.SchoolID != schoolID {
		return false
	}
	if role == model.RoleSchoolAdmin {
		return true
	}
	return role == model.RoleTeacher && challenge.CreatorID == userID
}

func (s *ChallengeService) ListChallenges(
	ctx context.Context,
	req *request.ListChallengesRequest,
	userID uint,
	schoolID uint,
	role string,
) ([]model.Challenge, int64, error) {
	return s.challengeRepo.List(
		ctx,
		role,
		schoolID,
		userID,
		req.Category,
		req.Difficulty,
		req.SourceType,
		req.Status,
		req.Keyword,
		req.IsPublic,
		req.GetPage(),
		req.GetPageSize(),
	)
}

func (s *ChallengeService) GetChallenge(ctx context.Context, id uint, userID, schoolID uint, role string) (*model.Challenge, error) {
	challenge, err := s.challengeRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if challenge == nil {
		return nil, errors.ErrNotFound
	}
	if !s.canAccessChallenge(challenge, userID, schoolID, role) {
		return nil, errors.ErrNotFound
	}
	return challenge, nil
}

func (s *ChallengeService) CreateChallenge(
	ctx context.Context,
	req *request.CreateChallengeRequest,
	creatorID uint,
	schoolID *uint,
	role string,
) (*model.Challenge, error) {
	runtimeProfile := resolveChallengeRuntimeProfile(req.RuntimeProfile, req.ChallengeOrchestration)
	orchestration := finalizeChallengeOrchestration(req.ChallengeOrchestration, runtimeProfile)
	if err := s.validateChallengeOrchestrationRequirements(ctx, orchestration); err != nil {
		return nil, err
	}
	isPublic := req.IsPublic && role == model.RolePlatformAdmin

	challenge := &model.Challenge{
		Title:                  req.Title,
		Description:            req.Description,
		Category:               req.Category,
		RuntimeProfile:         runtimeProfile,
		Difficulty:             req.Difficulty,
		BasePoints:             req.BasePoints,
		MinPoints:              req.MinPoints,
		DecayFactor:            req.DecayFactor,
		ContractCode:           req.ContractCode,
		SetupCode:              req.SetupCode,
		DeployScript:           req.DeployScript,
		CheckScript:            req.CheckScript,
		FlagType:               req.FlagType,
		FlagTemplate:           req.FlagTemplate,
		ChallengeOrchestration: orchestration,
		Hints:                  toModelJSONArray(req.Hints),
		Attachments:            toStringJSONArray(req.Attachments),
		Tags:                   model.StringList(req.Tags),
		CreatorID:              creatorID,
		SchoolID:               schoolID,
		IsPublic:               isPublic,
		Status:                 model.ChallengeStatusDraft,
		SourceType:             model.ChallengeSourceUserCreated,
	}

	if err := s.challengeRepo.Create(ctx, challenge); err != nil {
		return nil, err
	}
	return challenge, nil
}

func (s *ChallengeService) UpdateChallenge(ctx context.Context, id uint, userID, schoolID uint, role string, req *request.UpdateChallengeRequest) (*model.Challenge, error) {
	challenge, err := s.challengeRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if challenge == nil {
		return nil, errors.ErrNotFound
	}
	if !s.canManageChallenge(challenge, userID, schoolID, role) {
		return nil, errors.ErrNoPermission
	}

	if req.Title != nil {
		challenge.Title = *req.Title
	}
	if req.Description != nil {
		challenge.Description = *req.Description
	}
	if req.Category != nil {
		challenge.Category = *req.Category
	}
	if req.Difficulty != nil {
		challenge.Difficulty = *req.Difficulty
	}
	if req.BasePoints != nil {
		challenge.BasePoints = *req.BasePoints
	}
	if req.MinPoints != nil {
		challenge.MinPoints = *req.MinPoints
	}
	if req.DecayFactor != nil {
		challenge.DecayFactor = *req.DecayFactor
	}
	if req.ContractCode != nil {
		challenge.ContractCode = *req.ContractCode
	}
	if req.SetupCode != nil {
		challenge.SetupCode = *req.SetupCode
	}
	if req.DeployScript != nil {
		challenge.DeployScript = *req.DeployScript
	}
	if req.CheckScript != nil {
		challenge.CheckScript = *req.CheckScript
	}
	if req.FlagType != nil {
		challenge.FlagType = *req.FlagType
	}
	if req.FlagTemplate != nil {
		challenge.FlagTemplate = *req.FlagTemplate
	}
	if req.Hints != nil {
		challenge.Hints = toModelJSONArray(*req.Hints)
	}
	if req.Attachments != nil {
		challenge.Attachments = toStringJSONArray(*req.Attachments)
	}
	nextOrchestration := challenge.ChallengeOrchestration
	if req.ChallengeOrchestration != nil {
		nextOrchestration = *req.ChallengeOrchestration
	}
	if req.RuntimeProfile != nil || req.ChallengeOrchestration != nil {
		runtimeProfile := challenge.RuntimeProfile
		if req.RuntimeProfile != nil {
			runtimeProfile = *req.RuntimeProfile
		}
		challenge.RuntimeProfile = resolveChallengeRuntimeProfile(runtimeProfile, nextOrchestration)
	}
	if req.ChallengeOrchestration != nil {
		challenge.ChallengeOrchestration = finalizeChallengeOrchestration(nextOrchestration, challenge.RuntimeProfile)
		if validateErr := s.validateChallengeOrchestrationRequirements(ctx, challenge.ChallengeOrchestration); validateErr != nil {
			return nil, validateErr
		}
	}
	if req.Tags != nil {
		challenge.Tags = model.StringList(*req.Tags)
	}
	if req.IsPublic != nil {
		if *req.IsPublic && role != model.RolePlatformAdmin {
			return nil, errors.ErrInvalidParams.WithMessage("public challenge visibility requires review approval")
		}
		challenge.IsPublic = *req.IsPublic
	}
	if req.Status != nil {
		challenge.Status = *req.Status
	}

	if err := s.challengeRepo.Update(ctx, challenge); err != nil {
		return nil, err
	}
	return challenge, nil
}

func toModelJSONArray[T any](items []T) model.JSONArray {
	if len(items) == 0 {
		return model.JSONArray{}
	}

	result := make(model.JSONArray, 0, len(items))
	for _, item := range items {
		result = append(result, item)
	}
	return result
}

func toStringJSONArray(items []string) model.JSONArray {
	if len(items) == 0 {
		return model.JSONArray{}
	}

	result := make(model.JSONArray, 0, len(items))
	for _, item := range items {
		result = append(result, item)
	}
	return result
}

func (s *ChallengeService) DeleteChallenge(ctx context.Context, id uint, userID, schoolID uint, role string) error {
	challenge, err := s.challengeRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if !s.canManageChallenge(challenge, userID, schoolID, role) {
		return errors.ErrNoPermission
	}
	return s.challengeRepo.Delete(ctx, id)
}

func (s *ChallengeService) validateChallengeOrchestrationRequirements(ctx context.Context, orchestration model.ChallengeOrchestration) error {
	if !orchestration.NeedsEnvironment {
		return nil
	}
	if orchestration.Fork.Enabled {
		if strings.TrimSpace(orchestration.Fork.RPCURL) == "" {
			return errors.ErrInvalidParams.WithMessage("fork_replay 必须配置真实上游链 RPC URL")
		}
		if strings.TrimSpace(orchestration.Fork.Chain) == "" {
			return errors.ErrInvalidParams.WithMessage("fork_replay 必须配置链标识")
		}
		if orchestration.Fork.ChainID <= 0 {
			return errors.ErrInvalidParams.WithMessage("fork_replay 必须配置有效 chain_id")
		}
		if orchestration.Fork.BlockNumber <= 0 {
			return errors.ErrInvalidParams.WithMessage("fork_replay 必须固定 block_number，不能使用漂移状态")
		}
	}
	if s.imageRepo == nil {
		return errors.ErrInternal.WithMessage("image repository is not initialized")
	}
	images, err := s.imageRepo.ListAll(ctx)
	if err != nil {
		return errors.ErrDatabaseError.WithError(err)
	}
	if len(images) == 0 {
		return errors.ErrInvalidParams.WithMessage("未找到可用镜像，无法校验题目编排")
	}

	imageByRef := map[string]model.DockerImage{}
	for _, image := range images {
		fullName := strings.ToLower(strings.TrimSpace(image.FullName()))
		nameWithTag := strings.ToLower(strings.TrimSpace(fmt.Sprintf("%s:%s", image.Name, image.Tag)))
		nameOnly := strings.ToLower(strings.TrimSpace(image.Name))
		if fullName != "" {
			imageByRef[fullName] = image
		}
		if nameWithTag != "" {
			imageByRef[nameWithTag] = image
		}
		if nameOnly != "" {
			imageByRef[nameOnly] = image
		}
	}

	resolveImage := func(ref string) (*model.DockerImage, error) {
		key := strings.ToLower(strings.TrimSpace(ref))
		if key == "" {
			return nil, errors.ErrInvalidParams.WithMessage("题目编排镜像不能为空")
		}
		image, ok := imageByRef[key]
		if !ok {
			return nil, errors.ErrInvalidParams.WithMessage("题目编排镜像未登记或不可用: " + ref)
		}
		return &image, nil
	}

	workspaceImage, err := resolveImage(orchestration.Workspace.Image)
	if err != nil {
		return err
	}
	if err := ensureToolKeysSupportedByImage(orchestration.Workspace.InteractionTools, workspaceImage, "challenge.workspace"); err != nil {
		return err
	}

	serviceCapabilities := map[string]map[string]struct{}{}
	for _, serviceSpec := range orchestration.Services {
		serviceImage, serviceErr := resolveImage(serviceSpec.Image)
		if serviceErr != nil {
			return errors.ErrInvalidParams.WithMessage("题目服务 " + serviceSpec.Key + " 镜像校验失败: " + serviceErr.Error())
		}
		serviceCapabilities[serviceSpec.Key] = inferImageToolCapabilitySet(*serviceImage)
		for _, portSpec := range serviceSpec.Ports {
			toolKey, normalizeErr := normalizeChallengeExposeAs(portSpec.ExposeAs)
			if normalizeErr != nil {
				return normalizeErr
			}
			if toolKey == "" {
				continue
			}
			if _, ok := serviceCapabilities[serviceSpec.Key][toolKey]; !ok {
				return errors.ErrInvalidParams.WithMessage("题目服务 " + serviceSpec.Key + " 暴露了 " + toolKey + " 但镜像能力不支持")
			}
		}
	}

	for _, entry := range orchestration.Topology.ExposedEntrys {
		toolKey, normalizeErr := normalizeChallengeExposeAs(entry)
		if normalizeErr != nil {
			return normalizeErr
		}
		if toolKey == "" || toolKey == "ide" {
			continue
		}
		if toolKey == "rpc" && orchestration.Fork.Enabled {
			continue
		}
		if _, ok := inferImageToolCapabilitySet(*workspaceImage)[toolKey]; ok {
			continue
		}
		foundInService := false
		for _, caps := range serviceCapabilities {
			if _, ok := caps[toolKey]; ok {
				foundInService = true
				break
			}
		}
		if !foundInService {
			return errors.ErrInvalidParams.WithMessage("拓扑暴露入口 " + toolKey + " 未匹配到任何可用镜像能力")
		}
	}
	return nil
}

func normalizeChallengeExposeAs(value string) (string, error) {
	key := strings.TrimSpace(value)
	switch key {
	case "", "workspace":
		return "", nil
	case "ide", "terminal", "files", "logs", "rpc", "explorer", "api_debug", "visualization", "network":
		return key, nil
	default:
		return "", errors.ErrInvalidParams.WithMessage("不支持的 expose_as/tool_key: " + key)
	}
}

func (s *ChallengeService) RequestPublish(ctx context.Context, challengeID, userID, schoolID uint, role string, reason string) error {
	challenge, err := s.challengeRepo.GetByID(ctx, challengeID)
	if err != nil {
		return errors.ErrChallengeNotFound
	}
	if !s.canManageChallenge(challenge, userID, schoolID, role) {
		return errors.ErrNoPermission
	}
	if challenge.IsPublic {
		return errors.ErrInvalidParams.WithMessage("题目已经是公开状态")
	}

	publishRequest := &model.ChallengePublishRequest{
		ChallengeID: challengeID,
		ApplicantID: userID,
		Reason:      reason,
		Status:      model.ChallengePublishStatusPending,
	}
	return s.challengeRepo.CreatePublishRequest(ctx, publishRequest)
}
