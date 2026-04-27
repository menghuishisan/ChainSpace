package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/chainspace/backend/internal/pkg/logger"
	"github.com/chainspace/backend/internal/pkg/response"
	"github.com/chainspace/backend/pkg/errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Recovery 捕获请求处理过程中的 panic 并返回统一错误响应。
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logRecovery(c, recovered)
				c.AbortWithStatusJSON(http.StatusInternalServerError, response.Response{
					Code:    errors.ErrInternal.Code,
					Message: errors.ErrInternal.Message,
				})
			}
		}()

		c.Next()
	}
}

// RecoveryWithWriter 捕获 panic 并额外输出到自定义写入器。
func RecoveryWithWriter(out func(string)) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				out(fmt.Sprintf("[Recovery] panic recovered:\n%v\n%s", recovered, debug.Stack()))
				logRecovery(c, recovered)
				c.AbortWithStatusJSON(http.StatusInternalServerError, response.Response{
					Code:    errors.ErrInternal.Code,
					Message: errors.ErrInternal.Message,
				})
			}
		}()

		c.Next()
	}
}

// logRecovery 记录包含请求上下文的 panic 日志。
func logRecovery(c *gin.Context, recovered interface{}) {
	fields := []zap.Field{
		zap.Any("error", recovered),
		zap.String("stack", string(debug.Stack())),
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
		zap.String("ip", c.ClientIP()),
	}

	if requestID, exists := c.Get(ContextKeyRequestID); exists {
		if requestIDValue, ok := requestID.(string); ok && requestIDValue != "" {
			fields = append(fields, zap.String("request_id", requestIDValue))
		}
	}

	logger.Error("Panic recovered", fields...)
}
