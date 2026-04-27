package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/chainspace/backend/internal/model"
	"github.com/chainspace/backend/internal/pkg/logger"
	"github.com/chainspace/backend/internal/repository"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const auditWriteTimeout = 3 * time.Second

// AuditLogger 为写操作请求记录统一的审计日志。
func AuditLogger(logRepo *repository.OperationLogRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !shouldAuditMethod(c.Request.Method) {
			c.Next()
			return
		}

		requestBody := readRequestBody(c)
		c.Next()

		userID, exists := c.Get(ContextKeyUserID)
		if !exists {
			return
		}

		module, action := parseAuditModuleAction(c.FullPath(), c.Request.Method)
		if module == "" || action == "" {
			return
		}

		logEntry := &model.OperationLog{
			UserID:       userID.(uint),
			SchoolID:     extractAuditSchoolID(c),
			Module:       module,
			Action:       action,
			Description:  buildAuditDescription(module, action),
			RequestIP:    c.ClientIP(),
			UserAgent:    c.Request.UserAgent(),
			RequestData:  sanitizeAuditRequestData(requestBody),
			ResponseCode: c.Writer.Status(),
		}

		// 审计日志属于平台基础能力，不应随着请求上下文取消而丢失。
		go persistAuditLog(logRepo, logEntry)
	}
}

// shouldAuditMethod 判断当前 HTTP 方法是否属于需要记录的写操作。
func shouldAuditMethod(method string) bool {
	switch method {
	case "POST", "PUT", "PATCH", "DELETE":
		return true
	default:
		return false
	}
}

// readRequestBody 在不中断后续绑定的前提下复制请求体。
func readRequestBody(c *gin.Context) []byte {
	if c.Request.Body == nil {
		return nil
	}

	requestBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil
	}

	c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
	return requestBody
}

// parseAuditModuleAction 以路由模板和 HTTP 方法推导审计模块与动作。
func parseAuditModuleAction(fullPath, method string) (string, string) {
	trimmedPath := strings.TrimPrefix(fullPath, "/api/v1/")
	trimmedPath = strings.Trim(trimmedPath, "/")
	if trimmedPath == "" {
		return "", ""
	}

	parts := strings.Split(trimmedPath, "/")
	module := parts[0]
	action := actionFromMethod(method)

	if len(parts) > 1 {
		lastPart := parts[len(parts)-1]
		if derivedAction, ok := deriveActionFromPath(lastPart); ok {
			action = derivedAction
		}
	}

	return module, action
}

// actionFromMethod 根据 HTTP 方法映射基础动作。
func actionFromMethod(method string) string {
	switch method {
	case "POST":
		return "create"
	case "PUT", "PATCH":
		return "update"
	case "DELETE":
		return "delete"
	default:
		return ""
	}
}

// deriveActionFromPath 针对显式动作型子路由推导更具体的审计动作。
func deriveActionFromPath(lastPart string) (string, bool) {
	switch lastPart {
	case "login", "logout", "status", "password", "publish", "start", "stop", "pause", "resume":
		return lastPart, true
	default:
		return "", false
	}
}

// extractAuditSchoolID 从上下文中提取审计日志需要的学校信息。
func extractAuditSchoolID(c *gin.Context) *uint {
	if schoolID, ok := c.Get(ContextKeySchoolID); ok {
		if value, typeOK := schoolID.(uint); typeOK && value > 0 {
			return &value
		}
	}
	return nil
}

// buildAuditDescription 生成统一且中性的审计描述文本。
func buildAuditDescription(module, action string) string {
	return action + ":" + module
}

// sanitizeAuditRequestData 对请求体做脱敏后再写入审计日志。
func sanitizeAuditRequestData(requestBody []byte) model.JSONMap {
	if len(requestBody) == 0 {
		return nil
	}

	var requestData map[string]interface{}
	if err := json.Unmarshal(requestBody, &requestData); err != nil {
		return nil
	}

	maskSensitiveFields(requestData)
	return requestData
}

// maskSensitiveFields 递归脱敏常见敏感字段，避免原始凭据落入审计日志。
func maskSensitiveFields(data map[string]interface{}) {
	sensitiveFields := map[string]struct{}{
		"password":     {},
		"old_password": {},
		"new_password": {},
		"token":        {},
		"secret":       {},
		"api_key":      {},
	}

	for key, value := range data {
		if _, ok := sensitiveFields[strings.ToLower(key)]; ok {
			data[key] = "******"
			continue
		}

		switch typedValue := value.(type) {
		case map[string]interface{}:
			maskSensitiveFields(typedValue)
		case []interface{}:
			maskSensitiveSlice(typedValue)
		}
	}
}

// maskSensitiveSlice 递归处理数组中的嵌套对象。
func maskSensitiveSlice(items []interface{}) {
	for _, item := range items {
		switch typedItem := item.(type) {
		case map[string]interface{}:
			maskSensitiveFields(typedItem)
		case []interface{}:
			maskSensitiveSlice(typedItem)
		}
	}
}

// persistAuditLog 使用独立超时上下文异步写入审计日志。
func persistAuditLog(logRepo *repository.OperationLogRepository, logEntry *model.OperationLog) {
	ctx, cancel := context.WithTimeout(context.Background(), auditWriteTimeout)
	defer cancel()

	if err := logRepo.Create(ctx, logEntry); err != nil {
		logger.Warn("Write audit log failed", zap.Error(err))
	}
}
