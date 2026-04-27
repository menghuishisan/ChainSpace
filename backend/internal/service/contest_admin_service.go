package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chainspace/backend/internal/dto/request"
	"github.com/chainspace/backend/internal/dto/response"
	"github.com/chainspace/backend/internal/model"
	"github.com/chainspace/backend/internal/repository"
	"github.com/chainspace/backend/pkg/errors"
	"gorm.io/gorm"
)

// ContestAdminService handles contest lifecycle and management views.
type ContestAdminService struct {
	contestRepo          *repository.ContestRepository
	contestChallengeRepo *repository.ContestChallengeRepository
	teamRepo             *repository.TeamRepository
	teamMemberRepo       *repository.TeamMemberRepository
	imageRepo            *repository.DockerImageRepository
}

func NewContestAdminService(
	contestRepo *repository.ContestRepository,
	contestChallengeRepo *repository.ContestChallengeRepository,
	teamRepo *repository.TeamRepository,
	teamMemberRepo *repository.TeamMemberRepository,
	imageRepo *repository.DockerImageRepository,
) *ContestAdminService {
	return &ContestAdminService{
		contestRepo:          contestRepo,
		contestChallengeRepo: contestChallengeRepo,
		teamRepo:             teamRepo,
		teamMemberRepo:       teamMemberRepo,
		imageRepo:            imageRepo,
	}
}

func (s *ContestAdminService) CreateContest(ctx context.Context, creatorID uint, schoolID *uint, req *request.CreateContestRequest) (*response.ContestResponse, error) {
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		return nil, errors.ErrInvalidParams.WithMessage("开始时间格式错误，需要 RFC3339 格式")
	}
	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		return nil, errors.ErrInvalidParams.WithMessage("结束时间格式错误，需要 RFC3339 格式")
	}
	if !endTime.After(startTime) {
		return nil, errors.ErrInvalidParams.WithMessage("结束时间必须晚于开始时间")
	}

	contest := &model.Contest{
		CreatorID:           creatorID,
		Level:               req.Level,
		Title:               req.Title,
		Description:         req.Description,
		Type:                req.Type,
		Cover:               req.Cover,
		Rules:               req.Rules,
		BattleOrchestration: normalizeBattleOrchestration(req.BattleOrchestration, req.Type),
		StartTime:           startTime,
		EndTime:             endTime,
		Status:              model.ContestStatusDraft,
		IsPublic:            req.IsPublic,
		MaxParticipants:     req.MaxParticipants,
		TeamMinSize:         req.TeamMinSize,
		TeamMaxSize:         req.TeamMaxSize,
		DynamicScore:        req.DynamicScore,
		FirstBloodBonus:     req.FirstBloodBonus,
	}

	if schoolID != nil {
		contest.SchoolID = schoolID
	}
	if req.Level == "" {
		contest.Level = model.ContestLevelPractice
	}
	if req.TeamMinSize == 0 {
		contest.TeamMinSize = 1
	}
	if req.TeamMaxSize == 0 {
		contest.TeamMaxSize = 1
	}
	if req.RegistrationStart != "" {
		parsedTime, parseErr := time.Parse(time.RFC3339, req.RegistrationStart)
		if parseErr != nil {
			return nil, errors.ErrInvalidParams.WithMessage("报名开始时间格式错误，需要 RFC3339 格式")
		}
		contest.RegistrationStart = &parsedTime
	}
	if req.RegistrationEnd != "" {
		parsedTime, parseErr := time.Parse(time.RFC3339, req.RegistrationEnd)
		if parseErr != nil {
			return nil, errors.ErrInvalidParams.WithMessage("报名截止时间格式错误，需要 RFC3339 格式")
		}
		contest.RegistrationEnd = &parsedTime
	}
	if err := validateContestRegistrationWindow(startTime, contest.RegistrationStart, contest.RegistrationEnd); err != nil {
		return nil, err
	}
	if err := s.validateBattleOrchestrationRequirements(ctx, contest.Type, contest.BattleOrchestration); err != nil {
		return nil, err
	}

	if err := s.contestRepo.Create(ctx, contest); err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}
	return s.buildContestResponse(ctx, contest), nil
}

func (s *ContestAdminService) UpdateContest(ctx context.Context, contestID, userID, schoolID uint, role string, req *request.UpdateContestRequest) (*response.ContestResponse, error) {
	contest, err := s.contestRepo.GetByID(ctx, contestID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrContestNotFound
		}
		return nil, errors.ErrDatabaseError.WithError(err)
	}
	if err := ensureContestManageable(contest, userID, schoolID, role); err != nil {
		return nil, err
	}
	if req.Status != "" {
		return nil, errors.ErrInvalidParams.WithMessage("contest status must be changed via dedicated transition actions")
	}

	if req.Title != "" {
		contest.Title = req.Title
	}
	if req.Description != "" {
		contest.Description = req.Description
	}
	if req.Cover != "" {
		contest.Cover = req.Cover
	}
	if req.Rules != "" {
		contest.Rules = req.Rules
	}
	if req.StartTime != "" {
		parsedTime, parseErr := time.Parse(time.RFC3339, req.StartTime)
		if parseErr != nil {
			return nil, errors.ErrInvalidParams.WithMessage("开始时间格式错误，需要 RFC3339 格式")
		}
		contest.StartTime = parsedTime
	}
	if req.EndTime != "" {
		parsedTime, parseErr := time.Parse(time.RFC3339, req.EndTime)
		if parseErr != nil {
			return nil, errors.ErrInvalidParams.WithMessage("结束时间格式错误，需要 RFC3339 格式")
		}
		contest.EndTime = parsedTime
	}
	if !contest.EndTime.After(contest.StartTime) {
		return nil, errors.ErrInvalidParams.WithMessage("结束时间必须晚于开始时间")
	}
	if req.RegistrationStart != "" {
		parsedTime, parseErr := time.Parse(time.RFC3339, req.RegistrationStart)
		if parseErr != nil {
			return nil, errors.ErrInvalidParams.WithMessage("报名开始时间格式错误，需要 RFC3339 格式")
		}
		contest.RegistrationStart = &parsedTime
	}
	if req.RegistrationEnd != "" {
		parsedTime, parseErr := time.Parse(time.RFC3339, req.RegistrationEnd)
		if parseErr != nil {
			return nil, errors.ErrInvalidParams.WithMessage("报名截止时间格式错误，需要 RFC3339 格式")
		}
		contest.RegistrationEnd = &parsedTime
	}
	if err := validateContestRegistrationWindow(contest.StartTime, contest.RegistrationStart, contest.RegistrationEnd); err != nil {
		return nil, err
	}
	if req.IsPublic != nil {
		contest.IsPublic = *req.IsPublic
	}
	if req.MaxParticipants != nil {
		contest.MaxParticipants = *req.MaxParticipants
	}
	if req.TeamMinSize != nil {
		contest.TeamMinSize = *req.TeamMinSize
	}
	if req.TeamMaxSize != nil {
		contest.TeamMaxSize = *req.TeamMaxSize
	}
	if req.DynamicScore != nil {
		contest.DynamicScore = *req.DynamicScore
	}
	if req.FirstBloodBonus != nil {
		contest.FirstBloodBonus = *req.FirstBloodBonus
	}
	if req.BattleOrchestration != nil {
		contest.BattleOrchestration = normalizeBattleOrchestration(*req.BattleOrchestration, contest.Type)
		if validateErr := s.validateBattleOrchestrationRequirements(ctx, contest.Type, contest.BattleOrchestration); validateErr != nil {
			return nil, validateErr
		}
	}
	if req.Level != "" {
		contest.Level = req.Level
	}

	if err := s.contestRepo.Update(ctx, contest); err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}
	return s.buildContestResponse(ctx, contest), nil
}

func (s *ContestAdminService) transitionContestStatus(ctx context.Context, contest *model.Contest, targetStatus string) error {
	switch targetStatus {
	case model.ContestStatusPublished:
		if contest.Status != model.ContestStatusDraft {
			return errors.ErrInvalidParams.WithMessage("only draft contests can be published")
		}
	default:
		return errors.ErrInvalidParams.WithMessage("unsupported contest status transition")
	}

	contest.Status = targetStatus
	if err := s.contestRepo.Update(ctx, contest); err != nil {
		return errors.ErrDatabaseError.WithError(err)
	}
	return nil
}

func (s *ContestAdminService) GetContest(ctx context.Context, contestID, userID, schoolID uint, role string) (*response.ContestResponse, error) {
	contest, err := s.contestRepo.GetByID(ctx, contestID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrContestNotFound
		}
		return nil, errors.ErrDatabaseError.WithError(err)
	}
	if err := ensureContestAccessible(contest, userID, schoolID, role); err != nil {
		return nil, err
	}

	resp := s.buildContestResponse(ctx, contest)
	if userID > 0 {
		registered, _ := s.teamMemberRepo.IsInAnyTeam(ctx, contestID, userID)
		resp.IsRegistered = registered
	}
	return resp, nil
}

func (s *ContestAdminService) ListContests(ctx context.Context, schoolID uint, userID uint, role string, req *request.ListContestsRequest) ([]response.ContestResponse, int64, error) {
	repoStatus := normalizeContestListStatus(req.Status)
	contests, err := s.contestRepo.ListAll(ctx, schoolID, req.Type, req.Level, repoStatus, req.Keyword, req.IsPublic)
	if err != nil {
		return nil, 0, errors.ErrDatabaseError.WithError(err)
	}

	now := time.Now()
	filtered := make([]model.Contest, 0, len(contests))
	for _, contest := range contests {
		if !contestCanView(&contest, userID, schoolID, role) {
			continue
		}
		if req.Status != "" && !contest.MatchesStatusAt(req.Status, now) {
			continue
		}
		filtered = append(filtered, contest)
	}

	paged := paginateContests(filtered, req.GetPage(), req.GetPageSize())
	list := make([]response.ContestResponse, len(paged))
	for index, contest := range paged {
		list[index] = *s.buildContestResponse(ctx, &contest)
		if userID > 0 {
			registered, _ := s.teamMemberRepo.IsInAnyTeam(ctx, contest.ID, userID)
			list[index].IsRegistered = registered
		}
	}
	return list, int64(len(filtered)), nil
}

func (s *ContestAdminService) buildContestResponse(ctx context.Context, contest *model.Contest) *response.ContestResponse {
	resp := &response.ContestResponse{
		ID:                  contest.ID,
		SchoolID:            contest.SchoolID,
		CreatorID:           contest.CreatorID,
		Title:               contest.Title,
		Description:         contest.Description,
		Type:                contest.Type,
		Level:               contest.Level,
		Cover:               contest.Cover,
		Rules:               contest.Rules,
		BattleOrchestration: contest.BattleOrchestration,
		StartTime:           contest.StartTime,
		EndTime:             contest.EndTime,
		RegistrationStart:   contest.RegistrationStart,
		RegistrationEnd:     contest.RegistrationEnd,
		DynamicScore:        contest.DynamicScore,
		FirstBloodBonus:     contest.FirstBloodBonus,
		Status:              contest.CurrentStatus(),
		IsPublic:            contest.IsPublic,
		MaxParticipants:     contest.MaxParticipants,
		TeamMinSize:         contest.TeamMinSize,
		TeamMaxSize:         contest.TeamMaxSize,
		CreatedAt:           contest.CreatedAt,
	}

	if contest.Creator != nil {
		resp.CreatorName = contest.Creator.DisplayName()
	}

	participantCount, _ := s.teamRepo.CountParticipants(ctx, contest.ID)
	resp.ParticipantCount = participantCount
	challengeCount, _ := s.contestChallengeRepo.CountByContest(ctx, contest.ID)
	resp.ChallengeCount = challengeCount
	return resp
}

func normalizeContestListStatus(status string) string {
	switch status {
	case model.ContestStatusPublished, model.ContestStatusOngoing, model.ContestStatusEnded:
		return model.ContestStatusPublished
	default:
		return status
	}
}

func paginateContests(contests []model.Contest, page, pageSize int) []model.Contest {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	start := (page - 1) * pageSize
	if start >= len(contests) {
		return []model.Contest{}
	}

	end := start + pageSize
	if end > len(contests) {
		end = len(contests)
	}
	return contests[start:end]
}

func (s *ContestAdminService) validateBattleOrchestrationRequirements(ctx context.Context, contestType string, orchestration model.BattleOrchestration) error {
	if contestType != model.ContestTypeAgentBattle {
		return nil
	}
	if s.imageRepo == nil {
		return errors.ErrInternal.WithMessage("image repository is not initialized")
	}

	images, err := s.imageRepo.ListAll(ctx)
	if err != nil {
		return errors.ErrDatabaseError.WithError(err)
	}
	if len(images) == 0 {
		return errors.ErrInvalidParams.WithMessage("未找到可用镜像，无法校验对抗赛编排")
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
			return nil, errors.ErrInvalidParams.WithMessage("对抗赛编排镜像不能为空")
		}
		image, ok := imageByRef[key]
		if !ok {
			return nil, errors.ErrInvalidParams.WithMessage("对抗赛编排镜像未登记或不可用: " + ref)
		}
		return &image, nil
	}

	workspaceImage, err := resolveImage(orchestration.TeamWorkspace.Image)
	if err != nil {
		return err
	}
	if err := ensureToolKeysSupportedByImage(orchestration.TeamWorkspace.InteractionTools, workspaceImage, "battle.team_workspace"); err != nil {
		return err
	}

	if _, err := resolveImage(orchestration.SharedChain.Image); err != nil {
		return errors.ErrInvalidParams.WithMessage("共享链镜像校验失败: " + err.Error())
	}

	if strings.TrimSpace(orchestration.Judge.Image) != "" {
		if _, judgeErr := resolveImage(orchestration.Judge.Image); judgeErr != nil {
			return errors.ErrInvalidParams.WithMessage("裁判镜像校验失败: " + judgeErr.Error())
		}
	}

	return nil
}
