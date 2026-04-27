package middleware

import (
	"strconv"
	"sync"
	"time"

	"github.com/chainspace/backend/internal/config"
	"github.com/chainspace/backend/internal/pkg/response"
	"github.com/chainspace/backend/pkg/errors"
	"github.com/gin-gonic/gin"
)

// RateLimiter 实现当前项目实际使用的进程内令牌桶限流器。
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     int
	burst    int
	cleanup  time.Duration
}

// visitor 记录单个限流键的令牌桶状态。
type visitor struct {
	tokens     float64
	lastUpdate time.Time
}

// NewRateLimiter 根据全局配置创建统一限流器。
func NewRateLimiter(cfg *config.RateLimitConfig) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     cfg.RequestsPerSecond,
		burst:    cfg.Burst,
		cleanup:  time.Minute,
	}

	// 后台定期清理闲置键，避免访问记录长期累积。
	go rl.cleanupLoop()

	return rl
}

// Allow 按键名执行一次令牌桶检查。
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	currentVisitor, exists := rl.visitors[key]
	now := time.Now()

	if !exists {
		rl.visitors[key] = &visitor{
			tokens:     float64(rl.burst - 1),
			lastUpdate: now,
		}
		return true
	}

	// 根据上次访问时间补充令牌，再决定本次请求是否通过。
	elapsed := now.Sub(currentVisitor.lastUpdate).Seconds()
	currentVisitor.tokens += elapsed * float64(rl.rate)
	if currentVisitor.tokens > float64(rl.burst) {
		currentVisitor.tokens = float64(rl.burst)
	}
	currentVisitor.lastUpdate = now

	if currentVisitor.tokens >= 1 {
		currentVisitor.tokens--
		return true
	}

	return false
}

// cleanupLoop 定期清理长时间未访问的限流记录。
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		threshold := time.Now().Add(-rl.cleanup)
		for key, currentVisitor := range rl.visitors {
			if currentVisitor.lastUpdate.Before(threshold) {
				delete(rl.visitors, key)
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimit 为全局 API 请求提供统一限流中间件。
func RateLimit(cfg *config.RateLimitConfig) gin.HandlerFunc {
	if !cfg.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	limiter := NewRateLimiter(cfg)

	return func(c *gin.Context) {
		// 默认按客户端 IP 限流；登录后统一按用户 ID 聚合。
		key := c.ClientIP()
		if userID, exists := GetUserID(c); exists {
			key = strconv.FormatUint(uint64(userID), 10)
		}

		if !limiter.Allow(key) {
			response.Error(c, errors.ErrTooManyRequests)
			c.Abort()
			return
		}

		c.Next()
	}
}
