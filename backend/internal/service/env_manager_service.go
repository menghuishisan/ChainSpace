package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/chainspace/backend/internal/config"
	"github.com/chainspace/backend/internal/model"
	"github.com/chainspace/backend/internal/pkg/k8s"
	"github.com/chainspace/backend/internal/pkg/logger"
	"github.com/chainspace/backend/internal/pkg/safego"
	"github.com/chainspace/backend/internal/pkg/websocket"
	"github.com/chainspace/backend/internal/repository"
	"github.com/chainspace/backend/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type EnvManagerService struct {
	k8sClient         *k8s.Client
	provisioner       *RuntimeProvisionService
	sessionRepo       *repository.ExperimentSessionRepository
	sessionMemberRepo *repository.ExperimentSessionMemberRepository
	envRepo           *repository.ExperimentEnvRepository
	experimentRepo    *repository.ExperimentRepository
	dockerImageRepo   *repository.DockerImageRepository
	notifyRepo        *repository.NotificationRepository
	uploadService     *UploadService
	wsHub             *websocket.Hub
	cfg               *config.Config
}

func NewEnvManagerService(
	k8sClient *k8s.Client,
	sessionRepo *repository.ExperimentSessionRepository,
	sessionMemberRepo *repository.ExperimentSessionMemberRepository,
	envRepo *repository.ExperimentEnvRepository,
	experimentRepo *repository.ExperimentRepository,
	dockerImageRepo *repository.DockerImageRepository,
	notifyRepo *repository.NotificationRepository,
	uploadService *UploadService,
	wsHub *websocket.Hub,
	cfg *config.Config,
) *EnvManagerService {
	return &EnvManagerService{
		k8sClient:         k8sClient,
		provisioner:       NewRuntimeProvisionService(k8sClient),
		sessionRepo:       sessionRepo,
		sessionMemberRepo: sessionMemberRepo,
		envRepo:           envRepo,
		experimentRepo:    experimentRepo,
		dockerImageRepo:   dockerImageRepo,
		notifyRepo:        notifyRepo,
		uploadService:     uploadService,
		wsHub:             wsHub,
		cfg:               cfg,
	}
}

type EnvCreateRequest struct {
	ExperimentID uint
	UserID       uint
	SchoolID     uint
	SnapshotURL  string
}

type EnvInfo struct {
	EnvID     string     `json:"env_id"`
	Status    string     `json:"status"`
	AccessURL string     `json:"access_url"`
	ExpiresAt *time.Time `json:"expires_at"`
}

func (s *EnvManagerService) CreateEnv(ctx context.Context, req *EnvCreateRequest) (*EnvInfo, error) {
	if s.k8sClient == nil {
		return nil, errors.ErrInternal.WithMessage("experiment runtime is not initialized")
	}

	exp, err := s.experimentRepo.GetByID(ctx, req.ExperimentID)
	if err != nil {
		return nil, errors.ErrExperimentNotFound
	}
	blueprint := normalizeExperimentBlueprint(exp)

	if activeEnv, activeErr := s.envRepo.GetActiveByUser(ctx, req.UserID, req.ExperimentID); activeErr == nil && activeEnv != nil {
		return s.buildEnvInfo(activeEnv), nil
	}

	if blueprint.Mode == model.ExperimentModeCollaboration && s.sessionRepo != nil {
		if session, sessionErr := s.sessionRepo.GetReusableByExperiment(ctx, req.ExperimentID); sessionErr == nil && session != nil {
			if session.CurrentMemberCount < maxCollaborationMembers(blueprint) {
				return s.attachToSharedSession(ctx, session, req)
			}
		}
	}

	userActiveCount, _ := s.envRepo.CountActiveByUser(ctx, req.UserID)
	if userActiveCount >= int64(s.cfg.Experiment.MaxUserEnvs) {
		return nil, errors.ErrConflict.WithMessage(fmt.Sprintf("已达到并发环境上限(%d个)，请先关闭其他环境", s.cfg.Experiment.MaxUserEnvs))
	}

	schoolActiveCount, _ := s.envRepo.CountActiveBySchool(ctx, req.SchoolID)
	if schoolActiveCount >= int64(s.cfg.Experiment.MaxSchoolEnvs) {
		return nil, errors.ErrConflict.WithMessage("学校实验环境资源已满，请稍后重试")
	}

	envID := fmt.Sprintf("env-%d-%d-%d", req.ExperimentID, req.UserID, time.Now().Unix())
	timeout := s.cfg.Experiment.DefaultTimeout
	if exp.EstimatedTime > 0 {
		timeout = time.Duration(exp.EstimatedTime) * time.Minute
	}
	expiresAt := time.Now().Add(timeout)

	sessionKey := ""
	if blueprint.Mode == model.ExperimentModeCollaboration {
		sessionKey = fmt.Sprintf("session-%d-%d", req.ExperimentID, time.Now().UnixNano())
	}
	runtime := buildExperimentRuntimeState(envID, blueprint)
	session, err := s.ensureSessionForEnv(ctx, exp, req, blueprint, envID, expiresAt, sessionKey)
	if err != nil {
		return nil, err
	}
	env := &model.ExperimentEnv{
		EnvID:        envID,
		ExperimentID: req.ExperimentID,
		SessionID:    sessionIDPtr(session),
		UserID:       req.UserID,
		SchoolID:     req.SchoolID,
		Status:       model.EnvStatusCreating,
		ExpiresAt:    &expiresAt,
		SnapshotURL:  req.SnapshotURL,
	}
	applyRuntimeStateToEnv(env, runtime)
	if err := s.envRepo.Create(ctx, env); err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}
	if session != nil {
		if err := s.bindEnvToSession(ctx, session, env, req.UserID, blueprint); err != nil {
			return nil, err
		}
	}

	if err := s.createRuntimeInstances(ctx, exp, env, blueprint, runtime); err != nil {
		logger.Error("experiment runtime provisioning failed",
			zap.String("env_id", env.EnvID),
			zap.Uint("experiment_id", exp.ID),
			zap.Uint("user_id", req.UserID),
			zap.Error(err),
		)
		env.Status = model.EnvStatusFailed
		env.ErrorMessage = err.Error()
		_ = s.envRepo.Update(ctx, env)
		return nil, errors.ErrEnvStartFailed.WithError(err)
	}

	safego.GoWithTimeout("waitForExperimentRuntime:"+env.EnvID, 10*time.Minute, func(asyncCtx context.Context) {
		s.waitForRuntimeReady(asyncCtx, env, blueprint, runtime)
	})

	return s.buildEnvInfo(env), nil
}

func (s *EnvManagerService) attachToSharedSession(ctx context.Context, session *model.ExperimentSession, req *EnvCreateRequest) (*EnvInfo, error) {
	if session == nil {
		return nil, errors.ErrEnvNotFound
	}
	expiresAt := session.ExpiresAt
	primaryEnv, err := s.envRepo.GetByEnvID(ctx, session.PrimaryEnvID)
	if err != nil {
		return nil, errors.ErrEnvNotFound
	}
	runtime := buildRuntimeStateFromEnv(primaryEnv)
	env := &model.ExperimentEnv{
		EnvID:        fmt.Sprintf("env-%d-%d-%d", req.ExperimentID, req.UserID, time.Now().Unix()),
		ExperimentID: req.ExperimentID,
		SessionID:    &session.ID,
		UserID:       req.UserID,
		SchoolID:     req.SchoolID,
		Status:       primaryEnv.Status,
		ExpiresAt:    expiresAt,
	}
	applyRuntimeStateToEnv(env, runtime)
	if err := s.envRepo.Create(ctx, env); err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}
	if err := s.bindEnvToSession(ctx, session, env, req.UserID, normalizeExperimentBlueprint(primaryEnv.Experiment)); err != nil {
		return nil, err
	}
	return s.buildEnvInfo(env), nil
}

func (s *EnvManagerService) StopEnv(ctx context.Context, envID string, userID uint) error {
	env, err := s.envRepo.GetByEnvID(ctx, envID)
	if err != nil {
		return errors.ErrEnvNotFound
	}
	if userID != 0 && env.UserID != userID {
		return errors.ErrNoPermission
	}

	runtime := buildRuntimeStateFromEnv(env)
	if env.SessionID != nil && runtime.SessionMode == model.ExperimentModeCollaboration {
		count, countErr := s.envRepo.CountActiveBySession(ctx, *env.SessionID)
		if countErr == nil && count > 1 {
			if err := s.envRepo.UpdateStatus(ctx, env.EnvID, model.EnvStatusTerminated); err != nil {
				return errors.ErrDatabaseError.WithError(err)
			}
			s.handleSessionDetach(ctx, env)
			s.notifyEnvStatus(env.EnvID, model.EnvStatusTerminated, "experiment session detached")
			return nil
		}
	}

	return s.terminateEnv(ctx, env)
}

func (s *EnvManagerService) PauseEnv(ctx context.Context, envID string, userID uint) error {
	env, err := s.envRepo.GetByEnvID(ctx, envID)
	if err != nil {
		return errors.ErrEnvNotFound
	}
	if env.UserID != userID {
		return errors.ErrNoPermission
	}
	if env.Status != model.EnvStatusRunning {
		return errors.ErrEnvNotRunning
	}

	snapshotURL, snapshotErr := s.CreateSnapshot(ctx, envID, userID)
	if snapshotErr != nil {
		return snapshotErr
	}
	env.SnapshotURL = snapshotURL
	if err := s.deleteRuntimeInstances(ctx, buildRuntimeStateFromEnv(env)); err != nil {
		logger.Warn("pause experiment runtime cleanup failed", zap.String("env_id", envID), zap.Error(err))
	}
	env.Status = model.EnvStatusPaused
	if err := s.envRepo.Update(ctx, env); err != nil {
		return errors.ErrDatabaseError.WithError(err)
	}
	s.syncSessionStatus(ctx, env, model.EnvStatusPaused)
	s.notifyEnvStatus(env.EnvID, model.EnvStatusPaused, "experiment paused")
	return nil
}

func (s *EnvManagerService) ResumeEnv(ctx context.Context, envID string, userID uint) error {
	env, err := s.envRepo.GetByEnvID(ctx, envID)
	if err != nil {
		return errors.ErrEnvNotFound
	}
	if env.UserID != userID {
		return errors.ErrNoPermission
	}
	if env.Status != model.EnvStatusPaused {
		return errors.ErrBadRequest.WithMessage("environment is not paused")
	}

	exp, err := s.experimentRepo.GetByID(ctx, env.ExperimentID)
	if err != nil {
		return errors.ErrExperimentNotFound
	}
	blueprint := normalizeExperimentBlueprint(exp)
	runtime := buildRuntimeStateFromEnv(env)
	if err := s.createRuntimeInstances(ctx, exp, env, blueprint, runtime); err != nil {
		return errors.ErrEnvStartFailed.WithError(err)
	}
	env.Status = model.EnvStatusCreating
	if err := s.envRepo.Update(ctx, env); err != nil {
		return errors.ErrDatabaseError.WithError(err)
	}
	s.syncSessionStatus(ctx, env, model.EnvStatusCreating)

	safego.GoWithTimeout("resumeExperimentRuntime:"+env.EnvID, 10*time.Minute, func(asyncCtx context.Context) {
		s.waitForRuntimeReady(asyncCtx, env, blueprint, runtime)
	})
	return nil
}

func (s *EnvManagerService) ExtendEnv(ctx context.Context, envID string, userID uint, minutes int) error {
	env, err := s.envRepo.GetByEnvID(ctx, envID)
	if err != nil {
		return errors.ErrEnvNotFound
	}
	if env.UserID != userID {
		return errors.ErrNoPermission
	}
	if env.ExtendCount >= s.cfg.Experiment.MaxExtendTimes {
		return errors.ErrEnvExtendLimit
	}

	newExpireTime := env.ExpiresAt.Add(time.Duration(minutes) * time.Minute)
	if err := s.envRepo.UpdateExpireTime(ctx, envID, newExpireTime); err != nil {
		return errors.ErrDatabaseError.WithError(err)
	}
	return nil
}

func (s *EnvManagerService) CreateSnapshot(ctx context.Context, envID string, userID uint) (string, error) {
	env, err := s.envRepo.GetByEnvID(ctx, envID)
	if err != nil {
		return "", errors.ErrEnvNotFound
	}
	if env.UserID != userID {
		return "", errors.ErrNoPermission
	}
	if env.SnapshotAt != nil {
		if elapsed := time.Since(*env.SnapshotAt); elapsed < 20*time.Minute {
			return "", errors.ErrTooManyRequests.WithMessage(fmt.Sprintf("快照操作过于频繁，请在%d分钟后重试", int((20*time.Minute-elapsed).Minutes())+1))
		}
	}

	snapshotCount, countErr := s.envRepo.CountUserSnapshots(ctx, userID, env.ExperimentID)
	if countErr == nil && snapshotCount >= 10 {
		return "", errors.ErrConflict.WithMessage("快照数量已达上限(10个)，请删除旧快照后重试")
	}

	snapshotURL, snapshotErr := s.saveEnvSnapshot(ctx, env)
	if snapshotErr != nil {
		return "", errors.ErrInternal.WithError(snapshotErr)
	}
	if err := s.persistSnapshotMetadata(ctx, env, snapshotURL); err != nil {
		return "", errors.ErrDatabaseError.WithError(err)
	}
	return snapshotURL, nil
}

func (s *EnvManagerService) GetEnvStatus(ctx context.Context, envID string) (*EnvInfo, error) {
	env, err := s.envRepo.GetByEnvID(ctx, envID)
	if err != nil {
		return nil, errors.ErrEnvNotFound
	}
	s.syncEnvRuntimeStatus(ctx, env)
	return s.buildEnvInfo(env), nil
}

func (s *EnvManagerService) CleanupExpiredEnvs(ctx context.Context) error {
	expiredEnvs, err := s.envRepo.ListExpired(ctx)
	if err != nil {
		return err
	}

	for _, env := range expiredEnvs {
		if env.Status == model.EnvStatusRunning {
			if _, snapshotErr := s.saveEnvSnapshot(ctx, &env); snapshotErr != nil {
				logger.Warn("save snapshot before cleanup failed", zap.String("env_id", env.EnvID), zap.Error(snapshotErr))
			}
		}
		if err := s.terminateEnv(ctx, &env); err != nil {
			logger.Error("cleanup experiment env failed", zap.String("env_id", env.EnvID), zap.Error(err))
		}
	}
	return nil
}

func (s *EnvManagerService) NotifyExpiringEnvs(ctx context.Context, minutes int) error {
	expiringEnvs, err := s.envRepo.ListExpiring(ctx, minutes)
	if err != nil {
		return err
	}

	for _, env := range expiringEnvs {
		if env.ExpiresAt == nil {
			continue
		}
		msg := fmt.Sprintf("环境将在 %d 分钟后过期", int(time.Until(*env.ExpiresAt).Minutes()))
		s.notifyEnvStatus(env.EnvID, env.Status, msg)
		if s.notifyRepo != nil {
			notification := &model.Notification{
				UserID:      env.UserID,
				SchoolID:    &env.SchoolID,
				Type:        model.NotifyTypeExperiment,
				Title:       "实验环境即将过期",
				Content:     msg,
				RelatedID:   &env.ExperimentID,
				RelatedType: "experiment_env",
			}
			_ = s.notifyRepo.Create(ctx, notification)
		}
	}
	return nil
}

func (s *EnvManagerService) SyncAllEnvStatus(ctx context.Context) error {
	envs, err := s.envRepo.ListByStatus(ctx, model.EnvStatusRunning)
	if err != nil {
		return fmt.Errorf("list running experiment envs: %w", err)
	}
	for index := range envs {
		s.syncEnvRuntimeStatus(ctx, &envs[index])
	}
	return nil
}

func (s *EnvManagerService) createRuntimeInstances(ctx context.Context, exp *model.Experiment, env *model.ExperimentEnv, blueprint model.ExperimentBlueprint, runtime model.ExperimentRuntimeState) error {
	for _, instance := range runtime.Instances {
		podConfig, err := buildPodConfigForRuntimeInstance(exp, env, blueprint, runtime, instance)
		if err != nil {
			return err
		}
		s.applyRuntimeImageDefaultResources(ctx, instance, podConfig)
		if err := s.provisioner.StartInstance(ctx, podConfig, 0); err != nil {
			return fmt.Errorf("create runtime pod %s: %w", instance.Key, err)
		}
	}
	return nil
}

func (s *EnvManagerService) applyRuntimeImageDefaultResources(ctx context.Context, instance model.ExperimentRuntimeTarget, podConfig *k8s.PodConfig) {
	if podConfig == nil || s.dockerImageRepo == nil {
		return
	}
	if instance.Kind != "service" && instance.Kind != "simulation" {
		return
	}

	image, err := s.resolveDockerImageByRef(ctx, podConfig.Image)
	if err != nil || image == nil {
		return
	}

	if cpu := strings.TrimSpace(fmt.Sprintf("%v", image.DefaultResources["cpu"])); cpu != "" && cpu != "<nil>" {
		podConfig.CPU = normalizeCPUValue(cpu)
	}
	if memory := strings.TrimSpace(fmt.Sprintf("%v", image.DefaultResources["memory"])); memory != "" && memory != "<nil>" {
		podConfig.Memory = normalizeBinaryUnit(memory, podConfig.Memory)
	}
	if storage := strings.TrimSpace(fmt.Sprintf("%v", image.DefaultResources["storage"])); storage != "" && storage != "<nil>" {
		podConfig.Storage = normalizeBinaryUnit(storage, podConfig.Storage)
	}
}

func (s *EnvManagerService) resolveDockerImageByRef(ctx context.Context, imageRef string) (*model.DockerImage, error) {
	ref := strings.ToLower(strings.TrimSpace(imageRef))
	if ref == "" {
		return nil, nil
	}

	candidates := []string{ref}
	if slashIndex, colonIndex := strings.LastIndex(ref, "/"), strings.LastIndex(ref, ":"); colonIndex > slashIndex {
		candidates = append(candidates, ref[:colonIndex])
	}

	seen := map[string]struct{}{}
	for _, candidate := range candidates {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}
		if _, exists := seen[candidate]; exists {
			continue
		}
		seen[candidate] = struct{}{}

		image, err := s.dockerImageRepo.GetByName(ctx, candidate)
		if err == nil && image != nil {
			return image, nil
		}
	}

	return nil, nil
}

func (s *EnvManagerService) waitForRuntimeReady(ctx context.Context, env *model.ExperimentEnv, blueprint model.ExperimentBlueprint, runtime model.ExperimentRuntimeState) {
	for _, instance := range runtime.Instances {
		if err := s.k8sClient.WaitForPodReady(ctx, instance.PodName, 5*time.Minute); err != nil {
			logger.Error("experiment runtime instance failed to become ready", zap.String("env_id", env.EnvID), zap.String("instance", instance.Key), zap.Error(err))
			markRuntimeInstanceStatuses(env, runtime, model.EnvStatusFailed)
			env.Status = model.EnvStatusFailed
			env.ErrorMessage = fmt.Sprintf("runtime instance %s failed to become ready: %v", instance.Key, err)
			_ = s.envRepo.Update(ctx, env)
			s.notifyEnvStatus(env.EnvID, model.EnvStatusFailed, "experiment runtime startup failed")
			return
		}
	}

	markRuntimeInstanceStatuses(env, runtime, model.EnvStatusRunning)

	if err := s.applyExperimentContent(ctx, blueprint, runtime); err != nil {
		logger.Warn("apply experiment content failed", zap.String("env_id", env.EnvID), zap.Error(err))
	}

	env.Status = model.EnvStatusRunning
	now := time.Now()
	env.StartedAt = &now
	if err := s.envRepo.Update(ctx, env); err != nil {
		logger.Error("update experiment env after ready failed", zap.String("env_id", env.EnvID), zap.Error(err))
		return
	}
	s.syncSessionStatus(ctx, env, model.EnvStatusRunning)

	s.notifyEnvStatus(env.EnvID, model.EnvStatusRunning, "experiment runtime ready")
}

func (s *EnvManagerService) terminateEnv(ctx context.Context, env *model.ExperimentEnv) error {
	runtime := buildRuntimeStateFromEnv(env)
	if err := s.deleteRuntimeInstances(ctx, runtime); err != nil {
		logger.Warn("delete runtime instances failed", zap.String("env_id", env.EnvID), zap.Error(err))
	}
	if err := s.envRepo.UpdateStatus(ctx, env.EnvID, model.EnvStatusTerminated); err != nil {
		return errors.ErrDatabaseError.WithError(err)
	}
	s.syncSessionStatus(ctx, env, model.EnvStatusTerminated)
	s.notifyEnvStatus(env.EnvID, model.EnvStatusTerminated, "experiment terminated")
	return nil
}

func (s *EnvManagerService) deleteRuntimeInstances(ctx context.Context, runtime model.ExperimentRuntimeState) error {
	var firstErr error
	for _, instance := range runtime.Instances {
		if instance.PodName == "" {
			continue
		}
		if err := s.provisioner.StopInstance(ctx, instance.PodName); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (s *EnvManagerService) syncEnvRuntimeStatus(ctx context.Context, env *model.ExperimentEnv) {
	runtime := buildRuntimeStateFromEnv(env)
	if len(runtime.Instances) == 0 {
		return
	}

	statuses := make([]string, 0, len(runtime.Instances))
	instanceStatusMap := make(map[string]string, len(runtime.Instances))
	for _, instance := range runtime.Instances {
		status, err := s.k8sClient.GetPodStatus(ctx, instance.PodName)
		if err != nil {
			statuses = append(statuses, "Missing")
			instanceStatusMap[instance.Key] = model.EnvStatusFailed
			continue
		}
		statuses = append(statuses, status)
		switch status {
		case "Running":
			instanceStatusMap[instance.Key] = model.EnvStatusRunning
		case "Pending":
			instanceStatusMap[instance.Key] = model.EnvStatusCreating
		default:
			instanceStatusMap[instance.Key] = model.EnvStatusFailed
		}
	}

	newStatus := mapRuntimeStatuses(statuses)
	updatedInstance := false
	for index := range env.RuntimeInstances {
		if nextStatus, ok := instanceStatusMap[env.RuntimeInstances[index].InstanceKey]; ok && env.RuntimeInstances[index].Status != nextStatus {
			env.RuntimeInstances[index].Status = nextStatus
			updatedInstance = true
		}
	}
	if newStatus != "" && env.Status != newStatus {
		env.Status = newStatus
		_ = s.envRepo.Update(ctx, env)
		return
	}
	if updatedInstance {
		_ = s.envRepo.Update(ctx, env)
	}
}

func (s *EnvManagerService) buildEnvInfo(env *model.ExperimentEnv) *EnvInfo {
	return &EnvInfo{
		EnvID:     env.EnvID,
		Status:    env.Status,
		AccessURL: fmt.Sprintf("/api/v1/envs/%s/proxy", env.EnvID),
		ExpiresAt: env.ExpiresAt,
	}
}

func (s *EnvManagerService) syncSessionStatus(ctx context.Context, env *model.ExperimentEnv, status string) {
	if s.sessionRepo == nil || env == nil || env.SessionID == nil {
		return
	}
	session, err := s.sessionRepo.GetByID(ctx, *env.SessionID)
	if err != nil {
		return
	}
	memberCount := countJoinedSessionMembers(session)
	var startedAt *time.Time
	if status == model.EnvStatusRunning {
		now := time.Now()
		startedAt = &now
	}
	_ = s.sessionRepo.UpdateCounters(ctx, session.ID, memberCount, session.PrimaryEnvID, status, startedAt, env.ExpiresAt)
	if s.wsHub != nil {
		_ = s.wsHub.BroadcastToRoom("session:"+session.SessionKey, websocket.MessageTypeSessionUpdate, map[string]interface{}{
			"session_key":          session.SessionKey,
			"status":               status,
			"current_member_count": memberCount,
			"primary_env_id":       session.PrimaryEnvID,
		})
	}
}

func (s *EnvManagerService) handleSessionDetach(ctx context.Context, env *model.ExperimentEnv) {
	if s.sessionRepo == nil || env == nil || env.SessionID == nil {
		return
	}
	s.markSessionMemberStatus(ctx, *env.SessionID, env.UserID, "left")
	session, err := s.sessionRepo.GetByID(ctx, *env.SessionID)
	if err != nil {
		return
	}
	activeCount, err := s.envRepo.CountActiveBySession(ctx, session.ID)
	if err != nil {
		return
	}
	primaryEnvID := session.PrimaryEnvID
	if primaryEnvID == env.EnvID {
		if replacement, replacementErr := s.envRepo.GetFirstActiveBySession(ctx, session.ID, env.EnvID); replacementErr == nil && replacement != nil {
			primaryEnvID = replacement.EnvID
		}
	}
	memberCount := countJoinedSessionMembers(session)
	if memberCount == 0 {
		memberCount = int(activeCount)
	}
	_ = s.sessionRepo.UpdateCounters(ctx, session.ID, memberCount, primaryEnvID, model.EnvStatusRunning, nil, env.ExpiresAt)
}

func (s *EnvManagerService) ensureSessionForEnv(
	ctx context.Context,
	exp *model.Experiment,
	req *EnvCreateRequest,
	blueprint model.ExperimentBlueprint,
	envID string,
	expiresAt time.Time,
	sessionKey string,
) (*model.ExperimentSession, error) {
	if s.sessionRepo == nil {
		return nil, nil
	}

	if sessionKey == "" {
		sessionKey = envID
	}
	session := &model.ExperimentSession{
		SessionKey:         sessionKey,
		ExperimentID:       exp.ID,
		SchoolID:           req.SchoolID,
		Mode:               blueprint.Mode,
		Status:             model.EnvStatusCreating,
		PrimaryEnvID:       envID,
		MaxMembers:         maxCollaborationMembers(blueprint),
		CurrentMemberCount: 0,
		ExpiresAt:          &expiresAt,
	}
	if blueprint.Mode != model.ExperimentModeCollaboration {
		session.MaxMembers = 1
	}
	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}
	return session, nil
}

func (s *EnvManagerService) bindEnvToSession(ctx context.Context, session *model.ExperimentSession, env *model.ExperimentEnv, userID uint, blueprint model.ExperimentBlueprint) error {
	if session == nil || env == nil {
		return nil
	}
	if s.sessionMemberRepo != nil {
		if member, err := s.sessionMemberRepo.GetBySessionAndUser(ctx, session.ID, userID); err != nil {
			if err == gorm.ErrRecordNotFound {
				member := &model.ExperimentSessionMember{
					SessionID:       session.ID,
					UserID:          userID,
					RoleKey:         resolveSessionRoleKey(session, blueprint, userID),
					AssignedNodeKey: resolveAssignedNodeKey(session, blueprint, userID),
					JoinStatus:      "joined",
				}
				if createErr := s.sessionMemberRepo.Create(ctx, member); createErr != nil {
					return errors.ErrDatabaseError.WithError(createErr)
				}
			} else {
				return errors.ErrDatabaseError.WithError(err)
			}
		} else {
			member.RoleKey = resolveSessionRoleKey(session, blueprint, userID)
			member.AssignedNodeKey = resolveAssignedNodeKey(session, blueprint, userID)
			member.JoinStatus = "joined"
			member.JoinedAt = time.Now()
			if updateErr := s.sessionMemberRepo.Update(ctx, member); updateErr != nil {
				return errors.ErrDatabaseError.WithError(updateErr)
			}
		}
	}
	updatedSession, err := s.sessionRepo.GetByID(ctx, session.ID)
	if err != nil {
		return errors.ErrDatabaseError.WithError(err)
	}
	memberCount := countJoinedSessionMembers(updatedSession)
	now := time.Now()
	if err := s.sessionRepo.UpdateCounters(ctx, session.ID, memberCount, updatedSession.PrimaryEnvID, env.Status, &now, env.ExpiresAt); err != nil {
		return errors.ErrDatabaseError.WithError(err)
	}
	return nil
}

func (s *EnvManagerService) markSessionMemberStatus(ctx context.Context, sessionID, userID uint, joinStatus string) {
	if s.sessionMemberRepo == nil || sessionID == 0 || userID == 0 {
		return
	}
	member, err := s.sessionMemberRepo.GetBySessionAndUser(ctx, sessionID, userID)
	if err != nil || member == nil {
		return
	}
	if member.JoinStatus == joinStatus {
		return
	}
	member.JoinStatus = joinStatus
	_ = s.sessionMemberRepo.Update(ctx, member)
}

func (s *EnvManagerService) notifyEnvStatus(envID, status, message string) {
	if s.wsHub != nil {
		s.wsHub.BroadcastToRoom("env:"+envID, websocket.MessageTypeEnvStatus, map[string]interface{}{
			"env_id":  envID,
			"status":  status,
			"message": message,
		})
	}
}

func (s *EnvManagerService) saveEnvSnapshot(ctx context.Context, env *model.ExperimentEnv) (string, error) {
	runtime := buildRuntimeStateFromEnv(env)
	primaryPod := primaryPodName(runtime)
	if primaryPod == "" {
		return "", fmt.Errorf("primary runtime instance not found")
	}

	_, err := s.k8sClient.ExecCommand(ctx, primaryPod, []string{"tar", "-czf", "/tmp/snapshot.tar.gz", "-C", "/workspace", "."})
	if err != nil {
		return "", fmt.Errorf("package workspace snapshot: %w", err)
	}
	base64Data, err := s.k8sClient.ExecCommand(ctx, primaryPod, []string{"base64", "-w0", "/tmp/snapshot.tar.gz"})
	if err != nil {
		return "", fmt.Errorf("read workspace snapshot: %w", err)
	}
	snapshotBytes, err := base64.StdEncoding.DecodeString(strings.TrimSpace(base64Data))
	if err != nil {
		return "", fmt.Errorf("decode workspace snapshot: %w", err)
	}

	objectPath := fmt.Sprintf("snapshots/%s/%d.tar.gz", env.EnvID, time.Now().Unix())
	if s.uploadService == nil {
		return objectPath, nil
	}
	url, err := s.uploadService.Upload(ctx, "snapshots", objectPath, snapshotBytes)
	if err != nil {
		return "", fmt.Errorf("upload workspace snapshot: %w", err)
	}
	_, _ = s.k8sClient.ExecCommand(ctx, primaryPod, []string{"rm", "-f", "/tmp/snapshot.tar.gz"})
	return url, nil
}

func (s *EnvManagerService) persistSnapshotMetadata(ctx context.Context, env *model.ExperimentEnv, snapshotURL string) error {
	now := time.Now()
	env.SnapshotURL = snapshotURL
	env.SnapshotAt = &now
	return s.envRepo.Update(ctx, env)
}

func (s *EnvManagerService) applyExperimentContent(ctx context.Context, blueprint model.ExperimentBlueprint, runtime model.ExperimentRuntimeState) error {
	if len(blueprint.Content.Assets) == 0 && len(blueprint.Content.InitScripts) == 0 && len(blueprint.Workspace.InitScripts) == 0 && !hasNodeInitScripts(blueprint.Nodes) {
		return nil
	}

	primaryPod := primaryPodName(runtime)
	if primaryPod == "" {
		return fmt.Errorf("primary runtime instance not found")
	}

	for _, asset := range blueprint.Content.Assets {
		if s.uploadService == nil || asset.Bucket == "" || asset.ObjectPath == "" {
			continue
		}
		reader, err := s.uploadService.GetObject(ctx, asset.Bucket, asset.ObjectPath)
		if err != nil {
			if asset.Required {
				return err
			}
			continue
		}
		data, readErr := io.ReadAll(reader)
		_ = reader.Close()
		if readErr != nil {
			if asset.Required {
				return readErr
			}
			continue
		}

		targetPods, targetErr := resolveAssetTargetPods(runtime, asset.Target)
		if targetErr != nil {
			if asset.Required {
				return targetErr
			}
			continue
		}
		targetPath := resolveAssetMountPath(asset)
		for _, podName := range targetPods {
			if err := s.writeBytesToPod(ctx, podName, targetPath, data); err != nil && asset.Required {
				return err
			}
		}
	}

	for _, script := range blueprint.Content.InitScripts {
		if strings.TrimSpace(script) == "" {
			continue
		}
		if _, err := s.k8sClient.ExecCommand(ctx, primaryPod, []string{"sh", "-lc", script}); err != nil {
			return err
		}
	}

	for _, script := range blueprint.Workspace.InitScripts {
		if strings.TrimSpace(script) == "" {
			continue
		}
		if _, err := s.k8sClient.ExecCommand(ctx, primaryPod, []string{"sh", "-lc", script}); err != nil {
			return err
		}
	}

	for _, node := range blueprint.Nodes {
		if len(node.InitScripts) == 0 {
			continue
		}
		targetPod := runtimePodName(runtime, node.Key)
		if targetPod == "" {
			return fmt.Errorf("runtime node %s pod not found", node.Key)
		}
		for _, script := range node.InitScripts {
			if strings.TrimSpace(script) == "" {
				continue
			}
			if _, err := s.k8sClient.ExecCommand(ctx, targetPod, []string{"sh", "-lc", script}); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *EnvManagerService) writeBytesToPod(ctx context.Context, podName, targetPath string, data []byte) error {
	encoded := base64.StdEncoding.EncodeToString(data)
	command := strings.Join([]string{
		`TARGET="$1"`,
		`mkdir -p "$(dirname "$TARGET")"`,
		`cat <<'EOF' | base64 -d > "$TARGET"`,
		encoded,
		`EOF`,
	}, "\n")
	_, err := s.k8sClient.ExecCommand(ctx, podName, []string{"sh", "-lc", command, "sh", targetPath})
	return err
}

func buildExperimentRuntimeState(envID string, blueprint model.ExperimentBlueprint) model.ExperimentRuntimeState {
	instances := make([]model.ExperimentRuntimeTarget, 0, 1+len(blueprint.Nodes)+len(blueprint.Services))
	toolTargets := map[string]model.RuntimeToolRef{}

	workspacePorts := collectPortsForTools(blueprint.Workspace.InteractionTools)
	workspace := model.ExperimentRuntimeTarget{
		Key:              "workspace",
		Kind:             "workspace",
		PodName:          envID,
		Status:           model.EnvStatusPending,
		Ports:            workspacePorts,
		StudentFacing:    true,
		InteractionTools: blueprint.Workspace.InteractionTools,
	}
	instances = append(instances, workspace)
	for _, tool := range blueprint.Workspace.InteractionTools {
		if port := defaultPortForTool(tool); port > 0 {
			toolTargets[tool] = model.RuntimeToolRef{InstanceKey: workspace.Key, Port: port}
		}
	}

	for _, node := range blueprint.Nodes {
		instance := model.ExperimentRuntimeTarget{
			Key:              node.Key,
			Kind:             "node",
			PodName:          envID + "-" + sanitizeRuntimeKey(node.Key),
			Status:           model.EnvStatusPending,
			Ports:            resolveRuntimePortsForImage(node.Image, node.Ports, node.InteractionTools),
			StudentFacing:    node.StudentFacing,
			InteractionTools: node.InteractionTools,
		}
		instances = append(instances, instance)
		for _, tool := range node.InteractionTools {
			if _, exists := toolTargets[tool]; exists {
				continue
			}
			if port := portForTargetTool(tool, instance.Ports); port > 0 {
				toolTargets[tool] = model.RuntimeToolRef{InstanceKey: instance.Key, Port: port}
			}
		}
	}

	rpcTargets := resolveExperimentRuntimeRPCTargets(envID, blueprint)
	for _, serviceSpec := range blueprint.Services {
		serviceInstances := buildExperimentServiceRuntimeTargets(envID, serviceSpec, rpcTargets)
		for index, instance := range serviceInstances {
			instances = append(instances, instance)
			if index != 0 {
				continue
			}
			if toolKey, ok := serviceToolByKey[serviceSpec.Key]; ok {
				if port := portForTargetTool(toolKey, instance.Ports); port > 0 {
					toolTargets[toolKey] = model.RuntimeToolRef{InstanceKey: instance.Key, Port: port}
				}
			}
		}
	}

	appendBuiltInToolInstance := func(toolKey, instanceKey, instanceKind, image string, port int32) {
		if existing, exists := toolTargets[toolKey]; exists {
			if (toolKey != "explorer" && toolKey != "visualization") || existing.InstanceKey != "workspace" {
				return
			}
		}
		if !experimentBlueprintUsesTool(blueprint, toolKey) {
			return
		}

		instance := model.ExperimentRuntimeTarget{
			Key:              instanceKey,
			Kind:             instanceKind,
			PodName:          envID + "-" + sanitizeRuntimeKey(instanceKey),
			Status:           model.EnvStatusPending,
			Ports:            []int32{port},
			StudentFacing:    true,
			InteractionTools: []string{toolKey},
			EnvVars: map[string]string{
				"TOOL_KEY": toolKey,
			},
		}
		if image != "" {
			instance.EnvVars["RUNTIME_IMAGE"] = image
		}
		instances = append(instances, instance)
		toolTargets[toolKey] = model.RuntimeToolRef{InstanceKey: instance.Key, Port: port}
	}

	appendBuiltInToolInstance("visualization", "simulation", "simulation", resolveRuntimeImage("", "simulation", "visualization"), 8080)

	for _, tool := range blueprint.Tools {
		if _, exists := toolTargets[tool.Key]; exists {
			continue
		}
		targetPorts := []int32{}
		targetIndex := -1
		for _, instance := range instances {
			if instance.Key != tool.Target {
				continue
			}
			targetPorts = instance.Ports
			targetIndex = indexOfRuntimeInstance(instances, instance.Key)
			break
		}
		if port := portForTargetTool(tool.Key, targetPorts); port > 0 && tool.Target != "" {
			toolTargets[tool.Key] = model.RuntimeToolRef{InstanceKey: tool.Target, Port: port}
			if targetIndex >= 0 {
				instances[targetIndex].Ports = ensurePortOnRuntimeInstance(instances[targetIndex].Ports, port)
				instances[targetIndex].InteractionTools = ensureToolOnRuntimeInstance(instances[targetIndex].InteractionTools, tool.Key)
			}
		}
	}

	return model.ExperimentRuntimeState{
		SessionMode:        blueprint.Mode,
		PrimaryInstanceKey: workspace.Key,
		Instances:          instances,
		ToolTargets:        toolTargets,
	}
}

type experimentRuntimeRPCTargets struct {
	HTTPURL string
	WSURL   string
	ChainID string
}

func buildExperimentServiceRuntimeTargets(
	envID string,
	serviceSpec model.ExperimentServiceBlueprint,
	rpcTargets experimentRuntimeRPCTargets,
) []model.ExperimentRuntimeTarget {
	serviceTools := make([]string, 0, 1)
	if toolKey, ok := serviceToolByKey[serviceSpec.Key]; ok {
		serviceTools = append(serviceTools, toolKey)
	}
	ports := resolveRuntimePortsForImage(serviceSpec.Image, serviceSpec.Ports, serviceTools)
	mainEnvVars := cloneRuntimeEnvVars(serviceSpec.EnvVars)
	main := model.ExperimentRuntimeTarget{
		Key:              serviceSpec.Key,
		Kind:             "service",
		PodName:          experimentServicePodName(envID, serviceSpec.Key),
		Status:           model.EnvStatusPending,
		Ports:            ports,
		StudentFacing:    serviceSpec.StudentFacing,
		InteractionTools: serviceTools,
		EnvVars:          mainEnvVars,
	}

	instances := []model.ExperimentRuntimeTarget{main}
	switch serviceSpec.Key {
	case "blockscout":
		dbName := sanitizedExperimentServiceDBName(serviceSpec.Key, "blockscout")
		dbPassword := dbName + "-chainspace"
		if _, ok := mainEnvVars["DATABASE_URL"]; !ok {
			mainEnvVars["DATABASE_URL"] = fmt.Sprintf(
				"postgresql://%s:%s@%s:5432/%s",
				dbName,
				dbPassword,
				experimentServiceHost(envID, serviceSpec.Key+"-db"),
				dbName,
			)
		}
		if _, ok := mainEnvVars["ECTO_USE_SSL"]; !ok {
			mainEnvVars["ECTO_USE_SSL"] = "false"
		}
		if _, ok := mainEnvVars["SECRET_KEY_BASE"]; !ok {
			mainEnvVars["SECRET_KEY_BASE"] = "chainspace-blockscout-secret-key-base"
		}
		if _, ok := mainEnvVars["ETHEREUM_JSONRPC_VARIANT"]; !ok {
			mainEnvVars["ETHEREUM_JSONRPC_VARIANT"] = "geth"
		}
		if _, ok := mainEnvVars["PORT"]; !ok {
			mainEnvVars["PORT"] = strconv.Itoa(int(portForTargetTool("explorer", ports)))
		}
		if rpcTargets.HTTPURL != "" {
			if _, ok := mainEnvVars["ETHEREUM_JSONRPC_HTTP_URL"]; !ok {
				mainEnvVars["ETHEREUM_JSONRPC_HTTP_URL"] = rpcTargets.HTTPURL
			}
			if _, ok := mainEnvVars["ETHEREUM_JSONRPC_TRACE_URL"]; !ok {
				mainEnvVars["ETHEREUM_JSONRPC_TRACE_URL"] = rpcTargets.HTTPURL
			}
		}
		if rpcTargets.WSURL != "" {
			if _, ok := mainEnvVars["ETHEREUM_JSONRPC_WS_URL"]; !ok {
				mainEnvVars["ETHEREUM_JSONRPC_WS_URL"] = rpcTargets.WSURL
			}
		}
		if _, ok := mainEnvVars["NETWORK"]; !ok {
			mainEnvVars["NETWORK"] = "ChainSpace"
		}
		if _, ok := mainEnvVars["SUBNETWORK"]; !ok {
			mainEnvVars["SUBNETWORK"] = "Experiment"
		}
		if _, ok := mainEnvVars["COIN"]; !ok {
			mainEnvVars["COIN"] = "ETH"
		}
		instances = append(instances, buildExperimentPostgresRuntimeTarget(envID, serviceSpec.Key+"-db", dbName, dbPassword))
	case "chainlink":
		dbName := sanitizedExperimentServiceDBName(serviceSpec.Key, "chainlink")
		dbPassword := dbName + "-chainspace"
		if _, ok := mainEnvVars["CHAINLINK_DATABASE_URL"]; !ok {
			mainEnvVars["CHAINLINK_DATABASE_URL"] = fmt.Sprintf(
				"postgresql://%s:%s@%s:5432/%s?sslmode=disable",
				dbName,
				dbPassword,
				experimentServiceHost(envID, serviceSpec.Key+"-db"),
				dbName,
			)
		}
		if _, ok := mainEnvVars["CHAINLINK_HTTP_PORT"]; !ok {
			mainEnvVars["CHAINLINK_HTTP_PORT"] = strconv.Itoa(int(portForTargetTool("api_debug", ports)))
		}
		if _, ok := mainEnvVars["CHAINLINK_P2P_PORT"]; !ok {
			mainEnvVars["CHAINLINK_P2P_PORT"] = strconv.Itoa(int(preferredPortOrDefault(ports, 6689, 6689)))
		}
		if _, ok := mainEnvVars["CHAINLINK_EVM_CHAIN_ID"]; !ok {
			mainEnvVars["CHAINLINK_EVM_CHAIN_ID"] = rpcTargets.ChainID
		}
		if _, ok := mainEnvVars["CHAINLINK_API_EMAIL"]; !ok {
			mainEnvVars["CHAINLINK_API_EMAIL"] = "admin@chainspace.local"
		}
		if _, ok := mainEnvVars["CHAINLINK_API_PASSWORD"]; !ok {
			mainEnvVars["CHAINLINK_API_PASSWORD"] = "ChainspaceAdmin123!"
		}
		if _, ok := mainEnvVars["CHAINLINK_KEYSTORE_PASSWORD"]; !ok {
			mainEnvVars["CHAINLINK_KEYSTORE_PASSWORD"] = "ChainspaceKeystore123!"
		}
		if rpcTargets.HTTPURL != "" {
			if _, ok := mainEnvVars["CHAINLINK_EVM_HTTP_URL"]; !ok {
				mainEnvVars["CHAINLINK_EVM_HTTP_URL"] = rpcTargets.HTTPURL
			}
		}
		if rpcTargets.WSURL != "" {
			if _, ok := mainEnvVars["CHAINLINK_EVM_WS_URL"]; !ok {
				mainEnvVars["CHAINLINK_EVM_WS_URL"] = rpcTargets.WSURL
			}
		}
		instances = append(instances, buildExperimentPostgresRuntimeTarget(envID, serviceSpec.Key+"-db", dbName, dbPassword))
	case "thegraph":
		dbName := sanitizedExperimentServiceDBName(serviceSpec.Key, "graph")
		dbPassword := dbName + "-chainspace"
		if _, ok := mainEnvVars["postgres_host"]; !ok {
			mainEnvVars["postgres_host"] = experimentServiceHost(envID, serviceSpec.Key+"-db")
		}
		if _, ok := mainEnvVars["postgres_port"]; !ok {
			mainEnvVars["postgres_port"] = "5432"
		}
		if _, ok := mainEnvVars["postgres_user"]; !ok {
			mainEnvVars["postgres_user"] = dbName
		}
		if _, ok := mainEnvVars["postgres_pass"]; !ok {
			mainEnvVars["postgres_pass"] = dbPassword
		}
		if _, ok := mainEnvVars["postgres_db"]; !ok {
			mainEnvVars["postgres_db"] = dbName
		}
		if _, ok := mainEnvVars["ipfs"]; !ok {
			mainEnvVars["ipfs"] = experimentServiceHost(envID, serviceSpec.Key+"-ipfs") + ":5001"
		}
		if rpcTargets.HTTPURL != "" {
			if _, ok := mainEnvVars["ethereum"]; !ok {
				mainEnvVars["ethereum"] = "mainnet:" + rpcTargets.HTTPURL
			}
		}
		if _, ok := mainEnvVars["GRAPH_LOG"]; !ok {
			mainEnvVars["GRAPH_LOG"] = "info"
		}
		instances = append(
			instances,
			buildExperimentPostgresRuntimeTarget(envID, serviceSpec.Key+"-db", dbName, dbPassword),
			model.ExperimentRuntimeTarget{
				Key:           serviceSpec.Key + "-ipfs",
				Kind:          "service",
				PodName:       experimentServicePodName(envID, serviceSpec.Key+"-ipfs"),
				Status:        model.EnvStatusPending,
				Ports:         []int32{5001, 8080, 4001},
				StudentFacing: false,
				EnvVars: map[string]string{
					"RUNTIME_IMAGE": resolveRuntimeImage("", "ipfs"),
				},
			},
		)
	}

	instances[0].EnvVars = mainEnvVars
	return instances
}

func buildExperimentPostgresRuntimeTarget(envID, key, dbName, dbPassword string) model.ExperimentRuntimeTarget {
	return model.ExperimentRuntimeTarget{
		Key:           key,
		Kind:          "service",
		PodName:       experimentServicePodName(envID, key),
		Status:        model.EnvStatusPending,
		Ports:         []int32{5432},
		StudentFacing: false,
		EnvVars: map[string]string{
			"POSTGRES_DB":       dbName,
			"POSTGRES_USER":     dbName,
			"POSTGRES_PASSWORD": dbPassword,
			"RUNTIME_IMAGE":     "postgres:16-alpine",
		},
	}
}

func resolveExperimentRuntimeRPCTargets(envID string, blueprint model.ExperimentBlueprint) experimentRuntimeRPCTargets {
	targets := experimentRuntimeRPCTargets{
		ChainID: "31337",
	}

	for _, serviceSpec := range blueprint.Services {
		if !targetSupportsTool(blueprint, serviceSpec.Key, "rpc") {
			continue
		}
		ports := resolveRuntimePortsForImage(serviceSpec.Image, serviceSpec.Ports, []string{"rpc"})
		httpPort := portForTargetTool("rpc", ports)
		if httpPort <= 0 {
			httpPort = 8545
		}
		targets.HTTPURL = fmt.Sprintf("http://%s:%d", experimentServiceHost(envID, serviceSpec.Key), httpPort)
		targets.WSURL = fmt.Sprintf("ws://%s:%d", experimentServiceHost(envID, serviceSpec.Key), preferredPortOrDefault(ports, 8546, 8546))
		if chainID := strings.TrimSpace(serviceSpec.EnvVars["CHAIN_ID"]); chainID != "" {
			targets.ChainID = chainID
		}
		return targets
	}

	for _, node := range blueprint.Nodes {
		if !targetSupportsTool(blueprint, node.Key, "rpc") {
			continue
		}
		ports := resolveRuntimePortsForImage(node.Image, node.Ports, node.InteractionTools)
		httpPort := portForTargetTool("rpc", ports)
		if httpPort <= 0 {
			httpPort = 8545
		}
		targets.HTTPURL = fmt.Sprintf("http://%s-svc:%d", envID+"-"+sanitizeRuntimeKey(node.Key), httpPort)
		targets.WSURL = fmt.Sprintf("ws://%s-svc:%d", envID+"-"+sanitizeRuntimeKey(node.Key), preferredPortOrDefault(ports, 8546, 8546))
		return targets
	}

	return targets
}

func experimentServicePodName(envID, key string) string {
	return envID + "-svc-" + sanitizeRuntimeKey(key)
}

func experimentServiceHost(envID, key string) string {
	return experimentServicePodName(envID, key) + "-svc"
}

func sanitizedExperimentServiceDBName(key, fallback string) string {
	value := strings.ReplaceAll(sanitizeRuntimeKey(key), "-", "_")
	if value == "" {
		return fallback
	}
	return value
}

func cloneRuntimeEnvVars(values map[string]string) map[string]string {
	if len(values) == 0 {
		return map[string]string{}
	}
	result := make(map[string]string, len(values))
	for key, value := range values {
		result[key] = value
	}
	return result
}

func buildPodConfigForRuntimeInstance(
	exp *model.Experiment,
	env *model.ExperimentEnv,
	blueprint model.ExperimentBlueprint,
	runtime model.ExperimentRuntimeState,
	instance model.ExperimentRuntimeTarget,
) (*k8s.PodConfig, error) {
	switch instance.Kind {
	case "workspace":
		boot := buildRuntimeBootConfig(blueprint.Workspace.Image, exp, env, blueprint, runtime, instance)
		envVars := map[string]string{
			"EXPERIMENT_TYPE":         exp.Type,
			"EXPERIMENT_MODE":         blueprint.Mode,
			"CHAINSPACE_RUNTIME_KIND": "workspace",
		}
		for key, value := range boot.EnvVars {
			envVars[key] = value
		}
		return &k8s.PodConfig{
			EnvID:        instance.PodName,
			UserID:       env.UserID,
			SchoolID:     env.SchoolID,
			ExperimentID: exp.ID,
			Image:        blueprint.Workspace.Image,
			Command:      boot.Command,
			Args:         boot.Args,
			CPU:          blueprint.Workspace.Resources.CPU,
			Memory:       blueprint.Workspace.Resources.Memory,
			Storage:      blueprint.Workspace.Resources.Storage,
			Timeout:      time.Until(*env.ExpiresAt),
			Ports:        instance.Ports,
			ProbePort:    boot.ProbePort,
			EnvVars:      envVars,
		}, nil
	case "node":
		node, found := findBlueprintNode(blueprint, instance.Key)
		if !found {
			return nil, fmt.Errorf("runtime node %s not found in blueprint", instance.Key)
		}
		boot := buildRuntimeBootConfig(node.Image, exp, env, blueprint, runtime, instance)
		envVars := map[string]string{
			"EXPERIMENT_TYPE": exp.Type,
			"EXPERIMENT_MODE": blueprint.Mode,
			"NODE_KEY":        node.Key,
			"NODE_ROLE":       node.Role,
		}
		for key, value := range boot.EnvVars {
			envVars[key] = value
		}
		return &k8s.PodConfig{
			EnvID:        instance.PodName,
			UserID:       env.UserID,
			SchoolID:     env.SchoolID,
			ExperimentID: exp.ID,
			Image:        node.Image,
			Command:      boot.Command,
			Args:         boot.Args,
			CPU:          node.Resources.CPU,
			Memory:       node.Resources.Memory,
			Storage:      node.Resources.Storage,
			Timeout:      time.Until(*env.ExpiresAt),
			Ports:        instance.Ports,
			ProbePort:    boot.ProbePort,
			EnvVars:      envVars,
		}, nil
	case "service":
		serviceSpec, found := findBlueprintService(blueprint, instance.Key)
		envVars := map[string]string{
			"EXPERIMENT_TYPE": exp.Type,
			"EXPERIMENT_MODE": blueprint.Mode,
			"SERVICE_KEY":     instance.Key,
		}
		image := ""
		if found {
			image = serviceSpec.Image
			envVars["SERVICE_KEY"] = serviceSpec.Key
			envVars["SERVICE_ROLE"] = serviceSpec.Role
			for key, value := range serviceSpec.EnvVars {
				envVars[key] = value
			}
		} else {
			image = strings.TrimSpace(instance.EnvVars["RUNTIME_IMAGE"])
			if image == "" {
				image = resolveRuntimeImage("", instance.Key)
			}
			envVars["SERVICE_ROLE"] = strings.TrimSpace(instance.EnvVars["SERVICE_ROLE"])
		}
		boot := buildRuntimeBootConfig(image, exp, env, blueprint, runtime, instance)
		for key, value := range boot.EnvVars {
			envVars[key] = value
		}
		for key, value := range instance.EnvVars {
			envVars[key] = value
		}
		return &k8s.PodConfig{
			EnvID:        instance.PodName,
			UserID:       env.UserID,
			SchoolID:     env.SchoolID,
			ExperimentID: exp.ID,
			Image:        image,
			Command:      boot.Command,
			Args:         boot.Args,
			CPU:          blueprint.Workspace.Resources.CPU,
			Memory:       blueprint.Workspace.Resources.Memory,
			Storage:      blueprint.Workspace.Resources.Storage,
			Timeout:      time.Until(*env.ExpiresAt),
			Ports:        instance.Ports,
			ProbePort:    boot.ProbePort,
			EnvVars:      envVars,
		}, nil
	case "simulation":
		boot := buildRuntimeBootConfig(resolveRuntimeImage("", "simulation", "visualization"), exp, env, blueprint, runtime, instance)
		envVars := map[string]string{
			"EXPERIMENT_TYPE": exp.Type,
			"EXPERIMENT_MODE": blueprint.Mode,
			"INSTANCE_KIND":   "simulation",
		}
		for key, value := range boot.EnvVars {
			envVars[key] = value
		}
		return &k8s.PodConfig{
			EnvID:        instance.PodName,
			UserID:       env.UserID,
			SchoolID:     env.SchoolID,
			ExperimentID: exp.ID,
			Image:        resolveRuntimeImage("", "simulation", "visualization"),
			Command:      boot.Command,
			Args:         boot.Args,
			CPU:          blueprint.Workspace.Resources.CPU,
			Memory:       blueprint.Workspace.Resources.Memory,
			Storage:      blueprint.Workspace.Resources.Storage,
			Timeout:      time.Until(*env.ExpiresAt),
			Ports:        instance.Ports,
			ProbePort:    boot.ProbePort,
			EnvVars:      envVars,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported runtime instance kind: %s", instance.Kind)
	}
}

func primaryPodName(runtime model.ExperimentRuntimeState) string {
	for _, instance := range runtime.Instances {
		if instance.Key == runtime.PrimaryInstanceKey {
			return instance.PodName
		}
	}
	return ""
}

func runtimePodName(runtime model.ExperimentRuntimeState, instanceKey string) string {
	for _, instance := range runtime.Instances {
		if instance.Key == instanceKey {
			return instance.PodName
		}
	}
	return ""
}

func hasNodeInitScripts(nodes []model.ExperimentNodeBlueprint) bool {
	for _, node := range nodes {
		if len(node.InitScripts) > 0 {
			return true
		}
	}
	return false
}

func experimentBlueprintUsesTool(blueprint model.ExperimentBlueprint, toolKey string) bool {
	for _, key := range blueprint.Workspace.InteractionTools {
		if key == toolKey {
			return true
		}
	}
	for _, node := range blueprint.Nodes {
		for _, key := range node.InteractionTools {
			if key == toolKey {
				return true
			}
		}
	}
	for _, tool := range blueprint.Tools {
		if tool.Key == toolKey {
			return true
		}
	}
	return false
}

func resolveAssetMountPath(asset model.ExperimentContentBlueprintAsset) string {
	targetPath := strings.TrimSpace(asset.MountPath)
	if targetPath == "" {
		targetPath = "/workspace"
	}
	if strings.HasSuffix(targetPath, "/") || targetPath == "/workspace" {
		fileName := asset.Name
		if fileName == "" {
			fileName = asset.Key
		}
		targetPath = path.Join(targetPath, fileName)
	}
	return targetPath
}

func resolveAssetTargetPods(runtime model.ExperimentRuntimeState, target string) ([]string, error) {
	target = strings.TrimSpace(target)
	if target == "" || target == "workspace" {
		primaryPod := primaryPodName(runtime)
		if primaryPod == "" {
			return nil, fmt.Errorf("primary runtime instance not found")
		}
		return []string{primaryPod}, nil
	}

	instances := make([]model.ExperimentRuntimeTarget, 0, len(runtime.Instances))
	switch target {
	case "all_instances":
		instances = append(instances, runtime.Instances...)
	case "all_nodes":
		for _, instance := range runtime.Instances {
			if instance.Kind == "node" {
				instances = append(instances, instance)
			}
		}
	case "all_student_nodes":
		for _, instance := range runtime.Instances {
			if instance.Kind == "node" && instance.StudentFacing {
				instances = append(instances, instance)
			}
		}
	case "student_facing":
		for _, instance := range runtime.Instances {
			if instance.StudentFacing || instance.Key == runtime.PrimaryInstanceKey {
				instances = append(instances, instance)
			}
		}
	default:
		instanceKey := target
		if strings.Contains(target, ":") {
			parts := strings.SplitN(target, ":", 2)
			instanceKey = parts[1]
		}
		podName := runtimePodName(runtime, instanceKey)
		if podName == "" {
			return nil, fmt.Errorf("runtime instance %s pod not found", target)
		}
		return []string{podName}, nil
	}

	if len(instances) == 0 {
		return nil, fmt.Errorf("no runtime instances matched asset target %s", target)
	}

	pods := make([]string, 0, len(instances))
	for _, instance := range instances {
		if instance.PodName == "" {
			continue
		}
		pods = append(pods, instance.PodName)
	}
	if len(pods) == 0 {
		return nil, fmt.Errorf("no runtime pods matched asset target %s", target)
	}
	return pods, nil
}

func markRuntimeInstanceStatuses(env *model.ExperimentEnv, runtime model.ExperimentRuntimeState, status string) {
	for index := range env.RuntimeInstances {
		env.RuntimeInstances[index].Status = status
	}
	for index := range runtime.Instances {
		runtime.Instances[index].Status = status
	}
}

func maxCollaborationMembers(blueprint model.ExperimentBlueprint) int {
	if blueprint.Collaboration.MaxMembers > 0 {
		return blueprint.Collaboration.MaxMembers
	}
	return 4
}

func countJoinedSessionMembers(session *model.ExperimentSession) int {
	if session == nil {
		return 0
	}
	count := 0
	for _, member := range session.Members {
		if strings.EqualFold(member.JoinStatus, "joined") {
			count++
		}
	}
	return count
}

func sessionMemberIndex(session *model.ExperimentSession, userID uint) int {
	if session == nil {
		return 0
	}
	for index, member := range session.Members {
		if member.UserID == userID {
			return index
		}
	}
	return len(session.Members)
}

func sessionIDPtr(session *model.ExperimentSession) *uint {
	if session == nil {
		return nil
	}
	return &session.ID
}

func resolveSessionRoleKey(session *model.ExperimentSession, blueprint model.ExperimentBlueprint, userID uint) string {
	if session == nil || session.Mode != model.ExperimentModeCollaboration {
		return "owner"
	}
	memberIndex := sessionMemberIndex(session, userID)
	if memberIndex < len(blueprint.Collaboration.Roles) {
		return blueprint.Collaboration.Roles[memberIndex].Key
	}
	return fmt.Sprintf("member-%d", userID)
}

func resolveAssignedNodeKey(session *model.ExperimentSession, blueprint model.ExperimentBlueprint, userID uint) string {
	if session == nil || session.Mode != model.ExperimentModeCollaboration {
		return ""
	}
	memberIndex := sessionMemberIndex(session, userID)
	if memberIndex < len(blueprint.Collaboration.Roles) && len(blueprint.Collaboration.Roles[memberIndex].NodeKeys) > 0 {
		return blueprint.Collaboration.Roles[memberIndex].NodeKeys[0]
	}
	if memberIndex < len(blueprint.Nodes) {
		return blueprint.Nodes[memberIndex].Key
	}
	return ""
}

func mapRuntimeStatuses(statuses []string) string {
	hasPending := false
	hasFailed := false
	hasRunning := false
	for _, status := range statuses {
		switch status {
		case "Running":
			hasRunning = true
		case "Pending":
			hasPending = true
		case "Failed", "Unknown", "Missing":
			hasFailed = true
		}
	}
	switch {
	case hasFailed:
		return model.EnvStatusFailed
	case hasPending:
		return model.EnvStatusCreating
	case hasRunning:
		return model.EnvStatusRunning
	default:
		return model.EnvStatusTerminated
	}
}

func collectPortsForTools(tools []string) []int32 {
	seen := map[int32]struct{}{}
	ports := make([]int32, 0, len(tools))
	for _, tool := range tools {
		if port := defaultPortForTool(tool); port > 0 {
			if _, exists := seen[port]; exists {
				continue
			}
			seen[port] = struct{}{}
			ports = append(ports, port)
		}
	}
	return ports
}

func normalizeRuntimePorts(ports []int32, tools []string) []int32 {
	if len(ports) > 0 {
		return ports
	}
	return collectPortsForTools(tools)
}

func defaultPortForTool(tool string) int32 {
	switch tool {
	case "ide":
		return 8443
	case "terminal", "files", "logs", "network":
		return 7681
	case "rpc":
		return 8545
	case "api_debug":
		return 6688
	case "visualization":
		return 8080
	case "explorer":
		return 4000
	default:
		return 0
	}
}

func ensurePortOnRuntimeInstance(ports []int32, port int32) []int32 {
	for _, item := range ports {
		if item == port {
			return ports
		}
	}
	return append([]int32{port}, ports...)
}

func ensureToolOnRuntimeInstance(tools []string, tool string) []string {
	for _, item := range tools {
		if item == tool {
			return tools
		}
	}
	return append([]string{tool}, tools...)
}

func indexOfRuntimeInstance(instances []model.ExperimentRuntimeTarget, key string) int {
	for index := range instances {
		if instances[index].Key == key {
			return index
		}
	}
	return -1
}

func portForTargetTool(tool string, ports []int32) int32 {
	expected := defaultPortForTool(tool)
	if expected == 0 {
		return 0
	}
	for _, port := range ports {
		if port == expected {
			return port
		}
	}
	if len(ports) > 0 {
		return ports[0]
	}
	return expected
}

func findBlueprintNode(blueprint model.ExperimentBlueprint, key string) (model.ExperimentNodeBlueprint, bool) {
	for _, node := range blueprint.Nodes {
		if node.Key == key {
			return node, true
		}
	}
	return model.ExperimentNodeBlueprint{}, false
}

func findBlueprintService(blueprint model.ExperimentBlueprint, key string) (model.ExperimentServiceBlueprint, bool) {
	for _, serviceSpec := range blueprint.Services {
		if serviceSpec.Key == key {
			return serviceSpec, true
		}
	}
	return model.ExperimentServiceBlueprint{}, false
}

func sanitizeRuntimeKey(key string) string {
	key = strings.ToLower(strings.TrimSpace(key))
	key = strings.ReplaceAll(key, "_", "-")
	key = strings.ReplaceAll(key, " ", "-")
	return key
}
