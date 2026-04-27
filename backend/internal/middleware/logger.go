package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/chainspace/backend/internal/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	// ContextKeyRequestID 是请求链路追踪 ID 在 gin 上下文中的统一键名。
	ContextKeyRequestID = "request_id"
	requestIDHeader     = "X-Request-ID"
)

// Logger 记录统一的请求访问日志。
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		fields := []zap.Field{
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("query", c.Request.URL.RawQuery),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Duration("latency", time.Since(start)),
		}

		if requestID, exists := c.Get(ContextKeyRequestID); exists {
			if requestIDValue, ok := requestID.(string); ok && requestIDValue != "" {
				fields = append(fields, zap.String("request_id", requestIDValue))
			}
		}

		if userID, exists := c.Get(ContextKeyUserID); exists {
			fields = append(fields, zap.Uint("user_id", userID.(uint)))
		}

		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.String()))
		}

		switch {
		case c.Writer.Status() >= 500:
			logger.Error("Request completed with server error", fields...)
		case c.Writer.Status() >= 400:
			logger.Warn("Request completed with client error", fields...)
		default:
			logger.Info("Request completed", fields...)
		}
	}
}

// RequestID 为请求设置统一的链路追踪 ID。
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(requestIDHeader)
		if requestID == "" {
			requestID = generateRequestID()
		}

		c.Set(ContextKeyRequestID, requestID)
		c.Header(requestIDHeader, requestID)
		c.Next()
	}
}

// generateRequestID 生成高熵请求 ID，避免时间戳随机串碰撞。
func generateRequestID() string {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		return time.Now().UTC().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(buffer)
}
