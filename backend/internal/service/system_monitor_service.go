package service

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/chainspace/backend/internal/dto/response"
	"github.com/chainspace/backend/internal/model"
	"github.com/redis/go-redis/v9"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"gorm.io/gorm"
)

// SystemMonitorService 负责系统监控与健康检查。
type SystemMonitorService struct {
	db            *gorm.DB
	redis         *redis.Client
	uploadService *UploadService
	k8sEnabled    bool
}

func NewSystemMonitorService(db *gorm.DB, redis *redis.Client, uploadService *UploadService, k8sEnabled bool) *SystemMonitorService {
	return &SystemMonitorService{
		db:            db,
		redis:         redis,
		uploadService: uploadService,
		k8sEnabled:    k8sEnabled,
	}
}

func (s *SystemMonitorService) GetSystemStats(ctx context.Context, schoolID *uint) (*SystemStats, error) {
	stats := &SystemStats{
		ServerUptime: time.Since(serverStartTime).String(),
		GoVersion:    runtime.Version(),
		NumGoroutine: runtime.NumGoroutine(),
	}

	if schoolID != nil {
		s.db.Model(&model.User{}).Where("school_id = ?", *schoolID).Count(&stats.TotalUsers)
		stats.TotalSchools = 1
		s.db.Model(&model.Course{}).Where("school_id = ?", *schoolID).Count(&stats.TotalCourses)
		s.db.Model(&model.Experiment{}).Where("school_id = ?", *schoolID).Count(&stats.TotalExperiments)
		s.db.Model(&model.Contest{}).Where("school_id = ?", *schoolID).Count(&stats.TotalContests)
		s.db.Model(&model.ExperimentEnv{}).
			Joins("JOIN experiments ON experiments.id = experiment_envs.experiment_id").
			Where("experiments.school_id = ? AND experiment_envs.status IN ?", *schoolID, []string{model.EnvStatusRunning, model.EnvStatusCreating}).
			Count(&stats.ActiveEnvs)
	} else {
		s.db.Model(&model.User{}).Count(&stats.TotalUsers)
		s.db.Model(&model.School{}).Count(&stats.TotalSchools)
		s.db.Model(&model.Course{}).Count(&stats.TotalCourses)
		s.db.Model(&model.Experiment{}).Count(&stats.TotalExperiments)
		s.db.Model(&model.Contest{}).Count(&stats.TotalContests)
		s.db.Model(&model.ExperimentEnv{}).Where("status IN ?", []string{model.EnvStatusRunning, model.EnvStatusCreating}).Count(&stats.ActiveEnvs)
	}

	if s.redis != nil {
		onlineCount, _ := s.redis.SCard(ctx, "online_users").Result()
		stats.OnlineUsers = onlineCount
	}

	return stats, nil
}

func (s *SystemMonitorService) HealthCheck(ctx context.Context) map[string]interface{} {
	result := map[string]interface{}{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	}

	sqlDB, err := s.db.DB()
	if err != nil || sqlDB.Ping() != nil {
		result["database"] = "unhealthy"
		result["status"] = "degraded"
	} else {
		result["database"] = "healthy"
	}

	if s.redis != nil {
		if err := s.redis.Ping(ctx).Err(); err != nil {
			result["redis"] = "unhealthy"
			result["status"] = "degraded"
		} else {
			result["redis"] = "healthy"
		}
	} else {
		result["redis"] = "not_configured"
	}

	if s.uploadService != nil {
		if s.uploadService.IsHealthy(ctx) {
			result["minio"] = "healthy"
		} else {
			result["minio"] = "unhealthy"
			result["status"] = "degraded"
		}
	} else {
		result["minio"] = "not_configured"
	}

	if s.k8sEnabled {
		result["kubernetes"] = "enabled"
	} else {
		result["kubernetes"] = "disabled"
	}

	return result
}

func (s *SystemMonitorService) GetSystemMonitor(ctx context.Context) (*response.SystemMonitor, error) {
	_ = ctx
	monitor := &response.SystemMonitor{
		Uptime: int64(time.Since(serverStartTime).Seconds()),
	}

	if cpuPercent, err := cpu.Percent(time.Second, false); err == nil && len(cpuPercent) > 0 {
		monitor.CPUUsage = cpuPercent[0]
	}

	if vmStat, err := mem.VirtualMemory(); err == nil {
		monitor.MemoryUsage = vmStat.UsedPercent
		monitor.MemoryTotal = formatBytes(vmStat.Total)
		monitor.MemoryUsed = formatBytes(vmStat.Used)
	}

	if diskStat, err := disk.Usage("/"); err == nil {
		monitor.DiskUsage = diskStat.UsedPercent
		monitor.DiskTotal = formatBytes(diskStat.Total)
		monitor.DiskUsed = formatBytes(diskStat.Used)
	}

	if loadStat, err := load.Avg(); err == nil {
		monitor.LoadAverage = []float64{loadStat.Load1, loadStat.Load5, loadStat.Load15}
	} else {
		monitor.LoadAverage = []float64{0, 0, 0}
	}

	return monitor, nil
}

func (s *SystemMonitorService) GetContainerStats(ctx context.Context) (*response.ContainerStats, error) {
	_ = ctx
	stats := &response.ContainerStats{
		Containers: []response.ContainerInfo{},
	}

	var runningCount int64
	var stoppedCount int64
	s.db.Model(&model.ExperimentEnv{}).Where("status = ?", model.EnvStatusRunning).Count(&runningCount)
	s.db.Model(&model.ExperimentEnv{}).Where("status IN ?", []string{model.EnvStatusTerminated, model.EnvStatusFailed}).Count(&stoppedCount)

	stats.Running = int(runningCount)
	stats.Stopped = int(stoppedCount)
	stats.Total = stats.Running + stats.Stopped

	var envs []model.ExperimentEnv
	s.db.Where("status = ?", model.EnvStatusRunning).Order("created_at DESC").Limit(20).Find(&envs)
	for _, env := range envs {
		stats.Containers = append(stats.Containers, response.ContainerInfo{
			ID:          env.EnvID,
			Name:        fmt.Sprintf("exp-%d-env", env.ExperimentID),
			Status:      env.Status,
			CPUPercent:  0,
			MemoryUsage: "N/A",
			CreatedAt:   env.CreatedAt.Format(time.RFC3339),
		})
	}

	return stats, nil
}

func (s *SystemMonitorService) GetServiceHealth(ctx context.Context) ([]response.ServiceHealth, error) {
	services := make([]response.ServiceHealth, 0, 3)
	now := time.Now().Format(time.RFC3339)

	dbHealth := response.ServiceHealth{Name: "database", LastCheck: now}
	start := time.Now()
	sqlDB, err := s.db.DB()
	if err != nil || sqlDB.Ping() != nil {
		dbHealth.Status = "unhealthy"
		dbHealth.Message = "Database connection failed"
	} else {
		dbHealth.Status = "healthy"
		dbHealth.Latency = time.Since(start).Milliseconds()
	}
	services = append(services, dbHealth)

	redisHealth := response.ServiceHealth{Name: "redis", LastCheck: now}
	if s.redis != nil {
		start = time.Now()
		if _, err := s.redis.Ping(ctx).Result(); err != nil {
			redisHealth.Status = "unhealthy"
			redisHealth.Message = err.Error()
		} else {
			redisHealth.Status = "healthy"
			redisHealth.Latency = time.Since(start).Milliseconds()
		}
	} else {
		redisHealth.Status = "unknown"
		redisHealth.Message = "Redis not configured"
	}
	services = append(services, redisHealth)

	k8sHealth := response.ServiceHealth{Name: "kubernetes", LastCheck: now}
	if s.k8sEnabled {
		k8sHealth.Status = "healthy"
		k8sHealth.Message = "K8s integration enabled"
	} else {
		k8sHealth.Status = "unknown"
		k8sHealth.Message = "K8s integration disabled"
	}
	services = append(services, k8sHealth)

	return services, nil
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div := uint64(unit)
	exp := 0
	for value := bytes / unit; value >= unit; value /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
