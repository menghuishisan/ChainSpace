package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/chainspace/backend/internal/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// ConfigCache 配置缓存管理器
type ConfigCache struct {
	redis       *redis.Client
	localCache  sync.Map
	ttl         time.Duration
	prefix      string
	subscribers []func(key string, value interface{})
	mu          sync.RWMutex
}

// ConfigCacheOptions 配置缓存选项
type ConfigCacheOptions struct {
	Redis  *redis.Client
	TTL    time.Duration
	Prefix string
}

// NewConfigCache 创建配置缓存
func NewConfigCache(opts *ConfigCacheOptions) *ConfigCache {
	if opts.TTL == 0 {
		opts.TTL = 5 * time.Minute
	}
	if opts.Prefix == "" {
		opts.Prefix = "config:"
	}

	cc := &ConfigCache{
		redis:       opts.Redis,
		ttl:         opts.TTL,
		prefix:      opts.Prefix,
		subscribers: make([]func(string, interface{}), 0),
	}

	// 启动Redis订阅监听（用于集群同步）
	if opts.Redis != nil {
		go cc.subscribeConfigChanges()
	}

	return cc
}

// Get 获取配置（优先本地缓存 -> Redis -> 返回nil）
func (c *ConfigCache) Get(ctx context.Context, key string) (interface{}, bool) {
	// 1. 检查本地缓存
	if value, ok := c.localCache.Load(key); ok {
		entry := value.(*cacheEntry)
		if time.Now().Before(entry.expireAt) {
			return entry.value, true
		}
		c.localCache.Delete(key)
	}

	// 2. 检查Redis
	if c.redis != nil {
		val, err := c.redis.Get(ctx, c.prefix+key).Result()
		if err == nil {
			var value interface{}
			if json.Unmarshal([]byte(val), &value) == nil {
				c.setLocal(key, value)
				return value, true
			}
		}
	}

	return nil, false
}

// Set 设置配置（同时更新本地缓存和Redis，并通知其他节点）
func (c *ConfigCache) Set(ctx context.Context, key string, value interface{}) error {
	// 1. 更新本地缓存
	c.setLocal(key, value)

	// 2. 更新Redis
	if c.redis != nil {
		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("marshal config value: %w", err)
		}

		if err := c.redis.Set(ctx, c.prefix+key, data, c.ttl).Err(); err != nil {
			return fmt.Errorf("set redis: %w", err)
		}

		// 3. 发布配置变更通知（用于集群同步）
		changeMsg := map[string]interface{}{
			"key":   key,
			"value": value,
			"time":  time.Now().Unix(),
		}
		msgData, _ := json.Marshal(changeMsg)
		c.redis.Publish(ctx, c.prefix+"changes", msgData)
	}

	// 4. 通知本地订阅者
	c.notifySubscribers(key, value)

	return nil
}

// Delete 删除配置
func (c *ConfigCache) Delete(ctx context.Context, key string) error {
	c.localCache.Delete(key)

	if c.redis != nil {
		if err := c.redis.Del(ctx, c.prefix+key).Err(); err != nil {
			return fmt.Errorf("delete from redis: %w", err)
		}

		// 发布删除通知
		changeMsg := map[string]interface{}{
			"key":     key,
			"deleted": true,
			"time":    time.Now().Unix(),
		}
		msgData, _ := json.Marshal(changeMsg)
		c.redis.Publish(ctx, c.prefix+"changes", msgData)
	}

	return nil
}

// Refresh 刷新配置（从数据源重新加载）
func (c *ConfigCache) Refresh(ctx context.Context, key string, loader func() (interface{}, error)) error {
	value, err := loader()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	return c.Set(ctx, key, value)
}

// Subscribe 订阅配置变更
func (c *ConfigCache) Subscribe(handler func(key string, value interface{})) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.subscribers = append(c.subscribers, handler)
}

// setLocal 设置本地缓存
func (c *ConfigCache) setLocal(key string, value interface{}) {
	c.localCache.Store(key, &cacheEntry{
		value:    value,
		expireAt: time.Now().Add(c.ttl),
	})
}

// notifySubscribers 通知订阅者
func (c *ConfigCache) notifySubscribers(key string, value interface{}) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, handler := range c.subscribers {
		go handler(key, value)
	}
}

// subscribeConfigChanges 订阅Redis配置变更通知
func (c *ConfigCache) subscribeConfigChanges() {
	ctx := context.Background()
	pubsub := c.redis.Subscribe(ctx, c.prefix+"changes")
	defer pubsub.Close()

	ch := pubsub.Channel()
	for msg := range ch {
		var changeMsg map[string]interface{}
		if err := json.Unmarshal([]byte(msg.Payload), &changeMsg); err != nil {
			logger.Error("Failed to unmarshal config change message", zap.Error(err))
			continue
		}

		key, _ := changeMsg["key"].(string)
		if key == "" {
			continue
		}

		if deleted, ok := changeMsg["deleted"].(bool); ok && deleted {
			c.localCache.Delete(key)
			logger.Info("Config deleted from cache", zap.String("key", key))
		} else if value, ok := changeMsg["value"]; ok {
			c.setLocal(key, value)
			c.notifySubscribers(key, value)
			logger.Info("Config updated in cache", zap.String("key", key))
		}
	}
}

// cacheEntry 缓存条目
type cacheEntry struct {
	value    interface{}
	expireAt time.Time
}

// GetString 获取字符串配置
func (c *ConfigCache) GetString(ctx context.Context, key string, defaultValue string) string {
	if value, ok := c.Get(ctx, key); ok {
		if s, ok := value.(string); ok {
			return s
		}
	}
	return defaultValue
}

// GetInt 获取整数配置
func (c *ConfigCache) GetInt(ctx context.Context, key string, defaultValue int) int {
	if value, ok := c.Get(ctx, key); ok {
		switch v := value.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case int64:
			return int(v)
		}
	}
	return defaultValue
}

// GetBool 获取布尔配置
func (c *ConfigCache) GetBool(ctx context.Context, key string, defaultValue bool) bool {
	if value, ok := c.Get(ctx, key); ok {
		if b, ok := value.(bool); ok {
			return b
		}
	}
	return defaultValue
}

// GetDuration 获取时间间隔配置
func (c *ConfigCache) GetDuration(ctx context.Context, key string, defaultValue time.Duration) time.Duration {
	if value, ok := c.Get(ctx, key); ok {
		switch v := value.(type) {
		case string:
			if d, err := time.ParseDuration(v); err == nil {
				return d
			}
		case float64:
			return time.Duration(v) * time.Second
		case int64:
			return time.Duration(v) * time.Second
		}
	}
	return defaultValue
}

// Clear 清空所有缓存
func (c *ConfigCache) Clear(ctx context.Context) error {
	// 清空本地缓存
	c.localCache.Range(func(key, _ interface{}) bool {
		c.localCache.Delete(key)
		return true
	})

	// 清空Redis中的配置缓存
	if c.redis != nil {
		iter := c.redis.Scan(ctx, 0, c.prefix+"*", 100).Iterator()
		for iter.Next(ctx) {
			c.redis.Del(ctx, iter.Val())
		}
		if err := iter.Err(); err != nil {
			return fmt.Errorf("clear redis cache: %w", err)
		}
	}

	return nil
}
