package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/chainspace/backend/internal/config"
	"github.com/chainspace/backend/internal/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var rdb *redis.Client

// Init 初始化 Redis 连接并缓存全局客户端。
func Init(cfg *config.RedisConfig) (*redis.Client, error) {
	rdb = redis.NewClient(&redis.Options{
		Addr:         cfg.Addr(),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	logger.Info("Redis connected",
		zap.String("addr", cfg.Addr()),
		zap.Int("db", cfg.DB),
	)

	return rdb, nil
}

// Client 返回当前缓存的全局 Redis 客户端。
func Client() *redis.Client {
	return rdb
}

// Close 关闭当前全局 Redis 连接。
func Close() error {
	if rdb != nil {
		return rdb.Close()
	}
	return nil
}
