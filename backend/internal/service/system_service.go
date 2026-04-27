package service

import (
	"context"
	"time"

	"github.com/chainspace/backend/internal/dto/request"
	"github.com/chainspace/backend/internal/dto/response"
	"github.com/chainspace/backend/internal/model"
	"github.com/chainspace/backend/internal/repository"
	"github.com/chainspace/backend/pkg/errors"
)

// SystemService 负责系统配置、操作日志、跨校申请与系统级聚合入口。
type SystemService struct {
	configRepo         *repository.SystemConfigRepository
	challengeRepo      *repository.ChallengeRepository
	operationLogRepo   *repository.OperationLogRepository
	configService      *ConfigService
	crossSchoolService *CrossSchoolService
	monitorService     *SystemMonitorService
}

type SystemStats struct {
	TotalUsers       int64  `json:"total_users"`
	TotalSchools     int64  `json:"total_schools"`
	TotalCourses     int64  `json:"total_courses"`
	TotalExperiments int64  `json:"total_experiments"`
	TotalContests    int64  `json:"total_contests"`
	ActiveEnvs       int64  `json:"active_envs"`
	OnlineUsers      int64  `json:"online_users"`
	ServerUptime     string `json:"server_uptime"`
	GoVersion        string `json:"go_version"`
	NumGoroutine     int    `json:"num_goroutine"`
}

var serverStartTime = time.Now()

func NewSystemService(
	configRepo *repository.SystemConfigRepository,
	challengeRepo *repository.ChallengeRepository,
	operationLogRepo *repository.OperationLogRepository,
	configService *ConfigService,
	crossSchoolService *CrossSchoolService,
	monitorService *SystemMonitorService,
) *SystemService {
	return &SystemService{
		configRepo:         configRepo,
		challengeRepo:      challengeRepo,
		operationLogRepo:   operationLogRepo,
		configService:      configService,
		crossSchoolService: crossSchoolService,
		monitorService:     monitorService,
	}
}

func (s *SystemService) GetConfig(ctx context.Context, key string) (*model.SystemConfig, error) {
	if s.configService != nil {
		return s.configService.Get(ctx, key)
	}
	return s.configRepo.Get(ctx, key)
}

func (s *SystemService) SetConfig(ctx context.Context, key, value, description string) error {
	if s.configService != nil {
		return s.configService.Set(ctx, key, value, "string", description, "", false)
	}
	return s.configRepo.Set(ctx, key, value, "string", description, "", false)
}

func (s *SystemService) ListConfigs(ctx context.Context, group string) ([]model.SystemConfig, error) {
	if s.configService != nil {
		return s.configService.List(ctx, group, false)
	}
	return s.configRepo.List(ctx, group, false)
}

func (s *SystemService) LogOperation(ctx context.Context, userID uint, schoolID *uint, action, module string, targetID *uint, description string, ip string) error {
	logEntry := &model.OperationLog{
		UserID:      userID,
		SchoolID:    schoolID,
		Action:      action,
		Module:      module,
		TargetID:    targetID,
		Description: description,
		RequestIP:   ip,
	}
	return s.operationLogRepo.Create(ctx, logEntry)
}

func (s *SystemService) ListOperationLogs(ctx context.Context, req *request.ListOperationLogsRequest) ([]model.OperationLog, int64, error) {
	var schoolID uint
	var userID uint
	if req.SchoolID != nil {
		schoolID = *req.SchoolID
	}
	if req.UserID != nil {
		userID = *req.UserID
	}
	return s.operationLogRepo.List(ctx, schoolID, userID, req.Action, req.Resource, req.GetPage(), req.GetPageSize())
}

func (s *SystemService) CreateCrossSchoolApplication(ctx context.Context, req *request.CreateCrossSchoolApplicationRequest) (*model.CrossSchoolApplication, error) {
	if s.crossSchoolService == nil {
		return nil, errors.ErrInternal.WithMessage("cross school service not initialized")
	}
	return s.crossSchoolService.CreateApplication(ctx, req)
}

func (s *SystemService) HandleCrossSchoolApplication(ctx context.Context, id uint, approverID uint, approved bool, reason string) error {
	if s.crossSchoolService == nil {
		return errors.ErrInternal.WithMessage("cross school service not initialized")
	}
	return s.crossSchoolService.HandleApplication(ctx, id, approverID, approved, reason)
}

func (s *SystemService) ListCrossSchoolApplications(ctx context.Context, schoolID uint, status string, page, pageSize int) ([]model.CrossSchoolApplication, int64, error) {
	if s.crossSchoolService == nil {
		return nil, 0, errors.ErrInternal.WithMessage("cross school service not initialized")
	}
	return s.crossSchoolService.ListApplications(ctx, schoolID, status, page, pageSize)
}

func (s *SystemService) ListChallengePublishRequests(ctx context.Context, schoolID uint, status string, page, pageSize int) ([]model.ChallengePublishRequest, int64, error) {
	return s.challengeRepo.ListPublishRequests(ctx, schoolID, status, page, pageSize)
}

func (s *SystemService) HandleChallengePublishRequest(ctx context.Context, requestID, reviewerID uint, approved bool, comment string) error {
	publishRequest, err := s.challengeRepo.GetPublishRequestByID(ctx, requestID)
	if err != nil {
		return errors.ErrNotFound
	}

	now := time.Now()
	publishRequest.ReviewerID = &reviewerID
	publishRequest.ReviewedAt = &now
	publishRequest.ReviewComment = comment

	if approved {
		publishRequest.Status = model.ChallengePublishStatusApproved
		challenge, _ := s.challengeRepo.GetByID(ctx, publishRequest.ChallengeID)
		if challenge != nil {
			challenge.IsPublic = true
			_ = s.challengeRepo.Update(ctx, challenge)
		}
	} else {
		publishRequest.Status = model.ChallengePublishStatusRejected
	}

	return s.challengeRepo.UpdatePublishRequest(ctx, publishRequest)
}

func (s *SystemService) GetSystemStats(ctx context.Context, schoolID *uint) (*SystemStats, error) {
	if s.monitorService == nil {
		return nil, errors.ErrInternal.WithMessage("system monitor service not initialized")
	}
	return s.monitorService.GetSystemStats(ctx, schoolID)
}

func (s *SystemService) HealthCheck(ctx context.Context) map[string]interface{} {
	if s.monitorService == nil {
		return map[string]interface{}{"status": "unknown", "message": "system monitor service not initialized"}
	}
	return s.monitorService.HealthCheck(ctx)
}

func (s *SystemService) GetSystemMonitor(ctx context.Context) (*response.SystemMonitor, error) {
	if s.monitorService == nil {
		return nil, errors.ErrInternal.WithMessage("system monitor service not initialized")
	}
	return s.monitorService.GetSystemMonitor(ctx)
}

func (s *SystemService) GetContainerStats(ctx context.Context) (*response.ContainerStats, error) {
	if s.monitorService == nil {
		return nil, errors.ErrInternal.WithMessage("system monitor service not initialized")
	}
	return s.monitorService.GetContainerStats(ctx)
}

func (s *SystemService) GetServiceHealth(ctx context.Context) ([]response.ServiceHealth, error) {
	if s.monitorService == nil {
		return nil, errors.ErrInternal.WithMessage("system monitor service not initialized")
	}
	return s.monitorService.GetServiceHealth(ctx)
}
