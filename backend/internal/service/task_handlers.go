package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/chainspace/backend/internal/model"
	"github.com/chainspace/backend/internal/pkg/logger"
	"github.com/chainspace/backend/internal/pkg/mq"
	"github.com/chainspace/backend/internal/pkg/task"
	"github.com/chainspace/backend/internal/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	TaskTypeEnvCreate    = "env.create"
	TaskTypeEnvDestroy   = "env.destroy"
	TaskTypeVulnSync     = "vuln.sync"
	TaskTypeVulnEnrich   = "vuln.enrich"
	TaskTypeVulnConvert  = "vuln.convert"
	TaskTypeNotification = "notification"
)

type TaskHandlerService struct {
	taskManager          *task.Manager
	mqClient             *mq.Client
	mqConsumer           *mq.Consumer
	envManager           *EnvManagerService
	vulnerabilityService *VulnerabilityAdminService
	notifyRepo           *repository.NotificationRepository
}

func NewTaskHandlerService(
	taskManager *task.Manager,
	mqClient *mq.Client,
	envManager *EnvManagerService,
	vulnerabilityService *VulnerabilityAdminService,
	notifyRepo *repository.NotificationRepository,
) *TaskHandlerService {
	svc := &TaskHandlerService{
		taskManager:          taskManager,
		mqClient:             mqClient,
		envManager:           envManager,
		vulnerabilityService: vulnerabilityService,
		notifyRepo:           notifyRepo,
	}

	svc.registerHandlers()

	if mqClient != nil {
		svc.mqConsumer = mq.NewConsumer(mqClient)
		svc.registerMQHandlers()
		go svc.mqConsumer.Start(context.Background())
		logger.Info("已启用 RabbitMQ 任务队列")
	} else {
		logger.Info("使用数据库任务队列")
	}

	return svc
}

func (s *TaskHandlerService) registerMQHandlers() {
	s.mqConsumer.RegisterHandler(mq.QueueEnvCreate, func(ctx context.Context, msg *mq.Message) error {
		return s.handleEnvCreate(ctx, mustMarshal(msg.Payload))
	})
	s.mqConsumer.RegisterHandler(mq.QueueEnvDestroy, func(ctx context.Context, msg *mq.Message) error {
		return s.handleEnvDestroy(ctx, mustMarshal(msg.Payload))
	})
	s.mqConsumer.RegisterHandler(mq.QueueVulnSync, func(ctx context.Context, msg *mq.Message) error {
		return s.handleVulnSync(ctx, mustMarshal(msg.Payload))
	})
	s.mqConsumer.RegisterHandler(mq.QueueVulnConvert, func(ctx context.Context, msg *mq.Message) error {
		return s.handleVulnConvert(ctx, mustMarshal(msg.Payload))
	})
	s.mqConsumer.RegisterHandler(mq.QueueNotification, func(ctx context.Context, msg *mq.Message) error {
		return s.handleNotification(ctx, mustMarshal(msg.Payload))
	})
}

func mustMarshal(v interface{}) json.RawMessage {
	data, _ := json.Marshal(v)
	return data
}

func (s *TaskHandlerService) registerHandlers() {
	s.taskManager.RegisterHandler(TaskTypeEnvCreate, s.handleEnvCreate, &task.RetryStrategy{
		MaxRetries:    5,
		InitialDelay:  10 * time.Second,
		MaxDelay:      5 * time.Minute,
		BackoffFactor: 2.0,
	})

	s.taskManager.RegisterHandler(TaskTypeEnvDestroy, s.handleEnvDestroy, &task.RetryStrategy{
		MaxRetries:    3,
		InitialDelay:  5 * time.Second,
		MaxDelay:      time.Minute,
		BackoffFactor: 2.0,
	})

	s.taskManager.RegisterHandler(TaskTypeVulnSync, s.handleVulnSync, &task.RetryStrategy{
		MaxRetries:    3,
		InitialDelay:  30 * time.Second,
		MaxDelay:      10 * time.Minute,
		BackoffFactor: 2.0,
	})

	s.taskManager.RegisterHandler(TaskTypeVulnEnrich, s.handleVulnEnrich, &task.RetryStrategy{
		MaxRetries:    2,
		InitialDelay:  10 * time.Second,
		MaxDelay:      2 * time.Minute,
		BackoffFactor: 2.0,
	})

	s.taskManager.RegisterHandler(TaskTypeVulnConvert, s.handleVulnConvert, &task.RetryStrategy{
		MaxRetries:    3,
		InitialDelay:  5 * time.Second,
		MaxDelay:      time.Minute,
		BackoffFactor: 2.0,
	})

	s.taskManager.RegisterHandler(TaskTypeNotification, s.handleNotification, &task.RetryStrategy{
		MaxRetries:    3,
		InitialDelay:  2 * time.Second,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
	})
}

func (s *TaskHandlerService) handleEnvCreate(ctx context.Context, payload json.RawMessage) error {
	var data struct {
		ExperimentID uint   `json:"experiment_id"`
		UserID       uint   `json:"user_id"`
		SchoolID     uint   `json:"school_id"`
		SnapshotURL  string `json:"snapshot_url"`
	}

	if err := json.Unmarshal(payload, &data); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	logger.Info("Processing env create task",
		zap.Uint("experiment_id", data.ExperimentID),
		zap.Uint("user_id", data.UserID),
	)

	req := &EnvCreateRequest{
		ExperimentID: data.ExperimentID,
		UserID:       data.UserID,
		SchoolID:     data.SchoolID,
		SnapshotURL:  data.SnapshotURL,
	}

	_, err := s.envManager.CreateEnv(ctx, req)
	return err
}

func (s *TaskHandlerService) handleEnvDestroy(ctx context.Context, payload json.RawMessage) error {
	var data struct {
		EnvID string `json:"env_id"`
	}

	if err := json.Unmarshal(payload, &data); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	logger.Info("Processing env destroy task", zap.String("env_id", data.EnvID))
	return s.envManager.StopEnv(ctx, data.EnvID, 0)
}

func (s *TaskHandlerService) handleVulnSync(ctx context.Context, payload json.RawMessage) error {
	var data struct {
		SourceID uint `json:"source_id"`
	}

	if err := json.Unmarshal(payload, &data); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	logger.Info("Processing vuln sync task", zap.Uint("source_id", data.SourceID))
	return s.vulnerabilityService.SyncVulnerabilityData(ctx, data.SourceID)
}

func (s *TaskHandlerService) handleVulnEnrich(ctx context.Context, payload json.RawMessage) error {
	var data struct {
		VulnID uint `json:"vuln_id"`
	}

	if err := json.Unmarshal(payload, &data); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	logger.Info("Processing vuln enrich task", zap.Uint("vuln_id", data.VulnID))
	return s.vulnerabilityService.EnrichVulnerabilityCode(ctx, data.VulnID)
}

func (s *TaskHandlerService) handleVulnConvert(ctx context.Context, payload json.RawMessage) error {
	var data struct {
		VulnID    uint `json:"vuln_id"`
		CreatorID uint `json:"creator_id"`
	}

	if err := json.Unmarshal(payload, &data); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	creatorID := data.CreatorID
	if creatorID == 0 {
		creatorID = 1
	}

	logger.Info("Processing vuln convert task", zap.Uint("vuln_id", data.VulnID))
	_, err := s.vulnerabilityService.ConvertVulnerabilityToChallenge(ctx, data.VulnID, creatorID)
	return err
}

func (s *TaskHandlerService) handleNotification(ctx context.Context, payload json.RawMessage) error {
	var data struct {
		UserID  uint                   `json:"user_id"`
		Type    string                 `json:"type"`
		Title   string                 `json:"title"`
		Content string                 `json:"content"`
		Extra   map[string]interface{} `json:"extra"`
	}

	if err := json.Unmarshal(payload, &data); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	logger.Info("Processing notification task",
		zap.Uint("user_id", data.UserID),
		zap.String("type", data.Type),
	)

	notification := &model.Notification{
		UserID:  data.UserID,
		Type:    data.Type,
		Title:   data.Title,
		Content: data.Content,
	}

	return s.notifyRepo.Create(ctx, notification)
}

func (s *TaskHandlerService) SubmitEnvCreateTask(ctx context.Context, experimentID, userID, schoolID uint, snapshotURL string) (string, error) {
	payload := map[string]interface{}{
		"experiment_id": experimentID,
		"user_id":       userID,
		"school_id":     schoolID,
		"snapshot_url":  snapshotURL,
	}

	if s.mqClient != nil {
		msgID := uuid.New().String()
		err := s.mqClient.Publish(ctx, mq.QueueEnvCreate, &mq.Message{
			ID:        msgID,
			Type:      TaskTypeEnvCreate,
			Payload:   payload,
			CreatedAt: time.Now(),
		})
		return msgID, err
	}

	return s.taskManager.Submit(ctx, TaskTypeEnvCreate, payload)
}

func (s *TaskHandlerService) SubmitEnvDestroyTask(ctx context.Context, envID string) (string, error) {
	payload := map[string]interface{}{"env_id": envID}
	if s.mqClient != nil {
		msgID := uuid.New().String()
		err := s.mqClient.Publish(ctx, mq.QueueEnvDestroy, &mq.Message{
			ID:        msgID,
			Type:      TaskTypeEnvDestroy,
			Payload:   payload,
			CreatedAt: time.Now(),
		})
		return msgID, err
	}

	return s.taskManager.Submit(ctx, TaskTypeEnvDestroy, payload)
}

func (s *TaskHandlerService) SubmitVulnSyncTask(ctx context.Context, sourceID uint) (string, error) {
	if existingTaskID, ok := s.findActiveVulnSyncTask(ctx, sourceID); ok {
		return existingTaskID, nil
	}

	payload := map[string]interface{}{"source_id": sourceID}
	return s.taskManager.Submit(ctx, TaskTypeVulnSync, payload)
}

func (s *TaskHandlerService) findActiveVulnSyncTask(ctx context.Context, sourceID uint) (string, bool) {
	return s.findActiveTask(ctx, TaskTypeVulnSync, func(payload json.RawMessage) bool {
		var data struct {
			SourceID uint `json:"source_id"`
		}
		if err := json.Unmarshal(payload, &data); err != nil {
			return false
		}
		return data.SourceID == sourceID
	})
}

func (s *TaskHandlerService) SubmitVulnEnrichTask(ctx context.Context, vulnID uint) (string, error) {
	if existingTaskID, ok := s.findActiveVulnEnrichTask(ctx, vulnID); ok {
		return existingTaskID, nil
	}

	payload := map[string]interface{}{"vuln_id": vulnID}
	return s.taskManager.Submit(ctx, TaskTypeVulnEnrich, payload)
}

func (s *TaskHandlerService) findActiveVulnEnrichTask(ctx context.Context, vulnID uint) (string, bool) {
	return s.findActiveTask(ctx, TaskTypeVulnEnrich, func(payload json.RawMessage) bool {
		var data struct {
			VulnID uint `json:"vuln_id"`
		}
		if err := json.Unmarshal(payload, &data); err != nil {
			return false
		}
		return data.VulnID == vulnID
	})
}

func (s *TaskHandlerService) findActiveTask(ctx context.Context, taskType string, match func(payload json.RawMessage) bool) (string, bool) {
	if s.taskManager == nil {
		return "", false
	}

	statuses := []task.Status{task.StatusPending, task.StatusRunning, task.StatusRetrying}
	for _, status := range statuses {
		tasks, _, err := s.taskManager.ListTasks(ctx, status, taskType, 1, 50)
		if err != nil {
			continue
		}

		for i := range tasks {
			if match(tasks[i].Payload) {
				return tasks[i].ID, true
			}
		}
	}

	return "", false
}

func (s *TaskHandlerService) SubmitVulnConvertTask(ctx context.Context, vulnID uint) (string, error) {
	payload := map[string]interface{}{"vuln_id": vulnID}
	if s.mqClient != nil {
		msgID := uuid.New().String()
		err := s.mqClient.Publish(ctx, mq.QueueVulnConvert, &mq.Message{
			ID:        msgID,
			Type:      TaskTypeVulnConvert,
			Payload:   payload,
			CreatedAt: time.Now(),
		})
		return msgID, err
	}

	return s.taskManager.Submit(ctx, TaskTypeVulnConvert, payload)
}

func (s *TaskHandlerService) SubmitNotificationTask(ctx context.Context, userID uint, notifyType, title, content string) (string, error) {
	payload := map[string]interface{}{
		"user_id": userID,
		"type":    notifyType,
		"title":   title,
		"content": content,
	}

	if s.mqClient != nil {
		msgID := uuid.New().String()
		err := s.mqClient.Publish(ctx, mq.QueueNotification, &mq.Message{
			ID:        msgID,
			Type:      TaskTypeNotification,
			Payload:   payload,
			CreatedAt: time.Now(),
		})
		return msgID, err
	}

	return s.taskManager.Submit(ctx, TaskTypeNotification, payload)
}
