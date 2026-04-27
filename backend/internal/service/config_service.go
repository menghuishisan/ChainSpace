package service

import (
	"context"
	"fmt"

	"github.com/chainspace/backend/internal/model"
	"github.com/chainspace/backend/internal/pkg/cache"
	"github.com/chainspace/backend/internal/pkg/logger"
	"github.com/chainspace/backend/internal/repository"
	"go.uber.org/zap"
)

// ConfigService 配置服务（带缓存和热更新）
type ConfigService struct {
	configRepo  *repository.SystemConfigRepository
	configCache *cache.ConfigCache
}

// NewConfigService 创建配置服务
func NewConfigService(configRepo *repository.SystemConfigRepository, configCache *cache.ConfigCache) *ConfigService {
	svc := &ConfigService{
		configRepo:  configRepo,
		configCache: configCache,
	}

	// 订阅配置变更通知
	if configCache != nil {
		configCache.Subscribe(func(key string, value interface{}) {
			logger.Info("Config changed", zap.String("key", key))
		})
	}

	return svc
}

// Get 获取配置（优先从缓存获取）
func (s *ConfigService) Get(ctx context.Context, key string) (*model.SystemConfig, error) {
	// 1. 尝试从缓存获取
	if s.configCache != nil {
		if value, ok := s.configCache.Get(ctx, key); ok {
			if config, ok := value.(*model.SystemConfig); ok {
				return config, nil
			}
		}
	}

	// 2. 从数据库获取
	config, err := s.configRepo.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	// 3. 写入缓存
	if s.configCache != nil && config != nil {
		s.configCache.Set(ctx, key, config)
	}

	return config, nil
}

// GetValue 获取配置值
func (s *ConfigService) GetValue(ctx context.Context, key string, defaultValue string) string {
	config, err := s.Get(ctx, key)
	if err != nil || config == nil {
		return defaultValue
	}
	return config.Value
}

// Set 设置配置（同时更新数据库和缓存）
func (s *ConfigService) Set(ctx context.Context, key, value, configType, description, group string, isPublic bool) error {
	// 1. 更新数据库
	if err := s.configRepo.Set(ctx, key, value, configType, description, group, isPublic); err != nil {
		return fmt.Errorf("update database: %w", err)
	}

	// 2. 获取完整配置对象
	config, _ := s.configRepo.Get(ctx, key)

	// 3. 更新缓存（会自动广播到其他节点）
	if s.configCache != nil && config != nil {
		if err := s.configCache.Set(ctx, key, config); err != nil {
			logger.Warn("Failed to update config cache", zap.String("key", key), zap.Error(err))
		}
	}

	return nil
}

// Delete 删除配置
func (s *ConfigService) Delete(ctx context.Context, key string) error {
	// 1. 从数据库删除
	if err := s.configRepo.Delete(ctx, key); err != nil {
		return fmt.Errorf("delete from database: %w", err)
	}

	// 2. 从缓存删除
	if s.configCache != nil {
		if err := s.configCache.Delete(ctx, key); err != nil {
			logger.Warn("Failed to delete config from cache", zap.String("key", key), zap.Error(err))
		}
	}

	return nil
}

// List 获取配置列表
func (s *ConfigService) List(ctx context.Context, group string, publicOnly bool) ([]model.SystemConfig, error) {
	return s.configRepo.List(ctx, group, publicOnly)
}

// RefreshCache 刷新缓存（从数据库重新加载）
func (s *ConfigService) RefreshCache(ctx context.Context, key string) error {
	if s.configCache == nil {
		return nil
	}

	return s.configCache.Refresh(ctx, key, func() (interface{}, error) {
		return s.configRepo.Get(ctx, key)
	})
}

// RefreshAllCache 刷新所有缓存
func (s *ConfigService) RefreshAllCache(ctx context.Context) error {
	if s.configCache == nil {
		return nil
	}

	// 清空缓存
	if err := s.configCache.Clear(ctx); err != nil {
		return fmt.Errorf("clear cache: %w", err)
	}

	// 重新加载所有配置
	configs, err := s.configRepo.List(ctx, "", false)
	if err != nil {
		return fmt.Errorf("list configs: %w", err)
	}

	for _, config := range configs {
		configCopy := config
		if err := s.configCache.Set(ctx, config.Key, &configCopy); err != nil {
			logger.Warn("Failed to cache config", zap.String("key", config.Key), zap.Error(err))
		}
	}

	logger.Info("Config cache refreshed", zap.Int("count", len(configs)))
	return nil
}

// PreloadCache 预加载缓存
func (s *ConfigService) PreloadCache(ctx context.Context) error {
	return s.RefreshAllCache(ctx)
}
