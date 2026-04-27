package task

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/chainspace/backend/internal/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Status 任务状态
type Status string

const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
	StatusRetrying  Status = "retrying"
)

// Task 异步任务
type Task struct {
	ID          string          `gorm:"primaryKey;size:36" json:"id"`
	Type        string          `gorm:"size:50;index;not null" json:"type"`
	Payload     json.RawMessage `gorm:"type:jsonb" json:"payload"`
	Status      Status          `gorm:"size:20;index;default:pending" json:"status"`
	RetryCount  int             `gorm:"default:0" json:"retry_count"`
	MaxRetries  int             `gorm:"default:3" json:"max_retries"`
	LastError   string          `gorm:"type:text" json:"last_error"`
	NextRetryAt *time.Time      `gorm:"index" json:"next_retry_at"`
	StartedAt   *time.Time      `json:"started_at"`
	CompletedAt *time.Time      `json:"completed_at"`
	CreatedAt   time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 表名
func (Task) TableName() string {
	return "async_tasks"
}

// Handler 任务处理函数
type Handler func(ctx context.Context, payload json.RawMessage) error

// NonRetryableError 表示不应进行自动重试的确定性错误。
type NonRetryableError struct {
	Err error
}

func (e NonRetryableError) Error() string {
	if e.Err == nil {
		return ""
	}
	return e.Err.Error()
}

func (e NonRetryableError) Unwrap() error {
	return e.Err
}

func Permanent(err error) error {
	if err == nil {
		return nil
	}
	return NonRetryableError{Err: err}
}

func IsNonRetryable(err error) bool {
	var target NonRetryableError
	return errors.As(err, &target)
}

// RetryStrategy 重试策略
type RetryStrategy struct {
	MaxRetries    int
	InitialDelay  time.Duration
	MaxDelay      time.Duration
	BackoffFactor float64
}

// DefaultRetryStrategy 默认重试策略
var DefaultRetryStrategy = RetryStrategy{
	MaxRetries:    3,
	InitialDelay:  5 * time.Second,
	MaxDelay:      5 * time.Minute,
	BackoffFactor: 2.0,
}

// Manager 任务管理器
type Manager struct {
	db          *gorm.DB
	handlers    map[string]Handler
	strategies  map[string]RetryStrategy
	mu          sync.RWMutex
	workerCount int
	stopCh      chan struct{}
	wg          sync.WaitGroup
}

// NewManager 创建任务管理器
func NewManager(db *gorm.DB, workerCount int) *Manager {
	if workerCount <= 0 {
		workerCount = 5
	}
	return &Manager{
		db:          db,
		handlers:    make(map[string]Handler),
		strategies:  make(map[string]RetryStrategy),
		workerCount: workerCount,
		stopCh:      make(chan struct{}),
	}
}

// RegisterHandler 注册任务处理器
func (m *Manager) RegisterHandler(taskType string, handler Handler, strategy *RetryStrategy) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers[taskType] = handler
	if strategy != nil {
		m.strategies[taskType] = *strategy
	} else {
		m.strategies[taskType] = DefaultRetryStrategy
	}
}

// Submit 提交任务
func (m *Manager) Submit(ctx context.Context, taskType string, payload interface{}) (string, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal payload: %w", err)
	}

	m.mu.RLock()
	strategy, ok := m.strategies[taskType]
	m.mu.RUnlock()
	if !ok {
		strategy = DefaultRetryStrategy
	}

	task := &Task{
		ID:         uuid.New().String(),
		Type:       taskType,
		Payload:    payloadBytes,
		Status:     StatusPending,
		MaxRetries: strategy.MaxRetries,
	}

	if err := m.db.WithContext(ctx).Create(task).Error; err != nil {
		return "", fmt.Errorf("create task: %w", err)
	}

	return task.ID, nil
}

// Start 启动任务处理
func (m *Manager) Start() {
	// 启动工作协程
	for i := 0; i < m.workerCount; i++ {
		m.wg.Add(1)
		go m.worker()
	}

	// 启动重试检查协程
	m.wg.Add(1)
	go m.retryChecker()

	logger.Info("Task manager started", zap.Int("workers", m.workerCount))
}

// Stop 停止任务处理
func (m *Manager) Stop() {
	close(m.stopCh)
	m.wg.Wait()
	logger.Info("Task manager stopped")
}

// worker 工作协程
func (m *Manager) worker() {
	defer m.wg.Done()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.claimAndExecutePendingTask()
		}
	}
}

func (m *Manager) claimAndExecutePendingTask() {
	ctx := context.Background()

	var task Task
	err := m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("status = ? AND (next_retry_at IS NULL OR next_retry_at <= ?)", StatusPending, time.Now()).
			Order("created_at ASC").
			First(&task).Error; err != nil {
			return err
		}

		now := time.Now()
		result := tx.Model(&task).
			Where("status = ?", StatusPending).
			Updates(map[string]interface{}{
				"status":     StatusRunning,
				"started_at": now,
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		task.Status = StatusRunning
		task.StartedAt = &now
		return nil
	})

	if err != nil {
		if err != gorm.ErrRecordNotFound {
			logger.Error("Failed to fetch pending task", zap.Error(err))
		}
		return
	}

	m.executeTask(ctx, &task)
}

// executeTask 执行任务
func (m *Manager) executeTask(ctx context.Context, task *Task) {
	m.mu.RLock()
	handler, ok := m.handlers[task.Type]
	strategy := m.strategies[task.Type]
	m.mu.RUnlock()

	if !ok {
		logger.Error("No handler registered for task type", zap.String("type", task.Type))
		m.markTaskFailed(ctx, task, "no handler registered")
		return
	}

	// 执行处理器
	err := handler(ctx, task.Payload)
	if err != nil {
		logger.Error("Task execution failed",
			zap.String("task_id", task.ID),
			zap.String("type", task.Type),
			zap.Int("retry_count", task.RetryCount),
			zap.Error(err),
		)
		m.handleTaskError(ctx, task, err, strategy)
		return
	}

	// 标记完成
	m.markTaskCompleted(ctx, task)
}

// handleTaskError 处理任务错误
func (m *Manager) handleTaskError(ctx context.Context, task *Task, err error, strategy RetryStrategy) {
	task.RetryCount++
	task.LastError = err.Error()

	if IsNonRetryable(err) {
		m.markTaskFailed(ctx, task, err.Error())
		return
	}

	if task.RetryCount >= task.MaxRetries {
		m.markTaskFailed(ctx, task, err.Error())
		return
	}

	// 计算下次重试时间（指数退避）
	delay := strategy.InitialDelay
	for i := 0; i < task.RetryCount-1; i++ {
		delay = time.Duration(float64(delay) * strategy.BackoffFactor)
		if delay > strategy.MaxDelay {
			delay = strategy.MaxDelay
			break
		}
	}
	nextRetry := time.Now().Add(delay)

	m.db.WithContext(ctx).Model(task).Updates(map[string]interface{}{
		"status":        StatusPending,
		"retry_count":   task.RetryCount,
		"last_error":    task.LastError,
		"next_retry_at": nextRetry,
	})

	logger.Info("Task scheduled for retry",
		zap.String("task_id", task.ID),
		zap.Int("retry_count", task.RetryCount),
		zap.Time("next_retry_at", nextRetry),
	)
}

// markTaskCompleted 标记任务完成
func (m *Manager) markTaskCompleted(ctx context.Context, task *Task) {
	now := time.Now()
	m.db.WithContext(ctx).Model(task).Updates(map[string]interface{}{
		"status":       StatusCompleted,
		"completed_at": now,
	})
	logger.Info("Task completed", zap.String("task_id", task.ID), zap.String("type", task.Type))
}

// markTaskFailed 标记任务失败
func (m *Manager) markTaskFailed(ctx context.Context, task *Task, errMsg string) {
	now := time.Now()
	m.db.WithContext(ctx).Model(task).Updates(map[string]interface{}{
		"status":       StatusFailed,
		"last_error":   errMsg,
		"completed_at": now,
	})
	logger.Error("Task failed permanently",
		zap.String("task_id", task.ID),
		zap.String("type", task.Type),
		zap.String("error", errMsg),
	)
}

// retryChecker 重试检查协程
func (m *Manager) retryChecker() {
	defer m.wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			// 将超时的运行中任务重置为待处理
			ctx := context.Background()
			threshold := time.Now().Add(-30 * time.Minute)
			m.db.WithContext(ctx).Model(&Task{}).
				Where("status = ? AND started_at < ?", StatusRunning, threshold).
				Updates(map[string]interface{}{
					"status": StatusPending,
				})
		}
	}
}

// GetTask 获取任务详情
func (m *Manager) GetTask(ctx context.Context, taskID string) (*Task, error) {
	var task Task
	if err := m.db.WithContext(ctx).First(&task, "id = ?", taskID).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

// ListTasks 列出任务
func (m *Manager) ListTasks(ctx context.Context, status Status, taskType string, page, pageSize int) ([]Task, int64, error) {
	var tasks []Task
	var total int64

	query := m.db.WithContext(ctx).Model(&Task{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if taskType != "" {
		query = query.Where("type = ?", taskType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// CleanupOldTasks 清理旧任务
func (m *Manager) CleanupOldTasks(ctx context.Context, days int) error {
	threshold := time.Now().AddDate(0, 0, -days)
	return m.db.WithContext(ctx).
		Where("status IN ? AND completed_at < ?", []Status{StatusCompleted, StatusFailed}, threshold).
		Delete(&Task{}).Error
}
