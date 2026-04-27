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

type ExperimentService struct {
	experimentRepo     *repository.ExperimentRepository
	chapterRepo        *repository.ChapterRepository
	sessionRepo        *repository.ExperimentSessionRepository
	sessionMemberRepo  *repository.ExperimentSessionMemberRepository
	sessionMessageRepo *repository.ExperimentSessionMessageRepository
	envRepo            *repository.ExperimentEnvRepository
	submissionRepo     *repository.SubmissionRepository
	imageRepo          *repository.DockerImageRepository
	envManager         *EnvManagerService
	gradingService     *ExperimentGradingService
	runtimeService     *experimentRuntimeService
	sessionService     *experimentSessionService
	submissionService  *experimentSubmissionService
}

func NewExperimentService(
	experimentRepo *repository.ExperimentRepository,
	chapterRepo *repository.ChapterRepository,
	sessionRepo *repository.ExperimentSessionRepository,
	sessionMemberRepo *repository.ExperimentSessionMemberRepository,
	sessionMessageRepo *repository.ExperimentSessionMessageRepository,
	envRepo *repository.ExperimentEnvRepository,
	submissionRepo *repository.SubmissionRepository,
	imageRepo *repository.DockerImageRepository,
	envManager *EnvManagerService,
	gradingService *ExperimentGradingService,
) *ExperimentService {
	instance := &ExperimentService{
		experimentRepo:     experimentRepo,
		chapterRepo:        chapterRepo,
		sessionRepo:        sessionRepo,
		sessionMemberRepo:  sessionMemberRepo,
		sessionMessageRepo: sessionMessageRepo,
		envRepo:            envRepo,
		submissionRepo:     submissionRepo,
		imageRepo:          imageRepo,
		envManager:         envManager,
		gradingService:     gradingService,
	}
	instance.runtimeService = &experimentRuntimeService{core: instance}
	instance.sessionService = &experimentSessionService{core: instance}
	instance.submissionService = &experimentSubmissionService{core: instance}
	return instance
}

func (s *ExperimentService) CreateExperiment(ctx context.Context, schoolID, creatorID uint, req *request.CreateExperimentRequest) (*response.ExperimentResponse, error) {
	exp := &model.Experiment{
		SchoolID:      schoolID,
		ChapterID:     req.ChapterID,
		CreatorID:     creatorID,
		Title:         req.Title,
		Description:   req.Description,
		Type:          req.Type,
		Difficulty:    defaultInt(req.Difficulty, 1),
		MaxScore:      defaultInt(req.MaxScore, 100),
		PassScore:     defaultInt(req.PassScore, 60),
		AutoGrade:     req.AutoGrade,
		EstimatedTime: defaultInt(req.EstimatedTime, 60),
		SortOrder:     req.SortOrder,
		Status:        model.ExperimentStatusDraft,
		AllowLate:     req.AllowLate,
		LateDeduction: req.LateDeduction,
	}

	if req.StartTime != "" {
		startTime, err := time.Parse(time.RFC3339, req.StartTime)
		if err != nil {
			return nil, errors.ErrInvalidParams.WithMessage("开始时间格式错误，需要 RFC3339")
		}
		exp.StartTime = &startTime
	}
	if req.EndTime != "" {
		endTime, err := time.Parse(time.RFC3339, req.EndTime)
		if err != nil {
			return nil, errors.ErrInvalidParams.WithMessage("结束时间格式错误，需要 RFC3339")
		}
		exp.EndTime = &endTime
	}

	normalizedBlueprint := normalizeExperimentBlueprintSpec(req.Type, req.Blueprint)
	if err := s.validateBlueprintRuntimeRequirements(ctx, normalizedBlueprint); err != nil {
		return nil, err
	}
	applyBlueprintToExperiment(exp, normalizedBlueprint)
	if err := s.experimentRepo.Create(ctx, exp); err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	created, _ := s.experimentRepo.GetByID(ctx, exp.ID)
	resp := &response.ExperimentResponse{}
	return resp.FromExperiment(created), nil
}

func (s *ExperimentService) UpdateExperiment(ctx context.Context, expID, userID, schoolID uint, role string, req *request.UpdateExperimentRequest) (*response.ExperimentResponse, error) {
	exp, err := s.experimentRepo.GetByID(ctx, expID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrExperimentNotFound
		}
		return nil, errors.ErrDatabaseError.WithError(err)
	}
	if err := s.ensureManageExperiment(exp, userID, schoolID, role); err != nil {
		return nil, err
	}
	if req.Status != "" {
		return nil, errors.ErrInvalidParams.WithMessage("experiment status must be changed via dedicated actions")
	}
	if req.ChapterID != nil {
		chapter, chapterErr := s.requireManageChapter(ctx, *req.ChapterID, userID, schoolID, role)
		if chapterErr != nil {
			return nil, chapterErr
		}
		exp.ChapterID = chapter.ID
		exp.SchoolID = chapter.Course.SchoolID
	}

	if req.Title != "" {
		exp.Title = req.Title
	}
	if req.Description != "" {
		exp.Description = req.Description
	}
	if req.Type != "" {
		exp.Type = req.Type
	}
	if req.Difficulty != nil {
		exp.Difficulty = *req.Difficulty
	}
	if req.EstimatedTime != nil {
		exp.EstimatedTime = *req.EstimatedTime
	}
	if req.MaxScore != nil {
		exp.MaxScore = *req.MaxScore
	}
	if req.PassScore != nil {
		exp.PassScore = *req.PassScore
	}
	if req.AutoGrade != nil {
		exp.AutoGrade = *req.AutoGrade
	}
	if req.Blueprint != nil {
		normalizedBlueprint := normalizeExperimentBlueprintSpec(exp.Type, *req.Blueprint)
		if validateErr := s.validateBlueprintRuntimeRequirements(ctx, normalizedBlueprint); validateErr != nil {
			return nil, validateErr
		}
		applyBlueprintToExperiment(exp, normalizedBlueprint)
	} else {
		normalizedBlueprint := normalizeExperimentBlueprint(exp)
		if validateErr := s.validateBlueprintRuntimeRequirements(ctx, normalizedBlueprint); validateErr != nil {
			return nil, validateErr
		}
		applyBlueprintToExperiment(exp, normalizedBlueprint)
	}
	if req.SortOrder != nil {
		exp.SortOrder = *req.SortOrder
	}
	if req.AllowLate != nil {
		exp.AllowLate = *req.AllowLate
	}
	if req.LateDeduction != nil {
		exp.LateDeduction = *req.LateDeduction
	}
	if req.StartTime != "" {
		startTime, parseErr := time.Parse(time.RFC3339, req.StartTime)
		if parseErr != nil {
			return nil, errors.ErrInvalidParams.WithMessage("开始时间格式错误，需要 RFC3339")
		}
		exp.StartTime = &startTime
	}
	if req.EndTime != "" {
		endTime, parseErr := time.Parse(time.RFC3339, req.EndTime)
		if parseErr != nil {
			return nil, errors.ErrInvalidParams.WithMessage("结束时间格式错误，需要 RFC3339")
		}
		exp.EndTime = &endTime
	}

	if err := s.experimentRepo.Update(ctx, exp); err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	resp := &response.ExperimentResponse{}
	return resp.FromExperiment(exp), nil
}

func (s *ExperimentService) DeleteExperiment(ctx context.Context, expID, userID, schoolID uint, role string) error {
	exp, err := s.experimentRepo.GetByID(ctx, expID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrExperimentNotFound
		}
		return errors.ErrDatabaseError.WithError(err)
	}
	if err := s.ensureManageExperiment(exp, userID, schoolID, role); err != nil {
		return err
	}
	return s.experimentRepo.Delete(ctx, expID)
}

func (s *ExperimentService) PublishExperiment(ctx context.Context, expID, userID, schoolID uint, role string) error {
	exp, err := s.experimentRepo.GetByID(ctx, expID)
	if err != nil {
		return errors.ErrExperimentNotFound
	}
	if err := s.ensureManageExperiment(exp, userID, schoolID, role); err != nil {
		return err
	}
	if exp.Status != model.ExperimentStatusDraft {
		return errors.ErrInvalidParams.WithMessage("只能发布草稿状态的实验")
	}
	exp.Status = model.ExperimentStatusPublished
	return s.experimentRepo.Update(ctx, exp)
}

func (s *ExperimentService) GetExperiment(ctx context.Context, expID uint, userID *uint, schoolID uint, role string) (*response.ExperimentResponse, error) {
	exp, err := s.experimentRepo.GetByID(ctx, expID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrExperimentNotFound
		}
		return nil, errors.ErrDatabaseError.WithError(err)
	}
	if err := s.ensureViewExperiment(exp, schoolID, role); err != nil {
		return nil, err
	}

	resp := &response.ExperimentResponse{}
	resp.FromExperiment(exp)
	resp.SubmissionCount, _ = s.submissionRepo.CountByExperiment(ctx, expID)
	if userID != nil {
		if sub, latestErr := s.submissionRepo.GetLatestByStudent(ctx, expID, *userID); latestErr == nil && sub != nil {
			resp.MyScore = sub.Score
			resp.MyStatus = sub.Status
		}
	}
	return resp, nil
}

func (s *ExperimentService) ListExperiments(ctx context.Context, schoolID uint, role string, req *request.ListExperimentsRequest) ([]response.ExperimentResponse, int64, error) {
	status := req.Status
	if role == model.RoleStudent {
		status = model.ExperimentStatusPublished
	}

	exps, total, err := s.experimentRepo.List(ctx, schoolID, req.CourseID, req.ChapterID, req.Type, status, req.Keyword, req.GetPage(), req.GetPageSize())
	if err != nil {
		return nil, 0, errors.ErrDatabaseError.WithError(err)
	}

	list := make([]response.ExperimentResponse, 0, len(exps))
	for index := range exps {
		resp := &response.ExperimentResponse{}
		list = append(list, *resp.FromExperiment(&exps[index]))
	}
	return list, total, nil
}

func (s *ExperimentService) StartEnv(ctx context.Context, userID, schoolID uint, role string, req *request.StartEnvRequest) (*response.ExperimentEnvResponse, error) {
	return s.runtimeService.StartEnv(ctx, userID, schoolID, role, req)
}

func (s *ExperimentService) GetEnvStatus(ctx context.Context, envID string, userID, schoolID uint, role string) (*response.ExperimentEnvResponse, error) {
	return s.runtimeService.GetEnvStatus(ctx, envID, userID, schoolID, role)
}

func (s *ExperimentService) StopEnv(ctx context.Context, envID string, userID uint) error {
	return s.runtimeService.StopEnv(ctx, envID, userID)
}

func (s *ExperimentService) ExtendEnv(ctx context.Context, envID string, userID uint, req *request.ExtendEnvRequest) error {
	return s.runtimeService.ExtendEnv(ctx, envID, userID, req)
}

func (s *ExperimentService) PauseEnv(ctx context.Context, envID string, userID uint) error {
	return s.runtimeService.PauseEnv(ctx, envID, userID)
}

func (s *ExperimentService) ResumeEnv(ctx context.Context, envID string, userID uint) error {
	return s.runtimeService.ResumeEnv(ctx, envID, userID)
}

func (s *ExperimentService) CreateSnapshot(ctx context.Context, envID string, userID uint) (*response.ExperimentEnvResponse, error) {
	return s.runtimeService.CreateSnapshot(ctx, envID, userID)
}

func (s *ExperimentService) ListEnvs(ctx context.Context, userID, schoolID uint, role string, req *request.ListEnvsRequest) ([]response.ExperimentEnvResponse, int64, error) {
	return s.runtimeService.ListEnvs(ctx, userID, schoolID, role, req)
}

func (s *ExperimentService) ListSessions(ctx context.Context, userID, schoolID uint, role string, req *request.ListExperimentSessionsRequest) ([]response.ExperimentSessionResponse, int64, error) {
	return s.sessionService.ListSessions(ctx, userID, schoolID, role, req)
}

func (s *ExperimentService) GetSession(ctx context.Context, sessionKey string, userID, schoolID uint, role string) (*response.ExperimentSessionResponse, error) {
	return s.sessionService.GetSession(ctx, sessionKey, userID, schoolID, role)
}

func (s *ExperimentService) ListSessionMessages(ctx context.Context, sessionKey string, userID, schoolID uint, role string, req *request.ListExperimentSessionMessagesRequest) ([]response.ExperimentSessionMessageResponse, int64, error) {
	return s.sessionService.ListSessionMessages(ctx, sessionKey, userID, schoolID, role, req)
}

func (s *ExperimentService) SendSessionMessage(ctx context.Context, sessionKey string, userID, schoolID uint, role string, req *request.SendExperimentSessionMessageRequest) (*response.ExperimentSessionMessageResponse, error) {
	return s.sessionService.SendSessionMessage(ctx, sessionKey, userID, schoolID, role, req)
}

func (s *ExperimentService) UpdateSessionMember(
	ctx context.Context,
	sessionKey string,
	targetUserID, operatorUserID, schoolID uint,
	role string,
	req *request.UpdateExperimentSessionMemberRequest,
) (*response.ExperimentSessionResponse, error) {
	return s.sessionService.UpdateSessionMember(ctx, sessionKey, targetUserID, operatorUserID, schoolID, role, req)
}

func (s *ExperimentService) JoinSession(ctx context.Context, sessionKey string, userID, schoolID uint, role string) (*response.ExperimentSessionResponse, error) {
	return s.sessionService.JoinSession(ctx, sessionKey, userID, schoolID, role)
}

func (s *ExperimentService) LeaveSession(ctx context.Context, sessionKey string, userID, schoolID uint, role string) (*response.ExperimentSessionResponse, error) {
	return s.sessionService.LeaveSession(ctx, sessionKey, userID, schoolID, role)
}

func (s *ExperimentService) ListSessionLogs(ctx context.Context, sessionKey string, userID, schoolID uint, role string, source string, levels []string) ([]response.WorkspaceLogEntry, error) {
	return s.sessionService.ListSessionLogs(ctx, sessionKey, userID, schoolID, role, source, levels)
}

func (s *ExperimentService) SubmitExperiment(ctx context.Context, userID, schoolID uint, role string, req *request.SubmitExperimentRequest) (*response.SubmissionResponse, error) {
	return s.submissionService.SubmitExperiment(ctx, userID, schoolID, role, req)
}

func (s *ExperimentService) GradeSubmission(ctx context.Context, subID, graderID, schoolID uint, role string, req *request.GradeSubmissionRequest) (*response.SubmissionResponse, error) {
	return s.submissionService.GradeSubmission(ctx, subID, graderID, schoolID, role, req)
}

func (s *ExperimentService) ListSubmissions(ctx context.Context, userID, schoolID uint, role string, req *request.ListSubmissionsRequest) ([]response.SubmissionResponse, int64, error) {
	return s.submissionService.ListSubmissions(ctx, userID, schoolID, role, req)
}

func (s *ExperimentService) autoGradeSubmission(ctx context.Context, subID uint, exp *model.Experiment) {
	if s.gradingService != nil {
		s.gradingService.AutoGradeSubmission(ctx, subID, exp)
	}
}

func (s *ExperimentService) ensureManageExperiment(exp *model.Experiment, userID, schoolID uint, role string) error {
	if role == model.RolePlatformAdmin {
		return nil
	}
	if exp.SchoolID != schoolID {
		return errors.ErrNoPermission
	}
	if role == model.RoleSchoolAdmin {
		return nil
	}
	if role == model.RoleTeacher && exp.CreatorID == userID {
		return nil
	}
	return errors.ErrNoPermission
}

func (s *ExperimentService) requireManageChapter(ctx context.Context, chapterID, userID, schoolID uint, role string) (*model.Chapter, error) {
	if s.chapterRepo == nil {
		return nil, errors.ErrInternal.WithMessage("chapter repository is not initialized")
	}

	chapter, err := s.chapterRepo.GetByID(ctx, chapterID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrInvalidParams.WithMessage("target chapter not found")
		}
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	if chapter.Course == nil {
		return nil, errors.ErrInvalidParams.WithMessage("target chapter course relation is missing")
	}

	if role == model.RolePlatformAdmin {
		return chapter, nil
	}
	if chapter.Course.SchoolID != schoolID {
		return nil, errors.ErrNoPermission
	}
	if role == model.RoleSchoolAdmin {
		return chapter, nil
	}
	if role == model.RoleTeacher && chapter.Course.TeacherID == userID {
		return chapter, nil
	}

	return nil, errors.ErrNoPermission
}

func (s *ExperimentService) ensureViewExperiment(exp *model.Experiment, schoolID uint, role string) error {
	if role == model.RolePlatformAdmin {
		return nil
	}
	if exp.SchoolID != schoolID {
		return errors.ErrExperimentNotFound
	}
	if role == model.RoleStudent && exp.Status != model.ExperimentStatusPublished {
		return errors.ErrExperimentNotFound
	}
	return nil
}

func (s *ExperimentService) ensureEnvAccess(env *model.ExperimentEnv, userID, schoolID uint, role string) error {
	if role == model.RolePlatformAdmin {
		return nil
	}
	if env.SchoolID != schoolID {
		return errors.ErrNoPermission
	}
	if role == model.RoleSchoolAdmin || role == model.RoleTeacher {
		return nil
	}
	if env.UserID != userID {
		return errors.ErrNoPermission
	}
	return nil
}

func (s *ExperimentService) buildExperimentEnvResponse(ctx context.Context, env *model.ExperimentEnv, userID, schoolID uint, role string) (*response.ExperimentEnvResponse, error) {
	if env == nil {
		return nil, errors.ErrEnvNotFound
	}
	if err := s.ensureEnvAccess(env, userID, schoolID, role); err != nil {
		return nil, err
	}

	shapedEnv := env
	if role == model.RoleStudent {
		shapedEnv = shapeExperimentEnvForStudent(env, userID)
	}

	resp := &response.ExperimentEnvResponse{}
	result := resp.FromExperimentEnv(shapedEnv)
	s.filterExperimentRuntimeTools(ctx, shapedEnv, result, role)
	return result, nil
}

func (s *ExperimentService) filterExperimentRuntimeTools(
	_ context.Context,
	env *model.ExperimentEnv,
	resp *response.ExperimentEnvResponse,
	role string,
) {
	if env == nil || resp == nil || len(resp.Tools) == 0 {
		return
	}

	instanceByKey := make(map[string]model.ExperimentRuntimeInstance, len(env.RuntimeInstances))
	for _, instance := range env.RuntimeInstances {
		instanceByKey[instance.InstanceKey] = instance
	}
	toolMetaByKey := make(map[string][]model.ExperimentTool)
	allowedToolKeys := map[string]struct{}{}
	if env.Experiment != nil {
		for _, tool := range env.Experiment.Tools {
			toolMetaByKey[tool.ToolKey] = append(toolMetaByKey[tool.ToolKey], tool)
			if strings.TrimSpace(tool.ToolKey) != "" {
				allowedToolKeys[strings.TrimSpace(tool.ToolKey)] = struct{}{}
			}
		}
		if env.Experiment.Workspace != nil {
			for _, workspaceTool := range env.Experiment.Workspace.Tools {
				if strings.TrimSpace(workspaceTool.ToolKey) != "" {
					allowedToolKeys[strings.TrimSpace(workspaceTool.ToolKey)] = struct{}{}
				}
			}
		}
		for _, node := range env.Experiment.Nodes {
			for _, nodeTool := range node.Tools {
				if strings.TrimSpace(nodeTool.ToolKey) != "" {
					allowedToolKeys[strings.TrimSpace(nodeTool.ToolKey)] = struct{}{}
				}
			}
		}
		for _, serviceSpec := range env.Experiment.Services {
			if mappedTool, ok := serviceToolByKey[serviceSpec.ServiceKey]; ok {
				allowedToolKeys[mappedTool] = struct{}{}
			}
		}
	}

	servicePreferredKeys := map[string]struct{}{
		"rpc":           {},
		"explorer":      {},
		"api_debug":     {},
		"visualization": {},
	}
	hasNonWorkspaceForKey := make(map[string]bool, len(resp.Tools))
	for _, tool := range resp.Tools {
		if tool.InstanceKind != "workspace" {
			hasNonWorkspaceForKey[tool.Key] = true
		}
	}

	filtered := make([]response.RuntimeToolResponse, 0, len(resp.Tools))
	for _, tool := range resp.Tools {
		if len(allowedToolKeys) > 0 {
			if _, allowed := allowedToolKeys[tool.Key]; !allowed {
				continue
			}
		}
		instance, ok := instanceByKey[tool.InstanceKey]
		if !ok {
			continue
		}
		if metas, exists := toolMetaByKey[tool.Key]; exists && len(metas) > 0 {
			matched := false
			for _, meta := range metas {
				target := strings.TrimSpace(meta.Target)
				if target == "" {
					if tool.InstanceKind != "workspace" {
						continue
					}
					matched = true
					break
				}
				if target == tool.InstanceKey {
					matched = true
					break
				}
			}
			if !matched {
				if _, preferService := servicePreferredKeys[tool.Key]; !preferService || tool.InstanceKind == "workspace" {
					continue
				}
			}
			if !matched && tool.InstanceKind == "workspace" {
				continue
			}
		}
		if role == model.RoleStudent {
			if !tool.StudentFacing {
				continue
			}
			if !instance.StudentFacing && instance.Kind != "workspace" && instance.InstanceKey != env.PrimaryInstanceKey {
				continue
			}
		}
		if env.Status != model.EnvStatusRunning || instance.Status != model.EnvStatusRunning {
			continue
		}
		if tool.InstanceKind == "workspace" && hasNonWorkspaceForKey[tool.Key] {
			if _, preferService := servicePreferredKeys[tool.Key]; preferService {
				continue
			}
		}
		filtered = append(filtered, tool)
	}

	resp.Tools = filtered
}

func defaultInt(value, fallback int) int {
	if value == 0 {
		return fallback
	}
	return value
}

func buildSessionResponse(session *model.ExperimentSession) *response.ExperimentSessionResponse {
	if session == nil {
		return nil
	}
	resp := &response.ExperimentSessionResponse{
		ID:                 session.ID,
		SessionKey:         session.SessionKey,
		ExperimentID:       session.ExperimentID,
		Mode:               session.Mode,
		Status:             session.Status,
		PrimaryEnvID:       session.PrimaryEnvID,
		MaxMembers:         session.MaxMembers,
		CurrentMemberCount: session.CurrentMemberCount,
		StartedAt:          session.StartedAt,
		ExpiresAt:          session.ExpiresAt,
		Members:            make([]response.ExperimentSessionMemberResponse, 0, len(session.Members)),
	}
	for _, member := range session.Members {
		item := response.ExperimentSessionMemberResponse{
			UserID:          member.UserID,
			RoleKey:         member.RoleKey,
			AssignedNodeKey: member.AssignedNodeKey,
			JoinStatus:      member.JoinStatus,
			JoinedAt:        member.JoinedAt,
		}
		if member.User != nil {
			item.DisplayName = member.User.DisplayName()
			item.RealName = member.User.RealName
		}
		resp.Members = append(resp.Members, item)
	}
	return resp
}

func sessionContainsUser(session *model.ExperimentSession, userID uint) bool {
	for _, member := range session.Members {
		if member.UserID == userID && strings.EqualFold(member.JoinStatus, "joined") {
			return true
		}
	}
	return false
}

func sessionHasMember(session *model.ExperimentSession, userID uint) bool {
	for _, member := range session.Members {
		if member.UserID == userID {
			return true
		}
	}
	return false
}

func (s *ExperimentService) requireSessionAccess(ctx context.Context, sessionKey string, userID, schoolID uint, role string) (*model.ExperimentSession, error) {
	if s.sessionRepo == nil {
		return nil, errors.ErrEnvNotFound
	}
	session, err := s.sessionRepo.GetBySessionKey(ctx, sessionKey)
	if err != nil {
		return nil, errors.ErrEnvNotFound
	}
	if role != model.RolePlatformAdmin && session.SchoolID != schoolID {
		return nil, errors.ErrNoPermission
	}
	if role == model.RoleStudent && !sessionContainsUser(session, userID) {
		return nil, errors.ErrNoPermission
	}
	return session, nil
}

func (s *ExperimentService) requireSessionMembership(ctx context.Context, sessionKey string, userID, schoolID uint, role string) (*model.ExperimentSession, error) {
	if s.sessionRepo == nil {
		return nil, errors.ErrEnvNotFound
	}
	session, err := s.sessionRepo.GetBySessionKey(ctx, sessionKey)
	if err != nil {
		return nil, errors.ErrEnvNotFound
	}
	if role != model.RolePlatformAdmin && session.SchoolID != schoolID {
		return nil, errors.ErrNoPermission
	}
	if role == model.RoleStudent && !sessionHasMember(session, userID) {
		return nil, errors.ErrNoPermission
	}
	return session, nil
}

func (s *ExperimentService) requireManageSession(ctx context.Context, sessionKey string, userID, schoolID uint, role string) (*model.ExperimentSession, error) {
	session, err := s.requireSessionAccess(ctx, sessionKey, userID, schoolID, role)
	if err != nil {
		return nil, err
	}
	switch role {
	case model.RolePlatformAdmin:
		return session, nil
	case model.RoleSchoolAdmin, model.RoleTeacher:
		return session, nil
	}
	if session.PrimaryEnvID == "" {
		return nil, errors.ErrNoPermission
	}
	primaryEnv, err := s.envRepo.GetByEnvID(ctx, session.PrimaryEnvID)
	if err != nil {
		return nil, errors.ErrNoPermission
	}
	if primaryEnv.UserID != userID {
		return nil, errors.ErrNoPermission
	}
	return session, nil
}

func (s *ExperimentService) updateCurrentSessionJoinStatus(
	ctx context.Context,
	sessionKey string,
	userID, schoolID uint,
	role string,
	joinStatus string,
	allowRejoin bool,
) (*response.ExperimentSessionResponse, error) {
	var (
		session *model.ExperimentSession
		err     error
	)
	if allowRejoin {
		session, err = s.requireSessionAccess(ctx, sessionKey, userID, schoolID, role)
	} else {
		session, err = s.requireSessionMembership(ctx, sessionKey, userID, schoolID, role)
	}
	if err != nil {
		return nil, err
	}
	if s.sessionMemberRepo == nil {
		return nil, errors.ErrInternal.WithMessage("experiment session member repository is not initialized")
	}

	member, err := s.sessionMemberRepo.GetBySessionAndUser(ctx, session.ID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound.WithMessage("session member not found")
		}
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	member.JoinStatus = joinStatus
	if err := s.sessionMemberRepo.Update(ctx, member); err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	updatedSession, err := s.sessionRepo.GetByID(ctx, session.ID)
	if err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}
	if err := s.sessionRepo.UpdateCounters(
		ctx,
		updatedSession.ID,
		countJoinedSessionMembers(updatedSession),
		updatedSession.PrimaryEnvID,
		updatedSession.Status,
		nil,
		updatedSession.ExpiresAt,
	); err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	finalSession, err := s.sessionRepo.GetByID(ctx, session.ID)
	if err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}
	return buildSessionResponse(finalSession), nil
}

func (s *ExperimentService) envManagerLogs(ctx context.Context, env *model.ExperimentEnv, instance *model.ExperimentRuntimeInstance, source string, levels []string) ([]response.WorkspaceLogEntry, error) {
	if s.envManager == nil || s.envManager.k8sClient == nil || instance == nil {
		return nil, errors.ErrInternal.WithMessage("experiment runtime service is not initialized")
	}
	target := &WorkspaceAccessService{
		k8sClient: s.envManager.k8sClient,
	}
	return target.GetWorkspaceLogs(ctx, &WorkspaceRuntimeTarget{
		EnvID:   env.EnvID,
		PodName: instance.PodName,
		Status:  instance.Status,
	}, source, levels)
}

func shapeExperimentEnvForStudent(env *model.ExperimentEnv, userID uint) *model.ExperimentEnv {
	if env == nil {
		return nil
	}

	cloned := *env
	if env.Session == nil || env.SessionMode != model.ExperimentModeCollaboration {
		cloned.RuntimeInstances = filterStudentFacingInstances(env.RuntimeInstances, env.PrimaryInstanceKey)
		return &cloned
	}

	member := findExperimentSessionMember(env.Session, userID)
	allowedNodes, allowedTools := resolveExperimentMemberVisibility(env.Experiment, member)
	cloned.RuntimeInstances = filterVisibleCollaborationInstances(env.RuntimeInstances, env.PrimaryInstanceKey, allowedNodes, allowedTools)
	if !containsRuntimeInstance(cloned.RuntimeInstances, cloned.PrimaryInstanceKey) && len(cloned.RuntimeInstances) > 0 {
		cloned.PrimaryInstanceKey = cloned.RuntimeInstances[0].InstanceKey
	}
	return &cloned
}

func filterStudentFacingInstances(instances []model.ExperimentRuntimeInstance, primaryInstanceKey string) []model.ExperimentRuntimeInstance {
	filtered := make([]model.ExperimentRuntimeInstance, 0, len(instances))
	for _, instance := range instances {
		if instance.StudentFacing || instance.InstanceKey == primaryInstanceKey || instance.Kind == "workspace" {
			filtered = append(filtered, instance)
		}
	}
	return filtered
}

func filterVisibleCollaborationInstances(
	instances []model.ExperimentRuntimeInstance,
	primaryInstanceKey string,
	allowedNodes map[string]struct{},
	allowedTools map[string]struct{},
) []model.ExperimentRuntimeInstance {
	filtered := make([]model.ExperimentRuntimeInstance, 0, len(instances))
	for _, instance := range instances {
		if instance.StudentFacing || instance.InstanceKey == primaryInstanceKey || instance.Kind == "workspace" {
			filtered = append(filtered, instance)
			continue
		}
		if _, ok := allowedNodes[instance.InstanceKey]; ok {
			filtered = append(filtered, instance)
			continue
		}
		for _, tool := range instance.Tools {
			if _, ok := allowedTools[tool.ToolKey]; ok {
				filtered = append(filtered, instance)
				break
			}
		}
	}
	return filtered
}

func resolveExperimentMemberVisibility(exp *model.Experiment, member *model.ExperimentSessionMember) (map[string]struct{}, map[string]struct{}) {
	allowedNodes := map[string]struct{}{}
	allowedTools := map[string]struct{}{}
	if member == nil {
		return allowedNodes, allowedTools
	}
	if member.AssignedNodeKey != "" {
		allowedNodes[member.AssignedNodeKey] = struct{}{}
	}
	if exp == nil || exp.Collaboration == nil {
		return allowedNodes, allowedTools
	}

	for _, binding := range exp.Collaboration.Roles {
		if binding.RoleKey != member.RoleKey {
			continue
		}
		for _, node := range binding.NodeAssignments {
			if node.NodeKey != "" {
				allowedNodes[node.NodeKey] = struct{}{}
			}
		}
		for _, tool := range binding.ToolAssignments {
			if tool.ToolKey != "" {
				allowedTools[tool.ToolKey] = struct{}{}
			}
		}
		break
	}
	return allowedNodes, allowedTools
}

func findExperimentSessionMember(session *model.ExperimentSession, userID uint) *model.ExperimentSessionMember {
	if session == nil {
		return nil
	}
	for index := range session.Members {
		if session.Members[index].UserID == userID {
			return &session.Members[index]
		}
	}
	return nil
}

func containsRuntimeInstance(instances []model.ExperimentRuntimeInstance, instanceKey string) bool {
	for _, instance := range instances {
		if instance.InstanceKey == instanceKey {
			return true
		}
	}
	return false
}

func (s *ExperimentService) validateBlueprintRuntimeRequirements(ctx context.Context, blueprint model.ExperimentBlueprint) error {
	if s.imageRepo == nil {
		return errors.ErrInternal.WithMessage("image repository is not initialized")
	}
	images, err := s.imageRepo.ListAll(ctx)
	if err != nil {
		return errors.ErrDatabaseError.WithError(err)
	}
	if len(images) == 0 {
		return errors.ErrInvalidParams.WithMessage("未找到可用镜像，无法校验实验编排")
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
			return nil, errors.ErrInvalidParams.WithMessage("编排镜像不能为空")
		}
		image, ok := imageByRef[key]
		if !ok {
			return nil, errors.ErrInvalidParams.WithMessage("编排镜像未登记或不可用: " + ref)
		}
		return &image, nil
	}

	workspaceImage, err := resolveImage(blueprint.Workspace.Image)
	if err != nil {
		return err
	}
	if err := ensureToolKeysSupportedByImage(blueprint.Workspace.InteractionTools, workspaceImage, "workspace"); err != nil {
		return err
	}

	nodeImageByKey := map[string]*model.DockerImage{}
	for _, node := range blueprint.Nodes {
		nodeImage, nodeErr := resolveImage(node.Image)
		if nodeErr != nil {
			return errors.ErrInvalidParams.WithMessage("节点 " + node.Key + " 镜像校验失败: " + nodeErr.Error())
		}
		nodeImageByKey[node.Key] = nodeImage
		if toolErr := ensureToolKeysSupportedByImage(node.InteractionTools, nodeImage, "node:"+node.Key); toolErr != nil {
			return toolErr
		}
	}

	serviceImageByKey := map[string]*model.DockerImage{}
	for _, serviceSpec := range blueprint.Services {
		serviceImage, serviceErr := resolveImage(serviceSpec.Image)
		if serviceErr != nil {
			return errors.ErrInvalidParams.WithMessage("服务 " + serviceSpec.Key + " 镜像校验失败: " + serviceErr.Error())
		}
		serviceImageByKey[serviceSpec.Key] = serviceImage
	}

	for _, tool := range blueprint.Tools {
		target := strings.TrimSpace(tool.Target)
		if target == "" || target == "workspace" {
			if toolErr := ensureToolKeysSupportedByImage([]string{tool.Key}, workspaceImage, "tool:"+tool.Key+"@workspace"); toolErr != nil {
				return toolErr
			}
			continue
		}

		if serviceImage, ok := serviceImageByKey[target]; ok {
			if toolErr := ensureToolKeysSupportedByImage([]string{tool.Key}, serviceImage, "tool:"+tool.Key+"@service:"+target); toolErr != nil {
				return toolErr
			}
			continue
		}
		if nodeImage, ok := nodeImageByKey[target]; ok {
			if toolErr := ensureToolKeysSupportedByImage([]string{tool.Key}, nodeImage, "tool:"+tool.Key+"@node:"+target); toolErr != nil {
				return toolErr
			}
			continue
		}

		return errors.ErrInvalidParams.WithMessage("工具目标未定义: " + target)
	}

	return nil
}

func ensureToolKeysSupportedByImage(toolKeys []string, image *model.DockerImage, scope string) error {
	if image == nil || len(toolKeys) == 0 {
		return nil
	}
	capabilities := inferImageToolCapabilitySet(*image)
	for _, tool := range toolKeys {
		key := strings.TrimSpace(tool)
		if key == "" {
			continue
		}
		if _, ok := capabilities[key]; !ok {
			return errors.ErrInvalidParams.WithMessage("镜像能力不足: " + scope + " 需要工具 " + key + "，但镜像 " + image.FullName() + " 不支持")
		}
	}
	return nil
}

func inferImageToolCapabilitySet(image model.DockerImage) map[string]struct{} {
	result := map[string]struct{}{}
	add := func(key string) {
		if _, ok := runtimeToolCapabilityKeySet[key]; !ok {
			return
		}
		result[key] = struct{}{}
	}

	normalizedName := strings.ToLower(strings.TrimSpace(image.Name))
	if catalog, ok := runtimeImageToolCapabilityCatalog[normalizedName]; ok {
		for _, tool := range catalog {
			add(tool)
		}
	}

	for _, rawFeature := range toStringSlice(image.Features) {
		feature := strings.ToLower(strings.TrimSpace(rawFeature))
		if feature == "" {
			continue
		}
		if strings.HasPrefix(feature, "tool:") {
			add(strings.TrimSpace(strings.TrimPrefix(feature, "tool:")))
			continue
		}
		if strings.HasPrefix(feature, "tools=") {
			for _, item := range strings.Split(strings.TrimPrefix(feature, "tools="), ",") {
				add(strings.TrimSpace(item))
			}
		}
	}

	for key, value := range image.EnvVars {
		lowerKey := strings.ToLower(strings.TrimSpace(key))
		if lowerKey != "runtime_tools" && lowerKey != "interaction_tools" {
			continue
		}
		raw := strings.TrimSpace(fmt.Sprintf("%v", value))
		if raw == "" || raw == "<nil>" {
			continue
		}
		for _, item := range strings.Split(strings.ToLower(raw), ",") {
			add(strings.TrimSpace(item))
		}
	}

	return result
}

var runtimeToolCapabilityKeySet = map[string]struct{}{
	"ide":           {},
	"terminal":      {},
	"files":         {},
	"logs":          {},
	"rpc":           {},
	"explorer":      {},
	"api_debug":     {},
	"visualization": {},
	"network":       {},
}

var runtimeImageToolCapabilityCatalog = map[string][]string{
	"chainspace/eth-dev":    {"ide", "terminal", "files", "logs", "rpc"},
	"chainspace/security":   {"ide", "terminal", "files", "logs"},
	"chainspace/simulation": {"terminal", "logs", "visualization"},
	"chainspace/geth":       {"terminal", "logs", "rpc", "explorer", "api_debug", "network"},
	"chainspace/ipfs":       {"terminal", "logs", "api_debug"},
	"chainspace/blockscout": {"explorer"},
	"chainspace/chainlink":  {"api_debug"},
	"chainspace/thegraph":   {"api_debug"},
	"chainspace/fabric":     {"terminal", "logs", "rpc", "network"},
}
