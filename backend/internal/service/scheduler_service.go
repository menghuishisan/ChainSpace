package service

import (
	"context"
	"runtime/debug"
	"sync"
	"time"

	"github.com/chainspace/backend/internal/pkg/logger"
	"go.uber.org/zap"
)

// SchedulerService 定时任务服务
type SchedulerService struct {
	envManager    *EnvManagerService
	notifyService *NotificationService
	taskHandler   *TaskHandlerService
	tasks         map[string]*ScheduledTask
	stopChan      chan struct{}
	wg            sync.WaitGroup
	mu            sync.RWMutex
}

// ScheduledTask 定时任务
type ScheduledTask struct {
	Name     string
	Interval time.Duration
	Handler  func(ctx context.Context) error
	LastRun  time.Time
	Running  bool
}

// NewSchedulerService 创建定时任务服务
func NewSchedulerService(envManager *EnvManagerService, notifyService *NotificationService) *SchedulerService {
	return &SchedulerService{
		envManager:    envManager,
		notifyService: notifyService,
		tasks:         make(map[string]*ScheduledTask),
		stopChan:      make(chan struct{}),
	}
}

// SetTaskHandler 设置任务处理服务（用于异步任务提交）
func (s *SchedulerService) SetTaskHandler(taskHandler *TaskHandlerService) {
	s.taskHandler = taskHandler
}

// Start 启动定时任务服务
func (s *SchedulerService) Start() {
	logger.Info("Starting scheduler service")

	// 注册定时任务
	s.registerTasks()

	// 启动任务调度器
	s.wg.Add(1)
	go s.runScheduler()
}

// Stop 停止定时任务服务
func (s *SchedulerService) Stop() {
	logger.Info("Stopping scheduler service")
	close(s.stopChan)
	s.wg.Wait()
}

// registerTasks 注册所有定时任务
func (s *SchedulerService) registerTasks() {
	// 环境过期清理任务 - 每分钟执行
	s.RegisterTask("cleanup_expired_envs", time.Minute, func(ctx context.Context) error {
		return s.envManager.CleanupExpiredEnvs(ctx)
	})

	// 环境过期预警任务 - 每5分钟执行，提前30分钟提醒（docs要求）
	s.RegisterTask("notify_expiring_envs", 5*time.Minute, func(ctx context.Context) error {
		return s.envManager.NotifyExpiringEnvs(ctx, 30) // 30分钟提前提醒
	})

	// 清理过期通知 - 每天执行
	s.RegisterTask("cleanup_old_notifications", 24*time.Hour, func(ctx context.Context) error {
		if s.notifyService != nil {
			return s.notifyService.CleanupOldNotifications(ctx, 30)
		}
		return nil
	})

	// 环境状态同步任务 - 每30秒执行
	s.RegisterTask("sync_env_status", 30*time.Second, func(ctx context.Context) error {
		return s.syncAllEnvStatus(ctx)
	})

	// 清理旧异步任务 - 每天执行（使用异步任务系统）
	s.RegisterTask("cleanup_old_tasks", 24*time.Hour, func(ctx context.Context) error {
		if s.taskHandler != nil && s.taskHandler.taskManager != nil {
			return s.taskHandler.taskManager.CleanupOldTasks(ctx, 7) // 保留7天
		}
		return nil
	})
}

// RegisterTask 注册定时任务
func (s *SchedulerService) RegisterTask(name string, interval time.Duration, handler func(ctx context.Context) error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tasks[name] = &ScheduledTask{
		Name:     name,
		Interval: interval,
		Handler:  handler,
	}

	logger.Info("Registered scheduled task", zap.String("name", name), zap.Duration("interval", interval))
}

// runScheduler 运行调度器
func (s *SchedulerService) runScheduler() {
	defer s.wg.Done()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			logger.Info("Scheduler stopped")
			return
		case <-ticker.C:
			s.checkAndRunTasks()
		}
	}
}

// checkAndRunTasks 检查并运行到期的任务
func (s *SchedulerService) checkAndRunTasks() {
	s.mu.RLock()
	tasks := make([]*ScheduledTask, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}
	s.mu.RUnlock()

	now := time.Now()
	for _, task := range tasks {
		if task.Running {
			continue
		}

		if now.Sub(task.LastRun) >= task.Interval {
			go s.runTask(task)
		}
	}
}

// runTask 运行单个任务（带panic恢复，防止任务崩溃影响整个调度器）
func (s *SchedulerService) runTask(task *ScheduledTask) {
	s.mu.Lock()
	if task.Running {
		s.mu.Unlock()
		return
	}
	task.Running = true
	s.mu.Unlock()

	defer func() {
		// panic恢复
		if r := recover(); r != nil {
			logger.Error("Scheduled task panic recovered",
				zap.String("name", task.Name),
				zap.Any("panic", r),
				zap.String("stack", string(debug.Stack())),
			)
		}
		s.mu.Lock()
		task.Running = false
		task.LastRun = time.Now()
		s.mu.Unlock()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	startTime := time.Now()
	err := task.Handler(ctx)
	duration := time.Since(startTime)

	if err != nil {
		logger.Error("Scheduled task failed", zap.String("name", task.Name), zap.Duration("duration", duration), zap.Error(err))
	} else {
		logger.Debug("Scheduled task completed", zap.String("name", task.Name), zap.Duration("duration", duration))
	}
}

// syncAllEnvStatus 同步所有活跃环境状态
func (s *SchedulerService) syncAllEnvStatus(ctx context.Context) error {
	if s.envManager == nil {
		return nil
	}

	// 获取所有运行中的环境并同步状态
	return s.envManager.SyncAllEnvStatus(ctx)
}

// RunTaskNow 立即运行指定任务（用于手动触发）
func (s *SchedulerService) RunTaskNow(name string) error {
	s.mu.RLock()
	task, ok := s.tasks[name]
	s.mu.RUnlock()

	if !ok {
		return nil
	}

	go s.runTask(task)
	return nil
}

// GetTaskStatus 获取任务状态
func (s *SchedulerService) GetTaskStatus() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := make(map[string]interface{})
	for name, task := range s.tasks {
		status[name] = map[string]interface{}{
			"interval": task.Interval.String(),
			"last_run": task.LastRun,
			"running":  task.Running,
		}
	}
	return status
}
