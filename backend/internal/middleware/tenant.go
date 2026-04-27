package middleware

import (
	"strconv"

	"github.com/chainspace/backend/internal/model"
	"github.com/chainspace/backend/internal/pkg/response"
	"github.com/chainspace/backend/pkg/errors"
	"github.com/gin-gonic/gin"
)

// TenantIsolation 多租户数据隔离中间件
// 确保非平台管理员只能访问自己学校的数据
func TenantIsolation() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, roleExists := GetRole(c)
		if !roleExists {
			response.Error(c, errors.ErrLoginRequired)
			c.Abort()
			return
		}

		// 平台管理员不受限制
		if role == model.RolePlatformAdmin {
			c.Next()
			return
		}

		// 其他角色必须有学校ID
		schoolID, exists := GetSchoolID(c)
		if !exists || schoolID == 0 {
			response.Error(c, errors.ErrNoPermission.WithMessage("未关联学校"))
			c.Abort()
			return
		}

		c.Next()
	}
}

// ValidateSchoolAccess 验证学校访问权限
// 检查用户是否有权限访问指定的学校数据
func ValidateSchoolAccess(targetSchoolID uint) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, roleExists := GetRole(c)
		if !roleExists {
			response.Error(c, errors.ErrLoginRequired)
			c.Abort()
			return
		}

		// 平台管理员可以访问任何学校
		if role == model.RolePlatformAdmin {
			c.Next()
			return
		}

		// 其他角色只能访问自己学校的数据
		userSchoolID, exists := GetSchoolID(c)
		if !exists || userSchoolID != targetSchoolID {
			response.Error(c, errors.ErrNoPermission.WithMessage("无权访问其他学校数据"))
			c.Abort()
			return
		}

		c.Next()
	}
}

// ExtractSchoolIDFromPath 从Path参数中提取学校ID并验证权限
func ExtractSchoolIDFromPath(paramName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		schoolIDStr := c.Param(paramName)
		if schoolIDStr == "" {
			c.Next()
			return
		}

		schoolID, err := strconv.ParseUint(schoolIDStr, 10, 32)
		if err != nil {
			response.Error(c, errors.ErrInvalidParams.WithMessage("无效的学校ID"))
			c.Abort()
			return
		}

		role, _ := GetRole(c)
		if role != model.RolePlatformAdmin {
			userSchoolID, exists := GetSchoolID(c)
			if !exists || userSchoolID != uint(schoolID) {
				response.Error(c, errors.ErrNoPermission.WithMessage("无权访问其他学校数据"))
				c.Abort()
				return
			}
		}

		c.Set("path_school_id", uint(schoolID))
		c.Next()
	}
}

// SchoolContext 学校上下文中间件
// 为请求设置当前学校上下文
func SchoolContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := GetRole(c)

		if role == model.RolePlatformAdmin {
			if schoolIDHeader := c.GetHeader("X-School-ID"); schoolIDHeader != "" {
				schoolID, err := strconv.ParseUint(schoolIDHeader, 10, 32)
				if err == nil && schoolID > 0 {
					c.Set(ContextKeySchoolID, uint(schoolID))
				}
			}
		}

		c.Next()
	}
}
